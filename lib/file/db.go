package file

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/crypt"
	"ehang.io/nps/lib/rate"
)

type DbUtils struct {
	JsonDb *JsonDb
}

var (
	Db   *DbUtils
	once sync.Once
)

//init csv from file
func GetDb() *DbUtils {
	once.Do(func() {
		jsonDb := NewJsonDb(common.GetRunPath())
		jsonDb.LoadClientFromJsonFile()
		jsonDb.LoadTaskFromJsonFile()
		jsonDb.LoadHostFromJsonFile()
		Db = &DbUtils{JsonDb: jsonDb}
	})
	return Db
}

// GetMapKeys returns the int keys of m, optionally sorted by a flow
// field. The pointer signature is mandatory because sync.Map cannot
// be copied by value once it has been used.
func GetMapKeys(m *sync.Map, isSort bool, sortKey, order string) (keys []int) {
	if m == nil {
		return
	}
	if sortKey != "" && isSort {
		return sortClientByKey(m, sortKey, order)
	}
	m.Range(func(key, value interface{}) bool {
		if k, ok := key.(int); ok {
			keys = append(keys, k)
		}
		return true
	})
	sort.Ints(keys)
	return
}

func (s *DbUtils) GetClientList(start, length int, search, sort, order string, clientId int) ([]*Client, int) {
	list := make([]*Client, 0)
	var cnt int
	keys := GetMapKeys(&s.JsonDb.Clients, true, sort, order)
	for _, key := range keys {
		if value, ok := s.JsonDb.Clients.Load(key); ok {
			v := value.(*Client)
			if v.NoDisplay {
				continue
			}
			if clientId != 0 && clientId != v.Id {
				continue
			}
			if search != "" && !(v.Id == common.GetIntNoErrByStr(search) || strings.Contains(v.VerifyKey, search) || strings.Contains(v.Remark, search)) {
				continue
			}
			cnt++
			if start--; start < 0 {
				if length--; length >= 0 {
					list = append(list, v)
				}
			}
		}
	}
	return list, cnt
}

// GetIdByVerifyKey is kept for backwards compatibility (legacy
// secret/p2p flows still call it). It looks up a client by the
// historical md5(VerifyKey) token.
func (s *DbUtils) GetIdByVerifyKey(vKey string, addr string) (id int, err error) {
	var exist bool
	s.JsonDb.Clients.Range(func(key, value interface{}) bool {
		v := value.(*Client)
		if common.Getverifyval(v.VerifyKey) == vKey && v.Status {
			v.Addr = common.GetIpByAddr(addr)
			id = v.Id
			exist = true
			return false
		}
		return true
	})
	if exist {
		return
	}
	return 0, errors.New("not found")
}

// GetClientIdByMachineGuid looks up a bot by the MachineGuid the
// host reported. The lookup is in-memory; the Postgres backend has
// a matching index but we never go through it on the hot path.
func (s *DbUtils) GetClientIdByMachineGuid(guid string) (id int, ok bool) {
	if guid == "" {
		return 0, false
	}
	s.JsonDb.Clients.Range(func(_, value interface{}) bool {
		v, isClient := value.(*Client)
		if !isClient || v == nil {
			return true
		}
		if v.MachineGuid == guid {
			id = v.Id
			ok = true
			return false
		}
		return true
	})
	return
}

func (s *DbUtils) NewTask(t *Tunnel) (err error) {
	s.JsonDb.Tasks.Range(func(key, value interface{}) bool {
		v := value.(*Tunnel)
		if (v.Mode == "secret" || v.Mode == "p2p") && v.Password == t.Password {
			err = errors.New(fmt.Sprintf("secret mode keys %s must be unique", t.Password))
			return false
		}
		return true
	})
	if err != nil {
		return
	}
	t.Flow = new(Flow)
	s.JsonDb.Tasks.Store(t.Id, t)
	if b := GetBackend(); b != nil {
		if err = b.UpsertTask(t); err != nil {
			s.JsonDb.Tasks.Delete(t.Id)
		}
	}
	return
}

