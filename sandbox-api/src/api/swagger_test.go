package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSandboxHost(t *testing.T) {
	cases := []struct{ env, workspaceID, want string }{
		{"prod", "WS1", "sbx-box-ws1.us-pdx-1.b10.co"},
		{"dev", "WS1", "sbx-box-ws1.us-pdx-1.dev.b10.co"},
		{"", "WS1", ""},
		{"prod", "", ""},
	}
	for _, c := range cases {
		if got := sandboxHost(c.env, c.workspaceID, "box", "us-pdx-1"); got != c.want {
			t.Errorf("sandboxHost(%q, %q) = %q, want %q", c.env, c.workspaceID, got, c.want)
		}
	}
}

func TestPublicHostPrefersInjectedURL(t *testing.T) {
	t.Setenv("BASETEN_SANDBOX_URL", "") // the runner's own environment must not leak in
	t.Setenv("BASETEN_SANDBOX_ENV", "prod")
	t.Setenv("BASETEN_SANDBOX_TEAM_ID", "WS1")
	t.Setenv("BASETEN_SANDBOX_NAME", "box")
	t.Setenv("BASETEN_SANDBOX_REGION", "us-pdx-1")
	if got := publicHost(); got != "sbx-box-ws1.us-pdx-1.b10.co" {
		t.Fatalf("fallback: got %q", got)
	}

	t.Setenv("BASETEN_SANDBOX_URL", "https://sbx-box-ws1.us-pdx-1.b10.co")
	if got := publicHost(); got != "sbx-box-ws1.us-pdx-1.b10.co" {
		t.Fatalf("BASETEN_SANDBOX_URL: got %q", got)
	}

	t.Setenv("BASETEN_SANDBOX_URL", "https://sbx-box-ws1.us-pdx-1.b10.co?source=x#y")
	if got := publicHost(); got != "sbx-box-ws1.us-pdx-1.b10.co" {
		t.Fatalf("query and fragment must not be part of the host: got %q", got)
	}
}

// doc.json must follow the environment at request time, not at startup.
func TestSwaggerDocFollowsEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/swagger/*any", swaggerHandler())
	host := func() (string, string) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil))
		var doc struct{ Host, BasePath string }
		if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
			t.Fatalf("doc.json is not JSON: %v", err)
		}
		return doc.Host, doc.BasePath
	}

	t.Setenv("BASETEN_SANDBOX_URL", "") // exercise the fallback, whatever the runner exports
	t.Setenv("BASETEN_SANDBOX_ENV", "prod")
	t.Setenv("BASETEN_SANDBOX_TEAM_ID", "WS1")
	t.Setenv("BASETEN_SANDBOX_NAME", "box")
	t.Setenv("BASETEN_SANDBOX_REGION", "us-pdx-1")
	if h, base := host(); h != "sbx-box-ws1.us-pdx-1.b10.co" || base != "/" {
		t.Fatalf("got host %q basePath %q", h, base)
	}

	t.Setenv("BASETEN_SANDBOX_NAME", "fork")
	if h, _ := host(); h != "sbx-fork-ws1.us-pdx-1.b10.co" {
		t.Fatalf("host did not follow env change: %q", h)
	}
}
