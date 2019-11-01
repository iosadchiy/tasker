package tasker

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type Handlers struct {
	MaxWaitTime time.Duration
	Service     Service
}

func (h *Handlers) HandleCreate(w http.ResponseWriter, r *http.Request) {
	task, err := h.Service.Create(r.Context())
	if err != nil {
		encodeError(r.Context(), http.StatusInternalServerError, err, w)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	if _, err := w.Write([]byte(task.ID.String())); err != nil {
		log.Println(err)
	}
}

func (h *Handlers) HandleRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		encodeError(ctx, http.StatusBadRequest, err, w)
		return
	}
	task, err := h.Service.Read(ctx, id)
	if err == gorm.ErrRecordNotFound {
		encodeError(ctx, http.StatusNotFound, err, w)
		return
	}
	if err != nil {
		encodeError(ctx, http.StatusInternalServerError, err, w)
		return
	}

	encodeTask(ctx, w, task)
}

func (h *Handlers) HandlePoll(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.MaxWaitTime)
	defer cancel()
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		encodeError(ctx, http.StatusBadRequest, err, w)
		return
	}
	task, err := h.Service.Poll(ctx, id)
	if err != nil {
		encodeError(ctx, http.StatusInternalServerError, err, w)
		return
	}

	encodeTask(ctx, w, task)
}

func encodeError(_ context.Context, code int, err error, w http.ResponseWriter) {
	w.WriteHeader(code)
	if _, err = w.Write([]byte(err.Error())); err != nil {
		log.Println(err)
	}
}

func encodeTask(ctx context.Context, w http.ResponseWriter, task Task) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(task)
	if err != nil {
		encodeError(ctx, http.StatusInternalServerError, err, w)
		return
	}
}
