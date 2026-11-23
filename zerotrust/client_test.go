package zerotrust

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestZeroTrust returns a ZeroTrust object that calls a local test server with the given handler
func newTestZeroTrust(t *testing.T, handler http.HandlerFunc) *ZeroTrust {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)

	zt := NewZeroTrust(strings.TrimPrefix(srv.URL, "https://"), "test-token", false)
	zt.apiClient.SetHTTPClient(srv.Client())
	return zt
}

func TestRequestEscapesID(t *testing.T) {
	zt := newTestZeroTrust(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/zerotrust/get-protectsurface" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected authorization header %q", r.Header.Get("Authorization"))
		}
		if q := r.URL.Query(); q.Get("id") != "a&b=c" || q.Has("b") {
			t.Errorf("id should be escaped, got query %v", q)
		}
		w.Write([]byte(`{"items":[{"id":"a&b=c","name":"P_TEST"}]}`))
	})

	ps, err := zt.GetProtectSurfaceByID(nil, "a&b=c")
	if err != nil || ps.Name != "P_TEST" {
		t.Errorf("GetProtectSurfaceByID: %+v, %v", ps, err)
	}
}

func TestRequestEscapesMatchValue(t *testing.T) {
	zt := newTestZeroTrust(t, func(w http.ResponseWriter, r *http.Request) {
		if q := r.URL.Query(); q.Get("match_type") != "ipv6" || q.Get("match_method") != "ip" || q.Get("match_value") != "fe80::1+x" {
			t.Errorf("match values should be escaped, got query %v", q)
		}
		w.Write([]byte(`{"items":[{"name":"P_TEST"}]}`))
	})

	if _, err := zt.GetProtectSurfaceByStateTypeAndValue(nil, "ipv6", "ip", "fe80::1+x"); err != nil {
		t.Errorf("GetProtectSurfaceByStateTypeAndValue: %v", err)
	}
}

func TestRequestNotOK(t *testing.T) {
	zt := newTestZeroTrust(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	if _, err := zt.GetProtectSurfaceByID(nil, "unknown"); err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("a 404 should return an error mentioning it, got %v", err)
	}
}

func TestRequestNilContextTimeout(t *testing.T) {
	zt := newTestZeroTrust(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	})
	zt.apiClient.SetTimeout(1)

	start := time.Now()
	if _, err := zt.GetLocations(nil); err == nil {
		t.Errorf("a stalled server should time out")
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("a nil context should use the default timeout of 1s, took %s", elapsed)
	}
}
