package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Workiva/frugal/compiler"
	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
	"github.com/urfave/cli"
)

const defaultTopicDelim = "."

var (
	help    bool
	gen     string
	out     string
	delim   string
	audit   string
	recurse bool
	verbose bool
	version bool
	include string
)

func main() {
	app := cli.NewApp()
	app.Name = "frugal"
	app.Usage = "a tool for code generation"
	app.Version = globals.Version
	app.HideVersion = true
	app.HideHelp = true

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "help, h",
			Usage:       "show help",
			Destination: &help,
		},
		cli.StringFlag{
			Name:        "gen",
			Usage:       genUsage(),
			Destination: &gen,
		},
		cli.StringFlag{
			Name:        "out",
			Usage:       "set the output location for generated files (no gen-* folder will be created)",
			Destination: &out,
		},
		cli.StringFlag{
			Name:        "delim",
			Value:       defaultTopicDelim,
			Usage:       "set the delimiter for pub/sub topic tokens",
			Destination: &delim,
		},
		cli.BoolFlag{
			Name:        "recurse, r",
			Usage:       "generate included files",
			Destination: &recurse,
		},
		cli.BoolFlag{
			Name:        "verbose, v",
			Usage:       "verbose mode",
			Destination: &verbose,
		},
		cli.BoolFlag{
			Name:        "version",
			Usage:       "print the version",
			Destination: &version,
		}, cli.StringFlag{
			Name:        "audit",
			Usage:       "frugal file to run audit against",
			Destination: &audit,
		}, cli.StringFlag{
			Name:		"include, i",
			Usage:		"additional include directory",
			Destination: &include,
		},
	}

	app.Action = func(c *cli.Context) error {
		if help {
			cli.ShowAppHelp(c)
			os.Exit(0)
		}

		if version {
			cli.ShowVersion(c)
			os.Exit(0)
		}

		if len(c.Args()) == 0 {
			fmt.Printf("Usage: %s [options] file\n\n", app.Name)
			fmt.Printf("Use %s -help for a list of options\n", app.Name)
			os.Exit(1)
		}

		if gen == "" && audit == "" {
			fmt.Println("No output language specified")
			fmt.Printf("Usage: %s [options] file\n\n", app.Name)
			fmt.Printf("Use %s -help for a list of options\n", app.Name)
			os.Exit(1)
		}

		options := compiler.Options{
			Gen:     gen,
			Out:     out,
			Delim:   delim,
			Recurse: recurse,
			Verbose: verbose,
			IncludeDirs: strings.Split(include, ","),
		}

		// Handle panics for graceful error messages.
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Failed to generate %s:\n\t%s\n", options.File, r)
				os.Exit(1)
			}
		}()

		var err error
		auditor := parser.NewAuditor()
		for _, options.File = range c.Args() {
			if audit == "" {
				err = compiler.Compile(options)
			} else {
				err = auditor.Audit(audit, options.File, []string{include})
			}
			if err != nil {
				fmt.Printf("Failed to generate %s:\n\t%s\n", options.File, err.Error())
				os.Exit(1)
			}
		}

		return nil
	}

	app.Run(os.Args)
}

func genUsage() string {
	usage := "generate code with a registered generator and optional parameters " +
		"(lang[:key1=val1[,key2[,key3=val3]]])\n"
	langKeys := make([]string, 0, len(generator.Languages))
	for lang := range generator.Languages {
		langKeys = append(langKeys, lang)
	}
	sort.Strings(langKeys)
	langPrefix := ""
	for _, lang := range langKeys {
		options := generator.Languages[lang]
		optionsStr := ""
		optionKeys := make([]string, 0, len(options))
		for name := range options {
			optionKeys = append(optionKeys, name)
		}
		sort.Strings(optionKeys)
		for _, name := range optionKeys {
			optionsStr += fmt.Sprintf("\n\t        %s:\t%s", name, options[name])
		}
		usage += fmt.Sprintf("%s\t    %s:%s", langPrefix, lang, optionsStr)
		langPrefix = "\n"
	}
	return usage
}
