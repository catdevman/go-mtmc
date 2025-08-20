package main

import (
	"github.com/catdevman/go-mtmc/internal/emulator"
	"github.com/catdevman/go-mtmc/internal/web"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Create a new instance of the MTMC computer.
	computer := emulator.New()

	// Start the web server, which provides the user interface.
	server := web.NewServer(computer)
	go server.Start()

	// Start the computer's execution cycle in a separate goroutine.
	go computer.Run()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
}
