package pubsub

import (
	"fmt"
	"testing"

	"github.com/fission/fission-workflow/pkg/util/labels/kubelabels"
)

func TestPublisherSubscribe(t *testing.T) {
	pub := NewPublisher()
	defer pub.Close()
	sub := pub.Subscribe()

	if sub == nil {
		t.Error("Empty subscription provided")
	}
}

func TestPublish(t *testing.T) {
	pub := NewPublisher()
	sub := pub.Subscribe(SubscriptionOptions{
		Buf: 1,
	})

	msg := NewGenericMsg(kubelabels.New(map[string]string{
		"foo": "bar",
	}), "TestMsg")

	err := pub.Publish(msg)
	if err != nil {
		t.Error(err)
	}
	pub.Close()

	err = expectMsgs(sub, []Msg{
		msg,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestPublishBufferOverflow(t *testing.T) {
	pub := NewPublisher()
	sub := pub.Subscribe(SubscriptionOptions{
		Buf: 1,
	})
	sub2 := pub.Subscribe(SubscriptionOptions{
		Buf: 10,
	})

	firstMsg := NewGenericMsg(kubelabels.New(map[string]string{
		"foo": "bar",
	}), "TestMsg1")

	secondMsg := NewGenericMsg(kubelabels.New(map[string]string{
		"foo": "bar",
	}), "TestMsg2")

	err := pub.Publish(firstMsg)
	if err != nil {
		t.Error(err)
	}
	err = pub.Publish(secondMsg)
	if err != nil {
		t.Error(err)
	}
	pub.Close()

	err = expectMsgs(sub, []Msg{
		firstMsg,
	})
	if err != nil {
		t.Error(err)
	}

	err = expectMsgs(sub2, []Msg{
		firstMsg,
		secondMsg,
	})
	if err != nil {
		t.Error(err)
	}
}

// Note ensure that subscriptions are closed before this check
func expectMsgs(sub *Subscription, expectedMsgs []Msg) error {
	i := 0
	for msg := range sub.Ch {
		if i > len(expectedMsgs) {
			return fmt.Errorf("Received unexpected msg '%v'", msg)
		}
		if msg != expectedMsgs[i] {
			return fmt.Errorf("Received msg '%v' does not equal send msg '%v'", msg, expectedMsgs[i])
		}
		i = i + 1
	}
	if i != len(expectedMsgs) {
		return fmt.Errorf("Did not receive expected msgs: %v", expectedMsgs[i+1:])
	}
	return nil
}
