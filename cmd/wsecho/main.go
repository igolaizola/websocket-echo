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

	port := fs.Int("port", 0, "port number")

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
			if *port == 0 {
				return errors.New("missing port")
			}
			return wsecho.Serve(ctx, *port)
		},
	}
}
