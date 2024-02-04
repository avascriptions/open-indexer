package fetch

type LogEvent struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
	// --- Additional fields
	Timestamp uint64
}

type LogsResponse struct {
	Id      string      `json:"id"`
	JsonRpc string      `json:"jsonrpc"`
	Result  []*LogEvent `json:"result"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
