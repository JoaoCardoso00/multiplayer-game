package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const PORT = ":8080"

var (
	currentNumber int
	gameMutex     sync.Mutex
	connections   map[net.Conn]bool
)

var seed = rand.NewSource(time.Now().UnixNano())
var rng = rand.New(seed)

func main() {
	listener, err := net.Listen("tcp", PORT)
	connections = make(map[net.Conn]bool)

	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}

	defer listener.Close()

	fmt.Println("Listening on port", PORT)

	startNewRound()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		gameMutex.Lock()
		connections[conn] = true
		gameMutex.Unlock()

		go handleRequest(conn)
	}

}

func handleRequest(conn net.Conn) {
	fmt.Println("Recieved connection from: ", conn.RemoteAddr().String())

	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading guess:", err.Error())
			return
		}

		message = strings.TrimSpace(message)
		guess, err := strconv.Atoi(message)

		if err != nil {
			fmt.Println("Error converting guess:", err.Error())
			sendString(conn, "Please send a valid number.\n\n")
			continue
		}

		fmt.Println("Guess from ", conn.LocalAddr(), ": ", guess)

		if guess < currentNumber {
			message = "guess too low, try again\n"
		}

		if guess > currentNumber {
			message = "guess too high, try again\n"
		}

		if guess != currentNumber {
			sendString(conn, message)
			continue
		}

		broadcastMessage("Player " + conn.RemoteAddr().String() + " guessed correctly. Starting a new round!\n")
		startNewRound()

		continue

	}

}

func startNewRound() {
	gameMutex.Lock()
	defer gameMutex.Unlock()

	currentNumber = rng.Intn(100) + 1
	fmt.Println("New round started, number is:", currentNumber)
}

func sendString(conn net.Conn, message string) {
	data := []byte(message)

	_, err := conn.Write(data)
	if err != nil {
		fmt.Println("Error writing to connection:", err.Error())
	}
}

func broadcastMessage(message string) {
	gameMutex.Lock()
	defer gameMutex.Unlock()

	for conn := range connections {
		sendString(conn, message)
	}
}
