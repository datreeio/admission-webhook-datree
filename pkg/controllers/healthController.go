package controllers

import (
	"net/http"

	"github.com/datreeio/webhook-datree/pkg/responseWriter"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (h *HealthController) Health(w http.ResponseWriter, req *http.Request) {
	writer := responseWriter.New(w)
	writer.Write("OK")
}

func (h *HealthController) Ready(w http.ResponseWriter, req *http.Request) {
	writer := responseWriter.New(w)
	writer.Write("OK")
}
