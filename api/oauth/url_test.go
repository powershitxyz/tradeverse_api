package oauth

import (
	"chaos/api/log"
	"net/url"
	"testing"
)

func TestOAuthURL(t *testing.T) {
	url1 := "http://localhost:8083+"
	cb, err := url.Parse(url1)
	log.Info(cb, err)
}
