package infrastructure_test

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/iomallach/gchad/internal/server/domain"
	"github.com/iomallach/gchad/internal/server/infrastructure"
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
	readChan        chan readResult
	writes          []writeResult
	closed          bool
	mu              sync.Mutex
	writeTextError  error
	writePingError  error
	writeCloseError error
}

func NewMockConnection() *MockConnection {
	return &MockConnection{
		readChan: make(chan readResult),
		writes:   make([]writeResult, 0),
		closed:   false,
		mu:       sync.Mutex{},
	}
}

func (mc *MockConnection) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !mc.closed {
		close(mc.readChan)
		mc.closed = true
	}

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
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.closed {
		return io.ErrClosedPipe
	}

	mc.writes = append(mc.writes, writeResult{
		messageType: messageType,
		data:        data,
	})
	return nil
}
func (mc *MockConnection) WriteCloseMessage(data []byte) error {
	if mc.writeCloseError != nil {
		return mc.writeCloseError
	}
	return mc.writeMessage(CloseMessage, data)
}

func (mc *MockConnection) WriteTextMessage(data []byte) error {
	if mc.writeTextError != nil {
		return mc.writeTextError
	}
	return mc.writeMessage(TextMessage, data)
}

func (mc *MockConnection) WritePingMessage(data []byte) error {
	if mc.writePingError != nil {
		return mc.writePingError
	}
	return mc.writeMessage(PingMessage, data)
}

func (mc *MockConnection) EnqueueMessage(messageType int, data []byte) {
	mc.readChan <- readResult{messageType, data, nil}
}

func (mc *MockConnection) EnqueueError(err error) {
	mc.readChan <- readResult{0, nil, err}
}

func (mc *MockConnection) GetWrites() []writeResult {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	result := make([]writeResult, len(mc.writes))
	copy(result, mc.writes)
	return result
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

	writes := connection.GetWrites()
	foundMessage := false
	for _, write := range writes {
		if write.messageType == TextMessage && string(write.data) == string(userMsgBytes) {
			foundMessage = true
			break
		}
	}
	assert.True(t, foundMessage, "expected text message to be written")
}

// ReadMessages failure tests

func TestClient_ReadMessages_InvalidJSON(t *testing.T) {
	ctx := t.Context()

	configuration := NewTestingClientConfiguration()
	connection := NewMockConnection()
	defer connection.Close()
	spyLogger := NewSpyLogger()
	recv := make(chan domain.Messager, 3)
	client := infrastructure.NewClient("1", "Jane Doe", connection, recv, nil, configuration, spyLogger)

	go client.ReadMessages(ctx)

	// Send invalid JSON
	connection.EnqueueMessage(TextMessage, []byte("{invalid json"))

	time.Sleep(time.Millisecond * 50)

	// Should log error but not crash
	assert.GreaterOrEqual(t, len(spyLogger.Errors()), 1)
	assert.Contains(t, spyLogger.Errors()[0].msg, "failed to unmarshall")
	// recv channel should be empty
	assert.Len(t, recv, 0)
}

func TestClient_ReadMessages_ReadError(t *testing.T) {
	ctx := t.Context()

	configuration := NewTestingClientConfiguration()
	connection := NewMockConnection()
	defer connection.Close()
	spyLogger := NewSpyLogger()
	recv := make(chan domain.Messager, 3)
	client := infrastructure.NewClient("1", "Jane Doe", connection, recv, nil, configuration, spyLogger)

	go client.ReadMessages(ctx)

	// Enqueue a read error
	connection.EnqueueError(fmt.Errorf("connection broken"))

	time.Sleep(time.Millisecond * 50)

	// Should log error and continue
	assert.GreaterOrEqual(t, len(spyLogger.Errors()), 1)
	assert.Contains(t, spyLogger.Errors()[0].msg, "could not read the message")
	// recv channel should be empty
	assert.Len(t, recv, 0)
}

