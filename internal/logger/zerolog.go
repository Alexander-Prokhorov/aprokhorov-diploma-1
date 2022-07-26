package logger

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

type ZeroLogger struct {
	Logger zerolog.Logger
}

func NewZeroLogger(level string) (*ZeroLogger, error) {
	out := zerolog.NewConsoleWriter()
	logger := zerolog.New(out)
	zeroLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return &ZeroLogger{}, err
	}
	return &ZeroLogger{Logger: logger.Level(zeroLevel)}, nil
}

func (z *ZeroLogger) Debug(parent string, msg string) {
	z.Logger.Debug().Timestamp().Msg(fmt.Sprintf("\x1b[33m%s\x1b[0m: %s", strings.ToUpper(parent), msg))
}

func (z *ZeroLogger) Info(parent string, msg string) {
	z.Logger.Info().Timestamp().Msg(fmt.Sprintf("\x1b[33m%s\x1b[0m: %s", strings.ToUpper(parent), msg))
}

func (z *ZeroLogger) Warning(parent string, msg string) {
	z.Logger.Warn().Timestamp().Msg(fmt.Sprintf("\x1b[33m%s\x1b[0m: %s", strings.ToUpper(parent), msg))
}

func (z *ZeroLogger) Error(parent string, msg string) {
	z.Logger.Error().Timestamp().Msg(fmt.Sprintf("\x1b[33m%s\x1b[0m: %s", strings.ToUpper(parent), msg))
}

func (z *ZeroLogger) Fatal(parent string, msg string) {
	z.Logger.Fatal().Timestamp().Msg(fmt.Sprintf("\x1b[33m%s\x1b[0m: %s", strings.ToUpper(parent), msg))
}

func (z *ZeroLogger) Panic(parent string, msg string) {
	z.Logger.Panic().Timestamp().Msg(fmt.Sprintf("\x1b[33m%s\x1b[0m: %s", strings.ToUpper(parent), msg))
}
