package resty

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var ctx context.Context

type Requester struct {
	client     *resty.Client
	restyReq   *resty.Request
	restyRes   *resty.Response
	errHandler ErrorHandler
	headers    map[string]string
	baseURL    string
}

type Response struct {
	Body       []byte
	StatusCode int
	IsError    bool
}

type IHttpRequest interface {
	Get(endpoint string) (*Response, error)
	Post(endpoint string, body interface{}) (*Response, error)
	Put(endpoint string, body interface{}) (*Response, error)
	Delete(endpoint string) (*Response, error)
	SetHeaders(headers map[string]string) *Requester
	SetErrorHandler(h ErrorHandler) *Requester
	SetBaseURL(baseURL string) *Requester
	Unmarshal(v any) *Requester
}

const (
	REQ_TRACE_ENABLE       = true
	REQ_TRACE_ENABLE_PRINT = true
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ErrorHandler func(*Response) error

// New method creates a new httprequest client.
func NewClient() *resty.Client {
	client := resty.New()

	client.JSONMarshal = json.Marshal
	client.JSONUnmarshal = json.Unmarshal

	defaultRecTraceEnable := REQ_TRACE_ENABLE
	if os.Getenv("REQ_TRACE_ENABLE") != "" {
		defaultRecTraceEnable = (os.Getenv("REQ_TRACE_ENABLE") == "true")
	}
	if defaultRecTraceEnable {
		client.EnableTrace()
	} else {
		client.DisableTrace()
	}

	return client
}

// New method creates a new httprequest client.
func NewRequester(client *resty.Client) *Requester {
	HttpRequest := &Requester{
		client:   client,
		restyReq: client.R(),
	}

	return HttpRequest
}

// SetHeaders method sets multiple headers field and its values at one go in the client instance.
// These headers will be applied to all requests raised from this client instance. Also it can be
// overridden at request level headers options.
// For Example: To set `Content-Type` and `Accept` as `application/json`
//
//	request.SetHeaders(map[string]string{
//			"Content-Type": "application/json",
//			"Accept": "application/json",
//		})
func (req *Requester) SetHeaders(headers map[string]string) *Requester {
	req.headers = headers
	return req
}

// SetErrorHandler method is to register the response `ErrorHandler` for current `Requester`.
func (req *Requester) SetErrorHandler(h ErrorHandler) *Requester {
	req.errHandler = h

	return req
}

// SetErrorHandler method is to register the response `ErrorHandler` for current `Request`.
func (req *Requester) SetBaseURL(baseURL string) *Requester {
	req.baseURL = baseURL
	return req
}

// Post method performs the HTTP POST request for current `Request`.
func (r *Requester) Post(ctx context.Context, endpoint string, body []byte) (*Response, error) {
	return r.Execute(ctx, http.MethodPost, endpoint, body)
}

// Get method performs the HTTP GET request for current `Request`.
func (r *Requester) Get(ctx context.Context, endpoint string) (*Response, error) {
	return r.Execute(ctx, http.MethodGet, endpoint, nil)
}

// Put method performs the HTTP PUT request for current `Request`.
func (r *Requester) Put(ctx context.Context, endpoint string, body []byte) (*Response, error) {
	return r.Execute(ctx, http.MethodPut, endpoint, body)
}

// Delete method performs the HTTP DELETE request for current `Request`.
func (r *Requester) Delete(ctx context.Context, endpoint string) (*Response, error) {
	return r.Execute(ctx, http.MethodDelete, endpoint, nil)
}

// Unmarshal method unmarshals the HTTP response body to given struct.
func (req *Requester) Unmarshal(v any) *Requester {
	req.restyReq.SetResult(v)

	return req
}

// Execute method performs the HTTP request with given HTTP method, Endpoint and Body for current `Request`.
//
//	resp, err := httprequest.New("http://httpbin.org").Execute("GET", "/get", nil)
func (req *Requester) Execute(ctx context.Context, method, endpoint string, body []byte) (*Response, error) {
	span, ctxs := tracer.StartSpanFromContext(ctx, "post.process")
	defer span.Finish()

	rreq := req.restyReq

	if body != nil {
		rreq.SetBody(body)
	}

	req.client.SetBaseURL(req.baseURL)
	req.client.SetHeaders(req.headers)

	rreq.SetContext(ctxs)
	// Inject the span Context in the Request headers
	err := tracer.Inject(span.Context(), tracer.HTTPHeadersCarrier(req.client.Header))
	if err != nil {
		return nil, err
	}

	uriRequest := req.baseURL + endpoint
	rres, err := rreq.Execute(method, uriRequest)
	if err != nil {
		return nil, err
	}

	defaultRecTraceEnable := REQ_TRACE_ENABLE
	if os.Getenv("REQ_TRACE_ENABLE") != "" {
		defaultRecTraceEnable = (os.Getenv("REQ_TRACE_ENABLE") == "true")
	}
	if defaultRecTraceEnable {
		ti := rres.Request.TraceInfo()

		defaultRecTracePrintEnable := REQ_TRACE_ENABLE_PRINT
		if os.Getenv("REQ_TRACE_ENABLE_PRINT") != "" {
			defaultRecTracePrintEnable = (os.Getenv("REQ_TRACE_ENABLE_PRINT") == "true")
		}

		if defaultRecTracePrintEnable {
			fmt.Println("Request Info:")
			fmt.Println("  Body       		 :\n", string(body))
			fmt.Println("  Request URI       :\n", uriRequest)

			// Explore response object
			fmt.Println("Response Info:")
			fmt.Println("  Error      :", err)
			fmt.Println("  Status Code:", rres.StatusCode())
			fmt.Println("  Status     :", rres.Status())
			fmt.Println("  Proto      :", rres.Proto())
			fmt.Println("  Time       :", rres.Time())
			fmt.Println("  Received At:", rres.ReceivedAt())
			fmt.Println("  Body       :\n", rres)
			fmt.Println()

			// Explore trace info
			fmt.Println("Request Trace Info:")
			fmt.Println("  DNSLookup     :", ti.DNSLookup)
			fmt.Println("  ConnTime      :", ti.ConnTime)
			fmt.Println("  TCPConnTime   :", ti.TCPConnTime)
			fmt.Println("  TLSHandshake  :", ti.TLSHandshake)
			fmt.Println("  ServerTime    :", ti.ServerTime)
			fmt.Println("  ResponseTime  :", ti.ResponseTime)
			fmt.Println("  TotalTime     :", ti.TotalTime)
			fmt.Println("  IsConnReused  :", ti.IsConnReused)
			fmt.Println("  IsConnWasIdle :", ti.IsConnWasIdle)
			fmt.Println("  ConnIdleTime  :", ti.ConnIdleTime)
			fmt.Println("  RequestAttempt:", ti.RequestAttempt)
			fmt.Println()
			fmt.Println()
		} else {
			jsonResquestInfo := `{"Body":"%s","URI":"%s"}`

			fmt.Println(fmt.Sprintf(jsonResquestInfo, body, uriRequest))
			fmt.Println()

			jsonResponseInfo := `{"Error":"%v","Status Code":"%d","Status":"%s", "Proto":"%s",` +
				`"Time":"%v","Received At":"%v","Body":"%s"}`

			fmt.Println(fmt.Sprintf(jsonResponseInfo, err, rres.StatusCode(), rres.Status(), rres.Proto(), rres.Time(), rres.ReceivedAt(), rres.Body()))
			fmt.Println()

			jsonTracer := `{"DNSLookup":"%v","URI":"%s","ConnTime":"%v", "TCPConnTime":"%v",` +
				`"TLSHandshake":"%v","ServerTime":"%v","ResponseTime":"%v","TotalTime":"%v","IsConnReused":"%v","IsConnWasIdle":"%v",` +
				`"ConnIdleTime":"%v"}`

			fmt.Println(fmt.Sprintf(jsonTracer, ti.DNSLookup, uriRequest, ti.ConnTime, ti.TCPConnTime, ti.TLSHandshake, ti.ServerTime, ti.ResponseTime, ti.TotalTime, ti.IsConnReused, ti.IsConnWasIdle, ti.ConnIdleTime))
			fmt.Println()
			fmt.Println()
		}
	}

	res := parseResponse(rres)

	var respError error

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		respError = fmt.Errorf("%d-%s", res.StatusCode, string(res.Body))
		if req.errHandler != nil {
			respError = req.errHandler(res)
		}
	}

	return res, respError
}

func parseResponse(res *resty.Response) *Response {
	return &Response{res.Body(), res.StatusCode(), res.IsError()}
}
