package main

import "fmt"

type AVG struct {
	sum         float64
	p_value     float64
	p_utime     int64
	begin_utime int64
	period      int
	delta       float32
}

var (
	error_utime_zero error = fmt.Errorf("error utime is zero")
	error_not_found  error = fmt.Errorf("error not found")
)

func newAVG(period int, delta float32) *AVG {
	return &AVG{
		period: period,
		delta:  delta,
	}
}

func (a *AVG) add(value float64, utime int64) error {
	if utime == 0 {
		return error_utime_zero
	}
	if a.begin_utime == 0 {
		a.begin_utime = utime
	} else {
		a.sum += ((a.p_value + value) / 2) * float64(utime-a.p_utime)
	}
	a.p_value = value
	a.p_utime = utime
	return nil
}

func (a *AVG) calc_avg() (float64, int64, error) {
	if a.begin_utime == 0 {
		return 0, 0, error_not_found
	}
	avg := a.sum / float64(a.p_utime-a.begin_utime)
	return avg, (a.p_utime + a.begin_utime) / 2, nil
}

func (a *AVG) is_delta() bool {
	return false
}

func (a *AVG) is_end_period(utime int64) bool {
	if a.period == 0 {
		return true
	}
	// целочисленное деление на период используется для выравнивания точки сохранения, для сигналов с одинаковым периодом
	ptime := utime / int64(a.period)
	btime := a.begin_utime / int64(a.period)
	if ptime > btime {
		return true
	}
	return false
}

func (a *AVG) set_new_period() {
	a.begin_utime = 0
	a.p_utime = 0
	a.p_value = 0
	a.sum = 0
}
