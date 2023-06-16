package blockchain

import (
	"bytes"
	"fmt"
	"time"
)

type TransactionOutput struct {
	Signature []byte    `json:"signature"`
	Receiver  string    `json:"receiver"`
	Amount    int       `json:"amount"`
	Timestamp time.Time `json:"-"`
}

func (to *TransactionOutput) Buffer() *bytes.Buffer {
	return bytes.NewBufferString(fmt.Sprintf(
		"%s::%s::%v",
		to.Receiver,
		to.Amount,
		to.Timestamp,
	))
}

func (to *TransactionOutput) ToInput() *TransactionInput {
	return &TransactionInput{
		Signature:    to.Signature,
		Sender:       to.Receiver,
		Amount:       to.Amount,
		TxOutputHash: CalculateTxEntryHash(to),
	}
}

type TransactionOutputOptions struct {
	Receiver string
	Amount   int
}

func NewTransactionOutput(opts TransactionOutputOptions) *TransactionOutput {
	return &TransactionOutput{
		Receiver:  opts.Receiver,
		Amount:    opts.Amount,
		Timestamp: time.Now(),
	}
}
