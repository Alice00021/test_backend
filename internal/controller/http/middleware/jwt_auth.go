package middleware

import (
	"context"
	"errors"
	httpError "github.com/Alice00021/test_common/pkg/httpserver"
	"github.com/gin-gonic/gin"
	"slices"
	"strconv"
	"strings"
	er "test_go/internal/controller/http/errors"
	auth "test_go/internal/entity"
	"test_go/internal/usecase"
)

var ErrUnauthorized = errors.New("unauthorized")
var ErrNoAccessToken = errors.New("request does not contain an access token")

const userKey string = "x-user"

// IsRoleMiddleware - middleware check role.
func IsRoleMiddleware(role auth.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInfo, err := GetCurrentUser(c)
		if err != nil {
			er.ErrorResponse(c, err)
		}

		if !userInfo.IsEqualRole(role) {
			er.ErrorResponse(c, httpError.NewForbiddenError(auth.ErrAccessDenied))
			return
		}
		c.Next()
		return
	}
}

// IsRolesMiddleware - middleware check roles.
func IsRolesMiddleware(roles ...auth.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInfo, err := GetCurrentUser(c)
		if err != nil {
			er.ErrorResponse(c, err)
		}

		if !slices.Contains(roles, userInfo.Role) {
			er.ErrorResponse(c, httpError.NewForbiddenError(auth.ErrAccessDenied))
			return
		}
		c.Next()
		return
	}
}

// JwtAuthMiddleware - middleware authorization.
func JwtAuthMiddleware(uc usecase.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		t := strings.Split(authHeader, " ")
		if len(t) == 2 {
			authToken := t[1]
			userInfo, err := uc.Validation(context.Background(), authToken)
			if err == nil {
				c.Set(userKey, userInfo)
				c.Next()
				return
			}

			er.ErrorResponse(c, err)
			return
		}
		er.ErrorResponse(c, httpError.NewUnauthorizedError(ErrNoAccessToken))
	}
}

// JwtWsValidator - authorization ws client
func JwtWsValidator(uc usecase.Auth) func(token string) (map[string]*string, error) {
	return func(token string) (map[string]*string, error) {
		t := strings.Split(token, " ")
		if len(t) == 2 {
			authToken := t[1]
			userInfo, err := uc.Validation(context.Background(), authToken)
			if err != nil {
				return nil, err
			}

			userIdToStr := strconv.FormatInt(userInfo.ID, 10)
			userRoleToStr := string(userInfo.Role)

			attrs := make(map[string]*string)
			attrs["userId"] = &userIdToStr
			attrs["userRole"] = &userRoleToStr
			return attrs, nil
		}
		return nil, ErrUnauthorized
	}
}

func GetCurrentUser(c *gin.Context) (*auth.UserInfoToken, error) {
	if user, exists := c.Get(userKey); exists {
		if currentUser, ok := user.(*auth.UserInfoToken); ok {
			return currentUser, nil
		}
	}
	return nil, httpError.NewUnauthorizedError(ErrUnauthorized)
}
