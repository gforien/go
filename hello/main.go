package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "none"
)

var Version = func() string {
	return fmt.Sprintf("%s (commit %s, built on %s)", version, commit, date)
}

var Date = func() string {
	return time.Now().Format(time.UnixDate)
}

var Footer = func(t0 time.Time) string {
	var b strings.Builder
	b.WriteString("at " + Date() + "\n")
	b.WriteString("by hello " + Version() + "\n")
	b.WriteString("in " + FormatDuration(time.Since(t0)) + "\n")
	return b.String()
}

func main() {
	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	cliFlag      bool
	hostFlag     string
	httpFlag     bool
	httpPortFlag string
	listenFlag   bool // when true, act as a server
	resolveFlag  bool
	stdinFlag    bool
)

var root = &cobra.Command{
	Use: `hello [flags]                          (client mode)
  hello -l [flags]                       (server mode)`,
	Short: "A simple hello world application. Available via CLI and HTTP.",
	Args: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("http") && !viper.GetBool("listen") && viper.GetString("host") == "" {
			return fmt.Errorf("expects a host to connect to (e.g. 'hello --http --host example.com')")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("listen") {
			if viper.GetBool("http") {
				serve()
				return nil
			}
		}

		if viper.GetBool("http") {
			rawUrl := viper.GetString("host")
			if !strings.Contains(rawUrl, "://") {
				rawUrl = "http://" + rawUrl
			}

			u, err := url.Parse(rawUrl)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to parse URL: %v\n", err)
			}

			if u.Scheme == "" {
				u.Scheme = "http"
			}
			if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Hostname(), viper.GetString("http_port"))
			}
			fmt.Printf("Fetching %s\n", u.String())

			opts := []func(*fetchOpts){}
			if viper.GetBool("resolve") {
				opts = append(opts, withResolve())
			}

			for {
				_, _ = fetch(u, opts...)
				time.Sleep(10 * time.Second)
			}
		}

		if viper.GetBool("stdin") || (len(args) == 1 && args[0] == "-") {
			return stdinHandler()
		}

		res := cliHandler()
		_, err := os.Stdout.WriteString(res)
		return err
	},
	Version: Version(),
}

// FormatDuration returns a string like "153 µs", "12ms", "1.23 s"
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return strconv.FormatInt(d.Nanoseconds(), 10) + " ns"
	} else if d < time.Millisecond {
		return strconv.FormatInt(d.Microseconds(), 10) + " µs"
	} else if d < time.Second {
		// show in milliseconds with 1 decimal
		return strconv.FormatInt(d.Milliseconds(), 10) + " ms"
	} else {
		// show in seconds with 2 decimals
		s := float64(d) / float64(time.Second)
		return fmt.Sprintf("%.2f s", s)
	}
}

func init() {
	root.Flags().BoolVar(&cliFlag, "cli", false, "hello CLI")
	root.Flags().StringVar(&hostFlag, "host", "", "hostname")
	root.Flags().BoolVar(&httpFlag, "http", false, "use HTTP")
	root.Flags().StringVar(&httpPortFlag, "http-port", "80", "HTTP port")
	root.Flags().BoolVarP(&listenFlag, "listen", "l", false, "act as a server")
	root.Flags().BoolVar(&resolveFlag, "resolve", false, "execute a new DNS resolution each time")
	root.Flags().BoolVar(&stdinFlag, "stdin", false, "hello stdin")

	viper.SetEnvPrefix("hello")
	cobra.CheckErr(viper.BindPFlag("cli", root.Flags().Lookup("cli")))
	cobra.CheckErr(viper.BindPFlag("host", root.Flags().Lookup("host")))
	cobra.CheckErr(viper.BindPFlag("http", root.Flags().Lookup("http")))
	cobra.CheckErr(viper.BindPFlag("http_port", root.Flags().Lookup("http-port")))
	cobra.CheckErr(viper.BindPFlag("listen", root.Flags().Lookup("listen")))
	cobra.CheckErr(viper.BindPFlag("resolve", root.Flags().Lookup("resolve")))
	cobra.CheckErr(viper.BindPFlag("stdin", root.Flags().Lookup("stdin")))
	viper.AutomaticEnv()
}
