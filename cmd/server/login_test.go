package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleLoginRejectsEmptyCredentials(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
	}{
		{
			name:     "email is empty",
			email:    "",
			password: "password",
		},
		{
			name:     "password is empty",
			email:    "user@example.com",
			password: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := fmt.Sprintf(
				`{"email":%q,"password":%q}`,
				tt.email,
				tt.password,
			)
			req := httptest.NewRequest(
				http.MethodPost,
				"/login",
				strings.NewReader(requestBody),
			)
			recorder := httptest.NewRecorder()

			handleLogin(nil)(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(
					"status code = %d, want %d; body = %s",
					recorder.Code,
					http.StatusBadRequest,
					recorder.Body.String(),
				)
			}

			var response struct {
				Error string `json:"error"`
				Token string `json:"token"`
			}
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("decode response body: %v", err)
			}

			wantError := "email and password are required"
			if response.Error != wantError {
				t.Errorf("error = %q, want %q", response.Error, wantError)
			}

			if response.Token != "" {
				t.Errorf("token = %q, want empty string", response.Token)
			}
		})
	}
}
