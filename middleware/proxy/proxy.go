package main

import (
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/noxproject/nox/middleware/handler"
)

func transmit() {
	app := iris.New()
	app.Logger().SetLevel("debug")
	app.Use(recover.New())
	app.Use(logger.New())
	//允许跨域
	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	app.Use(crs)

	app.Get("/", func(ctx iris.Context) {
		ctx.HTML("<h1>Empty</h1>")
	})

	app.Post("/block/info", func(ctx iris.Context) {
		blockHash := ctx.PostValue("hash")
		blockInfo := handler.GetBlockInfo(blockHash)
		ctx.JSON(blockInfo)
	})

	app.Post("transaction", func(ctx iris.Context) {
		transactionHashStr := ctx.PostValue("transaction")
		transactionResult := handler.GetTransaction(transactionHashStr)
		if transactionResult != nil {
			ctx.JSON(transactionResult)
		} else {
			ctx.Writef("没有该交易")
		}
	})

	app.Post("balance", func(ctx iris.Context) {
		address := ctx.PostValue("address")
		balance := handler.GetAddressInfo(address)
		ctx.JSON(balance)

	})

	app.Run(iris.Addr(":9998"), iris.WithoutServerError(iris.ErrServerClosed))
}

func main() {
	transmit()
}
