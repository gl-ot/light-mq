package gate

import (
	"github.com/gl-ot/light-mq/pubsub/message/msgservice"
	log "github.com/sirupsen/logrus"
)

// todo every operation is not thread safe!!!

var streamingGates = make(map[string]map[string]chan *msgservice.Message)

// For publishers

func SendMessage(topic string, message *msgservice.Message) {
	subGroups := streamingGates[topic]
	for _, groupChan := range subGroups {
		groupChan <- message
	}
}

// For subscribers

func Open(topic string, subscriberGroup string) {
	subGroups, ok := streamingGates[topic]
	if !ok {
		subGroups = make(map[string]chan *msgservice.Message)
		streamingGates[topic] = subGroups
	}

	subGroups[subscriberGroup] = make(chan *msgservice.Message, 16) // todo put buffer size into config
}

func GetMessageChannel(topic string, group string) <-chan *msgservice.Message {
	subGroups, ok := streamingGates[topic]
	if !ok {
		log.Fatalf("Couldn't find topic, open gate before obtaining message channel, topic=%s", topic)
	}
	return subGroups[group]
}

func Close(topic string, subscriberGroup string) {
	subs, ok := streamingGates[topic]
	if ok {
		msgChan := subs[subscriberGroup]
		delete(subs, subscriberGroup)
		close(msgChan)
	} else {
		log.Warnf("Didn't find subscriber group in topic: topic=%s, group: %s", topic, subscriberGroup)
	}
}