package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"notification-service/internal/identity"
	"notification-service/internal/notification"
)

type NotificationHandler struct {
	service *notification.Service
}

type CreateNotificationRequest struct {
	RecipientID string `json:"recipient_id"`

	Type string `json:"type"`

	Title string `json:"title"`

	Body string `json:"body"`

	Resource CreateNotificationResource `json:"resource"`

	Action CreateNotificationAction `json:"action"`

	SourceEventID string `json:"source_event_id"`
}

type CreateNotificationResource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type CreateNotificationAction struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func NewNotificationHandler(
	service *notification.Service,
) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}

func (h *NotificationHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	requestIdentity, ok :=
		identity.FromContext(
			r.Context(),
		)

	if !ok {
		writeJSON(
			w,
			http.StatusUnauthorized,
			map[string]string{
				"error": "request identity not found",
			},
		)
		return
	}

	limit := int64(30)

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.ParseInt(
			value,
			10,
			64,
		)

		if err != nil {
			writeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "invalid limit",
				},
			)
			return
		}

		limit = parsed
	}

	notifications, err :=
		h.service.List(
			r.Context(),
			requestIdentity.AppID,
			requestIdentity.UserID,
			limit,
		)

	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		notifications,
	)
}

func (h *NotificationHandler) MarkAsRead(
	w http.ResponseWriter,
	r *http.Request,
) {
	requestIdentity, ok := identity.FromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "request identity not found",
		})
		return
	}

	id := r.PathValue("id")

	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "notification id is required",
		})
		return
	}

	err := h.service.MarkAsRead(
		r.Context(),
		requestIdentity.AppID,
		requestIdentity.UserID,
		id,
	)

	if errors.Is(err, notification.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "notification not found",
		})
		return
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{
		"success": true,
	})
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
func (h *NotificationHandler) UnreadCount(
	w http.ResponseWriter,
	r *http.Request,
) {
	requestIdentity, ok := identity.FromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "request identity not found",
		})
		return
	}

	count, err := h.service.UnreadCount(
		r.Context(),
		requestIdentity.AppID,
		requestIdentity.UserID,
	)

	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]int64{
			"count": count,
		},
	)
}
func (h *NotificationHandler) MarkAllAsRead(
	w http.ResponseWriter,
	r *http.Request,
) {
	requestIdentity, ok := identity.FromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "request identity not found",
		})
		return
	}

	err := h.service.MarkAllAsRead(
		r.Context(),
		requestIdentity.AppID,
		requestIdentity.UserID,
	)

	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]bool{
			"success": true,
		},
	)
}
func (h *NotificationHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	currentIdentity, ok :=
		requestIdentity(w, r)

	if !ok {
		return
	}

	var request CreateNotificationRequest

	decoder := json.NewDecoder(r.Body)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid request body",
			},
		)

		return
	}

	request.RecipientID =
		strings.TrimSpace(
			request.RecipientID,
		)

	request.Type =
		strings.TrimSpace(
			request.Type,
		)

	request.Title =
		strings.TrimSpace(
			request.Title,
		)

	request.Body =
		strings.TrimSpace(
			request.Body,
		)

	request.SourceEventID =
		strings.TrimSpace(
			request.SourceEventID,
		)

	if request.RecipientID == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "recipient_id is required",
			},
		)

		return
	}

	if request.Type == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "type is required",
			},
		)

		return
	}

	if request.Title == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "title is required",
			},
		)

		return
	}

	if request.SourceEventID == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "source_event_id is required",
			},
		)

		return
	}

	createdNotification, err :=
		h.service.Create(
			r.Context(),
			notification.CreateInput{
				AppID: currentIdentity.AppID,

				RecipientID: request.RecipientID,

				Type: request.Type,

				Actor: notification.Actor{
					ID: currentIdentity.UserID,
				},

				Resource: notification.Resource{
					Type: request.Resource.Type,

					ID: request.Resource.ID,
				},

				Title: request.Title,

				Body: request.Body,

				Action: notification.Action{
					Type: request.Action.Type,

					URL: request.Action.URL,
				},

				SourceEventID: request.SourceEventID,
			},
		)

	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to create notification",
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		createdNotification,
	)
}

func requestIdentity(
	w http.ResponseWriter,
	r *http.Request,
) (identity.Identity, bool) {
	currentIdentity, ok := identity.FromContext(
		r.Context(),
	)

	if !ok {
		writeJSON(
			w,
			http.StatusUnauthorized,
			map[string]string{
				"error": "request identity not found",
			},
		)

		return identity.Identity{}, false
	}

	return currentIdentity, true
}
