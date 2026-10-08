package integration_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"JoinUp/internal/controller"
	"JoinUp/internal/controller/dto"
	"JoinUp/internal/controller/handlers"
	repository "JoinUp/internal/repository/postgres"
	"JoinUp/internal/service"
	"JoinUp/internal/utils/jwt"
	appmigrations "JoinUp/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

var testApp struct {
	container testcontainers.Container
	pool      *pgxpool.Pool
	server    *echo.Echo
	jwtMgr    *jwt.JwtManager
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgrescontainer.Run(ctx,
		"postgres:16-alpine",
		postgrescontainer.WithDatabase("join_up_test"),
		postgrescontainer.WithUsername("postgres"),
		postgrescontainer.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(2*time.Minute),
		),
	)
	if err != nil {
		panic(err)
	}

	connString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	sqlDB, err := sql.Open("pgx", connString)
	if err != nil {
		panic(err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		panic(err)
	}

	if err := appmigrations.Up(ctx, sqlDB); err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		panic(err)
	}

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	uow := repository.NewUOW(pool)
	eventRepo := repository.NewEventRepo(pool)
	userRepo := repository.NewUserRepo(pool)
	imageRepo := repository.NewImageRepo(pool)
	reviewRepo := repository.NewReviewRepo(pool)
	eventSvc := service.NewEventService(&uow, &eventRepo, &userRepo, &imageRepo)
	jwtMgr := jwt.NewJwtManager("test_secret")
	userSvc := service.NewUserService(&userRepo, &imageRepo, &eventRepo, &jwtMgr)
	searchEngine := service.NewSearchEngine(&eventRepo)
	reviewSvc := service.NewReviewService(&uow, &reviewRepo, &eventRepo)
	validator := handlers.NewValidator()

	testApp.container = container
	testApp.pool = pool
	testApp.jwtMgr = &jwtMgr
	testApp.server = controller.NewServer(ctx, zap.NewNop(), &eventSvc, &userSvc, &searchEngine, &validator, &jwtMgr, &reviewSvc)

	code := m.Run()

	testApp.pool.Close()
	_ = testApp.container.Terminate(ctx)

	os.Exit(code)
}

func TestCreateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := prepareTestDB(t)

		body := map[string]any{
			"name":              "Moscow meetup",
			"desc":              "backend event",
			"event_time":        time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second).Format(time.RFC3339),
			"telegram_chat_url": "https://t.me/joinup_test",
			"city":              "Moscow",
			"location": map[string]any{
				"name":      "Center",
				"longitude": 37.6173,
				"latitude":  55.7558,
				"address":   "Red Square 1",
			},
		}

		rec := performAuthorizedJSONRequest(t, http.MethodPost, "/api/v1/user/event", body, 1, "user")
		require.Equal(t, http.StatusOK, rec.Code)

		var resp dto.EventIDResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Greater(t, resp.ID, 0)

		var name, city, locationName, address string
		var creatorID int
		var eventTime time.Time
		err := testApp.pool.QueryRow(ctx, `
			select e.name, e.city, e.creator_id, e.event_date, l.name, l.address
			from event e join location l on l.id = e.location_id
			where e.id = $1
		`, resp.ID).Scan(&name, &city, &creatorID, &eventTime, &locationName, &address)
		require.NoError(t, err)
		assert.Equal(t, "Moscow meetup", name)
		assert.Equal(t, "Moscow", city)
		assert.Equal(t, 1, creatorID)
		assert.Equal(t, body["event_time"], eventTime.UTC().Format(time.RFC3339))
		assert.Equal(t, "Center", locationName)
		assert.Equal(t, "Red Square 1", address)
	})

	t.Run("not found image", func(t *testing.T) {
		prepareTestDB(t)

		body := map[string]any{
			"name":       "Moscow meetup",
			"desc":       "backend event",
			"event_time": time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second).Format(time.RFC3339),
			"city":       "Moscow",
			"image_id":   9999,
			"location": map[string]any{
				"name":      "Center",
				"longitude": 37.6173,
				"latitude":  55.7558,
				"address":   "Red Square 1",
			},
		}

		rec := performAuthorizedJSONRequest(t, http.MethodPost, "/api/v1/user/event", body, 1, "user")
		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "image with such id not found")
	})
}

func TestGetEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := prepareTestDB(t)
		eventID := seedEvent(t, ctx)

		rec := performAuthorizedJSONRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/user/event/%d", eventID), nil, 1, "user")
		require.Equal(t, http.StatusOK, rec.Code)

		var resp dto.EventResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, eventID, resp.ID)
		assert.Equal(t, 1, resp.CreatorID)
		assert.Equal(t, "seeded event", resp.Name)
		assert.Equal(t, "Moscow", resp.City)
		require.NotNil(t, resp.Location)
		assert.Equal(t, "Seed location", resp.Location.Name)
	})

	t.Run("not found", func(t *testing.T) {
		prepareTestDB(t)

		rec := performAuthorizedJSONRequest(t, http.MethodGet, "/api/v1/user/event/9999", nil, 1, "user")
		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "no event with id = 9999")
	})
}

func TestUpdateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := prepareTestDB(t)
		eventID := seedEvent(t, ctx)

		body := map[string]any{
			"id":                eventID,
			"name":              "updated meetup",
			"desc":              "updated description",
			"event_time":        time.Date(2026, 7, 1, 18, 30, 0, 0, time.UTC).Format(time.RFC3339),
			"telegram_chat_url": "https://t.me/joinup_updated",
			"city":              "Saint Petersburg",
			"location": map[string]any{
				"name":      "New place",
				"longitude": 30.3141,
				"latitude":  59.9386,
				"address":   "Nevsky Prospect 1",
			},
		}

		rec := performAuthorizedJSONRequest(t, http.MethodPut, "/api/v1/user/event", body, 1, "user")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "OK")

		var name, city, locationName string
		err := testApp.pool.QueryRow(ctx, `
			select e.name, e.city, l.name
			from event e
			join location l on l.id = e.location_id
			where e.id = $1
		`, eventID).Scan(&name, &city, &locationName)
		require.NoError(t, err)
		assert.Equal(t, "updated meetup", name)
		assert.Equal(t, "Saint Petersburg", city)
		assert.Equal(t, "New place", locationName)
	})

	t.Run("not found", func(t *testing.T) {
		prepareTestDB(t)

		body := map[string]any{
			"id":         9999,
			"name":       "updated meetup",
			"desc":       "updated description",
			"event_time": time.Date(2026, 7, 1, 18, 30, 0, 0, time.UTC).Format(time.RFC3339),
			"city":       "Saint Petersburg",
			"location": map[string]any{
				"name":      "New place",
				"longitude": 30.3141,
				"latitude":  59.9386,
				"address":   "Nevsky Prospect 1",
			},
		}

		rec := performAuthorizedJSONRequest(t, http.MethodPut, "/api/v1/user/event", body, 1, "user")
		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "no event with id = 9999")
	})
}

func TestDeleteEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := prepareTestDB(t)
		eventID := seedEvent(t, ctx)

		rec := performAuthorizedJSONRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/user/event/%d", eventID), nil, 1, "user")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "OK")

		var deleted bool
		err := testApp.pool.QueryRow(ctx, `select deleted from event where id = $1`, eventID).Scan(&deleted)
		require.NoError(t, err)
		assert.True(t, deleted)
	})

	t.Run("not found", func(t *testing.T) {
		prepareTestDB(t)

		rec := performAuthorizedJSONRequest(t, http.MethodDelete, "/api/v1/user/event/9999", nil, 1, "user")
		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "no event with id = 9999")
	})
}

func prepareTestDB(t *testing.T) context.Context {
	t.Helper()

	ctx := context.Background()

	resetQueries := []string{
		`truncate table category restart identity cascade`,
		`truncate table member restart identity cascade`,
		`truncate table event restart identity cascade`,
		`truncate table users restart identity cascade`,
		`truncate table location restart identity cascade`,
		`truncate table image restart identity cascade`,
	}

	for _, query := range resetQueries {
		_, err := testApp.pool.Exec(ctx, query)
		require.NoError(t, err)
	}

	_, err := testApp.pool.Exec(ctx, `
		insert into users (id, name, age, login, password, created_at, city)
		values (1, 'misha', 22, 'admin', 'admin', now(), 'Moscow')
	`)
	require.NoError(t, err)

	return ctx
}

func seedEvent(t *testing.T, ctx context.Context) int {
	t.Helper()

	var locationID int
	err := testApp.pool.QueryRow(ctx, `
		insert into location (name, longitude, latitude, address)
		values ('Seed location', 37.6173, 55.7558, 'Red Square 1')
		returning id
	`).Scan(&locationID)
	require.NoError(t, err)

	var eventID int
	err = testApp.pool.QueryRow(ctx, `
		insert into event (name, description, created_at, updated_at, event_date, telegram_chat_url, city, creator_id, location_id, image_id, deleted)
		values ('seeded event', 'seeded description', now(), now(), now() + interval '24 hour', 'https://t.me/seeded', 'Moscow', 1, $1, null, false)
		returning id
	`, locationID).Scan(&eventID)
	require.NoError(t, err)

	return eventID
}

func performJSONRequest(t *testing.T, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, target, reader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	testApp.server.ServeHTTP(rec, req)

	return rec
}

func performAuthorizedJSONRequest(t *testing.T, method, target string, body any, userID int, role string) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, target, reader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	token, err := testApp.jwtMgr.NewToken(userID, role)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	authRec := httptest.NewRecorder()
	testApp.server.ServeHTTP(authRec, req)

	return authRec
}
