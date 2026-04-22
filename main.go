package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/anand-aman/orders-api/application"
)

func main() {
	app := application.New(application.LoadConfig())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	err := app.Start(ctx)

	if err != nil {
		fmt.Println("Failed to start the app: ", err)
	}
}
