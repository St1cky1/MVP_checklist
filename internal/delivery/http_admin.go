package delivery

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"MVP_checklist/internal/domain"
	"MVP_checklist/internal/usecase"
	"github.com/google/uuid"
)

type AdminHandler struct {
	templateUC  *usecase.TemplateUseCase
	analyticsUC *usecase.AnalyticsUseCase
}

func NewAdminHandler(templateUC *usecase.TemplateUseCase, analyticsUC *usecase.AnalyticsUseCase) *AdminHandler {
	return &AdminHandler{
		templateUC:  templateUC,
		analyticsUC: analyticsUC,
	}
}

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/admin/templates" && r.Method == http.MethodGet:
		h.handleListTemplates(w, r)
	case path == "/admin/templates" && r.Method == http.MethodPost:
		h.handleCreateTemplate(w, r)
	case path == "/admin/inspections" && r.Method == http.MethodGet:
		h.handleListInspections(w, r)
	case strings.HasPrefix(path, "/admin/inspections/") && strings.HasSuffix(path, "/export/csv") && r.Method == http.MethodGet:
		h.handleExportCSV(w, r)
	case strings.HasPrefix(path, "/admin/inspections/") && strings.HasSuffix(path, "/export/pdf") && r.Method == http.MethodGet:
		h.handleExportPDF(w, r)
	case strings.HasPrefix(path, "/admin/inspections/") && r.Method == http.MethodGet:
		h.handleGetInspectionDetail(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *AdminHandler) render(w http.ResponseWriter, name string, data interface{}) {
	tmpl, err := template.ParseFiles("templates/layout.html", "templates/admin/"+name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderData := map[string]interface{}{
		"IsAdmin": true,
		"Data":    data,
	}
	err = tmpl.ExecuteTemplate(w, "layout", renderData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *AdminHandler) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id, _ := uuid.Parse(parts[3])

	data, err := h.analyticsUC.ExportToCSV(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=inspection_%s.csv", id))
	w.Write(data)
}

func (h *AdminHandler) handleExportPDF(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id, _ := uuid.Parse(parts[3])

	data, err := h.analyticsUC.ExportToPDF(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=inspection_%s.pdf", id))
	w.Write(data)
}

func (h *AdminHandler) handleListInspections(w http.ResponseWriter, r *http.Request) {
	inspections, err := h.analyticsUC.ListInspections(r.Context(), nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.render(w, "inspections.html", inspections)
}

func (h *AdminHandler) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.templateUC.ListTemplates(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.render(w, "templates.html", templates)
}

func (h *AdminHandler) handleGetInspectionDetail(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id, err := uuid.Parse(parts[3])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	detail, err := h.analyticsUC.GetInspectionDetail(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.render(w, "detail.html", detail)
}

type createTemplateRequest struct {
	Role      domain.Role       `json:"role"`
	Questions []domain.Question `json:"questions"`
}

func (h *AdminHandler) handleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req createTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	template, err := h.templateUC.CreateTemplate(r.Context(), req.Role, req.Questions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(template)
}
