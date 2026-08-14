package delivery

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*client
	r       rate.Limit
	b       int
}

// NewIPRateLimiter membuat rate limiter (r = rate/detik, b = max burst capacity)
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		clients: make(map[string]*client),
		r:       r,
		b:       b,
	}

	// Clean up background goroutine untuk menghapus IP yang tidak aktif > 3 menit
	go i.cleanupClients()

	return i
}

func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	c, exists := i.clients[ip]
	if !exists {
		limiter := rate.NewLimiter(i.r, i.b)
		i.clients[ip] = &client{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	c.lastSeen = time.Now()
	return c.limiter
}

func (i *IPRateLimiter) cleanupClients() {
	for {
		time.Sleep(1 * time.Minute)

		i.mu.Lock()
		for ip, c := range i.clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(i.clients, ip)
			}
		}
		i.mu.Unlock()
	}
}

// RateLimitMiddleware mengembalikan Gin HandlerFunc (Batas: r req/detik, burst b)
func RateLimitMiddleware(requestsPerSecond float64, burst int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate.Limit(requestsPerSecond), burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		clientLimiter := limiter.getLimiter(ip)

		if !clientLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak request (429 Too Many Requests). Silakan coba beberapa detik lagi.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
