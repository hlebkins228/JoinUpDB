package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
)

const defaultBaseURL = "http://localhost:8080"

type client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

type menuItem struct {
	Label  string
	Action func(*client) error
}

type msgResponse struct {
	Msg string `json:"msg"`
}

type authResponse struct {
	Token string `json:"token"`
}

type eventIDResponse struct {
	ID int `json:"id"`
}

type eventLocation struct {
	Name      string   `json:"name"`
	Longitude *float64 `json:"longitude"`
	Latitude  *float64 `json:"latitude"`
	Address   string   `json:"address"`
}

type createEventRequest struct {
	Name            string         `json:"name"`
	Desc            string         `json:"desc,omitempty"`
	EventTime       time.Time      `json:"event_time"`
	TelegramChatURL string         `json:"telegram_chat_url,omitempty"`
	City            string         `json:"city"`
	Location        *eventLocation `json:"location"`
	ImageID         *int           `json:"image_id,omitempty"`
}

type updateEventRequest struct {
	ID              int            `json:"id"`
	Name            string         `json:"name"`
	Desc            string         `json:"desc,omitempty"`
	EventTime       time.Time      `json:"event_time"`
	TelegramChatURL string         `json:"telegram_chat_url,omitempty"`
	City            string         `json:"city"`
	Location        *eventLocation `json:"location"`
	ImageID         *int           `json:"image_id,omitempty"`
}

func main() {
	baseURL := defaultBaseURL
	if v := strings.TrimSpace(os.Getenv("JOINUP_BACKEND_URL")); v != "" {
		baseURL = strings.TrimRight(v, "/")
	}

	c := &client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	items := []menuItem{
		{Label: "1. Регистрация", Action: registerUser},
		{Label: "2. Авторизация", Action: authorize},
		{Label: "3. Создать event", Action: createEvent},
		{Label: "4. Получить event", Action: getEvent},
		{Label: "5. Обновить event", Action: updateEvent},
		{Label: "6. Удалить event", Action: deleteEvent},
		{Label: "7. Показать текущий токен", Action: showToken},
		{Label: "0. Выход", Action: func(*client) error { os.Exit(0); return nil }},
	}

	for {
		clearScreen()
		fmt.Printf("JoinUp CLI\nBackend: %s\nАвторизован: %t\n\n", c.baseURL, c.token != "")

		selector := promptui.Select{
			Label: "Выберите действие",
			Items: items,
			Size:  len(items),
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "▸ {{ .Label | cyan }}",
				Inactive: "  {{ .Label }}",
				Selected: "→ {{ .Label | green }}",
			},
		}

		idx, _, err := selector.Run()
		if err != nil {
			fmt.Printf("Ошибка меню: %v\n", err)
			return
		}

		clearScreen()
		fmt.Printf("JoinUp CLI\nBackend: %s\n\n", c.baseURL)
		if err := items[idx].Action(c); err != nil {
			fmt.Printf("Ошибка: %v\n", err)
		}
		waitForEnter()
	}
}

func authorize(c *client) error {
	login, err := promptString("Login", false)
	if err != nil {
		return err
	}
	password, err := promptString("Password", true)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/v1/auth", nil)
	if err != nil {
		return err
	}
	req.Header.Set("login", login)
	req.Header.Set("password", password)

	status, body, err := c.do(req)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		printResponse(status, body)
		return nil
	}

	var resp authResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	c.token = resp.Token
	fmt.Println("Авторизация успешна.")
	printPrettyJSON(body)
	return nil
}

func registerUser(c *client) error {
	name, err := promptString("Name", false)
	if err != nil {
		return err
	}
	age, err := promptInt("Age")
	if err != nil {
		return err
	}
	login, err := promptString("Login", false)
	if err != nil {
		return err
	}
	password, err := promptString("Password", true)
	if err != nil {
		return err
	}
	city, err := promptString("City", false)
	if err != nil {
		return err
	}
	telegramLogin, err := promptOptionalString("Telegram login (optional)")
	if err != nil {
		return err
	}
	avatarID, err := promptOptionalInt("Avatar ID (optional)")
	if err != nil {
		return err
	}

	body := map[string]any{
		"name":     name,
		"age":      age,
		"login":    login,
		"password": password,
		"city":     city,
	}
	if telegramLogin != nil {
		body["telegram_login"] = *telegramLogin
	}
	if avatarID != nil {
		body["avatar_id"] = *avatarID
	}

	return c.sendJSON(http.MethodPost, "/api/v1/user", body, false)
}

func createEvent(c *client) error {
	if err := ensureAuthorized(c); err != nil {
		return err
	}

	reqBody, err := promptEventPayload(nil)
	if err != nil {
		return err
	}

	return c.sendJSON(http.MethodPost, "/api/v1/user/event", reqBody, true)
}

func getEvent(c *client) error {
	if err := ensureAuthorized(c); err != nil {
		return err
	}

	id, err := promptInt("Event ID")
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/user/event/%d", c.baseURL, id), nil)
	if err != nil {
		return err
	}
	c.applyAuth(req)

	status, body, err := c.do(req)
	if err != nil {
		return err
	}

	printResponse(status, body)
	return nil
}

