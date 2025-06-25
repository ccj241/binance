package services

import (
	"context"
	"github.com/adshao/go-binance/v2"
)

type BinanceService struct {
	Client *binance.Client
}

func NewBinanceService(apiKey, secretKey string) *BinanceService {
	client := binance.NewClient(apiKey, secretKey)
	return &BinanceService{Client: client}
}

func (s *BinanceService) GetBalance() ([]*binance.Balance, error) {
	account, err := s.Client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	return account.Balances, nil
}
