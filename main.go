package main

import (
	"context"
	"first-little-server/application"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelFunc()

	app := application.NewApp(ctx, application.LoadConfig())

	err := app.Start(ctx)
	if err != nil {
		fmt.Println("failed to start app:", err)
	}
}
