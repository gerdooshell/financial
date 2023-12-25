package connection_config

import (
	"errors"
	"fmt"
	"github.com/gerdooshell/financial/environment"
	"github.com/gerdooshell/financial/lib/database/database_engine"
	"github.com/gerdooshell/financial/lib/database/database_hosttag"
	"github.com/go-yaml/yaml"
	"os"
)

type connectionConfig struct {
	Engine   database_engine.Engine   `yaml:"Engine"`
	HostTag  database_hosttag.HostTag `yaml:"HostTag"`
	Host     string                   `yaml:"Host"`
	User     string                   `yaml:"User"`
	Password string                   `yaml:"Password"`
	Port     int                      `yaml:"Port"`
	Database string                   `yaml:"Database"`
	SSL      bool                     `yaml:"SSL"`
}

type ConnectionConfig interface {
	GetConnectionString() string
	GetSignature() struct {
		Engine  database_engine.Engine
		HostTag database_hosttag.HostTag
	}
	GetEngine() database_engine.Engine
	GetHostTag() database_hosttag.HostTag
}

func FromConfigFile(absFilePath string, env environment.Environment) (ConnectionConfig, error) {
	data, err := os.ReadFile(absFilePath)
	if err != nil {
		return nil, err
	}
	var confMap map[environment.Environment]connectionConfig
	if err = yaml.Unmarshal(data, &confMap); err != nil {
		return nil, err
	}
	conf, ok := confMap[env]
	if !ok {
		return nil, errors.New(fmt.Sprintf("no config found for environment %v", env))
	}
	return &conf, nil
}

func (c *connectionConfig) GetConnectionString() string {
	if c.Engine == database_engine.Postgres {
		return c.toPostgresConnectionString()
	}
	return ""
}

func (c *connectionConfig) toPostgresConnectionString() string {
	sslMode := "disable"
	if c.SSL {
		sslMode = "enable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, sslMode)
}

func (c *connectionConfig) GetSignature() struct {
	Engine  database_engine.Engine
	HostTag database_hosttag.HostTag
} {
	return struct {
		Engine  database_engine.Engine
		HostTag database_hosttag.HostTag
	}{Engine: c.Engine, HostTag: c.HostTag}
}

func (c *connectionConfig) GetEngine() database_engine.Engine {
	return c.Engine
}

func (c *connectionConfig) GetHostTag() database_hosttag.HostTag {
	return c.HostTag
}
