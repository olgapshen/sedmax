package main

import (
	"fmt"
	"net"
	"os"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "1502"
	CONN_TYPE = "tcp"
)

const (
	CMD_PERSET_M_REG = 0x10
)

func main() {
	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

var Storage []byte

type PDUHeader struct {
	txid int16
	prid int16
	leng int16
	addr byte
	metd byte
}

type RequestPresetMReg struct {
	addr int16
	rcnt int16
	leng byte
	data []byte
}

type ResponsePresetMReg struct {
	addr int16
	amnt int16
}

func parseHeader(buf []byte) PDUHeader {
	var reqHeader PDUHeader
	reqHeader.txid = int16(buf[0]<<8) | int16(buf[1])
	reqHeader.prid = int16(buf[2]<<8) | int16(buf[3])
	reqHeader.leng = int16(buf[4]<<8) | int16(buf[5])
	reqHeader.addr = buf[6]
	reqHeader.metd = buf[7]

	return reqHeader
}

func parsePersetMReg(buf []byte) RequestPresetMReg {
	var command RequestPresetMReg
	command.addr = int16(buf[0]<<8) | int16(buf[1])
	command.rcnt = int16(buf[2]<<8) | int16(buf[3])
	command.leng = buf[4]
	command.data = buf[5 : command.leng+5]
	return command
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)

	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	arrHeader := buf[0:8]

	//fmt.Println(arrHeader)
	//fmt.Println(buf)

	reqHeader := parseHeader(arrHeader)
	//fmt.Printf("%v\n", reqHeader)
	//fmt.Println(buf)

	arrResp := make([]byte, 12)

	for i := 0; i < reqLen; i++ {
		fmt.Printf("%x ", buf[i])
	}
	fmt.Println()

	switch reqHeader.metd {
	case CMD_PERSET_M_REG:
		persetMReg := parsePersetMReg(buf[8 : reqHeader.leng+8])
		//arrResp := make([]byte, 12)
		copy(arrHeader, arrResp)
		arrResp[8] = byte(persetMReg.addr >> 8)
		arrResp[9] = byte(persetMReg.addr)
		arrResp[10] = byte(5 >> 8)
		arrResp[11] = byte(5)
		//fmt.Printf("%v\n", persetMReg)
	default:
		fmt.Println("Unknown MODBus method")
	}

	for i := 0; i < reqLen; i++ {
		fmt.Printf("%x ", buf[i])
	}

	fmt.Println()

	for i := 0; i < 12; i++ {
		//fmt.Printf("%x ", arrResp[i])
	}

	fmt.Println()

	// Close the connection when you're done with it.
	conn.Close()
}
