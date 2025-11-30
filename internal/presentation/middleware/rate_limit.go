package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
)

type RateLimitMiddleware struct {
	cache database.Cache
}

func NewRateLimitMiddleware(cache database.Cache) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		cache: cache,
	}
}

func (rm *RateLimitMiddleware) RateLimit(maxRequests int, windowSize time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()
		key := "ratelimit:" + clientIP

		rdb := rm.cache.GetRDB()
		count, err := rdb.Incr(context.Background(), key).Result()
		if err != nil {
			panic(err)
		}

		if count == 1 {
			rdb.Expire(context.Background(), key, windowSize)
		}

		if count > int64(maxRequests) {
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "too many requests, please try again later",
			})
			ctx.Abort()
			return
		}

		ttl, _ := rdb.TTL(context.Background(), key).Result()
		ctx.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
		ctx.Header("X-RateLimit-Remaining", strconv.FormatInt(int64(maxRequests)-count, 10))
		ctx.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

		ctx.Next()
	}
}

func (rm *RateLimitMiddleware) LoginRateLimit() gin.HandlerFunc {
	return rm.RateLimit(5, 15*time.Minute)
}

func (rm *RateLimitMiddleware) SignupRateLimit() gin.HandlerFunc {
	return rm.RateLimit(3, 1*time.Hour)
}
