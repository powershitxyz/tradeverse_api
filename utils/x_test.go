package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// rewrites requests to api.twitter.com to the provided base url
type rewriteTransport struct {
	base http.RoundTripper
	tgt  *url.URL
}

func (r *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "api.twitter.com" {
		// rewrite to test server, keep the path and query
		req.URL.Scheme = r.tgt.Scheme
		req.URL.Host = r.tgt.Host
	}
	return r.base.RoundTrip(req)
}

func TestTwitterOAuthFlow(t *testing.T) {
	// mock oauth + api server
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/request_token", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("oauth_token=reqtoken&oauth_token_secret=reqsecret&oauth_callback_confirmed=true"))
	})
	mux.HandleFunc("/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("oauth_token=acctok&oauth_token_secret=accsec&user_id=1&screen_name=tester"))
	})
	mux.HandleFunc("/1.1/account/verify_credentials.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserV1{ID: 1, IDStr: "1", Name: "Tester", ScreenName: "tester"})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	srvURL, _ := url.Parse(srv.URL)

	// build client and point oauth endpoints to mock server
	c := New(Config{ConsumerKey: "ck", ConsumerSecret: "cs", CallbackURL: srv.URL + "/cb"}, nil)
	c.oauthCfg.Endpoint.RequestTokenURL = srv.URL + "/oauth/request_token"
	c.oauthCfg.Endpoint.AuthorizeURL = srv.URL + "/oauth/authenticate"
	c.oauthCfg.Endpoint.AccessTokenURL = srv.URL + "/oauth/access_token"

	// rewrite api.twitter.com to mock server for verify_credentials
	prev := http.DefaultTransport
	http.DefaultTransport = &rewriteTransport{base: prev, tgt: srvURL}
	defer func() { http.DefaultTransport = prev }()

	// Start login
	authURL, requestToken, _, err := c.StartLogin(context.Background())
	if err != nil {
		t.Fatalf("StartLogin error: %v", err)
	}
	if !strings.Contains(authURL, "oauth_token=") {
		t.Fatalf("authURL missing oauth_token: %s", authURL)
	}
	if requestToken != "reqtoken" {
		t.Fatalf("unexpected request token: %s", requestToken)
	}

	// Simulate callback exchange + verify credentials
	tok, user, err := c.HandleCallback(context.Background(), "reqtoken", "verifier")
	if err != nil {
		t.Fatalf("HandleCallback error: %v", err)
	}
	if tok.AccessToken != "acctok" || tok.AccessSecret != "accsec" {
		t.Fatalf("unexpected access tokens: %#v", tok)
	}
	if user == nil || user.ScreenName != "tester" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

func TestMaskEmail(t *testing.T) {
	email := "test@example.com"
	maskedEmail := MaskEmail(email)

	t.Log(maskedEmail)
}

func TestMaskUserNo(t *testing.T) {
	userNo := "108934900173239"
	maskedUserNo := MaskUserNo(userNo)
	t.Log(maskedUserNo)
}
