package utils

import (
	"net"
	"testing"
)

func TestIsPrivateIP(t *testing.T) {
	private := []string{
		"127.0.0.1",
		"10.1.2.3",
		"172.16.0.1",
		"192.168.1.1",
		"169.254.1.1",
		"100.64.0.1",
		"::1",
		"fe80::1",
		"fc00::1",
		"0.0.0.0",
	}
	for _, ip := range private {
		if !isPrivateIP(ip) {
			t.Errorf("isPrivateIP(%q) = false, want true", ip)
		}
	}

	public := []string{"8.8.8.8", "1.1.1.1", "2001:4860:4860::8888"}
	for _, ip := range public {
		if isPrivateIP(ip) {
			t.Errorf("isPrivateIP(%q) = true, want false", ip)
		}
	}

	if isPrivateIP("not-an-ip") {
		t.Error(`isPrivateIP("not-an-ip") = true, want false`)
	}
}

func TestResolveAddr(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"127.0.0.1", "127.0.0.1"},
		{"127.0.0.1:7000", "127.0.0.1"},
		{"2001:db8::1", "2001:db8::1"},
		{"[2001:db8::1]:7000", "2001:db8::1"},
		{"8.8.8.8", "8.8.8.8"},
	}
	for _, c := range cases {
		got, err := resolveAddr(c.in)
		if err != nil {
			t.Errorf("resolveAddr(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("resolveAddr(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPingHostEmpty(t *testing.T) {
	res := PingHost("", 0)
	if res.Reachable {
		t.Error("empty address should not be reachable")
	}
	if res.Latency != -1 {
		t.Errorf("Latency = %d, want -1", res.Latency)
	}
}

func TestPingHostIPv6PortHandling(t *testing.T) {
	// Bracketed IPv6 with port must parse without double-appending :7000
	ln, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Skip("IPv6 loopback unavailable")
	}
	defer func() { _ = ln.Close() }()
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			_ = conn.Close()
		}
	}()

	addr := ln.Addr().String()
	res := PingHost(addr, 2e9)
	if !res.Reachable {
		t.Fatalf("PingHost(%q) not reachable: %s", addr, res.Error)
	}
}
