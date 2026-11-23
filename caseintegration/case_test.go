package caseintegration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestCaseIntegration returns a CaseIntegration object that calls a local test server with the given handler
func newTestCaseIntegration(t *testing.T, handler http.HandlerFunc) *CaseIntegration {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)

	ci := NewCaseIntegration(strings.TrimPrefix(srv.URL, "https://"), "test-token", false)
	ci.apiClient.SetHTTPClient(srv.Client())
	return ci
}

func TestCaseIDIsEscapedInPath(t *testing.T) {
	ci := newTestCaseIntegration(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/v3/case/integration/..%2Fother%3Fx=1/notes" || r.URL.RawQuery != "" {
			t.Errorf("case ID should stay a single path segment, got path %q query %q", r.URL.EscapedPath(), r.URL.RawQuery)
		}
		w.Write([]byte(`{"items":[]}`))
	})

	if err := ci.AddNoteToCase(nil, "../other?x=1", "note"); err != nil {
		t.Errorf("AddNoteToCase: %v", err)
	}
}
