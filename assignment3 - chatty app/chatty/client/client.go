package client

import (
	"../chatty"

	"fmt"
	"os"
	"bufio"
	"sync"
)
var (
	consMux sync.Mutex // restricts printing to the console
	wg sync.WaitGroup // synchronizes functions so that some exit only after all the workers are done
)
// constants to define user actions
const (
	SEND_MSG string = "!s"
	LIST = "!l"
	QUIT = "!q"
	CANCEL = "!c"
	HELP = "!h"
)
/* main function of the program
	inputs:
		user: the provided username
		serverPort: the port the server is listening on
		serverAddr: the IP address of the server
*/
func Start(user string, serverPort string, serverAddr string) {
	// check if username was provided
	if user == "" {
		fmt.Println("Please provide a non empty username")
		return
	}
	// Connect to chat server
	chConn, err := chatty.ServerConnect(user, serverAddr, serverPort)
	if err != nil {
		fmt.Printf("unable to connect to server: %s\n", err)
		return
	}
	// receive response from server
	msg, err := chatty.RecvMsg(chConn)
	// check for failed connection
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	// TCP connection was established
	fmt.Println("connected to server")
	
	// server did not accept the connection (username already in use)
	if msg.Action == chatty.ERROR {
		fmt.Println("ERROR: ", msg.Body)
		//disconnect before exitting
		chConn.Conn.Close()
		return
	}
	welcome()
	handleClient(user, chConn)
}

func handleClient(user string, chConn chatty.ChatConn) {
	// channel for passing the user input to this function
	in_chan := make (chan chatty.ChattyMsg)

	// channel for passing chatty messages received from the server
	conn_chan := make (chan chatty.ChattyMsg)
	// channel for notifying this function that the connection to server was lost
	disconnected := make (chan bool, 1)

	// go routine for communicating with server
	go recvr (chConn, conn_chan, disconnected)
	// go routine for interacting with user
	go getInput(in_chan)
	
	// main loop checks if messages are received from any channel and acts accordingly
	for {
		select { // tests if data is available on any channel (blocks if nothing is available)
		case <-disconnected: // connection to server was lost
			fmt.Println ("\nConnection lost. Exitting")
			os.Stdin.Close()
			wg.Wait() // wait for recvr() to finish its work
			return //exit

		case msg := <-in_chan: // user created a new chatty message
			if msg.Action == chatty.MSG {
				msg.Body = user + ": " + msg.Body // add username of the sender at the front of the message body
			}
			chatty.SendMsg(chConn, msg) // send chatty message to server

		case msg := <-conn_chan: // a chatty message is received from server
			switch msg.Action {
			case chatty.MSG: // a message from another client
				muxPrint(msg.Body) // print message to console
			case chatty.LIST: // server sent a list of other connected clients
				muxPrint(msg.Body)
			case chatty.ERROR: // error message from server
				muxPrint("error: " + msg.Body) // print error content
			case chatty.DISCONNECT: // the server allows client to close the connection
				muxPrint ("Disconnecting")
				
				chConn.Conn.Close() // close TCP connection
				os.Stdin.Close() // close input from console
				wg.Wait() // wait for recvr() and getInput() to exit
				return // exit
			}
		}
	}
}
/*	function to interact with the user and send chatty message structure to the main function Start()
	input:
		in_chan: the channel to pass the chatty message structure
*/
func getInput (in_chan chan chatty.ChattyMsg) {
//	wg.Add(1) // add getInput() to work group so that Start() can only exit after getInput() exits
//	defer wg.Done() // notify Start() that getInput() is done
	scanner := bufio.NewScanner(os.Stdin)
	muxPrint ("") // print prompt only
	for scanner.Scan() { // run loop as long as os.Stdin() is running
		cmd := scanner.Text() // get text read from console
		
		switch cmd {
		case SEND_MSG: // the user wants to send a message to some other user
			consMux.Lock() // get ownership of the console
			// get username of the destination
			fmt.Print("Enter destination username: ")
			scanner.Scan()
		    uname := scanner.Text()
			if uname == CANCEL {
				fmt.Println("message cancelled")
				consMux.Unlock()
				continue
			}
			// get content of the message
			fmt.Print("Enter message to send: ")
			scanner.Scan()
		    txt := scanner.Text()
			if txt == CANCEL {
				fmt.Println("message cancelled")
				consMux.Unlock()
				continue
			}
			consMux.Unlock() // release console mutex
			// create chatty message structure
			msg := chatty.ChattyMsg{Username: uname, Body: txt, Action: chatty.MSG}
			in_chan <- msg // put chatty message on the channel
		case LIST: // the user wants to get a list of other connected clients
			msg := chatty.ChattyMsg{Action: chatty.LIST, Body: ""} //create chatty message with the right purpose
			in_chan <- msg // put chatty message on the channel
		case QUIT: // the client wants to quit the chat app
			msg := chatty.ChattyMsg{Action: chatty.DISCONNECT} // create a DISCONNECT chatty message
			in_chan <- msg // put chatty message on the channel
			return // exit
		case HELP:
			help()

		default:
			muxPrint(cmd + " is not a valid command. Type \"" + HELP + "\" to see the list of valid commands")
		}
		muxPrint (" ")
	}
}
/*	function to receive messages from server
	input:
		chConn: 	the TCP connection along with the gob encoder and decoder
		conn_chan: 	the channel for passing chatty messages received from server
		disconnected: the channel for notifying Start that the TCP connection is closed
*/
func recvr (chConn chatty.ChatConn, conn_chan chan chatty.ChattyMsg, disconnected chan bool) {
	wg.Add(1) // add recvr to work group so that Start() can only exit after recvr exits
	defer wg.Done() // notify Start that recvr is done
	connected := true
	for connected {
		msg, err := chatty.RecvMsg(chConn) // receive message from server (blocks)
		if err == nil {
			conn_chan <- msg // pass chatty message onto the channel
		} else { // the connection was closed (either by the server or by Start())
			fmt.Println ("\rDisconnected")
			disconnected <- true
			connected = false
		}
	}
}
/*	function to print welcome message after the user is accepted by the server*/
func welcome () {
	muxPrint("Welcome to the chatty app! A terminal based chat application.\nYou can see the available commands by typing \"" + HELP + "\"")
}
/*	function to print the help when the user requests it*/
func help () {
	muxPrint ("List of commands:\n" +
				SEND_MSG + "\tsend a message\n" + 
				LIST + "\tget the list of other connected clients\n" +
				QUIT + "\tquit the app\n" +
				CANCEL + "\tcancel a message (while writing one)\n" +
				HELP + "\tget help")
}

/*	function to avoid race conditions on the console
	input:
		str: text to print on console
*/
func muxPrint (str string) {
	consMux.Lock() // get ownership of the console (blocks if another one is currently printing)
	fmt.Println("\r" + str) // print text
	fmt.Printf("%% ") // print prompt for user
	consMux.Unlock() // release console mutex
}
