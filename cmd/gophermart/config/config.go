package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

type Config struct {
	Server         string `env:"RUN_ADDRESS"`
	Database       string `env:"DATABASE_URI"`
	AccrualService string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DBName         string `env:"DATABASE_NAME"`
	LogLevel       int    `env:"GOPHERMART_LOGLEVEL"`
}

func (c *Config) EnvInit() error {
	err := env.Parse(c)
	if err != nil {
		return err
	}
	return nil
}

func (c Config) String() string {
	return fmt.Sprintf(
		"Server: %s, Database: %s, Database Name: %s, AccrualService: %s, LogLevel:%v",
		c.Server,
		c.Database,
		c.DBName,
		c.AccrualService,
		c.LogLevel,
	)
}

func NewServerConfig() *Config {
	return &Config{
		Server:         "",
		Database:       "",
		AccrualService: "",
		LogLevel:       5,
	}
}
