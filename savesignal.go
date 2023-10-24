package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lasaleks/svsignal/model"
)

// Сбор статистики работы сервиса
var SrvStatus struct {
	ValuesInBuffer      int `json:"ValuesInBuffer"`      // кол-во значений в буффере
	RecvedValues        int `json:"RecvedValues"`        // кол-во полученных сохранений значений сигнала
	NumberOfWriteValues int `json:"NumberOfWriteValues"` // кол-во сохраненных событий
	//
	HeapInuse    uint64 // количество байт, которые программа аллоцировала в динамической памяти
	StackInuse   uint64 // количество памяти, которое находится на стеке
	NumGoroutine int    //
}

const (
	_ = iota
	TYPE_IVALUE
	TYPE_FVALUE
	TYPE_MVALUE
)

type SValueInt struct {
	value   int64
	offline int
	utime   int64
}

type SValueFloat struct {
	value   float64
	offline int
	utime   int64
}

type SVSignalDB struct {
	db              *sql.DB // !!!! remove
	group_id        map[uint]*model.Group
	group_key       map[string]*model.Group
	signal_key      map[string]*model.Signal
	svalueint       map[string]*SValueInt
	svalueavg       map[string]*AVG
	CH_REQUEST_HTTP chan interface{}

	// cache insert
	bulk_insert_buffer_size int
	buffer_size             int

	buffer_write_i map[int64]*[]SValueInt
	size_buffer_i  int
	buffer_write_f map[int64]*[]SValueFloat
	size_buffer_f  int
	//
	lt_save     int64 //  время сохранения
	period_save int64
}

func newSVS() *SVSignalDB {
	return &SVSignalDB{
		group_id:        make(map[uint]*model.Group),
		group_key:       make(map[string]*model.Group),
		CH_REQUEST_HTTP: make(chan interface{}, 1),
		svalueint:       make(map[string]*SValueInt),
		svalueavg:       make(map[string]*AVG),
		buffer_write_i:  make(map[int64]*[]SValueInt),
		buffer_write_f:  make(map[int64]*[]SValueFloat),
		lt_save:         time.Now().Unix(),
	}
}

func (s *SVSignalDB) Run(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	defer func() {
		write_buffer_i(s.db, s.buffer_write_i, s.bulk_insert_buffer_size)
		write_buffer_f(s.db, s.buffer_write_f, s.bulk_insert_buffer_size)
		if cfg.SVSIGNAL.DEBUG_LEVEL >= 1 {
			log.Println("SaveSignal End")
		}
	}()

	var groups []model.Group
	// works because destination struct is passed in
	if result := DB.Find(&groups); result.Error != nil {
		// error load svsignal_systems
		log.Panicln(result.Statement.SQL, result.Error)
	}
	for _, group := range groups {
		s.group_key[group.Key] = &group
		s.group_id[group.ID] = &group
	}

	var signals []model.Signal
	//db.Joins("Company").Find(&users)
	// works because destination struct is passed in
	if result := DB.Joins("tag").Find(&signals); result.Error != nil {
		// error load svsignal_systems
		log.Panicln(result.Statement.SQL, result.Error)
	}
	for _, signal := range signals {
		group, ok := s.group_id[signal.GroupID]
		if !ok {
			continue
		}
		s.signal_key[fmt.Sprintf("%s.%s", group.Key, signal.Key)] = &signal
	}
	/*
		signals, err := load_signals(s.db)
		if err != nil {
			log.Panicf("error load svsignal_signals %v", err)
		}
		s.signals = *signals
	*/
	for {
		select {
		case <-ctx.Done():
			//log.Println("SaveSignal run Done")
			return
		case <-time.After(time.Second * 1):
			utime := time.Now().Unix()
			if s.lt_save+s.period_save < utime {
				write_buffer_i(s.db, s.buffer_write_i, s.bulk_insert_buffer_size)
				write_buffer_f(s.db, s.buffer_write_f, s.bulk_insert_buffer_size)
				s.lt_save = time.Now().Unix()
				s.buffer_write_i = make(map[int64]*[]SValueInt)
				s.buffer_write_f = make(map[int64]*[]SValueFloat)
				s.size_buffer_i = 0
				s.size_buffer_f = 0
				SrvStatus.ValuesInBuffer = 0
			}
			break
		case msg, ok := <-CH_SAVE_VALUE:
			if ok {
				s.save_value(&msg)
				SrvStatus.RecvedValues++
			} else {
				return
			}
		case msg, ok := <-CH_SET_SIGNAL:
			if ok {
				s.set_signal(msg)
			} else {
				return
			}
		case msg, ok := <-s.CH_REQUEST_HTTP:
			if ok {
				switch request := msg.(type) {
				case ReqListSignal:
					request.CH_RR_LIST_SIGNAL <- *s.response_list_signal()
				case ReqSignalData:
					signal_key := fmt.Sprintf("%s.%s", request.groupkey, request.signalkey)
					signal, ok := s.signal_key[signal_key]
					if !ok {
						request.CH_RESPONSE <- fmt.Errorf("error not found %s", signal_key)
					} else {
						name_group := ""
						group, ok := s.group_key[request.groupkey]
						if ok {
							name_group = group.Name
						}
						var value_signal *SValueInt
						if signal.TypeSave == 1 {
							value_signal = s.svalueint[signal_key]
							if value_signal != nil {
								value_signal = &SValueInt{value: value_signal.value, utime: value_signal.utime, offline: value_signal.offline}
							}
						}

						s.get_data_signal(s.db, request.CH_RESPONSE, name_group, signal, value_signal, request.begin, request.end, int(signal.TypeSave))
					}
				}
			}
		}
	}
}

