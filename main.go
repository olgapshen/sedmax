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
	HEADER_LENGTH           = 8
	RESP_PERS_MREQ_D_LENGTH = 4

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
		go handleTCPRequest(conn)
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
	head PDUHeader
	addr int16
	rcnt int16
	leng byte
	data []byte
}

type ResponsePresetMReg struct {
	head PDUHeader
	addr int16
	amnt int16
}

func sliceHeader(pkg []byte) []byte {
	return pkg[0:HEADER_LENGTH]
}

func sliceBody(pkg []byte, dataLength int16) []byte {
	return pkg[HEADER_LENGTH : dataLength+HEADER_LENGTH]
}

func parseHeader(buf []byte) PDUHeader {
	var reqHeader PDUHeader
	reqHeader.txid = int16(buf[0])<<8 | int16(buf[1])
	reqHeader.prid = int16(buf[2])<<8 | int16(buf[3])
	reqHeader.leng = int16(buf[4])<<8 | int16(buf[5])
	reqHeader.addr = buf[6]
	reqHeader.metd = buf[7]

	return reqHeader
}

func parsePersetMReg(buf []byte) RequestPresetMReg {
	var command RequestPresetMReg
	command.addr = int16(buf[0])<<8 | int16(buf[1])
	command.rcnt = int16(buf[2])<<8 | int16(buf[3])
	command.leng = buf[4]
	command.data = buf[5 : command.leng+5]
	return command
}

func serializeHeader(header PDUHeader) []byte {
	buf := make([]byte, HEADER_LENGTH)
	buf[0] = byte(header.txid >> 8)
	buf[1] = byte(header.txid)
	buf[2] = byte(header.prid >> 8)
	buf[3] = byte(header.prid)
	buf[4] = byte(header.leng >> 8)
	buf[5] = byte(header.leng)
	buf[6] = header.addr
	buf[7] = header.metd

	return buf
}

func serilizePersetMReg(resp ResponsePresetMReg) []byte {
	buf := make([]byte, HEADER_LENGTH+RESP_PERS_MREQ_D_LENGTH)
	copy(buf, serializeHeader(resp.head))
	buf[8] = byte(resp.addr >> 8)
	buf[9] = byte(resp.addr)
	buf[10] = byte(resp.amnt >> 8)
	buf[11] = byte(resp.amnt)

	return buf
}

func handlePersetMReg(request RequestPresetMReg) ResponsePresetMReg {
	var response ResponsePresetMReg
	response.head = request.head
	response.addr = request.addr
	response.amnt = request.rcnt

	return response
}

func handleTCPRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	bufReq := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(bufReq)

	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}

	arrHeader := sliceHeader(bufReq)
	reqHeader := parseHeader(arrHeader)
	reqBody := sliceBody(bufReq, reqHeader.leng)

	var bufResp []byte
	//arrResp := make([]byte, 12)

	switch reqHeader.metd {
	case CMD_PERSET_M_REG:
		persetMReg := parsePersetMReg(reqBody)
		persetMReg.head = reqHeader
		//arrResp := make([]byte, 12)
		//copy(arrResp, arrHeader)
		response := handlePersetMReg(persetMReg)
		bufResp = serilizePersetMReg(response)

		//fmt.Printf("%v\n", persetMReg)
	default:
		fmt.Println("Unknown MODBus method")
	}

	for i := 0; i < reqLen; i++ {
		//fmt.Printf("%x ", buf[i])
	}

	fmt.Println()

	for i := 0; i < 12; i++ {
		//fmt.Printf("%x ", arrResp[i])
	}

	fmt.Println()

	conn.Write(bufResp)

	// Close the connection when you're done with it.
	conn.Close()
}
