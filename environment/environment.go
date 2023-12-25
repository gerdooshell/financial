package environment

import "fmt"

type Environment string

const (
	Prod Environment = "Prod"
	Dev  Environment = "Dev"
)

var environment Environment

func SetEnvironment(env Environment) error {
	if environment != "" {
		return fmt.Errorf("environment is already set to %v", environment)
	}
	if env != Prod && env != Dev {
		return fmt.Errorf("invalid environment %v. choose between %v and %v", env, Dev, Prod)
	}
	environment = env
	return nil
}

func GetEnvironment() Environment {
	return environment
}
