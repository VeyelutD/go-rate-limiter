package ratelimiter

import (
	"sync"
	"time"
)

var ipBuckets = make(map[string]*TokenBucket)
var ipBucketsMutex sync.Mutex

// interval for the cleanup routine
var cleanupInterval = time.Minute * 5

// the size of the burst
var burst int64 = 10

// the maximum amount of tokens a bucket can hold
var maxTokens int64 = 5

// tokens will refill at the rate of 1 token for every 30 seconds
var tokenRefillRate = time.Second * 30

type TokenBucket struct {
	tokens     int64
	maxTokens  int64
	rate       time.Duration
	lastRefill time.Time
	lastUsed   time.Time
	sync.Mutex
}

func newTokenBucket(burst, maxTokens int64, rate time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:     burst,
		maxTokens:  maxTokens,
		rate:       rate,
		lastRefill: time.Now(),
	}
}

func CleanUpExpiredBuckets() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		ipBucketsMutex.Lock()
		now := time.Now()
		for ip, bucket := range ipBuckets {
			expiration := 1 * time.Minute
			if bucket.tokens < bucket.maxTokens/2 {
				expiration = 2 * time.Minute
			}
			if now.Sub(bucket.lastUsed) > expiration {
				delete(ipBuckets, ip)
			}
		}
		ipBucketsMutex.Unlock()
	}

}

func (tb *TokenBucket) IsRequestAllowed() bool {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int64(elapsed / tb.rate)
	if tokensToAdd > 0 {
		if tokensToAdd+tb.tokens > tb.maxTokens {
			tb.tokens = tb.maxTokens
		} else {
			tb.tokens = tb.tokens + tokensToAdd
		}
		tb.lastRefill = now
	}
	if tb.tokens > 0 {
		tb.tokens--
		tb.lastUsed = now
		return true
	}
	return false
}

func GetTokenBucketForIP(ip string) *TokenBucket {
	ipBucketsMutex.Lock()
	defer ipBucketsMutex.Unlock()
	bucket, ok := ipBuckets[ip]
	if ok {
		return bucket
	}
	bucket = newTokenBucket(burst, maxTokens, tokenRefillRate)
	ipBuckets[ip] = bucket
	return bucket
}
