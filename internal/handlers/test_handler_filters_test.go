package handlers

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"my-app/internal/models"
)

func TestFilterTests_AppliesAllFilters(t *testing.T) {
	subjectID := 1
	tests := []models.Test{
		{ID: 1, Title: "Math Easy", SubjectID: &subjectID, Difficulty: "Easy", ExamStandard: "Primary", Published: true},
		{ID: 2, Title: "Math Hard", SubjectID: &subjectID, Difficulty: "Hard", ExamStandard: "GCSE", Published: true},
		{ID: 3, Title: "Science Easy", Difficulty: "Easy", ExamStandard: "Primary", Published: false},
	}

	filters := testFilters{SubjectID: &subjectID, Difficulty: "Easy", Standard: "Primary", Published: boolPtr(true)}

	filtered := filterTests(tests, filters)

	if len(filtered) != 1 {
		t.Fatalf("expected 1 test after filtering, got %d", len(filtered))
	}

	if filtered[0].ID != 1 {
		t.Fatalf("expected test ID 1 to remain, got %d", filtered[0].ID)
	}
}

func TestFilterTests_PublishedDefaultsToTrue(t *testing.T) {
	tests := []models.Test{
		{ID: 1, Title: "Published", Published: true},
		{ID: 2, Title: "Draft", Published: false},
	}

	req := httptest.NewRequest("GET", "/tests", nil)
	filters := parseTestFilters(req)

	filtered := filterTests(tests, filters)
	if len(filtered) != 1 {
		t.Fatalf("expected only published tests by default, got %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("expected published test to remain, got id %d", filtered[0].ID)
	}
}

func TestParseTestFilters_AllowsExplicitUnpublished(t *testing.T) {
	form := url.Values{}
	form.Set("published", "false")
	req := httptest.NewRequest("GET", "/tests?"+form.Encode(), nil)

	filters := parseTestFilters(req)
	if filters.Published == nil || *filters.Published {
		t.Fatalf("expected published filter to be false when explicitly requested")
	}
}

func boolPtr(v bool) *bool { return &v }
