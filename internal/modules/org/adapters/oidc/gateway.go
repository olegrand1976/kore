package oidc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

type IDTokenClaims struct {
	Email string `json:"email"`
	Nonce string `json:"nonce"`
	jwt.RegisteredClaims
}

type Gateway struct {
	client *http.Client
	jwks   sync.Map
}

func NewGateway() *Gateway {
	return &Gateway{client: &http.Client{Timeout: 15 * time.Second}}
}

func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func S256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func BuildAuthorizeURL(issuer, clientID, redirectURI, scopes, state, codeChallenge string) (string, error) {
	candidates := authorizeURLs(issuer)
	var lastErr error
	for _, candidate := range candidates {
		u, err := url.Parse(candidate)
		if err != nil {
			lastErr = err
			continue
		}
		q := u.Query()
		q.Set("client_id", clientID)
		q.Set("response_type", "code")
		q.Set("redirect_uri", redirectURI)
		q.Set("scope", scopes)
		q.Set("state", state)
		q.Set("code_challenge", codeChallenge)
		q.Set("code_challenge_method", "S256")
		u.RawQuery = q.Encode()
		return u.String(), nil
	}
	return "", lastErr
}

func (g *Gateway) ExchangeCode(ctx context.Context, issuer, clientID, clientSecret, redirectURI, code, codeVerifier string) (TokenResponse, error) {
	candidates := tokenURLs(issuer)
	body := url.Values{}
	body.Set("grant_type", "authorization_code")
	body.Set("client_id", clientID)
	if clientSecret != "" {
		body.Set("client_secret", clientSecret)
	}
	body.Set("code", code)
	body.Set("redirect_uri", redirectURI)
	if codeVerifier != "" {
		body.Set("code_verifier", codeVerifier)
	}
	var lastErr error
	for _, tokenURL := range candidates {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(body.Encode()))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := g.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("token exchange failed: %s", string(raw))
			continue
		}
		var out TokenResponse
		if err := json.Unmarshal(raw, &out); err != nil {
			lastErr = err
			continue
		}
		return out, nil
	}
	return TokenResponse{}, lastErr
}

func (g *Gateway) ValidateIDToken(ctx context.Context, idToken, issuer, jwksURI, clientID string) (IDTokenClaims, error) {
	if idToken == "" {
		return IDTokenClaims{}, errors.New("missing id token")
	}
	keyFunc, err := g.jwksKeyFunc(ctx, jwksURI)
	if err != nil {
		return IDTokenClaims{}, err
	}
	token, err := jwt.ParseWithClaims(idToken, &IDTokenClaims{}, keyFunc,
		jwt.WithIssuer(issuer),
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
	)
	if err != nil {
		return IDTokenClaims{}, err
	}
	claims, ok := token.Claims.(*IDTokenClaims)
	if !ok || !token.Valid {
		return IDTokenClaims{}, errors.New("invalid id token claims")
	}
	if clientID != "" {
		found := false
		for _, aud := range claims.Audience {
			if aud == clientID {
				found = true
				break
			}
		}
		if !found && len(claims.Audience) > 0 {
			return IDTokenClaims{}, errors.New("audience mismatch")
		}
	}
	return *claims, nil
}

type jwkKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksDocument struct {
	Keys []jwkKey `json:"keys"`
}

func (g *Gateway) jwksKeyFunc(ctx context.Context, jwksURI string) (jwt.Keyfunc, error) {
	if jwksURI == "" {
		return nil, errors.New("jwks uri required")
	}
	doc, err := g.loadJWKS(ctx, jwksURI)
	if err != nil {
		return nil, err
	}
	keys := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, k := range doc.Keys {
		if k.Kty != "RSA" || k.N == "" || k.E == "" {
			continue
		}
		pub, err := rsaPublicKeyFromJWK(k.N, k.E)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}
	return func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		if kid != "" {
			if pub, ok := keys[kid]; ok {
				return pub, nil
			}
		}
		for _, pub := range keys {
			return pub, nil
		}
		return nil, errors.New("no matching jwk")
	}, nil
}

func (g *Gateway) loadJWKS(ctx context.Context, jwksURI string) (jwksDocument, error) {
	if cached, ok := g.jwks.Load(jwksURI); ok {
		return cached.(jwksDocument), nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURI, nil)
	if err != nil {
		return jwksDocument{}, err
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return jwksDocument{}, err
	}
	defer resp.Body.Close()
	var doc jwksDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return jwksDocument{}, err
	}
	g.jwks.Store(jwksURI, doc)
	return doc, nil
}

func rsaPublicKeyFromJWK(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nBytes)
	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}
	return &rsa.PublicKey{N: n, E: eInt}, nil
}

// ParseIDTokenPayload decodes claims without signature verification (tests / mock IdP).
func ParseIDTokenPayload(idToken string) (IDTokenClaims, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return IDTokenClaims{}, errors.New("invalid jwt format")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return IDTokenClaims{}, err
	}
	var claims IDTokenClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return IDTokenClaims{}, err
	}
	return claims, nil
}