func updateEvent(c *client) error {
	if err := ensureAuthorized(c); err != nil {
		return err
	}

	id, err := promptInt("Event ID")
	if err != nil {
		return err
	}

	reqBody, err := promptEventPayload(&id)
	if err != nil {
		return err
	}

	return c.sendJSON(http.MethodPut, "/api/v1/user/event", reqBody, true)
}

func deleteEvent(c *client) error {
	if err := ensureAuthorized(c); err != nil {
		return err
	}

	id, err := promptInt("Event ID")
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/user/event/%d", c.baseURL, id), nil)
	if err != nil {
		return err
	}
	c.applyAuth(req)

	status, body, err := c.do(req)
	if err != nil {
		return err
	}

	printResponse(status, body)
	return nil
}

func showToken(c *client) error {
	if c.token == "" {
		fmt.Println("Токен не установлен.")
		return nil
	}

	fmt.Println("Текущий JWT токен:")
	fmt.Println(c.token)
	return nil
}

func promptEventPayload(id *int) (any, error) {
	name, err := promptString("Name", false)
	if err != nil {
		return nil, err
	}
	desc, err := promptStringWithDefault("Description", "", false)
	if err != nil {
		return nil, err
	}
	eventTimeRaw, err := promptStringWithDefault("Event time (RFC3339)", time.Now().Add(24*time.Hour).UTC().Format(time.RFC3339), false)
	if err != nil {
		return nil, err
	}
	eventTime, err := time.Parse(time.RFC3339, eventTimeRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid RFC3339 time: %w", err)
	}
	telegramChatURL, err := promptStringWithDefault("Telegram chat URL", "", false)
	if err != nil {
		return nil, err
	}
	city, err := promptString("City", false)
	if err != nil {
		return nil, err
	}
	locationName, err := promptString("Location name", false)
	if err != nil {
		return nil, err
	}
	longitude, err := promptFloat("Longitude")
	if err != nil {
		return nil, err
	}
	latitude, err := promptFloat("Latitude")
	if err != nil {
		return nil, err
	}
	address, err := promptString("Address", false)
	if err != nil {
		return nil, err
	}
	imageID, err := promptOptionalInt("Image ID (optional)")
	if err != nil {
		return nil, err
	}

	location := &eventLocation{
		Name:      locationName,
		Longitude: &longitude,
		Latitude:  &latitude,
		Address:   address,
	}

	if id != nil {
		return updateEventRequest{
			ID:              *id,
			Name:            name,
			Desc:            desc,
			EventTime:       eventTime,
			TelegramChatURL: telegramChatURL,
			City:            city,
			Location:        location,
			ImageID:         imageID,
		}, nil
	}

	return createEventRequest{
		Name:            name,
		Desc:            desc,
		EventTime:       eventTime,
		TelegramChatURL: telegramChatURL,
		City:            city,
		Location:        location,
		ImageID:         imageID,
	}, nil
}

func (c *client) sendJSON(method, path string, body any, withAuth bool) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if withAuth {
		c.applyAuth(req)
	}

	status, respBody, err := c.do(req)
	if err != nil {
		return err
	}

	printResponse(status, respBody)
	return nil
}

func (c *client) applyAuth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

func (c *client) do(req *http.Request) (int, []byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, body, nil
}

func ensureAuthorized(c *client) error {
	if c.token == "" {
		return fmt.Errorf("сначала выполните авторизацию")
	}
	return nil
}

func printResponse(status int, body []byte) {
	fmt.Printf("HTTP %d\n", status)
	if len(bytes.TrimSpace(body)) == 0 {
		fmt.Println("<empty body>")
		return
	}
	printPrettyJSON(body)
}

func printPrettyJSON(body []byte) {
	var out bytes.Buffer
	if err := json.Indent(&out, body, "", "  "); err != nil {
		fmt.Println(string(body))
		return
	}
	fmt.Println(out.String())
}

func promptString(label string, mask bool) (string, error) {
	prompt := promptui.Prompt{Label: label}
	if mask {
		prompt.Mask = '*'
	}
	return prompt.Run()
}

func promptStringWithDefault(label, def string, mask bool) (string, error) {
	prompt := promptui.Prompt{Label: label, Default: def}
	if mask {
		prompt.Mask = '*'
	}
	return prompt.Run()
}

func promptInt(label string) (int, error) {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			_, err := strconv.Atoi(strings.TrimSpace(input))
			return err
		},
	}
	value, err := prompt.Run()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(value))
}

func promptOptionalInt(label string) (*int, error) {
	prompt := promptui.Prompt{Label: label}
	value, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func promptOptionalString(label string) (*string, error) {
	prompt := promptui.Prompt{Label: label}
	value, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	return &value, nil
}

func promptFloat(label string) (float64, error) {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			_, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
			return err
		},
	}
	value, err := prompt.Run()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(value), 64)
}

func waitForEnter() {
	prompt := promptui.Prompt{
		Label:     "Нажмите Enter для продолжения",
		IsConfirm: false,
		Default:   "",
	}
	_, _ = prompt.Run()
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
