package handlers

import (
	"JoinUp/internal/controller/dto"
	"JoinUp/internal/exceptions"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

func (h *EventHandler) PatchEvent(c *echo.Context) error {
	mediaType, _, err := mime.ParseMediaType(c.Request().Header.Get("Content-Type"))
	if err != nil || mediaType != "application/merge-patch+json" {
		return c.JSON(http.StatusUnsupportedMediaType, dto.Msg{Msg: "expected application/merge-patch+json"})
	}
	var patch map[string]json.RawMessage
	decoder := json.NewDecoder(c.Request().Body)
	if err := decoder.Decode(&patch); err != nil || patch == nil {
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: "expected a JSON object"})
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: "expected one JSON document"})
	}
	patchData, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	var fields dto.UpdateEventRequest
	decoder = json.NewDecoder(bytes.NewReader(patchData))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fields); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Msg{Msg: getBindMsg(err)})
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		return c.JSON(http.StatusUnprocessableEntity, dto.Msg{Msg: "invalid event id"})
	}
	userID, role, err := h.UserFromContext(c)
	if err != nil || userID == nil {
		return err
	}
	if allowed, err := h.checkPerms(c, id, *userID, role); !allowed {
		return err
	}
	model, err := h.eventSvc.GetEvent(c.Request().Context(), id)
	if errors.Is(err, exceptions.ErrNotFound) {
		return c.JSON(http.StatusNotFound, dto.Msg{Msg: err.Error()})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.Msg{Msg: MsgInternalServerError})
	}
	event := dto.EventFromModel(model)
	categoryIDs := make([]int, 0, len(event.Categories))
	for _, category := range event.Categories {
		categoryIDs = append(categoryIDs, category.ID)
	}
	current := map[string]any{
		"name": event.Name, "desc": event.Desc, "event_time": event.EventTime,
		"telegram_chat_url": event.TelegramChatURL, "city": event.City,
		"location": event.Location, "image_id": event.ImageID, "category_ids": categoryIDs,
	}
	data, err := json.Marshal(current)
	if err != nil {
		return err
	}
	var target map[string]json.RawMessage
	if err := json.Unmarshal(data, &target); err != nil {
		return err
	}
	mergeEventFields(target, patch)
	data, err = json.Marshal(target)
	if err != nil {
		return err
	}
	c.Request().Body = io.NopCloser(bytes.NewReader(data))
	c.Request().ContentLength = int64(len(data))
	c.Request().Header.Set("Content-Type", "application/json")
	return h.UpdateEvent(c)
}

func mergeEventFields(target, patch map[string]json.RawMessage) {
	for key, value := range patch {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			delete(target, key)
			continue
		}
		var nestedPatch map[string]json.RawMessage
		if json.Unmarshal(value, &nestedPatch) == nil && nestedPatch != nil {
			nestedTarget := make(map[string]json.RawMessage)
			_ = json.Unmarshal(target[key], &nestedTarget)
			if nestedTarget == nil {
				nestedTarget = make(map[string]json.RawMessage)
			}
			mergeEventFields(nestedTarget, nestedPatch)
			target[key], _ = json.Marshal(nestedTarget)
		} else {
			target[key] = value
		}
	}
}
