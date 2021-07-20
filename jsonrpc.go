package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
)

func jsonRPCHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	if r.Body == nil {
		http.NotFound(w, r)
		return
	}
	defer r.Body.Close()
	res := jsonRPCRequest{r.Body, &bytes.Buffer{}}
	codec := codec{codec: jsonrpc.NewServerCodec(&res)}
	rpc.ServeRequest(&codec)
	w.Header().Set("Content-Type", "application/json")
	if codec.isError {
		w.WriteHeader(400)
	}
	_, err := io.Copy(w, res.readWriter)
	if err != nil {
		log.Println("response error:", err)
	}
}

type jsonRPCRequest struct {
	reader     io.Reader
	readWriter io.ReadWriter
}

func (r *jsonRPCRequest) Read(p []byte) (n int, err error)  { return r.reader.Read(p) }
func (r *jsonRPCRequest) Write(p []byte) (n int, err error) { return r.readWriter.Write(p) }
func (r *jsonRPCRequest) Close() error                      { return nil }

type codec struct {
	codec   rpc.ServerCodec
	request *rpc.Request
	isError bool
}

func (c *codec) ReadRequestHeader(r *rpc.Request) error {
	c.request = r
	return c.codec.ReadRequestHeader(r)
}

func (c *codec) ReadRequestBody(x interface{}) error {
	err := c.codec.ReadRequestBody(x)
	b, _ := json.Marshal(x)
	log.Println("->", c.request.ServiceMethod, "-", strings.TrimSpace(string(b)))
	return err
}

func (c *codec) WriteResponse(r *rpc.Response, x interface{}) error {
	if r.Error == "" {
		b, _ := json.Marshal(x)
		log.Println("<-", r.ServiceMethod, "-", strings.TrimSpace(string(b)))
	} else {
		c.isError = true
		log.Println("<-", r.ServiceMethod, "-", r.Error)
	}
	return c.codec.WriteResponse(r, x)
}

func (c *codec) Close() error {
	return c.codec.Close()
}
