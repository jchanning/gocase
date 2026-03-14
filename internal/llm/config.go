package llm

import "os"

// Config holds OCI Generative AI connection settings.
type Config struct {
	TenancyOCID    string
	UserOCID       string
	Fingerprint    string
	PrivateKeyPath string
	Region         string
	CompartmentID  string
	ModelID        string
	Endpoint       string // optional override; derived from Region if empty
	MaxTokens      int
	Temperature    float64
}

// DefaultConfig returns a Config populated from environment variables
// with sensible defaults for optional fields.
func DefaultConfig() Config {
	cfg := Config{
		TenancyOCID:    os.Getenv("OCI_TENANCY_OCID"),
		UserOCID:       os.Getenv("OCI_USER_OCID"),
		Fingerprint:    os.Getenv("OCI_FINGERPRINT"),
		PrivateKeyPath: os.Getenv("OCI_PRIVATE_KEY_PATH"),
		Region:         os.Getenv("OCI_REGION"),
		CompartmentID:  os.Getenv("OCI_COMPARTMENT_ID"),
		ModelID:        os.Getenv("OCI_GENAI_MODEL_ID"),
		Endpoint:       os.Getenv("OCI_GENAI_ENDPOINT"),
		MaxTokens:      4096,
		Temperature:    0.3,
	}

	if cfg.Endpoint == "" && cfg.Region != "" {
		cfg.Endpoint = "https://inference.generativeai." + cfg.Region + ".oci.oraclecloud.com"
	}

	return cfg
}

// IsConfigured returns true when the minimum required OCI settings are present.
func (c Config) IsConfigured() bool {
	return c.TenancyOCID != "" &&
		c.UserOCID != "" &&
		c.Fingerprint != "" &&
		c.PrivateKeyPath != "" &&
		c.Region != "" &&
		c.CompartmentID != "" &&
		c.ModelID != ""
}
