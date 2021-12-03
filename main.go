package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"myutils"
	"myutils/rabbitmq"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const VERSION = "0.0.2"

var (
	config_file = flag.String("config-file", "etc/config.yaml", "path config file")
	pid_file    = flag.String("pid", "", "path pid file")
)

func main() {
	fmt.Println("Start svsignal ver:", VERSION)

	var wg sync.WaitGroup
	ctx := context.Background()

	flag.Parse()
	if len(*pid_file) > 0 {
		myutils.CreatePidFile(*pid_file)
	}
	// загрузка конфигурации
	var cfg Config
	cfg.parseConfig(*config_file)
	fmt.Println(cfg)

	// connect DB
	uri := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.CONFIG.MYSQL.USER, cfg.CONFIG.MYSQL.PASSWORD, cfg.CONFIG.MYSQL.HOST, cfg.CONFIG.MYSQL.PORT, cfg.CONFIG.MYSQL.DATABASE)
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
	//

	savesignal := newSVS()
	savesignal.db = db
	savesignal.debug_level = cfg.CONFIG.DEBUG_LEVEL
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.run(&wg, ctx_db)

	hub := newHub()
	hub.CH_SAVE_VALUE = savesignal.CH_SAVE_VALUE
	hub.CH_SET_SIGNAL = savesignal.CH_SET_SIGNAL
	//hub.CH_REQUEST_HTTP_DB = savesignal.CH_REQUEST_HTTP
	hub.debug_level = cfg.CONFIG.DEBUG_LEVEL
	ctx_hub, cancel_hub := context.WithCancel(ctx)
	wg.Add(1)
	go hub.run(&wg, ctx_hub)

	exchanges := []rabbitmq.ConsumerExchange{
		{
			Name:         "svsingal",
			ExchangeType: "topic",
			Keys: []string{
				fmt.Sprint("svs.*.*.#"),
			},
		},
	}

	consumer, err := rabbitmq.NewConsumer(cfg.CONFIG.RABBITMQ.URL, exchanges, cfg.CONFIG.RABBITMQ.QUEUE_NAME, "simple-consumer", hub.CH_MSG_AMPQ)
	if err != nil {
		log.Fatalf("ERROR %s", err)
	}

	//---http
	http := HttpSrv{Addr: cfg.CONFIG.HTTP.Address, svsignal: savesignal}
	http.hub = hub
	http.svsignal = savesignal
	wg.Add(1)
	go http.Run(&wg)
	//-------

	f_shutdown := func(ctx context.Context) {
		fmt.Println("ShutDown")
		err := consumer.Shutdown()
		cancel_hub()
		cancel_db()
		http.Close()
		if err != nil {
			log.Println("Consumer", err)
		}
	}
	wg.Add(1)
	go myutils.WaitSignalExit(&wg, ctx, f_shutdown)
	// ждем освобождение горутин
	wg.Wait()
}
