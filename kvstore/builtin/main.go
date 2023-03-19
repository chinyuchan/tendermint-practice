package main

import (
	"flag"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/spf13/viper"
	tdabcitypes "github.com/tendermint/tendermint/abci/types"
	tdconfig "github.com/tendermint/tendermint/config"
	tdflags "github.com/tendermint/tendermint/libs/cli/flags"
	tdlog "github.com/tendermint/tendermint/libs/log"
	tdnode "github.com/tendermint/tendermint/node"
	tdp2p "github.com/tendermint/tendermint/p2p"
	tdprivval "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"tendermint-practice/types"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/home/cdd/.tendermint/config/config.toml", "Path to config.toml")
}

func main() {
	db, err := badger.Open(badger.DefaultOptions("/home/cdd/data/kvstore"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open badger DB: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// 创建App对象
	app := types.NewKVStoreApp(db)

	// 创建结点
	node, err := newTendermint(app, configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to new tendermint: %v", err)
		os.Exit(2)
	}

	// 启动结点
	node.Start()
	defer func() {
		node.Stop()
		node.Wait()
	}()

	// 处理信号，退出
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func newTendermint(app tdabcitypes.Application, configFile string) (*tdnode.Node, error) {
	// parse config file
	config := tdconfig.DefaultConfig()
	config.RootDir = filepath.Dir(filepath.Dir(configFile))

	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("viper failed to read config file: %w", err)
	}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("viper failed to unmarshal config: %w", err)
	}
	if err := config.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("config is invalid: %w", err)
	}

	// create logger
	logger := tdlog.NewTMLogger(tdlog.NewSyncWriter(os.Stdout))
	logger, err := tdflags.ParseLogLevel(config.LogLevel, logger, tdconfig.DefaultLogLevel)

	// read pv
	pv := tdprivval.LoadFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())
	if err != nil {
		return nil, fmt.Errorf("load file pv: %w", err)
	}

	// read node key
	nodeKey, err := tdp2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, fmt.Errorf("load node key: %w", err)
	}

	// create node
	node, err := tdnode.NewNode(config, pv, nodeKey, proxy.NewLocalClientCreator(app), tdnode.DefaultGenesisDocProviderFunc(config), tdnode.DefaultDBProvider, tdnode.DefaultMetricsProvider(config.Instrumentation), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to new node: %w", err)
	}

	return node, nil
}
