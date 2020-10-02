package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "1502"
	CONN_TYPE = "tcp"
)

const (
	HEADER_LENGTH           = 8
	RESP_PERS_MREQ_D_LENGTH = 4
	REQ_PERS_MREQ_MD_LENGTH = 5

	CMD_PERSET_M_REG = 0x10
	CMD_READ_HOL_REG = 0x03
)

var Storage map[int16]int16
var Timeouts map[int16]time.Time

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
	data []int16
}

type RequestReadHReg struct {
	head PDUHeader
	addr int16
	rcnt int16
}

type ResponsePresetMReg struct {
	head PDUHeader
	addr int16
	amnt int16
}

type ResponseReadHReg struct {
	head PDUHeader
	leng byte
	data []int16
}

func joinBytes(buf []byte) int16 {
	return int16(buf[0])<<8 | int16(buf[1])
}

func splitBytes(buf []byte, value int16) {
	buf[0] = byte(value >> 8)
	buf[1] = byte(value)
}

func sliceHeader(pkg []byte) []byte {
	return pkg[0:HEADER_LENGTH]
}

func sliceBody(pkg []byte, dataLength int16) []byte {
	return pkg[HEADER_LENGTH : dataLength+HEADER_LENGTH]
}

func parseHeader(buf []byte) PDUHeader {
	var reqHeader PDUHeader
	reqHeader.txid = joinBytes(buf[0:])
	reqHeader.prid = joinBytes(buf[2:])
	reqHeader.leng = joinBytes(buf[4:])
	reqHeader.addr = buf[6]
	reqHeader.metd = buf[7]

	return reqHeader
}

func serializeHeader(header PDUHeader) []byte {
	buf := make([]byte, HEADER_LENGTH)
	splitBytes(buf[0:], header.txid)
	splitBytes(buf[2:], header.prid)
	splitBytes(buf[4:], header.leng)
	buf[6] = header.addr
	buf[7] = header.metd

	return buf
}

func parsePersetMReg(header PDUHeader, buf []byte) RequestPresetMReg {
	var request RequestPresetMReg
	request.head = header
	request.addr = joinBytes(buf[0:])
	request.rcnt = joinBytes(buf[2:])
	request.leng = buf[4]
	request.data = make([]int16, request.rcnt)
	rawData := buf[REQ_PERS_MREQ_MD_LENGTH : request.leng+REQ_PERS_MREQ_MD_LENGTH]

	var i int16
	for i = 0; i < request.rcnt; i++ {
		request.data[i] = joinBytes(rawData[i*2:])
	}

	return request
}

func serilizePersetMReg(resp ResponsePresetMReg) []byte {
	buf := make([]byte, HEADER_LENGTH+RESP_PERS_MREQ_D_LENGTH)
	copy(buf, serializeHeader(resp.head))
	splitBytes(buf[8:], resp.addr)
	splitBytes(buf[10:], resp.amnt)

	return buf
}

func parseReadHReg(header PDUHeader, buf []byte) RequestReadHReg {
	var request RequestReadHReg
	request.head = header
	request.addr = joinBytes(buf[0:])
	request.rcnt = joinBytes(buf[2:])
	return request
}

func serializeReadHReg(resp ResponseReadHReg) []byte {
	buf := make([]byte, HEADER_LENGTH+resp.leng+1)
	copy(buf, serializeHeader(resp.head))
	//splitBytes(buf[8:], resp.leng)
	//splitBytes(buf[10:], resp.amnt)
	return buf
}

func handlePersetMReg(request RequestPresetMReg) ResponsePresetMReg {
	var response ResponsePresetMReg
	response.head = request.head

	var i int16
	for i = 0; i < request.rcnt; i++ {
		addr := request.addr + i
		Storage[addr] = request.data[i]
		Timeouts[addr] = time.Now()
	}

	response.addr = request.addr
	response.amnt = request.rcnt

	return response
}

func handleReadHReg() {

}

func handleTCPRequest(conn net.Conn) {
	bufReq := make([]byte, 1024)
	_, err := conn.Read(bufReq)

	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}

	arrHeader := sliceHeader(bufReq)
	reqHeader := parseHeader(arrHeader)
	reqBody := sliceBody(bufReq, reqHeader.leng)

	var bufResp []byte

	switch reqHeader.metd {
	case CMD_PERSET_M_REG:
		persetMReg := parsePersetMReg(reqHeader, reqBody)
		response := handlePersetMReg(persetMReg)
		bufResp = serilizePersetMReg(response)
	case CMD_READ_HOL_REG:
		persetMReg := parsePersetMReg(reqHeader, reqBody)
		response := handlePersetMReg(persetMReg)
		bufResp = serilizePersetMReg(response)
	default:
		fmt.Println("Unknown MODBus method")
	}

	conn.Write(bufResp)

	fmt.Println(Storage)
	fmt.Println(Timeouts)

	conn.Close()
}

func main() {
	Storage = make(map[int16]int16)
	Timeouts = make(map[int16]time.Time)

	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		go handleTCPRequest(conn)
	}
}
