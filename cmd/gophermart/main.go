package main

import (
	"context"
	"flag"
	"fmt"

	"aprokhorov-diploma-1/cmd/gophermart/config"
	"aprokhorov-diploma-1/internal/storage"
)

func main() {
	//Init Config
	config := config.NewServerConfig()

	// Init flags
	flag.StringVar(&config.Server, "a", "127.0.0.1:8080", "Server ip:port")
	flag.StringVar(&config.Database, "d", "postgres://localhost:5432", "Database ip:port")
	flag.StringVar(&config.AccrualService, "r", "127.0.0.1:8082", "AccrualService ip:port")
	flag.StringVar(&config.DBName, "dn", "", "Database Name")
	flag.IntVar(&config.LogLevel, "l", 5, "Log Level, default:Warning")
	flag.Parse()

	fmt.Println(config)

	ctx := context.Background()
	if ctx != nil {
		ctx = context.Background()
	}

	postgres, err := storage.NewPostgresClient(ctx, config.Database, config.DBName)
	if err != nil {
		fmt.Println(err)
	}

	defer postgres.GracefulShutdown()
}
