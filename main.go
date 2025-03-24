package main

import (
	ratelimiter "github.com/veyelutd/go-rate-limiter/rate-limiter"
	"net/http"
)

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		bucket := ratelimiter.GetTokenBucketForIP(ip)
		if !bucket.IsRequestAllowed() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func main() {
	go ratelimiter.CleanUpExpiredBuckets()
	router := http.NewServeMux()
	router.HandleFunc("/ratelimit", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	})
	err := http.ListenAndServe(":8080", rateLimitMiddleware(router))
	if err != nil {
		panic(err)
	}
}
