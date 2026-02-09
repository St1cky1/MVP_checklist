package repository

import (
	"context"
	"fmt"
	"time"

	"MVP_checklist/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateTemplate(ctx context.Context, t *domain.ChecklistTemplate) error {
	query := `INSERT INTO checklist_templates (id, role, version, is_active) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, t.ID, string(t.Role), t.Version, t.IsActive)
	return err
}

func (r *PostgresRepository) CreateQuestion(ctx context.Context, q *domain.Question) error {
	query := `INSERT INTO questions (id, template_id, text, "order", min_photos, max_photos, is_required, reference_images) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query, q.ID, q.TemplateID, q.Text, q.Order, q.MinPhotos, q.MaxPhotos, q.IsRequired, q.ReferenceImages)
	return err
}

func (r *PostgresRepository) ListTemplates(ctx context.Context) ([]domain.ChecklistTemplate, error) {
	query := `SELECT id, role, version, is_active, created_at FROM checklist_templates ORDER BY role, version DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []domain.ChecklistTemplate
	for rows.Next() {
		var t domain.ChecklistTemplate
		err := rows.Scan(&t.ID, &t.Role, &t.Version, &t.IsActive, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (r *PostgresRepository) GetTemplateByRole(ctx context.Context, role domain.Role) (*domain.ChecklistTemplate, error) {
	query := `SELECT id, role, version, is_active, created_at FROM checklist_templates 
              WHERE role = $1 AND is_active = true ORDER BY version DESC LIMIT 1`

	var t domain.ChecklistTemplate
	err := r.db.QueryRow(ctx, query, string(role)).Scan(&t.ID, &t.Role, &t.Version, &t.IsActive, &t.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("template not found for role %s", role)
		}
		return nil, err
	}
	return &t, nil
}

func (r *PostgresRepository) GetTemplateByID(ctx context.Context, id uuid.UUID) (*domain.ChecklistTemplate, error) {
	query := `SELECT id, role, version, is_active, created_at FROM checklist_templates WHERE id = $1`

	var t domain.ChecklistTemplate
	err := r.db.QueryRow(ctx, query, id).Scan(&t.ID, &t.Role, &t.Version, &t.IsActive, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PostgresRepository) GetQuestionsByTemplateID(ctx context.Context, templateID uuid.UUID) ([]domain.Question, error) {
	query := `SELECT id, template_id, text, "order", min_photos, max_photos, is_required, reference_images, created_at 
              FROM questions WHERE template_id = $1 ORDER BY "order" ASC`

	rows, err := r.db.Query(ctx, query, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []domain.Question
	for rows.Next() {
		var q domain.Question
		err := rows.Scan(&q.ID, &q.TemplateID, &q.Text, &q.Order, &q.MinPhotos, &q.MaxPhotos, &q.IsRequired, &q.ReferenceImages, &q.CreatedAt)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, nil
}

func (r *PostgresRepository) CreateInspection(ctx context.Context, inspection *domain.Inspection) error {
	query := `INSERT INTO inspections (id, template_id, machine_serial, inspector_name, status, started_at) 
              VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query, inspection.ID, inspection.TemplateID, inspection.MachineSerial, inspection.InspectorName, string(inspection.Status), inspection.StartedAt)
	return err
}

