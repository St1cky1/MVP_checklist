package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"MVP_checklist/internal/domain"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
)

type AnalyticsUseCase struct {
	repo    domain.ChecklistRepository
	storage domain.FileStorage
}

func NewAnalyticsUseCase(repo domain.ChecklistRepository, storage domain.FileStorage) *AnalyticsUseCase {
	return &AnalyticsUseCase{repo: repo, storage: storage}
}

func (u *AnalyticsUseCase) ListInspections(ctx context.Context, role *domain.Role, status *domain.InspectionStatus) ([]domain.Inspection, error) {
	return u.repo.ListInspections(ctx, role, status)
}

func (u *AnalyticsUseCase) GetInspectionDetail(ctx context.Context, inspectionID uuid.UUID) (*domain.InspectionDetail, error) {
	inspection, err := u.repo.GetInspectionByID(ctx, inspectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspection: %w", err)
	}

	questions, err := u.repo.GetQuestionsByTemplateID(ctx, inspection.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template questions: %w", err)
	}

	answers, err := u.repo.GetInspectionAnswers(ctx, inspectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get answers: %w", err)
	}

	// Map answers to questions for detail view
	answerMap := make(map[uuid.UUID]domain.InspectionAnswer)
	for _, a := range answers {
		// Replace S3 keys with presigned URLs for viewing
		for i, key := range a.Photos {
			url, err := u.storage.GetURL(ctx, "checklist-photos", key)
			if err == nil {
				a.Photos[i] = url
			}
		}
		answerMap[a.QuestionID] = a
	}

	var details []domain.InspectionAnswerDetail
	for _, q := range questions {
		details = append(details, domain.InspectionAnswerDetail{
			Question: q,
			Answer:   answerMap[q.ID],
		})
	}

	return &domain.InspectionDetail{
		Inspection: *inspection,
		Answers:    details,
	}, nil
}

func (u *AnalyticsUseCase) ExportToCSV(ctx context.Context, inspectionID uuid.UUID) ([]byte, error) {
	detail, err := u.GetInspectionDetail(ctx, inspectionID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	w.Write([]string{"Machine Serial", "Inspector", "Status", "Started At", "Finished At"})
	finishedAt := ""
	if detail.Inspection.FinishedAt != nil {
		finishedAt = detail.Inspection.FinishedAt.String()
	}
	w.Write([]string{
		detail.Inspection.MachineSerial,
		detail.Inspection.InspectorName,
		string(detail.Inspection.Status),
		detail.Inspection.StartedAt.String(),
		finishedAt,
	})

	w.Write([]string{}) // Empty line
	w.Write([]string{"Question", "Comment", "Photos"})

	for _, d := range detail.Answers {
		photos := ""
		for _, p := range d.Answer.Photos {
			photos += p + " "
		}
		w.Write([]string{
			d.Question.Text,
			d.Answer.Comment,
			photos,
		})
	}

	w.Flush()
	return buf.Bytes(), nil
}

func (u *AnalyticsUseCase) ExportToPDF(ctx context.Context, inspectionID uuid.UUID) ([]byte, error) {
	detail, err := u.GetInspectionDetail(ctx, inspectionID)
	if err != nil {
		return nil, err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Note: Standard fonts don't support Cyrillic. 
	// In a real app, we would add a Unicode font here.
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("Inspection: %s", detail.Inspection.MachineSerial))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Inspector: %s", detail.Inspection.InspectorName))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Status: %s", detail.Inspection.Status))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Started: %s", detail.Inspection.StartedAt.Format("02.01.2006 15:04")))
	pdf.Ln(15)

	for _, d := range detail.Answers {
		pdf.SetFont("Arial", "B", 12)
		pdf.MultiCell(0, 8, d.Question.Text, "", "", false)
		
		pdf.SetFont("Arial", "I", 10)
		if d.Answer.Comment != "" {
			pdf.MultiCell(0, 6, fmt.Sprintf("Comment: %s", d.Answer.Comment), "", "", false)
		}
		
		if len(d.Answer.Photos) > 0 {
			pdf.SetFont("Arial", "", 10)
			pdf.Cell(0, 6, fmt.Sprintf("Photos: %d uploaded", len(d.Answer.Photos)))
			pdf.Ln(6)
		} else {
			pdf.SetFont("Arial", "", 10)
			pdf.Cell(0, 6, "No photos uploaded")
			pdf.Ln(6)
		}
		pdf.Ln(5)
	}

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

