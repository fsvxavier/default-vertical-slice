// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package fibertrace

import (
	"math"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
)

const (
	defaultServiceName = "fiber"
	defaultSpanName    = "http.request"
)

type config struct {
	isStatusError func(statusCode int) bool
	resourceNamer func(*fiber.Ctx) string
	serviceName   string
	spanName      string
	spanOpts      []ddtrace.StartSpanOption
	analyticsRate float64
}

// Option represents an option that can be passed to NewRouter.
type Option func(*config)

func defaults(cfg *config) {
	cfg.serviceName = defaultServiceName
	cfg.spanName = defaultSpanName
	cfg.isStatusError = isServerError
	cfg.resourceNamer = defaultResourceNamer
	cfg.analyticsRate = math.NaN()
}

// WithServiceName sets the given service name for the router.
func WithServiceName(name string) Option {
	return func(cfg *config) {
		cfg.serviceName = name
	}
}

// WithSpanOptions applies the given set of options to the spans started
// by the router.
func WithSpanOptions(opts ...ddtrace.StartSpanOption) Option {
	return func(cfg *config) {
		cfg.spanOpts = opts
	}
}

// WithAnalytics enables Trace Analytics for all started spans.
func WithAnalytics(on bool) Option {
	return func(cfg *config) {
		if on {
			cfg.analyticsRate = 1.0
		} else {
			cfg.analyticsRate = math.NaN()
		}
	}
}

// WithAnalyticsRate sets the sampling rate for Trace Analytics events
// correlated to started spans.
func WithAnalyticsRate(rate float64) Option {
	return func(cfg *config) {
		if rate >= 0.0 && rate <= 1.0 {
			cfg.analyticsRate = rate
		} else {
			cfg.analyticsRate = math.NaN()
		}
	}
}

// WithStatusCheck allow setting of a function to tell whether a status code is an error.
func WithStatusCheck(fn func(statusCode int) bool) Option {
	return func(cfg *config) {
		cfg.isStatusError = fn
	}
}

// WithResourceNamer specifies a function which will be used to
// obtain the resource name for a given request taking the go-fiber context
// as input.
func WithResourceNamer(fn func(*fiber.Ctx) string) Option {
	return func(cfg *config) {
		cfg.resourceNamer = fn
	}
}

func defaultResourceNamer(c *fiber.Ctx) string {
	r := c.Route()
	return r.Method + " " + r.Path
}

func isServerError(statusCode int) bool {
	return statusCode >= 500 && statusCode < 600
}
