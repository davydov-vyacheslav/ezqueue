package models

import "time"

type Queue struct {
	ID            string    `firestore:"-" json:"id"`
	Name          string    `firestore:"name" json:"name"`
	Description   string    `firestore:"description" json:"description"`
	EventTime     time.Time `firestore:"eventTime" json:"eventTime"`
	Location      string    `firestore:"location" json:"location"`
	CreatedAt     time.Time `firestore:"createdAt" json:"createdAt"`
	CreatedBy     string    `firestore:"createdBy" json:"createdBy"`
	Status        string    `firestore:"status" json:"status"` // active, closed, ... // FIXME: enum?
	ClosureReason string    `firestore:"closureReason" json:"closureReason"`
	MentorIDs     []string  `firestore:"mentorIds" json:"mentorIds"`
	CashierIDs    []string  `firestore:"activeCashierIds" json:"activeCashierIds"`
}
