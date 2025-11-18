package logger

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"os"
	"strings"
)

// Interface -.
type Interface interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
}

const (
	_defaultMaxFileSize = 5  // 5Mb
	_defaultMaxFileAge  = 30 // 30 days
	_defaultFileName    = "app.log"
	_defaultLevel       = "info"
)

// Logger -.
type Logger struct {
	logger      *zerolog.Logger
	Level       string
	MaxFileAge  int
	MaxFileSize int
	FileName    string
}

var _ Interface = (*Logger)(nil)

// New -.
func New(level string) *Logger {
	var l zerolog.Level

	switch strings.ToLower(level) {
	case "error":
		l = zerolog.ErrorLevel
	case "warn":
		l = zerolog.WarnLevel
	case "info":
		l = zerolog.InfoLevel
	case "debug":
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)

	skipFrameCount := 3
	logger := zerolog.New(os.Stdout).With().Timestamp().
		CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).Logger()

	return &Logger{
		logger: &logger,
	}
}

func NewMultipleWriter(opts ...Option) *Logger {
	l := &Logger{
		MaxFileAge:  _defaultMaxFileAge,
		MaxFileSize: _defaultMaxFileSize,
		FileName:    _defaultFileName,
		Level:       _defaultLevel,
	}

	for _, opt := range opts {
		opt(l)
	}

	var level zerolog.Level

	switch strings.ToLower(l.Level) {
	case "error":
		level = zerolog.ErrorLevel
	case "warn":
		level = zerolog.WarnLevel
	case "info":
		level = zerolog.InfoLevel
	case "debug":
		level = zerolog.DebugLevel
	default:
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	lumberjackLogger := &lumberjack.Logger{
		Filename: l.FileName,
		MaxSize:  l.MaxFileSize,
		MaxAge:   l.MaxFileAge,
		Compress: true,
	}

	levelWriter := zerolog.MultiLevelWriter(os.Stdout, lumberjackLogger)
	skipFrameCount := 3
	zeroLogger := zerolog.New(levelWriter).With().Timestamp().
		CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).Logger()

	l.logger = &zeroLogger

	return l
}

// Debug -.
func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.msg("debug", message, args...)
}

// Info -.
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(message, args...)
}

// Warn -.
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(message, args...)
}

// Error -.
func (l *Logger) Error(message interface{}, args ...interface{}) {
	if l.logger.GetLevel() == zerolog.DebugLevel {
		l.Debug(message, args...)
	}

	l.msg("error", message, args...)
}

// Fatal -.
func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	l.msg("fatal", message, args...)

	os.Exit(1)
}

func (l *Logger) log(message string, args ...interface{}) {
	if len(args) == 0 {
		l.logger.Info().Msg(message)
	} else {
		l.logger.Info().Msgf(message, args...)
	}
}

func (l *Logger) msg(level string, message interface{}, args ...interface{}) {
	switch msg := message.(type) {
	case error:
		l.log(msg.Error(), args...)
	case string:
		l.log(msg, args...)
	default:
		l.log(fmt.Sprintf("%s message %v has unknown type %v", level, message, msg), args...)
	}
}
