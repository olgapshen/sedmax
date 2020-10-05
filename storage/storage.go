package storage

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	S_OK      = 0
	W_TIMEOUT = 1
	E_EMPTY   = 2
)

var m_mutex = &sync.Mutex{}

var m_timeout int64
var m_storage map[int16]int16
var m_timeouts map[int16]time.Time

func init() {
	m_storage = make(map[int16]int16)
	m_timeouts = make(map[int16]time.Time)
}

func GetTimeout() int64 {
	return m_timeout
}

func SetTimeout(timeout int64) {
	atomic.StoreInt64(&m_timeout, timeout)
}

func StoreValue(addr int16, val int16) {
	m_mutex.Lock()
	m_storage[addr] = val
	m_timeouts[addr] = time.Now()
	m_mutex.Unlock()
}

func GetValue(addr int16) (byte, int16) {
	m_mutex.Lock()
	defer m_mutex.Unlock()
	elem, ok := m_storage[addr]

	if ok {
		seconds := time.Now().Sub(m_timeouts[addr]).Seconds()
		if int64(seconds) > m_timeout {
			return W_TIMEOUT, elem
		} else {
			return S_OK, elem
		}
	} else {
		return E_EMPTY, 0
	}
}
