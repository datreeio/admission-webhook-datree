package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	"io"
	"net/http"

	"github.com/datreeio/admission-webhook-datree/pkg/errorReporter"
	"github.com/google/uuid"

	"github.com/datreeio/admission-webhook-datree/pkg/responseWriter"
	"github.com/datreeio/admission-webhook-datree/pkg/services"
	admission "k8s.io/api/admission/v1"

	"github.com/datreeio/datree/pkg/cliClient"
	"github.com/datreeio/datree/pkg/deploymentConfig"
	"github.com/datreeio/datree/pkg/localConfig"
	"github.com/datreeio/datree/pkg/networkValidator"
	"github.com/datreeio/datree/pkg/utils"
)

type ValidationController struct{}

func NewValidationController() *ValidationController {
	return &ValidationController{}
}

func (c *ValidationController) Validate(w http.ResponseWriter, req *http.Request) {
	internalLogger := logger.New(uuid.NewString())

	var warningMessages []string
	writer := responseWriter.New(w)

	if req.Method != http.MethodPost {
		writer.NotAllowed("Method not allowed")
		return
	}

	err := headerValidation(req)
	if err != nil {
		internalLogger.LogError(fmt.Sprintf("header validation failed: %s", err))
		writer.BadRequest(err.Error())
		return
	}

	admissionReviewReq, err := ParseHTTPRequestBodyToAdmissionReview(req.Body)
	if err != nil {
		internalLogger.LogError(fmt.Sprintf("parsing request body failed: %s", err))
		writer.BadRequest(err.Error())
		return
	}

	// global panic errors handler
	defer func() {
		if panicErr := recover(); panicErr != nil {
			validator := networkValidator.NewNetworkValidator()
			newCliClient := cliClient.NewCliClient(deploymentConfig.URL, validator)
			newLocalConfigClient := localConfig.NewLocalConfigClient(newCliClient, validator)
			reporter := errorReporter.NewErrorReporter(newCliClient, newLocalConfigClient)
			reporter.ReportPanicError(panicErr)
			internalLogger.LogError(utils.ParseErrorToString(panicErr))
			warningMessages = append(warningMessages, "Datree failed to validate the applied resource. Check the pod logs for more details.")
			writer.WriteBody(services.ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, true, utils.ParseErrorToString(panicErr), warningMessages))
		}
	}()

	internalLogger.LogIncoming(admissionReviewReq)
	admissionReview, isSkipped := services.Validate(admissionReviewReq, &warningMessages, internalLogger)
	writer.WriteBody(admissionReview)
	internalLogger.LogOutgoing(admissionReview, isSkipped)
}

func headerValidation(req *http.Request) error {
	if req.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("Content-Type header is not application/json")
	}

	return nil
}

func ParseHTTPRequestBodyToAdmissionReview(body io.ReadCloser) (*admission.AdmissionReview, error) {
	var admissionReviewReq admission.AdmissionReview

	err := json.NewDecoder(body).Decode(&admissionReviewReq)
	if err != nil {
		return &admissionReviewReq, fmt.Errorf("%s", err)
	}
	if admissionReviewReq.Request == nil {
		return &admissionReviewReq, fmt.Errorf("request is nil")
	}

	return &admissionReviewReq, nil
}
