package controllers

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/datreeio/admission-webhook-datree/pkg/clients"
	"github.com/datreeio/admission-webhook-datree/pkg/config"
	"github.com/datreeio/admission-webhook-datree/pkg/services"
	"github.com/datreeio/datree/pkg/httpClient"
	"github.com/datreeio/datree/pkg/networkValidator"
	"github.com/stretchr/testify/assert"
	admission "k8s.io/api/admission/v1"
)

//go:embed test_fixtures/applyNotAllowedRequest.json
var applyRequestNotAllowedJson string

//go:embed test_fixtures/getPrerunDataResponse.json
var getPrerunDataResponse []byte

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
	validationController := mockValidationController(httpClient.Response{})
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
	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
	})
	validationController.Validate(responseRecorder, request)

	assert.Equal(t, responseRecorder.Code, http.StatusOK)
	assert.Equal(t, responseToAdmissionResponse(responseRecorder.Body.String()).Result.Message, "We're good!")
}

func TestValidateRequestBodyWithNotAllowedK8sResource(t *testing.T) {
	t.Setenv("DATREE_ENFORCE", "true")
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyRequestNotAllowedJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
		Body:       getPrerunDataResponse,
	})

	validationController.Validate(responseRecorder, request)
	assert.Equal(t, responseToAdmissionResponse(responseRecorder.Body.String()).Allowed, false)
}

func TestValidateRequestBodyWithNotAllowedK8sResourceEnforceModeOff(t *testing.T) {
	t.Setenv("DATREE_ENFORCE", "false")
	var applyRequestNotAllowed admission.AdmissionReview
	json.Unmarshal([]byte(applyRequestNotAllowedJson), &applyRequestNotAllowed)
	resourceKind := applyRequestNotAllowed.Request.Kind.Kind
	resourceName := applyRequestNotAllowed.Request.Name
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyRequestNotAllowedJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
		Body:       getPrerunDataResponse,
	})
	validationController.Validate(responseRecorder, request)

	admissionResponse := responseToAdmissionResponse(responseRecorder.Body.String())
	warningMessage := fmt.Sprintf("🚩 Object with name %s and kind %s failed the policy check, get the full report at: https://app.staging.datree.io/cli/invocations", resourceName, resourceKind)
	assert.Equal(t, admissionResponse.Allowed, true)
	assert.Contains(t, admissionResponse.Warnings[0], warningMessage)
	assert.Contains(t, admissionResponse.Warnings[0], "?webhook=true")
}

func TestValidateRequestBodyWithAllowedK8sResource(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyRequestAllowedJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
		Body:       getPrerunDataResponse,
	})

	validationController.Validate(responseRecorder, request)

	body := responseRecorder.Body.String()

	assert.Equal(t, responseToAdmissionResponse(body).Allowed, true)
}

func TestValidateRequestBodyWithFluxCDResource(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyAllowedRequestFluxCDJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
		Body:       getPrerunDataResponse,
	})
	validationController.Validate(responseRecorder, request)

	body := responseRecorder.Body.String()

	assert.Equal(t, responseToAdmissionResponse(body).Allowed, true)
}

func TestValidateRequestBodyWithFluxCDResourceWithoutLabels(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyAllowedRequestFluxCDJsonNoLabels))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
		Body:       getPrerunDataResponse,
	})
	validationController.Validate(responseRecorder, request)

	body := responseRecorder.Body.String()

	assert.Equal(t, responseToAdmissionResponse(body).Allowed, true)
}

func responseToAdmissionResponse(response string) *admission.AdmissionResponse {
	var admissionReview admission.AdmissionReview
	err := json.Unmarshal([]byte(response), &admissionReview)
	if err != nil {
		panic(err)
	}

	return admissionReview.Response
}

type MockHttpClient struct {
	mockedResponse httpClient.Response
}

func (mhc *MockHttpClient) Request(method string, resourceURI string, body interface{}, headers map[string]string) (httpClient.Response, error) {
	return mhc.mockedResponse, nil
}

func mockValidationController(mockedResponse httpClient.Response) *ValidationController {
	mockedHttpClient := &MockHttpClient{mockedResponse: mockedResponse}
	mockedCliServiceClient := clients.NewCustomCliServiceClient("", mockedHttpClient, nil, []string{}, networkValidator.NewNetworkValidator(), make(map[string]string))
	mockedValidationService := services.NewValidationServiceWithCustomCliServiceClient(mockedCliServiceClient)

	return &ValidationController{
		ValidationService: mockedValidationService,
	}
}
