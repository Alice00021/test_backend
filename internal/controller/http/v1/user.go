package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"test_go/internal/controller/http/errors"
	"test_go/internal/controller/http/middleware"
	"test_go/internal/controller/http/v1/request"
	"test_go/internal/usecase"
	httpError "test_go/pkg/httpserver"
	"test_go/pkg/logger"
)

type userRoutes struct {
	l  logger.Interface
	uc usecase.User
}

func NewUserRoutes(privateGroup *gin.RouterGroup, l logger.Interface, uc usecase.User) {
	r := &userRoutes{l, uc}
	{
		h := privateGroup.Group("/users")
		h.GET("/profile", r.getProfile)
		h.PATCH("/change-password", r.changePassword)
		h.PUT("/photo", r.setProfilePhoto)
		h.PATCH("/rating", r.updateRating)
		h.PATCH("/concurrent-test", r.simulateConcurrentUpdates)
	}
}

func (r *userRoutes) getProfile(c *gin.Context) {

	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		errors.ErrorResponse(c, err)
		return
	}

	res, err := r.uc.GetUser(c.Request.Context(), currentUser.ID)
	if err != nil {
		r.l.Error(err, "http - v1 - getProfile")
		errors.ErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (r *userRoutes) changePassword(c *gin.Context) {
	var req request.ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		r.l.Error(err, "http - v1 - changePassword")
		errors.ErrorResponse(c, httpError.NewBadRequestBodyError(err))
		return
	}
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		errors.ErrorResponse(c, err)
		return
	}

	inp := req.ToEntity()
	inp.ID = currentUser.ID

	if err := r.uc.ChangePassword(c.Request.Context(), inp); err != nil {
		r.l.Error(err, "http - v1 - changePassword")
		errors.ErrorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (r *userRoutes) setProfilePhoto(c *gin.Context) {
	var req request.UploadFileRequest
	if err := c.ShouldBind(&req); err != nil {
		r.l.Error(err, "http - v1 - setProfilePhoto")
		errors.ErrorResponse(c, httpError.NewBadRequestBodyError(err))
		return
	}

	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		errors.ErrorResponse(c, err)
		return
	}

	if err := r.uc.SetProfilePhoto(c.Request.Context(), currentUser.ID, req.File); err != nil {
		r.l.Error(err, "http - v1 - setProfilePhoto")
		errors.ErrorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (r *userRoutes) updateRating(c *gin.Context) {
	var req request.UpdateRatingRequest
	if err := c.ShouldBind(&req); err != nil {
		r.l.Error(err, "http - v1 - updateRating")
		errors.ErrorResponse(c, httpError.NewBadRequestBodyError(err))
		return
	}

	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		errors.ErrorResponse(c, err)
		return
	}

	if err := r.uc.UpdateRating(c.Request.Context(), currentUser.ID, req.Rating); err != nil {
		r.l.Error(err, "http - v1 - updateRating")
		errors.ErrorResponse(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (r *userRoutes) simulateConcurrentUpdates(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		errors.ErrorResponse(c, err)
		return
	}

	numGoroutines := 2
	deltas := []float32{15.0, 30.0}

	var wg sync.WaitGroup
	ctx := c.Request.Context()

	for i := 0; i < numGoroutines; i++ {
		delta := deltas[i]
		wg.Add(1)
		go func(d float32) {
			defer wg.Done()
			if err := r.uc.UpdateRating(ctx, currentUser.ID, d); err != nil {
				r.l.Error(err, "http - v1 - simulateConcurrentUpdates - goroutine error")
			}
		}(delta)
	}
	wg.Wait()

	user, err := r.uc.GetUser(ctx, currentUser.ID)
	if err != nil {
		r.l.Error(err, "http - v1 - simulateConcurrentUpdates - get user after update")
		errors.ErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}
