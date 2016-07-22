package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/Workiva/frugal/compiler"
	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
	"github.com/urfave/cli"
)

const defaultTopicDelim = "."

var (
	help               bool
	gen                string
	out                string
	delim              string
	sha                string
	token              string
	slug               string
	retainIntermediate bool
	recurse            bool
	verbose            bool
	version            bool
	compare            bool
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
			Name:        "retain-intermediate",
			Usage:       "retain generated intermediate thrift files",
			Destination: &retainIntermediate,
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
		}, cli.BoolFlag{
			Name:        "compare",
			Usage:       "do CI compare",
			Destination: &compare,
		}, cli.StringFlag{
			Name:        "sha",
			Usage:       "Sha to compare with. Only be used with compare flag",
			Destination: &sha,
		}, cli.StringFlag{
			Name:        "token",
			Usage:       "Token for fetching compare file from git. Only used with compare flag",
			Destination: &token,
		}, cli.StringFlag{
			Name:        "slug",
			Usage:       "slug for fetching compare file from git. Only used with compare flag",
			Destination: &slug,
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

		if gen == "" && !compare {
			fmt.Println("No output language specified")
			fmt.Printf("Usage: %s [options] file\n\n", app.Name)
			fmt.Printf("Use %s -help for a list of options\n", app.Name)
			os.Exit(1)
		}

		file := c.Args()[0]
		options := compiler.Options{
			File:               file,
			Gen:                gen,
			Out:                out,
			Delim:              delim,
			RetainIntermediate: retainIntermediate,
			Recurse:            recurse,
			Verbose:            verbose,
		}

		// Handle panics for graceful error messages.
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Failed to generate %s:\n\t%s\n", options.File, r)
				os.Exit(1)
			}
		}()

		if !compare {
			if err := compiler.Compile(options); err != nil {
				fmt.Printf("Failed to generate %s:\n\t%s\n", options.File, err.Error())
				os.Exit(1)
			}
		} else {
			// check sha, slug and token
			if sha == "" || token == "" || slug == "" {
				panic("specify compare fields with!\n$ frugal -compare -sha {commit} -token {token} -slug{git slug} {file strings}")
			}
			if err := parser.Compare(sha, slug, token, options.File); err != nil {
				fmt.Printf("Failed to do comparison %s:\n\t%s\n", options.File, err.Error())
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
