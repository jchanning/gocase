package llm

import (
	"testing"
)

func TestParseQuestions_ValidJSON(t *testing.T) {
	raw := `[
		{
			"question_text": "What is photosynthesis?",
			"options": ["A process", "A protein", "A cell", "A molecule"],
			"correct_index": 0,
			"points": 1
		},
		{
			"question_text": "Which organelle?",
			"options": ["Nucleus", "Chloroplast", "Mitochondria", "Ribosome"],
			"correct_index": 1,
			"points": 2
		}
	]`

	questions, err := parseQuestions(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(questions) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(questions))
	}
	if questions[0].QuestionText != "What is photosynthesis?" {
		t.Errorf("unexpected question text: %s", questions[0].QuestionText)
	}
	if questions[1].CorrectIndex != 1 {
		t.Errorf("expected correct_index 1, got %d", questions[1].CorrectIndex)
	}
	if questions[1].Points != 2 {
		t.Errorf("expected points 2, got %d", questions[1].Points)
	}
}

func TestParseQuestions_MarkdownFences(t *testing.T) {
	raw := "```json\n[{\"question_text\":\"Q1?\",\"options\":[\"A\",\"B\",\"C\",\"D\"],\"correct_index\":2,\"points\":1}]\n```"
	questions, err := parseQuestions(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(questions))
	}
	if questions[0].CorrectIndex != 2 {
		t.Errorf("expected correct_index 2, got %d", questions[0].CorrectIndex)
	}
}

func TestParseQuestions_PrefixText(t *testing.T) {
	raw := `Here are the questions:
[{"question_text":"Q?","options":["1","2","3","4"],"correct_index":0,"points":1}]
Hope this helps!`
	questions, err := parseQuestions(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(questions))
	}
}

func TestParseQuestions_InvalidCorrectIndex(t *testing.T) {
	raw := `[{"question_text":"Q?","options":["A","B","C","D"],"correct_index":5,"points":1}]`
	_, err := parseQuestions(raw)
	if err == nil {
		t.Fatal("expected error for invalid correct_index, got nil")
	}
}

func TestParseQuestions_WrongOptionCount(t *testing.T) {
	raw := `[{"question_text":"Q?","options":["A","B"],"correct_index":0,"points":1}]`
	_, err := parseQuestions(raw)
	if err == nil {
		t.Fatal("expected error for wrong option count, got nil")
	}
}

func TestParseQuestions_NoJSON(t *testing.T) {
	raw := "I cannot generate questions from this text."
	_, err := parseQuestions(raw)
	if err == nil {
		t.Fatal("expected error for no JSON, got nil")
	}
}

func TestParseQuestions_EmptyQuestionText(t *testing.T) {
	raw := `[{"question_text":"","options":["A","B","C","D"],"correct_index":0,"points":1}]`
	_, err := parseQuestions(raw)
	if err == nil {
		t.Fatal("expected error for empty question_text, got nil")
	}
}

func TestParseQuestions_DefaultPoints(t *testing.T) {
	raw := `[{"question_text":"Q?","options":["A","B","C","D"],"correct_index":0,"points":0}]`
	questions, err := parseQuestions(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if questions[0].Points != 1 {
		t.Errorf("expected default points 1, got %d", questions[0].Points)
	}
}

func TestBuildPrompt_Defaults(t *testing.T) {
	cfg := GenerateConfig{
		Subject:      "Science",
		ExamStandard: "GCSE",
		Difficulty:   "Medium",
	}
	prompt := buildPrompt("Some notes here.", cfg)
	if prompt == "" {
		t.Fatal("prompt should not be empty")
	}
	// Default should be 10 questions
	if !containsString(prompt, "10 multiple-choice") {
		t.Error("prompt should mention 10 questions by default")
	}
}

func TestBuildPrompt_ClampMax(t *testing.T) {
	cfg := GenerateConfig{
		Subject:      "Math",
		ExamStandard: "A-Level",
		Difficulty:   "Hard",
		NumQuestions: 50,
	}
	prompt := buildPrompt("Notes.", cfg)
	if !containsString(prompt, "30 multiple-choice") {
		t.Error("NumQuestions should be clamped to 30")
	}
}

func TestConfigIsConfigured(t *testing.T) {
	empty := Config{}
	if empty.IsConfigured() {
		t.Error("empty config should not be configured")
	}

	full := Config{
		TenancyOCID:    "ocid1.tenancy.oc1..abc",
		UserOCID:       "ocid1.user.oc1..abc",
		Fingerprint:    "aa:bb:cc",
		PrivateKeyPath: "/tmp/key.pem",
		Region:         "uk-london-1",
		CompartmentID:  "ocid1.compartment.oc1..abc",
		ModelID:        "cohere.command-r-plus",
	}
	if !full.IsConfigured() {
		t.Error("fully populated config should be configured")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
