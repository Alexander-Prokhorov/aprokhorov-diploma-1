package main

import (
	"context"
	"flag"
	"fmt"

	"aprokhorov-diploma-1/cmd/gophermart/config"
	"aprokhorov-diploma-1/internal/logger"
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
	flag.StringVar(&config.LogLevel, "l", "debug", "Log Level, default:debug")
	flag.Parse()

	//Init Logger
	var log logger.Logger
	var err error

	log, err = logger.NewZeroLogger(config.LogLevel)
	if err != nil {
		panic(err)
	}
	log.Info("main", "Start GopherMart Today!")
	log.Info("main", fmt.Sprint(config))

	ctx := context.Background()

	postgres, err := storage.NewPostgresClient(ctx, config.Database, config.DBName)
	if err != nil {
		log.Error("main", err.Error())
	}

	defer postgres.GracefulShutdown()
}
