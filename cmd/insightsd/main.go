package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	internal "github.com/subpop/insightsd/internal"
	insights "github.com/subpop/insightsd/pkg"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func main() {
	flags := []cli.Flag{
		&cli.StringFlag{Name: "config"},
		&cli.StringFlag{Name: "log-level", Value: "error"},
		altsrc.NewStringFlag(&cli.StringFlag{Name: "base-url"}),
		altsrc.NewStringFlag(&cli.StringFlag{Name: "auth"}),
		altsrc.NewStringFlag(&cli.StringFlag{Name: "username"}),
		altsrc.NewStringFlag(&cli.StringFlag{Name: "password"}),
	}

	app := cli.NewApp()
	app.Flags = flags
	app.Action = func(c *cli.Context) error {
		level, err := log.ParseLevel(c.String("log-level"))
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		log.SetLevel(level)

		var client *insights.Client
		switch strings.ToLower(c.String("auth")) {
		case "basic":
			client, err = insights.NewClientBasicAuth(c.String("base-url"),
				c.String("username"),
				c.String("password"))
			if err != nil {
				return cli.NewExitError(err, 1)
			}
		case "cert":
			client, err = insights.NewClientCertAuth(c.String("base-url"),
				c.String("ca-root"),
				c.String("cert"),
				c.String("key"))
			if err != nil {
				return cli.NewExitError(err, 1)
			}
		default:
			return cli.NewExitError(fmt.Errorf("error: invalid value for 'auth': %v", c.String("auth")), 1)
		}

		server, err := internal.NewDBusServer(client)
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		if err := server.Connect(); err != nil {
			return cli.NewExitError(err, 1)
		}
		defer server.Close()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
		<-quit

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
