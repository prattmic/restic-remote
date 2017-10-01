package auth0

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	gauth0 "github.com/auth0-community/go-auth0"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

// Validator performs server-side validation of access tokens.
type Validator struct {
	v *gauth0.JWTValidator
}

// NewValidator creates a Validator that validates access tokens.
//
// jwks is the URL of the JWKS keys for the API.
// issuer is the API issuer.
// audience are the unique identifiers of the APIs being validated.
func NewValidator(jwks, issuer string, audience []string) *Validator {
	sp := NewSecretProvider(jwks)
	c := gauth0.NewConfiguration(sp, audience, issuer, jose.RS256)

	return &Validator{
		v: gauth0.NewValidator(c),
	}
}

type errorResponse struct {
	Message string `json:"message"`
}

// MethodScopes is a map from HTTP request method to the scopes required for
// that methods. Missing methods are refused access.
type MethodScopes map[string][]string

// ValidateWithScope returns an http.Handler which validates that all requests
// have a valid access token with the specified scopes before calling into h.
//
// No scopes are checked if scopes is nil.
//
// Invalid requests receive a 401 response.
func (v *Validator) ValidateWithScopes(scopes MethodScopes, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, err := v.v.ValidateRequest(r)
		if err != nil {
			log.Printf("Token is not valid or missing token: %v", err)

			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(errorResponse{
				Message: "Missing or invalid token.",
			})

			return
		}

		if scopes != nil {
			mscopes, ok := scopes[r.Method]
			if !ok {
				log.Printf("Method %s missing from scopes %+v", r.Method, scopes)

				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(errorResponse{
					Message: "Invalid scopes.",
				})

				return
			}

			if len(mscopes) > 0 {
				// Ensure the token has the correct scopes.
				if err := v.checkScopes(r, t, mscopes); err != nil {
					log.Printf("Scopes from %+v not ok want %v: %v", t, mscopes, err)

					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(errorResponse{
						Message: "Invalid scopes.",
					})

					return
				}
			}
		}

		// We have a valid token and all desired tokens. Good to go!
		h.ServeHTTP(w, r)
	})
}

// checkScopes validates that t contains all of wantScopes.
func (v *Validator) checkScopes(r *http.Request, t *jwt.JSONWebToken, wantScopes []string) error {
	claims := map[string]interface{}{}
	if err := v.v.Claims(r, t, &claims); err != nil {
		return fmt.Errorf("error retrieving claims from %+v: %v", t, err)
	}

	s, ok := claims["scope"]
	if !ok {
		return fmt.Errorf("claims missing scope: %+v", claims)
	}

	str, ok := s.(string)
	if !ok {
		return fmt.Errorf("scope (%T) %+v isn't a string", s, s)
	}

	gotScopes := strings.Split(str, " ")

	got := make(map[string]struct{}, len(gotScopes))
	for _, s := range gotScopes {
		got[s] = struct{}{}
	}

	missing := []string{}
	for _, s := range wantScopes {
		if _, ok := got[s]; !ok {
			missing = append(missing, s)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing scopes %v from %v", missing, gotScopes)
	}

	return nil
}
