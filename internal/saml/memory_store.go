package saml

import (
	"encoding/json"
	"os"
	"strings"
	"sync"

	"github.com/seriousben/dev-identity-provider/internal/saml/samlidp"
	"github.com/seriousben/dev-identity-provider/internal/storage"
)

// MemoryStore is an implementation of Store that resides completely
// in memory.
type MemoryStore struct {
	storage Storage
	mu      sync.RWMutex
	data    map[string]string
}

// Get fetches the data stored in `key` and unmarshals it into `value`.
func (s *MemoryStore) Get(key string, value interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var (
		v  string
		ok bool
	)

	if ks := strings.Split(key, "/users/"); len(ks) == 2 {
		u, err := s.storage.GetUserByID(ks[1])
		if err == os.ErrNotExist {
			return samlidp.ErrNotFound
		}
		b, err := json.Marshal(u)
		if err != nil {
			return err
		}
		v = string(b)
	} else if ks := strings.Split(key, "/services/"); len(ks) == 2 {
		u, err := s.storage.GetServiceProviderByID(ks[1])
		if err == os.ErrNotExist {
			return samlidp.ErrNotFound
		}
		b, err := json.Marshal(u)
		if err != nil {
			return err
		}
		v = string(b)
	} else if ks := strings.Split(key, "/services-by-entity-id/"); len(ks) == 2 {
		u, err := s.storage.GetServiceProviderByEntityID(ks[1])
		if err == os.ErrNotExist {
			return samlidp.ErrNotFound
		}
		b, err := json.Marshal(u)
		if err != nil {
			return err
		}
		v = string(b)
	} else {
		v, ok = s.data[key]
		if !ok {
			return samlidp.ErrNotFound
		}
	}
	return json.Unmarshal([]byte(v), value)
}

// Put marshals `value` and stores it in `key`.
func (s *MemoryStore) Put(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data == nil {
		s.data = map[string]string{}
	}

	if ks := strings.Split(key, "/users/"); len(ks) == 2 {
		su := value.(*storage.User)
		err := s.storage.PutUser(ks[1], su)
		if err == os.ErrNotExist {
			return samlidp.ErrNotFound
		}
		if err != nil {
			return err
		}
		return nil
	} else if ks := strings.Split(key, "/services/"); len(ks) == 2 {
		sp := value.(*storage.ServiceProvider)
		sp.ID = ks[1]
		err := s.storage.PutServiceProvider(ks[1], sp)
		if err == os.ErrNotExist {
			return samlidp.ErrNotFound
		}
		if err != nil {
			return err
		}
		return nil
	}
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}

	s.data[key] = string(buf)
	return nil
}

// Delete removes `key`
func (s *MemoryStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ks := strings.Split(key, "/users/"); len(ks) == 2 {
		return s.storage.DeleteUser(ks[1])
	} else if ks := strings.Split(key, "/services/"); len(ks) == 2 {
		return s.storage.DeleteServiceProvider(ks[1])
	}
	delete(s.data, key)
	return nil
}

// List returns all the keys that start with `prefix`. The prefix is
// stripped from each returned value. So if keys are ["aa", "ab", "cd"]
// then List("a") would produce []string{"a", "b"}
func (s *MemoryStore) List(prefix string) ([]string, error) {
	switch prefix {
	case "/users/":
		us, err := s.storage.ListUsers()
		if err != nil {
			return nil, err
		}
		rv := make([]string, len(us))
		for i, v := range us {
			rv[i] = v.ID
		}
		return rv, nil
	case "/services/":
		sps, err := s.storage.ListServiceProviders()
		if err != nil {
			return nil, err
		}
		rv := make([]string, len(sps))
		for i, v := range sps {
			rv[i] = v.ID
		}
		return rv, nil
	}
	rv := []string{}
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			rv = append(rv, strings.TrimPrefix(k, prefix))
		}
	}
	return rv, nil
}
