package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
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

type SValueFloat struct {
	value   float64
	offline int
	utime   int64
}

type SVSignalDB struct {
	db              *sql.DB
	CH_SAVE_VALUE   chan ValueSignal
	CH_SET_SIGNAL   chan SetSignal
	groups          map[string]svsignal_group
	signals         map[string]*svsignal_signal
	svalueint       map[string]*SValueInt
	svalueavg       map[string]*AVG
	CH_REQUEST_HTTP chan interface{}
	debug_level     int

	// cache insert
	max_multiply_insert int
	cache_i             map[int64]*[]SValueInt
	size_cache_i        int
	cache_f             map[int64]*[]SValueFloat
	size_cache_f        int
	buff_size           int
}

func newSVS(cfg Config) *SVSignalDB {
	return &SVSignalDB{
		CH_SAVE_VALUE:       make(chan ValueSignal, 1),
		CH_SET_SIGNAL:       make(chan SetSignal, 1),
		CH_REQUEST_HTTP:     make(chan interface{}, 1),
		svalueint:           make(map[string]*SValueInt),
		svalueavg:           make(map[string]*AVG),
		cache_i:             make(map[int64]*[]SValueInt),
		cache_f:             make(map[int64]*[]SValueFloat),
		max_multiply_insert: cfg.CONFIG.MaxMultiplyInsert,
		buff_size:           cfg.CONFIG.BuffSize,
		debug_level:         cfg.CONFIG.DEBUG_LEVEL,
	}
}

func (s *SVSignalDB) run(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	defer func() {
		s.flush_cache(1)
		s.flush_cache(2)
	}()
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
						var value_signal *SValueInt
						if signal.type_save == 1 {
							value_signal = s.svalueint[signal_key]
							if value_signal != nil {
								value_signal = &SValueInt{value: value_signal.value, utime: value_signal.utime, offline: value_signal.offline}
							}
						}

						s.get_data_signal(s.db, request.CH_RESPONSE, name_group, signal, value_signal, request.begin, request.end, signal.type_save)
					}
				}
			}
		}
	}
}

func (s *SVSignalDB) response_list_signal() *ResponseListSignal {
	lsig := ResponseListSignal{Groups: make(map[string]*RLS_Groups)}
	for _, group := range s.groups {
		lsig.Groups[group.group_key] = &RLS_Groups{Name: group.name, Signals: make(map[string]RLS_Signal)}
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
		if data.tags != nil {
			for _, tag := range *data.tags {
				tags = append(tags, RLS_Tag{Tag: tag.tag, Value: tag.value})
			}
		}
		lsig.Groups[data.group_key].Signals[key] = RLS_Signal{
			Name:     data.name,
			TypeSave: data.type_save,
			Period:   data.period,
			Delta:    data.delta,
			Tags:     tags,
		}
	}
	return &lsig
}

