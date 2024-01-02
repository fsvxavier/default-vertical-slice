package nethttp

import (
	"encoding/json"
	"net/http"
	"regexp"
)

type Route struct {
	Handler http.Handler
	Method  string
	Pattern string
}
type Router struct {
	Prefix string
	routes []Route
}

func NewRouter() *Router {
	return &Router{}
}

type Handler func(r *http.Request) (statusCode int, data map[string]interface{})

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	statusCode, data := h(r)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (r *Router) getHandler(method, path string) http.Handler {
	for _, route := range r.routes {
		re := regexp.MustCompile(route.Pattern)
		if route.Method == method && re.MatchString(path) {
			return route.Handler
		}
	}
	return http.NotFoundHandler()
}

func (r *Router) AddRoute(method, path string, handler http.Handler) {
	r.routes = append(r.routes, Route{Method: method, Pattern: r.Prefix + path, Handler: handler})
}

func (r *Router) GET(path string, handler Handler) {
	r.AddRoute(http.MethodGet, path, handler)
}

func (r *Router) POST(path string, handler Handler) {
	r.AddRoute(http.MethodPost, path, handler)
}

func (r *Router) PUT(path string, handler Handler) {
	r.AddRoute(http.MethodPut, path, handler)
}

func (r *Router) DELETE(path string, handler Handler) {
	r.AddRoute(http.MethodDelete, path, handler)
}

func (r *Router) CONNECT(path string, handler Handler) {
	r.AddRoute(http.MethodConnect, path, handler)
}

func (r *Router) PATH(path string, handler Handler) {
	r.AddRoute(http.MethodPatch, path, handler)
}

func (r *Router) TRACE(path string, handler Handler) {
	r.AddRoute(http.MethodTrace, path, handler)
}

func (r *Router) HEAD(path string, handler Handler) {
	r.AddRoute(http.MethodHead, path, handler)
}

func (r *Router) OPTIONS(path string, handler Handler) {
	r.AddRoute(http.MethodOptions, path, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	handler := r.getHandler(method, path)

	// handler middlewares go here

	handler.ServeHTTP(w, req)
}

// trimRight is the equivalent of strings.TrimRight.
func trimRight(s string, cutset byte) string {
	lenStr := len(s)
	for lenStr > 0 && s[lenStr-1] == cutset {
		lenStr--
	}
	return s[:lenStr]
}

func getGroupPath(prefix, path string) string {
	if len(path) == 0 {
		return prefix
	}

	if path[0] != '/' {
		path = "/" + path
	}

	return trimRight(prefix, '/') + path
}

// Group is used for Routes with common prefix to define a new sub-router.
func (rtr *Router) Group(prefix string) *Router {
	rtr.Prefix = getGroupPath(rtr.Prefix, prefix)
	return rtr
}
