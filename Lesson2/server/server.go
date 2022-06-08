package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

type client chan<- string

var (
	entering   = make(chan client)
	leaving    = make(chan client)
	messages   = make(chan string)
	names      = map[string]string{}
	gameStart  = false
	gameResult = 0
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("server started at localhost:8001")

	go broadcaster()
	go sendTime(&messages, 60) // Запускаем функцию вывода времени.
	go playGame(&messages)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)

			continue
		}
		go handleConn(conn)
		go sendAdminMessage() // Запускаем обработчик консоли для отправки сообщений от Админа
	}
}

func sendAdminMessage() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		messages <- "Admin console message: " + scanner.Text()
	}
}

// sendTime - функция рассылает всем текущее время если игра запущено то рассылка прекращается.
func sendTime(chn *chan string, tim int64) {
	for {
		if !gameStart { // если игра не идет, то можно выводить время
			time.Sleep(time.Duration(tim) * time.Second)
			tm := time.Now().String()
			*chn <- tm
		}
	}
}

func playGame(chi *chan string) {
	for {
		if !gameStart { // Если нет активной игры, то стартуем
			time.Sleep(348 * time.Second)
			rand.Seed(time.Now().UnixNano())
			gameStart = true
			task := rand.Intn(3)
			operand1 := rand.Intn(100)
			operand2 := rand.Intn(100)
			switch task { // рандомно значение операции
			case 0:
				gameResult = operand1 + operand2 // Сложение
				operation := fmt.Sprintf("Play game?\nCount the answer: %d + %d = ?", operand1, operand2)
				*chi <- operation
			case 1:
				gameResult = operand1 - operand2 // Вычитание
				operation := fmt.Sprintf("Play game?\nCount the answer: %d - %d = ?", operand1, operand2)
				*chi <- operation
			case 2:
				gameResult = operand1 * operand2 // Умножение
				operation := fmt.Sprintf("Play game?\nCount the answer: %d * %d = ?", operand1, operand2)
				*chi <- operation
			case 3:
				for { // проверка деления на 0
					if operand2 == 0 {
						operand2 = rand.Intn(100)
					}

					break
				}
				gameResult = operand1 / operand2 // деление
				operation := fmt.Sprintf("Play game?\nCount the answer:%d / %d = ?", operand1, operand2)
				*chi <- operation
			}
		}
	}
}

func handleConn(conn net.Conn) {
	var cliName string

	ch := make(chan string)
	go clientWriter(conn, ch)

	input := bufio.NewScanner(conn)
	clAddr := conn.RemoteAddr().String()

	_, ok := names[clAddr]
	if !ok {
		input.Scan()

		names[clAddr] = input.Text()
	}

	cliName = names[clAddr]

	ch <- "You nickname " + cliName
	messages <- cliName + " has arrived"
	entering <- ch
	for input.Scan() {
		messages <- cliName + ": " + input.Text()
		if gameStart {
			cliResult, err := strconv.Atoi(input.Text())
			if err != nil {
				continue
			}
			if cliResult == gameResult {
				resultMessage := "The game is over! The winner is " + cliName
				messages <- resultMessage
				gameStart = false
			}
		}
	}
	leaving <- ch
	messages <- cliName + " has left"

	err := conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		_, err := fmt.Fprintln(conn, msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func broadcaster() {
	clients := make(map[client]bool)

	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}

		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}
