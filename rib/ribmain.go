package main

import (
	"flag"
	"git.apache.org/thrift.git/lib/go/thrift"
	"log"
	"os"
	"ribd"
)

var logger *log.Logger

func main() {
	var transport thrift.TServerTransport
	var err error
	var addr = "localhost:5000"

	logger = log.New(os.Stdout, "RIBD :", log.Ldate|log.Ltime|log.Lshortfile)

	transport, err = thrift.NewTServerSocket(addr)
	if err != nil {
		logger.Println("Failed to create Socket with:", addr)
	}
	paramsDir := flag.String("params", "", "Directory Location for config files")
	flag.Parse()
	handler := NewRouteServiceHandler(*paramsDir)
	processor := ribd.NewRouteServiceProcessor(handler)
	transportFactory := thrift.NewTBufferedTransportFactory(8192)
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	logger.Println("Starting RIB daemon")
	server.Serve()
}
