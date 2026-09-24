package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"JoinUp/internal/controller"
	"JoinUp/internal/controller/handlers"
	"JoinUp/internal/repository/mongo"
	"JoinUp/internal/repository/postgres"
	"JoinUp/internal/service"
	"JoinUp/internal/settings"
	"JoinUp/internal/utils/jwt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// @title JoinUp API
// @version 1.0
// @description JoinUp backend API documentation.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

/*
TODO:
Добавить пакет dto в controller и на каждый запрос/ответ (почти на каждый) своя dto
Передалать некоторую логику с учетом новых dto
Сделать PUT запрос с path параметром в виде id
*/

func main() {
	ctx, ctxCancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer ctxCancel()

	cfg, err := settings.ReadConfig()
	if err != nil {
		log.Fatalf("read config: %v", err)
	}

	var loggerCfg zap.Config
	switch cfg.Logging.Type {
	case settings.LoggerTypeProd:
		loggerCfg = zap.NewProductionConfig()
	case settings.LoggerTypeDev:
		loggerCfg = zap.NewDevelopmentConfig()
	default:
		panic("invalid logger type")
	}

	logger, err := loggerCfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		logger.Fatal("SERVER_PORT is not set")
	}
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		logger.Fatal("SERVER_HOST is not set")
	}

	var (
		uow        service.UOW
		eventRepo  service.EventRepo
		userRepo   service.UserRepo
		imageRepo  service.ImageRepo
		reviewRepo service.ReviewRepo
	)

	switch cfg.DB {
	case settings.DBTypePostgres:
		connString := settings.ReadDbConnectionConfig()
		connPool, err := pgxpool.New(ctx, connString)
		if err != nil {
			logger.Fatal("Unable to connect to DB", zap.Error(err))
		}
		defer connPool.Close()

		if err := connPool.Ping(ctx); err != nil {
			logger.Fatal("Unable to ping DB", zap.Error(err))
		}

		logger.Info("postgres db connection success")

		pgUow := postgres.NewUOW(connPool)
		pgEventRepo := postgres.NewEventRepo(connPool)
		pgUserRepo := postgres.NewUserRepo(connPool)
		pgImageRepo := postgres.NewImageRepo(connPool)
		pgReviewRepo := postgres.NewReviewRepo(connPool)

		uow = &pgUow
		eventRepo = &pgEventRepo
		userRepo = &pgUserRepo
		imageRepo = &pgImageRepo
		reviewRepo = &pgReviewRepo

	case settings.DBTypeMongo:
		connString := settings.ReadMongoConnectionConfig()
		mongoClient, err := mongodriver.Connect(options.Client().ApplyURI(connString))
		if err != nil {
			logger.Fatal("Unable to connect to MongoDB", zap.Error(err))
		}
		defer mongoClient.Disconnect(ctx)

		if err := mongoClient.Ping(ctx, nil); err != nil {
			logger.Fatal("Unable to ping MongoDB", zap.Error(err))
		}

		dbName := os.Getenv("MONGO_DB")
		if dbName == "" {
			dbName = "join_up"
		}

		logger.Info("mongo db connection success")

		if err := mongo.EnsureUserIndexes(ctx, mongoClient, dbName); err != nil {
			logger.Fatal("Unable to ensure user indexes", zap.Error(err))
		}
		if err := mongo.EnsureReviewIndexes(ctx, mongoClient, dbName); err != nil {
			logger.Fatal("Unable to ensure review indexes", zap.Error(err))
		}
		if err := mongo.EnsureCategories(ctx, mongoClient, dbName); err != nil {
			logger.Fatal("Unable to ensure categories", zap.Error(err))
		}

		mgUow := mongo.NewUOW(mongoClient)
		mgEventRepo := mongo.NewEventRepo(mongoClient, dbName)
		mgUserRepo := mongo.NewUserRepo(mongoClient, dbName)
		mgImageRepo := mongo.NewImageRepo(mongoClient, dbName)
		mgReviewRepo := mongo.NewReviewRepo(mongoClient, dbName)

		uow = &mgUow
		eventRepo = &mgEventRepo
		userRepo = &mgUserRepo
		imageRepo = &mgImageRepo
		reviewRepo = &mgReviewRepo
	}

	eventSvc := service.NewEventService(uow, eventRepo, userRepo, imageRepo)
	jwtMgr := jwt.NewJwtManager(cfg.JWT.Secret)
	userSvc := service.NewUserService(userRepo, imageRepo, eventRepo, &jwtMgr)
	searchEngine := service.NewSearchEngine(eventRepo)
	reviewSvc := service.NewReviewService(uow, reviewRepo, eventRepo)
	validator := handlers.NewValidator()

	server := controller.NewServer(ctx, logger, &eventSvc, &userSvc, &searchEngine, &validator, &jwtMgr, &reviewSvc)

	sc := echo.StartConfig{
		Address:         host + ":" + port,
		GracefulTimeout: 5 * time.Second,
	}

	fmt.Println("server started on http://" + host + ":" + port)
	fmt.Println("swagger ui on http://" + host + ":" + port + "/swagger/index.html")
	if err := sc.Start(ctx, server); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}

	logger.Info("server stop")
}
