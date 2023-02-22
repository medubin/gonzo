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
		return errors.New("You must pass a sub-command")
	}

	cmds := []Runner{
		NewGreetGenerateCommand(),
	}

	subcommand := os.Args[1]

	for _, cmd := range cmds {
		if cmd.Name() == subcommand {
			cmd.Init(os.Args[2:])
			return cmd.Run()
		}
	}

	return fmt.Errorf("Unknown subcommand: %s", subcommand)
}

func NewGreetGenerateCommand() *GenerateCommand {
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

	data, err := api.GenerateData(lines)
	if err != nil {
		return err
	}

	output, err := api.GenerateTypes(data)
	if err != nil {
		return err
	}

	err = fileio.WriteToFile(g.output, "types", output)
	if err != nil {
		return err
	}

	endpoints, err := api.GenerateEndpoints(data)
	if err != nil {
		return err
	}

	err = fileio.WriteEndpoints(g.output, endpoints)
	if err != nil {
		return err
	}

	server, err := api.GenerateServer()
	if err != nil {
		return err
	}

	err = fileio.SafeWriteToFile(g.output, "server", server)
	if err != nil {
		return err
	}

	return nil
}
