package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildRules(t *testing.T) {
	rules, err := BuildRules([]string{"sample"})
	if err != nil {
		t.Fatalf("failed to build rules")
	}
	if len(rules) < 10 {
		t.Fatalf("rules is suspiciously short")
	}
}

func TestHash(t *testing.T) {
	rules, err := BuildRules([]string{"sample"})
	if err != nil {
		t.Fatalf("failed to build rules")
	}
	// write the rules to a test file
	tDir := t.TempDir()
	tFile := filepath.Join(tDir, "hashtest")
	err = os.WriteFile(tFile, []byte(rules), 0644)
	if err != nil {
		t.Fatalf("failed to write test file for hashtest")
	}
	// hash the file
	f, err := os.Open(tFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	h1 := sha256.New()
	if _, err := io.Copy(h1, f); err != nil {
		t.Fatal(err)
	}
	h1sum := h1.Sum(nil)
	// call get sum from main
	rPath := fmt.Sprintf("/sum?rulesets=sample")
	req, err := http.NewRequest("GET", rPath, nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetSum)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	// compare the sums
	if rr.Body.String() != fmt.Sprintf("%x", h1sum) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), fmt.Sprintf("%x", h1sum))
	}
}
