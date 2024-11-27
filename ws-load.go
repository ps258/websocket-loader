package main

import (
  "crypto/tls"
  "flag"
  "fmt"
  "os"
  "log"
  "net/http"
  "sync"
  "time"

  "github.com/gorilla/websocket"
)

const (
  message   = "fred"
  writeWait = 10 * time.Second
)

var (
  numConnections int
  websocketURL   string
  printReplies   bool
)

func init() {
  flag.IntVar(&numConnections, "n", 0, "Number of concurrent connections")
  flag.StringVar(&websocketURL, "url", "", "WebSocket server URL")
  flag.BoolVar(&printReplies, "print", false, "Print replies from the server")
  flag.Parse()
  if websocketURL == "" || numConnections == 0 {
    fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
    flag.PrintDefaults()
    os.Exit(1)
  }
}

func main() {
  fmt.Printf("Connecting to %s with %d concurrent connections\n", websocketURL, numConnections)

  var wg sync.WaitGroup
  wg.Add(numConnections)

  for i := 0; i < numConnections; i++ {
    go func(id int) {
      defer wg.Done()
      connectWebSocket(id)
    }(i)
  }

  wg.Wait()
  fmt.Println("All connections completed")
}

func connectWebSocket(id int) {
  // Create a custom dialer that skips SSL verification
  dialer := websocket.Dialer{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
  }

  // Custom header to allow connection to servers that might require it
  header := http.Header{}
  header.Add("Origin", websocketURL)

  // Connect to the WebSocket server
  c, _, err := dialer.Dial(websocketURL, header)
  if err != nil {
    log.Printf("Connection %d failed: %v", id, err)
    return
  }
  defer c.Close()

  log.Printf("Connection %d established", id)

  // Send the message
  err = sendMessage(c, message)
  if err != nil {
    log.Printf("Connection %d send error: %v", id, err)
    return
  }
  log.Printf("Connection %d sent message: %s", id, message)

  // Read messages and discard them
  for {
    _, msg, err := c.ReadMessage()
    if err != nil {
      log.Printf("Connection %d read error: %v", id, err)
      break
    }
    // Discard the message, unless --print is given
    if printReplies {
      log.Printf("Connection %d received: %s", id, string(msg))
    }
  }
}

func sendMessage(c *websocket.Conn, message string) error {
  c.SetWriteDeadline(time.Now().Add(writeWait))
  return c.WriteMessage(websocket.TextMessage, []byte(message))
}
