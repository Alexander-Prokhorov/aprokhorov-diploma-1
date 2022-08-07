package accrual

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccrualService_FetchData(t *testing.T) {
	tests := []struct {
		name    string
		orderNo string
		want    Order
		wantErr bool
	}{
		{
			name:    "Live AccrualService Test",
			orderNo: "371449635398431",
			want:    Order{OrderID: "371449635398431", Status: "PROCESSED", Accrual: 0},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			a := NewAccrualService("localhost:8081", time.Second)
			order, err := a.FetchData(tt.orderNo)

			if !tt.wantErr {
				assert.NoError(t, err)
			}

			assert.EqualValues(t, tt.want, order)
		})
	}
}
