package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleOTK       Role = "OTK"
	RoleSticker   Role = "STICKER"
	RoleAds       Role = "ADS"
	RoleAssembler Role = "ASSEMBLER"
)

type ChecklistTemplate struct {
	ID        uuid.UUID
	Role      Role
	Version   int
	IsActive  bool
	CreatedAt time.Time
}

type Question struct {
	ID              uuid.UUID
	TemplateID      uuid.UUID
	Text            string
	Order           int
	MinPhotos       int
	MaxPhotos       int
	IsRequired      bool
	ReferenceImages []string
	CreatedAt       time.Time
}

type InspectionStatus string

const (
	StatusInProgress InspectionStatus = "in_progress"
	StatusCompleted  InspectionStatus = "completed"
)

type Inspection struct {
	ID            uuid.UUID
	TemplateID    uuid.UUID
	MachineSerial string
	InspectorName string
	Status        InspectionStatus
	StartedAt     time.Time
	FinishedAt    *time.Time
}

type InspectionAnswer struct {
	ID           uuid.UUID
	InspectionID uuid.UUID
	QuestionID   uuid.UUID
	Comment      string
	Photos       []string
	CreatedAt    time.Time
}

type InspectionDetail struct {
	Inspection Inspection
	Answers    []InspectionAnswerDetail
}

type InspectionAnswerDetail struct {
	Question Question
	Answer   InspectionAnswer
}

type ChecklistRepository interface {
	CreateTemplate(ctx context.Context, template *ChecklistTemplate) error
	CreateQuestion(ctx context.Context, question *Question) error
	ListTemplates(ctx context.Context) ([]ChecklistTemplate, error)
	GetTemplateByRole(ctx context.Context, role Role) (*ChecklistTemplate, error)
	GetTemplateByID(ctx context.Context, id uuid.UUID) (*ChecklistTemplate, error)
	GetQuestionsByTemplateID(ctx context.Context, templateID uuid.UUID) ([]Question, error)
	DeleteTemplateByRole(ctx context.Context, role Role) error
	DeactivateTemplatesByRole(ctx context.Context, role Role) error
	CreateInspection(ctx context.Context, inspection *Inspection) error
	GetInspectionByID(ctx context.Context, id uuid.UUID) (*Inspection, error)
	ListInspections(ctx context.Context, role *Role, status *InspectionStatus) ([]Inspection, error)
	GetInspectionAnswers(ctx context.Context, inspectionID uuid.UUID) ([]InspectionAnswer, error)
	SaveAnswer(ctx context.Context, answer *InspectionAnswer) error
	CompleteInspection(ctx context.Context, id uuid.UUID) error
}

type FileStorage interface {
	Upload(ctx context.Context, bucket, key string, data []byte) (string, error)
	GetURL(ctx context.Context, bucket, key string) (string, error)
}
