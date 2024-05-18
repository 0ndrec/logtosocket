package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var (
	logFilePath string
	endpoint    string
	address     string
	port        int
	upgrader    = websocket.Upgrader{}
	clients     = make(map[*websocket.Conn]bool)
	mu          sync.Mutex
)

func main() {
	flag.StringVar(&logFilePath, "logfile", "", "Path to the log file to monitor")
	flag.StringVar(&endpoint, "endpoint", "/ws/log", "WebSocket endpoint")
	flag.StringVar(&address, "address", "0.0.0.0", "Address to run the server on")
	flag.IntVar(&port, "port", 8080, "Port to run the server on")
	flag.Parse()

	if logFilePath == "" {
		log.Fatal("logfile is required")
	}

	http.HandleFunc(endpoint, handleConnections)
	go monitorLogFile(logFilePath)

	addr := fmt.Sprintf("%s:%d", address, port)
	log.Printf("Server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}
	defer ws.Close()

	mu.Lock()
	clients[ws] = true
	mu.Unlock()

	log.Printf("Client connected: %v", ws.RemoteAddr())

	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			mu.Lock()
			delete(clients, ws)
			mu.Unlock()
			log.Printf("Client disconnected: %v", ws.RemoteAddr())
			break
		}
	}
}

// MonitorLogFile watches a log file for new lines and sends them to all connected clients
func monitorLogFile(path string) {
	// Create a file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Open the log file and move the file pointer to the end of the file
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file.Seek(0, os.SEEK_END)

	// Start a goroutine to watch the file for changes
	go func() {
		// Loop forever and watch for file changes
		for {
			select {
			case event, ok := <-watcher.Events:
				// If the event is a write, read new lines from the file
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					readNewLines(file)
				}
			case err, ok := <-watcher.Errors:
				// If there's an error, log it and return
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

	// Add the file to the watcher
	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}

	// Block forever
	select {}
}

func readNewLines(file *os.File) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		sendToClients(line)
	}
}

func sendToClients(message string) {
	mu.Lock()
	defer mu.Unlock()
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("Write error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}
