package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitStatements(t *testing.T) {
	script := `
-- leading comment
CREATE TABLE a (
    id INT,
    note VARCHAR(10) DEFAULT 'x;y'
);

INSERT INTO a VALUES (1, "a;b");
UPDATE a SET note = 'z'
`
	got := splitStatements(script)
	require.Len(t, got, 3)
	assert.Contains(t, got[0], "DEFAULT 'x;y'")
	assert.Equal(t, `INSERT INTO a VALUES (1, "a;b")`, got[1])
	assert.Equal(t, "UPDATE a SET note = 'z'", got[2])
	assert.Empty(t, splitStatements("-- nothing here\n\n"))
}

func TestNormalizeDSN(t *testing.T) {
	dsn, err := normalizeDSN("user:pass@tcp(127.0.0.1:3306)/db?charset=utf8mb4")
	require.NoError(t, err)
	assert.Contains(t, dsn, "parseTime=true")
	assert.Contains(t, dsn, "clientFoundRows=true")
	assert.Contains(t, dsn, "charset=utf8mb4")

	_, err = normalizeDSN("user@tcp(h:1)/db?timeout=not-a-duration")
	require.Error(t, err)
}
