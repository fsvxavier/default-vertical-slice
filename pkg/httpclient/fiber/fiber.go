package fiber

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var (
	headerContentTypeJson = []byte("application/json")
	json                  = jsoniter.ConfigCompatibleWithStandardLibrary
)

type Request struct {
	structUnmarshal any
	headers         map[string]string
	client          *fiber.Client
	baseURL         string
}

type Response struct {
	Body       []byte
	StatusCode int
	IsError    bool
}

type IHttpRequest interface {
	Get(endpoint string) (*Response, error)
	Post(endpoint string, body []byte) (*Response, error)
	Put(endpoint string, body []byte) (*Response, error)
	Delete(endpoint string) (*Response, error)
}

func NewClient() *fiber.Client {
	client := fiber.AcquireClient()
	defer fiber.ReleaseClient(client)

	client = &fiber.Client{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	}

	return client
}

// New method creates a new httprequest client.
func New(url string) *Request {
	client := fiber.AcquireClient()
	defer fiber.ReleaseClient(client)

	client = &fiber.Client{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	}

	httpRequest := &Request{
		baseURL: url,
		client:  client,
	}

	return httpRequest
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
func (req *Request) SetHeaders(headers map[string]string) *Request {
	req.headers = headers
	return req
}

// SetErrorHandler method is to register the response `ErrorHandler` for current `Request`.
func (req *Request) SetBaseURL(baseURL string) *Request {
	req.baseURL = baseURL

	return req
}

// Post method performs the HTTP POST request for current `Request`.
func (req *Request) Post(ctx context.Context, endpoint string, body []byte) (*Response, error) {
	return req.Execute(ctx, fiber.MethodPost, endpoint, body)
}

// Get method performs the HTTP GET request for current `Request`.
func (req *Request) Get(ctx context.Context, endpoint string) (*Response, error) {
	return req.Execute(ctx, fiber.MethodGet, endpoint, nil)
}

// Put method performs the HTTP PUT request for current `Request`.
func (req *Request) Put(ctx context.Context, endpoint string, body []byte) (*Response, error) {
	return req.Execute(ctx, fiber.MethodPut, endpoint, body)
}

// Delete method performs the HTTP DELETE request for current `Request`.
func (req *Request) Delete(ctx context.Context, endpoint string) (*Response, error) {
	return req.Execute(ctx, fiber.MethodDelete, endpoint, nil)
}

// Unmarshal method unmarshals the HTTP response body to given struct.
func (req *Request) Unmarshal(v any) *Request {
	req.structUnmarshal = v
	return req
}

// Execute method performs the HTTP request with given HTTP method, Endpoint and Body for current `Request`.
func (req *Request) Execute(ctx context.Context, method, endpoint string, body []byte) (*Response, error) {
	ddSpan, ok := tracer.SpanFromContext(ctx)
	if ok {
		err := tracer.Inject(ddSpan.Context(), tracer.TextMapCarrier(req.headers))
		if err != nil {
			return nil, err
		}
	}

	// You may read the timeouts from some config
	readTimeout, err := time.ParseDuration("50ms")
	if err != nil {
		fmt.Println(err.Error())
	}
	writeTimeout, err := time.ParseDuration("50ms")
	if err != nil {
		fmt.Println(err.Error())
	}
	maxIdleConnDuration, err := time.ParseDuration("30m")
	if err != nil {
		fmt.Println(err.Error())
	}
	maxConnDuration, err := time.ParseDuration("30s")
	if err != nil {
		fmt.Println(err.Error())
	}
	maxConnWaitTimeout, err := time.ParseDuration("3s")
	if err != nil {
		fmt.Println(err.Error())
	}

	agent := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(agent)

	agent.HostClient = &fasthttp.HostClient{
		ReadTimeout:              readTimeout,
		MaxConns:                 100,
		WriteTimeout:             writeTimeout,
		MaxIdleConnDuration:      maxIdleConnDuration,
		MaxConnWaitTimeout:       maxConnWaitTimeout,
		MaxConnDuration:          maxConnDuration,
		IsTLS:                    false,
		NoDefaultUserAgentHeader: true, // Don't send: User-Agent: fasthttp
		DisablePathNormalizing:   true,
		// increase DNS cache time to an hour instead of default minute
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}

	agent.Request().Header.SetMethod(method)
	agent.Request().SetRequestURI(req.baseURL + endpoint)
	agent.InsecureSkipVerify()
	agent.Reuse()

	agent.Request().Header.SetContentTypeBytes(headerContentTypeJson)

	if body != nil {
		agent.Request().SetBody([]byte(body))
	}

	if req.headers == nil {
		req.headers = make(map[string]string)
	}

	for k, v := range req.headers {
		agent.Request().Header.Add(k, v)
	}

	err = agent.Parse()
	if err != nil {
		return nil, err
	}

	isErrors := false
	respStatusCode, respBody, respErrs := agent.Bytes()
	if len(respErrs) > 0 {
		isErrors = true
	}

	if req.structUnmarshal != nil {
		json.Unmarshal(respBody, req.structUnmarshal)
	}

	response := &Response{
		Body:       respBody,
		StatusCode: respStatusCode,
		IsError:    isErrors,
	}

	return response, nil
}
