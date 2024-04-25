package api

import (
	adapters "Anokha-main/internal/adapters/TradingAdapter"
	websocket "Anokha-main/internal/adapters/websocketAdapter"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupAPI(server *gin.Engine) {
	server.LoadHTMLFiles("public/LongTrade.html", "public/ShortTrade.html", "public/Position.html", "public/info.html", "public/main.html")
	server.Static("/public", "./public")

	server.GET("/info", ShowGetFuturesAccountInfo)
	server.POST("/info", GetFuturesAccountInfo)

	server.GET("/long", ShowTradePageL)
	server.GET("/longtradews", StartTradeWebSocketL) //Ignore this is for Websocket conn

	server.GET("/short", ShowTradePageS)
	server.GET("/shorttradews", StartTradeWebSocketS) //Ignore this is for Websocket conn

	server.GET("/position", ShowPositionPage)
	server.GET("/positionws", StartPositionWebSocket)
}

func ShowGetFuturesAccountInfo(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "info.html", nil)
}

func GetFuturesAccountInfo(ctx *gin.Context) {
	apiKey := ctx.PostForm("apiKey")
	secretKey := ctx.PostForm("secretKey")
	adapter := adapters.TradingStrategyAdapter{}
	info := adapter.AccountInfo(apiKey, secretKey)
	//ctx.JSON(http.StatusOK, info)
	ctx.HTML(http.StatusOK, "main.html", gin.H{
		"AvailableBalance": info.AvailableBalance,
	})
}

func ShowTradePageL(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "LongTrade.html", nil)
}

func StartTradeWebSocketL(ctx *gin.Context) {
	ws := websocket.WebSockets{}
	ws.ServeWebSocketLT(ctx.Writer, ctx.Request)
}

func ShowTradePageS(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "ShortTrade.html", nil)
}

func StartTradeWebSocketS(ctx *gin.Context) {
	ws := websocket.WebSockets{}
	ws.ServeWebSocketST(ctx.Writer, ctx.Request)
}

func ShowPositionPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "Position.html", nil)
}

func StartPositionWebSocket(ctx *gin.Context) {
	ws := websocket.WebSockets{}
	ws.ServeWebSocketP(ctx.Writer, ctx.Request)
}
