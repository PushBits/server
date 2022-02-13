package oauth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	ginserver "github.com/go-oauth2/gin-server"
	mysql "github.com/imrenagi/go-oauth2-mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/model"

	"gopkg.in/oauth2.v3"
	oauth_error "gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/generates"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

// The Database interface for encapsulating database access.
type Database interface {
	GetApplicationByToken(token string) (*model.Application, error)
	GetUserByName(name string) (*model.User, error)
	GetUserByID(id uint) (*model.User, error)
}

// AuthHandler is the oauth provider for authentication
type AuthHandler struct {
	db      Database
	manager *manage.Manager
	config  configuration.Authentication
}

// Initialize prepares the AuthHandler
func (a *AuthHandler) Initialize(db *database.Database, configAuth configuration.Authentication, configDatabase configuration.Database) error {
	a.db = db
	a.config = configAuth

	if len(a.config.Oauth.TokenKey) < 5 {
		panic("Your Oauth 2.0 token key is empty or not long enough to be secure. Please change it in the configuration file.")
	}

	// The manager handles the tokens
	a.manager = manage.NewDefaultManager()
	a.manager.SetAuthorizeCodeExp(time.Duration(24) * time.Hour)
	a.manager.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    time.Duration(24) * time.Hour,      // 1 day
		RefreshTokenExp:   time.Duration(24) * time.Hour * 30, // 30 days
		IsGenerateRefresh: true,
	})
	a.manager.SetRefreshTokenCfg(&manage.RefreshingConfig{
		AccessTokenExp:     time.Duration(24) * time.Hour,      // 1 day
		RefreshTokenExp:    time.Duration(24) * time.Hour * 30, // 30 days
		IsGenerateRefresh:  true,
		IsResetRefreshTime: true,
		IsRemoveAccess:     false,
		IsRemoveRefreshing: true,
	})
	a.manager.MapAccessGenerate(generates.NewJWTAccessGenerate([]byte(a.config.Oauth.TokenKey), jwt.SigningMethodHS512)) // unfortunately only symmetric algorithms seem to be supported

	// Define a storage for the tokens
	switch configDatabase.Dialect {
	case "mysql":
		dbOauth := sqlx.NewDb(db.GetSqldb(), "mysql")
		// Workaround to get rid of clients that still exist in the database but no longer in the config file
		_, err := dbOauth.Exec("DROP TABLE IF EXISTS oauth_clients")
		if err != nil {
			panic(err)
		}

		a.manager.MustTokenStorage(mysql.NewTokenStore(dbOauth))

		clientStore, _ := mysql.NewClientStore(dbOauth, mysql.WithClientStoreTableName("oauth_clients"))
		a.manager.MapClientStorage(clientStore)

		for _, client := range a.config.Oauth.Clients {
			if len(client.ClientSecret) < 5 {
				panic("Your Oauth 2.0 client secret is empty or not long enough to be secure. Please change it in the configuration file.")
			}

			clientStore.Create(&models.Client{
				ID:     client.ClientID,
				Secret: client.ClientSecret,
				Domain: client.ClientRedirect,
			})
		}

	case "sqlite3":
		a.manager.MustTokenStorage(store.NewFileTokenStore("pushbits_tokens.db"))
		clientStore := store.NewClientStore() // memory store
		a.manager.MapClientStorage(clientStore)

		for _, client := range a.config.Oauth.Clients {
			if len(client.ClientSecret) < 5 {
				panic("Your Oauth 2.0 client secret is empty or not long enough to be secure. Please change it in the configuration file.")
			}

			clientStore.Set(client.ClientID, &models.Client{
				ID:     client.ClientID,
				Secret: client.ClientSecret,
				Domain: client.ClientRedirect,
			})
		}

	default:
		log.Panicln("Unknown (oauth) storage dialect")
	}
	// Initialize and configure the token server
	ginserver.InitServer(a.manager)
	ginserver.SetAllowGetAccessRequest(true)
	ginserver.SetClientInfoHandler(server.ClientFormHandler)
	ginserver.SetUserAuthorizationHandler(a.UserAuthHandler())
	ginserver.SetPasswordAuthorizationHandler(a.passwordAuthorizationHandler())
	ginserver.SetInternalErrorHandler(a.InternalErrorHandler())
	ginserver.SetAllowedGrantType(
		oauth2.AuthorizationCode,
		//oauth2.PasswordCredentials,
		oauth2.Refreshing,
	)
	ginserver.SetAllowedResponseType(
		oauth2.Code,
	)
	ginserver.SetClientScopeHandler(ClientScopeHandler())
	ginserver.SetAccessTokenExpHandler(AccessTokenExpHandler())
	return nil
}

// PasswordAuthorizationHandler returns a PasswordAuthorizationHandler that handles username and password based authentication for access tokens
func (a *AuthHandler) passwordAuthorizationHandler() server.PasswordAuthorizationHandler {
	return func(username string, password string) (string, error) {
		log.Println("Received password based authentication request")

		user, err := a.db.GetUserByName(username)

		if err != nil || user == nil {
			return "", nil
		}

		if !credentials.ComparePassword(user.PasswordHash, []byte(password)) {
			return "", nil
		}

		return fmt.Sprintf("%d", user.ID), nil
	}
}

// UserAuthHandler extracts user information from an auth request
func (a *AuthHandler) UserAuthHandler() server.UserAuthorizationHandler {
	return func(w http.ResponseWriter, r *http.Request) (string, error) {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if user, err := a.db.GetUserByName(username); err != nil {
			return "", err
		} else if user != nil && credentials.ComparePassword(user.PasswordHash, []byte(password)) {
			return fmt.Sprint(user.ID), nil
		}
		return "", errors.New("no credentials provided")
	}
}

// ClientScopeHandler returns a ClientScopeHandler that allows or disallows scopes for access tokens
func ClientScopeHandler() server.ClientScopeHandler {
	return func(clientID, scope string) (allowed bool, err error) {
		if scope == "all" || scope == "" { // For now only allow generic scopes so there is place for future expansion
			return true, nil
		}

		return false, nil
	}
}

// AccessTokenExpHandler returns an AccessTokenExpHandler that sets the expiration time of access tokens
func AccessTokenExpHandler() server.AccessTokenExpHandler {
	return func(w http.ResponseWriter, r *http.Request) (exp time.Duration, err error) {
		return time.Duration(24) * time.Hour, nil
	}
}

// InternalErrorHandler handles errors for authentication, it will always return a server_error
func (a *AuthHandler) InternalErrorHandler() server.InternalErrorHandler {
	return func(err error) *oauth_error.Response {
		var re oauth_error.Response
		log.Println(err)

		re.Error = oauth_error.ErrServerError
		re.Description = oauth_error.Descriptions[oauth_error.ErrServerError]
		re.StatusCode = oauth_error.StatusCodes[oauth_error.ErrServerError]
		return &re
	}
}
