package main

import (
	"context"
	"first-little-server/application"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	app := application.NewApp()

	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelFunc()

	err := app.Start(ctx)
	if err != nil {
		fmt.Println("failed to start app:", err)
	}
}
