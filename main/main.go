package main

import (
	"../storage"
	"flag"
	"fmt"
	"net"
	"os"
)

const (
	HEADER_LENGTH           = 8
	HEADER_TAIL_LENGTH      = 2
	RES_PRES_MREG_D_LENGTH  = 4 // Data
	REQ_PRES_MREG_MD_LENGTH = 5 // Metadata

	CMD_PRESET_M_REG = 0x10
	CMD_READ_HOL_REG = 0x03

	ERR_READ_FAILED = 2
	ERR_UNKNOWN     = 100
)

type PDUHeader struct {
	txid int16
	prid int16
	leng int16
	addr byte
	metd byte // Method
}

type RequestPresetMReg struct {
	head PDUHeader
	addr int16
	rcnt int16
	leng byte
	data []byte
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
	data []byte
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
	request.data = buf[REQ_PRES_MREG_MD_LENGTH : request.leng+REQ_PRES_MREG_MD_LENGTH]

	return request
}

func serilizePersetMReg(resp ResponsePresetMReg) []byte {
	buf := make([]byte, HEADER_LENGTH+RES_PRES_MREG_D_LENGTH)
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
	buf := make([]byte, HEADER_LENGTH+1+resp.leng)
	copy(buf, serializeHeader(resp.head))

	if resp.leng > 0 {
		buf[HEADER_LENGTH] = resp.leng
		copy(buf[HEADER_LENGTH+1:], resp.data)
	} else {
		buf[HEADER_LENGTH] = resp.data[0]
	}

	return buf
}

func setResponseError(response *ResponseReadHReg, err byte) {
	response.head.metd |= 0x80
	response.leng = 0
	response.data = make([]byte, 1)
	response.data[0] = err
}

func handlePersetMReg(request RequestPresetMReg) ResponsePresetMReg {
	var response ResponsePresetMReg
	response.head = request.head
	response.head.leng = HEADER_TAIL_LENGTH + RES_PRES_MREG_D_LENGTH

	for i := int16(0); i < request.rcnt; i++ {
		addr := request.addr + i
		storage.StoreValue(addr, joinBytes(request.data[i*2:]))
	}

	response.addr = request.addr
	response.amnt = request.rcnt

	return response
}

func handleReadHReg(request RequestReadHReg) ResponseReadHReg {
	var response ResponseReadHReg
	response.head = request.head
	response.head.leng = HEADER_TAIL_LENGTH + 1
	var data = make([]byte, request.rcnt*2)

	for i := int16(0); i < request.rcnt; i++ {
		addr := request.addr + i
		status, elem := storage.GetValue(addr)

		switch status {
		case storage.E_EMPTY:
			setResponseError(&response, ERR_READ_FAILED)
			return response
		case storage.W_TIMEOUT:
			setResponseError(&response, ERR_READ_FAILED)
			return response
		case storage.S_OK:
			splitBytes(data[i*2:], elem)
		default:
			setResponseError(&response, ERR_UNKNOWN)
			return response
		}
	}

	response.head.leng += request.rcnt * 2
	response.leng = byte(request.rcnt) * 2
	response.data = data

	return response
}

func handleTCPRequest(conn net.Conn) {
	bufReq := make([]byte, 1024)
	_, err := conn.Read(bufReq)

	if err != nil {
		fmt.Println("Error reading: ", err.Error())
		return
	}

	arrHeader := sliceHeader(bufReq)
	reqHeader := parseHeader(arrHeader)
	reqBody := sliceBody(bufReq, reqHeader.leng)

	switch reqHeader.metd {
	case CMD_PRESET_M_REG:
		persetMReg := parsePersetMReg(reqHeader, reqBody)
		response := handlePersetMReg(persetMReg)
		conn.Write(serilizePersetMReg(response))
	case CMD_READ_HOL_REG:
		readHReg := parseReadHReg(reqHeader, reqBody)
		response := handleReadHReg(readHReg)
		conn.Write(serializeReadHReg(response))
	default:
		fmt.Println("Unhandled MODBus method: ", reqHeader.metd)
	}

	defer conn.Close()
}

func main() {
	hostPtr := flag.String("host", "localhost", "Host to listen to")
	portPtr := flag.Int("port", 1502, "ModBus port to run on")
	timeoutPtr := flag.Int64("timeout", 2, "Time the data to be expired (in seconds)")

	flag.Parse()

	storage.SetTimeout(*timeoutPtr)
	addr := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)

	l, err := net.Listen("tcp", addr)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer l.Close()
	fmt.Printf("Listening on %s, timeout set to: %d\n", addr, storage.GetTimeout())

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		go handleTCPRequest(conn)
	}
}
