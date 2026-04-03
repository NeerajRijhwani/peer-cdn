package tracker

// import (
// 	"encoding/json"
// 	"fmt"

// 	"github.com/NeerajRijhwani/peer-cdn/internal/storage"
// )

// func parseMessage(data []byte)(interface{},error){
// 	// convert the json message to the struct of given choiceand return the struct
// 	var baseMsg struct{
// 		Type string `json:type`
// 	}
// 	if err := json.Unmarshal(data, &baseMsg); err != nil {
// 		return nil, err
// 	}
//  switch baseMsg.Type {
// 	case storage.MsgTypeAnnounce:
// 		var msg storage.AnnounceRequest
// 		if err := json.Unmarshal(data, &msg); err != nil {
// 			return nil, err
// 		}
// 		return &msg, nil

// 	case storage.MsgTypeScrape:
// 		var msg storage.ScrapeRequest
// 		if err := json.Unmarshal(data, &msg); err != nil {
// 			return nil, err
// 		}
// 		return &msg, nil

// 	default:
// 		return nil, fmt.Errorf("unknown message type: %s", baseMsg.Type)
// 	}
// }

// func EncodeMessage(msg interface{}) ([]byte, error) {
// 	return json.Marshal(msg)  // convert the message to json and return
// }