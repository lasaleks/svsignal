package svsignalsave

import (
	"encoding/json"
	"fmt"

	"github.com/lasaleks/gormq"
)

type Tags struct {
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

type ValueSignal struct {
	Value   float64 `json:"value"`
	UTime   int64   `json:"utime"`
	Offline int64   `json:"offline"`
}

const (
	SVS_TYPE_SAVE_DISCREET = 1
	SVS_TYPE_SAVE_AVG      = 2
)

func PubSaveValue(out chan<- gormq.MessageAmpq, signal_key string, value float64, utime int64, offline int) error {
	data, err := json.Marshal(&ValueSignal{
		Value:   value,
		UTime:   utime,
		Offline: int64(offline),
	})
	if err != nil {
		return err
	}
	out <- gormq.MessageAmpq{
		Exchange:     "svsignal",
		Routing_key:  fmt.Sprintf("svs.save.%s", signal_key),
		Content_type: "text/plain",
		Data:         data,
	}
	return nil
}

type SetSignal struct {
	group_key  string
	signal_key string
	TypeSave   int     `json:"typesave"`
	Period     int     `json:"period"`
	Delta      float32 `json:"delta"`
	Name       string  `json:"name"`
	Tags       []Tags  `json:"tags"`
}

func PubSetSVSignal(out chan<- gormq.MessageAmpq, signal_key string, typesave int, period int, delta float32, name string, tags []Tags) error {
	data, err := json.Marshal(&SetSignal{
		TypeSave: typesave,
		Period:   period,
		Delta:    delta,
		Name:     name,
		Tags:     tags,
	})
	if err != nil {
		return err
	}

	out <- gormq.MessageAmpq{
		Exchange:     "svsignal",
		Routing_key:  fmt.Sprintf("svs.set.%s", signal_key),
		Content_type: "text/plain",
		Data:         data,
	}
	return nil
}
