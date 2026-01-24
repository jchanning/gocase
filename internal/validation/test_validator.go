package validation

import (
	"fmt"

	"my-app/internal/models"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// TestValidator validates test data
type TestValidator struct {
	errors []ValidationError
}

// NewTestValidator creates a new test validator
func NewTestValidator() *TestValidator {
	return &TestValidator{
		errors: []ValidationError{},
	}
}

// ValidateTest validates test data
func (v *TestValidator) ValidateTest(test *models.Test) bool {
	v.errors = []ValidationError{} // Reset errors

	if test.Title == "" {
		v.addError("title", "Test title is required")
	} else if len(test.Title) > 255 {
		v.addError("title", "Test title must not exceed 255 characters")
	}

	if test.Description == "" {
		v.addError("description", "Test description is required")
	}

	if test.ExamStandard == "" {
		v.addError("exam_standard", "Exam standard is required")
	} else if !isValidExamStandard(test.ExamStandard) {
		v.addError("exam_standard", "Invalid exam standard. Must be GCSE, A-Level, Primary, or Secondary")
	}

	if test.Difficulty == "" {
		v.addError("difficulty", "Difficulty is required")
	} else if !isValidDifficulty(test.Difficulty) {
		v.addError("difficulty", "Invalid difficulty. Must be Easy, Medium, or Hard")
	}

	if test.TimeLimitMinutes <= 0 {
		v.addError("time_limit_minutes", "Time limit must be greater than 0")
	}

	if test.PassingScore < 0 || test.PassingScore > 100 {
		v.addError("passing_score", "Passing score must be between 0 and 100")
	}

	return len(v.errors) == 0
}

// ValidateQuestion validates question data
func (v *TestValidator) ValidateQuestion(question *models.Question) bool {
	tempErrors := v.errors
	v.errors = []ValidationError{} // Reset errors for this validation

	if question.QuestionText == "" {
		v.addError("question_text", "Question text is required")
	} else if len(question.QuestionText) > 5000 {
		v.addError("question_text", "Question text must not exceed 5000 characters")
	}

	if question.Points <= 0 {
		v.addError("points", "Points must be greater than 0")
	}

	if len(question.Options) != 4 {
		v.addError("options", "A question must have exactly 4 answer options")
	} else {
		// Validate options and ensure one is marked as correct
		hasCorrect := false
		for i, opt := range question.Options {
			if opt.OptionText == "" {
				v.addError("options", fmt.Sprintf("Option %d text is required", i+1))
			}
			if opt.IsCorrect {
				if hasCorrect {
					v.addError("options", "Only one option can be marked as correct")
					break
				}
				hasCorrect = true
			}
		}
		if !hasCorrect {
			v.addError("options", "One option must be marked as correct")
		}
	}

	isValid := len(v.errors) == 0
	if !isValid {
		// Keep the test errors if this validation fails
		return isValid
	}
	// Restore previous errors if this validation passes
	v.errors = tempErrors
	return isValid
}

// ValidateAnswerOption validates answer option data
func (v *TestValidator) ValidateAnswerOption(option *models.AnswerOption) bool {
	v.errors = []ValidationError{} // Reset errors

	if option.OptionText == "" {
		v.addError("option_text", "Option text is required")
	} else if len(option.OptionText) > 1000 {
		v.addError("option_text", "Option text must not exceed 1000 characters")
	}

	if option.OptionOrder < 1 || option.OptionOrder > 4 {
		v.addError("option_order", "Option order must be between 1 and 4")
	}

	return len(v.errors) == 0
}

// GetErrors returns all validation errors
func (v *TestValidator) GetErrors() []ValidationError {
	return v.errors
}

// GetErrorMessages returns error messages as a map
func (v *TestValidator) GetErrorMessages() map[string]string {
	messages := make(map[string]string)
	for _, err := range v.errors {
		messages[err.Field] = err.Message
	}
	return messages
}

// Helper functions

func (v *TestValidator) addError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

func isValidExamStandard(standard string) bool {
	standards := []string{"GCSE", "A-Level", "Primary", "Secondary"}
	for _, s := range standards {
		if s == standard {
			return true
		}
	}
	return false
}

func isValidDifficulty(difficulty string) bool {
	difficulties := []string{"Easy", "Medium", "Hard"}
	for _, d := range difficulties {
		if d == difficulty {
			return true
		}
	}
	return false
}
