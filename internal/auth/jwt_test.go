package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssuer_RoundTrip(t *testing.T) {
	issuer, err := NewIssuer("secret", time.Hour)
	require.NoError(t, err)

	token, err := issuer.Issue(42)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(time.Hour), token.ExpiresAt, time.Minute)

	id, err := issuer.Parse(token.Value)
	require.NoError(t, err)
	assert.EqualValues(t, 42, id)
}

func TestIssuer_RejectsBadTokens(t *testing.T) {
	issuer, err := NewIssuer("secret", time.Hour)
	require.NoError(t, err)
	token, err := issuer.Issue(1)
	require.NoError(t, err)

	later, err := NewIssuer("secret", time.Hour)
	require.NoError(t, err)
	later.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	_, err = later.Parse(token.Value)
	require.Error(t, err, "expired token must be rejected")

	other, err := NewIssuer("another-secret", time.Hour)
	require.NoError(t, err)
	_, err = other.Parse(token.Value)
	require.Error(t, err, "token signed with another secret must be rejected")

	_, err = issuer.Parse("not-a-token")
	require.Error(t, err)
}

func TestNewIssuer_RequiresSecret(t *testing.T) {
	_, err := NewIssuer("", time.Hour)
	require.Error(t, err)
}
