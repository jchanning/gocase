package repository

import (
	"context"
	"fmt"

	"github.com/jchanning/gocase/internal/models"
)

// SyllabusRepository handles database operations for syllabi, sections, and topics.
type SyllabusRepository struct {
	pool dbQuerier
}

// NewSyllabusRepository creates a new SyllabusRepository.
func NewSyllabusRepository(pool dbQuerier) *SyllabusRepository {
	return &SyllabusRepository{pool: pool}
}

// GetAll returns all syllabi with subject join and topic count (admin/teacher view).
func (r *SyllabusRepository) GetAll(ctx context.Context) ([]models.Syllabus, error) {
	query := `
		SELECT sy.id, sy.subject_id, sy.exam_standard, sy.title,
		       COALESCE(sy.description, ''), sy.is_published, sy.created_by,
		       sy.created_at, sy.updated_at,
		       s.id, s.name, s.description,
		       (SELECT COUNT(*) FROM syllabus_topics st WHERE st.syllabus_id = sy.id)
		FROM syllabi sy
		LEFT JOIN subjects s ON sy.subject_id = s.id
		ORDER BY sy.created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Syllabus
	for rows.Next() {
		var sy models.Syllabus
		var subID *int
		var subName, subDesc *string
		if err := rows.Scan(
			&sy.ID, &sy.SubjectID, &sy.ExamStandard, &sy.Title,
			&sy.Description, &sy.IsPublished, &sy.CreatedBy,
			&sy.CreatedAt, &sy.UpdatedAt,
			&subID, &subName, &subDesc,
			&sy.TopicCount,
		); err != nil {
			return nil, err
		}
		if subID != nil {
			sy.Subject = &models.Subject{ID: *subID, Name: *subName, Description: *subDesc}
		}
		out = append(out, sy)
	}
	return out, rows.Err()
}

// GetPublished returns all published syllabi with subject join and topic count (student view).
func (r *SyllabusRepository) GetPublished(ctx context.Context) ([]models.Syllabus, error) {
	query := `
		SELECT sy.id, sy.subject_id, sy.exam_standard, sy.title,
		       COALESCE(sy.description, ''), sy.is_published, sy.created_by,
		       sy.created_at, sy.updated_at,
		       s.id, s.name, s.description,
		       (SELECT COUNT(*) FROM syllabus_topics st WHERE st.syllabus_id = sy.id)
		FROM syllabi sy
		LEFT JOIN subjects s ON sy.subject_id = s.id
		WHERE sy.is_published = TRUE
		ORDER BY s.name, sy.exam_standard, sy.title`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Syllabus
	for rows.Next() {
		var sy models.Syllabus
		var subID *int
		var subName, subDesc *string
		if err := rows.Scan(
			&sy.ID, &sy.SubjectID, &sy.ExamStandard, &sy.Title,
			&sy.Description, &sy.IsPublished, &sy.CreatedBy,
			&sy.CreatedAt, &sy.UpdatedAt,
			&subID, &subName, &subDesc,
			&sy.TopicCount,
		); err != nil {
			return nil, err
		}
		if subID != nil {
			sy.Subject = &models.Subject{ID: *subID, Name: *subName, Description: *subDesc}
		}
		out = append(out, sy)
	}
	return out, rows.Err()
}

// GetByID returns a single syllabus with its sections and topics (and linked tests per topic).
func (r *SyllabusRepository) GetByID(ctx context.Context, id int) (*models.Syllabus, error) {
	var sy models.Syllabus
	var subID *int
	var subName, subDesc *string
	err := r.pool.QueryRow(ctx, `
		SELECT sy.id, sy.subject_id, sy.exam_standard, sy.title,
		       COALESCE(sy.description, ''), sy.is_published, sy.created_by,
		       sy.created_at, sy.updated_at,
		       s.id, s.name, s.description
		FROM syllabi sy
		LEFT JOIN subjects s ON sy.subject_id = s.id
		WHERE sy.id = $1`, id).Scan(
		&sy.ID, &sy.SubjectID, &sy.ExamStandard, &sy.Title,
		&sy.Description, &sy.IsPublished, &sy.CreatedBy,
		&sy.CreatedAt, &sy.UpdatedAt,
		&subID, &subName, &subDesc,
	)
	if err != nil {
		return nil, err
	}
	if subID != nil {
		sy.Subject = &models.Subject{ID: *subID, Name: *subName, Description: *subDesc}
	}

	sections, err := r.getSectionsWithTopics(ctx, id)
	if err != nil {
		return nil, err
	}
	sy.Sections = sections

	count := 0
	for _, sec := range sections {
		count += len(sec.Topics)
	}
	sy.TopicCount = count

	return &sy, nil
}

func (r *SyllabusRepository) getSectionsWithTopics(ctx context.Context, syllabusID int) ([]models.SyllabusSection, error) {
	sRows, err := r.pool.Query(ctx, `
		SELECT id, syllabus_id, title, section_order, created_at
		FROM syllabus_sections
		WHERE syllabus_id = $1
		ORDER BY section_order, id`, syllabusID)
	if err != nil {
		return nil, err
	}
	defer sRows.Close()

	var sections []models.SyllabusSection
	for sRows.Next() {
		var sec models.SyllabusSection
		if err := sRows.Scan(&sec.ID, &sec.SyllabusID, &sec.Title, &sec.SectionOrder, &sec.CreatedAt); err != nil {
			return nil, err
		}
		sections = append(sections, sec)
	}
	if err := sRows.Err(); err != nil {
		return nil, err
	}

	for i, sec := range sections {
		topics, err := r.getTopicsWithTests(ctx, syllabusID, sec.ID)
		if err != nil {
			return nil, err
		}
		sections[i].Topics = topics
	}
	return sections, nil
}

func (r *SyllabusRepository) getTopicsWithTests(ctx context.Context, syllabusID, sectionID int) ([]models.SyllabusTopic, error) {
	tRows, err := r.pool.Query(ctx, `
		SELECT id, syllabus_id, section_id, title,
		       COALESCE(description, ''), estimated_hours, topic_order,
		       COALESCE(notes_content, ''), created_at
		FROM syllabus_topics
		WHERE syllabus_id = $1 AND section_id = $2
		ORDER BY topic_order, id`, syllabusID, sectionID)
	if err != nil {
		return nil, err
	}
	defer tRows.Close()

	var topics []models.SyllabusTopic
	for tRows.Next() {
		var t models.SyllabusTopic
		if err := tRows.Scan(
			&t.ID, &t.SyllabusID, &t.SectionID, &t.Title,
			&t.Description, &t.EstimatedHours, &t.TopicOrder,
			&t.NotesContent, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		tests, err := r.getLinkedTests(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		t.Tests = tests
		topics = append(topics, t)
	}
	return topics, tRows.Err()
}

func (r *SyllabusRepository) getLinkedTests(ctx context.Context, topicID int) ([]models.Test, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.id, t.title, COALESCE(t.description,''), t.exam_standard, t.difficulty, t.published
		FROM tests t
		JOIN syllabus_topic_tests stt ON stt.test_id = t.id
		WHERE stt.syllabus_topic_id = $1
		ORDER BY t.title`, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []models.Test
	for rows.Next() {
		var t models.Test
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.ExamStandard, &t.Difficulty, &t.Published); err != nil {
			return nil, err
		}
		tests = append(tests, t)
	}
	return tests, rows.Err()
}

