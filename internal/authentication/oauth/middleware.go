package oauth

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	ginserver "github.com/go-oauth2/gin-server"
	"gopkg.in/oauth2.v3"
)

// AuthenticationValidator returns a gin middleware for authenticating users based on a oauth access token
func (a AuthHandler) AuthenticationValidator() gin.HandlerFunc {
	return ginserver.HandleTokenVerify()
}

// UserSetter returns a gin HandlerFunc that takes the token from the AuthenticationValidator and sets the corresponding user object
func (a AuthHandler) UserSetter() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error
		ti, exists := ctx.Get(ginserver.DefaultConfig.TokenKey)
		if !exists {
			err = errors.New("no token available")
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		token, ok := ti.(oauth2.TokenInfo)
		if !ok {
			err = errors.New("wrong token format")
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		userID, err := strconv.ParseUint(token.GetUserID(), 10, 64)
		if err != nil {
			err = errors.New("user information of wrong format")
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		user, err := a.db.GetUserByID(uint(userID))
		if err != nil {
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		ctx.Set("user", user)
	}
}
