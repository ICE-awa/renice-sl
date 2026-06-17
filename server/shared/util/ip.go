package util

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

var (
	ErrNoPublicIP          = errors.New("private/internal ip not allowed")
	ErrBlockedIPRange      = errors.New("blocked ip range")
	ErrNoIP                = errors.New("host has no ip")
	ErrEmptyURL            = errors.New("empty url")
	ErrInvalidScheme       = errors.New("only http and https schemes are allowed")
	ErrEmptyHost           = errors.New("empty host")
	ErrLocalhostNotAllowed = errors.New("localhost is not allowed")
	ErrBlockedHost         = errors.New("blocked host")
)

func mustPrefix(s string) netip.Prefix {
	p, err := netip.ParsePrefix(s)
	if err != nil {
		panic(err)
	}
	return p
}

var blockedPrefixes = []netip.Prefix{
	mustPrefix("100.64.0.0/10"),
	mustPrefix("169.254.0.0/16"),
}

func validatePublicIP(ip netip.Addr) error {
	ip = ip.Unmap()

	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() {
		return ErrNoPublicIP
	}

	for _, p := range blockedPrefixes {
		if p.Contains(ip) {
			return ErrBlockedIPRange
		}
	}

	return nil
}

func validatePublicHost(ctx context.Context, host string) error {
	if ip, err := netip.ParseAddr(host); err == nil {
		return validatePublicIP(ip)
	}

	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return err
	}

	if len(ips) == 0 {
		return ErrNoIP
	}

	for _, ip := range ips {
		if err := validatePublicIP(ip); err != nil {
			return ErrBlockedIPRange
		}
	}

	return nil
}

func NormalizeAndValidateURL(ctx context.Context, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrEmptyURL
	}

	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	} else if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", ErrInvalidScheme
	}

	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if host == "" {
		return "", ErrEmptyHost
	}

	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return "", ErrLocalhostNotAllowed
	}

	if err := validatePublicHost(ctx, host); err != nil {
		return "", err
	}

	return u.String(), nil
}

func CheckURLReachable(ctx context.Context, rawURL string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			if err := validatePublicHost(req.Context(), req.URL.Hostname()); err != nil {
				return ErrBlockedHost
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; renice-sl/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusMethodNotAllowed {
		getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return "", fmt.Errorf("build GET request: %w", err)
		}
		getReq.Header.Set("User-Agent", "Mozilla/5.0 (compatible; renice-sl/1.0)")

		getResp, err := client.Do(getReq)
		if err != nil {
			return "", fmt.Errorf("GET request failed: %w", err)
		}
		getResp.Body.Close()
		return getResp.Request.URL.String(), nil
	}

	return resp.Request.URL.String(), nil
}
