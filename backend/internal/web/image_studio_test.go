//go:build embed

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestImageStudioStaticDirectoryAndMissingAsset(t *testing.T) {
	files := fstest.MapFS{"index.html": {Data: []byte("main-spa")}, "image-studio-app/index.html": {Data: []byte("studio-app")}, "image-studio-app/assets/app.js": {Data: []byte("studio-js")}}
	server := &FrontendServer{distFS: files, fileServer: http.FileServer(http.FS(files))}
	router := gin.New()
	router.Use(server.Middleware())
	for _, tc := range []struct {
		path, body string
		status     int
	}{
		{"/image-studio-app/", "studio-app", 200},
		{"/image-studio-app/assets/app.js", "studio-js", 200},
		{"/image-studio-app/assets/missing.js", "", 404},
	} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		require.Equal(t, tc.status, w.Code)
		require.Equal(t, tc.body, w.Body.String())
	}
}
