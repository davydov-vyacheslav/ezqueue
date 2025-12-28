package models

import "time"

type User struct {
	ID          string    `firestore:"id" json:"id"`
	Email       string    `firestore:"email" json:"email"`
	DisplayName string    `firestore:"displayName" json:"displayName"`
	PhotoURL    string    `firestore:"photoURL" json:"photoURL"`
	CreatedAt   time.Time `firestore:"createdAt" json:"createdAt"`
	//	FCMToken    string    `firestore:"fcmToken" json:"fcmToken"`
}
