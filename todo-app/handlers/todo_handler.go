package handlers

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"todo-app/repository"
)

const (
	maxTitleLength = 200
	maxBodyBytes   = 4 << 10 // 4 KB is plenty for one title
	requestTimeout = 5 * time.Second
)

type TodoHandler struct {
	repo *repository.TodoRepository
	tmpl *template.Template
}

func NewTodoHandler(repo *repository.TodoRepository, tmpl *template.Template) *TodoHandler {
	return &TodoHandler{repo: repo, tmpl: tmpl}
}

// RegisterRoutes uses Go 1.22+ patterns: "METHOD /path/{wildcard}".
func (h *TodoHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.Index) // {$} = exactly "/"
	mux.HandleFunc("GET /todos", h.List)
	mux.HandleFunc("POST /todos", h.Create)
	mux.HandleFunc("GET /todos/{id}/edit", h.EditForm)
	mux.HandleFunc("PUT /todos/{id}", h.Update)
	mux.HandleFunc("PATCH /todos/{id}/toggle", h.Toggle)
	mux.HandleFunc("DELETE /todos/{id}", h.Delete)
}

// Index renders the full page.
func (h *TodoHandler) Index(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	todos, err := h.repo.GetTodos(ctx)
	if err != nil {
		log.Printf("index: %v", err)
		http.Error(w, "Something went wrong. Please try again.", http.StatusInternalServerError)
		return
	}
	h.render(w, http.StatusOK, "index", todos)
}

// List renders only the list fragment.
func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	todos, err := h.repo.GetTodos(ctx)
	if err != nil {
		h.handleRepoError(w, "list", err)
		return
	}
	h.render(w, http.StatusOK, "todo-list", todos)
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	title, ok := h.readTitle(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	todo, err := h.repo.CreateTodo(ctx, title)
	if err != nil {
		h.handleRepoError(w, "create", err)
		return
	}
	h.render(w, http.StatusCreated, "todo-response", todo)
}

func (h *TodoHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	todo, err := h.repo.GetTodoByID(ctx, id)
	if err != nil {
		h.handleRepoError(w, "edit form", err)
		return
	}
	h.render(w, http.StatusOK, "edit-form", todo)
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	title, ok := h.readTitle(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	todo, err := h.repo.UpdateTodo(ctx, id, title)
	if err != nil {
		h.handleRepoError(w, "update", err)
		return
	}
	h.render(w, http.StatusOK, "todo-response", todo)
}

func (h *TodoHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	todo, err := h.repo.ToggleTodo(ctx, id)
	if err != nil {
		h.handleRepoError(w, "toggle", err)
		return
	}
	h.render(w, http.StatusOK, "todo-response", todo)
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	if err := h.repo.DeleteTodo(ctx, id); err != nil {
		h.handleRepoError(w, "delete", err)
		return
	}
	// Status 200 with an (almost) empty body: swapping nothing over the <li>
	// with outerHTML removes it from the page.
	h.render(w, http.StatusOK, "clear-message", nil)
}

// ---- helpers ----

func (h *TodoHandler) parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.renderError(w, http.StatusBadRequest, "Invalid todo ID.")
		return 0, false
	}
	return id, true
}

// readTitle limits the body size, parses the form and validates the title.
func (h *TodoHandler) readTitle(w http.ResponseWriter, r *http.Request) (string, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := r.ParseForm(); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			h.renderError(w, http.StatusRequestEntityTooLarge, "Request is too large.")
		} else {
			h.renderError(w, http.StatusBadRequest, "Could not read the form.")
		}
		return "", false
	}

	title := strings.TrimSpace(r.PostFormValue("title"))
	if title == "" {
		h.renderError(w, http.StatusUnprocessableEntity, "Please enter a title.")
		return "", false
	}
	if utf8.RuneCountInString(title) > maxTitleLength {
		h.renderError(w, http.StatusUnprocessableEntity, "Title must be 200 characters or fewer.")
		return "", false
	}
	return title, true
}

// handleRepoError logs the real error but shows the browser a safe message.
func (h *TodoHandler) handleRepoError(w http.ResponseWriter, op string, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		h.renderError(w, http.StatusNotFound, "That todo no longer exists.")
		return
	}
	log.Printf("%s: %v", op, err)
	h.renderError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
}

// renderError sends the message to the #message div via HTMX response headers.
func (h *TodoHandler) renderError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("HX-Retarget", "#message")
	w.Header().Set("HX-Reswap", "innerHTML")
	h.render(w, status, "error", msg)
}

// render executes into a buffer first, so a template failure never sends half a page.
func (h *TodoHandler) render(w http.ResponseWriter, status int, name string, data any) {
	var buf bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("render %q: %v", name, err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("write response: %v", err)
	}
}
