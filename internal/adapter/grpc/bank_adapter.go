package grpc

import (
	"context"
	"time"

	"github.com/fajaramaulana/go-grpc-micro-bank-proto/protogen/go/bank"
	"google.golang.org/genproto/googleapis/type/date"
)

// GetCurrentBalance retrieves the current balance for a given account number.
// It takes a context.Context and a *bank.CurrentBalanceRequest as input parameters.
// It returns a *bank.CurrentBalanceResponse and an error as output.
// The function uses the bankService to get the current balance for the specified account number.
// If an error occurs during the retrieval, it returns nil and the error.
// Otherwise, it constructs a *bank.CurrentBalanceResponse with the retrieved balance and the current date.
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
