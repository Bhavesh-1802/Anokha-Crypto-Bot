package ports

type TradingStrategyService interface {
	RunTradingStrategy(apiKey, secretKey string, signalChannel chan<- string)
	CheckUpdateOrder()
	IsPositionCompleted()
}
