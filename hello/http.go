package main

import (
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/spf13/viper"
)

func serve() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	mux := http.NewServeMux()
	mux.HandleFunc("/", httpHandler)

	h := loggingMiddleware(mux)

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

func httpHandler(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	var res []byte
	var err error

	h, _ := os.Hostname()
	res = append(res, []byte("Host: "+h+"\n")...)

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		res = append(res, []byte("IP:\n")...)
		slog.Debug("failed to retrieve local IP address",
			"error", err,
			"method", r.Method,
			"remote", r.RemoteAddr,
		)
	} else {
		defer func() {
			if err := conn.Close(); err != nil {
				slog.Debug("failed to close UDP socket",
					"error", err,
					"method", r.Method,
					"remote", r.RemoteAddr,
				)
			}
		}()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		res = append(res, []byte("IP: "+localAddr.IP.String()+"\n")...)
	}

	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		slog.Error("failed to dump request",
			"error", err,
			"method", r.Method,
			"remote", r.RemoteAddr,
		)
		http.Error(w, "failed to dump request: ", http.StatusInternalServerError)
		return
	}
	res = append(res, dump...)
	res = append(res, []byte(Footer(t0))...)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)
	if err != nil {
		slog.Error("failed to write response",
			"error", err,
			"method", r.Method,
			"remote", r.RemoteAddr,
		)
	}
}
