package domain

type ChatStats struct {
	MessagesReceived int
	MessagesSent     int
	ClientsInTheRoom int
}

func NewChatStats() *ChatStats {
	return &ChatStats{}
}

func (s *ChatStats) IncrementReceived() {
	s.MessagesReceived++
}

func (s *ChatStats) IncrementSent() {
	s.MessagesSent++
}

func (s *ChatStats) IncrementClients() {
	s.ClientsInTheRoom++
}

func (s *ChatStats) DecrementClients() {
	s.ClientsInTheRoom--
}

func (s *ChatStats) ResetClients(n int) {
	s.ClientsInTheRoom = n
}
