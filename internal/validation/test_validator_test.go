package validation

import (
	"strings"
	"testing"

	"github.com/jchanning/gocase/internal/models"
)

// validTest returns a fully valid Test ready for mutation in negative-path tests.
func validTest() *models.Test {
	return &models.Test{
		Title:            "Algebra Basics",
		Description:      "Core algebra for GCSE students",
		ExamStandard:     "GCSE",
		Difficulty:       "Medium",
		TimeLimitMinutes: 30,
		PassingScore:     60,
	}
}

// validQuestion returns a fully valid Question with four options, one correct.
func validQuestion() *models.Question {
	return &models.Question{
		QuestionText: "What is 2 + 2?",
		Points:       1,
		Options: []models.AnswerOption{
			{OptionText: "3", IsCorrect: false},
			{OptionText: "4", IsCorrect: true},
			{OptionText: "5", IsCorrect: false},
			{OptionText: "6", IsCorrect: false},
		},
	}
}

// validAnswerOption returns a fully valid AnswerOption.
func validAnswerOption() *models.AnswerOption {
	return &models.AnswerOption{
		OptionText:  "My answer",
		OptionOrder: 1,
	}
}

// ---- ValidateTest ----

func TestValidateTest_ValidInputPasses(t *testing.T) {
	v := NewTestValidator()
	if !v.ValidateTest(validTest()) {
		t.Fatalf("expected valid test to pass, got errors: %v", v.GetErrors())
	}
}

func TestValidateTest_AllExamStandards(t *testing.T) {
	standards := []string{"GCSE", "IGCSE", "A-Level", "Primary", "Secondary"}
	for _, std := range standards {
		t.Run(std, func(t *testing.T) {
			test := validTest()
			test.ExamStandard = std
			v := NewTestValidator()
			if !v.ValidateTest(test) {
				t.Fatalf("expected exam standard %q to be valid, got errors: %v", std, v.GetErrors())
			}
		})
	}
}

func TestValidateTest_AllDifficulties(t *testing.T) {
	for _, diff := range []string{"Easy", "Medium", "Hard"} {
		t.Run(diff, func(t *testing.T) {
			test := validTest()
			test.Difficulty = diff
			v := NewTestValidator()
			if !v.ValidateTest(test) {
				t.Fatalf("expected difficulty %q to be valid, got errors: %v", diff, v.GetErrors())
			}
		})
	}
}

func TestValidateTest_RequiresTitle(t *testing.T) {
	test := validTest()
	test.Title = ""
	v := NewTestValidator()
	if v.ValidateTest(test) {
		t.Fatal("expected validation to fail for empty title")
	}
	assertFieldError(t, v, "title")
}

func TestValidateTest_TitleMaxLength(t *testing.T) {
	test := validTest()
	test.Title = strings.Repeat("x", 256)
	v := NewTestValidator()
	if v.ValidateTest(test) {
		t.Fatal("expected validation to fail for title > 255 chars")
	}
	assertFieldError(t, v, "title")
}

func TestValidateTest_TitleAtMaxLength(t *testing.T) {
	test := validTest()
	test.Title = strings.Repeat("x", 255)
	v := NewTestValidator()
	if !v.ValidateTest(test) {
		t.Fatalf("expected 255-char title to pass, got errors: %v", v.GetErrors())
	}
}

func TestValidateTest_RequiresDescription(t *testing.T) {
	test := validTest()
	test.Description = ""
	v := NewTestValidator()
	if v.ValidateTest(test) {
		t.Fatal("expected validation to fail for empty description")
	}
	assertFieldError(t, v, "description")
}

func TestValidateTest_RequiresExamStandard(t *testing.T) {
	test := validTest()
	test.ExamStandard = ""
	v := NewTestValidator()
	if v.ValidateTest(test) {
		t.Fatal("expected validation to fail for empty exam standard")
	}
	assertFieldError(t, v, "exam_standard")
}

func TestValidateTest_RejectsInvalidExamStandard(t *testing.T) {
	test := validTest()
	test.ExamStandard = "SAT"
	v := NewTestValidator()
	if v.ValidateTest(test) {
		t.Fatal("expected validation to fail for unknown exam standard")
	}
	assertFieldError(t, v, "exam_standard")
}

