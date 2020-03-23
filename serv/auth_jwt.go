package serv

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	authHeader     = "Authorization"
	jwtAuth0   int = iota + 1
)

func jwtHandler(authc configAuth, next http.Handler) http.HandlerFunc {
	var key interface{}
	var jwtProvider int

	cookie := authc.Cookie

	if authc.JWT.Provider == "auth0" {
		jwtProvider = jwtAuth0
	}

	secret := authc.JWT.Secret
	publicKeyFile := authc.JWT.PubKeyFile

	switch {
	case len(secret) != 0:
		key = []byte(secret)

	case len(publicKeyFile) != 0:
		kd, err := ioutil.ReadFile(publicKeyFile)
		if err != nil {
			errlog.Fatal().Err(err).Send()
		}

		switch authc.JWT.PubKeyType {
		case "ecdsa":
			key, err = jwt.ParseECPublicKeyFromPEM(kd)

		case "rsa":
			key, err = jwt.ParseRSAPublicKeyFromPEM(kd)

		default:
			key, err = jwt.ParseECPublicKeyFromPEM(kd)

		}

		if err != nil {
			errlog.Fatal().Err(err).Send()
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var tok string

		if len(cookie) != 0 {
			ck, err := r.Cookie(cookie)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			tok = ck.Value
		} else {
			ah := r.Header.Get(authHeader)
			if len(ah) < 10 {
				next.ServeHTTP(w, r)
				return
			}
			tok = ah[7:]
		}

		token, err := jwt.ParseWithClaims(tok, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})

		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if claims, ok := token.Claims.(*jwt.StandardClaims); ok {
			ctx := r.Context()

			if jwtProvider == jwtAuth0 {
				sub := strings.Split(claims.Subject, "|")
				if len(sub) != 2 {
					ctx = context.WithValue(ctx, userIDProviderKey, sub[0])
					ctx = context.WithValue(ctx, userIDKey, sub[1])
				}
			} else {
				ctx = context.WithValue(ctx, userIDKey, claims.Subject)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		next.ServeHTTP(w, r)
	}
}
