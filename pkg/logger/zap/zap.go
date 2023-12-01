package zap

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const ENV_DEV = "dev"

type Field = zap.Field

type Logger struct {
	Config zap.Config
	Out    io.Writer
	Ctx    context.Context
	*zap.Logger
	Level         zap.AtomicLevel
	Zlg           *zap.Logger
	EncoderConfig zapcore.EncoderConfig
	TraceEnabled  bool
}

var Reflect = zap.Reflect

func NewLogger() *Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.LevelKey = "level"
	encoderCfg.NameKey = "logger"
	encoderCfg.MessageKey = "message"
	encoderCfg.StacktraceKey = "stack"
	encoderCfg.LineEnding = zapcore.DefaultLineEnding
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.CallerKey = "caller"
	encoderCfg.FunctionKey = zapcore.OmitKey
	encoderCfg.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

	lvl := zap.NewAtomicLevel()
	levelLog := zap.InfoLevel
	if os.Getenv("LOG_LEVEL") != "" {
		envLevel := os.Getenv("LOG_LEVEL")
		lvlParsed, _ := zap.ParseAtomicLevel(envLevel)
		levelLog = lvlParsed.Level()
	}
	lvl.SetLevel(levelLog)

	var options []zap.Option
	// if TraceEnabled() {
	// 	options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	// }
	if GetEnv() == ENV_DEV {
		options = append(options, zap.Development())
	}

	cfg := zap.Config{
		Level:             lvl,
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		InitialFields:     nil,
	}

	log, err := cfg.Build(options...)
	if err != nil {
		fmt.Println(err)
	}

	return &Logger{
		Zlg:           log,
		Config:        cfg,
		EncoderConfig: encoderCfg,
		Level:         lvl,
	}
}

func (lgr *Logger) Sync() {
	//nolint:errcheck // defer to closing
	defer lgr.Zlg.Sync()
}

func (lgr *Logger) WithContext(ctx context.Context) *Logger {
	lgr.Ctx = ctx
	lgr.setFields(ctx)
	return lgr
}

func (lgr *Logger) WithLevel(level string) *Logger {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		// Sets default level info
		lgr.Level.SetLevel(zapcore.InfoLevel)
		return lgr
	}
	lgr.Level.SetLevel(lvl.Level())
	return lgr
}

func (lgr *Logger) WithFields(fields map[string]interface{}) *Logger {
	var kvs []zap.Field
	if len(fields) > 0 {
		kvs = interface2field(fields)
	}

	ctx := lgr.Zlg.With(kvs...)
	lgr.Zlg = ctx
	return lgr
}

func (lgr *Logger) setFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	sfc, _ := ddtracer.SpanFromContext(ctx)
	fields = append(
		fields,
		zap.String("go_version", runtime.Version()),
		zap.String("app_name", os.Getenv("APP_NAME")),
		zap.String("dd.env", os.Getenv("DD_ENV")),
		zap.String("dd.service", os.Getenv("DD_SERVICE")),
		zap.String("dd.version", os.Getenv("DD_VERSION")),
		zap.Uint64("dd.trace_id", sfc.Context().TraceID()),
		zap.Uint64("dd.span_id", sfc.Context().SpanID()),
	)
	return fields
}

func (lgr *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	lgr.Zlg.Debug(msg, lgr.setFields(ctx, fields...)...)
}

func (lgr *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	lgr.Zlg.Info(msg, lgr.setFields(ctx, fields...)...)
}

func (lgr *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	lgr.Zlg.Warn(msg, lgr.setFields(ctx, fields...)...)
}

func (lgr *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	lgr.Zlg.Error(msg, lgr.setFields(ctx, fields...)...)
}

func (lgr *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	lgr.Zlg.Fatal(msg, lgr.setFields(ctx, fields...)...)
}

func (lgr *Logger) Panic(ctx context.Context, msg string, fields ...zap.Field) {
	lgr.Zlg.Panic(msg, lgr.setFields(ctx, fields...)...)
}

func (lgr *Logger) DebugCode(ctx context.Context, code, msg string, fields ...zap.Field) {
	ctxs, fmsg := Formatter(ctx, code, msg)
	lgr.Zlg.Debug(fmsg, lgr.setFields(ctxs, fields...)...)
}

func (lgr *Logger) InfoCode(ctx context.Context, code, msg string, fields ...zap.Field) {
	ctxs, fmsg := Formatter(ctx, code, msg)
	lgr.Zlg.Info(fmsg, lgr.setFields(ctxs, fields...)...)
}

func (lgr *Logger) WarnCode(ctx context.Context, code, msg string, fields ...zap.Field) {
	ctxs, fmsg := Formatter(ctx, code, msg)
	lgr.Zlg.Warn(fmsg, lgr.setFields(ctxs, fields...)...)
}

func (lgr *Logger) ErrorCode(ctx context.Context, code, msg string, fields ...zap.Field) {
	ctxs, fmsg := Formatter(ctx, code, msg)
	lgr.Zlg.Error(fmsg, lgr.setFields(ctxs, fields...)...)
}

func (lgr *Logger) FatalCode(ctx context.Context, code, msg string, fields ...zap.Field) {
	ctxs, fmsg := Formatter(ctx, code, msg)
	lgr.Zlg.Fatal(fmsg, lgr.setFields(ctxs, fields...)...)
}

func (lgr *Logger) PanicCode(ctx context.Context, code, msg string, fields ...zap.Field) {
	ctxs, fmsg := Formatter(ctx, code, msg)
	lgr.Zlg.Panic(fmsg, lgr.setFields(ctxs, fields...)...)
}