func (s *SVSignalDB) response_list_signal() *ResponseListSignal {
	lsig := ResponseListSignal{Groups: make(map[string]*RLS_Groups)}
	for _, group := range s.group_id {
		lsig.Groups[group.Key] = &RLS_Groups{Name: group.Name, Signals: make(map[string]RLS_Signal)}
	}
	// сортируем ключи
	keys := make([]string, 0, len(s.signal_key))
	for k := range s.signal_key {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		signal := s.signal_key[key]
		group := s.group_id[signal.GroupID]
		_, ok := lsig.Groups[group.Key]
		if !ok {
			lsig.Groups[group.Key] = &RLS_Groups{}
		}
		tags := []RLS_Tag{}
		if signal.Tags != nil {
			for _, tag := range signal.Tags {
				tags = append(tags, RLS_Tag{Tag: tag.Tag, Value: tag.Value})
			}
		}
		lsig.Groups[group.Key].Signals[key] = RLS_Signal{
			Name:     signal.Name,
			TypeSave: int(signal.TypeSave),
			Period:   signal.Period,
			Delta:    signal.Delta,
			Tags:     tags,
		}
	}
	return &lsig
}

func (s *SVSignalDB) save_value(val *ValueSignal) {
	sig_key := fmt.Sprintf("%s.%s", val.group_key, val.signal_key)
	if DEBUG_LEVEL >= 8 {
		log.Printf("save_value %s %+v", sig_key, val)
	}

	if cfg.SVSIGNAL.DEBUG_LEVEL >= 6 {
		log.Printf("SaveSignal:%s value:%v offline:%d utime:%d\n", sig_key, val.Value, val.Offline, val.UTime)
	}

	signal, ok := s.signal_key[sig_key]
	if !ok {
		if cfg.SVSIGNAL.DEBUG_LEVEL >= 1 {
			fmt.Printf("not found signal, key:%s", sig_key)
		}
		return
	}
	switch signal.TypeSave {
	case 1:
		valuei := int64(val.Value)
		offline := int(val.Offline)
		is_save := false
		pvalue, ok := s.svalueint[sig_key]
		if !ok {
			is_save = true
			pvalue = &SValueInt{value: valuei, offline: offline, utime: val.UTime}
			s.svalueint[sig_key] = pvalue
		}
		if pvalue.offline != offline {
			is_save = true
		}
		if pvalue.value != valuei {
			is_save = true
		}
		if is_save {
			s.insert_valuei(s.db, int64(signal.ID), valuei, val.UTime, val.Offline, cfg.SVSIGNAL.DEBUG_LEVEL >= 6)
		}
		pvalue.value = valuei
		pvalue.offline = offline
		pvalue.utime = val.UTime
		if DEBUG_LEVEL >= 8 {
			log.Printf("save_value %s %+v pvalue:%+v", sig_key, val, pvalue)
		}
	case 2:
		avg, ok := s.svalueavg[sig_key]
		if !ok {
			avg = newAVG(signal.Period, signal.Delta)
			s.svalueavg[sig_key] = avg
		}
		err, l_values := avg.add_value(SValueFloat{value: val.Value, utime: val.UTime, offline: int(val.Offline)})
		if err != nil {
			if DEBUG_LEVEL >= 1 {
				log.Printf("error add_value %+v %s", val, err)
			}
			return
		}
		for _, value := range l_values {
			//fmt.Printf("insert_valuef: %+v\n", value)
			s.insert_valuef(s.db, int64(signal.ID), value.value, value.utime, int64(value.offline))
		}
	case 3:
		break
	}
}

