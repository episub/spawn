package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var gqlConfigDefault = `
schema:
- schema.graphql
exec:
  filename: graph/generated.go
  package: graph
model:
  filename: models/models_gen.go
  package: models
resolver:
  filename: resolvers/resolver.go
  package: resolvers
  type: Resolver
`

var gqlSchemaDefault = `
# GraphQL schema example
#
# https://gqlgen.com/getting-started/

type Todo {
  id: ID!
  content: String!
  done: Boolean!
  user: User!
}

type User {
  id: ID!
  username: String!
  admin: Boolean!
}

type Query {
  todos: [Todo!]!
}

input NewTodo {
  text: String!
  userId: String!
}

type Mutation {
  createTodo(input: NewTodo!): Todo!
}

enum SortDirection {
  ASC
  DESC
}
`

var dockerComposeDefault = `
version: '3'
services:
  postgres:
    image: postgres:9.6
    ports:
    - "5432:5432"
    restart: always
    environment:
      POSTGRES_USER: spawn
      POSTGRES_PASSWORD: spawn
    volumes:
      - ./migrations:/docker-entrypoint-initdb.d
`

var initCmd = cli.Command{
	Name:  "init",
	Usage: "create a new spawn project",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "package",
			Value: "github.com/example/replace",
			Usage: "specify the name of the package for this project",
		},
		cli.StringFlag{Name: "config, c", Usage: "the config filename"},
		cli.StringFlag{Name: "schema", Usage: "where to write the schema stub to", Value: "schema.graphql"},
		cli.StringFlag{Name: "server", Usage: "where to write the server stub to", Value: "server/server.go"},
		cli.StringFlag{Name: "folder", Usage: "where to create the project"},
	},
	Action: func(ctx *cli.Context) {
		if len(ctx.String("folder")) > 0 {
			err := os.Chdir(ctx.String("folder"))
			if err != nil {
				exit(err)
			}
		}

		// Check if config already exists, if it does, assume init has been run
		// already:
		_, err := os.Stat("config.yaml")
		if err == nil {
			log.Printf("It appears init has already been run.  Please remove config.yaml to continue if you think this is mistaken.")
			return
		}

		log.Printf("init path: %s", ctx.String("folder"))

		// Ensure package name is loaded:
		err = loadPackageName()
		if err != nil {
			exit(err)
		}

		_ = os.Mkdir("migrations", 0755)
		_ = os.Mkdir("static", 0755)
		_ = os.Mkdir("templates", 0755)
		_ = os.Mkdir("loader", 0755)

		createFile(ctx, "schema.graphql", gqlSchemaDefault)
		createFile(ctx, "gqlgen.yml", gqlConfigDefault)
		createFile(ctx, "docker-compose.yml", dockerComposeDefault)
		copyTemplate("migrations/001-base.sql", "migrations/001-base.sql")
		createFileFromTemplate("gnorm.toml", "gnorm.toml")
		createFileFromTemplate("config.yaml", "config.yaml")

		// OPA policy related files
		_ = os.MkdirAll("policies/bundle/api/entity", 0755)
		_ = os.MkdirAll("policies/bundle/api/mutation", 0755)
		_ = os.MkdirAll("policies/bundle/api/query", 0755)

		generateGQL(ctx)
		createFileFromTemplate("server.go", "server.go")
		createFileFromTemplate("loader/init.gotmpl", "loader/init.go")
	},
}

// folder Returns the location to reference or create the file
func filePath(ctx *cli.Context, filename string) string {
	fd := ctx.String("folder")
	if len(fd) == 0 {
		log.Printf("filePath: %s", filename)
		return filename
	}
	final := ctx.String("folder") + "/" + filename
	final = strings.Replace(final, "//", "/", -1)

	log.Printf("filePath: %s", final)
	return final
}

func createFile(ctx *cli.Context, filename string, contents string) {
	filename = filePath(ctx, filename)
	_, err := os.Stat(filename)
	if !os.IsNotExist(err) {
		return
	}

	err = ioutil.WriteFile(filename, []byte(strings.TrimSpace(contents)), 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("unable to write file %s: %s", filename, err))
		os.Exit(1)
	}
}

// createFileFromTemplate input is the filename (under cmd/static) to use as template, and output  isi the file name to create
func createFileFromTemplate(input string, output string) {
	t := loadTemplateFromFile(input)

	f, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	err = t.Execute(f, packageName)

	if err != nil {
		panic(err)
	}
}

// GenerateGQL Generates gql stuff
func generateGQL(ctx *cli.Context) *config.Config {
	var cfg *config.Config
	var err error
	if configFilename := ctx.String("config"); configFilename != "" {
		cfg, err = config.LoadConfig(configFilename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	} else {
		cfg, err = config.LoadConfigFromDefaultLocations()
		if os.IsNotExist(errors.Cause(err)) {
			cfg = config.DefaultConfig()
		} else if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(2)
		}
	}

	if err = api.Generate(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(3)
	}

	return cfg
}