func TestValidateTest_RequiresDifficulty(t *testing.T) {
	test := validTest()
	test.Difficulty = ""
	v := NewTestValidator()
	if v.ValidateTest(test) {
		t.Fatal("expected validation to fail for empty difficulty")
	}
	assertFieldError(t, v, "difficulty")
}

func TestValidateTest_RejectsInvalidDifficulty(t *testing.T) {
	test := validTest()
	test.Difficulty = "Extreme"
	v := NewTestValidator()
	if v.ValidateTest(test) {
		t.Fatal("expected validation to fail for unknown difficulty")
	}
	assertFieldError(t, v, "difficulty")
}

func TestValidateTest_RequiresPositiveTimeLimit(t *testing.T) {
	for _, limit := range []int{0, -1, -100} {
		t.Run("limit", func(t *testing.T) {
			test := validTest()
			test.TimeLimitMinutes = limit
			v := NewTestValidator()
			if v.ValidateTest(test) {
				t.Fatalf("expected validation to fail for time_limit_minutes=%d", limit)
			}
			assertFieldError(t, v, "time_limit_minutes")
		})
	}
}

func TestValidateTest_PassingScoreBounds(t *testing.T) {
	cases := []struct {
		score   int
		wantErr bool
	}{
		{0, false},
		{100, false},
		{50, false},
		{-1, true},
		{101, true},
	}
	for _, tc := range cases {
		t.Run("score", func(t *testing.T) {
			test := validTest()
			test.PassingScore = tc.score
			v := NewTestValidator()
			got := v.ValidateTest(test)
			if got == tc.wantErr {
				t.Fatalf("score=%d: expected valid=%v but got valid=%v, errors=%v",
					tc.score, !tc.wantErr, got, v.GetErrors())
			}
		})
	}
}

// ---- ValidateQuestion ----

func TestValidateQuestion_ValidInputPasses(t *testing.T) {
	v := NewTestValidator()
	if !v.ValidateQuestion(validQuestion()) {
		t.Fatalf("expected valid question to pass, got errors: %v", v.GetErrors())
	}
}

func TestValidateQuestion_RequiresQuestionText(t *testing.T) {
	q := validQuestion()
	q.QuestionText = ""
	v := NewTestValidator()
	if v.ValidateQuestion(q) {
		t.Fatal("expected validation to fail for empty question text")
	}
	assertFieldError(t, v, "question_text")
}

func TestValidateQuestion_QuestionTextMaxLength(t *testing.T) {
	q := validQuestion()
	q.QuestionText = strings.Repeat("a", 5001)
	v := NewTestValidator()
	if v.ValidateQuestion(q) {
		t.Fatal("expected validation to fail for question text > 5000 chars")
	}
	assertFieldError(t, v, "question_text")
}

func TestValidateQuestion_RequiresPositivePoints(t *testing.T) {
	for _, pts := range []int{0, -1} {
		t.Run("points", func(t *testing.T) {
			q := validQuestion()
			q.Points = pts
			v := NewTestValidator()
			if v.ValidateQuestion(q) {
				t.Fatalf("expected validation to fail for points=%d", pts)
			}
			assertFieldError(t, v, "points")
		})
	}
}

func TestValidateQuestion_RequiresExactlyFourOptions(t *testing.T) {
	cases := []struct {
		count int
	}{
		{0}, {1}, {2}, {3}, {5},
	}
	opt := models.AnswerOption{OptionText: "opt", IsCorrect: false}
	for _, tc := range cases {
		t.Run("options", func(t *testing.T) {
			q := validQuestion()
			q.Options = make([]models.AnswerOption, tc.count)
			for i := range q.Options {
				q.Options[i] = opt
			}
			if tc.count > 0 {
				q.Options[0].IsCorrect = true
			}
			v := NewTestValidator()
			if v.ValidateQuestion(q) {
				t.Fatalf("expected validation to fail for %d options", tc.count)
			}
			assertFieldError(t, v, "options")
		})
	}
}

