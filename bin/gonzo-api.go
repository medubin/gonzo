package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/medubin/gonzo/code_generator/fileio"
	"github.com/medubin/gonzo/code_generator/generator"
	"github.com/medubin/gonzo/code_generator/utils"
)

type Languages int

// serverRegex matches server.go at any path depth (e.g., user_service/server.go).
var serverRegex = regexp.MustCompile(`(^|/)server\.[^/]+$`)
var typesRegex = regexp.MustCompile(`(^|/)types\.[^/]+$`)
var clientRegex = regexp.MustCompile(`(^|/)client\.[^/]+$`)

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

	subcommand := args[0]

	for _, cmd := range cmds {
		if cmd.Name() == subcommand {
			cmd.Init(args[1:])
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
	config := utils.GetLanguageStackConfig(g.language, g.stack)
	if config == "" {
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

	lines, err := fileio.ParseFile(g.input)
	if err != nil {
		return err
	}

	parser := generator.NewParser(string(lines), g.input)
	api, err := parser.Parse()
	if err != nil {
		return err
	}

	tmpl, err := generator.NewTemplateGenerator(config)
	if err != nil {
		return err
	}

	// For Go server generation, compute the full import path of the types package
	// so sub-packages can import it.
	var typesPackage string
	if g.language == "go" && g.stack == "server" {
		typesPackage, err = computeTypesPackage(g.output)
		if err != nil {
			// Non-fatal: generated code will compile but sub-packages won't have the import.
			fmt.Fprintf(os.Stderr, "warning: could not compute types package path: %v\n", err)
		}
	}

	results, err := tmpl.Generate(api, g.packageName, typesPackage)
	if err != nil {
		return err
	}

	for name, result := range results {
		safe := (!serverRegex.MatchString(name) && !typesRegex.MatchString(name) && !clientRegex.MatchString(name))
		err = fileio.WriteToFile(g.output, name, result, safe)
		if err != nil {
			return err
		}
	}

	return nil
}

// computeTypesPackage resolves the full Go module import path for the output directory.
// It walks up from the current directory to find go.mod and combines the module path
// with the relative path from the module root to outputDir.
func computeTypesPackage(outputDir string) (string, error) {
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return "", err
	}

	// Walk up from the output directory to find go.mod
	dir := absOutputDir
	var modulePath, moduleRoot string
	for {
		goModPath := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(goModPath)
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "module ") {
					modulePath = strings.TrimSpace(strings.TrimPrefix(line, "module "))
					moduleRoot = dir
					break
				}
			}
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}

	if modulePath == "" {
		return "", fmt.Errorf("module path not found in go.mod")
	}

	relPath, err := filepath.Rel(moduleRoot, absOutputDir)
	if err != nil {
		return "", err
	}

	// Normalize path separators for Go import paths
	relPath = strings.ReplaceAll(relPath, string(filepath.Separator), "/")

	return modulePath + "/" + relPath, nil
}
