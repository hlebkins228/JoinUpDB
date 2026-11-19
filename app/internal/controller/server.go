package controller

import (
	"JoinUp/api"
	"JoinUp/internal/controller/handlers"
	customWare "JoinUp/internal/controller/middleware"
	"JoinUp/internal/utils/jwt"
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

func NewServer(ctx context.Context, logger *zap.Logger, eventSvc handlers.EventService, userSvc handlers.UserService, search handlers.SearchEngine, validator *handlers.Validator, jwtMgr *jwt.JwtManager, reviewSvc handlers.ReviewService) *echo.Echo {
	e := echo.New()

	e.Validator = validator

	eventHandler := handlers.NewEventHandler(ctx, logger, eventSvc, userSvc, search)
	userHandler := handlers.NewUserHandler(ctx, logger, userSvc)
	reviewHandler := handlers.NewReviewHandler(ctx, logger, reviewSvc, userSvc)
	e.Use(middleware.Recover())
	e.Use(customWare.RequestID())
	e.Use(customWare.RequestLogger(logger))

	// API Routes (must be before static files)

	apiGroup := e.Group("/api/v1")
	apiGroup.POST("/users", userHandler.CreateUser)
	apiGroup.POST("/tokens", userHandler.Auth)
	apiGroup.GET("/images/:id", userHandler.GetImage)
	apiGroup.GET("/events/:id/reviews", reviewHandler.ListReviews)
	apiGroup.GET("/categories", eventHandler.ListCategories)
	apiGroup.GET("/events", eventHandler.SearchEvents)
	apiGroup.GET("/events/:id", eventHandler.GetEvent)

	protected := apiGroup.Group("")
	auth := customWare.JWTAuthorization(logger, jwtMgr)
	protected.Use(auth)
	protected.PUT("/users/me", userHandler.UpdateUser)
	protected.GET("/users/:id", userHandler.GetUser)
	protected.POST("/images", userHandler.UploadImage)
	protected.POST("/events", eventHandler.CreateEvent)
	protected.GET("/users/me/events", eventHandler.ListCreatedEvents)
	protected.GET("/users/me/participations", eventHandler.ListJoinedEvents)
	protected.PUT("/events/:id", eventHandler.UpdateEvent)
	protected.PATCH("/events/:id", eventHandler.PatchEvent)
	protected.PUT("/events/:id/participants/me", eventHandler.JoinEvent)
	protected.DELETE("/events/:id/participants/me", eventHandler.LeaveEvent)
	protected.POST("/events/:id/images", eventHandler.UploadEventImage)
	protected.PUT("/events/:id/categories/:categoryId", eventHandler.AddEventCategory)
	protected.DELETE("/events/:id", eventHandler.DeleteEvent)
	protected.POST("/events/:id/reviews", reviewHandler.CreateReview)

	registerV2(e, v2Operations{
		createToken:         userHandler.Auth,
		createUser:          userHandler.CreateUser,
		getUser:             userHandler.GetUser,
		replaceCurrentUser:  userHandler.UpdateUser,
		listCreatedEvents:   eventHandler.ListCreatedEvents,
		listJoinedEvents:    eventHandler.ListJoinedEvents,
		listCategories:      eventHandler.ListCategories,
		listEvents:          eventHandler.SearchEvents,
		createEvent:         eventHandler.CreateEvent,
		getEvent:            eventHandler.GetEvent,
		replaceEvent:        eventHandler.UpdateEvent,
		patchEvent:          eventHandler.PatchEvent,
		deleteEvent:         eventHandler.DeleteEvent,
		putParticipation:    eventHandler.JoinEvent,
		deleteParticipation: eventHandler.LeaveEvent,
		putEventCategory:    eventHandler.AddEventCategory,
		createImage:         userHandler.UploadImage,
		createEventImage:    eventHandler.UploadEventImage,
		getImage:            userHandler.GetImage,
		listReviews:         reviewHandler.ListReviews,
		createReview:        reviewHandler.CreateReview,
	}, auth, newV2Requests())

	e.GET("/openapi.yaml", func(c *echo.Context) error {
		return c.Blob(http.StatusOK, "application/yaml", api.Specification)
	})
	e.GET("/swagger/*", echo.WrapHandler(httpSwagger.Handler(httpSwagger.URL("/openapi.yaml"))))

	// Static files
	e.File("/", "frontend/index.html")
	e.Static("/static", "frontend")

	return e
}
