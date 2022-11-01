package integration_test

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticAssets(t *testing.T) {
	req := httptest.NewRequest("GET", "/robots.txt", nil)
	res := httptest.NewRecorder()
	newServer("").ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "text/plain; charset=UTF-8", res.Header().Get("content-type"))
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "User-agent: *\nDisallow: /\n", string(body))
}
