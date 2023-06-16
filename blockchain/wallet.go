package blockchain

func GetUTXOsByWalletAddress(chain Chain, addr string) ([]TransactionInput, error) {
	txs := make([]TransactionInput, 0)
	for _, txout := range chain.UTXO {
		txin := txout.ToInput()

		if txin.Sender == addr {
			txs = append(txs, *txin)
		}
	}

	return txs, nil
}
