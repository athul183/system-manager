package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sysagent/internal/api"
	"sysagent/internal/auth"
	"sysagent/internal/collector"
	"sysagent/internal/controller"
	"sysagent/internal/discovery"
	"sysagent/internal/models"
	"sysagent/internal/storage"
)

func main() {
	log.Println("Starting sysagent...")

	// 1. Initialize DB / Storage
	store, err := storage.NewStore("metrics.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 2. Generate a token for test printing
	token, _ := auth.GenerateToken()
	log.Println("=== Use this JWT token to connect from the Flutter app ===")
	log.Println(token)
	log.Println("==========================================================")

	// 3. Discover Applications
	disco := discovery.NewDiscoverer()
	apps := disco.Discover()
	log.Printf("Discovered %d applications to monitor\n", len(apps))

	// 4. Start Metrics Collection
	metricsChan := make(chan models.AppMetrics, 100)
	sampler := collector.NewSampler(apps, metricsChan)
	go sampler.Start()

	// 5. Start WebSocket & Control Loop
	commandChan := make(chan models.ControlCommand, 50)
	wsServer := api.NewWSServer(commandChan)

	// Start reading and executing commands from UI
	controller.StartCommandListener(commandChan)

	go func() {
		log.Println("WebSocket server listening on localhost:8080")
		http.Handle("/ws", wsServer)
		if err := http.ListenAndServe("127.0.0.1:8080", nil); err != nil {
			log.Fatalf("WebSocket listener failed: %v", err)
		}
	}()

	// 6. Router loop to handle DB saving and UI broadcasting
	go func() {
		batch := make([]models.AppMetrics, 0, 100)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case m := <-metricsChan:
				batch = append(batch, m)

			case <-ticker.C:
				// Push to DB & UI every 2 seconds
				if len(batch) > 0 {
					go wsServer.BroadcastMetrics(batch)
					if err := store.BatchInsert(batch); err != nil {
						log.Println("DB Insert error:", err)
					}
					batch = make([]models.AppMetrics, 0, 100)
				}
			}
		}
	}()

	// Wait for interrupt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down sysagent...")
	sampler.Stop()
}
