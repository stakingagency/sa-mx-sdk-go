package main

import (
	"errors"
	"os"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/stakingagency/sa-mx-sdk-go/abi2go/converter"
	"github.com/urfave/cli"
)

var (
	abi2goHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`
	// configPathFlag defines a flag for the path of the application's configuration file
	abiFileFlag = cli.StringFlag{
		Name:  "file",
		Usage: "the ABI file to convert",
		Value: "",
	}
)

var log = logger.GetOrCreate("abi2go")

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = abi2goHelpTemplate
	app.Name = "abi2go"
	app.Usage = "GO binding tool for MultiversX smart contracts ABI files"
	app.Flags = []cli.Flag{
		abiFileFlag,
	}
	app.Version = "v0.0.1"
	app.Authors = []cli.Author{
		{
			Name:  "Staking Agency",
			Email: "contact@staking.agency",
		},
	}

	app.Action = func(c *cli.Context) error {
		return convert(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}

func convert(ctx *cli.Context) error {
	abiFileName := ctx.GlobalString(abiFileFlag.Name)
	if abiFileName == "" {
		errorText := "ABI file not specified"
		log.Error(errorText)
		return errors.New(errorText)
	}

	err := logger.SetLogLevel("*:" + logger.LogDebug.String())
	if err != nil {
		log.Error("failed to set the log level", "error", err)
		return err
	}

	args := file.ArgsFileLogging{
		WorkingDir:      ".",
		DefaultLogsPath: "logs",
		LogFilePrefix:   "abi2go",
	}
	_, err = file.NewFileLogging(args)
	if err != nil {
		log.Error("failed to create the log file", "error", err)
		return err
	}

	log.Info("opening ABI file...")

	conv, err := converter.NewAbiConverter(abiFileName)
	if err != nil {
		log.Error("failed to open ABI file", "error", err)
		return err
	}

	log.Info("converting ABI file...")

	err = conv.Convert()
	if err != nil {
		log.Error("failed to convert ABI file", "error", err)
		return err
	}

	log.Info("ABI converted successfully")

	return nil
}