func (s *DbUtils) UpdateTask(t *Tunnel) error {
	s.JsonDb.Tasks.Store(t.Id, t)
	if b := GetBackend(); b != nil {
		return b.UpsertTask(t)
	}
	return nil
}

func (s *DbUtils) DelTask(id int) error {
	s.JsonDb.Tasks.Delete(id)
	if b := GetBackend(); b != nil {
		return b.DeleteTask(id)
	}
	return nil
}

//md5 password
func (s *DbUtils) GetTaskByMd5Password(p string) (t *Tunnel) {
	s.JsonDb.Tasks.Range(func(key, value interface{}) bool {
		if crypt.Md5(value.(*Tunnel).Password) == p {
			t = value.(*Tunnel)
			return false
		}
		return true
	})
	return
}

func (s *DbUtils) GetTask(id int) (t *Tunnel, err error) {
	if v, ok := s.JsonDb.Tasks.Load(id); ok {
		t = v.(*Tunnel)
		return
	}
	err = errors.New("not found")
	return
}

func (s *DbUtils) DelHost(id int) error {
	s.JsonDb.Hosts.Delete(id)
	if b := GetBackend(); b != nil {
		return b.DeleteHost(id)
	}
	return nil
}

func (s *DbUtils) IsHostExist(h *Host) bool {
	var exist bool
	s.JsonDb.Hosts.Range(func(key, value interface{}) bool {
		v := value.(*Host)
		if v.Id != h.Id && v.Host == h.Host && h.Location == v.Location && (v.Scheme == "all" || v.Scheme == h.Scheme) {
			exist = true
			return false
		}
		return true
	})
	return exist
}

func (s *DbUtils) NewHost(t *Host) error {
	if t.Location == "" {
		t.Location = "/"
	}
	if s.IsHostExist(t) {
		return errors.New("host has exist")
	}
	t.Flow = new(Flow)
	s.JsonDb.Hosts.Store(t.Id, t)
	if b := GetBackend(); b != nil {
		if err := b.UpsertHost(t); err != nil {
			s.JsonDb.Hosts.Delete(t.Id)
			return err
		}
	}
	return nil
}

func (s *DbUtils) GetHost(start, length int, id int, search string) ([]*Host, int) {
	list := make([]*Host, 0)
	var cnt int
	keys := GetMapKeys(&s.JsonDb.Hosts, false, "", "")
	for _, key := range keys {
		if value, ok := s.JsonDb.Hosts.Load(key); ok {
			v := value.(*Host)
			if search != "" && !(v.Id == common.GetIntNoErrByStr(search) || strings.Contains(v.Host, search) || strings.Contains(v.Remark, search)) {
				continue
			}
			if id == 0 || (v.Client != nil && v.Client.Id == id) {
				cnt++
				if start--; start < 0 {
					if length--; length >= 0 {
						list = append(list, v)
					}
				}
			}
		}
	}
	return list, cnt
}

func (s *DbUtils) DelClient(id int) error {
	s.JsonDb.Clients.Delete(id)
	if b := GetBackend(); b != nil {
		return b.DeleteClient(id)
	}
	return nil
}

func (s *DbUtils) NewClient(c *Client) error {
	if c.WebUserName != "" && !s.VerifyUserName(c.WebUserName, c.Id) {
		return errors.New("web login username duplicate, please reset")
	}
	// Pick a unique vkey before allocating any rate limiter so we
	// never spawn extra goroutines on the retry path.
	if c.VerifyKey == "" {
		const maxAttempts = 16
		for i := 0; i < maxAttempts; i++ {
			c.VerifyKey = crypt.GetRandomString(16)
			if s.VerifyVkey(c.VerifyKey, c.Id) {
				break
			}
			if i == maxAttempts-1 {
				return errors.New("could not allocate a unique vkey")
			}
		}
	} else if !s.VerifyVkey(c.VerifyKey, c.Id) {
		return errors.New("Vkey duplicate, please reset")
	}
	if c.Rate == nil {
		if c.RateLimit == 0 {
			c.Rate = rate.NewRate(int64(2 << 23))
		} else {
			c.Rate = rate.NewRate(int64(c.RateLimit * 1024))
		}
		c.Rate.Start()
	}
	if c.Id == 0 {
		c.Id = int(s.JsonDb.GetClientId())
	}
	if c.Flow == nil {
		c.Flow = new(Flow)
	}
	s.JsonDb.Clients.Store(c.Id, c)
	if b := GetBackend(); b != nil {
		if err := b.UpsertClient(c); err != nil {
			s.JsonDb.Clients.Delete(c.Id)
			return err
		}
	}
	return nil
}

