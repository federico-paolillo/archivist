package ssrf

import (
	"fmt"
	"net/netip"
	"strings"

	"codeberg.org/federico-paolillo/archivist/internal/arc"
)

// Phase identifies which SSRF policy boundary evaluated a target.
type Phase string

const (
	PhaseInitialURL Phase = "initial_url"
	PhaseRedirect   Phase = "redirect"
	PhaseDial       Phase = "dial"
)

// Reason identifies the policy decision reason used in diagnostics and logs.
type Reason string

const (
	ReasonAllowed             Reason = "allowed"
	ReasonRelativeURL         Reason = "relative_url"
	ReasonNotHTTPS            Reason = "not_https"
	ReasonUserinfo            Reason = "userinfo"
	ReasonEmptyHost           Reason = "empty_host"
	ReasonInvalidHostname     Reason = "invalid_hostname"
	ReasonNon443Port          Reason = "non_443_port"
	ReasonIPLiteral           Reason = "ip_literal"
	ReasonSingleLabelHost     Reason = "single_label_host"
	ReasonDeniedHostname      Reason = "denied_hostname"
	ReasonDNSResolutionFailed Reason = "dns_resolution_failed"
	ReasonDeniedIP            Reason = "denied_ip"
	ReasonNoResolvedIPs       Reason = "no_resolved_ips"
	ReasonRedirectLimit       Reason = "redirect_limit"
)

// Error carries SSRF decision diagnostics while unwrapping to the canonical ARC sentinel.
type Error struct {
	Phase       Phase
	Reason      Reason
	URL         string
	Host        string
	Port        string
	ResolvedIPs []netip.Addr
	BlockedIP   netip.Addr
	Err         error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("ssrf")

	if e.Phase != "" {
		b.WriteString(": phase=")
		b.WriteString(string(e.Phase))
	}

	if e.Reason != "" {
		b.WriteString(": reason=")
		b.WriteString(string(e.Reason))
	}

	if e.URL != "" {
		b.WriteString(": url=")
		b.WriteString(e.URL)
	}

	if e.Host != "" {
		b.WriteString(": host=")
		b.WriteString(e.Host)
	}

	if e.Port != "" {
		b.WriteString(": port=")
		b.WriteString(e.Port)
	}

	if e.BlockedIP.IsValid() {
		b.WriteString(": blocked_ip=")
		b.WriteString(e.BlockedIP.String())
	}

	if e.Err != nil {
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
	}

	return b.String()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func policyBlock(phase Phase, reason Reason, rawURL string, target target, opts ...func(*Error)) error {
	err := &Error{
		Phase:  phase,
		Reason: reason,
		URL:    redactURL(rawURL),
		Host:   target.Host,
		Port:   target.Port,
		Err:    arc.ErrSSRFDetected,
	}
	for _, opt := range opts {
		opt(err)
	}

	return err
}

func resolutionFailure(phase Phase, reason Reason, rawURL string, target target, cause error) error {
	return &Error{
		Phase:  phase,
		Reason: reason,
		URL:    redactURL(rawURL),
		Host:   target.Host,
		Port:   target.Port,
		Err:    fmt.Errorf("%w: %w", arc.ErrURLResolutionFailed, cause),
	}
}

func withResolvedIPs(ips []netip.Addr) func(*Error) {
	return func(err *Error) {
		err.ResolvedIPs = ips
	}
}

func withBlockedIP(ip netip.Addr) func(*Error) {
	return func(err *Error) {
		err.BlockedIP = ip
	}
}
