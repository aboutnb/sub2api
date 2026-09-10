package admin

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type adminSettingRepoStub struct {
	values map[string]string
}

func (s *adminSettingRepoStub) Get(context.Context, string) (*service.Setting, error) {
	panic("unexpected Get call")
}

func (s *adminSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if val, ok := s.values[key]; ok {
		return val, nil
	}
	return "", service.ErrSettingNotFound
}

func (s *adminSettingRepoStub) Set(_ context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *adminSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if val, ok := s.values[key]; ok {
			out[key] = val
		}
	}
	return out, nil
}

func (s *adminSettingRepoStub) SetMultiple(_ context.Context, values map[string]string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	for key, value := range values {
		s.values[key] = value
	}
	return nil
}

func (s *adminSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	return s.values, nil
}

func (s *adminSettingRepoStub) Delete(_ context.Context, key string) error {
	delete(s.values, key)
	return nil
}
