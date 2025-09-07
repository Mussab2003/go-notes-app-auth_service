package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/Mussab2003/go-notes-app-auth_service.git/application"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	app := application.New(ctx)
	
	if err := app.Start(ctx); err != nil {
		fmt.Println("Failed to start application:", err)
		os.Exit(1)
	}

}