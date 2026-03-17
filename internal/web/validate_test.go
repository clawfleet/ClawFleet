package web

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func slackJSONResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func withTestHTTPClient(t *testing.T, fn roundTripperFunc) {
	t.Helper()
	original := httpClient
	httpClient = &http.Client{Transport: fn}
	t.Cleanup(func() {
		httpClient = original
	})
}

func TestValidateChannelTokenSlackSuccess(t *testing.T) {
	withTestHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case "https://slack.com/api/auth.test":
			if got := req.Header.Get("Authorization"); got != "Bearer xoxb-valid" {
				t.Fatalf("unexpected bot token auth header: %q", got)
			}
			return slackJSONResponse(http.StatusOK, `{"ok":true,"user":"Claw Bot"}`), nil
		case "https://slack.com/api/apps.connections.open":
			if got := req.Header.Get("Authorization"); got != "Bearer xapp-valid" {
				t.Fatalf("unexpected app token auth header: %q", got)
			}
			return slackJSONResponse(http.StatusOK, `{"ok":true,"url":"wss://example"}`), nil
		default:
			t.Fatalf("unexpected request URL: %s", req.URL.String())
			return nil, nil
		}
	})

	if err := ValidateChannelToken("slack", "xoxb-valid", "xapp-valid", "", ""); err != nil {
		t.Fatalf("ValidateChannelToken returned error: %v", err)
	}
}

func TestValidateChannelTokenSlackRequiresAppToken(t *testing.T) {
	withTestHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected network request: %s", req.URL.String())
		return nil, nil
	})

	err := ValidateChannelToken("slack", "xoxb-valid", "", "", "")
	if err == nil || !strings.Contains(err.Error(), "Slack app token is required") {
		t.Fatalf("expected missing app token error, got: %v", err)
	}
}

func TestValidateChannelTokenSlackInvalidAppToken(t *testing.T) {
	withTestHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case "https://slack.com/api/auth.test":
			return slackJSONResponse(http.StatusOK, `{"ok":true,"user":"Claw Bot"}`), nil
		case "https://slack.com/api/apps.connections.open":
			return slackJSONResponse(http.StatusOK, `{"ok":false,"error":"invalid_auth"}`), nil
		default:
			t.Fatalf("unexpected request URL: %s", req.URL.String())
			return nil, nil
		}
	})

	err := ValidateChannelToken("slack", "xoxb-valid", "xapp-invalid", "", "")
	if err == nil || !strings.Contains(err.Error(), "invalid Slack app token or Socket Mode is not enabled: invalid_auth") {
		t.Fatalf("expected invalid app token error, got: %v", err)
	}
}
