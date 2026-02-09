package usecase

import (
	"context"

	"MVP_checklist/internal/domain"
	"github.com/google/uuid"
)

type TemplateUseCase struct {
	repo domain.ChecklistRepository
}

func NewTemplateUseCase(repo domain.ChecklistRepository) *TemplateUseCase {
	return &TemplateUseCase{repo: repo}
}

func (u *TemplateUseCase) CreateTemplate(ctx context.Context, role domain.Role, questions []domain.Question) (*domain.ChecklistTemplate, error) {
	// Find current version to increment
	version := 1
	oldTmpl, err := u.repo.GetTemplateByRole(ctx, role)
	if err == nil && oldTmpl != nil {
		version = oldTmpl.Version + 1
		// Deactivate old versions
		_ = u.repo.DeactivateTemplatesByRole(ctx, role)
	}

	template := &domain.ChecklistTemplate{
		ID:       uuid.New(),
		Role:     role,
		Version:  version,
		IsActive: true,
	}

	if err := u.repo.CreateTemplate(ctx, template); err != nil {
		return nil, err
	}

	for i := range questions {
		questions[i].ID = uuid.New()
		questions[i].TemplateID = template.ID
		if err := u.repo.CreateQuestion(ctx, &questions[i]); err != nil {
			return nil, err
		}
	}

	return template, nil
}

func (u *TemplateUseCase) ListTemplates(ctx context.Context) ([]domain.ChecklistTemplate, error) {
	return u.repo.ListTemplates(ctx)
}

func (u *TemplateUseCase) GetTemplateByRole(ctx context.Context, role domain.Role) (*domain.ChecklistTemplate, []domain.Question, error) {
	template, err := u.repo.GetTemplateByRole(ctx, role)
	if err != nil {
		return nil, nil, err
	}

	questions, err := u.repo.GetQuestionsByTemplateID(ctx, template.ID)
	if err != nil {
		return nil, nil, err
	}

	return template, questions, nil
}

func (u *TemplateUseCase) DeleteTemplateByRole(ctx context.Context, role domain.Role) error {
	return u.repo.DeleteTemplateByRole(ctx, role)
}
