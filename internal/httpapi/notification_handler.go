package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"notification-service/internal/identity"
	"notification-service/internal/notification"
)

type NotificationHandler struct {
	service *notification.Service
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
