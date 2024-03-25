package testgrp

import (
	"context"
	"go.uber.org/zap"
	"net/http"
)

type Handlers struct {
	Build string
	Log   *zap.SugaredLogger
}

func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (h Handlers) TestAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}
