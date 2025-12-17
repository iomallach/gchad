package application

type ChatClient interface {
	Connect(ulr string) error
	Disconnect() error
	SendMessage(message string)
	InboundMessages() <-chan any // TODO: use domain stuff here later
	Errors() <-chan error
}
