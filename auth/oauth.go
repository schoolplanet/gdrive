package auth

import (
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"os"
	"net/http"
	"time"
)

type authCodeFn func(string) func() string

func NewFileSourceClient(tokenFile string, authFn authCodeFn) (*http.Client, error) {
	
	// Read cached token
	token, exists, err := ReadToken(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read token: %s", err)
	}
	clientId, clientSecret := getOauthClientInfo()
	conf := getConfig(clientId, clientSecret)

	// Require auth code if token file does not exist
	// or refresh token is missing
	if !exists || token.RefreshToken == "" {
		authUrl := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
		authCode := authFn(authUrl)()
		token, err = conf.Exchange(oauth2.NoContext, authCode)
		if err != nil {
			return nil, fmt.Errorf("Failed to exchange auth code for token: %s", err)
		}
	}

	return oauth2.NewClient(
		oauth2.NoContext,
		FileSource(tokenFile, token, conf),
	), nil
}

func NewRefreshTokenClient(refreshToken string) *http.Client {
	clientId, clientSecret := getOauthClientInfo()
	conf := getConfig(clientId, clientSecret)

	token := &oauth2.Token{
		TokenType:    "Bearer",
		RefreshToken: refreshToken,
		Expiry:       time.Now(),
	}

	return oauth2.NewClient(
		oauth2.NoContext,
		conf.TokenSource(oauth2.NoContext, token),
	)
}

func NewAccessTokenClient(accessToken string) *http.Client {
	clientId, clientSecret := getOauthClientInfo()
	conf := getConfig(clientId, clientSecret)

	token := &oauth2.Token{
		TokenType:   "Bearer",
		AccessToken: accessToken,
	}

	return oauth2.NewClient(
		oauth2.NoContext,
		conf.TokenSource(oauth2.NoContext, token),
	)
}

func NewServiceAccountClient(serviceAccountFile string) (*http.Client, error) {
	content, exists, err := ReadFile(serviceAccountFile)
	if(!exists) {
		return nil, fmt.Errorf("Service account filename %q not found", serviceAccountFile)
	}

	if(err != nil) {
		return nil, err
	}

	conf, err := google.JWTConfigFromJSON(content, "https://www.googleapis.com/auth/drive")
	if(err != nil) {
		return nil, err
	}
	return conf.Client(oauth2.NoContext), nil
}

func getConfig(clientId, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/drive"},
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}
}

func getOauthClientInfo() (string, string) {
    // Use Client Id and Secret from env.
    clientId := os.Getenv("OAUTH_CLIENT_ID")
    clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
    return clientId, clientSecret
}
