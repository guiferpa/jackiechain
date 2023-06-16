package blockchain

import "github.com/guiferpa/jackiechain/wallet"

func GetUTXOsByWalletAddress(chain Chain, w wallet.Wallet) ([]TransactionInput, error) {
	addr := w.GetAddress()

	txs := make([]TransactionInput, 0)
	for _, txout := range chain.UTXO {
		txin := txout.ToInput()

		if txin.Sender == addr {
			txs = append(txs, *txin)
		}
	}

	return txs, nil
}
