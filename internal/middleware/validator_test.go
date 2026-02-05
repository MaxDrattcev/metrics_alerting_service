package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateContentType(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		contentType string
		wantStatus  int
		wantBody    string
	}{
		{
			name:        "POST with valid Content-Type",
			method:      http.MethodPost,
			contentType: "text/plain",
			wantStatus:  http.StatusOK,
			wantBody:    "OK",
		},
		{
			name:        "POST with invalid Content-Type",
			method:      http.MethodPost,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
			wantBody:    "Content-Type must be text/plain",
		},
		{
			name:        "POST with empty Content-Type",
			method:      http.MethodPost,
			contentType: "",
			wantStatus:  http.StatusBadRequest,
			wantBody:    "Content-Type must be text/plain",
		},
		{
			name:        "GET request (should pass)",
			method:      http.MethodGet,
			contentType: "application/json",
			wantStatus:  http.StatusOK,
			wantBody:    "OK",
		},
		{
			name:        "PUT request (should pass)",
			method:      http.MethodPut,
			contentType: "application/json",
			wantStatus:  http.StatusOK,
			wantBody:    "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(ValidateContentType())

			router.Any("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}
