// Package testgrp is for the class.
package testgrp

import (
	"context"
	"errors"
	"math/rand"
	"net/http"

	v1Web "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

// Handlers manages the set of user endpoints.
type Handlers struct {
	DB *sqlx.DB
}

// Status represents a test handler for now.
func (h Handlers) Status(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if n := rand.Intn(100); n%2 == 0 {
		return v1Web.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)
		//return errors.New("NON trusted error")
		// panic("ohh no we panic")
	}

	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}
