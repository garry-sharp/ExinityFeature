package api

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"time"
)

func BodyParseAndTimeout[T any](deadline time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var contentType ContentType
			if ct := r.Header.Get("Content-Type"); ct == "text/xml" || ct == "application/xml" {
				contentType = XML
			} else if r.Header.Get("Content-Type") == "application/json" {
				contentType = JSON
			} else {
				returnError("Unsupported content type", "", http.StatusBadRequest, w, JSON)
				return
			}

			target := new(T)
			if contentType == XML {
				if err := xml.NewDecoder(r.Body).Decode(target); err != nil {
					returnError("Invalid request", err.Error(), http.StatusBadRequest, w, XML)
					return
				}
			} else {
				if err := json.NewDecoder(r.Body).Decode(target); err != nil {
					returnError("Invalid request", err.Error(), http.StatusBadRequest, w, JSON)
					return
				}
			}
			ctx := context.WithValue(r.Context(), "request", *target)
			ctx = context.WithValue(ctx, "contentType", contentType)
			ctx, cancel := context.WithTimeout(ctx, deadline)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