func (r *PostgresRepository) GetInspectionByID(ctx context.Context, id uuid.UUID) (*domain.Inspection, error) {
	query := `SELECT id, template_id, machine_serial, inspector_name, status, started_at, finished_at 
              FROM inspections WHERE id = $1`

	var i domain.Inspection
	err := r.db.QueryRow(ctx, query, id).Scan(&i.ID, &i.TemplateID, &i.MachineSerial, &i.InspectorName, &i.Status, &i.StartedAt, &i.FinishedAt)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *PostgresRepository) ListInspections(ctx context.Context, role *domain.Role, status *domain.InspectionStatus) ([]domain.Inspection, error) {
	query := `SELECT i.id, i.template_id, i.machine_serial, i.inspector_name, i.status, i.started_at, i.finished_at 
              FROM inspections i 
              JOIN checklist_templates t ON i.template_id = t.id 
              WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if role != nil {
		query += fmt.Sprintf(" AND t.role = $%d", argIdx)
		args = append(args, string(*role))
		argIdx++
	}
	if status != nil {
		query += fmt.Sprintf(" AND i.status = $%d", argIdx)
		args = append(args, string(*status))
		argIdx++
	}
	query += " ORDER BY i.started_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inspections []domain.Inspection
	for rows.Next() {
		var i domain.Inspection
		err := rows.Scan(&i.ID, &i.TemplateID, &i.MachineSerial, &i.InspectorName, &i.Status, &i.StartedAt, &i.FinishedAt)
		if err != nil {
			return nil, err
		}
		inspections = append(inspections, i)
	}
	return inspections, nil
}

func (r *PostgresRepository) GetInspectionAnswers(ctx context.Context, inspectionID uuid.UUID) ([]domain.InspectionAnswer, error) {
	query := `SELECT ia.id, ia.inspection_id, ia.question_id, ia.comment, ia.created_at, 
              array_remove(array_agg(ap.file_url), NULL) as photos
              FROM inspection_answers ia
              LEFT JOIN answer_photos ap ON ia.id = ap.answer_id
              WHERE ia.inspection_id = $1
              GROUP BY ia.id`

	rows, err := r.db.Query(ctx, query, inspectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []domain.InspectionAnswer
	for rows.Next() {
		var a domain.InspectionAnswer
		err := rows.Scan(&a.ID, &a.InspectionID, &a.QuestionID, &a.Comment, &a.CreatedAt, &a.Photos)
		if err != nil {
			return nil, err
		}
		answers = append(answers, a)
	}
	return answers, nil
}

func (r *PostgresRepository) SaveAnswer(ctx context.Context, answer *domain.InspectionAnswer) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queryAnswer := `INSERT INTO inspection_answers (id, inspection_id, question_id, comment) 
                    VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET comment = $4`

	_, err = tx.Exec(ctx, queryAnswer, answer.ID, answer.InspectionID, answer.QuestionID, answer.Comment)
	if err != nil {
		return err
	}

	// Delete old photos for this answer if we are updating (simple approach for MVP)
	_, err = tx.Exec(ctx, "DELETE FROM answer_photos WHERE answer_id = $1", answer.ID)
	if err != nil {
		return err
	}

	queryPhoto := `INSERT INTO answer_photos (answer_id, file_url) VALUES ($1, $2)`
	for _, photo := range answer.Photos {
		_, err = tx.Exec(ctx, queryPhoto, answer.ID, photo)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) CompleteInspection(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE inspections SET status = $1, finished_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, string(domain.StatusCompleted), time.Now(), id)
	return err
}

func (r *PostgresRepository) DeleteTemplateByRole(ctx context.Context, role domain.Role) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get all template IDs for this role
	rows, err := tx.Query(ctx, "SELECT id FROM checklist_templates WHERE role = $1", string(role))
	if err != nil {
		return err
	}
	defer rows.Close()

	var templateIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			templateIDs = append(templateIDs, id)
		}
	}

	for _, templateID := range templateIDs {
		// Delete related records (simple approach for seeding)
		_, _ = tx.Exec(ctx, "DELETE FROM answer_photos WHERE answer_id IN (SELECT id FROM inspection_answers WHERE question_id IN (SELECT id FROM questions WHERE template_id = $1))", templateID)
		_, _ = tx.Exec(ctx, "DELETE FROM inspection_answers WHERE question_id IN (SELECT id FROM questions WHERE template_id = $1)", templateID)
		_, _ = tx.Exec(ctx, "DELETE FROM inspections WHERE template_id = $1", templateID)
		_, _ = tx.Exec(ctx, "DELETE FROM questions WHERE template_id = $1", templateID)
		_, _ = tx.Exec(ctx, "DELETE FROM checklist_templates WHERE id = $1", templateID)
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) DeactivateTemplatesByRole(ctx context.Context, role domain.Role) error {
	query := `UPDATE checklist_templates SET is_active = false WHERE role = $1`
	_, err := r.db.Exec(ctx, query, string(role))
	return err
}
