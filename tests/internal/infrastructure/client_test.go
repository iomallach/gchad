package infrastructure_test

import (
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/iomallach/gchad/internal/domain"
	"github.com/iomallach/gchad/internal/infrastructure"
	"github.com/stretchr/testify/assert"
)

const (
	CloseMessage = iota
	TextMessage
	PingMessage
)

type readResult struct {
	messageType int
	data        []byte
	err         error
}

type writeResult struct {
	messageType int
	data        []byte
}

type MockConnection struct {
	readChan  chan readResult
	writeChan chan writeResult
	closed    bool
	mu        sync.Mutex
}

func NewMockConnection() *MockConnection {
	return &MockConnection{
		readChan:  make(chan readResult),
		writeChan: make(chan writeResult),
		closed:    false,
		mu:        sync.Mutex{},
	}
}

func (mc *MockConnection) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	close(mc.readChan)
	close(mc.writeChan)
	mc.closed = true

	return nil
}

func (mc *MockConnection) ReadMessage() (int, []byte, error) {
	msg, ok := <-mc.readChan
	if !ok {
		return 0, nil, io.EOF
	}

	return msg.messageType, msg.data, msg.err
}

func (mc *MockConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

func (mc *MockConnection) writeMessage(messageType int, data []byte) error {
	if mc.closed {
		return io.ErrClosedPipe
	}

	mc.writeChan <- writeResult{
		messageType: messageType,
		data:        data,
	}
	return nil
}
func (mc *MockConnection) WriteCloseMessage(data []byte) error {
	return mc.writeMessage(CloseMessage, data)
}

func (mc *MockConnection) WriteTextMessage(data []byte) error {
	return mc.writeMessage(TextMessage, data)
}

func (mc *MockConnection) WritePingMessage(data []byte) error {
	return mc.writeMessage(PingMessage, data)
}

func (mc *MockConnection) EnqueueMessage(messageType int, data []byte) {
	mc.readChan <- readResult{messageType, data, nil}
}

func (mc *MockConnection) DrainWrites() []writeResult {
	writes := make([]writeResult, 0)

	for {
		select {
		case write := <-mc.writeChan:
			writes = append(writes, write)
		case <-time.After(time.Millisecond * 50):
			return writes
		}
	}
}

func mustMarshallMessage(msg domain.Messager) []byte {
	data, err := domain.MarshallMessage(msg)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal test message: %v", err))
	}

	return data
}

func NewTestingClientConfiguration() infrastructure.ClientConfiguration {
	return infrastructure.ClientConfiguration{
		WriteWait:       50 * time.Millisecond,
		PongWait:        50 * time.Millisecond,
		PingPeriod:      50 * time.Millisecond,
		RecieveChanWait: 50 * time.Millisecond,
		SendChannelSize: 3,
	}
}

func TestClient_ReadMessagesPumpSendsMessagesToRecv(t *testing.T) {
	ctx := t.Context()

	configuration := NewTestingClientConfiguration()
	connection := NewMockConnection()
	defer connection.Close()
	spyLogger := NewSpyLogger()
	recv := make(chan domain.Messager, 3)
	client := infrastructure.NewClient("1", "Jane Doe", connection, recv, nil, configuration, spyLogger)

	go client.ReadMessages(ctx)

	userMsg := domain.NewUserMessage("Hello test", time.Now(), "John Doe")
	connection.EnqueueMessage(
		TextMessage,
		mustMarshallMessage(userMsg),
	)

	time.Sleep(time.Millisecond * 50)

	assert.Len(t, spyLogger.Errors(), 0)
	receivedMsg := (<-recv).(*domain.UserMessage)
	assert.Equal(t, userMsg.Message, receivedMsg.Message)
	assert.Equal(t, userMsg.Timestamp.Truncate(time.Second), receivedMsg.Timestamp.Truncate(time.Second))
	assert.Equal(t, userMsg.From, receivedMsg.From)
}

func TestClient_WriteMessagesPumpSendsMessagesToRecv(t *testing.T) {
	ctx := t.Context()

	configuration := NewTestingClientConfiguration()
	connection := NewMockConnection()
	spyLogger := NewSpyLogger()
	recv := make(chan domain.Messager, 3)
	send := make(chan domain.Messager, 3)
	client := infrastructure.NewClient("1", "Jane Doe", connection, recv, send, configuration, spyLogger)

	go client.WriteMessages(ctx)

	userMsg := domain.NewUserMessage("Hello test", time.Now(), "Jane Doe")
	userMsgBytes := mustMarshallMessage(userMsg)
	// imitate notifier sending a single user message
	// it is then expected to be marshalled and written as a text message
	client.Send() <- userMsg

	time.Sleep(time.Millisecond * 50)

	assert.Len(t, spyLogger.Errors(), 0)

	select {
	case writtenMsg := <-connection.writeChan:
		assert.Equal(t, TextMessage, writtenMsg.messageType)
		assert.Equal(t, userMsgBytes, writtenMsg.data)
	case <-time.After(50 * time.Millisecond):
		t.Fatal("message was not written")
	}
}