func TestClient_ReadMessages_RecvChannelFull(t *testing.T) {
	ctx := t.Context()

	configuration := NewTestingClientConfiguration()
	connection := NewMockConnection()
	defer connection.Close()
	spyLogger := NewSpyLogger()
	recv := make(chan domain.Messager, 1) // Small buffer
	client := infrastructure.NewClient("1", "Jane Doe", connection, recv, nil, configuration, spyLogger)

	go client.ReadMessages(ctx)

	userMsg1 := domain.NewUserMessage("Message 1", time.Now(), "John Doe")
	connection.EnqueueMessage(TextMessage, mustMarshallMessage(userMsg1))

	time.Sleep(time.Millisecond * 50)

	userMsg2 := domain.NewUserMessage("Message 2", time.Now(), "John Doe")
	connection.EnqueueMessage(TextMessage, mustMarshallMessage(userMsg2))

	time.Sleep(time.Millisecond * 100) // Wait longer than RecieveChanWait

	errors := spyLogger.Errors()
	assert.GreaterOrEqual(t, len(errors), 1)
	foundFullChannelError := false
	for _, err := range errors {
		if err.msg == "message channel is full, skipping message" {
			foundFullChannelError = true
			break
		}
	}
	assert.True(t, foundFullChannelError, "expected 'channel full' error")
}

func TestClient_ReadMessages_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	configuration := NewTestingClientConfiguration()
	connection := NewMockConnection()
	defer connection.Close()
	spyLogger := NewSpyLogger()
	recv := make(chan domain.Messager, 3)
	client := infrastructure.NewClient("1", "Jane Doe", connection, recv, nil, configuration, spyLogger)

	go client.ReadMessages(ctx)

	// Send a message so ReadMessages processes it
	userMsg := domain.NewUserMessage("Test", time.Now(), "John Doe")
	connection.EnqueueMessage(TextMessage, mustMarshallMessage(userMsg))

	time.Sleep(time.Millisecond * 50)

	// Cancel context
	cancel()

	// Send another message to trigger the context check
	connection.EnqueueMessage(TextMessage, mustMarshallMessage(userMsg))

	time.Sleep(time.Millisecond * 50)

	// Should log debug about cancellation
	debugs := spyLogger.Debugs()
	if len(debugs) > 0 {
		foundCancellation := false
		for _, d := range debugs {
			if d.msg == "cancelling read pump" {
				foundCancellation = true
				break
			}
		}
		assert.True(t, foundCancellation, "expected cancellation debug log")
	}
}

// WriteMessages failure tests

func TestClient_WriteMessages_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	configuration := NewTestingClientConfiguration()
	connection := NewMockConnection()
	spyLogger := NewSpyLogger()
	recv := make(chan domain.Messager, 3)
	send := make(chan domain.Messager, 3)
	client := infrastructure.NewClient("1", "Jane Doe", connection, recv, send, configuration, spyLogger)

	done := make(chan bool)
	go func() {
		client.WriteMessages(ctx)
		done <- true
	}()

	time.Sleep(time.Millisecond * 50)

	// Cancel context
	cancel()

	// Wait for WriteMessages to exit
	select {
	case <-done:
		// WriteMessages exited due to context cancellation
	case <-time.After(200 * time.Millisecond):
		t.Fatal("WriteMessages should have exited after context cancellation")
	}

	writes := connection.GetWrites()
	foundCloseMessage := false
	for _, write := range writes {
		if write.messageType == CloseMessage {
			foundCloseMessage = true
			break
		}
	}
	assert.True(t, foundCloseMessage, "expected close message on context cancellation")
}

func TestClient_WriteMessages_SendChannelClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	configuration := NewTestingClientConfiguration()
	connection := NewMockConnection()
	spyLogger := NewSpyLogger()
	recv := make(chan domain.Messager, 3)
	send := make(chan domain.Messager, 3)
	client := infrastructure.NewClient("1", "Jane Doe", connection, recv, send, configuration, spyLogger)

	done := make(chan bool)
	go func() {
		client.WriteMessages(ctx)
		done <- true
	}()

	time.Sleep(time.Millisecond * 50)

	// Close send channel
	close(send)

	// Wait for WriteMessages to exit
	select {
	case <-done:
		// WriteMessages exited due to closed channel
	case <-time.After(200 * time.Millisecond):
		t.Fatal("WriteMessages should have exited after send channel closed")
	}

	errors := spyLogger.Errors()
	foundError := false
	for _, err := range errors {
		if err.msg == "send channel has been closed. Sending close message and terminating" {
			foundError = true
			break
		}
	}
	assert.True(t, foundError, "expected send channel closed error")

	writes := connection.GetWrites()
	foundCloseMessage := false
	for _, write := range writes {
		if write.messageType == CloseMessage {
			foundCloseMessage = true
			break
		}
	}
	assert.True(t, foundCloseMessage, "expected close message when send channel closed")
}

