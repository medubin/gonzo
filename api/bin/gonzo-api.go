package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	api "github.com/medubin/gonzo/api/generate"
	"github.com/medubin/gonzo/api/generate/fileio"
)

type Runner interface {
	Init([]string) error
	Run() error
	Name() string
}

func main() {
	if err := root(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func root(args []string) error {
	if len(args) < 1 {
		return errors.New("you must pass a sub-command")
	}

	cmds := []Runner{
		NewGenerateCommand(),
	}

	subcommand := os.Args[1]

	for _, cmd := range cmds {
		if cmd.Name() == subcommand {
			cmd.Init(os.Args[2:])
			return cmd.Run()
		}
	}

	return fmt.Errorf("unknown subcommand: %s", subcommand)
}

func NewGenerateCommand() *GenerateCommand {
	gc := &GenerateCommand{
		fs: flag.NewFlagSet("generate", flag.ContinueOnError),
	}
	gc.fs.StringVar(&gc.input, "input", "", "input file. Should end in .api")
	gc.fs.StringVar(&gc.output, "output", "", "output directory")

	return gc
}

type GenerateCommand struct {
	fs *flag.FlagSet

	input  string
	output string
}

func (g *GenerateCommand) Name() string {
	return g.fs.Name()
}

func (g *GenerateCommand) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *GenerateCommand) Run() error {
	lines, err := fileio.ParseFile(g.input + ".api")
	if err != nil {
		return err
	}

	types, endpoints, err := api.Generate(lines)
	if err != nil {
		return err
	}

	err = fileio.WriteToFile(g.output, "types", types)
	if err != nil {
		return err
	}

	err = fileio.WriteEndpoints(g.output, endpoints)
	if err != nil {
		return err
	}

	server := `package server

type ServerImpl struct{}
`
	err = fileio.SafeWriteToFile(g.output, "server", server)
	if err != nil {
		return err
	}

	return nil
}
