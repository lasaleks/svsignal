package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	goutils "github.com/lasaleks/go-utils"
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

type BulkInsertBuffer struct {
	ivalue     []model.IValue
	fvalue     []model.FValue
	ivalue_cnt int
	fvalue_cnt int
}

type RequestData struct {
	key         string
	begin       int64
	end         int64
	CH_RESPONSE chan interface{}
}

type SVSignalDB struct {
	db              *sql.DB // !!!! remove
	group_id        map[uint]*model.Group
	group_key       map[string]*model.Group
	signal_key      map[string]*model.Signal
	svalueint       map[string]*SValueInt
	svalueavg       map[string]*AVG
	CH_REQUEST_HTTP chan interface{}
	lt_save_values  int64
	bulkBuffer      BulkInsertBuffer
	mu              sync.RWMutex
}

func newSVS() *SVSignalDB {
	return &SVSignalDB{
		group_id:        make(map[uint]*model.Group),
		group_key:       make(map[string]*model.Group),
		signal_key:      make(map[string]*model.Signal),
		CH_REQUEST_HTTP: make(chan interface{}, 1),
		svalueint:       make(map[string]*SValueInt),
		svalueavg:       make(map[string]*AVG),
		bulkBuffer: BulkInsertBuffer{
			ivalue: make([]model.IValue, cfg.SVSIGNAL.BulkSize),
			fvalue: make([]model.FValue, cfg.SVSIGNAL.BulkSize),
		},
		lt_save_values: time.Now().Unix(),
	}
}

func (s *SVSignalDB) Load() {
	var groups []*model.Group
	// works because destination struct is passed in
	if result := DB.Find(&groups); result.Error != nil {
		// error load svsignal_systems
		log.Panicln(result.Statement.SQL, result.Error)
	}
	for _, group := range groups {
		s.group_key[group.Key] = group
		s.group_id[group.ID] = group
	}

	var signals []*model.Signal
	//db.Joins("Company").Find(&users)
	// works because destination struct is passed in
	//db.Joins("Company").Find(&users)
	if result := DB.Preload("Tags").Find(&signals); result.Error != nil {
		// error load svsignal_systems
		log.Println("not found signals", result.Statement.SQL, result.Error)
	}
	for _, signal := range signals {
		group, ok := s.group_id[signal.GroupID]
		if !ok {
			continue
		}
		group.Signals = append(group.Signals, signal)
		s.signal_key[fmt.Sprintf("%s.%s", group.Key, signal.Key)] = signal
	}
	/*
		signals, err := load_signals(s.db)
		if err != nil {
			log.Panicf("error load svsignal_signals %v", err)
		}
		s.signals = *signals
	*/
	fmt.Printf("groups:%d signals:%d", len(s.group_key), len(s.signal_key))
}

func (s *SVSignalDB) Run(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	defer func() {
		s.writeBulkBuffer(0)
		log.Println("SVSignalDB.Run End")
	}()

	for {
		select {
		case <-ctx.Done():
			return
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
						var value_signal *SValueInt
						if signal.TypeSave == 1 {
							value_signal = s.svalueint[signal_key]
							if value_signal != nil {
								value_signal = &SValueInt{value: value_signal.value, utime: value_signal.utime, offline: value_signal.offline}
							}
						}

						s.get_data_signal(s.db, request.CH_RESPONSE, signal, value_signal, request.begin, request.end)
					}
				}
			}
		case request := <-CH_REQUEST_DATA:
			signal, ok := s.signal_key[request.key]
			if !ok {
				request.CH_RESPONSE <- fmt.Errorf("not found %s", request.key)
			} else {
				var value_signal *SValueInt
				if signal.TypeSave == 1 {
					value_signal = s.svalueint[request.key]
					if value_signal != nil {
						value_signal = &SValueInt{value: value_signal.value, utime: value_signal.utime, offline: value_signal.offline}
					}
				}

				s.get_data_signal(s.db, request.CH_RESPONSE, signal, value_signal, request.begin, request.end)
			}
		case <-time.After(time.Second * 1):
		}
		if s.lt_save_values+cfg.SVSIGNAL.PeriodSave < time.Now().Unix() {
			s.writeBulkBuffer(0)
			s.lt_save_values = time.Now().Unix()
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
	signal, ok := s.signal_key[sig_key]
	if !ok {
		if cfg.SVSIGNAL.DEBUG_LEVEL >= 1 {
			log.Printf("not found signal, key:%s", sig_key)
		}
		return
	}
	if cfg.SVSIGNAL.DEBUG_LEVEL >= 6 {
		log.Printf("SaveSignal:%s value:%v offline:%d utime:%d\n", sig_key, val.Value, val.Offline, val.UTime)
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
			if s.bulkBuffer.ivalue_cnt >= cfg.SVSIGNAL.BulkSize {
				s.writeBulkBuffer(1)
			}
			s.bulkBuffer.ivalue[s.bulkBuffer.ivalue_cnt] = model.IValue{SignalID: signal.ID, UTime: val.UTime, Value: int32(val.Value), OffLine: goutils.IntToBool(int(val.Offline))}
			s.bulkBuffer.ivalue_cnt++

		}
		pvalue.value = valuei
		pvalue.offline = offline
		pvalue.utime = val.UTime
		if cfg.SVSIGNAL.DEBUG_LEVEL >= 8 {
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
			if cfg.SVSIGNAL.DEBUG_LEVEL >= 1 {
				log.Printf("error add_value %+v %s", val, err)
			}
			return
		}

		for _, value := range l_values {
			if s.bulkBuffer.fvalue_cnt >= cfg.SVSIGNAL.BulkSize {
				s.writeBulkBuffer(2)
			}
			s.bulkBuffer.fvalue[s.bulkBuffer.fvalue_cnt] = model.FValue{SignalID: signal.ID, UTime: value.utime, Value: value.value, OffLine: goutils.IntToBool(int(value.offline))}
			s.bulkBuffer.fvalue_cnt++
		}
	case 3:
		break
	}
}

