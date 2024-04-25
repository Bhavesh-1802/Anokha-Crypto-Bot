package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/markcheno/go-talib"
)

var CurrentEntryPrice string
var CurrentPositionPrice string
var CurrentPositionAmount string
var (
	apiKey    = "d38ff8ff03f3cb1de0693a83fedfa520b8d9a76f6e0f6456eccaeab327618e9f"
	secretKey = "6f06f82ebbfd7a4d2a544cef65f0f558ca348f88718f5d158af2cad22f42346c"
	symbol    = "BTCUSDT"
)

type TradingShortService struct{}

func (tss *TradingShortService) ShortFutureTrade(apiKey, secretKey string) string {
	// Initialize Binance Futures client
	client := futures.NewClient(apiKey, secretKey)

	symbol := symbol
	interval := "4h"
	limit := 500

	// Fetch historical candlestick data
	klines, err := client.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		Limit(limit).
		Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	var closes []float64
	for _, kline := range klines {
		closePrice, _ := strconv.ParseFloat(kline.Close, 64) // Convert string to float64
		closes = append(closes, closePrice)
	}

	// Calculate indicators
	rsiPeriod := 14
	bollingerPeriod := 20
	bollingerStdDev := 2.0
	emaPeriod := 21

	rsi := talib.Rsi(closes, rsiPeriod)
	bbUpper, _, _ := talib.BBands(closes, bollingerPeriod, bollingerStdDev, bollingerStdDev, talib.SMA)
	ema := talib.Ema(closes, emaPeriod)

	// Check buy condition
	lastIndex := len(closes) - 1

	if rsi[lastIndex] > 70 && closes[lastIndex] > ema[lastIndex] && closes[lastIndex] > bbUpper[lastIndex] {
		client := futures.NewClient(apiKey, secretKey)
		client.BaseURL = "https://testnet.binancefuture.com"

		symbol := symbol
		quantity := "0.01" // Quantity of the asset to buy

		leverage := 5 // Your desired leverage level

		leverageService := client.NewChangeLeverageService()

		// Set the symbol and leverage
		leverageService = leverageService.Symbol(symbol).Leverage(leverage)

		// Send the request to change leverage
		_, err := leverageService.Do(context.Background())
		if err != nil {
			// Handle error
			panic(err)
		}

		// Place the buy order
		orderResp, err := client.NewCreateOrderService().
			Symbol(symbol).
			Side(futures.SideTypeSell).
			Type(futures.OrderTypeMarket).
			Quantity(quantity).
			Do(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		CurrentEntryPrice = orderResp.Price
		CurrentPositionPricef, _ := strconv.ParseFloat(CurrentPositionPrice, 64)

		// Calculate target price for take-profit
		tpPricef := CurrentPositionPricef * 0.98
		tpPrice := fmt.Sprintf("%.2f", tpPricef)

		// Calculate stop-loss price (adjust as needed based on your strategy)
		slPricef := CurrentPositionPricef * 1.01
		slPrice := fmt.Sprintf("%.2f", slPricef)

		// Place the take-profit order
		TPorder, err := client.NewCreateOrderService().
			Symbol(symbol).
			Side(futures.SideTypeBuy).
			Type(futures.OrderTypeLimit).
			TimeInForce(futures.TimeInForceTypeGTC).
			Quantity(quantity).
			Price(tpPrice).
			Do(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		// Place the stop-loss order
		SLorder, err := client.NewCreateOrderService().
			Symbol(symbol).
			Side(futures.SideTypeBuy).
			Type(futures.OrderTypeStopMarket).
			Quantity(quantity).
			StopPrice(slPrice).
			Do(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Second * 5) // Pause to give orders some time to get processed

		TpPrice := TPorder.Price
		SlPrice := SLorder.Price

		return fmt.Sprintf("Order ID: %d\n AtPrice--> %s\n TakeProfit --> %s\n StopLoss --> %s\n", orderResp.OrderID, orderResp.Price, TpPrice, SlPrice)
	} else {
		result := "No Buy Signal...Checking Again"
		return result
	}
}

func (tss *TradingShortService) CheckUpdateOrder() {
	client := futures.NewClient(apiKey, secretKey)

	// Set the testnet base URL
	client.BaseURL = "https://testnet.binancefuture.com"

	// Fetch positions
	positions, err := client.NewGetPositionRiskService().Symbol(symbol).Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, position := range positions {

		CurrentPositionPrice = position.EntryPrice
		CurrentPositionAmount = position.PositionAmt

		if position.EntryPrice != CurrentEntryPrice {
			// Fetch open orders
			openOrders, err := client.NewListOpenOrdersService().Do(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			// Cancel open orders
			for _, order := range openOrders {
				_, err := client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
				if err != nil {
					log.Fatal(err)
				}

				quantity := CurrentPositionAmount

				CurrentPositionPricef, _ := strconv.ParseFloat(CurrentPositionPrice, 64)
				// Calculate target price for take-profit
				tpPricef := CurrentPositionPricef * 0.98
				tpPrice := fmt.Sprintf("%.2f", tpPricef)

				// Calculate stop-loss price (adjust as needed based on your strategy)
				slPricef := CurrentPositionPricef * 1.01
				slPrice := fmt.Sprintf("%.2f", slPricef)

				// Place the New take-profit order
				_, err = client.NewCreateOrderService().
					Symbol(symbol).
					Side(futures.SideTypeBuy).
					Type(futures.OrderTypeLimit).
					TimeInForce(futures.TimeInForceTypeGTC).
					Quantity(quantity).
					Price(tpPrice).
					Do(context.Background())
				if err != nil {
					log.Fatal(err)
				}

				// Place the New stop-loss order
				_, err = client.NewCreateOrderService().
					Symbol(symbol).
					Side(futures.SideTypeBuy).
					Type(futures.OrderTypeStopMarket).
					Quantity(quantity).
					StopPrice(slPrice).
					Do(context.Background())
				if err != nil {
					log.Fatal(err)
				}

			}

		}

	}

}

func (tss *TradingShortService) IsPositionCompleted() {
	client := futures.NewClient(apiKey, secretKey)

	// Set the testnet base URL
	client.BaseURL = "https://testnet.binancefuture.com"

	// Fetch positions
	positions, err := client.NewGetPositionRiskService().Symbol(symbol).Do(context.Background())
	if err != nil {
		fmt.Println("Error fetching positions:", err)
		return
	}

	for _, position := range positions {

		if position.EntryPrice == "0.0" {

			// Fetch open orders
			openOrders, err := client.NewListOpenOrdersService().Do(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			// Cancel open orders
			for _, order := range openOrders {
				_, err := client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
