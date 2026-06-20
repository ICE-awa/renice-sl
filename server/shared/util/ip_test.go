package util

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

// 回环/私有网段/CGNAT/link-local/multicast 测试
func TestValidatePublicIP_BlockedRanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ip   string
	}{
		// loopback
		{"IPv4 loopback", "127.0.0.1"},
		{"IPv6 loopback", "::1"},

		// RFC 1918 私有网段
		{"10.x.x.x", "10.0.0.1"},
		{"172.16.x.x", "172.16.0.1"},
		{"192.168.x.x", "192.168.1.1"},

		// CGNAT (100.64.0.0/10)
		{"CGNAT low", "100.64.0.1"},
		{"CGNAT high", "100.127.255.254"},

		// link-local
		{"IPv4 link-local", "169.254.1.1"},
		{"IPv6 link-local", "fe80::1"},

		// Multicast
		{"Ipv4 multicast", "224.0.0.1"},
		{"IPv6 multicast", "ff02::1"},

		// 未指定地址
		{"IPv4 unspecified", "0.0.0.0"},
		{"IPv6 unspecified", "::"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ip := netip.MustParseAddr(tt.ip)
			if err := validatePublicIP(ip); err == nil {
				t.Errorf("validatePublicIP(%s) should have failed", tt.ip)
			}
		})
	}
}

// 正常公网 ip 放行
func TestValidatePublicIP_AllowedRanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ip   string
	}{
		{"Google DNS", "8.8.8.8"},
		{"Cloudflare", "1.1.1.1"},
		{"Random public", "203.0.113.1"},
		{"IPv6 public", "2001:db8::1"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ip := netip.MustParseAddr(tt.ip)
			if err := validatePublicIP(ip); err != nil {
				t.Errorf("validatePublicIP(%s) = %v, want nil", tt.ip, err)
			}
		})
	}
}

// 检验无效输入（空串，没有 scheme，非 http/https，空 host，localhost
func TestNormalizeAndValidateURL_InvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"空字符串", "", ErrEmptyURL},
		{"纯空格", "   ", ErrEmptyURL},
		{"ftp scheme", "ftp://example.com", ErrInvalidScheme},
		{"javascript scheme", "javascript:alert(1)", ErrInvalidScheme},
		{"localhost", "http://localhost/path", ErrLocalhostNotAllowed},
		{"sub.localhost", "http://api.localhost:3000", ErrLocalhostNotAllowed},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NormalizeAndValidateURL(context.Background(), tt.input)
			if err == nil {
				t.Fatalf("NormalizeAndValidateURL(%q) = nil error, want %v", tt.input, tt.wantErr)
			}
			t.Logf("got expected error: %v", err)
		})
	}
}

// 解析后指向内网 ip 的域名被拒绝
func TestNormalizeAndValidateURL_PrivateHosts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"loopback IP", "http://127.0.0.1/admin"},
		{"private 10.x", "http://10.0.0.1:8080/api"},
		{"private 172.16.x", "https://172.16.0.1/"},
		{"private 192.168.x", "http://192.168.1.1/"},
		{"CGNAT", "http://100.64.0.1/metadata"},
		{"link-local", "http://169.254.169.254/latest/meta-data/"},
		{"IPv6 loopback", "http://[::1]:8080/"},

		{"AWS metadata", "http://169.254.169.254/latest/meta-data/iam/"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NormalizeAndValidateURL(context.Background(), tt.input)
			if err == nil {
				t.Errorf("NormalizeAndValidateURL(%q) = nil, want error (should block private IP)",
					tt.input)
			}
		})
	}
}

// 正常公网 URL
func TestNormalizeAndValidateURL_ValidURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"标准 https", "https://github.com"},
		{"带路径", "https://example.com/path/to/page"},
		{"带端口", "https://example.com:8443/api"},
		{"无 scheme 自动补全", "example.com/hello"},
		{"// 开头补全", "//example.com/world"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := NormalizeAndValidateURL(context.Background(), tt.input)
			if err != nil {
				t.Errorf("NormalizeAndValidateURL(%q) = error %v, want success",
					tt.input, err)
			}
			if result == "" {
				t.Error("result is empty string")
			}
			t.Logf("normalized: %s", result)
		})
	}
}

// 公网 302 到内网 ip 拦截
func TestCheckURLReachable_RedirectToPrivate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://10.0.0.1/evil", http.StatusFound)
	}))
	defer server.Close()

	_, err := CheckURLReachable(context.Background(), server.URL)
	if err == nil {
		t.Fatal("CheckURLReachable should fail when redirect targets private IP")
	}
	t.Logf("got expected error: %v", err)
}

// 重定向超过 5 跳被拒绝
func TestCheckURLReachable_TooManyRedirects(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/again", http.StatusFound)
	}))
	defer server.Close()

	_, err := CheckURLReachable(context.Background(), server.URL)
	if err == nil {
		t.Fatal("CheckURLReachable should fail after too many redirects")
	}
	t.Logf("got expected error: %v", err)
}
