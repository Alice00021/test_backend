package errors

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"test_go/internal/entity"
	httpError "test_go/pkg/httpserver"
)

func ErrorResponse(c *gin.Context, err error) {
	var httpErr httpError.HttpError
	if errors.As(err, &httpErr) {
		c.AbortWithStatusJSON(httpErr.Status, httpErr)
		return
	}

	if errors.Is(err, entity.ErrAccessDenied) {
		httpErr = httpError.NewForbiddenError(err.Error())
		c.AbortWithStatusJSON(httpErr.Status, httpErr)
		return
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, httpError.NewInternalServerError(err))
}
