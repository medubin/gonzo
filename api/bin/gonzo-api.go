package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/medubin/gonzo/api/code_generator/fileio"
	"github.com/medubin/gonzo/api/code_generator/generator"
	"github.com/medubin/gonzo/api/code_generator/utils"
)

type Languages int

var serverRegex = regexp.MustCompile(`^server\..+$`)
var typesRegex = regexp.MustCompile(`^types\..+$`)

const (
	Golang Languages = iota
)

type Stacks int

const (
	Server Stacks = iota
)

var StackLanguages = map[Stacks]map[Languages]string{
	Server: {
		Golang: "",
	},
}

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
	gc.fs.StringVar(&gc.stack, "stack", "", "server or client. Defaults to server")
	gc.fs.StringVar(&gc.language, "language", "", "language, can be go or typescript")
	gc.fs.StringVar(&gc.packageName, "package", "", "package name")

	return gc
}

type GenerateCommand struct {
	fs *flag.FlagSet

	input  string
	output string

	stack       string
	language    string
	packageName string
}

func (g *GenerateCommand) Name() string {
	return g.fs.Name()
}

func (g *GenerateCommand) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *GenerateCommand) Run() error {
	if !utils.IsLanguageStackAllowed(g.language, g.stack) {
		return fmt.Errorf("unsupported language stack combination: %s, %s", g.language, g.stack)
	}

	if g.input == "" {
		return fmt.Errorf("input required")
	}

	if g.output == "" {
		return fmt.Errorf("output required")
	}

	if g.packageName == "" {
		return fmt.Errorf("package name required")

	}

	lines, err := fileio.ParseFile(g.input + ".api")
	if err != nil {
		return err
	}

	parser := generator.NewParser(string(lines))
	api, err := parser.Parse()
	if err != nil {
		return err
	}

	template, err := generator.NewTemplateGenerator("api/code_generator/generator/languages/go/server/config.yaml")
	if err != nil {
		return err
	}

	results, err := template.Generate(api, g.packageName)
	if err != nil {
		return err
	}

	for name, result := range results {
		safe := (!serverRegex.MatchString(name) && !typesRegex.MatchString(name))
		err = fileio.WriteToFile(g.output, name, result, safe)
		if err != nil {
			return err
		}
	}

	// err = fileio.WriteToFile(g.output, "types", types,)
	// if err != nil {
	// 	return err
	// }

	// err = fileio.WriteEndpoints(g.output, endpoints)
	// if err != nil {
	// 	return err
	// }

	// 	server := `package server

	// type ServerImpl struct{}
	// `
	// err = fileio.SafeWriteToFile(g.output, "server_impl", server)
	// if err != nil {
	// 	return err
	// }

	return nil
}
