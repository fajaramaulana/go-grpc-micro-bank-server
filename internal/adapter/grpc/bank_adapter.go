package grpc

import (
	"context"
	"time"

	"github.com/fajaramaulana/go-grpc-micro-bank-proto/protogen/go/bank"
	"google.golang.org/genproto/googleapis/type/date"
)

func (a *GrpcAdapter) GetCurrentBalance(ctx context.Context, req *bank.CurrentBalanceRequest) (*bank.CurrentBalanceResponse, error) {
	now := time.Now()

	balance, err := a.bankService.GetCurrentBalance(req.GetAccountNumber())
	if err != nil {
		return nil, err
	}

	return &bank.CurrentBalanceResponse{
		Amount: balance,
		CurrentDate: &date.Date{
			Year:  int32(now.Year()),
			Month: int32(now.Month()),
			Day:   int32(now.Day()),
		},
	}, nil
}
