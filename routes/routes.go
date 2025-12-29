package routes

import (
	"ezqueue/auth"
	"ezqueue/common"
	"ezqueue/handlers"
)

func SetupRoutes(app *app.App, providers map[string]auth.Provider) {

	authHandler := handlers.NewAuthHandler(app, providers)
	queueHandler := handlers.NewQueueHandler(app)
	ticketHandler := handlers.NewTicketHandler(app)

	// public routes
	app.Router.POST("/auth/login", authHandler.Login)
	app.Router.POST("/auth/refresh", authHandler.Refresh)

	api := app.Router.Group("/api/v1")

	// Protected routes
	protected := api.Group("")
	protected.Use(authHandler.JWTAuth)
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
