package model

type EvmLog struct {
	Hash      string
	Address   string
	Topic0    string
	Topic1    string
	Topic2    string
	Topic3    string
	Data      string
	Block     uint64
	TrxIndex  uint32
	LogIndex  uint32
	Timestamp uint64
}

//func NewEvmLogFromMixData(mixData *MixData) *EvmLog {
//	logEvent := mixData.LogEvent
//	topic0 := ""
//	topic1 := ""
//	topic2 := ""
//	topic3 := ""
//	if len(logEvent.Topics) > 0 {
//		topic0 = logEvent.Topics[0]
//		if len(logEvent.Topics) > 1 {
//			topic1 = logEvent.Topics[1]
//			if len(logEvent.Topics) > 2 {
//				topic2 = logEvent.Topics[2]
//				if len(logEvent.Topics) > 3 {
//					topic3 = logEvent.Topics[3]
//				}
//			}
//		}
//	}
//	var log = EvmLog{
//		Hash:      mixData.Transaction.Hash,
//		Address:   logEvent.Address,
//		Topic0:    topic0,
//		Topic1:    topic1,
//		Topic2:    topic2,
//		Topic3:    topic3,
//		Data:      logEvent.Data,
//		Block:     mixData.BlockNumber,
//		TrxIndex:  mixData.TransactionIndex,
//		LogIndex:  mixData.LogIndex,
//		Timestamp: mixData.TimeStamp,
//		Type:      0,
//		Status:    0,
//	}
//
//	return &log
//}
