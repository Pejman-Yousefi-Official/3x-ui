package sub

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNormalizeHappRouting_Off(t *testing.T) {
	got, err := normalizeHappRouting([]byte("happ://routing/off"))
	if err != nil {
		t.Fatalf("normalizeHappRouting(off) error: %v", err)
	}
	if got != "happ://routing/off" {
		t.Fatalf("normalizeHappRouting(off) = %q, want happ://routing/off", got)
	}
}

func TestApplyCommonHeaders_HappClientHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := HappConfig{
		AutoDetect:          true,
		ProviderId:          "pid-test-123",
		NewUrl:              "https://new.example.com/sub",
		FallbackUrl:         "https://backup.example.com/sub",
		SubInfoColor:        "green",
		SubInfoText:         "Welcome to VIP Network",
		SubInfoButtonText:   "Telegram",
		SubInfoButtonLink:   "https://t.me/example",
		SubExpire:           true,
		SubExpireButtonLink: "https://renew.example.com",
		NotificationExpire:  true,
		NoLimit:             true,
		AlwaysHwid:          true,
		TunMode:             "gvisor",
		TunType:             "singbox",
		ExcludeRoutes:       "192.168.1.0/24, 10.0.0.0/8",
		ExcludeApns:         true,
		ColorProfile:        `{"serverRowBackgroundColor":"#21003D67"}`,
		PingType:            "proxy",
		AutoConnect:         true,
		AutoConnectType:     "lowestdelay",
		PerAppMode:          "on",
		PerAppList:          "com.google.chrome,com.meta.instagram",
	}

	controller := &SUBController{happConfig: cfg}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/sub/test", nil)
	ctx.Request.Header.Set("User-Agent", "Happ/1.2.0 (iPhone; iOS 17.5)")

	controller.ApplyCommonHeaders(ctx, "upload=0; download=100; total=1000; expire=1800000000", "12", "MyTitle", "", "", "", false, "", false)

	h := recorder.Header()
	if h.Get("Routing-Enable") != "0" {
		t.Fatalf("Routing-Enable = %q, want 0 for Happ with disabled routing", h.Get("Routing-Enable"))
	}
	if h.Get("Hide-Settings") != "0" {
		t.Fatalf("Hide-Settings = %q, want 0 for Happ with disabled hideSettings", h.Get("Hide-Settings"))
	}
	if h.Get("ProviderID") != "pid-test-123" {
		t.Fatalf("ProviderID = %q, want pid-test-123", h.Get("ProviderID"))
	}
	if h.Get("New-Url") != "https://new.example.com/sub" {
		t.Fatalf("New-Url = %q", h.Get("New-Url"))
	}
	if h.Get("Fallback-Url") != "https://backup.example.com/sub" {
		t.Fatalf("Fallback-Url = %q", h.Get("Fallback-Url"))
	}
	if h.Get("Sub-Info-Color") != "green" || h.Get("Sub-Info-Text") != "Welcome to VIP Network" {
		t.Fatalf("Sub-Info = %s / %s", h.Get("Sub-Info-Color"), h.Get("Sub-Info-Text"))
	}
	if h.Get("Sub-Info-Button-Text") != "Telegram" || h.Get("Sub-Info-Button-Link") != "https://t.me/example" {
		t.Fatalf("Sub-Info button = %s / %s", h.Get("Sub-Info-Button-Text"), h.Get("Sub-Info-Button-Link"))
	}
	if h.Get("Sub-Expire") != "1" || h.Get("Sub-Expire-Button-Link") != "https://renew.example.com" {
		t.Fatalf("Sub-Expire = %s / %s", h.Get("Sub-Expire"), h.Get("Sub-Expire-Button-Link"))
	}
	if h.Get("Notification-Subs-Expire") != "1" {
		t.Fatalf("Notification-Subs-Expire = %q", h.Get("Notification-Subs-Expire"))
	}
	if h.Get("No-Limit-Enabled") != "1" {
		t.Fatalf("No-Limit-Enabled = %q", h.Get("No-Limit-Enabled"))
	}
	if h.Get("Subscription-Always-Hwid-Enable") != "1" {
		t.Fatalf("Subscription-Always-Hwid-Enable = %q", h.Get("Subscription-Always-Hwid-Enable"))
	}
	if h.Get("Tun-Mode") != "gvisor" || h.Get("Tun-Type") != "singbox" {
		t.Fatalf("Tun mode/type = %s / %s", h.Get("Tun-Mode"), h.Get("Tun-Type"))
	}
	if h.Get("Exclude-Routes") != "192.168.1.0/24, 10.0.0.0/8" || h.Get("Exclude-Apns-Enable") != "true" {
		t.Fatalf("Exclude routes/apns = %s / %s", h.Get("Exclude-Routes"), h.Get("Exclude-Apns-Enable"))
	}
	if h.Get("Color-Profile") != cfg.ColorProfile {
		t.Fatalf("Color-Profile = %q", h.Get("Color-Profile"))
	}
	if h.Get("Ping-Type") != "proxy" {
		t.Fatalf("Ping-Type = %q", h.Get("Ping-Type"))
	}
	if h.Get("Subscription-Autoconnect") != "1" || h.Get("Subscription-Autoconnect-Type") != "lowestdelay" {
		t.Fatalf("Autoconnect = %s / %s", h.Get("Subscription-Autoconnect"), h.Get("Subscription-Autoconnect-Type"))
	}
	if h.Get("Per-App-Proxy-Mode") != "on" || h.Get("Per-App-Proxy-List") != "com.google.chrome,com.meta.instagram" {
		t.Fatalf("Per-App = %s / %s", h.Get("Per-App-Proxy-Mode"), h.Get("Per-App-Proxy-List"))
	}
}

func TestBuildHappPresetRouting(t *testing.T) {
	presets := []string{"iran-bypass", "china-direct", "adblock", "global"}
	for _, preset := range presets {
		link, err := BuildHappPresetRouting(preset)
		if err != nil {
			t.Fatalf("BuildHappPresetRouting(%s) error: %v", preset, err)
		}
		if !strings.HasPrefix(link, "happ://routing/onadd/") {
			t.Fatalf("BuildHappPresetRouting(%s) = %q, missing prefix", preset, link)
		}
		b64 := strings.TrimPrefix(link, "happ://routing/onadd/")
		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			t.Fatalf("failed to decode base64 for %s: %v", preset, err)
		}
		if !strings.Contains(string(decoded), `"GlobalProxy":"true"`) {
			t.Fatalf("decoded preset %s missing GlobalProxy: %s", preset, string(decoded))
		}
	}

	if _, err := BuildHappPresetRouting("invalid-preset"); err == nil {
		t.Fatalf("expected error on invalid preset")
	}
}

func TestAppendHappServerDescription(t *testing.T) {
	got := appendHappServerDescription("My Node", "VIP Server")
	wantSuffix := "?serverDescription=" + base64.StdEncoding.EncodeToString([]byte("VIP Server"))
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("appendHappServerDescription = %q, want suffix %q", got, wantSuffix)
	}

	gotExisting := appendHappServerDescription("My Node?foo=bar", "VIP Server")
	if !strings.Contains(gotExisting, "&serverDescription=") {
		t.Fatalf("appendHappServerDescription with existing query = %q", gotExisting)
	}

	if gotEmpty := appendHappServerDescription("My Node", ""); gotEmpty != "My Node" {
		t.Fatalf("appendHappServerDescription with empty desc = %q", gotEmpty)
	}
}
