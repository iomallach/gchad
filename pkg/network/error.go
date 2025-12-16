package network

import (
	"errors"
	"io"
	"net"

	"github.com/gorilla/websocket"
)

var (
	ErrConnectionClosedNormally   = errors.New("connection closed normally")
	ErrConnectionClosedAbnormally = errors.New("connection closed abnormally")
	ErrMessageTooLarge            = errors.New("message exceeds size limit")
	ErrNetworkFailure             = errors.New("network failure")
	ErrReadTimeOut                = errors.New("read timeout")

	ErrWriteAfterClose = errors.New("write after close")
	ErrWriteTimeout    = errors.New("write timeout")
	ErrWriteFailed     = errors.New("write unexpectedly failed")
)

func TranslateWriteError(err error) error {
	if err == nil {
		return nil
	}

	if err == websocket.ErrCloseSent {
		return ErrWriteAfterClose
	}

	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return ErrWriteTimeout
	}

	return ErrWriteFailed
}

func TranslateReadError(err error) error {
	if websocket.IsCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
	) {
		return ErrConnectionClosedNormally
	}

	if websocket.IsCloseError(
		err,
		websocket.CloseProtocolError,
		websocket.CloseUnsupportedData,
		websocket.CloseNoStatusReceived,
		websocket.CloseAbnormalClosure,
		websocket.CloseInvalidFramePayloadData,
		websocket.CloseInternalServerErr,
		websocket.CloseTryAgainLater,
	) {
		return ErrConnectionClosedAbnormally
	}

	if _, ok := err.(*websocket.CloseError); ok {
		return ErrConnectionClosedAbnormally
	}

	if err == io.EOF {
		return ErrConnectionClosedAbnormally
	}

	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return ErrReadTimeOut
	}

	if err == websocket.ErrReadLimit {
		return ErrMessageTooLarge
	}

	return ErrNetworkFailure
}