// GetTopicByID returns a single SyllabusTopic with linked tests.
func (r *SyllabusRepository) GetTopicByID(ctx context.Context, topicID int) (*models.SyllabusTopic, error) {
	var t models.SyllabusTopic
	err := r.pool.QueryRow(ctx, `
		SELECT id, syllabus_id, section_id, title,
		       COALESCE(description,''), estimated_hours, topic_order,
		       COALESCE(notes_content,''), created_at
		FROM syllabus_topics WHERE id = $1`, topicID).Scan(
		&t.ID, &t.SyllabusID, &t.SectionID, &t.Title,
		&t.Description, &t.EstimatedHours, &t.TopicOrder,
		&t.NotesContent, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	tests, err := r.getLinkedTests(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	t.Tests = tests
	return &t, nil
}

// Create inserts a new syllabus record.
func (r *SyllabusRepository) Create(ctx context.Context, sy *models.Syllabus) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO syllabi (subject_id, exam_standard, title, description, is_published, created_by)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, created_at, updated_at`,
		sy.SubjectID, sy.ExamStandard, sy.Title, sy.Description, sy.IsPublished, sy.CreatedBy,
	).Scan(&sy.ID, &sy.CreatedAt, &sy.UpdatedAt)
}

// Update saves changes to a syllabus's metadata fields.
func (r *SyllabusRepository) Update(ctx context.Context, sy *models.Syllabus) error {
	return r.pool.QueryRow(ctx, `
		UPDATE syllabi
		SET subject_id=$1, exam_standard=$2, title=$3, description=$4,
		    updated_at=CURRENT_TIMESTAMP
		WHERE id=$5
		RETURNING updated_at`,
		sy.SubjectID, sy.ExamStandard, sy.Title, sy.Description, sy.ID,
	).Scan(&sy.UpdatedAt)
}

// Delete removes a syllabus (cascade deletes sections, topics, and plan sessions).
func (r *SyllabusRepository) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM syllabi WHERE id=$1`, id)
	return err
}

// Publish marks a syllabus as visible to students.
func (r *SyllabusRepository) Publish(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE syllabi SET is_published=TRUE, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id)
	return err
}

// Unpublish hides a syllabus from students.
func (r *SyllabusRepository) Unpublish(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE syllabi SET is_published=FALSE, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id)
	return err
}

