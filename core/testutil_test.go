package core

import (
	"context"
	"fmt"
	"github.com/gl-ot/light-mq/config"
	"github.com/gl-ot/light-mq/core/message/idxrepo"
	"github.com/gl-ot/light-mq/core/offset/offsetrepo"
	"github.com/gl-ot/light-mq/testutil"
	"github.com/magiconair/properties/assert"
	"log"
	"testing"
	"time"
)

const (
	defualtGroup = "my-group"
)

func init() {
	log.SetFlags(0)
	log.SetOutput(testutil.StdWriter{})
}

func setup(t *testing.T, testName string) {
	err := testutil.LogSetup("pubsub_" + testName)
	if err != nil {
		t.Fatal(err)
	}
	offsetrepo.InitStorage()
	idxrepo.InitIndex()
}

func publish(t *testing.T) {
	publishWithId(t, 0)
}

func publishWithId(t *testing.T, pubId int) {
	log.Printf("Pub_%d starting publishing %d messages\n", pubId, publishCount)
	for n := 0; n < publishCount; n++ {
		err := Publish(topic, []byte(msg(pubId, n)))
		if err != nil {
			t.Fatal("Publish failed", err)
		}
	}
	log.Printf("Pub_%d finished publishing %d messages\n", pubId, publishCount)
}

func subscribe(t *testing.T) {
	subscribeGroupManyPubs(t, defualtGroup, 1)
}

func subscribeGroup(t *testing.T, group string) {
	subscribeGroupManyPubs(t, group, 1)
}

func subscribeManyPubs(t *testing.T, numberOfPubs int) {
	subscribeGroupManyPubs(t, defualtGroup, numberOfPubs)
}

func subscribeGroupManyPubs(t *testing.T, group string, numberOfPubs int) {
	s := newTestSubscriber(t, group)
	defer s.Close()

	startReceivingManyPubs(t, s, numberOfPubs)
}

func newTestSubscriber(t *testing.T, group string) *Subscriber {
	s, err := NewSub(topic, group)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func startReceivingManyPubs(t *testing.T, s *Subscriber, numberOfPubs int) {
	log.Printf("Starting subscribing %d messages\n", publishCount)

	// index represents pubId, and value is count of received messages
	msgCounts := make([]int, numberOfPubs)

	var ctx context.Context
	var cancel context.CancelFunc
	if config.Props.Stdout.Level == "debug" {
		ctx, cancel = context.WithCancel(context.Background())
	} else {
		ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	}

	err := s.Subscribe(ctx, func(message []byte) error {
		m := string(message)
		for i, v := range msgCounts {
			if m == msg(i, v) {
				msgCounts[i]++
				if areAllMessagesSent(msgCounts) {
					cancel()
				}
				return nil
			}
		}
		t.Fatalf("Message out of order: message=%s", message)
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %s", err)
	}
	log.Printf("Finished subscribing %d messages\n", publishCount)

	for _, v := range msgCounts {
		assert.Equal(t, v, publishCount, "Message count wrong! (Probably deadline limit)")
	}
}

func areAllMessagesSent(msgCounts []int) bool {
	all := true
	for _, v := range msgCounts {
		if v != publishCount {
			all = false
		}
	}
	return all
}

func msg(pubId, msgId int) string {
	return fmt.Sprintf("pub_%d_%s_%d", pubId, message, msgId)
}
