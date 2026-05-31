// Package ssrf provides Worker outbound URL and dial guards against SSRF.
package ssrf

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
	"github.com/imroc/req/v3"
	"golang.org/x/net/idna"
)

type resolver interface {
	LookupNetIP(ctx context.Context, network string, host string) ([]netip.Addr, error)
}

type dialer interface {
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)
}

// Guard validates outbound URLs, redirects, DNS results, and dial targets.
type Guard struct {
	logger  *slog.Logger
	policy  Policy
	resolve resolver
	dial    dialer
}

// Option configures a Guard.
type Option func(*Guard)

func New(logger *slog.Logger, opts ...Option) *Guard {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	g := &Guard{
		logger:  logger,
		policy:  DefaultPolicy(),
		resolve: net.DefaultResolver,
		dial:    &net.Dialer{},
	}
	for _, opt := range opts {
		opt(g)
	}

	return g
}

func WithPolicy(policy Policy) Option {
	return func(g *Guard) {
		if policy.MaxRedirects == 0 {
			policy.MaxRedirects = DefaultPolicy().MaxRedirects
		}

		g.policy = policy
	}
}

func WithResolver(resolve resolver) Option {
	return func(g *Guard) {
		if resolve != nil {
			g.resolve = resolve
		}
	}
}

func WithDialer(dial dialer) Option {
	return func(g *Guard) {
		if dial != nil {
			g.dial = dial
		}
	}
}

// RequestMiddleware validates the initial request URL before req sends it.
func (g *Guard) RequestMiddleware() req.RequestMiddleware {
	return func(_ *req.Client, request *req.Request) error {
		_, err := g.ValidateURL(request.RawURL, PhaseInitialURL)

		return err
	}
}

// RedirectPolicy validates redirects and enforces the configured redirect limit.
func (g *Guard) RedirectPolicy() req.RedirectPolicy {
	return func(request *http.Request, via []*http.Request) error {
		ctx := request.Context()

		if len(via) > g.policy.MaxRedirects {
			rawURL := ""
			if request.URL != nil {
				rawURL = request.URL.String()
			}

			target := targetFromURL(request.URL)
			err := policyBlock(PhaseRedirect, ReasonRedirectLimit, rawURL, target)
			g.logDecision(ctx, PhaseRedirect, false, ReasonRedirectLimit, rawURL, target, nil, netip.Addr{}, err)

			return err
		}

		_, err := g.ValidateURLContext(ctx, request.URL.String(), PhaseRedirect)

		return err
	}
}

// DialContext validates DNS answers at dial time and dials only an approved address.
func (g *Guard) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, splitErr := net.SplitHostPort(address)
	if splitErr != nil {
		target := target{Host: address}
		err := resolutionFailure(PhaseDial, ReasonEmptyHost, "", target, splitErr)
		g.logDecision(ctx, PhaseDial, false, ReasonEmptyHost, "", target, nil, netip.Addr{}, err)

		return nil, err
	}

	host = normalizeAddressHost(host)

	target := target{Host: host, Port: port}
	if port != httpsPort {
		err := policyBlock(PhaseDial, ReasonNon443Port, "", target)
		g.logDecision(ctx, PhaseDial, false, ReasonNon443Port, "", target, nil, netip.Addr{}, err)

		return nil, err
	}

	ips, lookupErr := g.resolve.LookupNetIP(ctx, "ip", host)
	if lookupErr != nil {
		err := resolutionFailure(PhaseDial, ReasonDNSResolutionFailed, "", target, lookupErr)
		g.logDecision(ctx, PhaseDial, false, ReasonDNSResolutionFailed, "", target, nil, netip.Addr{}, err)

		return nil, err
	}

	if len(ips) == 0 {
		err := resolutionFailure(PhaseDial, ReasonNoResolvedIPs, "", target, errors.New("no DNS results"))
		g.logDecision(ctx, PhaseDial, false, ReasonNoResolvedIPs, "", target, nil, netip.Addr{}, err)

		return nil, err
	}

	for _, ip := range ips {
		if deniedIP(ip, g.policy) {
			err := policyBlock(PhaseDial, ReasonDeniedIP, "", target, withResolvedIPs(ips), withBlockedIP(ip))
			g.logDecision(ctx, PhaseDial, false, ReasonDeniedIP, "", target, ips, ip, err)

			return nil, err
		}
	}

	g.logDecision(ctx, PhaseDial, true, ReasonAllowed, "", target, ips, netip.Addr{}, nil)

	conn, dialErr := g.dial.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
	if dialErr != nil {
		return nil, fmt.Errorf("ssrf: dial approved address: %w", dialErr)
	}

	return conn, nil
}

