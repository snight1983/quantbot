package utils

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger *zap.Logger
)

func InitLog(level, path, name, ver, logf string) {
	path = filepath.Join(path, logf)
	hook := lumberjack.Logger{
		Filename:   path,
		MaxSize:    128,
		MaxBackups: 30,
		MaxAge:     7,
		LocalTime:  true,
		Compress:   true,
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "linenum",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	atomicLevel := zap.NewAtomicLevel()
	switch level {
	case "debug":
		atomicLevel.SetLevel(zap.DebugLevel)
	case "error":
		atomicLevel.SetLevel(zap.ErrorLevel)
	default:
		atomicLevel.SetLevel(zap.InfoLevel)
	}

	eocoder := zapcore.NewConsoleEncoder(encoderConfig)
	core := zapcore.NewCore(
		eocoder,
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(&hook)),
		atomicLevel)

	caller := zap.AddCaller()
	development := zap.Development()

	filed := zap.Fields(zap.String("svr", name),
		zap.String("svrver", ver))
	Logger = zap.New(core, caller, development, filed)
}
