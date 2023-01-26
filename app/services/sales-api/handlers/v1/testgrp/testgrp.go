package testgrp

import (
	"context"
	"errors"
	"math/rand"
	"net/http"

	"github.com/ardanlabs/service/foundation/web"
)

// Status represents a test handler for now.
func Status(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if n := rand.Intn(100); n%2 == 0 {
		//return v1Web.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)
		return errors.New("NON trusted error")
	}

	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}
