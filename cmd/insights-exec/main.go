package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	insights "github.com/subpop/insightsd/pkg"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "base-url", Required: true},
		&cli.StringFlag{Name: "auth", Required: true},
		&cli.StringFlag{Name: "ca-root", TakesFile: true},
		&cli.StringFlag{Name: "cert", TakesFile: true},
		&cli.StringFlag{Name: "key", TakesFile: true},
		&cli.StringFlag{Name: "username"},
		&cli.StringFlag{Name: "password"},
	}

	app.Commands = []*cli.Command{
		{
			Name: "upload",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:      "file",
					TakesFile: true,
					Required:  true,
				},
				&cli.StringFlag{
					Name:     "collector",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				var client *insights.Client
				var err error

				switch strings.ToLower(c.String("auth")) {
				case "basic":
					username := c.String("username")
					if username == "" {
						reader := bufio.NewReader(os.Stdin)
						fmt.Print("Username: ")
						text, err := reader.ReadString('\n')
						if err != nil {
							return cli.NewExitError(err, 1)
						}
						username = strings.TrimSpace(text)
					}

					password := c.String("password")
					if password == "" {
						fmt.Print("Password: ")
						data, err := terminal.ReadPassword(int(os.Stdin.Fd()))
						if err != nil {
							return cli.NewExitError(err, 1)
						}
						password = strings.TrimSpace(string(data))
					}
					client, err = insights.NewClientBasicAuth(c.String("base-url"), username, password)
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
					return cli.NewExitError(fmt.Errorf("error: '%v' is not a valid option for '--auth'", c.String("auth")), 1)
				}

				if err := insights.Upload(client, c.String("file"), c.String("collector")); err != nil {
					return cli.NewExitError(err, 1)
				}

				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
