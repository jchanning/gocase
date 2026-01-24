package handlers

import (
	"context"
	"fmt"

	"my-app/internal/models"
	"my-app/internal/repository"
	"my-app/internal/validation"
)

// validateTestUpload runs server-side validation for incoming test data before DB writes.
func validateTestUpload(upload models.TestUpload) map[string]string {
	errors := make(map[string]string)

	validator := validation.NewTestValidator()
	tempTest := &models.Test{
		Title:            upload.Title,
		Description:      upload.Description,
		ExamStandard:     upload.ExamStandard,
		Difficulty:       upload.Difficulty,
		TimeLimitMinutes: upload.TimeLimitMinutes,
		PassingScore:     upload.PassingScore,
	}

	if !validator.ValidateTest(tempTest) {
		for field, msg := range validator.GetErrorMessages() {
			errors[field] = msg
		}
	}

	if upload.Subject == "" {
		errors["subject"] = "Subject is required"
	}

	if len(upload.Questions) == 0 {
		errors["questions"] = "At least one question is required"
		return errors
	}

	for idx, q := range upload.Questions {
		opts := make([]models.AnswerOption, 0, len(q.Options))
		for i, opt := range q.Options {
			opts = append(opts, models.AnswerOption{
				OptionText:  opt,
				IsCorrect:   i == q.CorrectIndex,
				OptionOrder: i + 1,
			})
		}

		question := &models.Question{
			QuestionText: q.QuestionText,
			Points:       q.Points,
			Options:      opts,
		}

		qValidator := validation.NewTestValidator()
		if !qValidator.ValidateQuestion(question) {
			for field, msg := range qValidator.GetErrorMessages() {
				// Prefix with question index for clarity
				errors[fmt.Sprintf("question_%d_%s", idx+1, field)] = msg
			}
		}
	}

	return errors
}

// persistTestUpload stores the validated test definition and returns the created test.
func persistTestUpload(ctx context.Context, repo *repository.TestRepository, upload models.TestUpload, createdBy int) (*models.Test, error) {
	subjectID, err := repo.GetOrCreateSubject(ctx, upload.Subject, "")
	if err != nil {
		return nil, err
	}

	var topicID *int
	if upload.Topic != "" {
		id, err := repo.GetOrCreateTopic(ctx, subjectID, upload.Topic, "")
		if err != nil {
			return nil, err
		}
		topicID = &id
	}

	test := &models.Test{
		Title:            upload.Title,
		Description:      upload.Description,
		SubjectID:        &subjectID,
		TopicID:          topicID,
		ExamStandard:     upload.ExamStandard,
		Difficulty:       upload.Difficulty,
		TimeLimitMinutes: upload.TimeLimitMinutes,
		PassingScore:     upload.PassingScore,
		CreatedBy:        &createdBy,
	}

	if err := repo.Create(ctx, test); err != nil {
		return nil, err
	}

	for i, q := range upload.Questions {
		question := &models.Question{
			TestID:        test.ID,
			QuestionText:  q.QuestionText,
			QuestionOrder: i + 1,
			Points:        normalizePoints(q.Points),
		}

		if q.ImageURL != "" {
			question.ImageURL = &q.ImageURL
		}

		if err := repo.CreateQuestion(ctx, question); err != nil {
			return nil, err
		}

		for j, optText := range q.Options {
			option := &models.AnswerOption{
				QuestionID:  question.ID,
				OptionText:  optText,
				IsCorrect:   j == q.CorrectIndex,
				OptionOrder: j + 1,
			}

			if err := repo.CreateAnswerOption(ctx, option); err != nil {
				return nil, err
			}
		}
	}

	return test, nil
}

func normalizePoints(points int) int {
	if points <= 0 {
		return 1
	}
	return points
}
