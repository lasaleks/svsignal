package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

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

type SVSignalDB struct {
	db              *sql.DB
	CH_SAVE_VALUE   chan ValueSignal
	systems         map[string]svsignal_system
	signals         map[string]svsignal_signal
	svalueint       map[string]SValueInt
	svalueavg       map[string]*AVG
	CH_REQUEST_HTTP chan interface{}
}

func newSVS() *SVSignalDB {
	return &SVSignalDB{
		CH_SAVE_VALUE:   make(chan ValueSignal, 1),
		CH_REQUEST_HTTP: make(chan interface{}, 1),
		svalueint:       make(map[string]SValueInt),
		svalueavg:       make(map[string]*AVG),
	}
}

func (s *SVSignalDB) run(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	systems, err := load_system(s.db)
	if err != nil {
		log.Panicf("error load svsignal_systems %v", err)
	}
	signals, err := load_signals(s.db)
	if err != nil {
		log.Panicf("error load svsignal_signals %v", err)
	}
	s.systems = *systems
	s.signals = *signals

	for {
		select {
		case <-ctx.Done():
			log.Println("SaveSignal run Done")
			return
		case msg, ok := <-s.CH_SAVE_VALUE:
			if ok {
				s.save_value(&msg)
			} else {
				return
			}
		case msg, ok := <-s.CH_REQUEST_HTTP:
			if ok {
				switch request := msg.(type) {
				case ReqListSignal:
					request.CH_RR_LIST_SIGNAL <- *s.response_list_signal()
					break
				case ReqSignalData:
					signal_key := fmt.Sprintf("%s.%s", request.groupkey, request.signalkey)
					signal, ok := s.signals[signal_key]
					if !ok {
						request.CH_RESPONSE <- fmt.Errorf("error not found %s", signal_key)
					} else {
						var tags []svsignal_tag
						if signal.tags != nil {
							tags = *signal.tags
						}
						signal.tags = &tags
						name_group := ""
						group, ok := s.systems[request.groupkey]
						if ok {
							name_group = group.name
						}
						request_data_signal(s.db, request.CH_RESPONSE, name_group, signal, request.begin, request.end, signal.type_save)
					}
					break
				}
			}
		}
	}
}

func (s *SVSignalDB) response_list_signal() *ResponseListSignal {
	lsig := ResponseListSignal{Groups: make(map[string]*RLS_Groups)}

	for _, group := range s.systems {
		lsig.Groups[group.system_key] = &RLS_Groups{Name: group.name, Signals: make([]RLS_Signal, 0)}
	}
	// сортируем ключи
	keys := make([]string, 0, len(s.signals))
	for k := range s.signals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		data := s.signals[key]
		_, ok := lsig.Groups[data.system_key]
		if !ok {
			lsig.Groups[data.system_key] = &RLS_Groups{}
		}
		tags := []RLS_Tag{}
		for _, tag := range *data.tags {
			tags = append(tags, RLS_Tag{Tag: tag.tag, Value: tag.value})
		}
		lsig.Groups[data.system_key].Signals = append(lsig.Groups[data.system_key].Signals, RLS_Signal{
			Id:        data.id,
			SignalKey: data.signal_key,
			Name:      data.name,
			TypeSave:  data.type_save,
			Period:    data.period,
			Delta:     data.delta,
			Tags:      tags,
		})
	}
	return &lsig
}

