package cron

import (
	"context"
	"fmt"
	"time"

	"aprokhorov-diploma-1/cmd/gophermart/accrual"
	"aprokhorov-diploma-1/internal/logger"
	"aprokhorov-diploma-1/internal/storage"
)

func StartOrderCheckProcess(ctx context.Context, signal <-chan time.Time, accrual *accrual.AccrualService, database storage.Storage, log logger.Logger) {
	parent := "Accrual:CheckTask"

	for {
		select {
		case <-signal:
			//log.Debug(parent, "Update Orders from Accrual Service")
			orders, err := database.GetOrdersUndone(ctx)
			if err != nil {
				log.Error(parent, err.Error())
			}

			for _, order := range orders {
				orderAccrual, err := accrual.FetchData(order.OrderID)
				if err != nil {
					log.Info(parent, err.Error())
				}

				if orderAccrual.OrderID != "" {
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
			}
		case <-ctx.Done():
			return
		}
	}
}
