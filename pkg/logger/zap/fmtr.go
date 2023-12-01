package zap

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	json "github.com/json-iterator/go"
)

type ErrorLogMessage struct {
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
	Code    string `json:"code"`
	File    string `json:"file"`
	Line    int    `json:"line"`
}

var traceID string

const (
	TWO = 2
)

func Formatter(ctx context.Context, code, message string) (ctxs context.Context, msg string) {
	if ctx.Value("trace_id") != nil {
		traceID = ctx.Value("trace_id").(string)
	}

	// Get file and line
	pc, file, line, _ := runtime.Caller(TWO) // 2 because we are calling from this file
	funcName := runtime.FuncForPC(pc).Name()
	funcNameSplit := strings.Split(funcName, "/")
	funcName = funcNameSplit[len(funcNameSplit)-1]

	logMessage, err := json.Marshal(ErrorLogMessage{
		Message: message,
		TraceID: traceID,
		Code:    code,
		File:    formatFilename(file),
		Line:    line,
	})
	if err != nil {
		fmt.Errorf("marshal Log error: %s", err.Error())
		return
	}

	return ctx, string(logMessage)
}

func formatFilename(filename string) string {
	fname := regexp.MustCompile(`exchange.+`).FindString(filename)
	if fname == "" {
		return filename
	}

	return fname
}