func (g *Guard) ValidateURL(rawURL string, phase Phase) (*url.URL, error) {
	return g.ValidateURLContext(context.Background(), rawURL, phase)
}

// ValidateURLContext validates rawURL for the given SSRF policy phase.
func (g *Guard) ValidateURLContext(ctx context.Context, rawURL string, phase Phase) (*url.URL, error) {
	parsed, parseErr := url.Parse(rawURL)
	if parseErr != nil {
		err := resolutionFailure(phase, ReasonInvalidHostname, rawURL, target{}, parseErr)
		g.logDecision(ctx, phase, false, ReasonInvalidHostname, rawURL, target{}, nil, netip.Addr{}, err)

		return nil, err
	}

	target := targetFromURL(parsed)

	normalized, validateErr := g.validateTarget(ctx, rawURL, phase, parsed, target)
	if validateErr != nil {
		return nil, validateErr
	}

	parsed.Scheme = "https"

	parsed.Host = net.JoinHostPort(normalized, httpsPort)
	if parsed.Port() == httpsPort && !hasExplicitPort(rawURL) {
		parsed.Host = normalized
	}

	target.Host = normalized
	target.Port = httpsPort
	g.logDecision(ctx, phase, true, ReasonAllowed, rawURL, target, nil, netip.Addr{}, nil)

	return parsed, nil
}

func (g *Guard) validateTarget(ctx context.Context, rawURL string, phase Phase, parsed *url.URL, target target) (string, error) {
	shapeErr := g.validateURLShape(ctx, rawURL, phase, parsed, target)
	if shapeErr != nil {
		return "", shapeErr
	}

	return g.validateHostname(ctx, rawURL, phase, parsed, target)
}

func (g *Guard) validateURLShape(ctx context.Context, rawURL string, phase Phase, parsed *url.URL, target target) error {
	if !parsed.IsAbs() {
		return g.blockURL(ctx, phase, ReasonRelativeURL, rawURL, target)
	}

	if !strings.EqualFold(parsed.Scheme, "https") {
		return g.blockURL(ctx, phase, ReasonNotHTTPS, rawURL, target)
	}

	if parsed.User != nil {
		return g.blockURL(ctx, phase, ReasonUserinfo, rawURL, target)
	}

	if parsed.Hostname() == "" {
		return g.blockURL(ctx, phase, ReasonEmptyHost, rawURL, target)
	}

	return nil
}

func (g *Guard) validateHostname(ctx context.Context, rawURL string, phase Phase, parsed *url.URL, target target) (string, error) {
	host := parsed.Hostname()
	normalizedRaw := normalizeAddressHost(host)
	target.Host = normalizedRaw

	_, parseErr := netip.ParseAddr(normalizedRaw)
	if parseErr == nil {
		return "", g.blockURL(ctx, phase, ReasonIPLiteral, rawURL, target)
	}

	normalized, normalizeErr := normalizeHostname(host)
	if normalizeErr != nil {
		return "", g.blockURL(ctx, phase, ReasonInvalidHostname, rawURL, target, func(e *Error) {
			e.Err = errors.Join(e.Err, normalizeErr)
		})
	}

	target.Host = normalized
	if port := parsed.Port(); port != "" && port != httpsPort {
		return "", g.blockURL(ctx, phase, ReasonNon443Port, rawURL, target)
	}

	target.Port = httpsPort

	if !validDomainName(normalized) {
		return "", g.blockURL(ctx, phase, ReasonInvalidHostname, rawURL, target)
	}

	if !strings.Contains(normalized, ".") {
		return "", g.blockURL(ctx, phase, ReasonSingleLabelHost, rawURL, target)
	}

	if deniedHostname(normalized, g.policy) {
		return "", g.blockURL(ctx, phase, ReasonDeniedHostname, rawURL, target)
	}

	return normalized, nil
}

