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

	mu.Lock()
	clients[ws] = true
	mu.Unlock()

	log.Printf("Client connected: %v", ws.RemoteAddr())

	go func() {
		defer func() {
			mu.Lock()
			delete(clients, ws)
			mu.Unlock()
			log.Printf("Client disconnected: %v", ws.RemoteAddr())
		}()
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

func monitorLogFile(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file.Seek(0, os.SEEK_END)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					readNewLines(file)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

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
		line := scanner.Bytes()
		sendToClients(line)
	}
}

func sendToClients(message []byte) {
	mu.Lock()
	defer mu.Unlock()
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Write error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}
