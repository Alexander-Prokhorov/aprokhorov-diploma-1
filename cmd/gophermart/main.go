package main

import (
	"flag"
	"fmt"

	"aprokhorov-diploma-1/cmd/gophermart/config"
)

func main() {
	//Init Config
	config := config.NewServerConfig()

	// Init flags
	flag.StringVar(&config.Server, "a", "127.0.0.1:8080", "Server ip:port")
	flag.StringVar(&config.Database, "d", "127.0.0.1:8081", "Database ip:port")
	flag.StringVar(&config.AccrualService, "r", "127.0.0.1:8082", "AccrualService ip:port")
	flag.IntVar(&config.LogLevel, "l", 5, "Log Level, default:Warning")
	flag.Parse()

	fmt.Println(config)
}
