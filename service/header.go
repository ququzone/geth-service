package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ququzone/go-common/env"
)

var headerService *HeaderService

type Subscriber interface {
	Receive(message string) error
}

type HeaderService struct {
	client      *ethclient.Client
	subscribers []Subscriber

	Number uint64
}

func (s *HeaderService) Json() string {
	return fmt.Sprintf(`{"header": {"number": %d}}`, s.Number)
}

func (s *HeaderService) AddSubscriber(sub Subscriber) {
	s.subscribers = append(s.subscribers, sub)
}

func GetHeaderService() (*HeaderService, error) {
	if headerService == nil {
		client, err := ethclient.Dial(env.GetNonEmpty("INFURA_WS_ENDPOINT"))
		if err != nil {
			return nil, err
		}

		current, err := client.BlockNumber(context.Background())
		if err != nil {
			return nil, err
		}

		header := make(chan *types.Header)
		if _, err := client.SubscribeNewHead(context.Background(), header); err != nil {
			return nil, err
		}

		headerService = &HeaderService{
			client: client,
			Number: current,
		}

		go func() {
			for h := range header {
				headerService.Number = h.Number.Uint64()
				for _, sub := range headerService.subscribers {
					if err := sub.Receive(headerService.Json()); err != nil {
						log.Printf("push message error: %v\n", err)
					}
				}
			}
		}()
	}

	return headerService, nil
}
