# [logtosocket]

## Description
LogToSocket is a Go application that monitors changes in a log file and sends newly added lines to connected WebSocket clients in real time. It uses the `fsnotify` library to watch for file changes and the `gorilla/websocket` library to handle WebSocket connections.

## Prerequisites
- Go (version 1.16 or later)

## Installation

### Install Go
1. Download and install Go from the official website: [https://golang.org/dl/](https://golang.org/dl/)
2. Follow the installation instructions for your operating system: [https://golang.org/doc/install](https://golang.org/doc/install)

### Install Required Libraries
Open a terminal and run the following commands to install the necessary Go packages:

```sh
go get -u github.com/fsnotify/fsnotify
go get -u github.com/gorilla/websocket
```

## Usage
Compile the application using the following command:

```sh
go build -o logtosocket.exe main.go
```

## Running the Application

Run the compiled application with the desired arguments:


```sh
./logtosocket.exe -logfile="path_to_logfile" -endpoint="/ws/log" -address="127.0.0.1" -port=8080
```

Command Line Arguments

    -logfile: Path to the log file to monitor (required)
    -endpoint: WebSocket endpoint (default: /ws/log)
    -address: Address to run the server on (default: 0.0.0.0)
    -port: Port to run the server on (default: 8080)
