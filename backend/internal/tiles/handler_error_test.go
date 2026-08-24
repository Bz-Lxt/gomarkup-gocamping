package tiles_test

import (
	"errors"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"gocamping/internal/tiles"
)

var errUpstreamInterrupted = errors.New("upstream connection interrupted")

type interruptedTransport struct{}

func (interruptedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       &interruptedPNGBody{},
		Request:    req,
	}, nil
}

type interruptedPNGBody struct {
	sent bool
}

func (b *interruptedPNGBody) Read(p []byte) (int, error) {
	if b.sent {
		return 0, io.EOF
	}
	b.sent = true
	n := copy(p, []byte{137, 80, 78, 71, 13, 10, 26, 10})
	return n, errUpstreamInterrupted
}

func (*interruptedPNGBody) Close() error { return nil }

func TestHandlerFallsBackWhenUpstreamBodyIsInterrupted(t *testing.T) {
	t.Setenv("TILE_PROVIDER", "osm")
	originalTransport := http.DefaultTransport
	http.DefaultTransport = interruptedTransport{}
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	router := chi.NewRouter()
	router.Get("/tiles/{z}/{x}/{y}.png", tiles.Handler)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/tiles/8/216/107.png", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "image/png", recorder.Header().Get("Content-Type"))
	_, err := png.Decode(recorder.Body)
	require.NoError(t, err, "tile response must remain decodable after an interrupted upstream body")
}
