# logtosocket
LogToSocket is a Go application that monitors changes in a log file and sends newly added lines to connected WebSocket clients in real time. It uses the `fsnotify` library to watch for file changes and the `gorilla/websocket` library to handle WebSocket connections.
