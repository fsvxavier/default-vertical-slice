package adapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	json "github.com/json-iterator/go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	domainerrors "github.com/fsvxavier/default-vertical-slice/internal/core/commons/errors"
	"github.com/fsvxavier/default-vertical-slice/internal/core/commons/types/logs"
	"github.com/fsvxavier/default-vertical-slice/internal/core/utils/helpers"
	"github.com/fsvxavier/default-vertical-slice/pkg/apierrors"
	log "github.com/fsvxavier/default-vertical-slice/pkg/logger/zap"
)

type (
	ControllerResponse struct {
		Data       any
		Error      error
		StatusCode int
	}

	ResponseError struct {
		Payload *apierrors.DockApiError
		TraceId string
		Status  int
	}
)

func ResponseAdapter(ctx context.Context, responseWriter *fiber.Ctx, res ControllerResponse) error {
	span, ddCtx := tracer.StartSpanFromContext(ctx, helpers.GetCurrentFuncName(), tracer.SpanType(ext.SpanTypeWeb))
	defer span.Finish()

	if res.Error != nil {
		return processHTTPError(ddCtx, responseWriter, res)
	} else {
		return processHTTPSuccess(ddCtx, responseWriter, res)
	}
}

func processHTTPSuccess(ctx context.Context, responseWriter *fiber.Ctx, res ControllerResponse) (err error) {
	traceId := responseWriter.Get("Trace-Id")
	responseWriter.Status(res.StatusCode)
	responseWriter.Response().Header.Add("Content-Type", "application/json")
	logMessage := logs.SuccessLogMessage{}
	if res.StatusCode != http.StatusNoContent {
		var err error
		var data interface{} = res.Data
		if _, ok := data.([]byte); ok {
			_, err = responseWriter.Write(data.([]byte))
		} else {
			data, _ := json.Marshal(data)
			if len(data) == 0 {
				data = []byte{}
			}
			_, err = responseWriter.Write(data)
		}
		logMessage = logs.SuccessLogMessage{
			TraceID:    traceId,
			HTTPStatus: res.StatusCode,
			Data:       res.Data,
			Error:      err,
		}
	} else {
		_, err := responseWriter.Write([]byte{})
		logMessage = logs.SuccessLogMessage{
			TraceID:    traceId,
			HTTPStatus: res.StatusCode,
			Error:      err,
		}
	}
	fields := append(
		[]log.Field{},
		log.Reflect("data", logMessage),
	)
	log.Info(ctx, "Success", fields...)
	return err
}

func processHTTPError(ctx context.Context, responseWriter *fiber.Ctx, res ControllerResponse) (err error) {
	requestPayload := make(map[string]any)
	json.Unmarshal(responseWriter.Body(), &requestPayload)

	err = res.Error
	traceId := responseWriter.Get("Trace-Id")
	status := 0
	var payload *apierrors.DockApiError
	logMessage := logs.ErrorLogMessage{}
	message := ""
	switch err := err.(type) {
	case *domainerrors.InvalidEntityError:
		status = http.StatusBadRequest
		payload = apierrors.NewDockApiError(status, statusCodeString(status), "Bad request")
		for attr, details := range err.Details {
			payload.AddErrorDetail(strings.ToLower(attr), details...)
		}
		message = fmt.Sprintf("%v: %v", err.Error(), err.EntityName)
		logMessage = logs.ErrorLogMessage{
			TraceID:    traceId,
			HTTPStatus: status,
			Error:      *payload,
			Entity:     err.EntityName,
		}
	case *domainerrors.UnsupportedMediaTypeError:
		status = http.StatusUnsupportedMediaType
		payload = apierrors.NewDockApiError(status, statusCodeString(status), "Unsupported media type")
		message = fmt.Sprintf("%v", err.Error())
		logMessage = logs.ErrorLogMessage{
			TraceID:    traceId,
			HTTPStatus: status,
			Error:      *payload,
		}
	case *domainerrors.UsecaseError:
		status = http.StatusUnprocessableEntity
		payload = apierrors.NewDockApiError(status, statusCodeString(status), err.Error())
		message = err.Error()
		logMessage = logs.ErrorLogMessage{
			TraceID:    traceId,
			HTTPStatus: status,
			Data:       requestPayload,
			Error:      *payload,
		}
	case *domainerrors.NotFoundError:
		status = http.StatusNotFound
		payload = apierrors.NewDockApiError(status, statusCodeString(status), err.Error())
		message = err.Error()
		logMessage = logs.ErrorLogMessage{
			TraceID:    traceId,
			HTTPStatus: status,
			Data:       requestPayload,
			Error:      *payload,
		}
	case *domainerrors.RepositoryError:
		status = http.StatusUnprocessableEntity
		payload = apierrors.NewDockApiError(status, statusCodeString(status), err.Error())
		message = err.Description
		logMessage = logs.ErrorLogMessage{
			TraceID:    traceId,
			HTTPStatus: status,
			Data:       requestPayload,
			Error:      *apierrors.NewDockApiError(status, statusCodeString(status), err.Error()),
		}
	case *domainerrors.ServerError:
		status = http.StatusInternalServerError
		payload = apierrors.NewDockApiError(status, statusCodeString(status), err.Error())
		message = err.InternalError.Error()
		logMessage = logs.ErrorLogMessage{
			TraceID:    traceId,
			HTTPStatus: status,
			Data:       err.Metadata,
			Error:      *payload,
		}
	case *domainerrors.ExternalIntegrationError:
		if err.Code >= 0 && err.Code <= 399 {
			status = 500
		} else {
			status = err.Code
		}

		switch status {
		case 403, 500, 502, 503, 504:
			status = http.StatusInternalServerError
			payload = apierrors.NewDockApiError(status, statusCodeString(status), "Unable to complete request")
		case 400, 404, 422, 429:
			payload = apierrors.NewDockApiError(status, statusCodeString(status), err.Error())
			json.Unmarshal(err.Data, &payload)
		}
		message = fmt.Sprintf("%v: %v", err.Error(), err.Extra())
		logMessage = logs.ErrorLogMessage{
			TraceID:    traceId,
			HTTPStatus: status,
			Data: map[string]any{
				"code":     err.Code,
				"response": string(err.Data),
				"request":  err.Metadata,
			},
			Error: *payload,
		}
	default:
		status = http.StatusInternalServerError
		payload = apierrors.NewDockApiError(status, statusCodeString(status), "Internal server error")
		message = err.Error()
		logMessage = logs.ErrorLogMessage{
			TraceID:    traceId,
			HTTPStatus: status,
			Data:       err,
			Error:      *payload,
		}
	}
	payload.SetId(traceId)

	fields := append(
		[]log.Field{},
		log.Reflect("data", logMessage),
	)
	log.Error(ctx, message, fields...)
	return responseWriter.Status(status).JSON(payload)
}

func (r *ResponseError) Error() string {
	return statusCodeString(r.Status)
}

func statusCodeString(code int) string {
	return fmt.Sprintf(`%v`, code)
}
