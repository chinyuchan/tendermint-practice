package main

import (
	"flag"
	"fmt"
	"github.com/dgraph-io/badger"
	abciserver "github.com/tendermint/tendermint/abci/server"
	"github.com/tendermint/tendermint/libs/log"
	"os"
	"os/signal"
	"syscall"
	"tendermint-practice/types"
)

func main() {
	// 打开DB
	db, err := badger.Open(badger.DefaultOptions("/home/cdd/data/kvstore"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open badger DB: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// 解析命令行参数
	flag.Parse()
	// 创建App对象
	app := types.NewKVStoreApp(db)
	// 创建Logger对象
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	// 创建Server对象
	server, err := abciserver.NewServer("tcp://127.0.0.1:26658", "socket", app)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create server: %v", err)
		os.Exit(1)
	}
	//
	server.SetLogger(logger)

	// 启动
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error starting socket server: %v", err)
		os.Exit(1)
	}
	defer server.Stop()

	// 处理信号，退出
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	os.Exit(0)
}
