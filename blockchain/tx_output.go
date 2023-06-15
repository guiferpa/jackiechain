package blockchain

import (
	"github.com/mr-tron/base58/base58"
)

type TransactionOutput struct {
	Signature []byte
	Receiver  string
	Amount    int
}

type TransactionOutputOptions struct {
	Receiver string
	Amount   int
}

func NewTransactionOutput(opts TransactionOutputOptions) *TransactionOutput {
	return &TransactionOutput{
		Receiver: opts.Receiver,
		Amount:   opts.Amount,
	}
}

func (to *TransactionOutput) SignTransactionOutput() error {
	// Getting public key from wallet address
	_, err := base58.Decode(to.Receiver)
	if err != nil {
		return err
	}

	return nil
}

func (to *TransactionOutput) ToInput() *TransactionInput {
	return &TransactionInput{
		Signature: to.Signature,
		Sender:    to.Receiver,
		Amount:    to.Amount,
	}
}

func NewSignedTransactionInput(opts TransactionOutputOptions) (*TransactionOutput, error) {
	tx := NewTransactionOutput(opts)
	tx.SignTransactionOutput()
	return tx, nil
}
