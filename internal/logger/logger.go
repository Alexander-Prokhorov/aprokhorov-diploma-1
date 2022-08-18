package logger

type Logger interface {
	Debug(parent string, message string)
	Info(parent string, message string)
	Warning(parent string, message string)
	Error(parent string, message string)
	Fatal(parent string, message string)
	Panic(parent string, message string)
}
