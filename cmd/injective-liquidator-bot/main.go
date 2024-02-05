package main

import (
	"fmt"
	"os"

	"github.com/InjectiveLabs/injective-liquidator-bot/internal/pkg/version"

	cli "github.com/jawher/mow.cli"
	log "github.com/xlab/suplog"

	"github.com/joho/godotenv"
)

var app = cli.App("injective-liquidator-bot", "Injective's Liquidator Bot.")

var (
	envFileName    *string
	envName        *string
	appLogLevel    *string
	svcWaitTimeout *string
)

func main() {
	envFileName = app.String(cli.StringOpt{
		Name:  "e env",
		Desc:  "File name/path with the environment configuration (.env)",
		Value: ".env",
	})

	err := godotenv.Load(*envFileName)
	if err != nil {
		log.Fatalf("Error loading %s file", *envFileName)
	}

	initGlobalOptions(
		&envName,
		&appLogLevel,
		&svcWaitTimeout,
	)

	app.Before = func() {
		log.DefaultLogger.SetLevel(logLevel(*appLogLevel))
	}

	app.Command("start", "Starts the liquidator main loop.", liquidatorCmd)
	app.Command("version", "Print the version information and exit.", versionCmd)

	_ = app.Run(os.Args)
}

func versionCmd(c *cli.Cmd) {
	c.Action = func() {
		fmt.Println(version.Version())
	}
}
