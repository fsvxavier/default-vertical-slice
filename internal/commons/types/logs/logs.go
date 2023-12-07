package logs

import "github.com/fsvxavier/default-vertical-slice/pkg/apierrors"

type (
	Log struct {
		TraceId any               `json:"trace_id,omitempty"`
		Url     string            `json:"url,omitempty"`
		Headers map[string]string `json:"headers,omitempty"`
		Body    string            `json:"body,omitempty"`
	}

	SuccessLogMessage struct {
		Data       any    `json:"data,omitempty"`
		Error      error  `json:"error,omitempty"`
		TraceID    string `json:"trace_id,omitempty"`
		ClientId   string `json:"client_id,omitempty"`
		HTTPStatus int    `json:"http_status,omitempty"`
	}

	HttpResponseLogMessage struct {
		Request    HttpRequestLogMessage `json:"request,omitempty"`
		TraceID    string                `json:"trace_id,omitempty"`
		ClientId   string                `json:"client_id,omitempty"`
		Response   string                `json:"response,omitempty"`
		HTTPStatus int                   `json:"http_status,omitempty"`
	}

	HttpRequestLogMessage struct {
		TraceId                 any               `json:"trace_id,omitempty"`
		ClientId                any               `json:"client_id,omitempty"`
		BookId                  any               `json:"book_id,omitempty"`
		AccountId               any               `json:"account_id,omitempty"`
		PersonId                any               `json:"person_id,omitempty"`
		ProductId               any               `json:"product_id,omitempty"`
		OperationStepInstanceId any               `json:"operation_step_instance_id,omitempty"`
		OperationInstanceId     any               `json:"operation_instance_id,omitempty"`
		Path                    string            `json:"path,omitempty"`
		Params                  any               `json:"params,omitempty"`
		Headers                 map[string]string `json:"headers,omitempty"`
		Query                   any               `json:"query,omitempty"`
		Body                    any               `json:"body,omitempty"`
		Method                  string            `json:"method,omitempty"`
	}

	ErrorLogMessage struct {
		Data       interface{}            `json:"data,omitempty"`
		TraceID    string                 `json:"trace_id,omitempty"`
		ClientId   string                 `json:"client_id,omitempty"`
		Entity     string                 `json:"entity,omitempty"`
		Error      apierrors.DockApiError `json:"response,omitempty"`
		HTTPStatus int                    `json:"http_status,omitempty"`
	}
)
