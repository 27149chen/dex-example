package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/27149chen/dex-example/pkg/template"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

const (
	exampleAppState = "loutest"

	clientID = "example-app"
	clientSecret = "ZXhhbXBsZS1hcHAtc2VjcmV0"
	redirectURI = "http://example.test.8slan.com/callback"
)

func oauth2Config(scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     a.provider.Endpoint(),
		Scopes:       scopes,
		RedirectURL:  redirectURI,
	}
}

func Login(c *gin.Context) {
	var scopes []string
	if extraScopes := c.PostForm("extra_scopes"); extraScopes != "" {
		scopes = strings.Split(extraScopes, " ")
	}
	var clients []string
	if crossClients := c.PostForm("cross_client"); crossClients != "" {
		clients = strings.Split(crossClients, " ")
	}
	for _, client := range clients {
		scopes = append(scopes, "audience:server:client_id:"+client)
	}
	connectorID := ""
	if id := c.PostForm("connector_id"); id != "" {
		connectorID = id
	}

	authCodeURL := ""
	scopes = append(scopes, oidc.ScopeOpenID, "profile", "email")
	authCodeURL = a.oauth2Config(scopes).AuthCodeURL(exampleAppState)
	if connectorID != "" {
		authCodeURL = authCodeURL + "&connector_id=" + connectorID
	}

	http.Redirect(w, r, authCodeURL, http.StatusSeeOther)
}

func (a *app) handleCallback(w http.ResponseWriter, r *http.Request) {
	var (
		err   error
		token *oauth2.Token
	)

	ctx := oidc.ClientContext(r.Context(), a.client)
	oauth2Config := a.oauth2Config(nil)
	switch r.Method {
	case http.MethodGet:
		// Authorization redirect callback from OAuth2 auth flow.
		if errMsg := r.FormValue("error"); errMsg != "" {
			http.Error(w, errMsg+": "+r.FormValue("error_description"), http.StatusBadRequest)
			return
		}
		code := r.FormValue("code")
		if code == "" {
			http.Error(w, fmt.Sprintf("no code in request: %q", r.Form), http.StatusBadRequest)
			return
		}
		if state := r.FormValue("state"); state != exampleAppState {
			http.Error(w, fmt.Sprintf("expected state %q got %q", exampleAppState, state), http.StatusBadRequest)
			return
		}
		token, err = oauth2Config.Exchange(ctx, code)
	case http.MethodPost:
		// Form requests from frontend to refresh a token.
		refresh := r.FormValue("refresh_token")
		if refresh == "" {
			http.Error(w, fmt.Sprintf("no refresh_token in request: %q", r.Form), http.StatusBadRequest)
			return
		}
		t := &oauth2.Token{
			RefreshToken: refresh,
			Expiry:       time.Now().Add(-time.Hour),
		}
		token, err = oauth2Config.TokenSource(ctx, t).Token()
	default:
		http.Error(w, fmt.Sprintf("method not implemented: %s", r.Method), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get token: %v", err), http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token in token response", http.StatusInternalServerError)
		return
	}

	idToken, err := a.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to verify ID token: %v", err), http.StatusInternalServerError)
		return
	}

	accessToken, ok := token.Extra("access_token").(string)
	if !ok {
		http.Error(w, "no access_token in token response", http.StatusInternalServerError)
		return
	}

	var claims json.RawMessage
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, fmt.Sprintf("error decoding ID token claims: %v", err), http.StatusInternalServerError)
		return
	}

	buff := new(bytes.Buffer)
	if err := json.Indent(buff, []byte(claims), "", "  "); err != nil {
		http.Error(w, fmt.Sprintf("error indenting ID token claims: %v", err), http.StatusInternalServerError)
		return
	}

	template.RenderToken(w, a.redirectURI, rawIDToken, accessToken, token.RefreshToken, buff.String())
}
