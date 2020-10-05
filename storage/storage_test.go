package storage

import (
	"testing"
	"time"
)

const (
	TIMEOUT = 0xFFFFFFFE
)

func TestSetTimeout(t *testing.T) {
	SetTimeout(TIMEOUT)
	if m_timeout != TIMEOUT {
		t.Error("Timeout set failure")
	}
}

func TestGetTimeout(t *testing.T) {
	SetTimeout(TIMEOUT)
	timeout := GetTimeout()

	if timeout != TIMEOUT {
		t.Error("Timeout get/set failure")
	}
}

func TestStoreValue(t *testing.T) {
	t_now := time.Now()
	addr := int16(0x0F)
	valu := int16(0xFF)
	StoreValue(addr, valu)
	t_stored := m_timeouts[addr]
	v_stored := m_storage[addr]

	if v_stored != valu {
		t.Error("Value didn't stored")
	}

	seconds := t_now.Sub(t_stored).Seconds()

	if seconds > 1 {
		t.Error("Too old value for time")
	}
}

func TestGetValue(t *testing.T) {
	addr := int16(0x0F)
	valu := int16(0xFF)
	SetTimeout(1)
	StoreValue(addr, valu)
	code, v_stored := GetValue(addr)

	if code != S_OK {
		t.Error("Failed to store data: ", code)
	} else if v_stored != valu {
		t.Error("Retreived value is incorrect: ", valu, v_stored)
	}

	code, v_stored = GetValue(addr + 1)

	if code != E_EMPTY {
		t.Error("Non empty cell for unused address")
	}

	time.Sleep(1100 * time.Millisecond)

	code, v_stored = GetValue(addr)

	if code != W_TIMEOUT {
		t.Error("Immortal cell value")
	}
}