func (s *DbUtils) VerifyVkey(vkey string, id int) (res bool) {
	res = true
	s.JsonDb.Clients.Range(func(key, value interface{}) bool {
		v := value.(*Client)
		if v.VerifyKey == vkey && v.Id != id {
			res = false
			return false
		}
		return true
	})
	return res
}

func (s *DbUtils) VerifyUserName(username string, id int) (res bool) {
	res = true
	s.JsonDb.Clients.Range(func(key, value interface{}) bool {
		v := value.(*Client)
		if v.WebUserName == username && v.Id != id {
			res = false
			return false
		}
		return true
	})
	return res
}

func (s *DbUtils) UpdateClient(t *Client) error {
	s.JsonDb.Clients.Store(t.Id, t)
	if t.Rate == nil {
		if t.RateLimit == 0 {
			t.Rate = rate.NewRate(int64(2 << 23))
		} else {
			t.Rate = rate.NewRate(int64(t.RateLimit * 1024))
		}
		t.Rate.Start()
	}
	if b := GetBackend(); b != nil {
		return b.UpsertClient(t)
	}
	return nil
}

func (s *DbUtils) IsPubClient(id int) bool {
	client, err := s.GetClient(id)
	if err == nil {
		return client.NoDisplay
	}
	return false
}

func (s *DbUtils) GetClient(id int) (c *Client, err error) {
	if v, ok := s.JsonDb.Clients.Load(id); ok {
		c = v.(*Client)
		return
	}
	err = errors.New("client not found")
	return
}

func (s *DbUtils) GetClientIdByVkey(vkey string) (id int, err error) {
	var exist bool
	s.JsonDb.Clients.Range(func(key, value interface{}) bool {
		v := value.(*Client)
		if crypt.Md5(v.VerifyKey) == vkey {
			exist = true
			id = v.Id
			return false
		}
		return true
	})
	if exist {
		return
	}
	err = errors.New("client not found")
	return
}

func (s *DbUtils) GetHostById(id int) (h *Host, err error) {
	if v, ok := s.JsonDb.Hosts.Load(id); ok {
		h = v.(*Host)
		return
	}
	err = errors.New("The host could not be parsed")
	return
}

//get key by host from x
func (s *DbUtils) GetInfoByHost(host string, r *http.Request) (h *Host, err error) {
	var hosts []*Host
	//Handling Ported Access
	host = common.GetIpByAddr(host)
	s.JsonDb.Hosts.Range(func(key, value interface{}) bool {
		v := value.(*Host)
		if v.IsClose {
			return true
		}
		//Remove http(s) http(s)://a.proxy.com
		//*.proxy.com *.a.proxy.com  Do some pan-parsing
		if v.Scheme != "all" && v.Scheme != r.URL.Scheme {
			return true
		}
		tmpHost := v.Host
		if strings.Contains(tmpHost, "*") {
			tmpHost = strings.Replace(tmpHost, "*", "", -1)
			if strings.Contains(host, tmpHost) {
				hosts = append(hosts, v)
			}
		} else if v.Host == host {
			hosts = append(hosts, v)
		}
		return true
	})

	for _, v := range hosts {
		//If not set, default matches all
		if v.Location == "" {
			v.Location = "/"
		}
		if strings.Index(r.RequestURI, v.Location) == 0 {
			if h == nil || (len(v.Location) > len(h.Location)) {
				h = v
			}
		}
	}
	if h != nil {
		return
	}
	err = errors.New("The host could not be parsed")
	return
}
