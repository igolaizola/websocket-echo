package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"

	"github.com/igolaizola/wsecho"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
)

// Build flags
var Version = ""
var Commit = ""
var Date = ""

func main() {
	// Create signal based context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Launch command
	cmd := newCommand()
	if err := cmd.ParseAndRun(ctx, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func newCommand() *ffcli.Command {
	fs := flag.NewFlagSet("wsecho", flag.ExitOnError)

	return &ffcli.Command{
		ShortUsage: "wsecho [flags] <subcommand>",
		FlagSet:    fs,
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			newVersionCommand(),
			newServeCommand(),
			newPingCommand(),
		},
	}
}

func newVersionCommand() *ffcli.Command {
	return &ffcli.Command{
		Name:       "version",
		ShortUsage: "wsecho version",
		ShortHelp:  "print version",
		Exec: func(ctx context.Context, args []string) error {
			v := Version
			if v == "" {
				if buildInfo, ok := debug.ReadBuildInfo(); ok {
					v = buildInfo.Main.Version
				}
			}
			if v == "" {
				v = "dev"
			}
			versionFields := []string{v}
			if Commit != "" {
				versionFields = append(versionFields, Commit)
			}
			if Date != "" {
				versionFields = append(versionFields, Date)
			}
			fmt.Println(strings.Join(versionFields, " "))
			return nil
		},
	}
}

func newServeCommand() *ffcli.Command {
	cmd := "serve"
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	_ = fs.String("config", "", "config file (optional)")

	addr := fs.String("addr", ":1337", "address to listen on")

	return &ffcli.Command{
		Name:       cmd,
		ShortUsage: fmt.Sprintf("wsecho %s [flags] <key> <value data...>", cmd),
		Options: []ff.Option{
			ff.WithConfigFileFlag("config"),
			ff.WithConfigFileParser(ff.PlainParser),
			ff.WithEnvVarPrefix("WSECHO"),
		},
		ShortHelp: fmt.Sprintf("wsecho %s command", cmd),
		FlagSet:   fs,
		Exec: func(ctx context.Context, args []string) error {
			if *addr == "" {
				return errors.New("missing address")
			}
			return wsecho.Serve(ctx, *addr)
		},
	}
}

func newPingCommand() *ffcli.Command {
	cmd := "ping"
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	_ = fs.String("config", "", "config file (optional)")

	host := fs.String("host", "ws://localhost:1337", "address to ping, e.g. ws://localhost:1337")
	n := fs.Int("n", 10, "number of pings to send")
	size := fs.Int("size", 32, "size of each ping message")

	return &ffcli.Command{
		Name:       cmd,
		ShortUsage: fmt.Sprintf("wsecho %s [flags] <key> <value data...>", cmd),
		Options: []ff.Option{
			ff.WithConfigFileFlag("config"),
			ff.WithConfigFileParser(ff.PlainParser),
			ff.WithEnvVarPrefix("WSECHO"),
		},
		ShortHelp: fmt.Sprintf("wsecho %s command", cmd),
		FlagSet:   fs,
		Exec: func(ctx context.Context, args []string) error {
			if *host == "" {
				return errors.New("missing host")
			}
			if *n < 1 {
				return errors.New("n must be greater than 0")
			}
			if *size < 1 {
				return errors.New("size must be greater than 0")
			}
			return wsecho.Ping(ctx, *host, *n, *size)
		},
	}
}
