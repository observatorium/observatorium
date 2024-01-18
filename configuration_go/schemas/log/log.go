package log

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type LogFormat string

const (
	LogFormatLogfmt LogFormat = "logfmt"
	LogFormatJSON   LogFormat = "json"
)
