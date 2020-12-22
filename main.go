package main

import (
	"log"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/ququzone/geth-service/service"
	"github.com/ququzone/geth-service/web"
)

func main() {
	hs, err := service.GetHeaderService()
	if err != nil {
		log.Fatalf("get header service error: %v\n", err)
	}
	hs.AddSubscriber(web.NewWebsocketPool())

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/ws", web.Websocket)

	// Start server
	e.Logger.Fatal(e.Start(":8081"))
}