func TestValidateQuestion_RequiresOneCorrectOption(t *testing.T) {
	q := validQuestion()
	for i := range q.Options {
		q.Options[i].IsCorrect = false
	}
	v := NewTestValidator()
	if v.ValidateQuestion(q) {
		t.Fatal("expected validation to fail when no option is correct")
	}
	assertFieldError(t, v, "options")
}

func TestValidateQuestion_RejectsMultipleCorrectOptions(t *testing.T) {
	q := validQuestion()
	q.Options[0].IsCorrect = true
	q.Options[1].IsCorrect = true
	v := NewTestValidator()
	if v.ValidateQuestion(q) {
		t.Fatal("expected validation to fail when multiple options are correct")
	}
	assertFieldError(t, v, "options")
}

func TestValidateQuestion_RejectsEmptyOptionText(t *testing.T) {
	q := validQuestion()
	q.Options[2].OptionText = ""
	v := NewTestValidator()
	if v.ValidateQuestion(q) {
		t.Fatal("expected validation to fail when an option has empty text")
	}
	assertFieldError(t, v, "options")
}

// ---- ValidateAnswerOption ----

func TestValidateAnswerOption_ValidInputPasses(t *testing.T) {
	v := NewTestValidator()
	if !v.ValidateAnswerOption(validAnswerOption()) {
		t.Fatalf("expected valid answer option to pass, got errors: %v", v.GetErrors())
	}
}

func TestValidateAnswerOption_RequiresOptionText(t *testing.T) {
	opt := validAnswerOption()
	opt.OptionText = ""
	v := NewTestValidator()
	if v.ValidateAnswerOption(opt) {
		t.Fatal("expected validation to fail for empty option text")
	}
	assertFieldError(t, v, "option_text")
}

func TestValidateAnswerOption_OptionTextMaxLength(t *testing.T) {
	opt := validAnswerOption()
	opt.OptionText = strings.Repeat("x", 1001)
	v := NewTestValidator()
	if v.ValidateAnswerOption(opt) {
		t.Fatal("expected validation to fail for option text > 1000 chars")
	}
	assertFieldError(t, v, "option_text")
}

func TestValidateAnswerOption_OptionTextAtMaxLength(t *testing.T) {
	opt := validAnswerOption()
	opt.OptionText = strings.Repeat("x", 1000)
	v := NewTestValidator()
	if !v.ValidateAnswerOption(opt) {
		t.Fatalf("expected 1000-char option text to pass, got errors: %v", v.GetErrors())
	}
}

func TestValidateAnswerOption_OptionOrderBounds(t *testing.T) {
	cases := []struct {
		order   int
		wantErr bool
	}{
		{1, false},
		{2, false},
		{3, false},
		{4, false},
		{0, true},
		{5, true},
		{-1, true},
	}
	for _, tc := range cases {
		t.Run("order", func(t *testing.T) {
			opt := validAnswerOption()
			opt.OptionOrder = tc.order
			v := NewTestValidator()
			got := v.ValidateAnswerOption(opt)
			if got == tc.wantErr {
				t.Fatalf("OptionOrder=%d: expected valid=%v but got valid=%v, errors=%v",
					tc.order, !tc.wantErr, got, v.GetErrors())
			}
		})
	}
}

// ---- GetErrorMessages ----

func TestGetErrorMessages_ReturnsFieldMap(t *testing.T) {
	test := validTest()
	test.Title = ""
	test.Description = ""
	v := NewTestValidator()
	v.ValidateTest(test)

	msgs := v.GetErrorMessages()
	if _, ok := msgs["title"]; !ok {
		t.Error("expected 'title' in error messages map")
	}
	if _, ok := msgs["description"]; !ok {
		t.Error("expected 'description' in error messages map")
	}
}

// ---- helper ----

func assertFieldError(t *testing.T, v *TestValidator, field string) {
	t.Helper()
	for _, e := range v.GetErrors() {
		if e.Field == field {
			return
		}
	}
	t.Fatalf("expected a validation error for field %q, got errors: %v", field, v.GetErrors())
}
