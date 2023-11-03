package main

import (
	"context"
	"fmt"

	goutils "github.com/lasaleks/go-utils"
	svsignal "github.com/lasaleks/svsignal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// func (m *Beacons) GetStateBeacon(ctx context.Context, addr *sbeacon.MsgRequestStateBeacon) (*sbeacon.MsgStateBeacon, error) {
func (s *SVSignalDB) CreateGroup(ctx context.Context, gr *svsignal.Group) (*svsignal.Nothing, error) {
	return nil, nil
}

func (s *SVSignalDB) UpdateGroup(ctx context.Context, gr *svsignal.Group) (*svsignal.Nothing, error) {
	return nil, nil
}

func (s *SVSignalDB) DeleteGroup(ctx context.Context, gr *svsignal.Group) (*svsignal.Nothing, error) {
	return nil, nil
}

func (s *SVSignalDB) GetAllGroup(ctx context.Context, _ *svsignal.Nothing) (*svsignal.Groups, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := svsignal.Groups{Groups: make(map[string]*svsignal.Group)}
	for _, group := range s.group_key {
		res.Groups[group.Key] = &svsignal.Group{Key: group.Key, Name: group.Name}
	}
	return &res, nil
}

func (s *SVSignalDB) GetGroup(ctx context.Context, key *svsignal.MsgKey) (*svsignal.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	group, found := s.group_key[key.Key]
	if !found {
		return nil, status.Errorf(codes.NotFound, "not found")
	}

	return &svsignal.Group{Key: group.Key, Name: group.Name}, nil
}

func (s *SVSignalDB) CreateSignal(ctx context.Context, _ *svsignal.Signal) (*svsignal.Nothing, error) {
	return nil, nil
}

func (s *SVSignalDB) UpdateSignal(ctx context.Context, _ *svsignal.Signal) (*svsignal.Nothing, error) {
	return nil, nil
}

func (s *SVSignalDB) DeleteSignal(ctx context.Context, _ *svsignal.Signal) (*svsignal.Nothing, error) {
	return nil, nil
}

func (s *SVSignalDB) GetSignals(ctx context.Context, key *svsignal.MsgKey) (*svsignal.Signals, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	group, found := s.group_key[key.Key]
	if !found {
		return nil, status.Errorf(codes.NotFound, "not found group")
	}
	res := svsignal.Signals{Signals: make(map[string]*svsignal.Signal, len(group.Signals))}
	for _, sig := range group.Signals {
		s := &svsignal.Signal{
			Key:      sig.Key,
			GroupKey: group.Key,
			Name:     group.Name,
			TypeSave: svsignal.TypeSignal(sig.TypeSave),
			Period:   int32(sig.Period),
			Delta:    sig.Delta,
			Tags:     make([]*svsignal.Tag, len(sig.Tags)),
		}

		for i := 0; i < len(sig.Tags); i++ {
			s.Tags[i] = &svsignal.Tag{Tag: sig.Tags[i].Tag, Value: sig.Tags[i].Value}
		}
		res.Signals[sig.Key] = s
	}

	return &res, nil
}

func (s *SVSignalDB) GetSignal(ctx context.Context, key *svsignal.MsgKey) (*svsignal.Signal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sig, found := s.signal_key[key.Key]
	if !found {
		return nil, status.Errorf(codes.NotFound, "not found signal")
	}
	group, found := s.group_id[sig.GroupID]
	if !found {
		return nil, status.Errorf(codes.NotFound, "not found group")
	}

	res := svsignal.Signal{
		Key:      sig.Key,
		GroupKey: group.Key,
		Name:     group.Name,
		TypeSave: svsignal.TypeSignal(sig.TypeSave),
		Period:   int32(sig.Period),
		Delta:    sig.Delta,
		Tags:     make([]*svsignal.Tag, len(sig.Tags)),
	}

	for i := 0; i < len(sig.Tags); i++ {
		res.Tags[i] = &svsignal.Tag{Tag: sig.Tags[i].Tag, Value: sig.Tags[i].Value}
	}
	return &res, nil
}

func (s *SVSignalDB) CreateTag(ctx context.Context, query *svsignal.SignalTag) (*svsignal.Nothing, error) {
	return nil, nil
}

func (s *SVSignalDB) UpdateTag(ctx context.Context, query *svsignal.SignalTag) (*svsignal.Nothing, error) {
	return nil, nil
}

func (s *SVSignalDB) DeleteTag(ctx context.Context, query *svsignal.SignalTag) (*svsignal.Nothing, error) {
	return nil, nil
}

func (s *SVSignalDB) GetData(ctx context.Context, request *svsignal.RequestData) (*svsignal.ResponseData, error) {
	var CH_RESPONSE chan interface{} = make(chan interface{}, 1)
	CH_REQUEST_DATA <- RequestData{
		begin:       request.Begin,
		end:         request.End,
		key:         request.SignalKey,
		CH_RESPONSE: CH_RESPONSE,
	}

	resp := <-CH_RESPONSE
	switch response := resp.(type) {
	case error:
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("%v", response))
	case ResponseDataSignalT1:
		res := svsignal.ResponseData{TypeSignal: svsignal.TypeSignal_TYPE_IVALUE, IValues: make([]*svsignal.IValue, len(response.Values))}
		for i := 0; i < len(response.Values); i++ {
			res.IValues[i] = &svsignal.IValue{UTime: response.Values[i][1], Value: int32(response.Values[i][2]), NULL: goutils.IntToBool(int(response.Values[i][3]))}
		}
		return &res, nil
	case ResponseDataSignalT2:
		res := svsignal.ResponseData{TypeSignal: svsignal.TypeSignal_TYPE_FVALUE, FValues: make([]*svsignal.FValue, len(response.Values))}
		for i := 0; i < len(response.Values); i++ {
			res.FValues[i] = &svsignal.FValue{UTime: response.Values[i][1].(int64), Value: response.Values[i][2].(float64), NULL: goutils.IntToBool(int(response.Values[i][3].(int64)))}
		}
		return &res, nil
	}
	/*
		sig, found := s.signal_key[key.Key]
		if !found {
			return nil, status.Errorf(codes.NotFound, "not found signal")
		}*/
	return nil, nil
}

func (s *SVSignalDB) SaveValue(ctx context.Context, query *svsignal.MsgSaveValue) (*svsignal.Nothing, error) {
	return nil, nil
}
