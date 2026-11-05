//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"JoinUp/internal/controller/dto"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
)

func TestE2EEventLifecycle(t *testing.T) {
	runner.Run(t, "Register, login, create, read and delete event", func(pt provider.T) {
		baseURL := strings.TrimRight(os.Getenv("BACKEND_URL"), "/") + "/api/v1"
		pt.Require().NotEmpty(os.Getenv("BACKEND_URL"))
		client := &http.Client{Timeout: 10 * time.Second}
		defer client.CloseIdleConnections()
		var user dto.UserIDResponse
		var auth dto.AuthResponse
		var eventID dto.EventIDResponse
		var event dto.EventResponse
		request := func(method, path string, payload any, response any) {
			var body []byte
			var err error
			if payload != nil {
				body, err = json.Marshal(payload)
				pt.Require().NoError(err)
			}
			req, err := http.NewRequestWithContext(context.Background(), method, baseURL+path, bytes.NewReader(body))
			pt.Require().NoError(err)
			req.Header.Set("Content-Type", "application/json")
			if auth.Token != "" {
				req.Header.Set("Authorization", "Bearer "+auth.Token)
			}
			if path == "/auth" {
				req.Header.Set("login", "e2e_user")
				req.Header.Set("password", "password")
			}
			resp, err := client.Do(req)
			pt.Require().NoError(err)
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			pt.Require().NoError(err)
			pt.Require().Equal(http.StatusOK, resp.StatusCode, "%s %s: %s", method, path, data)
			if response != nil {
				pt.Require().NoError(json.Unmarshal(data, response))
			}
		}

		pt.WithNewStep("Register", func(_ provider.StepCtx) {
			request(http.MethodPost, "/user", dto.CreateUserRequest{
				Name: "E2E User", Age: 25, Login: "e2e_user", Password: "password", City: "Moscow",
			}, &user)
		})
		pt.WithNewStep("Login", func(_ provider.StepCtx) {
			request(http.MethodGet, "/auth", nil, &auth)
			pt.Require().NotEmpty(auth.Token)
		})
		pt.WithNewStep("Create event", func(_ provider.StepCtx) {
			latitude, longitude := 55.75, 37.62
			request(http.MethodPost, "/user/event", dto.CreateEventRequest{
				Name: "E2E meetup", City: "Moscow",
				EventTime: time.Date(2100, 1, 1, 12, 0, 0, 0, time.UTC),
				Location:  &dto.EventLocation{Name: "Park", Address: "Park entrance", Latitude: &latitude, Longitude: &longitude},
			}, &eventID)
		})
		path := fmt.Sprintf("/user/event/%d", eventID.ID)
		pt.WithNewStep("Read event", func(_ provider.StepCtx) {
			request(http.MethodGet, path, nil, &event)
			pt.Assert().Equal("E2E meetup", event.Name)
			pt.Assert().Equal(user.ID, event.CreatorID)
		})
		pt.WithNewStep("Delete event", func(_ provider.StepCtx) {
			request(http.MethodDelete, path, nil, nil)
		})
		pt.WithNewStep("Verify soft deletion", func(_ provider.StepCtx) {
			request(http.MethodGet, path, nil, &event)
			pt.Assert().True(event.Deleted)
		})
	})
}
