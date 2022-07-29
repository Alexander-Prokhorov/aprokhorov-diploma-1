package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type ZeroLogger struct {
	Logger zerolog.Logger
}

func NewZeroLogger(level string) (*ZeroLogger, error) {
	logger := zerolog.New(os.Stdin)
	zeroLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return &ZeroLogger{}, err
	}
	return &ZeroLogger{Logger: logger.Level(zeroLevel)}, nil
}

func (z *ZeroLogger) Debug(parent string, msg string) {
	z.Logger.Debug().Timestamp().Msg(fmt.Sprintf("%s: %s", strings.ToUpper(parent), msg))
}

func (z *ZeroLogger) Info(parent string, msg string) {
	z.Logger.Info().Timestamp().Msg(fmt.Sprintf("%s: %s", strings.ToUpper(parent), msg))
}

func (z *ZeroLogger) Warning(parent string, msg string) {
	z.Logger.Warn().Timestamp().Msg(fmt.Sprintf("%s: %s", strings.ToUpper(parent), msg))
}

func (z *ZeroLogger) Error(parent string, msg string) {
	z.Logger.Error().Timestamp().Msg(fmt.Sprintf("%s: %s", strings.ToUpper(parent), msg))
}

func (z *ZeroLogger) Fatal(parent string, msg string) {
	z.Logger.Fatal().Timestamp().Msg(fmt.Sprintf("%s: %s", strings.ToUpper(parent), msg))
}

func (z *ZeroLogger) Panic(parent string, msg string) {
	z.Logger.Panic().Timestamp().Msg(fmt.Sprintf("%s: %s", strings.ToUpper(parent), msg))
}
