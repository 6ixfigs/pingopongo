package rest

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
)

type Endpoint struct {
	Router  chi.Router
	handler *Handler
}

func NewEndpoint(db *sql.DB) *Endpoint {
	return &Endpoint{Router: chi.NewRouter(), handler: newHandler(db)}
}

func (e *Endpoint) MountHandlers() {
	e.Router.Post("/commands", e.handler.parse)
}
