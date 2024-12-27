package main

import (
	app "github.com/genefriendway/onchain-handler/cmd/app"
	"github.com/genefriendway/onchain-handler/conf"
	_ "github.com/genefriendway/onchain-handler/docs"
)

func main() {
	config := conf.GetConfiguration()
	app.RunApp(config)
}
