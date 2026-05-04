package file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/rate"
)

// JsonBackend persists records in three flat JSON files (clients,
// tasks, hosts). Every mutation rewrites the corresponding file
// atomically. This is the original storage strategy; it is fast at
// small scale but slows down once the panel has thousands of bots
// because every write touches the whole file.
type JsonBackend struct {
	taskFile   string
	hostFile   string
	clientFile string

	mu sync.Mutex
}

func NewJsonBackend(runPath string) *JsonBackend {
	return &JsonBackend{
		taskFile:   filepath.Join(runPath, "conf", "tasks.json"),
		hostFile:   filepath.Join(runPath, "conf", "hosts.json"),
		clientFile: filepath.Join(runPath, "conf", "clients.json"),
	}
}

func (b *JsonBackend) Close() error { return nil }

func (b *JsonBackend) LoadAll(clients, tasks, hosts *sync.Map) (mc, mt, mh int32, err error) {
	loadFile(b.clientFile, func(s string) {
		c := new(Client)
		if json.Unmarshal([]byte(s), c) != nil {
			return
		}
		if c.RateLimit > 0 {
			c.Rate = rate.NewRate(int64(c.RateLimit * 1024))
		} else {
			c.Rate = rate.NewRate(int64(2 << 23))
		}
		c.Rate.Start()
		c.NowConn = 0
		clients.Store(c.Id, c)
		if int32(c.Id) > mc {
			mc = int32(c.Id)
		}
	})
	loadFile(b.taskFile, func(s string) {
		t := new(Tunnel)
		if json.Unmarshal([]byte(s), t) != nil {
			return
		}
		if t.Client != nil {
			if v, ok := clients.Load(t.Client.Id); ok {
				t.Client = v.(*Client)
			} else {
				return
			}
		}
		tasks.Store(t.Id, t)
		if int32(t.Id) > mt {
			mt = int32(t.Id)
		}
	})
	loadFile(b.hostFile, func(s string) {
		h := new(Host)
		if json.Unmarshal([]byte(s), h) != nil {
			return
		}
		if h.Client != nil {
			if v, ok := clients.Load(h.Client.Id); ok {
				h.Client = v.(*Client)
			} else {
				return
			}
		}
		hosts.Store(h.Id, h)
		if int32(h.Id) > mh {
			mh = int32(h.Id)
		}
	})
	return
}

func (b *JsonBackend) UpsertClient(c *Client) error {
	return b.flush(b.clientFile, GetDb().JsonDb.Clients)
}
func (b *JsonBackend) DeleteClient(id int) error {
	return b.flush(b.clientFile, GetDb().JsonDb.Clients)
}
func (b *JsonBackend) UpsertTask(t *Tunnel) error {
	return b.flush(b.taskFile, GetDb().JsonDb.Tasks)
}
func (b *JsonBackend) DeleteTask(id int) error {
	return b.flush(b.taskFile, GetDb().JsonDb.Tasks)
}
func (b *JsonBackend) UpsertHost(h *Host) error {
	return b.flush(b.hostFile, GetDb().JsonDb.Hosts)
}
func (b *JsonBackend) DeleteHost(id int) error {
	return b.flush(b.hostFile, GetDb().JsonDb.Hosts)
}
func (b *JsonBackend) FlushClients(m *sync.Map) error { return b.flush(b.clientFile, *m) }
func (b *JsonBackend) FlushTasks(m *sync.Map) error   { return b.flush(b.taskFile, *m) }
func (b *JsonBackend) FlushHosts(m *sync.Map) error   { return b.flush(b.hostFile, *m) }

// flush rewrites the whole file atomically (tmp + rename), as the
// original implementation did.
func (b *JsonBackend) flush(path string, m sync.Map) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	f, err := os.Create(path + ".tmp")
	if err != nil {
		return err
	}
	m.Range(func(_, value interface{}) bool {
		var data []byte
		switch v := value.(type) {
		case *Client:
			if v.NoStore {
				return true
			}
			data, _ = json.Marshal(v)
		case *Tunnel:
			if v.NoStore {
				return true
			}
			data, _ = json.Marshal(v)
		case *Host:
			if v.NoStore {
				return true
			}
			data, _ = json.Marshal(v)
		default:
			return true
		}
		if data == nil {
			return true
		}
		if _, err := f.Write(data); err != nil {
			return false
		}
		if _, err := f.Write([]byte("\n" + common.CONN_DATA_SEQ)); err != nil {
			return false
		}
		return true
	})
	_ = f.Sync()
	_ = f.Close()
	return os.Rename(path+".tmp", path)
}

func loadFile(path string, cb func(string)) {
	b, err := common.ReadAllFromFile(path)
	if err != nil {
		return
	}
	for _, v := range strings.Split(string(b), "\n"+common.CONN_DATA_SEQ) {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		cb(v)
	}
}
