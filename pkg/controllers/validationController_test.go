package controllers

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	"github.com/datreeio/admission-webhook-datree/pkg/openshiftService"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/datreeio/admission-webhook-datree/pkg/enums"
	"github.com/datreeio/admission-webhook-datree/pkg/errorReporter"
	servicestate "github.com/datreeio/admission-webhook-datree/pkg/serviceState"
	"github.com/stretchr/testify/mock"

	"github.com/datreeio/admission-webhook-datree/pkg/k8sMetadataUtil"

	"github.com/datreeio/admission-webhook-datree/pkg/clients"
	"github.com/datreeio/admission-webhook-datree/pkg/config"
	"github.com/datreeio/datree/pkg/httpClient"
	"github.com/datreeio/datree/pkg/networkValidator"
	"github.com/stretchr/testify/assert"
	admission "k8s.io/api/admission/v1"
	"k8s.io/client-go/kubernetes/fake"
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

func setMockEnv(t *testing.T) {
	t.Setenv(enums.Token, "test-token")
	t.Setenv(enums.ClusterName, "test-cluster-name")
	t.Setenv(enums.Policy, "Default")
	t.Setenv(enums.Enforce, "true")
}

func TestHeaderValidation(t *testing.T) {
	setMockEnv(t)
	request := httptest.NewRequest(http.MethodPost, "/validate", nil)
	responseRecorder := httptest.NewRecorder()

	request.Header.Set("Content-Type", "text/html")
	validationController := mockValidationController(httpClient.Response{})
	validationController.Validate(responseRecorder, request)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, "Content-Type header is not application/json", strings.TrimSpace(responseRecorder.Body.String()))
}

func TestValidateHttpMethod(t *testing.T) {
	setMockEnv(t)
	request := httptest.NewRequest(http.MethodGet, "/validate", nil)
	responseRecorder := httptest.NewRecorder()

	request.Header.Set("Content-Type", "application/json")
	validationController := mockValidationController(httpClient.Response{})
	validationController.Validate(responseRecorder, request)

	assert.Equal(t, http.StatusMethodNotAllowed, responseRecorder.Code)
	assert.Equal(t, "Method not allowed", strings.TrimSpace(responseRecorder.Body.String()))
}

func TestValidateRequestBodyEmpty(t *testing.T) {
	setMockEnv(t)
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(""))
	responseRecorder := httptest.NewRecorder()

	request.Header.Set("Content-Type", "application/json")
	validationController := mockValidationController(httpClient.Response{})
	validationController.Validate(responseRecorder, request)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Equal(t, "EOF", strings.TrimSpace(responseRecorder.Body.String()))
}

