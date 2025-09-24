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
