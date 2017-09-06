package pubsub

import (
	"fmt"
	"testing"

	"github.com/fission/fission-workflow/pkg/util/labels/kubelabels"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestPublisherSubscribe(t *testing.T) {
	pub := NewPublisher()
	defer pub.Close()
	sub := pub.Subscribe()

	assert.NotNil(t, sub)
}

func TestPublish(t *testing.T) {
	pub := NewPublisher()
	sub := pub.Subscribe(SubscriptionOptions{
		Buf: 1,
	})

	msg := NewGenericMsg(kubelabels.New(map[string]string{
		"foo": "bar",
	}), time.Now(), "TestMsg")

	err := pub.Publish(msg)
	assert.NoError(t, err)
	pub.Close()

	err = expectMsgs(sub, []Msg{
		msg,
	})
	assert.NoError(t, err)
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
	}), time.Now(), "TestMsg1")

	secondMsg := NewGenericMsg(kubelabels.New(map[string]string{
		"foo": "bar",
	}), time.Now(), "TestMsg2")

	err := pub.Publish(firstMsg)
	assert.NoError(t, err)

	err = pub.Publish(secondMsg)
	assert.NoError(t, err)
	pub.Close()

	err = expectMsgs(sub, []Msg{
		firstMsg,
	})
	assert.NoError(t, err)

	err = expectMsgs(sub2, []Msg{
		firstMsg,
		secondMsg,
	})
	assert.NoError(t, err)
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
