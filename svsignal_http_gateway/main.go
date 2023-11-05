package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	goutils "github.com/lasaleks/go-utils"
	"github.com/lasaleks/svsignal/config"
)

var VERSION string
var BUILD string

func GetVersion() string {
	return fmt.Sprintf("%s build:%s", VERSION, BUILD)
}

var (
	config_file = flag.String("config-file", "etc/config.yaml", "path config file")
	pid_file    = flag.String("pid-file", "", "path pid file")
	get_version = flag.Bool("version", false, "--version")
)

var (
	DEBUG_LEVEL = 0
	//redis_cli_db0 *redis.Client
	//now_unixtime  NowUnixTimeI
	cfg config.Config
)

func main() {
	var wg sync.WaitGroup
	ctx := context.Background()

	flag.Parse()
	fmt.Println(VERSION+" build:", BUILD)
	if *get_version {
		return
	}

	if len(*pid_file) > 0 {
		goutils.CreatePidFile(*pid_file)
		defer os.Remove(*pid_file)
	}

	// загрузка конфигурации
	if err := cfg.ParseConfig(*config_file); err != nil {
		log.Panicf("file %s parse config error %v", *config_file, err)
	}
	fmt.Printf("%+v\n", cfg)

	proxy := HttpProxy{
		proxyAddr:     cfg.SVSIGNAL_HTTP_GATEWAY.PROXY_HTTP_ADDR,
		grpcAddr:      cfg.SVSIGNAL_HTTP_GATEWAY.GPRC_ADDRESS,
		proxyUnixSock: cfg.SVSIGNAL_HTTP_GATEWAY.PROXY_UNIXSOCKET,
	}
	err := proxy.initUnixSocketServer()
	if err != nil {
		panic(err)
	}

	wg.Add(1)
	go proxy.run(&wg)

	f_shutdown := func(ctx context.Context) {
		wg.Done()
		fmt.Println("ShutDown")
		proxy.close()
	}
	wg.Add(1)
	go goutils.WaitSignalExit(&wg, ctx, f_shutdown)
	// ждем освобождение горутин
	wg.Wait()
}
