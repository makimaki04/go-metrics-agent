package middleware

import (
	"net"
	"net/http"
	"strings"
)

func TrustedSubnet(ipNet *net.IPNet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ipNet == nil {
				next.ServeHTTP(w, r)
				return
			}

			ipStr := strings.TrimSpace(r.Header.Get("X-Real-IP"))
			ip := net.ParseIP(ipStr)
			if ip == nil {
				http.Error(w, "forbidden: X-Real-IP not in trusted subnet", http.StatusForbidden)
				return
			}

			ok := ipNet.Contains(ip)
			if !ok {
				http.Error(w, "forbidden: X-Real-IP not in trusted subnet", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
