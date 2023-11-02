package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lasaleks/svsignal/model"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	source      = flag.String("source", "apache2:apache2data@tcp(mysql:3306)/insiteexpert_v4?charset=utf8&parseTime=True&loc=Local", "")
	source_type = flag.String("source-type", "mysql_data_v1", "mysql/mysql_data_v1/sqlite")
	dest_type   = flag.String("dest-type", "sqlite", "mysql/sqlite")
	dest        = flag.String("dest", "svsignal.db", "")
	get_version = flag.Bool("version", false, "version")
)

var new_logger = logger.New(
	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	logger.Config{
		SlowThreshold:             time.Second,   // Slow SQL threshold
		LogLevel:                  logger.Silent, // Log level
		IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
		ParameterizedQueries:      true,          // Don't include params in the SQL log
		Colorful:                  false,         // Disable color
	},
)

var config = gorm.Config{
	//PrepareStmt:            true,
	//SkipDefaultTransaction: true,
	Logger: new_logger,
}

var EXECS = []string{
	"PRAGMA journal_mode = WAL",
	"PRAGMA synchronous = OFF",
}

var VERSION string
var BUILD string

func main() {
	flag.Parse()
	fmt.Println(VERSION+" build:", BUILD)
	if *get_version {
		return
	}

	var source_db *gorm.DB
	var dest_db *gorm.DB
	var err error

	switch *source_type {
	case "sqlite":
		source_db, err = gorm.Open(sqlite.Open(*source), &config)
		if err != nil {
			panic("failed to connect source database")
		}
		for _, exec := range EXECS {
			fmt.Println(exec)
			source_db.Exec(exec)
		}
	case "mysql":
		source_db, err = gorm.Open(mysql.Open(*source), &config)
		if err != nil {
			log.Panicln("failed to connect database", *source)
		}
	case "mysql_data_v1":
		source_db, err = gorm.Open(mysql.Open(*source), &config)
		if err != nil {
			log.Panicln("failed to connect database", err)
		}
	default:
		log.Panicln("typeDB error value:", *source_type)
	}

	switch *dest_type {
	case "sqlite":
		dest_db, err = gorm.Open(sqlite.Open(*dest), &config)
		if err != nil {
			log.Panicln("failed to connect dest database err:", err)
		}
		for _, exec := range EXECS {
			fmt.Println(exec)
			dest_db.Exec(exec)
		}
	case "mysql":
		dest_db, err = gorm.Open(mysql.Open(*dest), &config)
		if err != nil {
			log.Panicln("failed to connect database", *source)
		}
	default:
		log.Panicln("typeDB error value:", *source_type)
	}

	db, err := source_db.DB()
	if err != nil {
		log.Panicln(err)
	}

	model.Migrate(dest_db)

	rows, err := db.Query("SELECT id, group_key, name FROM svsignal_group")
	if err != nil {
		log.Panicln(err)
	}
	defer rows.Close()
	for rows.Next() {
		group := model.Group{}
		err := rows.Scan(&group.ID, &group.Key, &group.Name)
		if err != nil {
			log.Panicln(err)
		}
		res := dest_db.Create(&group)
		if res.Error != nil {
			log.Println(res.Error)
		}
	}

	rows, err = db.Query("SELECT `id`, `group_id`, `signal_key`, `name`, `type_save`, `period`, `delta` FROM `svsignal_signal`")
	if err != nil {
		log.Panicln(err)
	}

	for rows.Next() {
		sig := model.Signal{}
		err := rows.Scan(&sig.ID, &sig.GroupID, &sig.Key, &sig.Name, &sig.TypeSave, &sig.Period, &sig.Delta)
		if err != nil {
			log.Panicln(err)
		}
		res := dest_db.Create(&sig)
		if res.Error != nil {
			log.Println(res.Error)
		}
	}
	rows.Close()

	rows, err = db.Query("SELECT `id`, `signal_id`, `tag`, `value` FROM `svsignal_tag`")
	if err != nil {
		log.Panicln(err)
	}
	for rows.Next() {
		tag := model.Tag{}
		err := rows.Scan(&tag.ID, &tag.SignalID, &tag.Tag, &tag.Value)
		if err != nil {
			log.Panicln(err)
		}
		res := dest_db.Create(&tag)
		if res.Error != nil {
			log.Println(res.Error)
		}
	}
	rows.Close()

	limit := 1000
	begin := 0

	fvalues := make([]model.FValue, limit)
	values_cnt := 0
	min_id := 0
	max_id := 0
	row := db.QueryRow("SELECT min(id), max(id) from svsignal_fvalue")
	if row.Err() == nil {
		row.Scan(&min_id, &max_id)
		begin = min_id
	}

	for begin < max_id {
		fmt.Printf("<query svsignal_fvalue id>=%d and id <%d", begin, begin+limit)
		rows, err = db.Query("SELECT `id`, `signal_id`, `utime`, `value`, `offline` FROM `svsignal_fvalue` where id>=? and id<? order by id", begin, begin+limit)
		if err != nil {
			log.Panicln(err)
		}
		for rows.Next() {
			val := model.FValue{}
			offline := 0
			err := rows.Scan(&val.ID, &val.SignalID, &val.UTime, &val.Value, &offline)
			if err != nil {
				log.Println(err)
			}
			if offline == 1 {
				val.OffLine = true
			}
			fvalues[values_cnt] = val
			values_cnt++
		}
		fmt.Printf(">\n")
		rows.Close()
		fmt.Printf("<insert rows:%d", values_cnt)
		tx := dest_db.Begin()
		res := tx.Create(fvalues[:values_cnt])
		if res.Error != nil {
			log.Println(res.Error)
		}
		tx.Commit()
		fmt.Printf(">\n")
		/*if values_cnt < limit {
			break
		}*/
		values_cnt = 0
		begin = begin + limit
	}

	limit = 1000
	begin = 0

	ivalues := make([]model.IValue, limit)
	values_cnt = 0
	min_id = 0
	max_id = 0
	row = db.QueryRow("SELECT min(id), max(id) from svsignal_ivalue")
	if row.Err() == nil {
		row.Scan(&min_id, &max_id)
		begin = min_id
	}

	for begin < max_id {
		fmt.Printf("<query svsignal_ivalue id>=%d and id <%d", begin, begin+limit)
		rows, err = db.Query("SELECT `id`, `signal_id`, `utime`, `value`, `offline` FROM `svsignal_ivalue` where id>=? and id<? order by id", begin, begin+limit)
		if err != nil {
			log.Panicln(err)
		}
		for rows.Next() {
			val := model.IValue{}
			offline := 0
			err := rows.Scan(&val.ID, &val.SignalID, &val.UTime, &val.Value, &offline)
			if err != nil {
				log.Println(err)
			}
			if offline == 1 {
				val.OffLine = true
			}
			ivalues[values_cnt] = val
			values_cnt++
		}
		fmt.Printf(">\n")
		rows.Close()
		fmt.Printf("<insert rows:%d", values_cnt)
		tx := dest_db.Begin()
		res := tx.Create(ivalues[:values_cnt])
		if res.Error != nil {
			log.Println(res.Error)
		}
		tx.Commit()
		fmt.Printf(">\n")
		/*if values_cnt < limit {
			break
		}*/
		values_cnt = 0
		begin = begin + limit
	}

}
