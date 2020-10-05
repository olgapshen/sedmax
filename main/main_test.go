package main

import (
	"testing"
)

func TestParseHeader(t *testing.T) {
	rawData := [...]byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0f, 0x01, 0x10}
	header := parseHeader(rawData[:])

	if header.txid != 1 {
		t.Error("Bad transaction id")
	}

	if header.prid != 0 {
		t.Error("Bad protocol id")
	}

	if header.leng != 0x0f {
		t.Error("Bad length")
	}

	if header.addr != 1 {
		t.Error("Bad device address")
	}

	if header.metd != CMD_PRESET_M_REG {
		t.Error("Bad command")
	}
}

func TestSerializeHeader(t *testing.T) {
	var header PDUHeader
	header.txid = 1
	header.prid = 0
	header.leng = 8
	header.addr = 1
	header.metd = CMD_READ_HOL_REG

	rawData := serializeHeader(header)

	expected := [...]byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x08, 0x01, 0x03}

	if len(rawData) != len(expected) {
		t.Error("Different length")
	}

	for i := 0; i < len(rawData); i++ {
		if rawData[i] != expected[i] {
			t.Error("Data is mismatch")
		}
	}
}
