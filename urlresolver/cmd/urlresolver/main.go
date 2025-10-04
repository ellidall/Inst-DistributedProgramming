package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	flagFilePathSystem = "f"
	flagFilePath       = "-f"

	emptyString = ""
)

type Args struct {
	FilePath string
}

func main() {
	args, err := parseArgs()
	if err != nil {
		panic(err)
	}

	cfg, err := getConfig(args)
	if err != nil {
		panic(err)
	}

	err = runServer(cfg)
	if err != nil {
		panic(err)
	}
}

func parseArgs() (*Args, error) {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	flagSet.Usage = func() {
		fmt.Printf("Usage: %s %s <file_path>\n", flagFilePath, os.Args[0])
		fmt.Println("Flags:")
		flagSet.PrintDefaults()
	}

	var configPath string
	flagSet.StringVar(&configPath, flagFilePathSystem, emptyString, "Path to the JSON configuration file")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("invalid flag provided")
	}

	if flagSet.NArg() > 0 {
		flagSet.Usage()
		return nil, fmt.Errorf("unexpected positional arguments: %v", flagSet.Args())
	}

	if configPath == emptyString {
		flagSet.Usage()
		return nil, fmt.Errorf("required flag %s not provided", flagFilePath)
	}

	return &Args{
		FilePath: configPath,
	}, nil
}
