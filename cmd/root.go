package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"runtime"
	"text/template"

	"github.com/urfave/cli"

	// Required since otherwise dep will prune away these unused packages before codegen has a chance to run
	_ "github.com/99designs/gqlgen/handler"
)

var packageName string
var packageRx = regexp.MustCompile(`^module (.+)$`)

// Execute Run spawn
func Execute() {
	loadTemplates()

	app := cli.NewApp()
	app.Name = "spawn"
	app.Usage = genCmd.Usage
	app.Description = "Tools and libraries for setting up and maintaining a project using the Episub Stack."
	app.HideVersion = true
	app.Flags = genCmd.Flags
	app.Action = genCmd.Action
	app.Commands = []cli.Command{
		genCmd,
		initCmd,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

}

func loadTemplates() {
	loaderTemplate = loadTemplateFromFile("loader/generated.gotmpl")
	filterTemplate = loadTemplateFromFile("models/filter.gotmpl")
	postgresTemplate = loadTemplateFromFile("loader/gen.gotmpl")
	postgresFileTemplate = loadTemplateFromFile("loader/gen__file_management.gotmpl")
	resolverTemplate = loadTemplateFromFile("resolvers/gen.gotmpl")
}

// loadTemplateFromFile Loads template from the package's local directory, under static folder
func loadTemplateFromFile(input string) *template.Template {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}

	source, err := ioutil.ReadFile(path.Dir(filename) + "/static/" + input)
	if err != nil {
		panic(err)
	}

	return template.Must(template.New("").Funcs(templateFuncs).Parse(string(source)))
}

// loadPackageName Grabs the package/module name for this project from go.mod
func loadPackageName() error {
	if len(packageName) > 0 {
		return nil
	}
	file, err := os.Open("go.mod")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	line := scanner.Text()
	if err := scanner.Err(); err != nil {
		return err
	}

	// Grab it from first line:
	matches := packageRx.FindStringSubmatch(line)
	if len(matches) != 2 {
		log.Printf("Line: %s", line)
		return fmt.Errorf("Could not find module name in first line of go.mod.")
	}

	packageName = matches[1]

	log.Printf("package name: %s", packageName)

	return nil
}
