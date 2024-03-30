package deposits

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"regexp"
	"time"

	g "github.com/f-taxes/csv_import/grpc_client"
	"github.com/f-taxes/csv_import/proto"
	"github.com/kataras/golog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type columnFunc func(val string, tx *proto.Transfer) error

type DepositProcessor struct {
	column  map[string]columnFunc
	headers []string
}

var re = regexp.MustCompile(`^Transfer from (.*) to .+$`)

func NewDepositProcessor() *DepositProcessor {
	return &DepositProcessor{
		column: map[string]columnFunc{
			"ID": func(val string, t *proto.Transfer) error {
				t.TxID = val
				return nil
			},
			"Coin": func(val string, t *proto.Transfer) error {
				t.Asset = val
				t.Action = proto.TransferAction_DEPOSIT
				return nil
			},
			"Fee": func(val string, t *proto.Transfer) error {
				t.Fee = val
				t.FeeCurrency = "USD"
				return nil
			},
			"Size": func(val string, t *proto.Transfer) error {
				t.Amount = val
				return nil
			},
			"Time": func(val string, t *proto.Transfer) error {
				ts, err := time.Parse("2006-01-02 15:04:05.999999 -0700 -0700", val)
				t.Ts = timestamppb.New(ts)
				return err
			},
			"Notes": func(val string, t *proto.Transfer) error {
				t.Comment = val

				matches := re.FindStringSubmatch(val)
				if len(matches) > 1 {
					t.Source = matches[1]

					if t.Source == "main account" {
						t.Source = "Main"
					}
				}

				return nil
			},
		},
	}
}

func (t *DepositProcessor) Parse(content []byte, account, fileName string) {
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

		transfer := &proto.Transfer{}

		for i, h := range t.headers {
			val := rec[i]

			if fn, ok := t.column[h]; ok {
				fn(val, transfer)
			}
		}

		transfer.Account = account
		transfer.Destination = account

		err = g.GrpcClient.SubmitTransfer(context.Background(), transfer)
		if err != nil {
			golog.Errorf("Failed to send transaction to F-Taxes: %v", err)
			break
		}
	}
}

func (t *DepositProcessor) getValue(rec []string, name string, defValue ...string) string {
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

func (t *DepositProcessor) headerIndex(name string) int {
	for i, v := range t.headers {
		if v == name {
			return i
		}
	}

	return -1
}
