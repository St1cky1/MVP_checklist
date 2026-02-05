package delivery

import (
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	"MVP_checklist/internal/domain"
	"MVP_checklist/internal/usecase"
	"github.com/google/uuid"
)

type PublicHandler struct {
	inspectionUC *usecase.InspectionUseCase
}

func NewPublicHandler(inspectionUC *usecase.InspectionUseCase) *PublicHandler {
	return &PublicHandler{
		inspectionUC: inspectionUC,
	}
}

func (h *PublicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/" && r.Method == http.MethodGet:
		h.render(w, "index.html", nil)
	case path == "/inspections/start" && r.Method == http.MethodPost:
		h.handleStartInspection(w, r)
	case strings.HasPrefix(path, "/inspections/") && strings.HasSuffix(path, "/question") && r.Method == http.MethodGet:
		h.handleShowQuestion(w, r)
	case strings.HasPrefix(path, "/inspections/") && strings.HasSuffix(path, "/answer") && r.Method == http.MethodPost:
		h.handleSaveAnswer(w, r)
	case strings.HasPrefix(path, "/inspections/") && strings.HasSuffix(path, "/success") && r.Method == http.MethodGet:
		h.handleShowSuccess(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *PublicHandler) render(w http.ResponseWriter, name string, data interface{}) {
	tmpl, err := template.ParseFiles("templates/layout.html", "templates/public/"+name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderData := map[string]interface{}{
		"IsAdmin": false,
		"Data":    data,
	}
	err = tmpl.ExecuteTemplate(w, "layout", renderData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *PublicHandler) handleStartInspection(w http.ResponseWriter, r *http.Request) {
	role := domain.Role(r.FormValue("role"))
	machineSerial := r.FormValue("machine_serial")
	inspectorName := r.FormValue("inspector_name")

	inspection, _, err := h.inspectionUC.StartInspection(r.Context(), role, machineSerial, inspectorName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/inspections/"+inspection.ID.String()+"/question?step=1", http.StatusSeeOther)
}

func (h *PublicHandler) handleShowQuestion(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	inspectionID, _ := uuid.Parse(parts[2])
	step, _ := strconv.Atoi(r.URL.Query().Get("step"))

	inspection, err := h.inspectionUC.GetInspectionByID(r.Context(), inspectionID)
	if err != nil {
		http.Error(w, "Inspection not found", http.StatusNotFound)
		return
	}

	// For MVP, we get all questions and pick the one for the current step
	_, questions, err := h.inspectionUC.GetQuestionsByTemplateID(r.Context(), inspection.TemplateID)
	if err != nil || step > len(questions) || step < 1 {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	h.render(w, "question.html", map[string]interface{}{
		"InspectionID":  inspectionID,
		"MachineSerial": inspection.MachineSerial,
		"Question":      questions[step-1],
		"CurrentStep":   step,
		"TotalSteps":    len(questions),
		"Progress":      (step * 100) / len(questions),
		"Role":          inspection.TemplateID, // simplified
	})
}

func (h *PublicHandler) handleSaveAnswer(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	inspectionID, _ := uuid.Parse(parts[2])

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	questionID, _ := uuid.Parse(r.FormValue("question_id"))
	step, _ := strconv.Atoi(r.FormValue("step"))
	comment := r.FormValue("comment")

	var photos [][]byte
	files := r.MultipartForm.File["photos"]
	for _, fh := range files {
		f, _ := fh.Open()
		data, _ := io.ReadAll(f)
		photos = append(photos, data)
		f.Close()
	}

	err = h.inspectionUC.SaveAnswer(r.Context(), inspectionID, questionID, comment, photos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if more questions
	inspection, _ := h.inspectionUC.GetInspectionByID(r.Context(), inspectionID)
	_, questions, _ := h.inspectionUC.GetQuestionsByTemplateID(r.Context(), inspection.TemplateID)

	if step >= len(questions) {
		h.inspectionUC.CompleteInspection(r.Context(), inspectionID)
		http.Redirect(w, r, "/inspections/"+inspectionID.String()+"/success", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/inspections/"+inspectionID.String()+"/question?step="+strconv.Itoa(step+1), http.StatusSeeOther)
	}
}

func (h *PublicHandler) handleShowSuccess(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	inspectionID, _ := uuid.Parse(parts[2])

	inspection, _ := h.inspectionUC.GetInspectionByID(r.Context(), inspectionID)
	h.render(w, "success.html", inspection)
}
