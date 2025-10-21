package utils

import "context"

const (
	LoggerKey   = "logger"
	LoggerValue = "true"
)

func AddListenerPropertyToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, LoggerKey, LoggerValue)
}

func CheckListenerPropertyFromContext(ctx context.Context) bool {
	value := ctx.Value(LoggerKey)
	if value != nil {
		return true
	}

	return false
}

func CheckListenerPropertyFromHeaders(h map[string]interface{}) bool {
	for k, v := range h {
		if k == LoggerKey && v.(string) == LoggerValue {
			return true
		}
	}

	return false
}
