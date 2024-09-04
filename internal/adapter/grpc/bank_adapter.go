package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/fajaramaulana/go-grpc-micro-bank-proto/protogen/go/bank"
	domainBank "github.com/fajaramaulana/go-grpc-micro-bank-server/internal/application/domain/bank"
	"github.com/fajaramaulana/go-grpc-micro-bank-server/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/datetime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		log.Error().Msg(fmt.Sprintf("account %v not found", req.AccountNumber))
		return nil, status.Errorf(
			codes.FailedPrecondition,
			"account %v not found", req.AccountNumber,
		)
	}
	// get exchange rate
	exchangeRate, err := a.bankService.FindExchangeRate("USD", "IDR", now)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Exchange rate from %v to %v on %v not found", "USD", "IDR", now.Format(time.RFC3339)))
		return nil, status.Errorf(
			codes.FailedPrecondition,
			"Exchange rate from %v to %v on %v not found", "USD", "IDR", now.Format(time.RFC3339),
		)
	}

	// convert balance to IDR
	balanceExchange := balance * exchangeRate

	return &bank.CurrentBalanceResponse{
		Amount:        balance,
		AmountConvert: balanceExchange,
		CurrentDate: &date.Date{
			Year:  int32(now.Year()),
			Month: int32(now.Month()),
			Day:   int32(now.Day()),
		},
	}, nil
}

func (a *GrpcAdapter) FetchExchangeRates(req *bank.ExchangeRateRequest, stream grpc.ServerStreamingServer[bank.ExchangeRateResponse]) error {
	context := stream.Context()

	for {
		select {
		case <-context.Done():
			log.Info().Msg("Client Cancelled stream")
			return nil
		default:
			now := time.Now().Truncate(time.Second)
			rate, err := a.bankService.FindExchangeRate(req.FromCurrency, req.ToCurrency, now)

			if err != nil {
				s := status.New(codes.InvalidArgument,
					"Currency not valid. Please use valid currency for both from and to")
				s, _ = s.WithDetails(&errdetails.ErrorInfo{
					Domain: "my-bank-website.com",
					Reason: "INVALID_CURRENCY",
					Metadata: map[string]string{
						"from_currency": req.FromCurrency,
						"to_currency":   req.ToCurrency,
					},
				})

				return s.Err()
			}

			stream.Send(
				&bank.ExchangeRateResponse{
					FromCurrency: req.FromCurrency,
					ToCurrency:   req.ToCurrency,
					Rate:         rate,
					Timestamp:    now.Format(time.RFC3339),
				},
			)
			log.Info().Msg(fmt.Sprintf("Exchange rate sent to client, %v to %v : %v\n", req.FromCurrency,
				req.ToCurrency, rate))

			time.Sleep(3 * time.Second)
		}
	}
}

func (a *GrpcAdapter) SummarizeTransactions(stream grpc.ClientStreamingServer[bank.Transaction, bank.TransactionSummary]) error {
	trxSum := domainBank.TransactionSummary{
		SummaryDate: time.Now(),
		SumIn:       0,
		SumOut:      0,
		SumTotal:    0,
	}

	account := ""

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			res := bank.TransactionSummary{
				AccountNumber: account,
				SumAmountIn:   trxSum.SumIn,
				SumAmountOut:  trxSum.SumOut,
				SumAmount:     trxSum.SumTotal,
				Timestamp: &datetime.DateTime{
					Year:    int32(trxSum.SummaryDate.Year()),
					Month:   int32(trxSum.SummaryDate.Month()),
					Day:     int32(trxSum.SummaryDate.Day()),
					Hours:   int32(time.Now().Hour()),
					Minutes: int32(time.Now().Minute()),
					Seconds: int32(time.Now().Second()),
					Nanos:   int32(time.Now().Nanosecond()),
				},
			}

			return stream.SendAndClose(&res)
		}

		if err != nil {
			logErr := util.LogError("Error while reading from client : "+err.Error(), "", "Bank Adapter GRPC - SummarizeTransactions - stream.Recv()")
			log.Fatal().Msg(logErr)
		}

		if req.Amount < 0 {
			errMsg := fmt.Sprintf("Requested amount %v is negative", req.Amount)
			logErr := util.LogError(errMsg, "", "Bank Adapter GRPC - SummarizeTransactions - check negative req.Amount")
			log.Error().Msg(logErr)
			s := status.New(codes.InvalidArgument, errMsg)
			s, _ = s.WithDetails(&errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "amount",
						Description: errMsg,
					},
				},
			})

			return s.Err()
		}

		account = req.AccountNumber
		ts, err := util.ToTime(req.Timestamp)

		if err != nil {
			logErr := util.LogError(fmt.Sprintf("Error while parsing timestamp %v : %v", req.Timestamp, err), "", "Bank Adapter GRPC - SummarizeTransactions - util.ToTime")
			log.Fatal().Msg(logErr)
		}

		trxType := domainBank.TransactionTypeUnknown

		if req.Type == bank.TransactionType_TRANSACTION_TYPE_IN {
			trxType = domainBank.TransactionTypeIn
		} else if req.Type == bank.TransactionType_TRANSACTION_TYPE_OUT {
			trxType = domainBank.TransactionTypeOut
		}
		trxCurrent := domainBank.Transaction{
			Amount:          req.Amount,
			Timestamp:       ts,
			TransactionType: trxType,
		}

		accountUuid, err := a.bankService.CreateTransaction(req.AccountNumber, trxCurrent)

		if err != nil && accountUuid == uuid.Nil {
			logErr := util.LogError(fmt.Sprintf("Invalid account number: %v", err), "", "Bank Adapter GRPC - SummarizeTransactions - a.bankService.CreateTransaction")
			log.Error().Msg(logErr)
			s := status.New(codes.InvalidArgument, err.Error())
			s, _ = s.WithDetails(&errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "account_number",
						Description: "Invalid account number",
					},
				},
			})

			return s.Err()
		} else if err != nil && accountUuid != uuid.Nil {
			errMsg := fmt.Sprintf("Requested amount %v exceed available balance", req.Amount)
			logErr := util.LogError(errMsg, "", "Bank Adapter GRPC - SummarizeTransactions - a.bankService.CreateTransaction")
			log.Error().Msg(logErr)
			s := status.New(codes.InvalidArgument, err.Error())
			s, _ = s.WithDetails(&errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "amount",
						Description: errMsg,
					},
				},
			})

			return s.Err()
		}

		if err != nil {
			logErr := util.LogError("Error while creating transaction : "+err.Error(), "", "Bank Adapter GRPC - SummarizeTransactions - a.bankService.CreateTransaction")
			log.Error().Msg(logErr)
		}

		err = a.bankService.CalculateTransactionSummary(&trxSum, trxCurrent)

		if err != nil {
			return err
		}
	}
}