func (s *SVSignalDB) save_value(val *ValueSignal) {
	_, ok := s.systems[val.system_key]
	if !ok {
		// create systems
		err := create_new_system(s.db, val.system_key)
		if err != nil {
			log.Printf("error create new system: system_key:%s; error:%v", val.system_key, err)
		} else {
			//fmt.Println("create systems", val.system_key, "Ok")
			s.systems[val.system_key] = svsignal_system{system_key: val.system_key, name: ""}
		}

	}
	sig_key := fmt.Sprintf("%s.%s", val.system_key, val.signal_key)
	signal, ok := s.signals[sig_key]
	if !ok {
		// create signals
		//fmt.Println("create signal", val.signal_key)
		id, err := create_new_signal(s.db, val.system_key, val.signal_key, val.TypeSave)
		if err != nil {
			log.Println("Error create signal", val, err)
			return
		}
		//log.Println("create signal", val, id, "OK")
		signal = svsignal_signal{id: id, system_key: val.system_key, signal_key: val.signal_key, type_save: val.TypeSave, period: 60}
		s.signals[sig_key] = signal
	}
	switch val.TypeSave {
	case 1:
		is_save := false
		pvalue, ok := s.svalueint[sig_key]
		if !ok {
			is_save = true
		}
		if pvalue.offline != int(val.Offline) {
			is_save = true
		}
		valuei := int64(val.Value)
		if pvalue.value != valuei {
			is_save = true
		}
		if is_save {
			s.svalueint[sig_key] = SValueInt{value: valuei, offline: int(val.Offline), utime: val.UTime}
			insert_valuei(s.db, signal.id, valuei, val.UTime, val.Offline)
		}
	case 2:
		is_save := false
		avg, ok := s.svalueavg[sig_key]
		if !ok {
			avg = newAVG(signal.period, signal.delta)
			s.svalueavg[sig_key] = avg
		}
		if avg.is_end_period(val.UTime) {
			is_save = true
		}
		if avg.is_delta() {
			is_save = true
		}
		if is_save {
			value_avg, utime_avg, err := avg.calc_avg()
			if err != nil {
			} else {
				err := insert_valuef(s.db, signal.id, value_avg, utime_avg, 0)
				if err != nil {
					fmt.Println("error insert fvalue", err)
				}
			}
			avg.set_new_period()
		}
		avg.add(val.Value, val.UTime)
	case 3:
		break
	}
}

func insert_valuei(db *sql.DB, system_id int64, value int64, utime int64, offline int64) error {
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

	str_sql := "INSERT INTO svsignal_ivalue(signal_id, value, utime, offline) VALUES (?,?,?,?)"
	//fmt.Println(str_sql, system_id, value, utime, offline)
	if _, err := tx.Exec(str_sql, system_id, value, utime, offline); err != nil {
		fmt.Println("Error", err)
		return err
	}
	return nil
}

func insert_valuef(db *sql.DB, system_id int64, value float64, utime int64, offline int64) error {
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

	str_sql := "INSERT INTO svsignal_fvalue(signal_id, value, utime, offline) VALUES (?,?,?,?)"
	//fmt.Println(str_sql, system_id, value, utime, offline)
	if _, err := tx.Exec(str_sql, system_id, value, utime, offline); err != nil {
		fmt.Println("Error", err)
		return err
	}
	return nil
}

func create_new_system(db *sql.DB, system_key string) error {
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

	str_sql := "INSERT INTO svsignal_system(system_key, name) VALUES (?, ?)"
	//fmt.Println(str_sql)
	if _, err := tx.Exec(str_sql, system_key, ""); err != nil {
		fmt.Println("Error", err)
		return err
	}
	return nil
}

