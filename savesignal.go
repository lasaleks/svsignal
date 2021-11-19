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
	CH_SET_SIGNAL   chan SetSignal
	groups          map[string]svsignal_group
	signals         map[string]*svsignal_signal
	svalueint       map[string]SValueInt
	svalueavg       map[string]*AVG
	CH_REQUEST_HTTP chan interface{}
	debug_level     int
}

func newSVS() *SVSignalDB {
	return &SVSignalDB{
		CH_SAVE_VALUE:   make(chan ValueSignal, 1),
		CH_SET_SIGNAL:   make(chan SetSignal, 1),
		CH_REQUEST_HTTP: make(chan interface{}, 1),
		svalueint:       make(map[string]SValueInt),
		svalueavg:       make(map[string]*AVG),
	}
}

func (s *SVSignalDB) run(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	groups, err := load_groups(s.db)
	if err != nil {
		log.Panicf("error load svsignal_systems %v", err)
	}
	signals, err := load_signals(s.db)
	if err != nil {
		log.Panicf("error load svsignal_signals %v", err)
	}
	s.groups = *groups
	s.signals = *signals

	for {
		select {
		case <-ctx.Done():
			//log.Println("SaveSignal run Done")
			return
		case msg, ok := <-s.CH_SAVE_VALUE:
			if ok {
				s.save_value(&msg)
			} else {
				return
			}
		case msg, ok := <-s.CH_SET_SIGNAL:
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
						group, ok := s.groups[request.groupkey]
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

	for _, group := range s.groups {
		lsig.Groups[group.group_key] = &RLS_Groups{Name: group.name, Signals: make([]RLS_Signal, 0)}
	}
	// сортируем ключи
	keys := make([]string, 0, len(s.signals))
	for k := range s.signals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		data := s.signals[key]
		_, ok := lsig.Groups[data.group_key]
		if !ok {
			lsig.Groups[data.group_key] = &RLS_Groups{}
		}
		tags := []RLS_Tag{}
		for _, tag := range *data.tags {
			tags = append(tags, RLS_Tag{Tag: tag.tag, Value: tag.value})
		}
		lsig.Groups[data.group_key].Signals = append(lsig.Groups[data.group_key].Signals, RLS_Signal{
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
	group, ok := s.groups[val.group_key]
	if !ok {
		// create systems
		id, err := create_new_group(s.db, val.group_key)
		if err != nil {
			log.Printf("error create new system: system_key:%s; error:%v", val.group_key, err)
			return
		} else {
			//fmt.Println("create systems", val.system_key, "Ok")
			group = svsignal_group{id: id, group_key: val.group_key, name: ""}
			s.groups[val.group_key] = group
		}

	}
	sig_key := fmt.Sprintf("%s.%s", val.group_key, val.signal_key)

	if s.debug_level >= 4 {
		log.Printf("SaveSignal:%s value:%v offline:%d utime:%d\n", sig_key, val.Value, val.Offline, val.UTime)
	}

	signal, ok := s.signals[sig_key]
	if !ok {
		// create signals
		//fmt.Println("create signal", val.signal_key)
		id, err := create_new_signal(s.db, group.id, val.signal_key, "", val.TypeSave, 60, 10000)
		if err != nil {
			log.Println("Error create signal", val, err)
			return
		}
		//log.Println("create signal", val, id, "OK")
		signal = &svsignal_signal{id: id, group_id: group.id, group_key: val.group_key, signal_key: val.signal_key, type_save: val.TypeSave, period: 60}
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

	str_sql := "INSERT INTO svsignal_group(system_key, name) VALUES (?, ?)"
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

	sql := "UPDATE svsignal_signal SET name=?, type_save=?, period=?, delta=? WHERE signal_id=?"
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

	rows, err := db.Query("SELECT id, group_id, g.group_key, signal_key, name, type_save, period, delta FROM svsignal_signal inner join svsignal_group g on g.id=group_id")
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

func request_data_signal(db *sql.DB, out chan interface{}, name_group string, signal *svsignal_signal, begin int64, end int64, type_table int) {
	/*
		type_table 0 - none; 1 - svsignal_ivalue; 2 - svsignal_fvalue; 3 - svsignal_mvalue
	*/
	if signal == nil {
		out <- fmt.Errorf("error request_data_signal signal:nil")
		return
	}
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
			GroupKey:   signal.group_key,
			GroupName:  name_group,
			SignalKey:  signal.signal_key,
			SignalName: signal.name,
			TypeSave:   type_table,
			Values:     ivalues,
			Tags:       tags,
		}
	case TYPE_FVALUE:
		out <- ResponseDataSignalT2{
			GroupKey:   signal.group_key,
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

func (s *SVSignalDB) set_signal(setsig SetSignal) {
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
	signal, ok := s.signals[sig_key]
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
		s.signals[sig_key] = signal
	} else {
		if signal.name != setsig.Name || signal.type_save != setsig.TypeSave || signal.period != setsig.Period || signal.delta != setsig.Delta {
			err := update_signal(s.db, signal.id, setsig.Name, setsig.TypeSave, setsig.Period, setsig.Delta)
			if err != nil {
				log.Println("Error update signal", setsig, err)
				return
			}
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
	}
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
	str_sql := "INSERT INTO svsignal_tag(system_id, tag, value) VALUES (?,?,?)"
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
