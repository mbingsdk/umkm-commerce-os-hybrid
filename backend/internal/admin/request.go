package admin

import (
	"net"
	"net/http"
	"strings"
)

type AuditInput struct {
	ActorUserID string
	Action      string
	TargetType  string
	TargetID    string
	BeforeData  any
	AfterData   any
	IPAddress   string
	UserAgent   string
}

func RequestMetadata(r *http.Request) (ipAddress string, userAgent string) {
	userAgent = strings.TrimSpace(r.UserAgent())

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host, userAgent
	}

	return strings.TrimSpace(r.RemoteAddr), userAgent
}
