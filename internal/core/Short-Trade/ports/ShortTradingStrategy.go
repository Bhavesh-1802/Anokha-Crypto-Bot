package ports

type ShortTradingStrategy interface {
	RunShortTradingStrategy(apiKey, secretKey string, ShortSignalChannel chan<- string)
	CheckUpdateOrder()
	IsPositionCompleted()
}
