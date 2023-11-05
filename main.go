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
	"github.com/lasaleks/svsignal/config"
	"github.com/lasaleks/svsignal/model"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var VERSION string
var BUILD string

var (
	CH_SAVE_VALUE   chan ValueSignal
	CH_SET_SIGNAL   chan SetSignal
	CH_MSG_AMPQ     chan gormq.MessageAmpq
	CH_REQUEST_DATA chan RequestData

	cfg        config.Config
	DB         *gorm.DB
	hub        *Hub
	savesignal *SVSignalDB
)

var (
	config_file = flag.String("config-file", "etc/config.yaml", "path config file")
	pid_file    = flag.String("pid-file", "", "path pid file")
	get_version = flag.Bool("version", false, "version")
)

func connectDataBase() {
	if cfg.SVSIGNAL.MYSQL != nil {
		uri := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.SVSIGNAL.MYSQL.USER, cfg.SVSIGNAL.MYSQL.PASSWORD, cfg.SVSIGNAL.MYSQL.HOST, cfg.SVSIGNAL.MYSQL.PORT, cfg.SVSIGNAL.MYSQL.DATABASE)
		fmt.Println(uri)
		var err error

		db, err := sql.Open("mysql", uri)

		if err != nil {
			log.Println("Error Open DB", err)
		}

		// See "Important settings" section.
		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)

		err = db.Ping()
		if err != nil {
			log.Panicln("connect to DB", err)
		}

		DB, err = gorm.Open(mysql.New(mysql.Config{
			Conn: db,
		}), &gorm.Config{})
		if err != nil {
			log.Panicln("connect to DB", err)
		}
	}

	if cfg.SVSIGNAL.SQLite != nil {
		db, err := gorm.Open(sqlite.Open(cfg.SVSIGNAL.SQLite.FILE), &gorm.Config{})
		if err != nil {
			log.Panicf("failed to connect DB; err:%s", err)
		}
		DB = db
		for _, exec := range cfg.SVSIGNAL.SQLite.PRAGMA {
			db.Exec(exec)
		}
	}
}

func main() {

	var wg sync.WaitGroup
	ctx := context.Background()

	CH_SAVE_VALUE = make(chan ValueSignal, 1)
	CH_SET_SIGNAL = make(chan SetSignal, 1)
	CH_MSG_AMPQ = make(chan gormq.MessageAmpq, 1)
	CH_REQUEST_DATA = make(chan RequestData, 1)

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
		log.Panicln(err)
	}
	fmt.Printf("%+v\n", cfg)

	connectDataBase()
	if cfg.SVSIGNAL.DEBUG_LEVEL >= 6 {
		DB = DB.Debug()
	}
	sqlDB, err := DB.DB()
	if err != nil {
		log.Panicln("DB err:", err)
	}
	defer sqlDB.Close()

	// migrate DB
	model.Migrate(DB)

	savesignal = newSVS()
	savesignal.Load()
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.Run(&wg, ctx_db)

	conn_rmq, err := gormq.NewConnect(cfg.SVSIGNAL.RABBITMQ.URL)
	if err != nil {
		log.Panicln("connect rabbitmq", err)
	}

	hub = newHub()
	//hub.CH_REQUEST_HTTP_DB = savesignal.CH_REQUEST_HTTP
	hub.debug_level = cfg.SVSIGNAL.DEBUG_LEVEL
	ctx_hub, cancel_hub := context.WithCancel(ctx)
	wg.Add(1)
	go hub.run(&wg, ctx_hub)

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
		CH_MSG_AMPQ,
	)
	if err != nil {
		log.Panicln("consumer error\n", err)
	}

	//---http
	/*http := HttpSrv{
		Addr:       cfg.SVSIGNAL.HTTP.Address,
		UnixSocket: cfg.SVSIGNAL.HTTP.UnixSocket,
		svsignal:   savesignal,
	}
	err = http.initUnixSocketServer()
	if err != nil {
		panic(err)
	}
	http.svsignal = savesignal
	wg.Add(1)
	go http.Run(&wg)*/
	//-------

	// GRPC
	grpc_srv := gRPCServer{Addr: cfg.SVSIGNAL.GRPCServer.Address}
	wg.Add(1)
	go grpc_srv.run(&wg)

	f_shutdown := func(ctx context.Context) {
		fmt.Println("ShutDown")
		grpc_srv.Close()
		// close rabbitmq
		conn_rmq.Close()
		chCons.Close()

		cancel_hub()
		cancel_db()
		//http.Close()
	}
	wg.Add(1)
	go goutils.WaitSignalExit(&wg, ctx, f_shutdown)
	// ждем освобождение горутин
	wg.Wait()
	fmt.Println("End")
}
