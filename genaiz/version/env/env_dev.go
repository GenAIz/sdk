//go:build dev

package env

import "strings"

func DefaultProtocolPrefix(host string) string {
	// Only localhost in development will ever be allowed to default to http
	if strings.HasPrefix(host, "localhost") {
		return "http:/"
	}

	return "https:/"
}

func GetVersionTag() string {
	return "-dev"
}

func IsAllowedProtocol(string) bool {
	return true
}
