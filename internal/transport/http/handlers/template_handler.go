package handlers

import (
	"net/http"

	taskdomain "example.com/taskservice/internal/domain/task"
	templateusecase "example.com/taskservice/internal/usecase/task_template"
)

type TemplateHandler struct {
	usecase templateusecase.Usecase
}

func NewTemplateHandler(usecase templateusecase.Usecase) *TemplateHandler {
	return &TemplateHandler{usecase: usecase}
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req templateMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	createdTemplate, createdTasks, err := h.usecase.Create(r.Context(), templateusecase.TemplateInput{
		Title:       req.Title,
		Description: req.Description,
		Recurrence:  req.Recurrence.toDomain(),
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, templateCreateResponseDTO{
		Template: newTemplateDTO(createdTemplate),
		Tasks:    newTasksDTO(createdTasks),
	})
}

func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req templateMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updatedTemplate, createdTasks, err := h.usecase.Update(r.Context(), id, templateusecase.TemplateInput{
		Title:       req.Title,
		Description: req.Description,
		Recurrence:  req.Recurrence.toDomain(),
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, templateUpdateResponseDTO{
		Template: newTemplateDTO(updatedTemplate),
		Tasks:    newTasksDTO(createdTasks),
	})
}

func (h *TemplateHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	template, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTemplateDTO(template))
}

func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	templates, err := h.usecase.List(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]templateDTO, 0, len(templates))
	for i := range templates {
		response = append(response, newTemplateDTO(&templates[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (dto recurrenceDTO) toDomain() taskdomain.Recurrence {
	return taskdomain.Recurrence{
		Type:          dto.Type,
		EveryNDays:    dto.EveryNDays,
		DayOfMonth:    dto.DayOfMonth,
		MonthParity:   dto.MonthParity,
		SpecificDates: dto.SpecificDates,
	}
}
