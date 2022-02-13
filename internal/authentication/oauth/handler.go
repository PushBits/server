package oauth

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ginserver "github.com/go-oauth2/gin-server"
	"gopkg.in/oauth2.v3"
)

// RevokeAccessRequest holds data required in a revoke request
type RevokeAccessRequest struct {
	Access string `json:"access_token"`
}

type LongtermTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// GetTokenInfo answers with information about an access token
func (a *AuthHandler) GetTokenInfo(c *gin.Context) {
	ti, err := a.tokenFromContext(c)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	tdi := TokenDisplayInfo{}
	tdi.ReadFromTi(ti)

	c.JSON(200, tdi)
}

// RevokeAccess revokes an access token
func (a *AuthHandler) RevokeAccess(c *gin.Context) {
	var request RevokeAccessRequest

	err := c.BindJSON(&request)
	if err != nil || request.Access == "" {
		log.Println("Error when reading request.")
		c.AbortWithError(http.StatusUnprocessableEntity, errors.New("missing access_token"))
		return
	}

	err = a.manager.RemoveAccessToken(request.Access)
	if err != nil {
		log.Println("Error when revoking: ", err)
		c.AbortWithError(http.StatusNotFound, errors.New("unknown access token"))
		return
	}

	c.JSON(200, request)
}

// LongtermToken handles request for longterm access tokens
func (a *AuthHandler) LongtermToken(c *gin.Context) {
	var tokenGenerateRequest oauth2.TokenGenerateRequest
	var request LongtermTokenRequest
	var ltdi LongtermTokenDisplayInfo

	err := c.BindJSON(&request)
	if err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusUnprocessableEntity, errors.New("missing or malformated request"))
		return
	}

	userTi, err := a.tokenFromContext(c)

	if err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	tokenGenerateRequest.UserID = userTi.GetUserID()
	tokenGenerateRequest.Scope = userTi.GetScope()
	tokenGenerateRequest.ClientID = request.ClientID
	tokenGenerateRequest.ClientSecret = request.ClientSecret
	tokenGenerateRequest.AccessTokenExp = time.Hour * 24 * 365 * 5 // 5 years

	ti, err := a.manager.GenerateAccessToken(oauth2.Implicit, &tokenGenerateRequest)
	if err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ltdi.ReadFromTi(ti)

	c.JSON(200, ltdi)
}

func (a *AuthHandler) tokenFromContext(c *gin.Context) (oauth2.TokenInfo, error) {
	err := errors.New("token not found")

	data, exists := c.Get(ginserver.DefaultConfig.TokenKey)
	if !exists {
		log.Println("token does not exist in context.")
		return nil, err
	}

	ti, ok := data.(oauth2.TokenInfo)
	if !ok {
		log.Println("token from context has wrong format.")
		return nil, err
	}
	return ti, nil
}