/*
func (s *SVSignalDB) save_value_avg(sig_key *string, signal *svsignal_signal, val *ValueSignal) {
	is_save := false
	avg, ok := s.svalueavg[*sig_key]
	if !ok {
		avg = newAVG(signal.period, signal.delta)
		s.svalueavg[*sig_key] = avg
	}
	reason_save := 0
	if avg.is_end_period(val.UTime) {
		is_save = true
		reason_save = 1
	} else if avg.is_delta_prev(val.Value) {
		is_save = true
		reason_save = 2
	} else if avg.is_delta_avg(val.Value) {
		reason_save = 3
	}

	if is_save {
		err, calc_avg, value_avg, utime_avg := avg.calc_avg()

		if err != nil {
		} else {
			switch reason_save {
			case 2:
				if calc_avg {
					s.insert_valuef(s.db, signal.id, value_avg, utime_avg, 0)
					s.insert_valuef(s.db, signal.id, avg.p_value, avg.p_utime, 0)
				}
				s.insert_valuef(s.db, signal.id, val.Value, val.UTime, 0)
			case 3:

			default:
				s.insert_valuef(s.db, signal.id, value_avg, utime_avg, 0)
			}
		}
		avg.set_new_period()
	}
	avg.add(val.Value, val.UTime)
}*/

func setupBindVars(stmt, bindVars string, len int) string {
	bindVars += ","
	stmt = fmt.Sprintf(stmt, strings.Repeat(bindVars, len))
	return strings.TrimSuffix(stmt, ",")
}

func insert_values(db *sql.DB, sql string, values []interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	result, err := tx.Exec(sql, values...)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	last_id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	/*if affected != int64(len(vals)) {
	}*/
	if DEBUG_LEVEL >= 4 {
		log.Println("Insert - RowsAffected", affected, "LastInsertId: ", last_id)
	}
	return nil
}

func write_buffer_i(db *sql.DB, buffer_i map[int64]*[]SValueInt, max_multiply_insert int) error {
	values := make([]interface{}, max_multiply_insert*4)
	len_vals := 0
	rows := 0
	stmt := "INSERT INTO svsignal_ivalue(signal_id, value, utime, offline) VALUES %s"
	for signal_id := range buffer_i {
		vals := (buffer_i)[signal_id]
		for _, val := range *vals {
			if rows >= max_multiply_insert {
				sql := setupBindVars(stmt, "(?,?,?,?)", rows)
				err := insert_values(db, sql, values[0:len_vals])
				if err != nil {
					fmt.Printf("error insert rows:%d, %s", rows, err)
				}
				len_vals = 0
				rows = 0
			}
			values[len_vals] = signal_id
			len_vals++
			values[len_vals] = val.value
			len_vals++
			values[len_vals] = val.utime
			len_vals++
			values[len_vals] = int(val.offline)
			len_vals++
			rows++
			SrvStatus.NumberOfWriteValues++
		}
	}

	if len_vals > 0 {
		sql := setupBindVars(stmt, "(?,?,?,?)", rows)
		err := insert_values(db, sql, values[0:len_vals])
		if err != nil {
			fmt.Printf("error insert rows:%d, %s", rows, err)
		}
		len_vals = 0
		rows = 0
	}

	return nil
}

