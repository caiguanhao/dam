package main

import (
	"bytes"
	"fmt"
	"log"
	"net/rpc"
	"sync"
	"time"

	"go.bug.st/serial"
)

type Port struct {
	serial.Port

	jsonrpc  *DAM
	channels *sync.Map
}

func (port Port) GetChannels() *sync.Map {
	return port.channels
}

func (port Port) read() {
	buf := make([]byte, 20)
	data := []byte{}
	for {
		n, err := port.Read(buf)
		if err != nil {
			log.Println("read:", err)
			return
		}

		if n == 0 {
			log.Println("read: EOF")
			return
		}
		data = append(data, buf[:n]...)
		data = process(data, port.channels)
	}
}

var clients sync.Map
var channels sync.Map
var port = &Port{
	jsonrpc: &DAM{
		Clients: &clients,
	},
	channels: &channels,
}

func init() {
	clients.Store("", port)
	err := rpc.RegisterName("DAM", port.jsonrpc)
	if err != nil {
		log.Fatalln(err)
	}
}

func openPort(name string) {
	for {
		p, err := serial.Open(name, &serial.Mode{
			BaudRate: 9600,
		})
		if err == nil {
			port.Port = p
			break
		}
		log.Println("open:", err)
		time.Sleep(2 * time.Second)
	}
	go func() {
		port.read()
		openPort(name)
	}()
	var i int
	if port.jsonrpc.GetAddress(&BasicArgs{}, &i) == nil {
		log.Println("successfully opened", name)
	}
}

func process(_data []byte, multiChannels ...*sync.Map) []byte {
	for i := 0; i < len(_data)-2; i++ {
		pos := findValidData(_data, i)
		if pos == -1 {
			continue
		}
		data := _data[i:pos]
		_data = _data[pos:]
		i = -1 // i will be reset to 0 after this loop
		log.Println("received:", data)
		key := fmt.Sprintf("%d-%d", int(data[0]), int(data[1]))
		for _, channels := range multiChannels {
			if channel, ok := channels.LoadAndDelete(key); ok {
				channel.(chan []byte) <- data
			}
		}
	}
	return _data
}

func findValidData(data []byte, i int) int {
	pos := i + 2 + 2 + 2 + 2
	if pos <= len(data) {
		bytes := data[i:pos]
		if isValid(bytes) {
			return pos
		}
	}
	if i+2 < len(data) {
		size := int(data[i+2])
		pos = i + 2 + 1 + size + 2
		if pos <= len(data) {
			bytes := data[i:pos]
			if isValid(bytes) {
				return pos
			}
		}
	}
	return -1
}

func isValid(input []byte) bool {
	size := len(input)
	if size < 3 {
		return false
	}
	return bytes.Equal(checksum(input[:size-2]), input[size-2:])
}
