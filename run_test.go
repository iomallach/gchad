package gchad_test

import (
	"testing"

	"github.com/iomallach/gchad"
)

func TestMapConnectMeMessageSucceeds(t *testing.T) {
	body := []byte(`{"type":"connect_me","data":{"name":"John Doe"}}`)
	expected := gchad.ConnectMeMessage{Name: "John Doe"}

	res, err := gchad.MapConnectMeMessage(body)
	if err != nil {
		t.Errorf("expected no error, got %s", err.Error())
	}

	if *res != expected {
		t.Errorf("expected name=%s, got name=%s", "John Doe", res.Name)
	}
}

func TestMapConnectMeMessageFailsOnInvalidJSON(t *testing.T) {
	body := []byte(`invalid json`)

	res, err := gchad.MapConnectMeMessage(body)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}

	if res != nil {
		t.Errorf("expected nil result, got %v", res)
	}
}

func TestMapConnectMeMessageFailsOnWrongType(t *testing.T) {
	body := []byte(`{"type":"user_msg","data":{"timestamp":"2023-01-01T00:00:00Z","message":"Hello"}}`)

	res, err := gchad.MapConnectMeMessage(body)
	if err == nil {
		t.Error("expected error for wrong message type, got nil")
	}

	if res != nil {
		t.Errorf("expected nil result, got %v", res)
	}
}