func (a *GrpcAdapter) TransferMultiple(stream grpc.BidiStreamingServer[bank.TransferRequest, bank.TransferResponse]) error {
	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Client cancelled stream")
			return nil // Context was cancelled, exit gracefully
		default:
			req, err := stream.Recv()
			if err == io.EOF {
				log.Info().Msg("Client has closed sending, finishing stream.")
				return nil // Normal stream closure
			}

			if err != nil {
				log.Error().Msg(fmt.Sprintf("Error while reading from client: %s", err.Error()))
				return err // Return the error to the client
			}

			transferTrx := domainBank.TransferTransaction{
				FromAccountNumber: req.AccountNumberSender,
				ToAccountNumber:   req.AccountNumberReciever,
				Currency:          req.GetCurrency(),
				Amount:            req.GetAmount(),
				Notes:             req.Notes,
			}

			_, transferSuccess, err := a.bankService.Transfer(transferTrx)
			if err != nil {
				return buildTransferErrorStatusGrpc(err, req) // Handle and send detailed gRPC error status
			}

			res := &bank.TransferResponse{
				AccountNumberSender:   req.AccountNumberSender,
				AccountNumberReciever: req.AccountNumberReciever,
				Currency:              req.Currency,
				Amount:                req.Amount,
				Timestamp:             util.CurrentDatetime(),
			}

			if transferSuccess {
				res.Status = bank.TransferStatus_TRANSFER_STATUS_SUCCESS
			} else {
				res.Status = bank.TransferStatus_TRANSFER_STATUS_FAILED
			}

			// Send response to the client
			if err := stream.Send(res); err != nil {
				log.Error().Msg(fmt.Sprintf("Error while sending response to client: %s", err.Error()))
				return err // Return the send error to the client
			}
		}
	}
}

func buildTransferErrorStatusGrpc(err error, req *bank.TransferRequest) error {
	switch {
	case errors.Is(err, domainBank.ErrTransferSourceAccountNotFound):
		s := status.New(codes.FailedPrecondition, err.Error())
		s, _ = s.WithDetails(&errdetails.PreconditionFailure{
			Violations: []*errdetails.PreconditionFailure_Violation{
				{
					Type:        "INVALID_ACCOUNT",
					Subject:     "Source account not found",
					Description: fmt.Sprintf("source account (from %v) not found", req.AccountNumberSender),
				},
			},
		})

		return s.Err()
	case errors.Is(err, domainBank.ErrTransferDestinationAccountNotFound):
		s := status.New(codes.FailedPrecondition, err.Error())
		s, _ = s.WithDetails(&errdetails.PreconditionFailure{
			Violations: []*errdetails.PreconditionFailure_Violation{
				{
					Type:        "INVALID_ACCOUNT",
					Subject:     "Destination account not found",
					Description: fmt.Sprintf("destination account (to %v) not found", req.AccountNumberReciever),
				},
			},
		})

		return s.Err()
	case errors.Is(err, domainBank.ErrTransferRecordFailed):
		s := status.New(codes.Internal, err.Error())
		s, _ = s.WithDetails(&errdetails.Help{
			Links: []*errdetails.Help_Link{
				{
					Url:         "my-bank-website.com/faq",
					Description: "Bank FAQ",
				},
			},
		})

		return s.Err()
	case errors.Is(err, domainBank.ErrTransferTransactionPair):
		s := status.New(codes.InvalidArgument, err.Error())
		s, _ = s.WithDetails(&errdetails.ErrorInfo{
			Domain: "my-bank-website.com",
			Reason: "TRANSACTION_PAIR_FAILED",
			Metadata: map[string]string{
				"from_account": req.AccountNumberSender,
				"to_account":   req.AccountNumberReciever,
				"currency":     req.Currency,
				"amount":       fmt.Sprintf("%f", req.Amount),
			},
		})

		return s.Err()
	default:
		s := status.New(codes.Unknown, err.Error())
		return s.Err()
	}
}