func TestValidateRequestBodyMissingRequestProperty(t *testing.T) {
	setMockEnv(t)
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"id":1}`))
	responseRecorder := httptest.NewRecorder()

	request.Header.Set("Content-Type", "application/json")
	validationController := mockValidationController(httpClient.Response{})
	validationController.Validate(responseRecorder, request)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	assert.Contains(t, strings.TrimSpace(responseRecorder.Body.String()), "request is nil")
}

func TestValidateRequestBody(t *testing.T) {
	config.WebhookVersion = "0.0.1"
	setMockEnv(t)
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

	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	assert.Equal(t, "We're good!", responseToAdmissionResponse(responseRecorder.Body.String()).Result.Message)
}

func TestValidateRequestBodyWithNotAllowedK8sResource(t *testing.T) {
	setMockEnv(t)
	t.Setenv("DATREE_ENFORCE", "true")
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyRequestNotAllowedJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
		Body:       getPrerunDataResponse,
	})

	validationController.Validate(responseRecorder, request)
	assert.Equal(t, false, responseToAdmissionResponse(responseRecorder.Body.String()).Allowed)
}

func TestValidateRequestBodyWithNotAllowedK8sResourceBypassed(t *testing.T) {
	setMockEnv(t)
	t.Setenv("DATREE_ENFORCE", "true")
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyRequestNotAllowedJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
		Body:       getPrerunDataResponse,
	})

	validationController.ValidationService.State.SetBypassPermissions(&servicestate.BypassPermissions{UserAccounts: []string{"admin"}})

	validationController.Validate(responseRecorder, request)
	assert.Equal(t, true, responseToAdmissionResponse(responseRecorder.Body.String()).Allowed)
}

func TestValidateRequestBodyWithNotAllowedK8sResourceEnforceModeOff(t *testing.T) {
	setMockEnv(t)
	t.Setenv(enums.Enforce, "false")
	var applyRequestNotAllowed admission.AdmissionReview

	err := json.Unmarshal([]byte(applyRequestNotAllowedJson), &applyRequestNotAllowed)
	if err != nil {
		fmt.Printf("json unmarshal error: %s \n", err.Error())
	}
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyRequestNotAllowedJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
		Body:       getPrerunDataResponse,
	})
	validationController.Validate(responseRecorder, request)

	admissionResponse := responseToAdmissionResponse(responseRecorder.Body.String())

	expectedWarningMessages := []string{
		"🚩 Object with name \"my-deployment\" and kind \"Scale\" failed the policy check",
		"👉 Get the full report https://app.staging.datree.io/cli/invocations/",
	}
	assert.Equal(t, true, admissionResponse.Allowed)
	assert.Contains(t, admissionResponse.Warnings[0], expectedWarningMessages[0])
	assert.Contains(t, admissionResponse.Warnings[1], expectedWarningMessages[1])
	assert.Contains(t, admissionResponse.Warnings[1], "webhook=true")
}

func TestValidateRequestBodyWithPolicyNotExists(t *testing.T) {
	setMockEnv(t)
	t.Setenv(enums.Enforce, "true")

	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyRequestNotAllowedJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
		Body: getAndMutatePrerunResponse(func(prerunResponse *clients.ClusterEvaluationPrerunDataResponse) {
			prerunResponse.ActivePolicies = []string{"NotExistsPolicy"}
		}),
	})

	validationController.Validate(responseRecorder, request)
	admissionResponse := responseToAdmissionResponse(responseRecorder.Body.String())

	assert.Equal(t, true, admissionResponse.Allowed)
	assert.Contains(t, admissionResponse.Warnings[0], "Policy NotExistsPolicy not found, skipping evaluation")
}

func TestValidateRequestBodyWithAllowedK8sResource(t *testing.T) {
	setMockEnv(t)
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyRequestAllowedJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
		Body:       getPrerunDataResponse,
	})

	validationController.Validate(responseRecorder, request)

	body := responseRecorder.Body.String()

	assert.Equal(t, true, responseToAdmissionResponse(body).Allowed)
}

func TestValidateRequestBodyWithFluxCDResource(t *testing.T) {
	setMockEnv(t)
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyAllowedRequestFluxCDJson))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
		Body:       getPrerunDataResponse,
	})
	validationController.Validate(responseRecorder, request)

	body := responseRecorder.Body.String()

	assert.Equal(t, true, responseToAdmissionResponse(body).Allowed)
}

func TestValidateRequestBodyWithFluxCDResourceWithoutLabels(t *testing.T) {
	setMockEnv(t)
	request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(applyAllowedRequestFluxCDJsonNoLabels))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()

	validationController := mockValidationController(httpClient.Response{
		StatusCode: http.StatusOK,
		Body:       getPrerunDataResponse,
	})
	validationController.Validate(responseRecorder, request)

	body := responseRecorder.Body.String()

	assert.Equal(t, true, responseToAdmissionResponse(body).Allowed)
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

type MockErrorReporterClient struct {
	mock.Mock
}

func (m *MockErrorReporterClient) ReportError(reportCliErrorRequest clients.ReportErrorRequest, uri string) (StatusCode int, Error error) {
	m.Called(reportCliErrorRequest, uri)
	return 200, nil
}

func mockValidationController(mockedResponse httpClient.Response) *ValidationController {
	mockedHttpClient := &MockHttpClient{mockedResponse: mockedResponse}
	mockedCliServiceClient := clients.NewCustomCliServiceClient("", mockedHttpClient, nil, []string{}, networkValidator.NewNetworkValidator(), make(map[string]string))
	mockK8sMetadataUtil := &k8sMetadataUtil.K8sMetadataUtil{
		ClientSet: fake.NewSimpleClientset(),
	}

	mockState := servicestate.New()
	mockState.SetClusterUuid("test-cluster-uuid")
	mockState.SetK8sVersion("1.18.0")

	mockErrorReporterClient := &MockErrorReporterClient{}
	mockErrorReporterClient.On("ReportError", mock.Anything, mock.Anything).Return(200, nil)
	mockErrorReporter := errorReporter.NewErrorReporter(mockErrorReporterClient, mockState)

	mockLogger := &logger.Logger{}

	mockOpenshiftService := &openshiftService.OpenshiftService{}

	return NewValidationController(mockedCliServiceClient, mockState, mockErrorReporter, mockK8sMetadataUtil, mockLogger, mockOpenshiftService)
}

func convertPrerunResponseJsonToStruct(prerunResponse []byte) *clients.ClusterEvaluationPrerunDataResponse {
	var prerunResponseStruct clients.ClusterEvaluationPrerunDataResponse
	err := json.Unmarshal(prerunResponse, &prerunResponseStruct)
	if err != nil {
		panic(err)
	}
	return &prerunResponseStruct
}

func convertPrerunResponseStructToBytes(prerunResponse *clients.ClusterEvaluationPrerunDataResponse) []byte {
	prerunResponseJson, err := json.Marshal(prerunResponse)
	if err != nil {
		panic(err)
	}
	return prerunResponseJson
}

func getAndMutatePrerunResponse(mutate func(*clients.ClusterEvaluationPrerunDataResponse)) []byte {
	prerunResponseStruct := convertPrerunResponseJsonToStruct(getPrerunDataResponse)
	mutate(prerunResponseStruct)
	return convertPrerunResponseStructToBytes(prerunResponseStruct)
}
