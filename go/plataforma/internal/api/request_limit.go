package api

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type window struct {
	start time.Time
	count int
}
type jobLimiter struct {
	mu      sync.Mutex
	clients map[string]window
	now     func() time.Time
}

func configuredJobLimit() int {
	raw, ok := os.LookupEnv("JOB_POSTS_PER_MINUTE")
	if !ok {
		return 10
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 10
	}
	return n // 0 desactiva el límite durante una campaña local aislada.
}

func limitJobCreation(next http.Handler, maximum int) http.Handler {
	if maximum == 0 {
		return next
	}
	l := &jobLimiter{clients: make(map[string]window), now: time.Now}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/jobs" {
			next.ServeHTTP(w, r)
			return
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = strings.TrimSpace(r.RemoteAddr)
		}
		now := l.now()
		l.mu.Lock()
		for ip, v := range l.clients {
			if now.Sub(v.start) >= time.Minute {
				delete(l.clients, ip)
			}
		}
		v := l.clients[host]
		if now.Sub(v.start) >= time.Minute {
			v = window{start: now}
		}
		if v.count >= maximum {
			l.clients[host] = v
			l.mu.Unlock()
			w.Header().Set("Retry-After", "60")
			responderError(w, http.StatusTooManyRequests, "limite_solicitudes", "Se alcanzó el límite temporal de ejecuciones. Intenta más tarde.")
			return
		}
		v.count++
		l.clients[host] = v
		l.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
