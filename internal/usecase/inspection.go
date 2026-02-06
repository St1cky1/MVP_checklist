package usecase

import (
	"context"
	"fmt"
	"time"

	"MVP_checklist/internal/domain"
	"github.com/google/uuid"
)

type InspectionUseCase struct {
	repo    domain.ChecklistRepository
	storage domain.FileStorage
}

func NewInspectionUseCase(repo domain.ChecklistRepository, storage domain.FileStorage) *InspectionUseCase {
	return &InspectionUseCase{repo: repo, storage: storage}
}

func (u *InspectionUseCase) StartInspection(ctx context.Context, role domain.Role, machineSerial, inspectorName string) (*domain.Inspection, []domain.Question, error) {
	template, err := u.repo.GetTemplateByRole(ctx, role)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get template for role %s: %w", role, err)
	}

	inspection := &domain.Inspection{
		ID:            uuid.New(),
		TemplateID:    template.ID,
		MachineSerial: machineSerial,
		InspectorName: inspectorName,
		Status:        domain.StatusInProgress,
		StartedAt:     time.Now(),
	}

	if err := u.repo.CreateInspection(ctx, inspection); err != nil {
		return nil, nil, fmt.Errorf("failed to create inspection: %w", err)
	}

	questions, err := u.repo.GetQuestionsByTemplateID(ctx, template.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get questions: %w", err)
	}

	return inspection, questions, nil
}

func (u *InspectionUseCase) SaveAnswer(ctx context.Context, inspectionID, questionID uuid.UUID, comment string, photos [][]byte) error {
	inspection, err := u.repo.GetInspectionByID(ctx, inspectionID)
	if err != nil {
		return err
	}
	if inspection.Status == domain.StatusCompleted {
		return fmt.Errorf("inspection already completed")
	}

	var photoKeys []string
	for i, data := range photos {
		key := fmt.Sprintf("inspections/%s/%s/%d.jpg", inspectionID, questionID, i)
		uploadedKey, err := u.storage.Upload(ctx, "", key, data)
		if err != nil {
			return fmt.Errorf("failed to upload photo %d: %w", i, err)
		}
		photoKeys = append(photoKeys, uploadedKey)
	}

	answer := &domain.InspectionAnswer{
		ID:           uuid.New(), // In a real scenario, we might want to find existing answer to update
		InspectionID: inspectionID,
		QuestionID:   questionID,
		Comment:      comment,
		Photos:       photoKeys,
		CreatedAt:    time.Now(),
	}

	return u.repo.SaveAnswer(ctx, answer)
}

func (u *InspectionUseCase) GetInspectionByID(ctx context.Context, id uuid.UUID) (*domain.Inspection, error) {
	return u.repo.GetInspectionByID(ctx, id)
}

func (u *InspectionUseCase) GetInspectionWithRole(ctx context.Context, id uuid.UUID) (*domain.Inspection, domain.Role, error) {
	inspection, err := u.repo.GetInspectionByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	template, err := u.repo.GetTemplateByID(ctx, inspection.TemplateID)
	if err != nil {
		return inspection, "", nil // return inspection even if template not found
	}
	return inspection, template.Role, nil
}

func (u *InspectionUseCase) GetQuestionsByTemplateID(ctx context.Context, templateID uuid.UUID) (*domain.ChecklistTemplate, []domain.Question, error) {
	// For simplicity, we just return questions
	questions, err := u.repo.GetQuestionsByTemplateID(ctx, templateID)
	return nil, questions, err
}

func (u *InspectionUseCase) CompleteInspection(ctx context.Context, inspectionID uuid.UUID) error {
	return u.repo.CompleteInspection(ctx, inspectionID)
}
