package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	natsproxy "github.com/regnull/nats-proxy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	defaultHumanReadableLog = false
	defaultLogLevel         = "info"
	defaultLogEnableColor   = true
)

type Args struct {
	LogEnableColor   bool
	LogLevel         string
	HumanReadableLog bool
	NatsUrl          string
}

func main() {
	var args Args

	flag.BoolVar(&args.LogEnableColor, "log-enable-color", defaultLogEnableColor, "enable color in logging")
	flag.StringVar(&args.LogLevel, "log-level", defaultLogLevel, "log level")
	flag.BoolVar(&args.HumanReadableLog, "human-readable-log", defaultHumanReadableLog, "human readable log")
	flag.StringVar(&args.NatsUrl, "nats-url", nats.DefaultURL, "NATS url")
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

	proxy, err := natsproxy.NewNatsClient(conn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create NATS proxy")
	}

	proxy.GET("/hello", func(c *natsproxy.Context) {
		user := struct {
			Message string
		}{
			"Hello there!",
		}
		c.JSON(200, user)
	})

	defer conn.Close()

	// Waiting for signal to close the client
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
