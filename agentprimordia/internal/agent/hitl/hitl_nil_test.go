package hitl

import (
	"context"
	"errors"
	"testing"
)

func TestRequestInterrupt_NilResponseIsRejected(t *testing.T) {
	ch := make(chan *HumanResponse, 1)
	m := NewHITLManager(HITLConfig{
		InterruptPoints: []InterruptPoint{{Type: InterruptToolConfirm}},
		HumanInputChan:  ch,
	})
	ch <- nil

	resp, err := m.RequestInterrupt(context.Background(), &InterruptRequest{Reason: InterruptToolConfirm})
	if !errors.Is(err, ErrNilHumanResponse) {
		t.Fatalf("expected ErrNilHumanResponse, got resp=%+v err=%v", resp, err)
	}
}
