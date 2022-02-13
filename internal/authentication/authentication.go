package authentication

import (
	"errors"
	"log"
	"net/http"

	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/model"

	"github.com/gin-gonic/gin"
)

const (
	headerName = "X-Gotify-Key"
)

type (
	// AuthenticationValidator defines a type for authenticating a user
	AuthenticationValidator func() gin.HandlerFunc
	// UserSetter defines a type for setting a user object
	UserSetter func() gin.HandlerFunc
)

// AuthHandler defines the minimal interface for an auth handler
type AuthHandler interface {
	AuthenticationValidator() gin.HandlerFunc
	UserSetter() gin.HandlerFunc
}

// The Database interface for encapsulating database access.
type Database interface {
	GetApplicationByToken(token string) (*model.Application, error)
	GetUserByName(name string) (*model.User, error)
	GetUserByID(id uint) (*model.User, error)
}

// Authenticator is the provider for authentication middleware.
type Authenticator struct {
	DB                      Database
	Config                  configuration.Authentication
	AuthenticationValidator AuthenticationValidator
	UserSetter              UserSetter
}

type hasUserProperty func(user *model.User) bool

func (a *Authenticator) requireUserProperty(has hasUserProperty, errorMessage string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := errors.New("user not found")

		u, exists := ctx.Get("user")

		if !exists {
			log.Println("no user object in context")
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		user, ok := u.(*model.User)

		if !ok {
			log.Println("user object from context has wrong format")
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		if !has(user) {
			ctx.AbortWithError(http.StatusForbidden, errors.New(errorMessage))
			return
		}
	}
}

// RequireUser returns a Gin middleware which requires valid user credentials to be supplied with the request.
func (a *Authenticator) RequireUser() []gin.HandlerFunc {
	funcs := make([]gin.HandlerFunc, 0)
	funcs = append(funcs, a.RequireValidAuthentication())
	funcs = append(funcs, a.UserSetter())
	return funcs
}

// RequireAdmin returns a Gin middleware which requires valid admin credentials to be supplied with the request.
func (a *Authenticator) RequireAdmin() []gin.HandlerFunc {
	funcs := make([]gin.HandlerFunc, 0)
	funcs = append(funcs, a.RequireValidAuthentication())
	funcs = append(funcs, a.UserSetter())
	funcs = append(funcs, a.requireUserProperty(func(user *model.User) bool {
		return user.IsAdmin
	}, "user does not have permission: admin"))

	return funcs
}

func (a *Authenticator) tokenFromQueryOrHeader(ctx *gin.Context) string {
	if token := a.tokenFromQuery(ctx); token != "" {
		return token
	} else if token := a.tokenFromHeader(ctx); token != "" {
		return token
	}

	return ""
}

func (a *Authenticator) tokenFromQuery(ctx *gin.Context) string {
	return ctx.Request.URL.Query().Get("token")
}

func (a *Authenticator) tokenFromHeader(ctx *gin.Context) string {
	return ctx.Request.Header.Get(headerName)
}

// RequireApplicationToken returns a Gin middleware which requires an application token to be supplied with the request.
func (a *Authenticator) RequireApplicationToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := a.tokenFromQueryOrHeader(ctx)

		app, err := a.DB.GetApplicationByToken(token)
		if err != nil {
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		ctx.Set("app", app)
	}
}

// RequireValidAuthentication returns a Gin middleware which requires a valid authentication
func (a *Authenticator) RequireValidAuthentication() gin.HandlerFunc {
	return a.AuthenticationValidator()
}

// RegisterHandler registers an authentication handler
func (a *Authenticator) RegisterHandler(handler AuthHandler) {
	a.UserSetter = handler.UserSetter
	a.AuthenticationValidator = handler.AuthenticationValidator
}
