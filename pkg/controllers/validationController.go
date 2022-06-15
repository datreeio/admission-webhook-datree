package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/datreeio/webhook-datree/pkg/errorReporter"

	"github.com/datreeio/webhook-datree/pkg/responseWriter"
	"github.com/datreeio/webhook-datree/pkg/services"
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

// Validate TODO: think about the name of the controller
func (c *ValidationController) Validate(w http.ResponseWriter, req *http.Request) {
	writer := responseWriter.New(w)

	defer func() {
		if panicErr := recover(); panicErr != nil {
			validator := networkValidator.NewNetworkValidator()
			newCliClient := cliClient.NewCliClient(deploymentConfig.URL, validator)
			newLocalConfigClient := localConfig.NewLocalConfigClient(newCliClient, validator)
			reporter := errorReporter.NewErrorReporter(newCliClient, newLocalConfigClient)
			reporter.ReportPanicError(panicErr)
			admissionReviewReq, err := ParseHTTPRequestBodyToAdmissionReview(req.Body)
			if err != nil {
				writer.BadRequest(err.Error())
				return
			}
			writer.WriteBody(services.ParseEvaluationResponseIntoAdmissionReview(admissionReviewReq.Request.UID, false, utils.ParseErrorToString(panicErr)))
		}
	}()

	err := headerValidation(req)
	if err != nil {
		writer.BadRequest(err.Error())
		return
	}

	switch req.Method {
	case http.MethodPost:
		admissionReviewReq, err := ParseHTTPRequestBodyToAdmissionReview(req.Body)
		if err != nil {
			writer.BadRequest(err.Error())
			return
		}
		res := services.Validate(admissionReviewReq)
		writer.WriteBody(res)
		return
	default:
		writer.NotAllowed("Method not allowed")
	}
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
	if err != nil || admissionReviewReq.Request == nil {
		return &admissionReviewReq, fmt.Errorf("request is nil %s", err)
	}

	return &admissionReviewReq, nil
}
