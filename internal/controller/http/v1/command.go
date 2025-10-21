package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"test_go/internal/controller/http/errors"
	"test_go/internal/usecase"
	"test_go/pkg/logger"
)

type commandRoutes struct {
	l  logger.Interface
	uc usecase.Command
}

func NewCommandRoutes(privateGroup *gin.RouterGroup, l logger.Interface, uc usecase.Command) {
	r := &commandRoutes{l, uc}
	{
		h := privateGroup.Group("/commands")
		h.POST("", r.updateCommands)
	}
}

func (r *commandRoutes) updateCommands(c *gin.Context) {

	if err := r.uc.UpdateCommands(c.Request.Context()); err != nil {
		r.l.Error(err, "http - v1 - updateCommands")
		errors.ErrorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}
