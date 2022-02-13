package basicauth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/model"
)

// The Database interface for encapsulating database access.
type Database interface {
	GetApplicationByToken(token string) (*model.Application, error)
	GetUserByName(name string) (*model.User, error)
}

// AuthHandler is the basic auth provider for authentication
type AuthHandler struct {
	db Database
}

// Initialize prepares the AuthHandler
func (a *AuthHandler) Initialize(db *database.Database) error {
	a.db = db
	return nil
}

// AuthenticationValidator returns a gin HandlerFunc that takes the basic auth credentials and validates them
func (a AuthHandler) AuthenticationValidator() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var user *model.User
		user, err := a.userFromBasicAuth(ctx)

		if err != nil {
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		if user == nil {
			ctx.AbortWithError(http.StatusForbidden, errors.New("authentication failed"))
			return
		}
	}
}

// UserSetter returns a gin HandlerFunc that takes the basic auth credentials and sets the corresponding user object
func (a AuthHandler) UserSetter() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var user *model.User
		user, err := a.userFromBasicAuth(ctx)

		if err != nil {
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		if user == nil {
			ctx.AbortWithError(http.StatusForbidden, errors.New("authentication failed"))
			return
		}

		ctx.Set("user", user)
	}
}

func (a *AuthHandler) userFromBasicAuth(ctx *gin.Context) (*model.User, error) {
	if name, password, ok := ctx.Request.BasicAuth(); ok {
		if user, err := a.db.GetUserByName(name); err != nil {
			return nil, err
		} else if user != nil && credentials.ComparePassword(user.PasswordHash, []byte(password)) {
			return user, nil
		} else {
			return nil, errors.New("credentials were invalid")
		}
	}

	return nil, errors.New("no credentials were supplied")
}
