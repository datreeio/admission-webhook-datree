package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	responseRecorder := httptest.NewRecorder()

	healthController := NewHealthController()
	healthController.Health(responseRecorder, request)

	assert.Equal(t, responseRecorder.Code, http.StatusOK)
	assert.Equal(t, strings.TrimSpace(responseRecorder.Body.String()), "OK")
}

func TestReady(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)
	responseRecorder := httptest.NewRecorder()

	healthController := NewHealthController()
	healthController.Ready(responseRecorder, request)

	assert.Equal(t, responseRecorder.Code, http.StatusOK)
	assert.Equal(t, strings.TrimSpace(responseRecorder.Body.String()), "OK")
}