func create_new_signal(db *sql.DB, system_key string, signal_key string, type_save int) (int64, error) {
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

	str_sql := "INSERT INTO svsignal_signal(system_key, signal_key, name, type_save, period, delta) VALUES (?,?,?,?,?,?)"
	//fmt.Println(str_sql)
	var result sql.Result
	// var err error
	if result, err = tx.Exec(str_sql, system_key, signal_key, "", type_save, 60, 10000); err != nil {
		fmt.Println("Error", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

type svsignal_system struct {
	name       string
	system_key string
}

func load_system(db *sql.DB) (*map[string]svsignal_system, error) {
	// Prepare statement for reading data
	rows, err := db.Query("SELECT system_key, name FROM svsignal_system")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	systems := make(map[string]svsignal_system)
	for rows.Next() {
		sys := svsignal_system{}
		err := rows.Scan(&sys.system_key, &sys.name)
		if err != nil {
			fmt.Println(err)
			continue
		}
		systems[sys.system_key] = sys
	}
	return &systems, nil
}

type svsignal_signal struct {
	id         int64
	system_key string
	signal_key string
	name       string
	type_save  int
	period     int
	delta      float32
	tags       *[]svsignal_tag
}

func load_signals(db *sql.DB) (*map[string]svsignal_signal, error) {

	tags, err := load_signal_tags(db)
	if err != nil {
		tags = nil
	}
	// Prepare statement for reading data
	rows, err := db.Query("SELECT id, system_key, signal_key, name, type_save, period, delta FROM svsignal_signal")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	signals := make(map[string]svsignal_signal)
	for rows.Next() {
		sig := svsignal_signal{}
		err := rows.Scan(&sig.id, &sig.system_key, &sig.signal_key, &sig.name, &sig.type_save, &sig.period, &sig.delta)
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
		signals[fmt.Sprintf("%s.%s", sig.system_key, sig.signal_key)] = sig
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

func request_data_signal(db *sql.DB, out chan interface{}, name_group string, signal svsignal_signal, begin int64, end int64, type_table int) {
	/*
		type_table 0 - none; 1 - svsignal_ivalue; 2 - svsignal_fvalue; 3 - svsignal_mvalue
	*/
	var sql string = ""
	switch type_table {
	case TYPE_IVALUE:
		sql = fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_ivalue WHERE signal_id=%d and utime >= %d and utime <=%d", signal.id, begin, end)
		break
	case TYPE_FVALUE:
		sql = fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_fvalue WHERE signal_id=%d and utime >= %d and utime <=%d", signal.id, begin, end)
		break
	case TYPE_MVALUE:
		sql = fmt.Sprintf("SELECT id, utime, max, min, mean, median, offline FROM svsignal_mvalue WHERE signal_id=%d and utime >= %d and utime <=%d", signal.id, begin, end)
		break
	default:
		out <- fmt.Errorf("error request data signal; type not found %d", type_table)
		return
	}
	var ivalues [][4]int64
	var fvalues [][4]interface{}
	if sql != "" {
		// fmt.Println(sql)
		rows, err := db.Query(sql)
		if err != nil {
			fmt.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			var ivalue [4]int64
			var fvalue [4]interface{}
			var err error
			switch type_table {
			case TYPE_IVALUE:
				err = rows.Scan(&ivalue[0], &ivalue[1], &ivalue[2], &ivalue[3])

			case TYPE_FVALUE:
				var id, utime, offline int64
				var fval float32
				err = rows.Scan(&id, &utime, &fval, &offline)
				fvalue[0], fvalue[1], fvalue[2], fvalue[3] = id, utime, fval, offline

			case TYPE_MVALUE:
				break
			}
			if err != nil {
				fmt.Println(err)
				continue
			}
			switch type_table {
			case TYPE_IVALUE:
				ivalues = append(ivalues, ivalue)
			case TYPE_FVALUE:
				fvalues = append(fvalues, fvalue)
			case TYPE_MVALUE:
				break
			}
		}
	}
	tags := []RLS_Tag{}
	for _, tag := range *signal.tags {
		tags = append(tags, RLS_Tag{Tag: tag.tag, Value: tag.value})
	}
	switch type_table {
	case TYPE_IVALUE:
		out <- ResponseDataSignalT1{
			GroupKey:   signal.system_key,
			GroupName:  name_group,
			SignalKey:  signal.signal_key,
			SignalName: signal.name,
			TypeSave:   type_table,
			Values:     ivalues,
			Tags:       tags,
		}
	case TYPE_FVALUE:
		out <- ResponseDataSignalT2{
			GroupKey:   signal.system_key,
			GroupName:  name_group,
			SignalKey:  signal.signal_key,
			SignalName: signal.name,
			TypeSave:   type_table,
			Values:     fvalues,
			Tags:       tags,
		}
	case TYPE_MVALUE:
		break
	}
}
