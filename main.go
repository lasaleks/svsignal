package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	goutils "github.com/lasaleks/go-utils"
	"github.com/lasaleks/gormq"
)

var VERSION string
var BUILD string

var DEBUG_LEVEL = 0

var (
	config_file = flag.String("config-file", "etc/config.yaml", "path config file")
	pid_file    = flag.String("pid", "", "path pid file")
	get_version = flag.Bool("version", false, "version")
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
	var cfg Config
	if err := cfg.ParseConfig(*config_file); err != nil {
		log.Panicln(err)
	}
	fmt.Printf("%+v\n", cfg)

	// connect DB
	uri := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.CONFIG_SERVER.MYSQL.USER, cfg.CONFIG_SERVER.MYSQL.PASSWORD, cfg.CONFIG_SERVER.MYSQL.HOST, cfg.CONFIG_SERVER.MYSQL.PORT, cfg.CONFIG_SERVER.MYSQL.DATABASE)
	fmt.Println(uri)
	var err error

	db, err := sql.Open("mysql", uri)

	if err != nil {
		log.Println("Error Open", err)
	}
	defer db.Close()

	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	err = migratedb(db)
	if err != nil {
		panic(err)
	}

	savesignal := newSVS(cfg)
	savesignal.db = db
	savesignal.debug_level = cfg.SVSIGNAL.DEBUG_LEVEL
	DEBUG_LEVEL = cfg.SVSIGNAL.DEBUG_LEVEL
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.run(&wg, ctx_db)

	hub := newHub()
	hub.CH_SAVE_VALUE = savesignal.CH_SAVE_VALUE
	hub.CH_SET_SIGNAL = savesignal.CH_SET_SIGNAL
	//hub.CH_REQUEST_HTTP_DB = savesignal.CH_REQUEST_HTTP
	hub.debug_level = cfg.SVSIGNAL.DEBUG_LEVEL
	ctx_hub, cancel_hub := context.WithCancel(ctx)
	wg.Add(1)
	go hub.run(&wg, ctx_hub)

	conn_rmq, err := gormq.NewConnect(cfg.SVSIGNAL.RABBITMQ.URL)
	if err != nil {
		log.Panicln("connect rabbitmq", err)
	}

	chCons, err := gormq.NewChannelConsumer(
		&wg, conn_rmq, []gormq.ExhangeOptions{
			{
				Name:         "svsignal",
				ExchangeType: "topic",
				Keys:         []string{"svs.*.*.#"},
			},
		},
		gormq.QueueOption{
			QOS:  cfg.SVSIGNAL.RABBITMQ.QOS,
			Name: cfg.SVSIGNAL.RABBITMQ.QUEUE_NAME,
		},
		hub.CH_MSG_AMPQ,
	)
	if err != nil {
		log.Panicln("consumer error\n", err)
	}

	//---http
	http := HttpSrv{
		Addr:       cfg.SVSIGNAL.HTTP.Address,
		UnixSocket: cfg.SVSIGNAL.HTTP.UnixSocket,
		svsignal:   savesignal,
		cfg:        &cfg,
	}
	err = http.initUnixSocketServer()
	if err != nil {
		panic(err)
	}
	http.hub = hub
	http.svsignal = savesignal
	wg.Add(1)
	go http.Run(&wg)
	//-------

	f_shutdown := func(ctx context.Context) {
		fmt.Println("ShutDown")
		// close rabbitmq
		conn_rmq.Close()
		chCons.Close()

		cancel_hub()
		cancel_db()
		http.Close()

	}
	wg.Add(1)
	go goutils.WaitSignalExit(&wg, ctx, f_shutdown)
	// ждем освобождение горутин
	wg.Wait()
	fmt.Println("End")
}
