package httphandler

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

var validID = regexp.MustCompile(`^[a-z0-9]{15}$`)

// requireAuth wraps a handler with authentication check.
// Pass an optional core.App to also accept a ?token= query param (needed for EventSource/SSE).
func requireAuth(next func(*core.RequestEvent) error, apps ...core.App) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if e.Auth == nil && len(apps) > 0 {
			// Prefer Authorization header so the token stays out of access logs.
			// PocketBase SDK sends the bare token; Bearer prefix is also accepted.
			token := e.Request.Header.Get("Authorization")
			token = strings.TrimPrefix(token, "Bearer ")
			if token == "" {
				token = e.Request.URL.Query().Get("token")
			}
			if token != "" {
				if record, err := apps[0].FindAuthRecordByToken(token, core.TokenTypeAuth); err == nil {
					e.Auth = record
				}
			}
		}
		if e.Auth == nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		return next(e)
	}
}

// requireAdmin wraps requireAuth and additionally requires the auth record to
// hold the admin role. Used for every state-changing endpoint.
func requireAdmin(next func(*core.RequestEvent) error, apps ...core.App) func(*core.RequestEvent) error {
	return requireAuth(func(e *core.RequestEvent) error {
		if e.Auth.GetString("role") != "admin" {
			return e.JSON(http.StatusForbidden, map[string]string{"error": "admin role required"})
		}
		return next(e)
	}, apps...)
}
