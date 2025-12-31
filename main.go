package main

import (
	"context"
	"ezqueue/auth"
	"ezqueue/auth/providers"
	"ezqueue/common"
	"ezqueue/routes"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// @title           Easy Queue API
// @version         1.0
// @description     Queue service API
// @host            https://ezq.onrender.com
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	godotenv.Load()

	app, err := initializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.FSClient.Close()

	var providersMap = map[string]auth.Provider{
		"google": &providers.GoogleProvider{
			ClientID: os.Getenv("GOOGLE_CLIENT_ID"),
		},
	}

	routes.SetupRoutes(app, providersMap)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initializeApp() (*common.App, error) {
	ctx := context.Background()

	opt := option.WithAuthCredentialsFile(option.ServiceAccount, os.Getenv("FIREBASE_CREDENTIALS"))
	firebaseApp, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		return nil, err
	}

	fsClient, err := firebaseApp.Firestore(ctx)
	if err != nil {
		return nil, err
	}

	router := gin.Default()
	router.Use(corsMiddleware())

	return &common.App{
		FirebaseApp: firebaseApp,
		AuthClient:  authClient,
		FSClient:    fsClient,
		Router:      router,
	}, nil
}

// FIXME: setup cors properly
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
