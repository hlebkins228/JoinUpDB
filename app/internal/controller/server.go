package controller

import (
	_ "JoinUp/docs"
	"JoinUp/internal/controller/handlers"
	customWare "JoinUp/internal/controller/middleware"
	"JoinUp/internal/utils/jwt"
	"context"

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
	reviewHandler := handlers.NewReviewHandler(ctx, logger, reviewSvc)
	e.Use(middleware.Recover())
	e.Use(customWare.RequestID())
	e.Use(customWare.RequestLogger(logger))

	// API Routes (must be before static files)

	apiGroup := e.Group("/api/v1")
	apiGroup.POST("/user", userHandler.CreateUser)
	apiGroup.GET("/auth", userHandler.Auth)
	apiGroup.GET("/image/:id", userHandler.GetImage)
	apiGroup.GET("/event/:id/review", reviewHandler.ListReviews)
	apiGroup.GET("/user/category", eventHandler.ListCategories)
	apiGroup.GET("/user/event/search", eventHandler.SearchEvents)
	apiGroup.GET("/user/event/:id", eventHandler.GetEvent)

	userGroup := apiGroup.Group("/user")
	userGroup.Use(customWare.JWTAuthorization(logger, jwtMgr))

	userGroup.PUT("", userHandler.UpdateUser)
	userGroup.GET("/:id", userHandler.GetUser)
	userGroup.POST("/image", userHandler.UploadImage)
	userGroup.POST("/event", eventHandler.CreateEvent)
	userGroup.GET("/event/created", eventHandler.ListCreatedEvents)
	userGroup.GET("/event/joined", eventHandler.ListJoinedEvents)
	userGroup.PUT("/event", eventHandler.UpdateEvent)
	userGroup.PUT("/event/:id", eventHandler.UpdateEvent)
	userGroup.POST("/event/:id/join", eventHandler.JoinEvent)
	userGroup.DELETE("/event/:id/join", eventHandler.LeaveEvent)
	userGroup.PUT("/event/:id/image", eventHandler.UploadEventImage)
	userGroup.POST("/event/:id/category", eventHandler.AddEventCategory)
	userGroup.DELETE("/event/:id", eventHandler.DeleteEvent)
	userGroup.POST("/event/:id/review", reviewHandler.CreateReview)
	e.GET("/swagger/*", echo.WrapHandler(httpSwagger.Handler()))

	// Static files
	e.File("/", "frontend/index.html")
	e.Static("/static", "frontend")

	return e
}
