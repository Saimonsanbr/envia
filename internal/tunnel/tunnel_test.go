package tunnel

import "testing"

func TestParseCloudflareURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2026-09-08T14:00:10Z INF |  https://suspended-exemption-terrace-ing.trycloudflare.com                                 |", "https://suspended-exemption-terrace-ing.trycloudflare.com"},
		{"Visit at https://abc123.trycloudflare.com for your tunnel", "https://abc123.trycloudflare.com"},
		{"no url here", ""},
		{"mixed \x1b[32mhttps://foo-bar.trycloudflare.com\x1b[0m more text", "https://foo-bar.trycloudflare.com"},
		{"https://multiple-1.trycloudflare.com and https://multiple-2.trycloudflare.com", "https://multiple-1.trycloudflare.com"},
	}
	for _, tc := range tests {
		got := ParseCloudflareURL(tc.input)
		if got != tc.want {
			t.Errorf("ParseCloudflareURL(%q)=%q want %q", tc.input, got, tc.want)
		}
	}
}

func TestParseBoreURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"listening at bore.pub:27999", "http://bore.pub:27999"},
		{"\x1b[2m2026-09-08T14:02:20.708476Z\x1b[0m listening at bore.pub:12345", "http://bore.pub:12345"},
		{"bore.pub:8080 something", "http://bore.pub:8080"},
		{"no bore url", ""},
	}
	for _, tc := range tests {
		got := ParseBoreURL(tc.input)
		if got != tc.want {
			t.Errorf("ParseBoreURL(%q)=%q want %q", tc.input, got, tc.want)
		}
	}
}

func TestCloudflareRegexNotFixedLines(t *testing.T) {
	// Should not depend on line count
	lines := []string{
		"2026-09-08T14:00:02Z INF Thank you for trying Cloudflare Tunnel.",
		"2026-09-08T14:00:02Z INF Requesting new quick Tunnel on trycloudflare.com...",
		"2026-09-08T14:00:10Z INF +--------------------------------------------------------------------------------------------+",
		"2026-09-08T14:00:10Z INF |  Your quick Tunnel has been created! Visit it at (it may take some time to be reachable):  |",
		"2026-09-08T14:00:10Z INF |  https://suspended-exemption-terrace-ing.trycloudflare.com                                 |",
		"2026-09-08T14:00:10Z INF +--------------------------------------------------------------------------------------------+",
	}
	found := ""
	for _, line := range lines {
		if u := ParseCloudflareURL(line); u != "" {
			found = u
			break
		}
	}
	if found != "https://suspended-exemption-terrace-ing.trycloudflare.com" {
		t.Fatalf("failed to find url in varied lines got %q", found)
	}
}
