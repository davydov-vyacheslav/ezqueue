package models

import "time"

type Ticket struct {
	ID              string    `firestore:"-" json:"id"`
	QueueID         string    `firestore:"queueId" json:"queueId"`
	UserID          string    `firestore:"userId" json:"userId"`
	TicketNumber    int       `firestore:"ticketNumber" json:"ticketNumber"`
	CreatedAt       time.Time `firestore:"createdAt" json:"createdAt"`
	CompletedAt     time.Time `firestore:"completedAt" json:"completedAt"`
	StartedAt       time.Time `firestore:"startedAt" json:"startedAt"`
	CashierID       string    `firestore:"cashierId" json:"cashierId"`
	CashierName     string    `firestore:"cashierName" json:"cashierName"` // FIXME?
	Status          string    `firestore:"status" json:"status"`           // waiting, processing, served, deleted, timeout FIXME: enum
	PositionInQueue int       `firestore:"positionInQueue" json:"positionInQueue"`
}
