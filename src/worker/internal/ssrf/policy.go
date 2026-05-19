package ssrf

import (
	"net/netip"
	"slices"
	"strings"
)

const httpsPort = "443"

// Policy defines the non-weakenable SSRF defaults plus additive deny entries.
type Policy struct {
	ExtraDeniedHostnames []string
	ExtraDeniedCIDRs     []netip.Prefix
	MaxRedirects         int
}

func DefaultPolicy() Policy {
	return Policy{
		MaxRedirects: 1,
	}
}

func deniedHostname(host string, policy Policy) bool {
	switch {
	case host == "localhost":
		return true
	case strings.HasSuffix(host, ".localhost"):
		return true
	case host == "host.docker.internal":
		return true
	case host == "gateway.docker.internal":
		return true
	case host == "metadata.google.internal":
		return true
	case host == "169.254.169.254.nip.io":
		return true
	}

	return slices.Contains(policy.ExtraDeniedHostnames, host)
}

func deniedIP(ip netip.Addr, policy Policy) bool {
	if ip.Is4In6() {
		unmapped := ip.Unmap()
		if deniedIP(unmapped, policy) {
			return true
		}
	}

	for _, prefix := range defaultDeniedCIDRs {
		if prefix.Contains(ip) {
			return true
		}
	}

	for _, prefix := range policy.ExtraDeniedCIDRs {
		if prefix.Contains(ip) {
			return true
		}
	}

	return false
}

var defaultDeniedCIDRs = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("::ffff:0:0/96"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
}
