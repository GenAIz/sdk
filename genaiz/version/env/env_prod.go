//go:build prod

package env

import "strings"

func DefaultProtocolPrefix(host string) string {
	return "https:/"
}

func GetVersionTag() string {
	return ""
}

func IsAllowedProtocol(url string) bool {
	return strings.HasPrefix(url, "https://")
}
