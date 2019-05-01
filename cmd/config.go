package cmd

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Config Store values read from config.yaml file
type Config struct {
	PackageName string   `yaml:"packageName"`
	Generate    Generate `yaml:"generate"`
}

// Generate Stores generate values from yaml
type Generate struct {
	// ProtectGnorm When true, prevents gnorm's default files from being
	// overwritten
	ProtectGnorm bool               `yaml:"protectGnorm"`
	SchemaName   string             `yaml:"schemaName"`
	Resolvers    []ResolverGenerate `yaml:"resolvers"`
	Postgres     []PostgresGenerate `yaml:"postgres"`
}

// ResolverGenerate Which resolver related things to generate code for
type ResolverGenerate struct {
	SingularModelName string `yaml:"singularName"`
	PluralModelName   string `yaml:"pluralName"`
	PrimaryKey        string `yaml:"primaryKey"`
	PrimaryKeyType    string `yaml:"primaryKeyType"`
	Create            bool   `yaml:"create"`        // Build a create function
	Update            bool   `yaml:"update"`        // Build an update function
	PrepareCreate     bool   `yaml:"prepareCreate"` // Provide a prepare function for you (set to false if you want to set one yourself)
	Query             bool   `yaml:"query"`         // Creates a queryX function used for pagination via a connections type method
}

// PostgresGenerate Which postgres helper functions to generate code for
type PostgresGenerate struct {
	ModelName      string `yaml:"modelName"`      // Name of model used by GraphQL
	ModelStruct    string `yaml:"modelStruct"`    // Struct to use for GraphQL model
	ModelPackage   string `yaml:"modelPackage"`   // Path of the package containing the model
	PmName         string `yaml:"postgresName"`   // Name of postgres data object
	PK             string `yaml:"primaryKey"`     // Go struct for database name for primary key field
	PrimaryKeyType string `yaml:"primaryKeyType"` // Go type for primary key
	Create         bool   `yaml:"create"`         // Generate create/update related functions
}

func readConfig(filename string) (Config, error) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(input, &config)
	if err != nil {
		return config, err
	}

	return config, err
}

// UnmarshalYAML Allows us to set default values when not provided in the
// config
func (r *ResolverGenerate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawR ResolverGenerate

	raw := rawR{
		PrimaryKey:     "ID",
		PrimaryKeyType: "string",
	}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	*r = ResolverGenerate(raw)
	return nil
}
