package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func serve() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	mux := http.NewServeMux()
	mux.Handle("/", HandlerWithError(httpHandler))

	h := loggingMiddleware(coreHeaders(mux))

	addr := net.JoinHostPort(viper.GetString("host"), viper.GetString("http_port"))
	slog.Info("Listening at " + addr)

	if err := http.ListenAndServe(addr, h); err != nil {
		panic(err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"duration", time.Since(start),
		)
	})
}

// coreHeaders adds several headers
func coreHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "hello/"+version)
		w.Header().Set("Server", "hello/"+version)
		if val := os.Getenv("POD_NAME"); val != "" {
			w.Header().Set("X-Served-By", val)
		}
		envs := []string{
			"POD_NAME", "POD_IP",
			"NODE_NAME", "NODE_IP",
			"CONTAINER_PORT", "SVC", "SVC_PORT", "CLUSTER_IP", "EXTERNAL_IP",
		}
		for _, env := range envs {
			if val := os.Getenv(env); val != "" {
				w.Header().Set(toHeader(env), val)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func toHeader(s string) string {
	words := strings.Split(s, "_")
	for i := range words {
		words[i] = strings.ToTitle(words[i])
	}
	words = append([]string{"X"}, words...)
	return strings.Join(words, "-")
}

type HandlerWithError func(w http.ResponseWriter, r *http.Request) error

func (h HandlerWithError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err != nil {
		// Centralized error handling
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Error("httpHandler error", "err", err, "method", r.Method, "path", r.URL.Path)
	}
}

func httpHandler(w http.ResponseWriter, r *http.Request) error {
	var err error

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("OK")); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return nil
}

//nolint:unused
func getPreferredOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP 8.8.8.8:80: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			panic(fmt.Errorf("failed to close UDP connection: %w", err))
		}
	}()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
