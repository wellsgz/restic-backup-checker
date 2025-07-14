package onedrive

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

// Microsoft's well-known public client ID for device code flow
const (
	PublicClientID = "d3590ed6-52b3-4102-aeff-aad2292ab01c" // Microsoft Graph PowerShell public client
	DeviceCodeURL  = "https://login.microsoftonline.com/common/oauth2/v2.0/devicecode"
	TokenURL       = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
)

// Authenticator handles OneDrive OAuth2 authentication using device code flow
type Authenticator struct {
	httpClient *http.Client
}

// DeviceCodeResponse represents the response from the device code endpoint
type DeviceCodeResponse struct {
	UserCode                string `json:"user_code"`
	DeviceCode              string `json:"device_code"`
	VerificationURI         string `json:"verification_uri"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	Message                 string `json:"message"`
	VerificationURIComplete string `json:"verification_uri_complete"`
}

// TokenResponse represents the response from the token endpoint
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// NewAuthenticator creates a new OneDrive authenticator using device code flow
func NewAuthenticator() *Authenticator {
	return &Authenticator{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Authenticate performs device code flow authentication
func (a *Authenticator) Authenticate() (*oauth2.Token, error) {
	// Step 1: Get device code
	deviceCode, err := a.getDeviceCode()
	if err != nil {
		return nil, fmt.Errorf("failed to get device code: %w", err)
	}

	// Step 2: Display instructions to user
	fmt.Printf("\n🔐 OneDrive Authentication Required\n")
	fmt.Printf("Please visit: %s\n", deviceCode.VerificationURI)
	fmt.Printf("Enter this code: %s\n\n", deviceCode.UserCode)
	fmt.Printf("Waiting for authorization...")

	// Step 3: Poll for token
	token, err := a.pollForToken(deviceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	fmt.Printf("\n✅ Successfully authenticated!\n\n")
	return token, nil
}

// getDeviceCode requests a device code from Microsoft
func (a *Authenticator) getDeviceCode() (*DeviceCodeResponse, error) {
	data := url.Values{}
	data.Set("client_id", PublicClientID)
	data.Set("scope", "https://graph.microsoft.com/Files.Read.All offline_access")

	resp, err := a.httpClient.PostForm(DeviceCodeURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed with status %d", resp.StatusCode)
	}

	var deviceCode DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceCode); err != nil {
		return nil, fmt.Errorf("failed to decode device code response: %w", err)
	}

	return &deviceCode, nil
}

// pollForToken polls the token endpoint until authentication is complete
func (a *Authenticator) pollForToken(deviceCode *DeviceCodeResponse) (*oauth2.Token, error) {
	data := url.Values{}
	data.Set("client_id", PublicClientID)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", deviceCode.DeviceCode)

	timeout := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)
	interval := time.Duration(deviceCode.Interval) * time.Second

	for time.Now().Before(timeout) {
		resp, err := a.httpClient.PostForm(TokenURL, data)
		if err != nil {
			return nil, fmt.Errorf("failed to poll token endpoint: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read token response: %w", err)
		}

		var tokenResp TokenResponse
		if err := json.Unmarshal(body, &tokenResp); err != nil {
			return nil, fmt.Errorf("failed to decode token response: %w", err)
		}

		if tokenResp.Error != "" {
			if tokenResp.Error == "authorization_pending" {
				fmt.Print(".")
				time.Sleep(interval)
				continue
			}
			return nil, fmt.Errorf("authentication failed: %s - %s", tokenResp.Error, tokenResp.ErrorDescription)
		}

		if tokenResp.AccessToken != "" {
			return &oauth2.Token{
				AccessToken:  tokenResp.AccessToken,
				RefreshToken: tokenResp.RefreshToken,
				TokenType:    tokenResp.TokenType,
				Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
			}, nil
		}

		time.Sleep(interval)
	}

	return nil, fmt.Errorf("authentication timeout")
}

// RefreshToken refreshes an expired OAuth2 token
func (a *Authenticator) RefreshToken(token *oauth2.Token) (*oauth2.Token, error) {
	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	data := url.Values{}
	data.Set("client_id", PublicClientID)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", token.RefreshToken)
	data.Set("scope", "https://graph.microsoft.com/Files.Read.All offline_access")

	resp, err := a.httpClient.PostForm(TokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("token refresh failed: %s - %s", tokenResp.Error, tokenResp.ErrorDescription)
	}

	return &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}
