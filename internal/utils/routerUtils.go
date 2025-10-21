package utils

import (
	"github.com/google/uuid"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ParseParams struct {
	Context *gin.Context
	Key     string
	Default interface{}
}

func ParseUint64(param string) (uint64, error) {
	return strconv.ParseUint(param, 10, 64)
}

func ParseUUID(param string) (uuid.UUID, error) {
	return uuid.Parse(param)
}

func ParseInt64(param string) (int64, error) {
	return strconv.ParseInt(param, 10, 64)
}

func ParsePathParam[T any](params ParseParams, parseFunc func(string) (T, error)) (T, error) {
	return ParseParam(
		func() string { return params.Context.Param(params.Key) },
		GetDefault[T](params.Default),
		parseFunc,
	)
}

func ParseQueryParam[T any](params ParseParams, parseFunc func(string) (T, error)) (T, error) {
	return ParseParam(
		func() string { return params.Context.Query(params.Key) },
		GetDefault[T](params.Default),
		parseFunc,
	)
}

func ParseParam[T any](getParamFunc func() string, defaultValue *T, parseFunc func(string) (T, error)) (T, error) {
	param := getParamFunc()

	if param == "" {
		if defaultValue != nil {
			return *defaultValue, nil
		}
		var zeroValue T
		return zeroValue, nil
	}

	parsedValue, err := parseFunc(param)
	if err != nil {
		if defaultValue != nil {
			return *defaultValue, err
		}
		var zeroValue T
		return zeroValue, err
	}

	return parsedValue, nil
}

func GetDefault[T any](def interface{}) *T {
	if def != nil {
		if v, ok := def.(T); ok {
			return &v
		}
	}
	return nil
}
