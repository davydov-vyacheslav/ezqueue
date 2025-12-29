package handlers

import (
	"ezqueue/common"
	"ezqueue/models"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

type QueueHandler struct {
	App *app.App
}

func NewQueueHandler(application *app.App) *QueueHandler {
	return &QueueHandler{App: application}
}

type CreateQueueRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	EventTime   time.Time `json:"eventTime"`
}

type JoinQueueRequest struct {
	UniqueID string `json:"uniqueId" binding:"required"`
}

// FIXME: UUID generator?
// FIXME: to Utils
func generateUniqueID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (h *QueueHandler) CreateQueue(c *gin.Context) {
	var req CreateQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	queueID := generateUniqueID()

	queue := models.Queue{
		ID:          queueID,
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		EventTime:   req.EventTime,
		CreatedAt:   time.Now(),
		CreatedBy:   userID,
		Status:      "active",
		MentorIDs:   []string{},
		CashierIDs:  []string{},
	}

	_, err := h.App.FSClient.Collection("queues").Doc(queueID).Set(c.Request.Context(), queue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create queue"})
		return
	}

	c.JSON(http.StatusCreated, queue)
}

func (h *QueueHandler) ListQueues(c *gin.Context) {
	status := c.DefaultQuery("status", "active")

	iter := h.App.FSClient.Collection("queues").
		Where("status", "==", status).
		Documents(c.Request.Context())

	var queues []models.Queue
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queues"})
			return
		}

		var queue models.Queue
		if err := doc.DataTo(&queue); err != nil {
			continue
		}
		queues = append(queues, queue)
	}

	c.JSON(http.StatusOK, queues)
}

func (h *QueueHandler) GetQueue(c *gin.Context) {
	queueID := c.Param("id")

	doc, err := h.App.FSClient.Collection("queues").Doc(queueID).Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Queue not found"})
		return
	}

	var queue models.Queue
	if err := doc.DataTo(&queue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse queue"})
		return
	}

	c.JSON(http.StatusOK, queue)
}

func (h *QueueHandler) JoinQueue(c *gin.Context) {
	var req JoinQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")

	// Find queue by uniqueID
	iter := h.App.FSClient.Collection("queues").
		Where("uniqueId", "==", req.UniqueID).
		Limit(1).
		Documents(c.Request.Context())

	doc, err := iter.Next()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Queue not found"})
		return
	}

	var queue models.Queue
	if err := doc.DataTo(&queue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse queue"})
		return
	}

	if queue.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Queue is closed"})
		return
	}

	// Check if user already in queue
	existingMembership := h.App.FSClient.Collection("queueMemberships").
		Where("userId", "==", userID).
		Where("queueId", "==", queue.ID).
		Where("status", "==", "active").
		Limit(1).
		Documents(c.Request.Context())

	if _, err := existingMembership.Next(); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already in this queue"})
		return
	}

	// Create ticket
	ticketID := fmt.Sprintf("ticket_%d", time.Now().UnixNano())
	ticket := models.Ticket{
		ID:        ticketID,
		QueueID:   queue.ID,
		UserID:    userID,
		CreatedAt: time.Now(),
		Status:    "waiting",
	}

	_, err = h.App.FSClient.Collection("tickets").Doc(ticketID).Set(c.Request.Context(), ticket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
		return
	}

	// Update queue current number
	_, err = h.App.FSClient.Collection("queues").Doc(queue.ID).Update(
		c.Request.Context(),
		[]firestore.Update{
			{Path: "currentNumber", Value: "TO BE DONE: ticekt number"}, // FIXME
		},
	)

	c.JSON(http.StatusCreated, ticket)
}

func (h *QueueHandler) CloseQueue(c *gin.Context) {
	// TODO: only owner can close queue
	// TODO: close all opened ticekts as 'closed by queue close'
	queueID := c.Param("id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.App.FSClient.Collection("queues").Doc(queueID).Update(
		c.Request.Context(),
		[]firestore.Update{
			{Path: "status", Value: "closed"},
			{Path: "closureReason", Value: req.Reason},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close queue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Queue closed"})
}

func (h *QueueHandler) AssignMentors(c *gin.Context) {
	queueID := c.Param("id")

	var req struct {
		MentorIDs []string `json:"mentorIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.App.FSClient.Collection("queues").Doc(queueID).Update(
		c.Request.Context(),
		[]firestore.Update{
			{Path: "mentorIds", Value: req.MentorIDs},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign mentors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mentors assigned"})
}
