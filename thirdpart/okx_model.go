package thirdpart

type TransactionListResponse struct {
	Cursor          string        `json:"cursor"`
	TransactionList []Transaction `json:"transactionList"`
}

type Transaction struct {
	ChainIndex   string        `json:"chainIndex"`
	TxHash       string        `json:"txHash"`
	MethodID     string        `json:"methodId"`
	Nonce        string        `json:"nonce"`
	TxTime       string        `json:"txTime"`
	From         []AddressInfo `json:"from"`
	To           []AddressInfo `json:"to"`
	TokenAddress string        `json:"tokenAddress"`
	Amount       string        `json:"amount"`
	Symbol       string        `json:"symbol"`
	TxFee        string        `json:"txFee"`
	TxStatus     string        `json:"txStatus"`
	HitBlacklist bool          `json:"hitBlacklist"`
	Tag          string        `json:"tag"`
	IType        string        `json:"itype"`
}

type AddressInfo struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}
