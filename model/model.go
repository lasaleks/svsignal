package model

import (
	"log"

	"gorm.io/gorm"
)

type Group struct {
	ID      uint `gorm:"primarykey"`
	Key     string
	Name    string
	Signals []*Signal
}

type Signal struct {
	ID       uint `gorm:"primarykey"`
	GroupID  uint
	Key      string
	Name     string
	TypeSave int8
	Period   int
	Delta    float32
	Tags     []Tag
	//MValues  []MValue
}

type MValue struct {
	ID       uint  `gorm:"primarykey"`
	SignalID uint  `gorm:"index:idx_m_sig_utime,priority:2;"`
	UTime    int64 `gorm:"index:idx_m_sig_utime,priority:1;column:utime;"`
	Max      float32
	Min      float32
	Mean     float32
	Median   float32
	OffLine  bool
}

type IValue struct {
	ID       uint  `gorm:"primarykey"`
	SignalID uint  `gorm:"index:idx_i_sig_utime,priority:2;"`
	UTime    int64 `gorm:"index:idx_i_sig_utime,priority:1;column:utime;"`
	Value    int32
	OffLine  bool
}

type FValue struct {
	ID       uint  `gorm:"primarykey"`
	SignalID uint  `gorm:"index:idx_f_sig_utime,priority:2;"`
	UTime    int64 `gorm:"index:idx_f_sig_utime,priority:1;column:utime;"`
	Value    float64
	OffLine  bool
}

type Tag struct {
	ID       uint `gorm:"primarykey"`
	SignalID uint
	Tag      string
	Value    string
}

func (Tag) TableName() string {
	return "tags"
}

func Migrate(db *gorm.DB) {
	if err := db.AutoMigrate(&Group{}, &Signal{}, &MValue{}, &IValue{}, &FValue{}, &Tag{}); err != nil {
		log.Panicln(err)
	}
}
