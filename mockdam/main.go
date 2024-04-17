package main

import (
	"bytes"
	"flag"

	"github.com/caiguanhao/mockserialport"
	"go.bug.st/serial"
)

func main() {
	status := []byte{254, 1, 2, 0, 0, 173, 232, 0}
	opts := &mockserialport.Options{
		InputFile:  "ttyIN",
		OutputFile: "ttyOUT",
		PidFile:    "socat.pid",
		SocatPath:  "socat",
		BaudRate:   9600,
		ExtraOpts:  "",
		Verbose:    true,
		Open: func(path string, baudrate int) (mockserialport.Port, error) {
			return serial.Open(path, &serial.Mode{
				BaudRate: baudrate,
			})
		},
		Process: func(mock *mockserialport.Mock, input []byte) []byte {
			if bytes.Equal(input, []byte{254, 4, 3, 232, 0, 1, 165, 181}) { // GetAddress
				mock.Write([]byte{254, 4, 2, 0, 1, 108, 228, 0})
			} else if bytes.Equal(input, []byte{254, 1, 0, 0, 0, 16, 41, 201}) { // GetStatuses
				mock.Write(status)
			} else if bytes.Equal(input, []byte{254, 15, 0, 0, 0, 16, 2, 255, 255, 166, 100}) { // OpenAll
				mock.Write([]byte{254, 15, 0, 0, 0, 16, 64, 8})
				status = []byte{254, 1, 2, 255, 255, 172, 88, 0}
			} else if bytes.Equal(input, []byte{254, 15, 0, 0, 0, 16, 2, 0, 0, 167, 212}) { // CloseAll
				mock.Write([]byte{254, 15, 0, 0, 0, 16, 64, 8})
				status = []byte{254, 1, 2, 0, 0, 173, 232, 0}
			} else if bytes.Equal(input, []byte{254, 16, 0, 3, 0, 2, 4, 0, 4, 0, 10, 65, 107}) { // OpenClose
				mock.Write([]byte{254, 16, 0, 3, 0, 2, 165, 199})
				status = []byte{254, 1, 2, 0, 0, 173, 232, 0}
			} else if bytes.Equal(input, []byte{254, 16, 0, 3, 0, 2, 4, 0, 2, 0, 10, 161, 106}) { // CloseOpen
				mock.Write([]byte{254, 16, 0, 3, 0, 2, 165, 199})
				status = []byte{254, 1, 2, 1, 0, 172, 120, 0}
			} else if bytes.Equal(input, []byte{254, 5, 0, 0, 0, 0, 217, 197}) { // Close
				mock.Write([]byte{254, 5, 0, 0, 0, 0, 217, 197})
				status = []byte{254, 1, 2, 0, 0, 173, 232, 0}
			} else if bytes.Equal(input, []byte{254, 5, 0, 0, 255, 0, 152, 53}) { // Open
				mock.Write([]byte{254, 5, 0, 0, 255, 0, 152, 53})
				status = []byte{254, 1, 2, 1, 0, 172, 120, 0}
			}
			return nil
		},
	}

	opts.SetFlags(flag.CommandLine)
	flag.Parse()
	mock := mockserialport.New(opts)
	if err := mock.Start(); err != nil {
		panic(err)
	}
}
