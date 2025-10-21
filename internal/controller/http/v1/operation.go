package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"test_go/internal/controller/http/errors"
	"test_go/internal/controller/http/v1/request"
	"test_go/internal/usecase"
	"test_go/internal/utils"
	httpError "test_go/pkg/httpserver"
	"test_go/pkg/logger"
)

type operationRoutes struct {
	l  logger.Interface
	uc usecase.Operation
}

func NewOperationRoutes(privateGroup *gin.RouterGroup, l logger.Interface, uc usecase.Operation) {
	r := &operationRoutes{l, uc}
	{
		h := privateGroup.Group("/operation")
		h.POST("", r.createOperation)
		h.PUT("/:id", r.updateOperation)
		h.DELETE("/:id", r.deleteOperation)
	}
}

func (r *operationRoutes) createOperation(c *gin.Context) {
	var req request.CreateOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.l.Error(err, "http - v1 - createOperation")
		errors.ErrorResponse(c, httpError.NewBadRequestBodyError(err))
		return
	}
	res, err := r.uc.CreateOperation(c.Request.Context(), req.ToEntity())
	if err != nil {
		r.l.Error(err, "http - v1 - createOperation")
		errors.ErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (r *operationRoutes) updateOperation(c *gin.Context) {
	id, err := utils.ParsePathParam(utils.ParseParams{Context: c, Key: "id"}, utils.ParseInt64)
	if err != nil {
		r.l.Error(err, "http - v1 - updateOperation")
		errors.ErrorResponse(c, httpError.NewBadPathParamsError(err))
		return
	}

	var req request.UpdateOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.l.Error(err, "http - v1 - updateOperation")
		errors.ErrorResponse(c, httpError.NewBadRequestBodyError(err))
		return
	}
	inp := req.ToEntity()
	inp.ID = id

	if err = r.uc.UpdateOperation(c.Request.Context(), inp); err != nil {
		r.l.Error(err, "http - v1 - updateOperation")
		errors.ErrorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (r *operationRoutes) deleteOperation(c *gin.Context) {
	id, err := utils.ParsePathParam(utils.ParseParams{Context: c, Key: "id"}, utils.ParseInt64)
	if err != nil {
		r.l.Error(err, "http - v1 - deleteOperation")
		errors.ErrorResponse(c, httpError.NewBadPathParamsError(err))
		return
	}

	err = r.uc.DeleteOperation(c.Request.Context(), id)
	if err != nil {
		r.l.Error(err, "http - v1 - deleteOperation")
		errors.ErrorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}
