package main

import (
	"fmt"
	"math"
	"reflect"
)

// среднее значение по методу трапеций сумма по методу трапеций
type AVG struct {
	sum                 float64
	begin_utime         int64
	period              int
	delta               float32
	period_save_offline int

	prev_value      SValueFloat
	prev_avg        SValueFloat
	prev_save_in_db SValueFloat
}

var (
	error_utime_zero error = fmt.Errorf("error utime is zero")
	error_not_found  error = fmt.Errorf("error not found")
)

func newAVG(period int, delta float32) *AVG {
	return &AVG{
		period:              period,
		delta:               delta,
		period_save_offline: 600,
	}
}

func is_end_period(utime int64, begin_utime int64, period int64) bool {
	if period == 0 {
		return true
	}
	// целочисленное деление на период используется для выравнивания точки сохранения, для сигналов с одинаковым периодом
	ptime := utime / period
	btime := begin_utime / period
	if ptime > btime {
		return true
	}
	return false
}

func is_delta(value SValueFloat, prev_value SValueFloat, delta float64) bool {
	if prev_value.utime != 0 && math.Abs(prev_value.value-value.value) >= delta {
		return true
	}
	return false
}

func calc_avg_2(begin int64, end int64, sum float64) (error, bool, float64, int64) {
	if begin == 0 {
		return error_not_found, false, 0, 0
	}
	if end == begin {
		return nil, false, 0, 0
	}
	value := sum / float64(end-begin)
	return nil, true, value, (end + begin) / 2
}

// добавить значение
func (a *AVG) add_value(value SValueFloat) (error, []SValueFloat) {
	if value.utime == 0 {
		return error_utime_zero, []SValueFloat{}
	}

	/*сохранение offline*/
	if value.offline == 1 {
		if a.prev_value.offline == 0 {
			err, is_calc, avg, utime := calc_avg_2(a.begin_utime, a.prev_value.utime, a.sum)
			if err != nil {
			} else if is_calc {
				a.prev_value = value
				a.prev_save_in_db = value
				a.prev_avg = SValueFloat{value: avg, utime: utime, offline: 0}
				a.sum = 0
				a.begin_utime = 0
				return nil, []SValueFloat{
					{value: avg, utime: utime, offline: 0},
					{value: 0, utime: value.utime, offline: value.offline},
				}
			}
		} else if a.prev_value.offline == 1 && is_end_period(value.utime, a.prev_save_in_db.utime, int64(a.period_save_offline)) {
			// Сохраняем в БД значения offline, если после прошлого сохранения в БД, прошло больше чем period_save_offline.
			a.prev_value = value
			a.prev_save_in_db = value
			a.sum = 0
			a.begin_utime = 0
			return nil, []SValueFloat{{value: 0, utime: value.utime, offline: value.offline}}
		}
		a.prev_value = value
		return nil, []SValueFloat{}
	} else if a.prev_value.offline == 1 {
		if is_delta(value, a.prev_avg, float64(a.delta)) || is_end_period(value.utime, a.prev_avg.utime, int64(a.period)) {
			//сохраняем значение полсе offline, если вышло после прошлого сохранения за период, или больше предыдущего ср. на дельту
			a.prev_save_in_db = value
			a.sum = 0
			a.begin_utime = value.utime
			a.prev_avg = value
			return nil, []SValueFloat{value}
		}
		a.begin_utime = value.utime
		a.prev_avg = value
		a.prev_value = value
		return nil, []SValueFloat{}
	}

	/*if a.begin_utime == 0 {
		// новый период
		a.begin_utime = value.utime
		a.prev_value = value
		return nil, []SValueFloat{}
	}*/

	if is_delta(value, a.prev_value, float64(a.delta)) {
		err, is_calc, avg, utime := calc_avg_2(a.begin_utime, a.prev_value.utime, a.sum)
		if err == nil && is_calc {
			s_p_value := a.prev_value
			a.prev_value = value
			a.prev_avg = value // SValueFloat{value: avg, utime: utime, offline: 0}
			a.prev_save_in_db = value
			a.begin_utime = value.utime
			a.sum = 0
			return nil, []SValueFloat{
				{value: avg, utime: utime, offline: 0},
				s_p_value,
				value,
			}
		}
		a.prev_value = value
		a.prev_avg = value
		a.prev_save_in_db = value
		a.begin_utime = value.utime
		a.sum = 0
		return nil, []SValueFloat{
			value,
		}
	}
	if is_delta(value, a.prev_avg, float64(a.delta)) {
		err, is_calc, avg, utime := calc_avg_2(a.begin_utime, a.prev_value.utime, a.sum)
		if err == nil && is_calc {
			a.prev_value = value
			a.prev_avg = value // SValueFloat{value: avg, utime: utime, offline: 0}
			a.prev_save_in_db = value
			a.begin_utime = value.utime
			a.sum = 0
			return nil, []SValueFloat{
				{value: avg, utime: utime, offline: 0},
				value,
			}
		}
		a.prev_value = value
		a.prev_avg = value
		a.prev_save_in_db = value
		a.begin_utime = value.utime
		a.sum = 0
		return nil, []SValueFloat{
			value,
		}
	}
	if is_end_period(value.utime, a.begin_utime, int64(a.period)) {
		err, is_calc, avg, utime := calc_avg_2(a.begin_utime, a.prev_value.utime, a.sum)
		if err == nil && is_calc {
			a.prev_value = value
			a.prev_avg = SValueFloat{value: avg, utime: utime, offline: 0}
			a.prev_save_in_db = a.prev_avg
			a.begin_utime = value.utime
			a.sum = 0
			return nil, []SValueFloat{
				{value: avg, utime: utime, offline: 0},
			}
		}
		if reflect.DeepEqual(a.prev_value, a.prev_save_in_db) {
			a.prev_value = value
			a.begin_utime = value.utime
			a.sum = 0
			return nil, []SValueFloat{}
		}
		ret := []SValueFloat{
			a.prev_value,
		}
		a.prev_save_in_db = a.prev_value
		a.prev_avg = a.prev_value
		a.prev_value = value
		a.begin_utime = value.utime
		a.sum = 0
		return nil, ret
	}

	if value.utime != a.prev_value.utime {
		a.sum += ((a.prev_value.value + value.value) / 2) * float64(value.utime-a.prev_value.utime)
	}

	a.prev_value = value
	return nil, []SValueFloat{}
}
