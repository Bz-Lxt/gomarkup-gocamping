package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestJWTRoundtrip(t *testing.T) {
	tok, err := SignJWT("s3cret", 42, "leader", time.Hour)
	require.NoError(t, err)
	c, err := ParseJWT("s3cret", tok)
	require.NoError(t, err)
	require.Equal(t, int64(42), c.UserID)
	require.Equal(t, "leader", c.Role)
}

func TestJSONEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	JSON(rec, req, 200, map[string]int{"n": 1})
	require.Equal(t, 200, rec.Code)
	require.Contains(t, rec.Body.String(), `"code":0`)
}
