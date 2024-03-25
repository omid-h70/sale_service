package checkgrp

import (
	"context"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"os"
	"service/domain/sys/database"
	"service/tooling"
	"time"
)

type Handlers struct {
	Build string
	Log   *zap.SugaredLogger
	DB    *sqlx.DB
}

func (h Handlers) Readiness(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	data := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	var statusCode = http.StatusOK
	if err := database.StatusCheck(ctx, h.DB); err != nil {
		data.Status = "not ready"
		statusCode = http.StatusInternalServerError
	}

	err := response(w, statusCode, data)
	if err != nil {
		h.Log.Errorw("readiness", "error", err)
	}
	h.Log.Infow("readiness", "statuscode", statusCode)
}

func (h Handlers) Liveness(w http.ResponseWriter, r *http.Request) {

	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	data := struct {
		Status    string `json:"status,omitempty"`
		Build     string `json:"build,omitempty"`
		Host      string `json:"host,omitempty"`
		Pod       string `json:"pod,omitempty"`
		PodIP     string `json:"pod_ip,omitempty"`
		Node      string `json:"node,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	}{
		Status:    "up",
		Build:     h.Build,
		Host:      host,
		Pod:       os.Getenv("KUBERNETES_POD_NAME"),
		PodIP:     os.Getenv("KUBERNETES_NAME_SPACE_POD_IP"),
		Node:      os.Getenv("KUBERNETES_NODENAME"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}

	statusCode := http.StatusOK
	err = response(w, statusCode, data)
	if err != nil {
		h.Log.Errorw("readiness", "error", err)
	}

	h.Log.Infow(tooling.GetCallerName(), "status_code", statusCode)
}

func response(w http.ResponseWriter, statusCode int, data any) error {
	return nil
}
