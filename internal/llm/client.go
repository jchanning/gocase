package llm

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"my-app/internal/models"
)

// QuestionGenerator is the interface for generating questions from text.
type QuestionGenerator interface {
	GenerateQuestions(ctx context.Context, notesText string, cfg GenerateConfig) ([]models.QuestionUpload, error)
	IsAvailable() bool
}

// Client implements QuestionGenerator using the OCI Generative AI REST API.
type Client struct {
	cfg        Config
	httpClient *http.Client
	privateKey *rsa.PrivateKey

	// Simple concurrency limiter: one in-flight generation at a time.
	mu sync.Mutex
}

// NewClient creates a new OCI GenAI client. Returns an error if the
// private key cannot be loaded; call cfg.IsConfigured() first.
func NewClient(cfg Config) (*Client, error) {
	key, err := loadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load OCI private key: %w", err)
	}

	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		privateKey: key,
	}, nil
}

// IsAvailable reports whether the LLM backend is configured and reachable.
func (c *Client) IsAvailable() bool {
	return c != nil && c.cfg.IsConfigured() && c.privateKey != nil
}

// GenerateQuestions sends the notes text to OCI GenAI and returns parsed questions.
// It retries once if the LLM response fails to parse.
func (c *Client) GenerateQuestions(ctx context.Context, notesText string, cfg GenerateConfig) ([]models.QuestionUpload, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prompt := buildPrompt(notesText, cfg)

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := c.callGenerateText(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("OCI GenAI API call failed: %w", err)
		}

		questions, err := parseQuestions(raw)
		if err != nil {
			lastErr = err
			continue // retry once on parse failure
		}
		return questions, nil
	}

	return nil, fmt.Errorf("failed to parse LLM output after retries: %w", lastErr)
}

// ── OCI GenAI REST call ──────────────────────────────────────────────

// ociGenerateRequest is the request body for the OCI GenerateText action.
type ociGenerateRequest struct {
	CompartmentID    string              `json:"compartmentId"`
	ServingMode      ociServingMode      `json:"servingMode"`
	InferenceRequest ociInferenceRequest `json:"inferenceRequest"`
}

type ociServingMode struct {
	ServingType string `json:"servingType"`
	ModelID     string `json:"modelId"`
}

type ociInferenceRequest struct {
	RuntimeType string  `json:"runtimeType"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
}

// ociGenerateResponse is a minimal representation of the API response.
type ociGenerateResponse struct {
	InferenceResponse struct {
		GeneratedTexts [][]struct {
			Text string `json:"text"`
		} `json:"generatedTexts"`
		// For chat-style models
		Choices []struct {
			Message struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"message"`
			Text string `json:"text"`
		} `json:"choices"`
	} `json:"inferenceResponse"`
}

func (c *Client) callGenerateText(ctx context.Context, prompt string) (string, error) {
	reqBody := ociGenerateRequest{
		CompartmentID: c.cfg.CompartmentID,
		ServingMode: ociServingMode{
			ServingType: "ON_DEMAND",
			ModelID:     c.cfg.ModelID,
		},
		InferenceRequest: ociInferenceRequest{
			RuntimeType: "COHERE",
			Prompt:      prompt,
			MaxTokens:   c.cfg.MaxTokens,
			Temperature: c.cfg.Temperature,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := c.cfg.Endpoint + "/20231130/actions/generateText"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	if err := c.signRequest(req, bodyBytes); err != nil {
		return "", fmt.Errorf("sign request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OCI API returned %d: %s", resp.StatusCode, string(respBytes))
	}

	var ociResp ociGenerateResponse
	if err := json.Unmarshal(respBytes, &ociResp); err != nil {
		return "", fmt.Errorf("decode OCI response: %w", err)
	}

	// Extract text from the response — handle both Cohere and chat-style formats
	if texts := ociResp.InferenceResponse.GeneratedTexts; len(texts) > 0 && len(texts[0]) > 0 {
		return texts[0][0].Text, nil
	}
	if choices := ociResp.InferenceResponse.Choices; len(choices) > 0 {
		if choices[0].Text != "" {
			return choices[0].Text, nil
		}
		if len(choices[0].Message.Content) > 0 {
			return choices[0].Message.Content[0].Text, nil
		}
	}

	return "", fmt.Errorf("empty response from OCI GenAI")
}

// ── OCI HTTP Signature Auth ──────────────────────────────────────────

func (c *Client) signRequest(req *http.Request, body []byte) error {
	now := time.Now().UTC().Format(http.TimeFormat)
	req.Header.Set("date", now)

	bodyHash := sha256.Sum256(body)
	bodyB64 := base64.StdEncoding.EncodeToString(bodyHash[:])
	req.Header.Set("x-content-sha256", bodyB64)
	req.Header.Set("content-length", fmt.Sprintf("%d", len(body)))

	host := req.URL.Host
	target := fmt.Sprintf("%s %s", strings.ToLower(req.Method), req.URL.RequestURI())

	headers := "(request-target) date host x-content-sha256 content-type content-length"
	sigString := fmt.Sprintf(
		"(request-target): %s\ndate: %s\nhost: %s\nx-content-sha256: %s\ncontent-type: %s\ncontent-length: %d",
		target, now, host, bodyB64, req.Header.Get("Content-Type"), len(body),
	)

	hashed := sha256.Sum256([]byte(sigString))
	sig, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return err
	}

	keyID := fmt.Sprintf("%s/%s/%s", c.cfg.TenancyOCID, c.cfg.UserOCID, c.cfg.Fingerprint)
	authHeader := fmt.Sprintf(
		`Signature version="1",keyId="%s",algorithm="rsa-sha256",headers="%s",signature="%s"`,
		keyID, headers, base64.StdEncoding.EncodeToString(sig),
	)

	req.Header.Set("authorization", authHeader)
	return nil
}

// ── Key loading ──────────────────────────────────────────────────────

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in %s", path)
	}

	// Try PKCS#8 first, then PKCS#1
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, fmt.Errorf("key is not RSA")
	}

	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}
	return rsaKey, nil
}
