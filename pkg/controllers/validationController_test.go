package controllers

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/datreeio/admission-webhook-datree/pkg/config"
	"github.com/stretchr/testify/assert"
)

//go:embed test_fixtures/applyNotAllowedRequest.json
var applyRequestNotAllowedJson string

//go:embed test_fixtures/applyAllowedRequest.json
var applyRequestAllowedJson string

//go:embed test_fixtures/applyAllowedRequestFluxCD.json
var applyAllowedRequestFluxCDJson string

//go:embed test_fixtures/applyAllowedRequestFluxCDNoLabels.json
var applyAllowedRequestFluxCDJsonNoLabels string

func TestHeaderValidation(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/validate", nil)
	responseRecorder := httptest.NewRecorder()

	request.Header.Set("Content-Type", "text/html")
	validationController := NewValidationController()
	validationController.Validate(responseRecorder, request)

	assert.Equal(t, responseRecorder.Code, http.StatusBadRequest)
	assert.Equal(t, strings.TrimSpace(responseRecorder.Body.String()), "Content-Type header is not application/json")
}

func TestValidateHttpMethod(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/validate", nil)
	responseRecorder := httptest.NewRecorder()

	request.Header.Set("Content-Type", "application/json")
	validationController := NewValidationController()
	validationController.Validate(responseRecorder, request)

	assert.Equal(t, responseRecorder.Code, http.StatusMethodNotAllowed)
	assert.Equal(t, strings.TrimSpace(responseRecorder.Body.String()), "Method not allowed")
}

func TestValidateRequestBodyEmpty(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(""))
	responseRecorder := httptest.NewRecorder()

	request.Header.Set("Content-Type", "application/json")
	validationController := NewValidationController()
	validationController.Validate(responseRecorder, request)

	assert.Equal(t, responseRecorder.Code, http.StatusBadRequest)
	assert.Equal(t, strings.TrimSpace(responseRecorder.Body.String()), "EOF")
}

func TestValidateRequestBodyMissingRequestProperty(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"id":1}`))
	responseRecorder := httptest.NewRecorder()

	request.Header.Set("Content-Type", "application/json")
	validationController := NewValidationController()
	validationController.Validate(responseRecorder, request)

	assert.Equal(t, responseRecorder.Code, http.StatusBadRequest)
	assert.Contains(t, strings.TrimSpace(responseRecorder.Body.String()), "request is nil")
}

func TestValidateRequestBody(t *testing.T) {
	config.WebhookVersion = "0.0.1"
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{
  "request": {
    "uid": "123",
    "object": {
      "metadata": {
        "managedFields": [
          {
            "manager": "kube-controller"
          }
        ]
      }
    },
		"dryRun": false
  }
}`))
	responseRecorder := httptest.NewRecorder()

	request.Header.Set("Content-Type", "application/json")
	validationController := NewValidationController()
	validationController.Validate(responseRecorder, request)

	assert.Equal(t, responseRecorder.Code, http.StatusOK)
	assert.Contains(t, strings.TrimSpace(responseRecorder.Body.String()), "We're good!")
}

func TestValidateRequestBodyWithNotAllowedK8sResource(t *testing.T) {
	// fakeUUID := "b218c483-b5cc-4839-8c3c-1bf3e8e6e839"
	// t.Setenv(enums.Token, fakeUUID)
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyRequestNotAllowedJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := NewValidationController()
	validationController.Validate(responseRecorder, request)

	assert.Contains(t, strings.TrimSpace(responseRecorder.Body.String()), "\"allowed\":false")
}

func TestValidateRequestBodyWithAllowedK8sResource(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyRequestAllowedJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := NewValidationController()
	validationController.Validate(responseRecorder, request)

	body := responseRecorder.Body.String()

	assert.Contains(t, strings.TrimSpace(body), "\"allowed\":true")
}

func TestValidateRequestBodyWithFluxCDResource(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyAllowedRequestFluxCDJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := NewValidationController()
	validationController.Validate(responseRecorder, request)

	body := responseRecorder.Body.String()

	assert.Contains(t, strings.TrimSpace(body), "\"allowed\":true")
}

func TestValidateRequestBodyWithFluxCDResourceWithoutLabels(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyAllowedRequestFluxCDJsonNoLabels))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := NewValidationController()
	validationController.Validate(responseRecorder, request)

	body := responseRecorder.Body.String()

	assert.Contains(t, strings.TrimSpace(body), "\"allowed\":true")
}
