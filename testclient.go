package testclient

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/centrifugal/centrifuge-go"

	"github.com/Unity-Technologies/mp-prism/testclient/handler"
	"github.com/Unity-Technologies/mp-prism/testclient/model"
)

func NewTestClient() {

}

func New() {
	cfg := model.LoadConfig("./cmd/testclient/tests.yaml")

	for _, test := range cfg.Tests {
		doneChan := make(chan bool)
		go runTest(test, cfg, doneChan)
		select {
		case <-doneChan:
			continue
		case <-time.Tick(cfg.TestTimeout):
			log.Fatalf("test failed: timeout")
		}
	}
}

func runTest(test model.Test, cfg *model.Config, done chan bool) {

	log.Printf("running test: %s", test.Name)

	testWaitGroup := getTestWaitGroup(test)

	clients, subs := prepareClient(test.Expected, cfg, testWaitGroup)

	sendRequests(test.Requests, cfg.TestServiceURL)

	testWaitGroup.Wait()

	log.Printf("%s: passed\n", test.Name)

	// clean up
	for _, sub := range subs {
		_ = sub.Close()
	}
	for _, client := range clients {
		_ = client.Close()
	}
	done <- true
}

func getTestWaitGroup(test model.Test) *sync.WaitGroup {
	testWaitGroup := &sync.WaitGroup{}
	expectMsgCount := 0
	for _, user := range test.Expected {
		for _, alias := range user.Aliases {
			expectMsgCount += len(alias.Messages)
		}
	}
	log.Printf("expecte %d messages for user %s", expectMsgCount, test.Name)
	testWaitGroup.Add(expectMsgCount)
	return testWaitGroup
}

func newClient(wsURL string) (*centrifuge.Client, *handler.DefaultHandler) {
	h := &handler.DefaultHandler{}
	c := centrifuge.New(wsURL, centrifuge.DefaultConfig())
	c.OnConnect(h)
	c.OnDisconnect(h)
	c.OnMessage(h)
	c.OnError(h)

	c.OnServerSubscribe(h)
	c.OnServerUnsubscribe(h)

	return c, h
}

func prepareClient(users []model.User, cfg *model.Config, wg *sync.WaitGroup) (clients []*centrifuge.Client, subs []*centrifuge.Subscription) {
	for _, user := range users {

		client, defaultHandler := newClient(cfg.DoormanURL)
		token := getToken(cfg)
		client.SetToken(token)
		clients = append(clients, client)

		for _, alias := range user.Aliases {
			sub := newSubscription(client, alias, defaultHandler, wg, alias.Messages)

			if err := sub.Subscribe(); err != nil {
				log.Fatalln(err)
			}

			subs = append(subs, sub)
		}
		if err := client.Connect(); err != nil {
			log.Fatalln(err)
		}
	}
	return
}

func getToken(cfg *model.Config) string {
	client := http.Client{}
	req, err := http.NewRequest("POST", cfg.TokenURL, nil)
	if err != nil {
		log.Fatalf("cannot create POST request for token: %v", err)
	}
	req.Header.Set(cfg.ProjectIDHeader, cfg.ProjectID)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error sending request for token: %v", err)
	}
	bToken, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %v", err)
	}
	token := &model.Token{}
	if err = json.Unmarshal(bToken, token); err != nil {
		log.Fatalf("error unmarshal response body: %v", err)
	}

	return token.IDToken
}

func newSubscription(client *centrifuge.Client, alias model.ResponseFromAlias, defaultHandler *handler.DefaultHandler, wg *sync.WaitGroup, messages []string) *centrifuge.Subscription {
	sub, err := client.NewSubscription(alias.ID)
	if err != nil {
		log.Fatalln(err)
	}
	aliasPublishHandler := handler.NewAliasPublisherHandler(wg, alias.ID, messages)
	sub.OnPublish(aliasPublishHandler)
	sub.OnSubscribeSuccess(defaultHandler)
	sub.OnSubscribeError(defaultHandler)
	sub.OnUnsubscribe(defaultHandler)

	return sub
}
func sendRequests(requests []model.Request, tsURL string) {
	client := http.Client{}
	for _, req := range requests {
		r, err := http.NewRequest("POST", tsURL, bytes.NewBuffer([]byte(req.Body)))
		if err != nil {
			log.Fatalf("cannot create request: %v", err)
		}
		resp, err := client.Do(r)
		if err != nil {
			log.Fatalf("error sending request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("response status: %v", resp.StatusCode)
		}
		_ = resp.Body.Close()
	}
}
