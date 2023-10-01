package transaction

type TxOut struct {
	Receiver string
	Value    int64
}

type TxOutSlice []TxOut

func GenerateUTxOFromTxOut(txo TxOut) UTxO {
	return UTxO{
		Receiver: txo.Receiver,
		Value:    txo.Value,
	}
}
