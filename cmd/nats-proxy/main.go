package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/nats-io/nats.go"
	natsproxy "github.com/regnull/nats-proxy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	defaultHost             = ""
	defaultPort             = 8080
	defaultHumanReadableLog = false
	defaultLogLevel         = "info"
	defaultLogEnableColor   = true
	defaultAllowOrigin      = "*"
	defaultAllowMethods     = "POST, GET"
	defaultAllowHeaders     = "*"
)

type Args struct {
	LogEnableColor   bool
	LogLevel         string
	HumanReadableLog bool
	Host             string
	Port             int
	NatsUrl          string
	SubjectPrefix    string
	AllowOrigin      string
	AllowHeaders     string
	AllowMethods     string
}

func main() {
	var args Args

	flag.BoolVar(&args.LogEnableColor, "log-enable-color", defaultLogEnableColor, "enable color in logging")
	flag.StringVar(&args.LogLevel, "log-level", defaultLogLevel, "log level")
	flag.BoolVar(&args.HumanReadableLog, "human-readable-log", defaultHumanReadableLog, "human readable log")
	flag.StringVar(&args.Host, "host", defaultHost, "host")
	flag.IntVar(&args.Port, "port", defaultPort, "port")
	flag.StringVar(&args.NatsUrl, "nats-url", nats.DefaultURL, "NATS url")
	flag.StringVar(&args.SubjectPrefix, "subj-prefix", "", "NATS subject prefix")
	flag.StringVar(&args.AllowOrigin, "allow-origin", defaultAllowOrigin, "pre-flight allow origin")
	flag.StringVar(&args.AllowMethods, "allow-methods", defaultAllowMethods, "pre-flight allow methods")
	flag.StringVar(&args.AllowHeaders, "allow-headers", defaultAllowHeaders, "pre-flight allow headers")
	flag.Parse()

	if args.HumanReadableLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "01/02 15:04:05", NoColor: !args.LogEnableColor})
	}

	logLevel, err := zerolog.ParseLevel(args.LogLevel)
	if err != nil {
		log.Fatal().Str("level", args.LogLevel).Msg("invalid log level")
	}
	zerolog.SetGlobalLevel(logLevel)

	conn, err := nats.Connect(args.NatsUrl)
	if err != nil {
		log.Fatal().Str("nats-url", args.NatsUrl).Err(err).Msg("failed to connect to NATS")
	}
	log.Info().Str("nats-url", args.NatsUrl).Msg("connected to NATS")

	proxy, err := natsproxy.NewNatsProxy(conn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create NATS proxy")
	}
	proxy = proxy.
		WithPrefix(args.SubjectPrefix).
		WithAlowOrigin(args.AllowOrigin).
		WithAllowMethods(args.AllowMethods).
		WithAllowHeaders(args.AllowHeaders)

	defer conn.Close()
	http.ListenAndServe(fmt.Sprintf("%s:%d", args.Host, args.Port), proxy)
}
