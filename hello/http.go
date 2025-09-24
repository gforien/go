package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Options for [fetch], using functional options pattern.
type fetchOpts struct {
	resolve func(host string) string
}

func withResolve() func(*fetchOpts) {
	return func(opts *fetchOpts) {
		resolve := func(host string) string {
			if ip := net.ParseIP(host); ip != nil {
				return host
			}

			ips, err := (&net.Resolver{PreferGo: true}).LookupHost(context.Background(), host)
			if err != nil {
				panic(fmt.Errorf("DNS lookup failed: %w", err))
			}
			if len(ips) == 0 {
				panic(fmt.Errorf("no IPs found for %s", host))
			}

			fmt.Printf("* Host %s resolved to %s\n", host, ips[0])
			fmt.Printf("* IPs: %s\n", strings.Join(ips, ", "))
			fmt.Printf("*\n")

			ip := ips[0]
			return ip
		}
		opts.resolve = resolve
	}
}

func fetch(u *url.URL, opts ...(func(*fetchOpts))) (*http.Response, []byte) {
	client := http.DefaultClient

	// apply options
	options := &fetchOpts{}
	for _, o := range opts {
		o(options)
	}

	if options.resolve != nil {
		if ip := options.resolve(u.Hostname()); ip != u.Hostname() {
			transport := &http.Transport{
				DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
					d := net.Dialer{Timeout: 5 * time.Second}
					return d.DialContext(ctx, network, net.JoinHostPort(ip, u.Port()))
				},
			}
			client = &http.Client{Transport: transport}
		}
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	cobra.CheckErr(err)
	dumpReq, err := httputil.DumpRequestOut(req, true)
	cobra.CheckErr(err)
	for _, line := range strings.Split(string(dumpReq), "\n") {
		os.Stdout.WriteString("> " + line + "\n")
	}

	res, err := client.Do(req)
	cobra.CheckErr(err)
	body, err := io.ReadAll(res.Body)
	cobra.CheckErr(err)
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Debug("failed to close response body",
				"error", err,
			)
		}
	}()

	// read the entire response body and print it to stdout
	dumpRes, err := httputil.DumpResponse(res, false)
	cobra.CheckErr(err)
	for _, line := range strings.Split(string(dumpRes), "\n") {
		os.Stdout.WriteString("< " + line + "\n")
	}

	os.Stdout.WriteString(string(body) + "\n")
	return res, body
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
