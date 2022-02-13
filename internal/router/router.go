package router

import (
	"log"

	"github.com/pushbits/server/internal/api"
	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/authentication/basicauth"
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/authentication/oauth"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/dispatcher"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	ginserver "github.com/go-oauth2/gin-server"
)

// Create a Gin engine and setup all routes.
func Create(debug bool, cm *credentials.Manager, db *database.Database, dp *dispatcher.Dispatcher, config *configuration.Configuration) *gin.Engine {
	log.Println("Setting up HTTP routes.")

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize gin
	r := gin.Default()
	r.Use(location.Default())

	// Set up authentication and handler
	auth := authentication.Authenticator{
		DB:     db,
		Config: config.Authentication,
	}

	switch config.Authentication.Method {
	case "oauth":
		authHandler := oauth.AuthHandler{}
		authHandler.Initialize(db, config.Authentication, config.Database)
		auth.RegisterHandler(authHandler)

		// Register oauth endpoints
		oauthGroup := r.Group("/oauth2")
		{
			oauthGroup.POST("/token", ginserver.HandleTokenRequest)
			oauthGroup.POST("/auth", ginserver.HandleAuthorizeRequest)
			oauthGroup.GET("/tokeninfo", auth.RequireValidAuthentication(), authHandler.GetTokenInfo)
			oauthGroup.POST("/revoke", append(auth.RequireAdmin(), authHandler.RevokeAccess)...)
			oauthGroup.POST("/longtermtoken", auth.RequireValidAuthentication(), authHandler.LongtermToken)
		}
	case "basic":
		authHandler := basicauth.AuthHandler{}
		authHandler.Initialize(db)
		auth.RegisterHandler(authHandler)
	default:
		panic("Unknown authentication method set. Please use one of basic, oauth.")
	}

	applicationHandler := api.ApplicationHandler{DB: db, DP: dp}
	healthHandler := api.HealthHandler{DB: db}
	notificationHandler := api.NotificationHandler{DB: db, DP: dp}
	userHandler := api.UserHandler{AH: &applicationHandler, CM: cm, DB: db, DP: dp}

	applicationGroup := r.Group("/application")
	applicationGroup.Use(auth.RequireUser()...)
	{
		applicationGroup.POST("", applicationHandler.CreateApplication)
		applicationGroup.GET("", applicationHandler.GetApplications)

		applicationGroup.GET("/:id", api.RequireIDInURI(), applicationHandler.GetApplication)
		applicationGroup.DELETE("/:id", api.RequireIDInURI(), applicationHandler.DeleteApplication)
		applicationGroup.PUT("/:id", api.RequireIDInURI(), applicationHandler.UpdateApplication)
	}

	r.GET("/health", healthHandler.Health)

	r.POST("/message", auth.RequireApplicationToken(), notificationHandler.CreateNotification)
	r.DELETE("/message/:messageid", api.RequireMessageIDInURI(), auth.RequireApplicationToken(), notificationHandler.DeleteNotification)

	userGroup := r.Group("/user")
	userGroup.Use(auth.RequireAdmin()...)
	{
		userGroup.POST("", userHandler.CreateUser)
		userGroup.GET("", userHandler.GetUsers)

		userGroup.GET("/:id", api.RequireIDInURI(), userHandler.GetUser)
		userGroup.DELETE("/:id", api.RequireIDInURI(), userHandler.DeleteUser)
		userGroup.PUT("/:id", api.RequireIDInURI(), userHandler.UpdateUser)
	}

	return r
}
