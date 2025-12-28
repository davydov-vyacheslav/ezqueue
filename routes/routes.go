package routes

import (
	"ezqueue/app"
	"ezqueue/handlers"
)

func SetupRoutes(app *app.App) {

	authHandler := handlers.NewAuthHandler(app)
	queueHandler := handlers.NewQueueHandler(app)
	ticketHandler := handlers.NewTicketHandler(app)

	api := app.Router.Group("/api/v1")

	// Public routes
	api.POST("/auth/google", authHandler.HandleGoogleAuth)

	// Protected routes
	protected := api.Group("")
	protected.Use(app.AuthMiddleware())
	{
		// User routes
		protected.GET("/users/me", authHandler.GetCurrentUser)

		// Queue routes
		protected.GET("/queues", queueHandler.ListQueues)
		protected.POST("/queues", queueHandler.CreateQueue)
		protected.GET("/queues/:id", queueHandler.GetQueue)
		protected.POST("/queues/join", queueHandler.JoinQueue)
		protected.POST("/queues/:id/close", queueHandler.CloseQueue)
		protected.POST("/queues/:id/mentors", queueHandler.AssignMentors)

		// Tickets routes
		protected.DELETE("/tickets/:id", ticketHandler.DeleteTicket)
		protected.GET("/queues/:id/tickets", ticketHandler.GetQueueTickets)
		protected.GET("/tickets/my", ticketHandler.GetMyTickets)

		// Cashier routes
		//protected.POST("/cashiers/register", cashierHandler.registerCashier)
		//protected.PUT("/cashiers/:id/status", cashierHandler.updateCashierStatus)
		//protected.POST("/cashiers/:id/subscribe", cashierHandler.subscribeToCashierQueue)
		//protected.POST("/cashiers/:id/select-ticket", cashierHandler.selectNextTicket)
		//protected.PUT("/cashiers/tickets/:ticketId/complete", cashierHandler.completeTicket)
	}
}
