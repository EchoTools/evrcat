package cat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

type EVRCatServer struct {
	pool sync.Pool
}

func NewEVRCatServer(hashmap map[string]string) *EVRCatServer {
	replacements := make([]string, 0, len(hashmap)*2)
	for k, v := range hashmap {
		replacements = append(replacements, k, v)
	}

	return &EVRCatServer{
		pool: sync.Pool{
			New: func() interface{} {
				return strings.NewReplacer(replacements...)
			},
		},
	}
}

func (cs *EVRCatServer) Start(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Fprintln(os.Stderr, "Server started. Listening on port", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go cs.handleConnection(conn)
	}
}

func (cs *EVRCatServer) handleConnection(conn net.Conn) {
	fmt.Fprintln(os.Stderr, "Accepted connection from", conn.RemoteAddr())
	defer conn.Close()

	replacer := cs.pool.Get().(*strings.Replacer)
	defer cs.pool.Put(replacer)

	readBytes := 0
	sentBytes := 0

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		readBytes += len(line)
		// Process the line using your custom function
		line = replacer.Replace(line)
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading from connection:", err)
		}
		// Send the processed line back to the client
		n, err := fmt.Fprintln(conn, line)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing to connection:", err)
			break
		}
		sentBytes += n
	}
	fmt.Fprintf(os.Stderr, "Connection from %s closed. Read %d bytes, sent %d bytes\n", conn.RemoteAddr(), readBytes, sentBytes)
}