func (g *Guard) blockURL(
	ctx context.Context,
	phase Phase,
	reason Reason,
	rawURL string,
	target target,
	opts ...func(*Error),
) error {
	err := policyBlock(phase, reason, rawURL, target, opts...)
	g.logDecision(ctx, phase, false, reason, rawURL, target, nil, netip.Addr{}, err)

	return err
}

type target struct {
	Host string
	Port string
}

func targetFromURL(parsed *url.URL) target {
	if parsed == nil {
		return target{}
	}

	port := parsed.Port()
	if port == "" && strings.EqualFold(parsed.Scheme, "https") && parsed.Hostname() != "" {
		port = httpsPort
	}

	return target{
		Host: normalizeAddressHost(parsed.Hostname()),
		Port: port,
	}
}

func normalizeHostname(host string) (string, error) {
	host = normalizeAddressHost(host)

	ascii, err := idna.Lookup.ToASCII(host)
	if err != nil {
		return "", fmt.Errorf("ssrf: normalize hostname: %w", err)
	}

	ascii = strings.TrimSuffix(strings.ToLower(ascii), ".")
	if ascii == "" {
		return "", errors.New("empty hostname")
	}

	return ascii, nil
}

func normalizeAddressHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

func validDomainName(host string) bool {
	if host == "" || len(host) > 253 {
		return false
	}

	labels := strings.SplitSeq(host, ".")
	for label := range labels {
		if !validDomainLabel(label) {
			return false
		}
	}

	return true
}

func validDomainLabel(label string) bool {
	if label == "" || len(label) > 63 {
		return false
	}

	if label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}

	for _, r := range label {
		if !validDomainRune(r) {
			return false
		}
	}

	return true
}

func validDomainRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= '0' && r <= '9':
		return true
	case r == '-':
		return true
	default:
		return false
	}
}

func hasExplicitPort(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	return parsed.Port() != ""
}

func (g *Guard) logDecision(
	ctx context.Context,
	phase Phase,
	allowed bool,
	reason Reason,
	rawURL string,
	target target,
	resolvedIPs []netip.Addr,
	blockedIP netip.Addr,
	decisionErr error,
) {
	level := slog.LevelDebug
	decision := "allow"

	if !allowed {
		level = slog.LevelWarn
		decision = "block"
	}

	attrs := []slog.Attr{
		slog.String("event", "ssrf_decision"),
		slog.String("phase", string(phase)),
		slog.String("decision", decision),
		slog.String("reason", string(reason)),
		slog.String("host", target.Host),
		slog.String("port", target.Port),
	}

	if !allowed {
		if code, ok := arc.CodeOf(decisionErr); ok {
			attrs = append(attrs, slog.String("arc_code", string(code)))
		}
	}

	if rawURL != "" {
		attrs = append(attrs,
			slog.String("url_redacted", redactURL(rawURL)),
			slog.String("url_hash", hashURL(rawURL)),
		)
	}

	if len(resolvedIPs) > 0 {
		attrs = append(attrs, slog.Any("resolved_ips", ipStrings(resolvedIPs)))
	}

	if blockedIP.IsValid() {
		attrs = append(attrs, slog.String("blocked_ip", blockedIP.String()))
	}

	g.logger.LogAttrs(ctx, level, "ssrf policy decision", attrs...)
}

func redactURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String()
}

func hashURL(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))

	return hex.EncodeToString(sum[:])
}

func ipStrings(ips []netip.Addr) []string {
	values := make([]string, 0, len(ips))
	for _, ip := range ips {
		values = append(values, ip.String())
	}

	return values
}
