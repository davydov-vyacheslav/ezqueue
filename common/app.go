package common

import (
	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

type App struct {
	FirebaseApp *firebase.App
	AuthClient  *auth.Client
	FSClient    *firestore.Client
	Router      *gin.Engine
}
