package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func getDefaultRules() (string, error) {
	defaultRules, err := getRuleFile("_default")
	if err != nil {
		log.Error().Msg("couldn't load default rules file!")
		return "", err
	}
	return defaultRules, nil
}

func getRuleFile(rulefile string) (string, error) {
	extensions := []string{"yaml", "yml"}
	for _, ext := range extensions {
		path := fmt.Sprintf("./rules/%s.%s", rulefile, ext)
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		fileString := string(fileBytes)
		return fileString, nil
	}
	log.Error().Str("rulefile", rulefile).Msg("couldn't read rule file with yaml or yml extension")
	return "", errors.New("rule file not found")
}

func getRuleSets() ([]string, error) {
	files, err := ioutil.ReadDir("./rules/")
	if err != nil {
		log.Error().Err(err).Msg("Couldn't iterate rules directory")
		return nil, err
	}
	rulesets := make([]string, 0)
	for _, file := range files {
		// NOTE: no subdirectories right now
		if file.IsDir() {
			continue
		}
		// skip the default config
		if strings.ToLower(file.Name()) == "_default.yml" {
			continue
		}
		lowerName := strings.ToLower(file.Name())
		lowerName = strings.TrimSuffix(lowerName, ".yml")
		lowerName = strings.TrimSuffix(lowerName, ".yaml")
		rulesets = append(rulesets, lowerName)
	}
	return rulesets, nil
}

func BuildRules(rulesets []string) (string, error) {
	defaultRules, err := getDefaultRules()
	if err != nil {
		// logged in prev func
		return "", err
	}
	availableRulesets, err := getRuleSets()
	if err != nil {
		// logged in prev func
		return "", err
	}
	// make sure that rulesets is lower and then try and
	// match, returning if we find one. otherwise return the default
	// rules and log
	for _, ruleset := range rulesets {
		ruleset = strings.ToLower(ruleset)
		if ruleset == "default" {
			log.Debug().Msg("skipping default ruleset")
			continue
		}
		if !contains(availableRulesets, ruleset) {
			log.Warn().Str("ruleset", ruleset).Msg("tried to request a non-existant ruleset, continuing")
			continue
		}
		appendRules, err := getRuleFile(ruleset)
		if err != nil {
			log.Warn().Err(err).Str("ruleset", ruleset).Msg("failed to get ruleset, continuing")
			continue
		}
		defaultRules = fmt.Sprintf("%s\n\n%s\n", defaultRules, appendRules)
	}
	return defaultRules, nil
}
