package main

import (
	"github.com/genefriendway/onchain-handler/conf"
	_ "github.com/genefriendway/onchain-handler/docs"
	app "github.com/genefriendway/onchain-handler/internal"
)

func main() {
	config := conf.GetConfiguration()
	app.RunApp(config)
}
