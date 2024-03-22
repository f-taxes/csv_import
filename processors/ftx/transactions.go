package ftx

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"strings"
	"time"

	"github.com/f-taxes/csv_import/global"
	g "github.com/f-taxes/csv_import/grpc_client"
	"github.com/f-taxes/csv_import/proto"
	"github.com/kataras/golog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type columnFunc func(val string, tx *proto.SrcTx) error

type TransactionProcessor struct {
	column  map[string]columnFunc
	headers []string
}

func NewTransactionProcessor() *TransactionProcessor {
	return &TransactionProcessor{
		column: map[string]columnFunc{
			"ID": func(val string, tx *proto.SrcTx) error {
				tx.TxID = val
				return nil
			},
			"BaseCurrency": func(val string, tx *proto.SrcTx) error {
				if val != "" {
					tx.Base = val
				}
				return nil
			},
			"QuoteCurrency": func(val string, tx *proto.SrcTx) error {
				if val != "" {
					tx.Quote = val
				}
				return nil
			},
			"Fee": func(val string, tx *proto.SrcTx) error {
				tx.Fee = val
				return nil
			},
			"FeeCurrency": func(val string, tx *proto.SrcTx) error {
				tx.FeeCurrency = val
				return nil
			},
			"FeeRate": func(val string, tx *proto.SrcTx) error {
				tx.FeeRate = val
				return nil
			},
			"Market": func(val string, tx *proto.SrcTx) error {
				tx.Ticker = val
				tx.MarketType = proto.MarketType_SPOT
				if strings.HasSuffix(val, "-PERP") {
					tx.MarketType = proto.MarketType_FUTURE
					tx.Base = strings.TrimSuffix(val, "-PERP")
					tx.Quote = "USD"
				}
				return nil
			},
			"Liquidity": func(val string, tx *proto.SrcTx) error {
				tx.OrderType = proto.OrderType_TAKER
				if val == "maker" {
					tx.OrderType = proto.OrderType_MAKER
				}
				return nil
			},
			"OrderID": func(val string, tx *proto.SrcTx) error {
				tx.OrderID = val
				return nil
			},
			"Price": func(val string, tx *proto.SrcTx) error {
				tx.Price = val
				return nil
			},
			"Side": func(val string, tx *proto.SrcTx) error {
				tx.Action = proto.TxAction_BUY

				if val == "sell" {
					tx.Action = proto.TxAction_SELL
				}
				return nil
			},
			"Size": func(val string, tx *proto.SrcTx) error {
				tx.Amount = val
				return nil
			},
			"Time": func(val string, tx *proto.SrcTx) error {
				ts, err := time.Parse("2006-01-02 15:04:05.999999 -0700 -0700", val)
				tx.Ts = timestamppb.New(ts)
				return err
			},
		},
	}
}

func (t *TransactionProcessor) Parse(content []byte, account, fileName string) {
	reader := csv.NewReader(bytes.NewReader(content))

	line := 0
	t.headers = []string{}

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}

		line++
		if err != nil {
			golog.Errorf("Error while parsing %s: %v", fileName, err)
			break
		}

		if line == 1 {
			t.headers = rec
			continue
		}

		tx := &proto.SrcTx{}

		for i, h := range t.headers {
			val := rec[i]

			if fn, ok := t.column[h]; ok {
				fn(val, tx)
			}
		}

		recType := t.getValue(rec, "Type")

		if recType == "otc" {
			tx.Comment = "OTC Trade"
		}

		tx.Account = account
		tx.Value = global.StrToDecimal(tx.Amount).Mul(global.StrToDecimal(tx.Price)).String()

		err = g.GrpcClient.SubmitTransaction(context.Background(), tx)
		if err != nil {
			golog.Errorf("Failed to send transaction to F-Taxes: %v", err)
			break
		}
	}
}

func (t *TransactionProcessor) getValue(rec []string, name string, defValue ...string) string {
	idx := t.headerIndex(name)

	if idx == -1 {
		if len(defValue) > 0 {
			return defValue[0]
		}
		return ""
	}

	val := rec[idx]
	if val == "" && len(defValue) > 0 {
		return defValue[0]
	}

	return val
}

func (t *TransactionProcessor) headerIndex(name string) int {
	for i, v := range t.headers {
		if v == name {
			return i
		}
	}

	return -1
}