/*-------------------------------------------------------------------*/
// Debug logs a message at level Debug on the context logger.
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	instance := NewLogger()
	instance.Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	instance := NewLogger()
	instance.Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	instance := NewLogger()
	instance.Warn(ctx, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	instance := NewLogger()
	instance.Error(ctx, msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	instance := NewLogger()
	instance.Fatal(ctx, msg, fields...)
}

func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	instance := NewLogger()
	instance.Panic(ctx, msg, fields...)
}

/*-------------------------------------------------------------------*/
// Debug logs a message at level Debug on the standard logger.
func DebugCode(ctx context.Context, code, msg string, fields ...zap.Field) {
	instance := NewLogger()
	ctxs, fmsg := Formatter(ctx, code, msg)
	instance.Debug(ctxs, fmsg, fields...)
}

func InfoCode(ctx context.Context, code, msg string, fields ...zap.Field) {
	instance := NewLogger()
	ctxs, fmsg := Formatter(ctx, code, msg)
	instance.Info(ctxs, fmsg, fields...)
}

func WarnCode(ctx context.Context, code, msg string, fields ...zap.Field) {
	instance := NewLogger()
	ctxs, fmsg := Formatter(ctx, code, msg)
	instance.Warn(ctxs, fmsg, fields...)
}

func ErrorCode(ctx context.Context, code, msg string, fields ...zap.Field) {
	instance := NewLogger()
	ctxs, fmsg := Formatter(ctx, code, msg)
	instance.Error(ctxs, fmsg, fields...)
}

func FatalCode(ctx context.Context, code, msg string, fields ...zap.Field) {
	instance := NewLogger()
	ctxs, fmsg := Formatter(ctx, code, msg)
	instance.Fatal(ctxs, fmsg, fields...)
}

func PanicCode(ctx context.Context, code, msg string, fields ...zap.Field) {
	instance := NewLogger()
	ctxs, fmsg := Formatter(ctx, code, msg)
	instance.Panic(ctxs, fmsg, fields...)
}

/*-------------------------------------------------------------------*/
// Debug logs a message at level Debug on the standard logger.
func DebugOutCtx(msg string, fields ...zap.Field) {
	instance := NewLogger()
	instance.Debug(context.TODO(), msg, fields...)
}

func InfoOutCtx(msg string, fields ...zap.Field) {
	instance := NewLogger()
	instance.Info(context.TODO(), msg, fields...)
}

func WarnOutCtx(msg string, fields ...zap.Field) {
	instance := NewLogger()
	instance.Warn(context.TODO(), msg, fields...)
}

func ErrorOutCtx(msg string, fields ...zap.Field) {
	instance := NewLogger()
	instance.Error(context.TODO(), msg, fields...)
}

func FatalOutCtx(msg string, fields ...zap.Field) {
	instance := NewLogger()
	instance.Fatal(context.TODO(), msg, fields...)
}

func PanicOutCtx(msg string, fields ...zap.Field) {
	instance := NewLogger()
	instance.Panic(context.TODO(), msg, fields...)
}

/*-------------------------------------------------------------------*/
// Debugf logs a message at level Debug on the standard logger.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	instance := NewLogger()
	instance.setFields(ctx)
	instance.Zlg.Sugar().Debugf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(ctx context.Context, format string, args ...interface{}) {
	instance := NewLogger()
	instance.setFields(ctx)
	instance.Zlg.Sugar().Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(ctx context.Context, format string, args ...interface{}) {
	instance := NewLogger()
	instance.setFields(ctx)
	instance.Zlg.Sugar().Warnf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	instance := NewLogger()
	instance.setFields(ctx)
	instance.Zlg.Sugar().Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(ctx context.Context, format string, args ...interface{}) {
	instance := NewLogger()
	instance.setFields(ctx)
	instance.Zlg.Sugar().Panicf(format, args...)
}

/*-------------------------------------------------------------------*/
// Debugln logs a message at level Debug on the standard logger.
func Debugln(args ...interface{}) {
	instance := NewLogger()
	instance.Zlg.Sugar().Debugln(args...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(args ...interface{}) {
	instance := NewLogger()
	instance.Zlg.Sugar().Infoln(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(args ...interface{}) {
	instance := NewLogger()
	instance.Zlg.Sugar().Warnln(args...)
}

// Warningln logs a message at level Warn on the standard logger.
func Warningln(args ...interface{}) {
	instance := NewLogger()
	instance.Zlg.Sugar().Warnln(args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(args ...interface{}) {
	instance := NewLogger()
	instance.Zlg.Sugar().Errorln(args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func Fatalln(args ...interface{}) {
	instance := NewLogger()
	instance.Zlg.Sugar().Fatalln(args...)
}

// Panicln logs a message at level Panic on the standard logger.
func Panicln(args ...interface{}) {
	instance := NewLogger()
	instance.Zlg.Sugar().Panicln(args...)
}

func (lgr *Logger) SetOutput(w io.Writer) *Logger {
	lgr.Zlg = zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(lgr.EncoderConfig),
			zapcore.Lock(zapcore.AddSync(w)),
			lgr.Level,
		))
	return lgr
}

func TraceEnabled() bool {
	enabled := os.Getenv("DD_LOG_TRACER_ENABLED")
	return strings.EqualFold(enabled, "true")
}

func GetEnv() string {
	env := os.Getenv("ENV")
	if env != "" {
		return env
	}
	return ENV_DEV
}

func interface2field(v interface{}) (rField []zap.Field) {
	switch mType := v.(type) {
	case []interface{}:
		if len(mType)/2 != 0 {
			// invalid num of params to key value, must be even
			return
		}
		for i := 0; i < len(mType); i += 2 {
			rField = append(rField, zap.Any(fmt.Sprintf("%v", mType[i]), mType[i+1]))
		}
	case map[string]interface{}:
		for key, val := range mType {
			rField = append(rField, zap.Any(key, val))
		}
	}
	return
}
