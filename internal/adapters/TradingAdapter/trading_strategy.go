package adapters

import (
	trading "Anokha-main/internal/core/Long-Trade"
	"Anokha-main/internal/core/Long-Trade/services"
	shortservices "Anokha-main/internal/core/Short-Trade/services"
	"time"
)

type TradingStrategyAdapter struct{}

func (tsa *TradingStrategyAdapter) RunTradingStrategy(apiKey, secretKey string, signalChannel chan<- string) {
	tradingService := services.TradingService{}

	for {
		result := tradingService.FutureTrade(apiKey, secretKey)
		signalChannel <- result
		time.Sleep(time.Second * 10)
	}
}

func (tsa *TradingStrategyAdapter) AccountInfo(apiKey, secretKey string) *trading.Account {
	tradingService := services.TradingService{}

	info := tradingService.GetFuturesAccountInfo(apiKey, secretKey)

	return info
}

func (tsa *TradingStrategyAdapter) RunLivePosition(PositionChannel chan<- string) {
	tradingService := services.TradingService{}
	for {
		result := tradingService.LivePositions()
		PositionChannel <- result
		time.Sleep(time.Millisecond * 10)
	}
}

func (tsa *TradingStrategyAdapter) RunShortTradingStrategy(apiKey, secretKey string, ShortSignalChannel chan<- string) {
	tradingService := shortservices.TradingShortService{}

	for {
		result := tradingService.ShortFutureTrade(apiKey, secretKey)
		ShortSignalChannel <- result
		time.Sleep(time.Second * 10)
	}
}