func TestClient_WriteMessages_PingSentPeriodically(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	configuration := NewTestingClientConfiguration()
	configuration.PingPeriod = 20 * time.Millisecond // Short period for testing
	connection := NewMockConnection()
	spyLogger := NewSpyLogger()
	recv := make(chan domain.Messager, 3)
	send := make(chan domain.Messager, 3)
	client := infrastructure.NewClient("1", "Jane Doe", connection, recv, send, configuration, spyLogger)

	done := make(chan bool)
	go func() {
		client.WriteMessages(ctx)
		done <- true
	}()

	// Wait for at least one ping
	time.Sleep(time.Millisecond * 100)

	cancel()

	select {
	case <-done:
		// WriteMessages exited
	case <-time.After(200 * time.Millisecond):
		t.Fatal("WriteMessages should have exited after context cancellation")
	}

	// Should have at least one ping message
	writes := connection.GetWrites()
	pingCount := 0
	for _, write := range writes {
		if write.messageType == PingMessage {
			pingCount++
		}
	}

	assert.GreaterOrEqual(t, pingCount, 1, "expected at least one ping message")
}

func TestClient_WriteMessages_WriteTextMessageError(t *testing.T) {
	ctx := t.Context()

	configuration := NewTestingClientConfiguration()
	connection := NewMockConnection()
	connection.writeTextError = fmt.Errorf("write failed")
	spyLogger := NewSpyLogger()
	recv := make(chan domain.Messager, 3)
	send := make(chan domain.Messager, 3)
	client := infrastructure.NewClient("1", "Jane Doe", connection, recv, send, configuration, spyLogger)

	go client.WriteMessages(ctx)

	userMsg := domain.NewUserMessage("Hello test", time.Now(), "Jane Doe")
	client.Send() <- userMsg

	time.Sleep(time.Millisecond * 50)

	// Should log error about write failure
	errors := spyLogger.Errors()
	assert.GreaterOrEqual(t, len(errors), 1)
	assert.Contains(t, errors[0].msg, "failed to write message")
}

func TestClient_WriteMessages_WritePingError(t *testing.T) {
	ctx := t.Context()

	configuration := NewTestingClientConfiguration()
	configuration.PingPeriod = 20 * time.Millisecond
	connection := NewMockConnection()
	connection.writePingError = fmt.Errorf("ping write failed")
	spyLogger := NewSpyLogger()
	recv := make(chan domain.Messager, 3)
	send := make(chan domain.Messager, 3)
	client := infrastructure.NewClient("1", "Jane Doe", connection, recv, send, configuration, spyLogger)

	done := make(chan bool)
	go func() {
		client.WriteMessages(ctx)
		done <- true
	}()

	// Wait for ping to be attempted and WriteMessages to exit
	select {
	case <-done:
		// WriteMessages exited due to ping error
	case <-time.After(100 * time.Millisecond):
		t.Fatal("WriteMessages should have exited after ping error")
	}

	// Should log error about ping failure
	errors := spyLogger.Errors()
	assert.GreaterOrEqual(t, len(errors), 1)
	assert.Contains(t, errors[0].msg, "failed to write ping message")
}

func TestClient_WriteMessages_WriteCloseMessageError(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	configuration := NewTestingClientConfiguration()
	connection := NewMockConnection()
	connection.writeCloseError = fmt.Errorf("close write failed")
	spyLogger := NewSpyLogger()
	recv := make(chan domain.Messager, 3)
	send := make(chan domain.Messager, 3)
	client := infrastructure.NewClient("1", "Jane Doe", connection, recv, send, configuration, spyLogger)

	done := make(chan bool)
	go func() {
		client.WriteMessages(ctx)
		done <- true
	}()

	time.Sleep(time.Millisecond * 50)

	// Cancel context to trigger close message
	cancel()

	// Wait for WriteMessages to finish
	select {
	case <-done:
		// WriteMessages exited
	case <-time.After(100 * time.Millisecond):
		t.Fatal("WriteMessages should have exited")
	}

	// Should log error about close message failure
	errors := spyLogger.Errors()
	assert.GreaterOrEqual(t, len(errors), 1)
	assert.Contains(t, errors[0].msg, "failed to write close message")
}
