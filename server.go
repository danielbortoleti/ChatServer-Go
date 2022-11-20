package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type client chan<- string // canal de mensagem

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
	users    = make(map[string]client)
	pv_msg   = make(chan string)
)

func broadcaster() {
	clients := make(map[client]bool) // todos os clientes conectados

	for {
		select {
		case msg := <-messages:
			// broadcast de mensagens. Envio para todos
			for cli := range clients {
				cli <- msg
			}
			fmt.Println(clients)

		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		case msg := <-pv_msg:
			arg := strings.Split(msg, " ")
			fmt.Println(arg)

			cliente := users[arg[2]]
			fmt.Println(cliente)
			cliente <- msg

			for c, _ := range clients {
				fmt.Println(c)
				if users[arg[2]] == c {
					arg[3] = reverse(arg[3])
					users[arg[2]] <- " retornou: " + arg[3]
					break

				}
			}
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	apelido := conn.RemoteAddr().String()
	ch <- "vc é " + apelido
	messages <- apelido + " chegou!"
	entering <- ch
	users[apelido] = ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		entrada := strings.Split(input.Text(), " ")
		comandos := entrada[0]

		switch comandos {
		case ".nick":
			messages <- "Username alterado para " + entrada[1]
			apelido = entrada[1]
			users[apelido] = ch

		case ".sair":
			leaving <- ch
			messages <- apelido + " se foi "
			conn.Close()

		case ".pv":
			pv_msg <- apelido + " -> " + entrada[1] + " " + entrada[2]

		default:
			messages <- apelido + ":" + input.Text()
		}

	}
}

func reverse(s string) string {
	rns := []rune(s) // convert to rune
	for i, j := 0, len(rns)-1; i < j; i, j = i+1, j-1 {

		rns[i], rns[j] = rns[j], rns[i]
	}

	// return the reversed string.
	return string(rns)
}

func main() {
	fmt.Println("Iniciando servidor...")
	// Gerando uma conexão TCP
	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	for {
		// Permite conexõoes com o servidor criado
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		// Salva Conexão
		// connections = append(connections, conn)
		go handleConn(conn)
	}

}
