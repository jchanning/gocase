package llm

import (
	"encoding/json"
	"fmt"
	"strings"

	"my-app/internal/models"
)

// GenerateConfig controls the question generation request.
type GenerateConfig struct {
	Subject      string
	ExamStandard string
	Difficulty   string
	NumQuestions int
	Points       int
}

// buildPrompt constructs the system + user prompt for question generation.
func buildPrompt(notesText string, cfg GenerateConfig) string {
	if cfg.NumQuestions <= 0 {
		cfg.NumQuestions = 10
	}
	if cfg.NumQuestions > 30 {
		cfg.NumQuestions = 30
	}
	if cfg.Points <= 0 {
		cfg.Points = 1
	}

	return fmt.Sprintf(`You are an expert educational assessment creator. Given the study notes below, generate exactly %d multiple-choice questions suitable for a %s %s exam at %s difficulty.

RULES:
1. Each question must have exactly 4 answer options.
2. Exactly one option must be correct.
3. correct_index is 0-based (0 to 3).
4. Points per question: %d.
5. Questions should cover different parts of the notes.
6. Options should be plausible — avoid obviously wrong distractors.
7. Include a concise explanation (1-3 sentences) for why the correct answer is right.
8. Respond ONLY with a valid JSON array, no markdown fences.

REQUIRED JSON FORMAT (array of objects):
[
  {
    "question_text": "...",
    "options": ["A", "B", "C", "D"],
    "correct_index": 0,
    "points": %d,
    "explanation": "The correct answer is A because ..."
  }
]

STUDY NOTES:
%s`, cfg.NumQuestions, cfg.ExamStandard, cfg.Subject, cfg.Difficulty, cfg.Points, cfg.Points, notesText)
}

// parseQuestions extracts a []QuestionUpload from the LLM's raw text response.
func parseQuestions(raw string) ([]models.QuestionUpload, error) {
	// Strip markdown code fences if the LLM wrapped its output
	text := strings.TrimSpace(raw)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	// Find the first '[' and last ']' to isolate the JSON array
	start := strings.Index(text, "[")
	end := strings.LastIndex(text, "]")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON array found in LLM response")
	}
	text = text[start : end+1]

	var questions []models.QuestionUpload
	if err := json.Unmarshal([]byte(text), &questions); err != nil {
		return nil, fmt.Errorf("invalid JSON from LLM: %w", err)
	}

	// Validate each question
	for i, q := range questions {
		if len(q.Options) != 4 {
			return nil, fmt.Errorf("question %d has %d options, expected 4", i+1, len(q.Options))
		}
		if q.CorrectIndex < 0 || q.CorrectIndex > 3 {
			return nil, fmt.Errorf("question %d has invalid correct_index %d", i+1, q.CorrectIndex)
		}
		if q.QuestionText == "" {
			return nil, fmt.Errorf("question %d has empty question_text", i+1)
		}
		if q.Points <= 0 {
			questions[i].Points = 1
		}
		// explanation is optional — no error if absent
	}

	return questions, nil
}
