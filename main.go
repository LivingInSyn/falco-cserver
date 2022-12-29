package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type RulesetRequest struct {
	Rulesets []string `json:"rulesets"`
}

func configLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	_, dob := os.LookupEnv("FAM_DEBUG")
	if dob {
		log.Info().Msg("Log level set to DEBUG")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		log.Info().Msg("Log level set to default")
	}
	log.Info().Msg("Logger setup")
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func GetRules(w http.ResponseWriter, r *http.Request) {
	rs := r.URL.Query().Get("rulesets")
	if rs == "" {
		log.Warn().Msg("got request without rulesets")
		respondWithError(w, http.StatusBadRequest, "rulesets parameter is required")
		return
	}
	rulesets := strings.Split(rs, ",")
	rr := RulesetRequest{Rulesets: rulesets}
	rulesFile, err := BuildRules(rr.Rulesets)
	if err != nil {
		log.Error().Err(err).Msg("bad request to GetRules")
		respondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	// return as plain text
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/text")
	w.Write([]byte(rulesFile))
}

func GetSum(w http.ResponseWriter, r *http.Request) {
	rs := r.URL.Query().Get("rulesets")
	if rs == "" {
		log.Warn().Msg("got request without rulesets to sum")
		respondWithError(w, http.StatusBadRequest, "rulesets parameter is required")
		return
	}
	rulesets := strings.Split(rs, ",")
	rr := RulesetRequest{Rulesets: rulesets}
	rulesFile, err := BuildRules(rr.Rulesets)
	if err != nil {
		log.Error().Err(err).Msg("bad request to GetSum")
		respondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	h := sha256.New()
	h.Write([]byte(rulesFile))
	sum := h.Sum(nil)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("%x", sum)))
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	// Logger config
	configLogger()
	// Load port info
	port := os.Getenv("PORT")
	if port == "" {
		log.Warn().Msg("no PORT env var set, defaulting to 8080")
		port = "8080"
	}
	port = fmt.Sprintf(":%s", port)
	// setup routes
	r := mux.NewRouter()
	r.HandleFunc("/", HealthCheck)
	r.HandleFunc("/ruleset", GetRules).Methods("GET")
	r.HandleFunc("/sum", GetSum).Methods("GET")
	//setup auth middleware
	amw := authenticationMiddleware{make(map[string]string)}
	amw.Populate()
	r.Use(amw.Middleware)
	// start the server
	err := http.ListenAndServe(port, r)
	log.Fatal().Err(err).Msg("http server stopped")
}
