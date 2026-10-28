package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateAndRequired(t *testing.T) {
	issuer, err := NewIssuer("secret", time.Hour)
	require.NoError(t, err)
	token, err := issuer.Issue(7)
	require.NoError(t, err)

	var (
		gotID  int64
		gotErr error
	)
	h := NewMiddleware(issuer).Authenticate(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotID, _ = UserID(r.Context())
		_, gotErr = Required(r.Context(), nil)
	}))

	tests := []struct {
		name    string
		header  string
		wantID  int64
		wantErr string
	}{
		{name: "valid bearer", header: "Bearer " + token.Value, wantID: 7},
		{name: "lowercase scheme", header: "bearer " + token.Value, wantID: 7},
		{name: "missing header", wantErr: "missing access token"},
		{name: "wrong scheme", header: "Basic abc", wantErr: "invalid access token"},
		{name: "garbage token", header: "Bearer nope", wantErr: "invalid access token"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotErr = 0, nil
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			h.ServeHTTP(httptest.NewRecorder(), req)

			assert.Equal(t, tt.wantID, gotID)
			if tt.wantErr == "" {
				assert.NoError(t, gotErr)
				return
			}
			require.Error(t, gotErr)
			assert.Contains(t, gotErr.Error(), tt.wantErr)
		})
	}
}