func write_buffer_f(db *sql.DB, buffer_f map[int64]*[]SValueFloat, max_multiply_insert int) error {
	values := make([]interface{}, max_multiply_insert*4)
	len_vals := 0
	rows := 0
	stmt := "INSERT INTO svsignal_fvalue(signal_id, value, utime, offline) VALUES %s"
	for signal_id := range buffer_f {
		vals := (buffer_f)[signal_id]
		for _, val := range *vals {
			if len_vals/4 >= max_multiply_insert {
				if len_vals > 0 {
					sql := setupBindVars(stmt, "(?,?,?,?)", rows)
					err := insert_values(db, sql, values[0:len_vals])
					if err != nil {
						fmt.Printf("error insert rows:%d, %s", rows, err)
					}
					len_vals = 0
					rows = 0
				}
			}
			values[len_vals] = signal_id
			len_vals++
			values[len_vals] = val.value
			len_vals++
			values[len_vals] = val.utime
			len_vals++
			values[len_vals] = int(val.offline)
			len_vals++
			rows++
			SrvStatus.NumberOfWriteValues++
		}
	}

	if len_vals > 0 {
		sql := setupBindVars(stmt, "(?,?,?,?)", rows)
		err := insert_values(db, sql, values[0:len_vals])
		if err != nil {
			fmt.Printf("error insert rows:%d, %s", rows, err)
		}
		len_vals = 0
		rows = 0
	}

	return nil
}

func (s *SVSignalDB) insert_valuei(db *sql.DB, signal_id int64, value int64, utime int64, offline int64, debug bool) error {
	if cfg.SVSIGNAL.DEBUG_LEVEL >= 6 {
		fmt.Println("insert_valuei", signal_id, value, utime, offline)
	}
	cache, ok := s.buffer_write_i[signal_id]
	if !ok {
		cache = &[]SValueInt{{value: value, utime: utime, offline: int(offline)}}
		s.buffer_write_i[signal_id] = cache
	} else {
		*cache = append(*cache, SValueInt{value: value, utime: utime, offline: int(offline)})
	}
	s.size_buffer_i++

	if s.buffer_size <= (s.size_buffer_i+s.size_buffer_f) && s.size_buffer_i >= s.bulk_insert_buffer_size {
		buff := s.buffer_write_i
		write_buffer_i(s.db, buff, s.bulk_insert_buffer_size)
		s.buffer_write_i = make(map[int64]*[]SValueInt)
		s.size_buffer_i = 0
	}

	SrvStatus.ValuesInBuffer = s.size_buffer_i + s.size_buffer_f
	return nil
}

func (s *SVSignalDB) insert_valuef(db *sql.DB, signal_id int64, value float64, utime int64, offline int64) error {
	if DEBUG_LEVEL >= 6 {
		fmt.Println("insert_valuef", signal_id, value, utime, offline)
	}
	cache, ok := s.buffer_write_f[signal_id]
	if !ok {
		cache = &[]SValueFloat{
			{value: value, utime: utime, offline: int(offline)},
		}
		s.buffer_write_f[signal_id] = cache
	} else {
		*cache = append(*cache, SValueFloat{value: value, utime: utime, offline: int(offline)})
	}
	s.size_buffer_f++

	if s.buffer_size <= (s.size_buffer_i+s.size_buffer_f) && s.size_buffer_f >= s.bulk_insert_buffer_size {
		buff := s.buffer_write_f
		write_buffer_f(s.db, buff, s.bulk_insert_buffer_size)
		s.buffer_write_f = make(map[int64]*[]SValueFloat)
		s.size_buffer_f = 0
	}

	SrvStatus.ValuesInBuffer = s.size_buffer_i + s.size_buffer_f
	return nil
}

