package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aprokhorov-diploma-1/cmd/gophermart/accrual"
	"aprokhorov-diploma-1/cmd/gophermart/config"
	"aprokhorov-diploma-1/cmd/gophermart/handlers"
	"aprokhorov-diploma-1/internal/cache"
	"aprokhorov-diploma-1/internal/hasher"
	"aprokhorov-diploma-1/internal/logger"
	"aprokhorov-diploma-1/internal/storage"
	"aprokhorov-diploma-1/internal/verificator"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	//Init Config
	config := config.NewServerConfig()

	// Init flags
	flag.StringVar(&config.Server, "a", "127.0.0.1:8080", "Server ip:port")
	flag.StringVar(&config.Database, "d", "", "Database ip:port")
	flag.StringVar(&config.AccrualService, "r", "http://127.0.0.1:8081", "AccrualService ip:port")
	flag.StringVar(&config.AccrualFrequency, "rf", "50us", "AccrualService Frequency, default:1s")
	flag.StringVar(&config.DBName, "dn", "", "Database Name")
	flag.StringVar(&config.LogLevel, "l", "debug", "Log Level, default:debug")
	flag.StringVar(&config.AuthCacheTimeout, "at", "300s", "Auth Cache Timeout, default:300s")
	flag.StringVar(&config.AuthCacheHouseKeeperTime, "ah", "1h", "Auth Cache HouseKeeper Interval, default:1h")
	flag.Parse()

	//Init Logger
	var log logger.Logger
	var err error

	log, err = logger.NewZeroLogger(config.LogLevel)
	if err != nil {
		panic(err)
	}

	// Init Config from Env
	err = config.EnvInit()
	if err != nil {
		log.Error("main", "Can't get enviroment")
	}

	log.Info("main", "Start GopherMart Today!")
	log.Info("main", fmt.Sprint(config))

	// Init system calls
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Init context
	ctx := context.Background()

	// Init Database
	database, err := storage.NewPostgresClient(ctx, config.Database, config.DBName)
	if err != nil {
		log.Fatal("main", err.Error())
	}
	defer database.GracefulShutdown()

	// Init Hasher
	mainHasher := hasher.NewHMAC()

	// Init AuthCache
	authCacheTimeout, err := time.ParseDuration(config.AuthCacheTimeout)
	if err != nil {
		log.Fatal("main", err.Error())
	}
	authCache := cache.NewMemCache(authCacheTimeout, log)

	authCacheHousekeeper, err := time.ParseDuration(config.AuthCacheHouseKeeperTime)
	if err != nil {
		log.Fatal("main", err.Error())
	}
	authHousekeeperTicker := time.NewTicker(authCacheHousekeeper)

	go func() {
		for {
			<-authHousekeeperTicker.C
			err := authCache.HouseKeeper()
			if err != nil {
				log.Error("MemCache:HouseKeeper", err.Error())
			}
			// TODO: add closing channel
		}
	}()

	// Init Verificator
	verificator, err := verificator.NewLuhn()
	if err != nil {
		log.Fatal("main", err.Error())
	}

	/*
		POST /api/user/register — регистрация пользователя;
		POST /api/user/login — аутентификация пользователя;
		POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
		GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
		GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
		POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
		GET /api/user/balance/withdrawals -- ошибка в ТЗ, правильный /api/user/withdrawals
	*/
	r := chi.NewRouter()
	r.Use(middleware.Logger)      // Access Log
	r.Use(middleware.Compress(5)) // Support for gzip

	r.Route("/api/user", func(r chi.Router) {
		r.Route("/", func(r chi.Router) {
			r.Use(handlers.CheckHeaders(log)) // Check content-type == app/json for post.request
			r.Post("/register", handlers.Authorize(true, database, authCache, mainHasher, log))
			r.Post("/login", handlers.Authorize(false, database, authCache, mainHasher, log))
		})

		r.Route("/orders", func(r chi.Router) {
			r.Use(handlers.AuthMiddleware(authCache, log)) // Check Authorization Token
			r.Post("/", handlers.NewOrder(database, verificator, log))
			r.Get("/", handlers.GetOrders(database, log))
		})

		r.Route("/balance", func(r chi.Router) {
			r.Use(handlers.CheckHeaders(log))              // Check content-type == app/json for post.request
			r.Use(handlers.AuthMiddleware(authCache, log)) // Check Authorization Token
			r.Get("/", handlers.GetBalance(database, log))
			r.Post("/withdraw", handlers.AddWithdraw(database, log))
		})
		r.Get("/withdrawals", handlers.GetWithdrawals(database, log))
	})

	// Init Server
	server := &http.Server{
		Addr:    config.Server,
		Handler: r,
	}

	go func() {
		log.Fatal("main", server.ListenAndServe().Error())
	}()

	log.Info("main", "Server Started")

	// Accrual Service Operations
	frequency, err := time.ParseDuration(config.AccrualFrequency)
	if err != nil {
		log.Error("main", err.Error())
	}
	accrual := accrual.NewAccrualService(config.AccrualService, frequency)

	ticketAccrual := time.NewTicker(frequency)

	go func(ctx context.Context, database storage.Storage) {
		parent := "Accrual:CheckTask"

		for {
			select {
			case <-ticketAccrual.C:
				//log.Debug(parent, "Update Orders from Accrual Service")
				orders, err := database.GetOrdersUndone(ctx)
				if err != nil {
					log.Error(parent, err.Error())
				}

				for _, order := range orders {
					orderAccrual, err := accrual.FetchData(order.OrderId)
					if err != nil {
						log.Info(parent, err.Error())
					}

					log.Debug(parent, fmt.Sprint(orderAccrual))
					err = database.ModifyOrder(ctx, orderAccrual.OrderID, orderAccrual.Status, orderAccrual.Accrual)
					if err != nil {
						log.Info(parent, err.Error())
					}

					balance, err := database.GetBalance(ctx, order.Login)
					if err != nil {
						log.Info(parent, err.Error())
					}

					newBalance := balance.CurrentScore + orderAccrual.Accrual
					err = database.ModifyBalance(ctx, order.Login, newBalance, balance.TotalWithdrawals)
					if err != nil {
						log.Info(parent, err.Error())
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}(ctx, database)

	<-done
	log.Info("main", "Shutdown")

}
