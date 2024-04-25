package services

import (
	trading "Anokha-main/internal/core/Long-Trade"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/markcheno/go-talib"
)

var (
	apiKey    = "d38ff8ff03f3cb1de0693a83fedfa520b8d9a76f6e0f6456eccaeab327618e9f"
	secretKey = "6f06f82ebbfd7a4d2a544cef65f0f558ca348f88718f5d158af2cad22f42346c"
	symbol    = "BTCUSDT"
)
var (
	PSymbol           string
	PLiquidationPrice string
	PEntryPrice       string
	PPositionAmt      string
	PUnRealizedProfit string
	PLeverage         string
)

var CurrentEntryPrice string
var CurrentPositionPrice string
var CurrentPositionAmount string

type TradingService struct{}

func (ts *TradingService) FutureTrade(apiKey, secretKey string) string {
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
	_, _, bbLower := talib.BBands(closes, bollingerPeriod, bollingerStdDev, bollingerStdDev, talib.SMA)
	ema := talib.Ema(closes, emaPeriod)

	// Check buy condition
	lastIndex := len(closes) - 1

	if rsi[lastIndex] > 30 && closes[lastIndex] < bbLower[lastIndex] && closes[lastIndex] > ema[lastIndex] {
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
			Side(futures.SideTypeBuy).
			Type(futures.OrderTypeMarket).
			Quantity(quantity).
			Do(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		CurrentEntryPrice = orderResp.Price
		CurrentPositionPricef, _ := strconv.ParseFloat(CurrentPositionPrice, 64)

		// Calculate target price for take-profit
		tpPricef := CurrentPositionPricef * 1.01
		tpPrice := fmt.Sprintf("%.2f", tpPricef)

		// Calculate stop-loss price (adjust as needed based on your strategy)
		slPricef := CurrentPositionPricef * 0.98
		slPrice := fmt.Sprintf("%.2f", slPricef)

		// Place the take-profit order
		TPorder, err := client.NewCreateOrderService().
			Symbol(symbol).
			Side(futures.SideTypeSell).
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
			Side(futures.SideTypeSell).
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

func (ts *TradingService) GetFuturesAccountInfo(apiKey, secretKey string) *trading.Account {
	// Create a new futures client
	client := futures.NewClient(apiKey, secretKey)
	client.BaseURL = "https://testnet.binancefuture.com"

	// Fetch account information
	accountInfo, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil
	}

	// Convert accountInfo to AccountInfo structure
	accountJSON, err := json.Marshal(accountInfo)
	if err != nil {
		return nil
	}

	var account trading.Account
	err = json.Unmarshal(accountJSON, &account)
	if err != nil {
		return nil
	}

	return &account
}

func (ts *TradingService) LivePositions() string {
	client := futures.NewClient(apiKey, secretKey)

	// Set the testnet base URL
	client.BaseURL = "https://testnet.binancefuture.com"

	// Fetch positions
	positions, err := client.NewGetPositionRiskService().Symbol(symbol).Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, position := range positions {

		PSymbol = position.Symbol
		PLiquidationPrice = position.LiquidationPrice
		PEntryPrice = position.EntryPrice
		PPositionAmt = position.PositionAmt
		PUnRealizedProfit = position.UnRealizedProfit
		PLeverage = position.Leverage

	}
	return fmt.Sprintf("Symbol: %s\n  Entry Price: %s\n Position Amount: %s\n Leverage: %s\n Liquidation Price: %s\n P/L: %s\n",
		PSymbol, PEntryPrice, PPositionAmt, PLeverage, PLiquidationPrice, PUnRealizedProfit)

}

func (ts *TradingService) CheckUpdateOrder() {
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
				tpPricef := CurrentPositionPricef * 1.01
				tpPrice := fmt.Sprintf("%.2f", tpPricef)

				// Calculate stop-loss price (adjust as needed based on your strategy)
				slPricef := CurrentPositionPricef * 0.98
				slPrice := fmt.Sprintf("%.2f", slPricef)

				// Place the New take-profit order
				_, err = client.NewCreateOrderService().
					Symbol(symbol).
					Side(futures.SideTypeSell).
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
					Side(futures.SideTypeSell).
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

func (ts *TradingService) IsPositionCompleted() {
	client := futures.NewClient(apiKey, secretKey)

	client.BaseURL = "https://testnet.binancefuture.com"
	positions, err := client.NewGetPositionRiskService().Symbol(symbol).Do(context.Background())
	if err != nil {
		fmt.Println("Error While fetching Positions : ", err)
		return
	}
	for _, position := range positions {
		if position.EntryPrice == "0.0" {

			openOrders, err := client.NewListOpenOrdersService().Do(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			for _, order := range openOrders {
				_, err := client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
