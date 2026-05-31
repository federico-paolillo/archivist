package ssrf

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"testing"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateURLAllowsHTTPSWithoutExplicitPort(t *testing.T) {
	guard, _ := newTestGuard()

	parsed, err := guard.ValidateURL("https://example.com/path", PhaseInitialURL)

	require.NoError(t, err)
	require.Equal(t, "https", parsed.Scheme)
	require.Equal(t, "example.com", parsed.Host)
}

func TestValidateURLAllowsExplicitPort443(t *testing.T) {
	guard, _ := newTestGuard()

	parsed, err := guard.ValidateURL("https://example.com:443/path", PhaseInitialURL)

	require.NoError(t, err)
	require.Equal(t, "example.com:443", parsed.Host)
}

func TestValidateURLRejectsSuspiciousTargets(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		reason Reason
	}{
		{name: "http", rawURL: "http://example.com/path", reason: ReasonNotHTTPS},
		{name: "file", rawURL: "file:///etc/passwd", reason: ReasonNotHTTPS},
		{name: "relative", rawURL: "/article", reason: ReasonRelativeURL},
		{name: "userinfo", rawURL: "https://user:secret@example.com/path", reason: ReasonUserinfo},
		{name: "non 443 port", rawURL: "https://example.com:8443/path", reason: ReasonNon443Port},
		{name: "ipv4 literal", rawURL: "https://127.0.0.1/path", reason: ReasonIPLiteral},
		{name: "ipv6 literal", rawURL: "https://[::1]/path", reason: ReasonIPLiteral},
		{name: "single label", rawURL: "https://worker/path", reason: ReasonSingleLabelHost},
		{name: "localhost", rawURL: "https://localhost/path", reason: ReasonSingleLabelHost},
		{name: "localhost suffix", rawURL: "https://foo.localhost/path", reason: ReasonDeniedHostname},
		{name: "docker host", rawURL: "https://host.docker.internal/path", reason: ReasonDeniedHostname},
		{name: "metadata host", rawURL: "https://metadata.google.internal/path", reason: ReasonDeniedHostname},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guard, _ := newTestGuard()

			_, err := guard.ValidateURL(tt.rawURL, PhaseInitialURL)

			require.ErrorIs(t, err, arc.ErrSSRFDetected)
			ssrfErr := assertSSRFFailure(t, err)
			assert.Equal(t, tt.reason, ssrfErr.Reason)
		})
	}
}

func TestDialContextAllowsOnlyPublicResolvedIPs(t *testing.T) {
	resolver := fakeResolver{ips: map[string][]netip.Addr{
		"example.com": {netip.MustParseAddr("93.184.216.34")},
	}}
	dialer := &fakeDialer{}
	guard, _ := newTestGuard(WithResolver(resolver), WithDialer(dialer))

	conn, err := guard.DialContext(t.Context(), "tcp", "example.com:443")
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	assert.Equal(t, "93.184.216.34:443", dialer.address)
}

func TestDialContextRejectsMixedDNSAnswers(t *testing.T) {
	resolver := fakeResolver{ips: map[string][]netip.Addr{
		"example.com": {
			netip.MustParseAddr("93.184.216.34"),
			netip.MustParseAddr("10.0.0.10"),
		},
	}}
	guard, _ := newTestGuard(WithResolver(resolver))

	_, err := guard.DialContext(t.Context(), "tcp", "example.com:443")

	require.ErrorIs(t, err, arc.ErrSSRFDetected)
	ssrfErr := assertSSRFFailure(t, err)
	assert.Equal(t, ReasonDeniedIP, ssrfErr.Reason)
	assert.Equal(t, "10.0.0.10", ssrfErr.BlockedIP.String())
}

func TestDialContextDNSFailureMapsToARC001(t *testing.T) {
	resolver := fakeResolver{err: errors.New("no such host")}
	guard, _ := newTestGuard(WithResolver(resolver))

	_, err := guard.DialContext(t.Context(), "tcp", "example.com:443")

	require.ErrorIs(t, err, arc.ErrURLResolutionFailed)
	require.NotErrorIs(t, err, arc.ErrSSRFDetected)
	ssrfErr := assertSSRFFailure(t, err)
	assert.Equal(t, ReasonDNSResolutionFailed, ssrfErr.Reason)
}

func TestRedirectPolicyAllowsOneRedirectAndRejectsSecond(t *testing.T) {
	guard, _ := newTestGuard()
	redirectPolicy := guard.RedirectPolicy()
	redirectRequest, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.com/next", nil)
	require.NoError(t, err)
	firstRequest, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://source.example/start", nil)
	require.NoError(t, err)

	require.NoError(t, redirectPolicy(redirectRequest, []*http.Request{firstRequest}))

	secondRequest, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://example.com/final", nil)
	require.NoError(t, err)
	err = redirectPolicy(secondRequest, []*http.Request{firstRequest, redirectRequest})

	require.ErrorIs(t, err, arc.ErrSSRFDetected)
	ssrfErr := assertSSRFFailure(t, err)
	assert.Equal(t, ReasonRedirectLimit, ssrfErr.Reason)
}

func TestLogDecisionRedactsSensitiveURLParts(t *testing.T) {
	guard, logs := newTestGuard()

	_, err := guard.ValidateURL("https://user:secret@example.com/path?token=secret#frag", PhaseInitialURL)

	require.ErrorIs(t, err, arc.ErrSSRFDetected)
	output := logs.String()
	assert.Contains(t, output, `"level":"WARN"`)
	assert.Contains(t, output, `"decision":"block"`)
	assert.Contains(t, output, `"arc_code":"ARC-017"`)
	assert.Contains(t, output, `"url_redacted":"https://example.com/path"`)
	assert.NotContains(t, output, "user:secret")
	assert.NotContains(t, output, "token=secret")
	assert.NotContains(t, output, "frag")
}

func TestLogDecisionUsesDebugForAllow(t *testing.T) {
	guard, logs := newTestGuard()

	_, err := guard.ValidateURL("https://example.com/path", PhaseInitialURL)

	require.NoError(t, err)
	output := logs.String()
	assert.Contains(t, output, `"level":"DEBUG"`)
	assert.Contains(t, output, `"decision":"allow"`)
	assert.NotContains(t, output, "arc_code")
}

func TestLogDecisionUsesARC001ForDNSFailure(t *testing.T) {
	resolver := fakeResolver{err: errors.New("no such host")}
	guard, logs := newTestGuard(WithResolver(resolver))

	_, err := guard.DialContext(t.Context(), "tcp", "example.com:443")

	require.ErrorIs(t, err, arc.ErrURLResolutionFailed)
	output := logs.String()
	assert.Contains(t, output, `"level":"WARN"`)
	assert.Contains(t, output, `"decision":"block"`)
	assert.Contains(t, output, `"reason":"dns_resolution_failed"`)
	assert.Contains(t, output, `"arc_code":"ARC-001"`)
}

func newTestGuard(opts ...Option) (*Guard, *bytes.Buffer) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	guard := New(logger, opts...)

	return guard, &logs
}

func assertSSRFFailure(t *testing.T, err error) *Error {
	t.Helper()

	ssrfErr, ok := errors.AsType[*Error](err)
	require.True(t, ok)

	return ssrfErr
}

type fakeResolver struct {
	ips map[string][]netip.Addr
	err error
}

func (r fakeResolver) LookupNetIP(_ context.Context, _ string, host string) ([]netip.Addr, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.ips[host], nil
}

type fakeDialer struct {
	address string
}

func (d *fakeDialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	d.address = address
	left, right := net.Pipe()
	_ = right.Close()

	return left, nil
}
