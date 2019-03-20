// Copyright (c) 2014-2015 The btcsuite developers
// Copyright (c) 2015-2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
  "os/signal"
	"path/filepath"
  "syscall"
	"time"

	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpcclient"
)

var rpcuser = flag.String("rpcuser", "", "RPC Username defined in dcrd.conf")
var rpcpass = flag.String("rpcpass", "", "RPC Password defined in dcrd.conf")

func main() {
	flag.Parse()
	// Only override the handlers for notifications you care about.
	// Also note most of these handlers will only be called if you register
	// for notifications.  See the documentation of the rpcclient.
	// NotificationHandlers type for more details about each handler.
	ntfnHandlers := rpcclient.NotificationHandlers{
		OnBlockConnected: func(blockHeader []byte, transactions [][]byte) {
		//	log.Printf("Block connected: %v %v", blockHeader, transactions)
		},
		OnBlockDisconnected: func(blockHeader []byte) {
		//	log.Printf("Block disconnected: %v", blockHeader)
		},
	}

	// Connect to local dcrd RPC server using websockets.
	dcrdHomeDir := dcrutil.AppDataDir("dcrd", false)
	certs, err := ioutil.ReadFile(filepath.Join(dcrdHomeDir, "rpc.cert"))
	if err != nil {
		log.Fatal(err)
	}
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:9109",
		Endpoint:     "ws",
		User:         *rpcuser,
		Pass:         *rpcpass,
		Certificates: certs,
	}
	client, err := rpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		log.Fatal(err)
	}

	// Register for block connect and disconnect notifications.
	if err := client.NotifyBlocks(); err != nil {
		log.Fatal(err)
	}
	log.Println("NotifyBlocks: Registration Complete")

	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Block count: %d", blockCount)

	c := make(chan os.Signal)
  signal.Notify(c, os.Interrupt, syscall.SIGTERM)
  go func() {
      <-c
			log.Println("SIGTERM received!")
			for s := 5; s > 0; s-- {
				data := s
				message := fmt.Sprintf("Client shutdown in %d seconds...", data)
				log.Println(message)
				time.Sleep(time.Millisecond * 1000)
			}
			client.Shutdown()
			log.Println("Client shutdown complete.")
  }()

	// Wait until the client either shuts down gracefully (or the user
	// terminates the process with Ctrl+C).
	client.WaitForShutdown()
}