func create_new_group(db *sql.DB, system_key string) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	str_sql := "INSERT INTO svsignal_group(group_key, name) VALUES (?, ?)"
	var result sql.Result
	if result, err = tx.Exec(str_sql, system_key, ""); err != nil {
		fmt.Println("Error", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func create_new_signal(db *sql.DB, group_id int64, signal_key string, name string, type_save int, period int, delta float32) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	str_sql := "INSERT INTO svsignal_signal(group_id, signal_key, name, type_save, period, delta) VALUES (?,?,?,?,?,?)"
	var result sql.Result
	if result, err = tx.Exec(str_sql, group_id, signal_key, name, type_save, period, delta); err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func update_signal(db *sql.DB, signal_id int64, name string, type_save int, period int, delta float32) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	sql := "UPDATE svsignal_signal SET name=?, type_save=?, period=?, delta=? WHERE id=?"
	if _, err := tx.Exec(sql, name, type_save, period, delta, signal_id); err != nil {
		return err
	}
	return nil
}

type svsignal_group struct {
	id        int64
	name      string
	group_key string
}

func load_groups(db *sql.DB) (*map[string]svsignal_group, error) {
	// Prepare statement for reading data
	rows, err := db.Query("SELECT id, group_key, name FROM svsignal_group")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make(map[string]svsignal_group)
	for rows.Next() {
		g := svsignal_group{}
		err := rows.Scan(&g.id, &g.group_key, &g.name)
		if err != nil {
			fmt.Println(err)
			continue
		}
		groups[g.group_key] = g
	}
	return &groups, nil
}

type svsignal_signal struct {
	id         int64
	group_id   int64
	group_key  string
	signal_key string
	name       string
	type_save  int
	period     int
	delta      float32
	tags       *[]svsignal_tag
}

func load_signals(db *sql.DB) (*map[string]*svsignal_signal, error) {

	tags, err := load_signal_tags(db)
	if err != nil {
		tags = nil
	}
	// Prepare statement for reading data

	rows, err := db.Query("SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	signals := make(map[string]*svsignal_signal)
	for rows.Next() {
		sig := svsignal_signal{}
		err := rows.Scan(&sig.id, &sig.group_id, &sig.group_key, &sig.signal_key, &sig.name, &sig.type_save, &sig.period, &sig.delta)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if tags != nil {
			tag, ok := (*tags)[sig.id]
			if ok {
				sig.tags = tag
			}
		}
		signals[fmt.Sprintf("%s.%s", sig.group_key, sig.signal_key)] = &sig
	}
	return &signals, nil
}

type svsignal_tag struct {
	id        int64
	signal_id int64
	tag       string
	value     string
}

func load_signal_tags(db *sql.DB) (*map[int64]*[]svsignal_tag, error) {
	// Prepare statement for reading data
	rows, err := db.Query("SELECT id, signal_id, tag, `value` FROM svsignal_tag")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make(map[int64]*[]svsignal_tag)
	for rows.Next() {
		tag := svsignal_tag{}
		err := rows.Scan(&tag.id, &tag.signal_id, &tag.tag, &tag.value)
		if err != nil {
			fmt.Println(err)
			continue
		}
		_, ok := tags[tag.signal_id]
		if !ok {
			tags[tag.signal_id] = &[]svsignal_tag{}
		}
		*tags[tag.signal_id] = append(*tags[tag.signal_id], tag)
	}
	return &tags, nil
}

func (s *SVSignalDB) get_data_signal_i(signal *model.Signal, value_signal_i *SValueInt, begin int64, end int64) (*[][4]int64, error) {
	ivalues := &[][4]int64{}

	// определяем наличие данных в буффере на вставку
	var firts_utime_in_cache *int64
	var last_utime_in_cache *int64
	cache, ok := s.buffer_write_i[int64(signal.ID)]
	if ok {
		len_cache := len(*cache)
		if len_cache > 0 {
			firts_utime_in_cache = &(*cache)[0].utime
		}
		if len_cache > 1 {
			last_utime_in_cache = &(*cache)[len_cache-1].utime
		}
	}

	add_data_from_buff := false
	add_end_point := true
	query_db_begin := true
	query_db := true
	query_db_end := true

	if firts_utime_in_cache != nil && last_utime_in_cache != nil {
		// читаем данные только из буффера
		if begin >= *firts_utime_in_cache && begin < *last_utime_in_cache { // 1. begin >= first && begin < end
			add_data_from_buff = true
			query_db = false
			query_db_end = false
			if begin != *firts_utime_in_cache {
				query_db_begin = false
			}
		} else if begin < *firts_utime_in_cache && end > *firts_utime_in_cache { // 2. begin < first && end > first
			// читаем данные из БД и дополняем данными из буффера
			add_data_from_buff = true
			if end < *last_utime_in_cache {
				query_db_end = false
			}

		} else if begin < *firts_utime_in_cache && end < *firts_utime_in_cache { // 3. begin < first && end < first
			// читаем данные только из БД
			add_data_from_buff = false
		} else if begin > *firts_utime_in_cache && begin > *last_utime_in_cache { // 4. begin > firts && begin > last
			add_data_from_buff = false
		}
	}

	if query_db_begin {
		// запрос значения до заданного периода begin-end
		//IValues := []model.IValue{}
		//DB.Where("id = ?", signal.ID).Select().Find(&IValues)
		//DB.Find(&ivalues, "id = ? and id=(select max(id) from svsignal_ivalue where utime<%d and signal_id=%d)", "jinzhu", 20)
		row := s.db.QueryRow(fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_ivalue WHERE signal_id=%d and id=(select max(id) from svsignal_ivalue where utime<%d and signal_id=%d)", signal.ID, begin, signal.ID))
		if row != nil {
			var ivalue [4]int64
			err := row.Scan(&ivalue[0], &ivalue[1], &ivalue[2], &ivalue[3])
			if err == nil {
				*ivalues = append(*ivalues, ivalue)
			}
		}
	}

	if query_db {
		// запрашиваем данные за период из бд
		rows, err := s.db.Query(fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_ivalue WHERE signal_id=%d and utime >= %d and utime <=%d", signal.ID, begin, end))
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var ivalue [4]int64
			var err error
			err = rows.Scan(&ivalue[0], &ivalue[1], &ivalue[2], &ivalue[3])
			if err != nil {
				fmt.Println(err)
				continue
			}
			*ivalues = append(*ivalues, ivalue)
		}

		if query_db_end {
			// запрос значения после заданного периода begin-end
			row := s.db.QueryRow(fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_ivalue WHERE signal_id=%d and id=(select max(id) from svsignal_ivalue where utime>%d and signal_id=%d)", signal.ID, end, signal.ID))
			if row != nil {
				var ivalue [4]int64
				err := row.Scan(&ivalue[0], &ivalue[1], &ivalue[2], &ivalue[3])
				if err == nil {
					*ivalues = append(*ivalues, ivalue)
				} else {
					if !add_data_from_buff && firts_utime_in_cache != nil {
						value := (*cache)[0]
						*ivalues = append(*ivalues, [4]int64{0, value.utime, value.value, int64(value.offline)})
					}
				}
			}
		}
	}

	if add_data_from_buff {
		var prev_row *SValueInt
		add_first := false
		for idx, value := range *cache {
			if begin <= value.utime {
				if end >= value.utime {
					if !query_db_begin && !add_first && prev_row != nil {
						*ivalues = append(*ivalues, [4]int64{0, prev_row.utime, prev_row.value, int64(prev_row.offline)})
						add_first = true
					}
					*ivalues = append(*ivalues, [4]int64{0, value.utime, value.value, int64(value.offline)})
				} else {
					add_end_point = false
					*ivalues = append(*ivalues, [4]int64{0, value.utime, value.value, int64(value.offline)})
					break
				}
			}
			prev_row = &(*cache)[idx]
		}
	}

	// задания дополнительной точки с временем равным полученному последнему значению.
	if add_end_point && value_signal_i != nil {
		len := len(*ivalues)
		if len > 0 {
			len--
			if (*ivalues)[len][1] < value_signal_i.utime && (*ivalues)[len][2] == value_signal_i.value && (*ivalues)[len][3] == int64(value_signal_i.offline) {
				*ivalues = append(*ivalues, (*ivalues)[len])
				(*ivalues)[len+1][1] = value_signal_i.utime
			}
		}
	}
	return ivalues, nil
}

func (s *SVSignalDB) get_data_signal_f(signal *model.Signal, begin int64, end int64) (*[][4]interface{}, error) {
	fvalues := &[][4]interface{}{}

	// определяем наличие данных в буффере на вставку
	var firts_utime_in_cache *int64
	var last_utime_in_cache *int64
	cache, ok := s.buffer_write_f[int64(signal.ID)]
	if ok {
		len_cache := len(*cache)
		if len_cache > 0 {
			firts_utime_in_cache = &(*cache)[0].utime
		}
		if len_cache > 1 {
			last_utime_in_cache = &(*cache)[len_cache-1].utime
		}
	}

	add_data_from_buff := false
	query_db := true

	if firts_utime_in_cache != nil && last_utime_in_cache != nil {
		// читаем данные только из буффера
		if begin >= *firts_utime_in_cache && begin < *last_utime_in_cache { // 1. begin >= first && begin < end
			add_data_from_buff = true
			query_db = false
		} else if begin < *firts_utime_in_cache && end > *firts_utime_in_cache { // 2. begin < first && end > first
			// читаем данные из БД и дополняем данными из буффера
			add_data_from_buff = true
		} else if begin < *firts_utime_in_cache && end < *firts_utime_in_cache { // 3. begin < first && end < first
			// читаем данные только из БД
			add_data_from_buff = false
		} else if begin > *firts_utime_in_cache && begin > *last_utime_in_cache { // 4. begin > firts && begin > last
			add_data_from_buff = false
		}
	}

	if query_db {
		// запрашиваем данные за период из бд
		rows, err := s.db.Query(fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_fvalue WHERE signal_id=%d and utime >= %d and utime <=%d", signal.ID, begin, end))
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {

			var id, utime, offline int64
			var value float64
			err := rows.Scan(&id, &utime, &value, &offline)
			if err != nil {
				fmt.Println(err)
				continue
			}
			*fvalues = append(*fvalues, [4]interface{}{id, utime, value, offline})
		}
	}

	if add_data_from_buff {
		for _, value := range *cache {
			if begin <= value.utime {
				if end >= value.utime {
					*fvalues = append(*fvalues, [4]interface{}{0, value.utime, value.value, value.offline})
				} else {
					break
				}
			}
		}
	}

	return fvalues, nil
}

func (s *SVSignalDB) get_data_signal(db *sql.DB, out chan interface{}, name_group string, signal *model.Signal, value_signal_i *SValueInt, begin int64, end int64, type_table int) {
	if signal == nil {
		out <- fmt.Errorf("error get_data_signal signal:nil")
		return
	}
	tags := []RLS_Tag{}
	for _, tag := range signal.Tags {
		tags = append(tags, RLS_Tag{Tag: tag.Tag, Value: tag.Value})
	}

	switch type_table {
	case TYPE_IVALUE:
		ivalues, err := s.get_data_signal_i(signal, value_signal_i, begin, end)
		if err != nil {
			out <- fmt.Errorf("error key:%s %s", signal.Key, err)
		} else {
			group := s.group_id[signal.GroupID]
			out <- ResponseDataSignalT1{
				GroupKey:   group.Key,
				GroupName:  name_group,
				SignalKey:  signal.Key,
				SignalName: signal.Name,
				TypeSave:   type_table,
				Values:     *ivalues,
				Tags:       tags,
			}
		}
	case TYPE_FVALUE:
		fvalues, err := s.get_data_signal_f(signal, begin, end)
		if err != nil {
			out <- fmt.Errorf("error key:%s %s", signal.Key, err)
		} else {
			group := s.group_id[signal.GroupID]
			out <- ResponseDataSignalT2{
				GroupKey:   group.Key,
				GroupName:  name_group,
				SignalKey:  signal.Key,
				SignalName: signal.Name,
				TypeSave:   type_table,
				Values:     *fvalues,
				Tags:       tags,
			}
		}
	case TYPE_MVALUE:
	default:
		out <- fmt.Errorf("error request data signal; type not found %d", type_table)
		return
	}
}

func (s *SVSignalDB) set_signal(setsig SetSignal) {
	/*
		group, ok := s.groups[setsig.group_key]
		if !ok {
			// create systems
			id, err := create_new_group(s.db, setsig.group_key)
			if err != nil {
				log.Printf("error create new system: group_key:%s; error:%v", setsig.group_key, err)
			} else {
				//fmt.Println("create systems", val.system_key, "Ok")
				group = svsignal_group{id: id, group_key: setsig.group_key, name: ""}
				s.groups[setsig.group_key] = group
			}
		}
		sig_key := fmt.Sprintf("%s.%s", setsig.group_key, setsig.signal_key)
		signal, ok := s.signal_key[sig_key]
		if !ok {
			// create signals
			//fmt.Println("create signal", val.signal_key)
			id, err := create_new_signal(s.db, group.id, setsig.signal_key, setsig.Name, setsig.TypeSave, setsig.Period, setsig.Delta)
			if err != nil {
				log.Println("Error create signal", setsig, err)
				return
			}
			//log.Println("create signal", val, id, "OK")
			signal = &svsignal_signal{
				id:         id,
				group_key:  setsig.group_key,
				signal_key: setsig.signal_key,
				name:       setsig.Name,
				type_save:  setsig.TypeSave,
				period:     setsig.Period,
				delta:      setsig.Delta,
			}
			s.signal_key[sig_key] = signal
		} else {
			fmt.Printf("update signal %++v\n", setsig)
			if signal.name != setsig.Name || signal.type_save != setsig.TypeSave || signal.period != setsig.Period || signal.delta != setsig.Delta {
				err := update_signal(s.db, signal.id, setsig.Name, setsig.TypeSave, setsig.Period, setsig.Delta)
				if err != nil {
					log.Println("Error update signal", setsig, err)
					return
				}
				signal.name = setsig.Name
				signal.type_save = setsig.TypeSave
				signal.period = setsig.Period
				signal.delta = setsig.Delta
			}
		}
		if signal.tags == nil {
			signal.tags = &[]svsignal_tag{}
		}
		not_remove := []int64{}
		for _, utag := range setsig.Tags {
			create := true
			for i := 0; i < len(*signal.tags); i++ {
				tag := &(*signal.tags)[i]
				if tag.tag == utag.Tag {
					not_remove = append(not_remove, tag.id)
					create = false
					if tag.value != utag.Value {
						// update
						err := update_tag(s.db, tag.id, utag.Tag, utag.Value)
						if err != nil {
							fmt.Println(err)
						} else {
							tag.tag = utag.Tag
							tag.value = utag.Value
						}
					}
					break
				}
			}
			if create {
				// create
				if new_tag_id, err := create_tag(s.db, signal.id, utag.Tag, utag.Value); err == nil {
					*(signal.tags) = append(*signal.tags, svsignal_tag{id: new_tag_id, signal_id: signal.id, tag: utag.Tag, value: utag.Value})
					not_remove = append(not_remove, new_tag_id)
				} else {
					fmt.Println("error create_tag", err)
				}
			}
		}
		// удаляем лишние метки
		for i := 0; i < len(*signal.tags); {
			remove := true
			for _, not_remove_id := range not_remove {
				if not_remove_id == (*signal.tags)[i].id {
					remove = false
					break
				}
			}
			if remove {
				*signal.tags = append((*signal.tags)[:i], (*signal.tags)[i+1:]...)
				i = 0
				err := delete_tag(s.db, (*signal.tags)[i].id)
				if err != nil {
					fmt.Println(err)
				}
				continue
			}
			i++
		}*/
}

func create_tag(db *sql.DB, signal_id int64, tag string, value string) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()
	str_sql := "INSERT INTO svsignal_tag(signal_id, tag, value) VALUES (?,?,?)"
	var result sql.Result
	// var err error
	if result, err = tx.Exec(str_sql, signal_id, tag, value); err != nil {
		fmt.Println("Error", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func update_tag(db *sql.DB, tag_id int64, tag string, value string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()
	str_sql := "UPDATE svsignal_tag SET tag=?, value=? WHERE id=?"
	if _, err = tx.Exec(str_sql, tag, value, tag_id); err != nil {
		return err
	}
	return nil
}

func delete_tag(db *sql.DB, tag_id int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()
	str_sql := "delete from svsignal_tag WHERE id=?"
	if _, err = tx.Exec(str_sql, tag_id); err != nil {
		fmt.Println("Error", err)
		return err
	}
	return nil
}
