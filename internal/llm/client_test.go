package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCompleteParsesContentAndPostsModel(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"hello [1]"}}]}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "qwen2.5:14b", "secret", 5*time.Second)
	out, err := c.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, 0.2, 256)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if out != "hello [1]" {
		t.Errorf("content = %q", out)
	}
	if gotPath != "/chat/completions" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer secret" {
		t.Errorf("auth = %q", gotAuth)
	}
	if gotBody["model"] != "qwen2.5:14b" {
		t.Errorf("model = %v", gotBody["model"])
	}
}

func TestCompleteNoAuthHeaderWhenKeyEmpty(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "m", "", 5*time.Second)
	if _, err := c.Complete(context.Background(), []Message{{Role: "user", Content: "x"}}, 0, 16); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("expected no auth header, got %q", gotAuth)
	}
}
