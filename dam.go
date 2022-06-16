package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	ErrTimeout      = errors.New("timeout")
	ErrProcessing   = errors.New("already processing")
	ErrNoContent    = errors.New("no content")
	ErrNoSuchClient = errors.New("no such client")
)

type (
	DAM struct {
		Clients *sync.Map
	}

	BasicArgs struct {
		ClientId string
	}

	AddressArgs struct {
		BasicArgs
		Address int
	}

	NumberArgs struct {
		AddressArgs
		Number int
	}

	OpenCloseArgs struct {
		NumberArgs
		Decisecond int
	}

	Client interface {
		GetChannels() *sync.Map
		Write([]byte) (int, error)
	}

	Status struct {
		Number int
		On     bool
	}
)

func (d *DAM) GetAddress(args *BasicArgs, reply *int) error {
	channelKey, bytes := build(254, 4, toInt(0x03, 0xe8), toInt(0x00, 0x01))
	ret, err := d.write(args.ClientId, channelKey, bytes)
	if err != nil {
		return err
	}
	*reply = toInt(ret[3], ret[4])
	return nil
}

func (d *DAM) GetStatuses(args *AddressArgs, reply *[]Status) error {
	channelKey, bytes := build(254, 1, 0, 16)
	ret, err := d.write(args.ClientId, channelKey, bytes)
	if err != nil {
		return err
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < 8; j++ {
			*reply = append(*reply, Status{
				Number: 8*i + j + 1,
				On:     int(ret[3+i])&(1<<j) != 0,
			})
		}
	}
	return nil
}

func (d *DAM) OpenAll(args *AddressArgs, reply *bool) (err error) {
	channelKey, bytes := build(254, 0x0F, 0, 16, 0xFF, 0xFF)
	_, err = d.write(args.ClientId, channelKey, bytes)
	*reply = err == nil
	return err
}

func (d *DAM) CloseAll(args *AddressArgs, reply *bool) (err error) {
	channelKey, bytes := build(254, 0x0F, 0, 16, 0x00, 0x00)
	_, err = d.write(args.ClientId, channelKey, bytes)
	*reply = err == nil
	return err
}

func (d *DAM) Open(args *NumberArgs, reply *bool) (err error) {
	channelKey, bytes := build(254, 0x05, args.Number-1, 65280)
	_, err = d.write(args.ClientId, channelKey, bytes)
	*reply = err == nil
	return err
}

func (d *DAM) Close(args *NumberArgs, reply *bool) (err error) {
	channelKey, bytes := build(254, 0x05, args.Number-1, 0)
	_, err = d.write(args.ClientId, channelKey, bytes)
	*reply = err == nil
	return err
}

func (d *DAM) OpenClose(args *OpenCloseArgs, reply *bool) (err error) {
	addr := 5*(args.Number-1) + 3
	ds := args.Decisecond
	if ds < 1 {
		ds = 10
	}
	channelKey, bytes := build(254, 0x10, addr, 2, 0, 4, byte(ds>>8), byte(ds))
	_, err = d.write(args.ClientId, channelKey, bytes)
	*reply = err == nil
	return err
}

func (d *DAM) CloseOpen(args *OpenCloseArgs, reply *bool) (err error) {
	addr := 5*(args.Number-1) + 3
	ds := args.Decisecond
	if ds < 1 {
		ds = 10
	}
	channelKey, bytes := build(254, 0x10, addr, 2, 0, 2, byte(ds>>8), byte(ds))
	_, err = d.write(args.ClientId, channelKey, bytes)
	*reply = err == nil
	return err
}

func (d *DAM) write(clientId, channelKey string, input []byte) (output []byte, err error) {
	if len(input) == 0 {
		err = ErrNoContent
		return
	}
	if d.Clients == nil {
		err = ErrNoSuchClient
		return
	}
	_client, ok := d.Clients.Load(clientId)
	if !ok {
		err = ErrNoSuchClient
		return
	}
	client, ok := _client.(Client)
	if !ok {
		err = ErrNoSuchClient
		return
	}
	channels := client.GetChannels()
	channel, hasChannel := channels.LoadOrStore(channelKey, make(chan []byte))
	if hasChannel {
		err = ErrProcessing
		return
	} else {
		defer channels.Delete(channelKey)
	}
	var n int
	n, err = client.Write(input)
	if err != nil {
		log.Println("error writting", input, err)
		return
	}
	if clientId == "" {
		log.Printf("%d bytes written: % X", n, input)
	} else {
		log.Printf("%s %d bytes written: % X", clientId, n, input)
	}
	timeoutChan := time.After(time.Duration(1000) * time.Millisecond)
	for {
		select {
		case output = <-channel.(chan []byte):
			return
		case <-timeoutChan:
			err = ErrTimeout
			return
		}
	}
}

func toInt(a, b byte) int {
	return int(a)<<8 | int(b)
}

func build(addr, function byte, start, size int, command ...byte) (string, []byte) {
	b := []byte{addr, function, byte(start >> 8), byte(start), byte(size >> 8), byte(size)}
	if len(command) > 0 {
		b = append(b, byte(len(command)))
		b = append(b, command...)
	}
	return fmt.Sprintf("%d-%d", addr, function), append(b, checksum(b)...)
}
