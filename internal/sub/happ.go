package sub

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var happUserAgentRegex = regexp.MustCompile(`(?i)\bhapp\b`)

// HappConfig holds all Happ client customization parameters.
type HappConfig struct {
	AutoDetect          bool
	ProviderId          string
	NewUrl              string
	FallbackUrl         string
	SubInfoColor        string
	SubInfoText         string
	SubInfoButtonText   string
	SubInfoButtonLink   string
	SubExpire           bool
	SubExpireButtonLink string
	NotificationExpire  bool
	NoLimit             bool
	AlwaysHwid          bool
	TunMode             string
	TunType             string
	ExcludeRoutes       string
	ExcludeApns         bool
	ColorProfile        string
	PingType            string
	AutoConnect         bool
	AutoConnectType     string
	PerAppMode          string
	PerAppList          string
}

// IsHappClient checks if the client user-agent identifies as Happ.
func IsHappClient(userAgent string) bool {
	return happUserAgentRegex.MatchString(userAgent)
}

// ApplyHappHeaders sets standard and advanced Happ subscription headers.
func ApplyHappHeaders(c *gin.Context, cfg HappConfig, isHapp bool) {
	if cfg.ProviderId != "" {
		c.Writer.Header().Set("ProviderID", cfg.ProviderId)
	}
	if cfg.NewUrl != "" {
		c.Writer.Header().Set("New-Url", cfg.NewUrl)
	}
	if cfg.FallbackUrl != "" {
		c.Writer.Header().Set("Fallback-Url", cfg.FallbackUrl)
	}
	if text := strings.TrimSpace(cfg.SubInfoText); text != "" {
		color := strings.TrimSpace(cfg.SubInfoColor)
		switch strings.ToLower(color) {
		case "primary", "info":
			color = "blue"
		case "success":
			color = "green"
		case "warning", "danger":
			color = "red"
		case "":
			color = "blue"
		}
		c.Writer.Header().Set("Sub-Info-Color", color)
		c.Writer.Header().Set("Sub-Info-Text", text)
		if btnText := strings.TrimSpace(cfg.SubInfoButtonText); btnText != "" {
			c.Writer.Header().Set("Sub-Info-Button-Text", btnText)
		}
		if btnLink := strings.TrimSpace(cfg.SubInfoButtonLink); btnLink != "" {
			c.Writer.Header().Set("Sub-Info-Button-Link", btnLink)
		}
	}
	if cfg.SubExpire {
		c.Writer.Header().Set("Sub-Expire", "1")
		if link := strings.TrimSpace(cfg.SubExpireButtonLink); link != "" {
			c.Writer.Header().Set("Sub-Expire-Button-Link", link)
		}
	}
	if cfg.NotificationExpire {
		c.Writer.Header().Set("Notification-Subs-Expire", "1")
	}
	if cfg.NoLimit {
		c.Writer.Header().Set("No-Limit-Enabled", "1")
	}
	if cfg.AlwaysHwid {
		c.Writer.Header().Set("Subscription-Always-Hwid-Enable", "1")
	}
	if cfg.TunMode != "" {
		c.Writer.Header().Set("Tun-Mode", cfg.TunMode)
	}
	if cfg.TunType != "" {
		c.Writer.Header().Set("Tun-Type", cfg.TunType)
	}
	if routes := strings.TrimSpace(cfg.ExcludeRoutes); routes != "" {
		c.Writer.Header().Set("Exclude-Routes", routes)
	}
	if cfg.ExcludeApns {
		c.Writer.Header().Set("Exclude-Apns-Enable", "true")
	}
	if profile := strings.TrimSpace(cfg.ColorProfile); profile != "" {
		c.Writer.Header().Set("Color-Profile", profile)
	}
	if ping := strings.TrimSpace(cfg.PingType); ping != "" {
		if strings.EqualFold(ping, "http") {
			ping = "proxy"
		}
		c.Writer.Header().Set("Ping-Type", ping)
	}
	if cfg.AutoConnect {
		c.Writer.Header().Set("Subscription-Autoconnect", "1")
		autoType := strings.TrimSpace(cfg.AutoConnectType)
		switch strings.ToLower(autoType) {
		case "fastest":
			autoType = "lowestdelay"
		case "last":
			autoType = "lastused"
		}
		if autoType != "" {
			c.Writer.Header().Set("Subscription-Autoconnect-Type", autoType)
		}
	}
	if mode := strings.TrimSpace(cfg.PerAppMode); mode != "" && mode != "off" {
		switch strings.ToLower(mode) {
		case "include":
			mode = "on"
		case "exclude":
			mode = "bypass"
		}
		c.Writer.Header().Set("Per-App-Proxy-Mode", mode)
		if list := strings.TrimSpace(cfg.PerAppList); list != "" {
			c.Writer.Header().Set("Per-App-Proxy-List", list)
		}
	}
}

// BuildHappPresetRouting creates a ready-to-use happ:// routing deeplink.
func BuildHappPresetRouting(preset string) (string, error) {
	directIPs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		"224.0.0.0/4",
		"255.255.255.255",
	}
	profile := map[string]any{
		"GlobalProxy":      "true",
		"RemoteDNSType":    "DoH",
		"RemoteDNSDomain":  "https://cloudflare-dns.com/dns-query",
		"RemoteDNSIP":      "1.1.1.1",
		"DomesticDNSType":  "DoH",
		"DomesticDNSDomain": "https://dns.google/dns-query",
		"DomesticDNSIP":    "8.8.8.8",
		"Geoipurl":         "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geoip.dat",
		"Geositeurl":       "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geosite.dat",
		"DnsHosts": map[string]string{
			"cloudflare-dns.com": "1.1.1.1",
			"dns.google":         "8.8.8.8",
		},
		"DirectSites":    []string{},
		"DirectIp":       directIPs,
		"ProxySites":     []string{},
		"ProxyIp":        []string{},
		"BlockSites":     []string{"geosite:ads"},
		"BlockIp":        []string{"geoip:ads"},
		"DomainStrategy": "IPIfNonMatch",
		"FakeDNS":        "false",
	}

	switch preset {
	case "iran-bypass":
		profile["Name"] = "Iran Bypass"
		profile["DirectSites"] = []string{"geosite:ir"}
		profile["DirectIp"] = append([]string{"geoip:ir"}, directIPs...)
	case "china-direct":
		profile["Name"] = "China Direct"
		profile["DirectSites"] = []string{"geosite:cn", "geosite:geolocation-cn"}
		profile["DirectIp"] = append([]string{"geoip:cn"}, directIPs...)
	case "adblock":
		profile["Name"] = "AdBlock"
	case "global":
		profile["Name"] = "Global"
		profile["BlockSites"] = []string{}
		profile["BlockIp"] = []string{}
	default:
		return "", errors.New("unknown Happ routing preset: " + preset)
	}

	compact, err := json.Marshal(profile)
	if err != nil {
		return "", err
	}
	return "happ://routing/onadd/" + base64.StdEncoding.EncodeToString(compact), nil
}
