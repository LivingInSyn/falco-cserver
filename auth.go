package main

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var authConfFile = "/secrets/auth.yml"

type userConf struct {
	Users map[string]string `yaml:"users"`
}

type authenticationMiddleware struct {
	tokenUsers map[string]string
}

func NewAuthMiddleware() *authenticationMiddleware {
	// if the FCS_AUTH env var is set then use that,
	// otherwise we're going to read the file defined
	// in authConfFile
	var secretReader io.Reader
	authFileString, envSet := os.LookupEnv("FCS_AUTH")
	if envSet {
		secretReader = strings.NewReader(authFileString)
	} else {
		f, err := os.Open(authConfFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Couldn't open config file")
		}
		defer f.Close()
		secretReader = f
	}

	amw := &authenticationMiddleware{
		tokenUsers: make(map[string]string),
	}
	amw.populate(secretReader)
	return amw
}

func (amw *authenticationMiddleware) populate(reader io.Reader) {
	var u userConf
	decoder := yaml.NewDecoder(reader)
	err := decoder.Decode(&u)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal auth.yml")
	}
	amw.tokenUsers = u.Users
}

func parseBasicAuth(r *http.Request) (string, string, error) {
	auth := r.Header.Get("Authorization")
	if auth != "" {
		authr := strings.Replace(auth, "Basic ", "", 1)
		authd, err := base64.StdEncoding.DecodeString(authr)
		if err != nil {
			log.Error().Err(err).Str("auth_header", auth).Msg("got bad auth header")
			return "", "", errors.New("got a bad auth header")
		}
		authds := string(authd)
		auths := strings.Split(authds, ":")
		if len(auths) != 2 {
			log.Error().Err(err).Str("auth_header", auth).Msg("got bad auth header")
			return "", "", errors.New("got a bad auth header")
		}
		username := auths[0]
		token := auths[1]
		return username, token, nil
	}
	return "", "", errors.New("no auth token")
}

func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// skip the healthcheck
		if r.RequestURI == "/" && r.Method == "GET" {
			log.Info().Str("IP", r.RemoteAddr).Msg("healthcheck request - always allowed")
			next.ServeHTTP(w, r)
			return
		}
		// parse auth
		token := r.Header.Get("X-Auth-Token")
		username := ""
		var err error
		// if token is blank, check the basic auth header and pull the user and token out of it
		if token == "" {
			username, token, err = parseBasicAuth(r)
			if err != nil {
				//error is logged in parseBasic Auth, just return the function
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
		// note: we know we're not checking the username explicitly, just warning if it's wrong
		// but that's the same as checking just the auth token
		if user, found := amw.tokenUsers[token]; found {
			if username != "" && username != user {
				log.Warn().
					Str("user_from_map", user).
					Str("user_from_token", username).
					Msg("username in map doesn't match username in basic auth token")
			}
			// We found the token in our map
			log.Info().Str("User", user).Str("IP", r.RemoteAddr).Str("path", r.URL.Path).Msg("user authenticated")
			next.ServeHTTP(w, r)
		} else {
			log.Warn().Str("IP", r.RemoteAddr).Str("path", r.URL.Path).Msg("forbidden request")
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}
