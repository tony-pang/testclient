package handler

import (
	"log"
	"sync"

	"github.com/centrifugal/centrifuge-go"
)

type DefaultHandler struct{}

func (h *DefaultHandler) OnConnect(_ *centrifuge.Client, e centrifuge.ConnectEvent) {
	log.Printf("Connected to chat with ID %s", e.ClientID)
}

func (h *DefaultHandler) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	log.Printf("Error: %s", e.Message)
}

func (h *DefaultHandler) OnMessage(_ *centrifuge.Client, e centrifuge.MessageEvent) {
	log.Printf("Message from server: %s", string(e.Data))
}

func (h *DefaultHandler) OnDisconnect(_ *centrifuge.Client, e centrifuge.DisconnectEvent) {
	log.Printf("Disconnected from chat: %s", e.Reason)
}

func (h *DefaultHandler) OnServerSubscribe(_ *centrifuge.Client, e centrifuge.ServerSubscribeEvent) {
	log.Printf("Subscribe to server-side channel %s: (resubscribe: %t, recovered: %t)", e.Channel, e.Resubscribed, e.Recovered)
}

func (h *DefaultHandler) OnServerUnsubscribe(_ *centrifuge.Client, e centrifuge.ServerUnsubscribeEvent) {
	log.Printf("Unsubscribe from server-side channel %s", e.Channel)
}

func (h *DefaultHandler) OnServerPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	log.Printf("Publication from server-side channel %s: %s", e.Channel, e.Data)
}

func (h *DefaultHandler) OnPublish(sub *centrifuge.Subscription, e centrifuge.PublishEvent) {
	log.Printf("Someone says via channel %s", sub.Channel())
}

func (h *DefaultHandler) OnSubscribeSuccess(sub *centrifuge.Subscription, e centrifuge.SubscribeSuccessEvent) {
	log.Printf("Subscribed on channel %s, resubscribed: %v, recovered: %v", sub.Channel(), e.Resubscribed, e.Recovered)
}

func (h *DefaultHandler) OnSubscribeError(sub *centrifuge.Subscription, e centrifuge.SubscribeErrorEvent) {
	log.Printf("Subscribed on channel %s failed, error: %s", sub.Channel(), e.Error)
}

func (h *DefaultHandler) OnUnsubscribe(sub *centrifuge.Subscription, _ centrifuge.UnsubscribeEvent) {
	log.Printf("Unsubscribed from channel %s", sub.Channel())
}

type AliasPublisherHandler struct {
	WG       *sync.WaitGroup
	AliasID  string
	expected map[string]bool
	lock     sync.Mutex
	missing  int
}

func NewAliasPublisherHandler(wg *sync.WaitGroup, aliasID string, expected []string) *AliasPublisherHandler {
	expectedMap := make(map[string]bool)
	for _, s := range expected {
		expectedMap[s] = false
	}
	return &AliasPublisherHandler{
		WG:       wg,
		AliasID:  aliasID,
		expected: expectedMap,
		lock:     sync.Mutex{},
		missing:  len(expected),
	}
}

func (h *AliasPublisherHandler) OnPublish(sub *centrifuge.Subscription, e centrifuge.PublishEvent) {
	if h.AliasID != sub.Channel() {
		return
	}
	message := string(e.Data)
	h.lock.Lock()
	alreadyFound, found := h.expected[message]
	if !found {
		log.Fatalf("unexpected message received from alias %s: %v", h.AliasID, message)
	}
	if alreadyFound {
		log.Fatalf("duplicated message received from alias %s: %v", h.AliasID, message)
	}
	h.expected[message] = true
	h.missing--
	log.Printf("message received as expected via channel %s: %s", sub.Channel(), message)
	if h.missing == 0 {
		log.Printf("all expected messages have been received for alias: %v", h.AliasID)
	}

	h.lock.Unlock()
	h.WG.Done()
}
