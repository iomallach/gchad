package infrastructure

import (
	"encoding/json"
	"fmt"

	"github.com/iomallach/gchad/internal/client/application"
	"github.com/iomallach/gchad/internal/client/domain"
)

func UnmarshallMessage(message []byte) (application.Message, error) {
	var envelope domain.Envelope

	if err := json.Unmarshal(message, &envelope); err != nil {
		return nil, err
	}

	switch envelope.Type {
	case domain.TypeChatMessage:
		var msg *domain.ChatMessage
		if err := json.Unmarshal(envelope.Payload, msg); err != nil {
			return nil, err
		}
		return msg, nil

	case domain.TypeUserJoinedMessage:
		var msg *domain.UserJoinedMessage
		if err := json.Unmarshal(envelope.Payload, msg); err != nil {
			return nil, err
		}
		return msg, nil

	case domain.TypeUserLeftMessage:
		var msg *domain.UserLeftMessage
		if err := json.Unmarshal(envelope.Payload, msg); err != nil {
			return nil, err
		}
		return msg, nil

	default:
		// should be unreachable due to unmarshalling into envelope
		return nil, fmt.Errorf("unexpected message type: %v", envelope.Type)
	}
}