// typeValues 0 - All, 1 - IValue, 2 - FValue
func (s *SVSignalDB) writeBulkBuffer(typeValues int) {
	if s.bulkBuffer.ivalue_cnt == 0 && s.bulkBuffer.fvalue_cnt == 0 {
		return
	}
	tx := DB.Begin()
	if typeValues == 1 || typeValues == 0 && s.bulkBuffer.ivalue_cnt > 0 {
		res := tx.Create(s.bulkBuffer.ivalue[:s.bulkBuffer.ivalue_cnt])
		if res.Error != nil {
			log.Println("Create error:", res.Error)
		}
		s.bulkBuffer.ivalue_cnt = 0
	}
	if typeValues == 2 || typeValues == 0 && s.bulkBuffer.fvalue_cnt > 0 {
		res := tx.Create(s.bulkBuffer.fvalue[:s.bulkBuffer.fvalue_cnt])
		if res.Error != nil {
			log.Println("Create error:", res.Error)
		}
		s.bulkBuffer.fvalue_cnt = 0
	}

	res := tx.Commit()
	if res.Error != nil {
		log.Println("Commit error:", res.Error)
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

func (s *SVSignalDB) get_data_signal_i(signal *model.Signal, value_signal_i *SValueInt, begin int64, end int64) (*[][4]int64, error) {
	ivalues := &[][4]int64{}

	add_end_point := true

	// запрос значения до заданного периода begin-end
	value := model.IValue{}
	res := DB.Model(&model.IValue{}).Where("id=(?)", DB.Model(&model.IValue{}).Select("max(id)").Where("signal_id=? and utime<?", signal.ID, begin)).First(&value)
	if res.Error == nil {
		*ivalues = append(*ivalues, [4]int64{int64(value.ID), value.UTime, int64(value.Value), int64(goutils.BoolToInt(value.OffLine))})
	}

	// запрашиваем данные за период из бд
	values := []model.IValue{}
	res = DB.Model(&model.IValue{}).Where("signal_id = ? and utime >= ? and utime <=?", signal.ID, begin, end).Find(&values)
	if res.Error == nil {
		for _, val := range values {
			*ivalues = append(*ivalues, [4]int64{int64(val.ID), val.UTime, int64(val.Value), int64(goutils.BoolToInt(val.OffLine))})
		}
	}

	// запрос значения после заданного периода begin-end
	value = model.IValue{}
	res = DB.Model(&model.IValue{}).Where("id=(?)", DB.Model(&model.IValue{}).Select("max(id)").Where("signal_id=? and utime>?", signal.ID, end)).First(&value)
	if res.Error == nil {
		*ivalues = append(*ivalues, [4]int64{int64(value.ID), value.UTime, int64(value.Value), int64(goutils.BoolToInt(value.OffLine))})
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
	// запрашиваем данные за период из бд
	values := []model.FValue{}
	res := DB.Model(&model.FValue{}).Where("signal_id = ? and utime >= ? and utime <=?", signal.ID, begin, end).Find(&values)
	if res.Error != nil {
		return nil, res.Error
	}

	for _, val := range values {
		*fvalues = append(*fvalues, [4]interface{}{int64(val.ID), val.UTime, val.Value, int64(goutils.BoolToInt(val.OffLine))})
	}

	return fvalues, nil
}

func (s *SVSignalDB) get_data_signal(db *sql.DB, out chan interface{}, signal *model.Signal, value_signal_i *SValueInt, begin int64, end int64) {
	if signal == nil {
		out <- fmt.Errorf("signal:nil error")
		return
	}

	group, found := s.group_id[signal.GroupID]
	if !found {
		out <- fmt.Errorf("not found group")
		return
	}

	tags := []RLS_Tag{}
	for _, tag := range signal.Tags {
		tags = append(tags, RLS_Tag{Tag: tag.Tag, Value: tag.Value})
	}

	switch signal.TypeSave {
	case TYPE_IVALUE:
		ivalues, err := s.get_data_signal_i(signal, value_signal_i, begin, end)
		if err != nil {
			out <- fmt.Errorf("error key:%s %s", signal.Key, err)
		} else {
			out <- ResponseDataSignalT1{
				GroupKey:   group.Key,
				GroupName:  group.Name,
				SignalKey:  signal.Key,
				SignalName: signal.Name,
				TypeSave:   int(signal.TypeSave),
				Values:     *ivalues,
				Tags:       tags,
			}
		}
	case TYPE_FVALUE:
		fvalues, err := s.get_data_signal_f(signal, begin, end)
		if err != nil {
			out <- fmt.Errorf("error key:%s %s", signal.Key, err)
		} else {
			out <- ResponseDataSignalT2{
				GroupKey:   group.Key,
				GroupName:  group.Name,
				SignalKey:  signal.Key,
				SignalName: signal.Name,
				TypeSave:   int(signal.TypeSave),
				Values:     *fvalues,
				Tags:       tags,
			}
		}
	case TYPE_MVALUE:
	default:
		out <- fmt.Errorf("error request data signal; type not found %d", signal.TypeSave)
		return
	}
}

func (s *SVSignalDB) set_signal(setsig SetSignal) {
	group, ok := s.group_key[setsig.group_key]
	if !ok {
		new_group := &model.Group{Key: setsig.group_key}
		res := DB.Create(new_group)
		if res.Error != nil {
			log.Println("create group err:", res.Error)
			return
		}
		group = new_group
		s.group_id[group.ID] = group
		s.group_key[setsig.group_key] = group
	}
	sig_key := fmt.Sprintf("%s.%s", setsig.group_key, setsig.signal_key)
	signal, ok := s.signal_key[sig_key]
	if !ok {
		// create signals
		//fmt.Println("create signal", val.signal_key)
		//str_sql := "INSERT INTO svsignal_signal(group_id, signal_key, name, type_save, period, delta) VALUES (?,?,?,?,?,?)"
		new_signal := &model.Signal{GroupID: group.ID, Key: setsig.signal_key, Name: setsig.Name, TypeSave: int8(setsig.TypeSave), Period: setsig.Period, Delta: setsig.Delta}
		res := DB.Create(new_signal)
		if res.Error != nil {
			log.Println("create signal err:", res.Error)
		}
		signal = new_signal
		s.signal_key[sig_key] = signal
	} else {
		fmt.Printf("update signal %++v\n", setsig)
		if signal.Name != setsig.Name || int(signal.TypeSave) != setsig.TypeSave || signal.Period != setsig.Period || signal.Delta != setsig.Delta {
			res := DB.Save(signal)
			if res.Error != nil {
				log.Println("Error update signal", setsig, res.Error)
				return
			}
			signal.Name = setsig.Name
			signal.TypeSave = int8(setsig.TypeSave)
			signal.Period = setsig.Period
			signal.Delta = setsig.Delta

		}
	}

	/*if signal.tags == nil {
		signal.tags = &[]svsignal_tag{}
	}*/
	not_remove := []int64{}
	for _, utag := range setsig.Tags {
		create := true
		for i := 0; i < len(signal.Tags); i++ {
			tag := signal.Tags[i]
			if tag.Tag == utag.Tag {
				not_remove = append(not_remove, int64(tag.ID))
				create = false
				if tag.Value != utag.Value {
					// update
					tag.Value = utag.Value
					if res := DB.Save(tag); res.Error != nil {
						log.Println("error tag save err:", res.Error)
					}
				}
				break
			}
		}
		if create {
			// create
			new_tag := model.Tag{SignalID: signal.ID, Tag: utag.Tag, Value: utag.Value}
			res := DB.Create(&new_tag)
			if res.Error != nil {
				log.Println("create tag err:", res.Error)
			} else {
				signal.Tags = append(signal.Tags, new_tag)
				not_remove = append(not_remove, int64(new_tag.ID))
			}
		}
	}
	// удаляем лишние метки
	for i := 0; i < len(signal.Tags); {
		remove := true
		for _, not_remove_id := range not_remove {
			if not_remove_id == int64(signal.Tags[i].ID) {
				remove = false
				break
			}
		}
		if remove {
			DB.Delete(&signal.Tags[i])
			signal.Tags = append(signal.Tags[:i], signal.Tags[i+1:]...)
			continue
		}
		i++
	}
}
