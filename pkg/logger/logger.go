package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// Interface -.
type Interface interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
}

// Logger -.
type Logger struct {
	logger *zerolog.Logger
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
	logger := zerolog.New(os.Stdout).With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).Logger()

	return &Logger{
		logger: &logger,
	}
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

// zap logger for logging
/*
package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"s
	"go.uber.org/zap/zapcore"
)

// Interface defines the logger interface.
type Interface interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
}

// Logger is a wrapper around zap.Logger.
type Logger struct {
	logger *zap.SugaredLogger
}

var _ Interface = (*Logger)(nil)

// New creates a new logger instance.
func New(level string) *Logger {
	zapLogger := newZapLogger(level)
	return &Logger{
		logger: zapLogger.Sugar(),
	}
}

func newZapLogger(level string) *zap.Logger {
	// Set the log level
	atom := zap.NewAtomicLevel()

	switch strings.ToLower(level) {
	case "debug":
		atom.SetLevel(zapcore.DebugLevel)
	case "info":
		atom.SetLevel(zapcore.InfoLevel)
	case "warn":
		atom.SetLevel(zapcore.WarnLevel)
	case "error":
		atom.SetLevel(zapcore.ErrorLevel)
	default:
		atom.SetLevel(zapcore.InfoLevel)
	}

	// Encoder config (JSON logs for prod, Console for dev)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg), // Change to zapcore.NewConsoleEncoder for dev
		zapcore.Lock(os.Stdout),            // Output
		atom,                               // Log level
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

// Debug logs a debug message.
func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.msg("debug", message, args...)
}

// Info logs an info message.
func (l *Logger) Info(message string, args ...interface{}) {
	if len(args) == 0 {
		l.logger.Info(message)
	} else {
		l.logger.Infof(message, args...)
	}
}

// Warn logs a warning message.
func (l *Logger) Warn(message string, args ...interface{}) {
	if len(args) == 0 {
		l.logger.Warn(message)
	} else {
		l.logger.Warnf(message, args...)
	}
}

// Error logs an error message.
func (l *Logger) Error(message interface{}, args ...interface{}) {
	l.msg("error", message, args...)
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	l.msg("fatal", message, args...)
	l.logger.Sync() // flush logs before exit
	os.Exit(1)
}

func (l *Logger) msg(level string, message interface{}, args ...interface{}) {
	switch msg := message.(type) {
	case error:
		l.log(level, msg.Error(), args...)
	case string:
		l.log(level, msg, args...)
	default:
		l.log(level, fmt.Sprintf("%s message %v has unknown type %T", level, message, message), args...)
	}
}

func (l *Logger) log(level string, msg string, args ...interface{}) {
	switch level {
	case "debug":
		if len(args) == 0 {
			l.logger.Debug(msg)
		} else {
			l.logger.Debugf(msg, args...)
		}
	case "info":
		if len(args) == 0 {
			l.logger.Info(msg)
		} else {
			l.logger.Infof(msg, args...)
		}
	case "warn":
		if len(args) == 0 {
			l.logger.Warn(msg)
		} else {
			l.logger.Warnf(msg, args...)
		}
	case "error":
		if len(args) == 0 {
			l.logger.Error(msg)
		} else {
			l.logger.Errorf(msg, args...)
		}
	case "fatal":
		if len(args) == 0 {
			l.logger.Fatal(msg)
		} else {
			l.logger.Fatalf(msg, args...)
		}
	default:
		l.logger.Info(msg)
	}
}
*/
