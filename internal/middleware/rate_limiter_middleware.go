package middleware

import (
	"gin/user-management-api/internal/utils"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type Client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*Client)
)

func getClientIP(ctx *gin.Context) string {
	ip := ctx.ClientIP()
	if ip == "" {
		ip = ctx.Request.RemoteAddr
	}

	return ip
}

func getRateLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	client, exists := clients[ip]
	if !exists {
		requestSecStr := utils.GetIntEnv("RATE_LIMITER_REQUEST_SEC", 5)
		brustStr := utils.GetIntEnv("RATE_LIMITER_REQUEST_BURST", 10)

		limiter := rate.NewLimiter(rate.Limit(requestSecStr), brustStr) // 5 request/sec, brust 10
		newClient := &Client{limiter, time.Now()}
		clients[ip] = newClient
		return limiter
	}

	client.lastSeen = time.Now()
	return client.limiter
}

func CleanupClients() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, client := range clients {
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}
}

// ab -n 20 -c 1 -H "X-API-Key:2a2cc361-9801-4036-8200-3088e14a403e" http://localhost:8080/api/v1/users
func RateLimiterMiddleware(rateLimiterLogger *zerolog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := getClientIP(ctx)

		limiter := getRateLimiter(ip)

		if !limiter.Allow() {
			if shouldLogRateLimit(ip) {
				rateLimiterLogger.Warn().
				Str("method", ctx.Request.Method).
				Str("path", ctx.Request.URL.Path).
				Str("query", ctx.Request.URL.RawQuery).
				Str("client_ip", ctx.ClientIP()).
				Str("user_agent", ctx.Request.UserAgent()).
				Str("referer", ctx.Request.Referer()).
				Str("protocol", ctx.Request.Host).
				Str("host", ctx.Request.Method).
				Str("remote_addr", ctx.Request.RemoteAddr).
				Interface("headers", ctx.Request.Header).
				Msg("rate limiter execested")
			}

			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many request",
				"message": "Bạn đã gửi quá nhiêu request. Hãy thử lại sau",
			})
			return
		}

		ctx.Next()
	}
}

var rateLimitLogCache = sync.Map{}

const rateLimitLogTTL = 10 * time.Second

func shouldLogRateLimit(ip string) bool {
	now := time.Now()
	if val, ok := rateLimitLogCache.Load(ip); ok {
		if t, ok := val.(time.Time); ok && now.Sub(t) < rateLimitLogTTL {
			return false
		}
	}

	rateLimitLogCache.Store(ip,now)
	return true
}
