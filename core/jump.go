package core

import (
	"encoding/gob"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/term"
)

const jumpKeywordName = "jump_keyword"

var jumpKeywordExpireSeconds = config.DaySeconds

type JumpKeywordStorage struct {
	data map[string]int64

	now int64

	lock sync.Mutex
}

func NewJumpKeywordStorage() (*JumpKeywordStorage, error) {
	s := &JumpKeywordStorage{
		data: make(map[string]int64),
		now:  time.Now().Unix(),
	}
	err := s.init()
	if err != nil {
		return nil, errors.Trace(err, "init jump keyword storage")
	}
	return s, nil
}

func (s *JumpKeywordStorage) isOutDate(seconds int64) bool {
	delta := s.now - seconds
	if delta <= 0 {
		return true
	}
	return delta >= jumpKeywordExpireSeconds
}

func (s *JumpKeywordStorage) Add(kw string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data[kw] = s.now
}

func (s *JumpKeywordStorage) List() []string {
	s.lock.Lock()
	defer s.lock.Unlock()

	kws := make([]string, 0, len(s.data))
	for kw, seconds := range s.data {
		if s.isOutDate(seconds) {
			delete(s.data, kw)
			continue
		}
		kws = append(kws, kw)
	}
	sort.Strings(kws)
	return kws
}

type parseJumpKeywordStorageError struct {
	path string
	err  error
}

func (err *parseJumpKeywordStorageError) Error() string {
	return err.err.Error()
}

func (err *parseJumpKeywordStorageError) Extra() {
	term.Println()
	term.Printf("The repository data is broken, please fix or delete it: %s", err.path)
}

func (s *JumpKeywordStorage) init() error {
	path := config.GetLocalDir(jumpKeywordName)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Trace(err, "open jump keyword file")
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&s.data)
	if err != nil {
		return &parseJumpKeywordStorageError{
			path: path,
			err:  err,
		}
	}
	return nil
}

func (s *JumpKeywordStorage) Close() error {
	if len(s.data) == 0 {
		return nil
	}
	path := config.GetLocalDir(jumpKeywordName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Trace(err, "open data file")
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(s.data)
}
