package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/datreeio/admission-webhook-datree/pkg/openshiftService"

	"github.com/datreeio/admission-webhook-datree/pkg/clients"
	"github.com/datreeio/admission-webhook-datree/pkg/k8sMetadataUtil"
	servicestate "github.com/datreeio/admission-webhook-datree/pkg/serviceState"

	"github.com/datreeio/admission-webhook-datree/pkg/logger"

	"github.com/datreeio/admission-webhook-datree/pkg/errorReporter"
	"github.com/google/uuid"

	"github.com/datreeio/admission-webhook-datree/pkg/responseWriter"
	"github.com/datreeio/admission-webhook-datree/pkg/services"
	admission "k8s.io/api/admission/v1"

	"github.com/datreeio/datree/pkg/utils"
)

type ValidationController struct {
	ValidationService *services.ValidationService
	ErrorReporter     *errorReporter.ErrorReporter
	logger            *logger.Logger
}

func NewValidationController(cliServiceClient *clients.CliClient, state *servicestate.ServiceState, errorReporter *errorReporter.ErrorReporter, k8sMetadataUtilInstance *k8sMetadataUtil.K8sMetadataUtil, logger *logger.Logger, openshiftService *openshiftService.OpenshiftService) *ValidationController {
	validationService := &services.ValidationService{
		CliServiceClient: cliServiceClient,
		State:            state,
		K8sMetadataUtil:  k8sMetadataUtilInstance,
		ErrorReporter:    errorReporter,
		OpenshiftService: openshiftService,
		Logger:           logger,
	}

	return &ValidationController{
		ValidationService: validationService,
		ErrorReporter:     errorReporter,
		logger:            logger,
	}
}

func (c *ValidationController) Validate(w http.ResponseWriter, req *http.Request) {
	c.logger.SetRequestId(uuid.NewString())

	var warningMessages []string
	writer := responseWriter.New(w)

	if req.Method != http.MethodPost {
		writer.NotAllowed("Method not allowed")
		return
	}

	err := headerValidation(req)
	if err != nil {
		c.logger.LogAndReportUnexpectedError(fmt.Sprintf("header validation failed: %s", err))
		writer.BadRequest(err.Error())
		return
	}

	admissionReviewReq, err := ParseHTTPRequestBodyToAdmissionReview(req.Body)
	if err != nil {
		c.logger.LogAndReportUnexpectedError(fmt.Sprintf("parsing request body failed: %s", err))
		writer.BadRequest(err.Error())
		return
	}

	// global panic errors handler
	defer func() {
		if panicErr := recover(); panicErr != nil {
			c.ErrorReporter.ReportPanicError(panicErr)
			c.logger.LogError(utils.ParseErrorToString(panicErr))
			warningMessages = append(warningMessages, "Datree failed to validate the applied resource. Check the pod logs for more details.")
			writer.WriteBody(services.ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, true, utils.ParseErrorToString(panicErr), warningMessages))
		}
	}()

	c.logger.LogAdmissionRequest(admissionReviewReq, false, logger.Incoming)
	admissionReview, isSkipped := c.ValidationService.Validate(admissionReviewReq, &warningMessages)
	writer.WriteBody(admissionReview)

	admissionReview.Request = admissionReviewReq.Request
	c.logger.LogAdmissionRequest(admissionReview, isSkipped, logger.Outgoing)
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
