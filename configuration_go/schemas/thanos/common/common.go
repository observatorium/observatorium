package common

import "time"

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

// Taken from https://github.com/thanos-io/thanos/blob/release-0.32/pkg/model/timeduration.go#L17
type TimeOrDurationValue struct {
	Time *time.Time
	Dur  *time.Duration
}

// String returns either time or duration.
func (tdv *TimeOrDurationValue) String() string {
	switch {
	case tdv.Time != nil:
		return tdv.Time.String()
	case tdv.Dur != nil:
		if v := *tdv.Dur; v < 0 {
			return "-" + (-v).String()
		}
		return tdv.Dur.String()
	}

	return "nil"
}
