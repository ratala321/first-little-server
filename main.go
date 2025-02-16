package main

import (
	"context"
	"first-little-server/application"
	"fmt"
)

func main() {
	app := application.NewApp()

	err := app.Start(context.TODO())

	if err != nil {
		fmt.Println("failed to start app:", err)
	}
}
