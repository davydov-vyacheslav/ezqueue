package handlers

import (
	"ezqueue/common"
	"ezqueue/models"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

type TicketHandler struct {
	App *common.App
}

func NewTicketHandler(application *common.App) *TicketHandler {
	return &TicketHandler{App: application}
}

func (h *TicketHandler) GetMyTickets(c *gin.Context) {
	userID := c.GetString("userID")

	iter := h.App.FSClient.Collection("tickets").
		Where("userId", "==", userID).
		Where("status", "in", []string{"waiting", "processing"}).
		OrderBy("createdAt", firestore.Desc).
		Documents(c.Request.Context())

	var tickets []models.Ticket
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.GetErrorResponse("Failed to fetch tickets"))
			return
		}

		var ticket models.Ticket
		if err := doc.DataTo(&ticket); err != nil {
			continue
		}
		tickets = append(tickets, ticket)
	}

	c.JSON(http.StatusOK, tickets)
}

// TODO: to be removed as part of /queues/:id endpoint
func (h *TicketHandler) GetQueueTickets(c *gin.Context) {
	queueID := c.Param("id")

	iter := h.App.FSClient.Collection("tickets").
		Where("queueId", "==", queueID).
		OrderBy("ticketNumber", firestore.Asc).
		Documents(c.Request.Context())

	var tickets []models.Ticket
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.GetErrorResponse("Failed to fetch tickets"))
			return
		}

		var ticket models.Ticket
		if err := doc.DataTo(&ticket); err != nil {
			continue
		}
		tickets = append(tickets, ticket)
	}

	c.JSON(http.StatusOK, tickets)
}

// TODO: to be remastered: owner, mentor can close user ticket;
//
//	user can close its ticket
func (h *TicketHandler) DeleteTicket(c *gin.Context) {
	ticketID := c.Param("id")

	_, err := h.App.FSClient.Collection("tickets").Doc(ticketID).Update(
		c.Request.Context(),
		[]firestore.Update{
			{Path: "status", Value: "deleted"},
			{Path: "completedAt", Value: time.Now()},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.GetErrorResponse("Failed to delete ticket"))
		return
	}

	// Update membership
	iter := h.App.FSClient.Collection("queueMemberships").
		Where("ticketId", "==", ticketID).
		Limit(1).
		Documents(c.Request.Context())

	if doc, err := iter.Next(); err == nil {
		doc.Ref.Update(c.Request.Context(), []firestore.Update{
			{Path: "status", Value: "removed"},
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket deleted"})
}
