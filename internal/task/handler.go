package task

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jzxq45/golang-todoapp/internal/httpx"
)

const maxJSONBodyBytes = 1 << 20

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/tasks", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateTaskInput
	if err := decodeJSONBody(w, r, &input); err != nil {
		httpx.Error(w, *err)
		return
	}

	task, err := h.service.Create(r.Context(), input)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httpx.Created(w, task)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.Error(w, httpx.ErrInvalidID)
		return
	}

	task, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httpx.OK(w, task)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	filter := Filter{}
	if doneStr := r.URL.Query().Get("done"); doneStr != "" {
		done, err := strconv.ParseBool(doneStr)
		if err != nil {
			httpx.Error(w, httpx.ErrInvalidQueryParam("done", "must be true or false"))
			return
		}
		filter.Done = &done
	}

	tasks, err := h.service.List(r.Context(), filter)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httpx.OK(w, tasks)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.Error(w, httpx.ErrInvalidID)
		return
	}

	var input UpdateTaskInput
	if err := decodeJSONBody(w, r, &input); err != nil {
		httpx.Error(w, *err)
		return
	}

	task, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		h.handleError(w, err)
		return
	}

	httpx.OK(w, task)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.Error(w, httpx.ErrInvalidID)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}

	httpx.NoContent(w)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, httpx.ErrNotFound)
	case errors.Is(err, ErrTitleRequired):
		httpx.Error(w, httpx.ErrValidation("title", "title is required"))
	case errors.Is(err, ErrTitleTooLong):
		httpx.Error(w, httpx.ErrValidation("title", "title must be at most 200 characters"))
	default:
		httpx.Error(w, httpx.ErrInternal)
	}
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, destination any) *httpx.ErrorDetail {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBodyBytes))
	if err := decoder.Decode(destination); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return &httpx.ErrPayloadTooLarge
		}
		return &httpx.ErrInvalidJSON
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return &httpx.ErrPayloadTooLarge
		}
		return &httpx.ErrInvalidJSON
	}
	return nil
}
