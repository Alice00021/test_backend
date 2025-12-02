package v1

import (
	httpError "github.com/Alice00021/test_common/pkg/httpserver"
	"github.com/Alice00021/test_common/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"test_go/internal/controller/http/errors"
	"test_go/internal/controller/http/v1/request"
	"test_go/internal/usecase"
)

type operationRoutes struct {
	l  logger.Interface
	uc usecase.OperationMongo
}

func NewOperationRoutes(privateGroup *gin.RouterGroup, l logger.Interface, uc usecase.OperationMongo) {
	r := &operationRoutes{l, uc}
	{
		h := privateGroup.Group("/operation")
		h.GET("", r.getOperations)
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
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
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
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
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

func (r *operationRoutes) getOperations(c *gin.Context) {
	res, err := r.uc.GetOperations(c.Request.Context())
	if err != nil {
		r.l.Error(err, "http - v1 - getOperations")
		errors.ErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}