// AddSection inserts a new section into a syllabus.
func (r *SyllabusRepository) AddSection(ctx context.Context, sec *models.SyllabusSection) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO syllabus_sections (syllabus_id, title, section_order)
		VALUES ($1,$2,$3)
		RETURNING id, created_at`,
		sec.SyllabusID, sec.Title, sec.SectionOrder,
	).Scan(&sec.ID, &sec.CreatedAt)
}

// UpdateSection saves changes to a section's title or order.
func (r *SyllabusRepository) UpdateSection(ctx context.Context, sec *models.SyllabusSection) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE syllabus_sections SET title=$1, section_order=$2 WHERE id=$3`,
		sec.Title, sec.SectionOrder, sec.ID)
	return err
}

// DeleteSection removes a section. Topics in this section have section_id set to NULL.
func (r *SyllabusRepository) DeleteSection(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM syllabus_sections WHERE id=$1`, id)
	return err
}

// AddTopic inserts a new topic into a syllabus section.
func (r *SyllabusRepository) AddTopic(ctx context.Context, t *models.SyllabusTopic) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO syllabus_topics
		    (syllabus_id, section_id, title, description, estimated_hours, topic_order, notes_content)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, created_at`,
		t.SyllabusID, t.SectionID, t.Title, t.Description,
		t.EstimatedHours, t.TopicOrder, t.NotesContent,
	).Scan(&t.ID, &t.CreatedAt)
}

// UpdateTopic saves changes to a topic.
func (r *SyllabusRepository) UpdateTopic(ctx context.Context, t *models.SyllabusTopic) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE syllabus_topics
		SET section_id=$1, title=$2, description=$3,
		    estimated_hours=$4, topic_order=$5, notes_content=$6
		WHERE id=$7`,
		t.SectionID, t.Title, t.Description,
		t.EstimatedHours, t.TopicOrder, t.NotesContent, t.ID)
	return err
}

// DeleteTopic removes a topic and its test links.
func (r *SyllabusRepository) DeleteTopic(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM syllabus_topics WHERE id=$1`, id)
	return err
}

// LinkTest creates a link between a syllabus topic and an existing test.
func (r *SyllabusRepository) LinkTest(ctx context.Context, topicID, testID int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO syllabus_topic_tests (syllabus_topic_id, test_id)
		VALUES ($1,$2)
		ON CONFLICT DO NOTHING`, topicID, testID)
	return err
}

// UnlinkTest removes a link between a syllabus topic and a test.
func (r *SyllabusRepository) UnlinkTest(ctx context.Context, topicID, testID int) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM syllabus_topic_tests WHERE syllabus_topic_id=$1 AND test_id=$2`,
		topicID, testID)
	return err
}

// GetAllTopicsOrdered returns all topics for a syllabus in curriculum order (for scheduling).
func (r *SyllabusRepository) GetAllTopicsOrdered(ctx context.Context, syllabusID int) ([]models.SyllabusTopic, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT st.id, st.syllabus_id, st.section_id, st.title,
		       COALESCE(st.description,''), st.estimated_hours, st.topic_order,
		       COALESCE(st.notes_content,''), st.created_at
		FROM syllabus_topics st
		LEFT JOIN syllabus_sections ss ON ss.id = st.section_id
		WHERE st.syllabus_id = $1
		ORDER BY COALESCE(ss.section_order, 9999), ss.id, st.topic_order, st.id`, syllabusID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []models.SyllabusTopic
	for rows.Next() {
		var t models.SyllabusTopic
		if err := rows.Scan(
			&t.ID, &t.SyllabusID, &t.SectionID, &t.Title,
			&t.Description, &t.EstimatedHours, &t.TopicOrder,
			&t.NotesContent, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		topics = append(topics, t)
	}
	return topics, rows.Err()
}

// SearchTests returns published tests filtered by optional subjectID for the link-test panel.
func (r *SyllabusRepository) SearchTests(ctx context.Context, subjectID *int, titleFilter string) ([]models.Test, error) {
	query := `
		SELECT t.id, t.title, COALESCE(t.description,''), t.exam_standard, t.difficulty, t.published,
		       s.id, s.name
		FROM tests t
		LEFT JOIN subjects s ON s.id = t.subject_id
		WHERE t.published = TRUE`
	args := []any{}
	argN := 1

	if subjectID != nil {
		query += fmt.Sprintf(` AND t.subject_id = $%d`, argN)
		args = append(args, *subjectID)
		argN++
	}
	if titleFilter != "" {
		query += fmt.Sprintf(` AND t.title ILIKE $%d`, argN)
		args = append(args, "%"+titleFilter+"%")
	}
	query += ` ORDER BY t.title LIMIT 50`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []models.Test
	for rows.Next() {
		var t models.Test
		var subID *int
		var subName *string
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.ExamStandard, &t.Difficulty, &t.Published,
			&subID, &subName,
		); err != nil {
			return nil, err
		}
		if subID != nil {
			t.Subject = &models.Subject{ID: *subID, Name: *subName}
		}
		tests = append(tests, t)
	}
	return tests, rows.Err()
}
