// Package proxy provides a generic path-preserving reverse proxy handler:
// the gateway forwards a request unchanged (including the original
// Authorization header) to a backend service's own REST API, which matches
// its own api-docs/openapi/*.yaml 1:1 and re-verifies auth itself.
package proxy

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func New(targetBaseURL string) gin.HandlerFunc {
	target, err := url.Parse(targetBaseURL)
	if err != nil {
		panic("invalid proxy target URL " + targetBaseURL + ": " + err.Error())
	}
	rp := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		rp.ServeHTTP(c.Writer, c.Request)
	}
}
