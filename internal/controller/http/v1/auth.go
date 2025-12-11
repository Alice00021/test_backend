package v1

import (
	"github.com/gin-gonic/gin"

	httpError "github.com/Alice00021/test_common/pkg/httpserver"
	"github.com/Alice00021/test_common/pkg/logger"
	"net/http"
	"test_go/internal/controller/http/errors"
	"test_go/internal/controller/http/v1/request"
	"test_go/internal/usecase"
)

type authRoutes struct {
	l  logger.Interface
	uc usecase.Auth
}

func NewAuthRoutes(publicGroup *gin.RouterGroup, l logger.Interface, uc usecase.Auth) {
	r := &authRoutes{l, uc}
	{
		h := publicGroup.Group("/auth")
		h.POST("/register", r.register)
		h.POST("/login", r.login)
		h.GET("/verify", r.verifyEmail)
	}
}

func (r *authRoutes) register(c *gin.Context) {
	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.l.Error(err, "http - v1 - register")
		errors.ErrorResponse(c, httpError.NewBadRequestBodyError(err))
		return
	}
	inp := req.ToEntity()

	res, err := r.uc.Register(c.Request.Context(), inp)
	if err != nil {
		r.l.Error(err, "http - v1 - register")
		errors.ErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (r *authRoutes) login(c *gin.Context) {
	var req request.AuthenticateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.l.Error(err, "http - v1 - login")
		errors.ErrorResponse(c, httpError.NewBadRequestBodyError(err))
		return
	}

	res, err := r.uc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		r.l.Error(err, "http - v1 - login")
		errors.ErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (r *authRoutes) verifyEmail(c *gin.Context) {
	var req request.VerifyEmailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		r.l.Error(err, "http - v1 - verifyEmail")
		errors.ErrorResponse(c, httpError.NewBadQueryParamsError(err))
		return
	}

	if err := r.uc.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		r.l.Error(err, "http - v1 - verifyEmail")
		errors.ErrorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}