func (s *SVSignalDB) save_value(val *ValueSignal) {
	sig_key := fmt.Sprintf("%s.%s", val.group_key, val.signal_key)

	if s.debug_level >= 6 {
		log.Printf("SaveSignal:%s value:%v offline:%d utime:%d\n", sig_key, val.Value, val.Offline, val.UTime)
	}

	signal, ok := s.signals[sig_key]
	if !ok {
		if s.debug_level >= 1 {
			fmt.Printf("not found signal, key:%s", sig_key)
		}
		return
	}
	switch signal.type_save {
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
			s.insert_valuei(s.db, signal.id, valuei, val.UTime, val.Offline, s.debug_level >= 6)
		}
		pvalue.value = valuei
		pvalue.offline = offline
		pvalue.utime = val.UTime
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
				err := s.insert_valuef(s.db, signal.id, value_avg, utime_avg, 0, s.debug_level >= 6)
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

func setupBindVars(stmt, bindVars string, len int) string {
	bindVars += ","
	stmt = fmt.Sprintf(stmt, strings.Repeat(bindVars, len))
	return strings.TrimSuffix(stmt, ",")
}

func (s *SVSignalDB) flush_cache(type_save int) error {
	vals := []interface{}{}
	var sql string
	len_vals := 0
	end_idx := func(len_arr int) int {
		if len_arr >= s.max_multiply_insert {
			return s.max_multiply_insert
		}
		return len_arr
	}
	switch type_save {
	case 1:
		for signal_id := range s.cache_i {
			values := s.cache_i[signal_id]
			for _, val := range (*values)[:end_idx(len(*values))] {
				vals = append(vals, signal_id, val.value, val.utime, int(val.offline))
				len_vals++
			}
		}
		sql = setupBindVars("INSERT INTO svsignal_ivalue(signal_id, value, utime, offline) VALUES %s", "(?,?,?,?)", len_vals)
	case 2:
		for signal_id := range s.cache_f {
			values := s.cache_f[signal_id]
			for _, val := range (*values)[:end_idx(len(*values))] {
				vals = append(vals, signal_id, val.value, val.utime, int(val.offline))
				len_vals++
			}
		}
		sql = setupBindVars("INSERT INTO svsignal_fvalue(signal_id, value, utime, offline) VALUES %s", "(?,?,?,?)", len_vals)
	default:
		return fmt.Errorf("Error type save")
	}

	if len_vals == 0 {
		return nil
	}

	if s.debug_level >= 4 {
		log.Println("Insert svsignal_ivalue len:", len_vals)
	}

	tx, err := s.db.Begin()
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

	result, err := tx.Exec(sql, vals...)
	if err != nil {
		fmt.Println("Error", err)
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("Error", err)
		return err
	}
	last_id, err := result.LastInsertId()
	if err != nil {
		fmt.Println("Error", err)
		return err
	}
	/*if affected != int64(len(vals)) {
	}*/
	if s.debug_level >= 4 {
		log.Println("Insert - RowsAffected", affected, "LastInsertId: ", last_id)
	}
	switch type_save {
	case 1:
		s.cache_i = make(map[int64]*[]SValueInt)
		s.size_cache_i = 0
	case 2:
		s.cache_i = make(map[int64]*[]SValueInt)
		s.size_cache_f = 0
	}

	return nil
}

func (s *SVSignalDB) insert_valuei(db *sql.DB, signal_id int64, value int64, utime int64, offline int64, debug bool) error {
	if s.debug_level >= 6 {
		fmt.Println("insert_valuei", signal_id, value, utime, offline)
	}
	cache, ok := s.cache_i[signal_id]
	if !ok {
		cache = &[]SValueInt{{value: value, utime: utime, offline: int(offline)}}
		s.cache_i[signal_id] = cache
	} else {
		*cache = append(*cache, SValueInt{value: value, utime: utime, offline: int(offline)})
	}
	s.size_cache_i++

	if s.buff_size < (s.size_cache_i+s.size_cache_f) && s.size_cache_i >= s.max_multiply_insert {
		s.flush_cache(1)
	}
	return nil
}

func (s *SVSignalDB) insert_valuef(db *sql.DB, signal_id int64, value float64, utime int64, offline int64, debug bool) error {
	if s.debug_level >= 6 {
		fmt.Println("insert_valuef", signal_id, value, utime, offline)
	}
	cache, ok := s.cache_f[signal_id]
	if !ok {
		cache = &[]SValueFloat{
			{value: value, utime: utime, offline: int(offline)},
		}
		s.cache_f[signal_id] = cache
	} else {
		*cache = append(*cache, SValueFloat{value: value, utime: utime, offline: int(offline)})
	}
	s.size_cache_f++

	if s.buff_size < (s.size_cache_i+s.size_cache_f) && s.size_cache_f >= s.max_multiply_insert {
		s.flush_cache(2)
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

func (s *SVSignalDB) get_data_signal_i(signal *svsignal_signal, value_signal_i *SValueInt, begin int64, end int64) (*[][4]int64, error) {
	ivalues := &[][4]int64{}

	// определяем наличие данных в буффере на вставку
	var firts_utime_in_cache *int64
	var last_utime_in_cache *int64
	cache, ok := s.cache_i[signal.id]
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
		row := s.db.QueryRow(fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_ivalue WHERE signal_id=%d and id=(select max(id) from svsignal_ivalue where utime<%d and signal_id=%d)", signal.id, begin, signal.id))
		if row != nil {
			var ivalue [4]int64
			err := row.Scan(&ivalue[0], &ivalue[1], &ivalue[2], &ivalue[3])
			if err != nil {
				fmt.Println(err)
			} else {
				*ivalues = append(*ivalues, ivalue)
			}
		}
	}

	if query_db {
		// запрашиваем данные за период из бд
		rows, err := s.db.Query(fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_ivalue WHERE signal_id=%d and utime >= %d and utime <=%d", signal.id, begin, end))
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
			row := s.db.QueryRow(fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_ivalue WHERE signal_id=%d and id=(select max(id) from svsignal_ivalue where utime>%d and signal_id=%d)", signal.id, end, signal.id))
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

func (s *SVSignalDB) get_data_signal_f(signal *svsignal_signal, begin int64, end int64) (*[][4]interface{}, error) {
	fvalues := &[][4]interface{}{}

	// определяем наличие данных в буффере на вставку
	var firts_utime_in_cache *int64
	var last_utime_in_cache *int64
	cache, ok := s.cache_f[signal.id]
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
		rows, err := s.db.Query(fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_fvalue WHERE signal_id=%d and utime >= %d and utime <=%d", signal.id, begin, end))
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

func (s *SVSignalDB) get_data_signal(db *sql.DB, out chan interface{}, name_group string, signal *svsignal_signal, value_signal_i *SValueInt, begin int64, end int64, type_table int) {
	if signal == nil {
		out <- fmt.Errorf("error get_data_signal signal:nil")
		return
	}
	tags := []RLS_Tag{}
	for _, tag := range *signal.tags {
		tags = append(tags, RLS_Tag{Tag: tag.tag, Value: tag.value})
	}

	switch type_table {
	case TYPE_IVALUE:
		ivalues, err := s.get_data_signal_i(signal, value_signal_i, begin, end)
		if err != nil {
			out <- fmt.Errorf("error key:%s %s", signal.signal_key, err)
		} else {
			out <- ResponseDataSignalT1{
				GroupKey:   signal.group_key,
				GroupName:  name_group,
				SignalKey:  signal.signal_key,
				SignalName: signal.name,
				TypeSave:   type_table,
				Values:     *ivalues,
				Tags:       tags,
			}
		}
	case TYPE_FVALUE:
		fvalues, err := s.get_data_signal_f(signal, begin, end)
		if err != nil {
			out <- fmt.Errorf("error key:%s %s", signal.signal_key, err)
		} else {
			out <- ResponseDataSignalT2{
				GroupKey:   signal.group_key,
				GroupName:  name_group,
				SignalKey:  signal.signal_key,
				SignalName: signal.name,
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

/*
func (s *SVSignalDB) get_data_signal(db *sql.DB, out chan interface{}, name_group string, signal *svsignal_signal, value_signal_i *SValueInt, begin int64, end int64, type_table int) {
	// type_table 0 - none; 1 - svsignal_ivalue; 2 - svsignal_fvalue; 3 - svsignal_mvalue
	if signal == nil {
		out <- fmt.Errorf("error get_data_signal signal:nil")
		return
	}

	var sql string = ""
	switch type_table {
	case TYPE_IVALUE:
		sql = fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_ivalue WHERE signal_id=%d and utime >= %d and utime <=%d", signal.id, begin, end)
	case TYPE_FVALUE:
		sql = fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_fvalue WHERE signal_id=%d and utime >= %d and utime <=%d", signal.id, begin, end)
	case TYPE_MVALUE:
		sql = fmt.Sprintf("SELECT id, utime, max, min, mean, median, offline FROM svsignal_mvalue WHERE signal_id=%d and utime >= %d and utime <=%d", signal.id, begin, end)
	default:
		out <- fmt.Errorf("error request data signal; type not found %d", type_table)
		return
	}

	var ivalues [][4]int64
	var fvalues [][4]interface{}

	var firts_utime_in_cache *int64
	var last_utime_in_cache *int64
	// получаем период данных в кеше
	switch type_table {
	case TYPE_IVALUE:
		cache, ok := s.cache_i[signal.id]
		if ok {
			len_cache := len(*cache)
			if len_cache > 0 {
				firts_utime_in_cache = &(*cache)[0].utime
			}
			if len_cache > 1 {
				last_utime_in_cache = &(*cache)[len_cache-1].utime
			}
		}
	case TYPE_FVALUE:
		cache, ok := s.cache_f[signal.id]
		if ok {
			len_cache := len(*cache)
			if len_cache > 0 {
				firts_utime_in_cache = &(*cache)[0].utime
			}
			if len_cache > 1 {
				last_utime_in_cache = &(*cache)[len_cache-1].utime
			}
		}
	}

	// запрос значения до заданного периода begin-end
	if type_table == TYPE_IVALUE {
		sql_begin := fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_ivalue WHERE signal_id=%d and id=(select max(id) from svsignal_ivalue where utime<%d and signal_id=%d)", signal.id, begin, signal.id)
		row := db.QueryRow(sql_begin)
		if row != nil {
			var ivalue [4]int64
			err := row.Scan(&ivalue[0], &ivalue[1], &ivalue[2], &ivalue[3])
			if err == nil {
				ivalues = append(ivalues, ivalue)
			}
		}
	}

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

				//case TYPE_MVALUE:break
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
				//case TYPE_MVALUE:break
			}
		}
	}

	if type_table == TYPE_IVALUE {
		// запрос значения после заданного периода begin-end
		sql_begin := fmt.Sprintf("SELECT id, utime, value, offline FROM svsignal_ivalue WHERE signal_id=%d and id=(select max(id) from svsignal_ivalue where utime>%d and signal_id=%d)", signal.id, end, signal.id)
		row := db.QueryRow(sql_begin)
		if row != nil {
			var ivalue [4]int64
			err := row.Scan(&ivalue[0], &ivalue[1], &ivalue[2], &ivalue[3])
			if err == nil {
				ivalues = append(ivalues, ivalue)
			}
		}

		// задания дополнительной точки с временем равным полученному последнему значению.
		if value_signal_i != nil {
			len := len(ivalues)
			if len > 0 {
				len--
				if ivalues[len][1] < value_signal_i.utime && ivalues[len][2] == value_signal_i.value && ivalues[len][3] == int64(value_signal_i.offline) {
					ivalues = append(ivalues, ivalues[len])
					ivalues[len+1][1] = value_signal_i.utime
				}
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
}*/

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
