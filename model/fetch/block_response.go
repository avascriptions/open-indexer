package fetch

type Transaction struct {
	BlockHash            string        `json:"blockHash"`
	BlockNumber          string        `json:"blockNumber"`
	From                 string        `json:"from"`
	Gas                  string        `json:"gas"`
	GasPrice             string        `json:"gasPrice"`
	MaxFeePerGas         string        `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string        `json:"maxPriorityFeePerGas"`
	Hash                 string        `json:"hash"`
	Input                string        `json:"input"`
	Nonce                string        `json:"nonce"`
	To                   string        `json:"to"`
	TransactionIndex     string        `json:"transactionIndex"`
	Value                string        `json:"value"`
	Type                 string        `json:"type"`
	AccessList           []interface{} `json:"accessList"`
	ChainId              string        `json:"chainId"`
	V                    string        `json:"v"`
	R                    string        `json:"r"`
	S                    string        `json:"s"`
	YParity              string        `json:"yParity"`
	// --- Additional fields
	Timestamp uint64
}

type Block struct {
	Id string `json:"id"`

	Difficulty       string         `json:"difficulty"`
	ExtraData        string         `json:"extraData"`
	GasLimit         string         `json:"gasLimit"`
	GasUsed          string         `json:"gasUsed"`
	Hash             string         `json:"hash"`
	LogsBloom        string         `json:"logsBloom"`
	Miner            string         `json:"miner"`
	MixHash          string         `json:"mixHash"`
	Nonce            string         `json:"nonce"`
	Number           string         `json:"number"`
	ParentHash       string         `json:"parentHash"`
	ReceiptsRoot     string         `json:"receiptsRoot"`
	Sha3Uncles       string         `json:"sha3Uncles"`
	Size             string         `json:"size"`
	StateRoot        string         `json:"stateRoot"`
	Timestamp        string         `json:"timestamp"`
	TotalDifficulty  string         `json:"totalDifficulty"`
	Transactions     []*Transaction `json:"transactions"`
	TransactionsRoot string         `json:"transactionsRoot"`
	Uncles           []interface{}  `json:"uncles"`
	//Withdrawals      []struct {
	//	Address        string `json:"address"`
	//	Amount         string `json:"amount"`
	//	Index          string `json:"index"`
	//	ValidatorIndex string `json:"validatorIndex"`
	//} `json:"withdrawals"`
	//WithdrawalsRoot string `json:"withdrawalsRoot"`
}

type BlockResponse struct {
	Id      string `json:"id"`
	JsonRpc string `json:"jsonrpc"`
	Result  *Block `json:"result"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
