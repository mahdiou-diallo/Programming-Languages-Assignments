package server

import (
	"../chatty"

	"fmt"
	"net"
	"sync"
	"encoding/gob"
)
type muxConn struct {
	chConn chatty.ChatConn
	mux sync.Mutex
}
// channel for safely passing the list of clients between go routines
var (
	list_chan chan map[string]*muxConn
	consMux sync.Mutex
)

func Start() {
	listen, port, err := chatty.OpenListener()
	fmt.Printf("Listening on port %d\n", port)

	if err != nil {
		fmt.Println(err)
		return
	}

	clientsList := make(map[string]*muxConn) 	//will be used to hold the list of currently connected clients
	list_chan  = make(chan map[string]*muxConn, 1)
	//put the list on the channel
	list_chan <- clientsList

	for {
		conn, err := listen.Accept() // this blocks until connection or error
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	chatConn := chatty.ChatConn{gob.NewEncoder(conn), gob.NewDecoder(conn), conn}
	clientConn := new(muxConn)
	clientConn.chConn = chatConn

	msg, err := chatty.RecvMsg(chatConn)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
	if msg.Action != chatty.CONNECT {
		msg.Action = chatty.ERROR
		msg.Body = "Protocol Error"
		clientConn.sendMsg(msg)
		return
	}
	//get clientsList from channel
	clientsList := <-list_chan
	//check that username is not in use
	if _, ok := clientsList[msg.Username]; ok {
		msg.Action = chatty.ERROR
		msg.Body = "Username " + msg.Username + " not available. Please choose a different one"
		clientConn.sendMsg(msg)
		list_chan <- clientsList // release clients list
		return
	}
	//add username
	clientsList[msg.Username] = clientConn
	msg.Body = "ok"
	//release clients list
	list_chan <- clientsList

	clientConn.sendMsg(msg)
	uname := msg.Username
	
	fmt.Printf("new connection accepted. user: %s\n", uname)
	
	//last action that the connection handler should do before exitting
	defer removeClient(uname)
	
	//receive messages from the client and do the task given by its purpose
	msg, err = chatty.RecvMsg(chatConn)
	for err == nil {
		switch msg.Action {
		case chatty.MSG:
			//get clientsList from channel
			clientsList := <-list_chan
			//find destination Encoder
			if destConn, ok := clientsList[msg.Username]; ok { // destination username is currently connected
				
				destConn.sendMsg(msg)
			} else { // destination username not currently connected
				msg.Action = chatty.ERROR
				msg.Body = msg.Username + " is not currently connected"
				msg.Username = ""
				clientConn.sendMsg(msg)
			}
			//release clientsList
			list_chan <- clientsList
			
		case chatty.LIST:
			msg.Body = "List of connected clients:"
			msg.Body += listClients(uname)
			clientConn.sendMsg(msg)
		case chatty.DISCONNECT:
			return
		default:
			//wrong purpose, protocol error
			//send error message
			msg = chatty.ChattyMsg{Action: chatty.ERROR, Body: "invalid purpose. Disconnecting."}
			clientConn.sendMsg(msg)
			//close connection
			clientConn.chConn.Conn.Close()
			//exit
			return
		}
		msg, err = chatty.RecvMsg(chatConn)
	}
}

func (destConn *muxConn) sendMsg (msg chatty.ChattyMsg) {
	//use mutex to ensure no race condition
	destConn.mux.Lock()
	//transmit message to destination
	chatty.SendMsg(destConn.chConn, msg)
	//release mutex
	destConn.mux.Unlock()
}

func muxPrint (str string) {
	consMux.Lock()
	fmt.Println(str)
	consMux.Unlock()
}

func listClients (uname string) string {
	//get clientsList from channel
	clientsList := <-list_chan
	//make list of clients
	list := ""
	for username, _ := range clientsList {
		if username != uname {
			list += "\n" + username
		}
	}
	if list == "" {
		list = " (None)"
	}
	//release clients list
	list_chan <- clientsList
	return list
}
func removeClient(uname string) {
	//get clientsList from channel
	clientsList := <-list_chan
	//find destination Encoder
	clientConn, _ := clientsList[uname]
	//remove client from list
	delete (clientsList, uname)
	//release clientsList
	list_chan <- clientsList

	//respond to client
	msg := chatty.ChattyMsg{Action: chatty.DISCONNECT}
	clientConn.sendMsg(msg)

	fmt.Println("\n" + uname + " disconnected.")
	//exit
}

