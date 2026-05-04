package file

import (
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"

	"ehang.io/nps/lib/common"
)

// newSyncMaps returns a fresh trio of sync.Maps used by the
// import-from-json helper.
func newSyncMaps() (*sync.Map, *sync.Map, *sync.Map) {
	return &sync.Map{}, &sync.Map{}, &sync.Map{}
}

func NewJsonDb(runPath string) *JsonDb {
	return &JsonDb{
		RunPath:        runPath,
		TaskFilePath:   filepath.Join(runPath, "conf", "tasks.json"),
		HostFilePath:   filepath.Join(runPath, "conf", "hosts.json"),
		ClientFilePath: filepath.Join(runPath, "conf", "clients.json"),
	}
}

// JsonDb keeps every record in three sync.Maps and delegates
// persistence to the configured Backend. The struct still carries the
// name "JsonDb" for backwards compatibility with the rest of the
// codebase, but it transparently supports any backend (json files,
// postgres, ...).
type JsonDb struct {
	Tasks            sync.Map
	Hosts            sync.Map
	HostsTmp         sync.Map
	Clients          sync.Map
	RunPath          string
	ClientIncreaseId int32
	TaskIncreaseId   int32
	HostIncreaseId   int32
	TaskFilePath     string
	HostFilePath     string
	ClientFilePath   string
}

func (s *JsonDb) backend() Backend {
	if b := GetBackend(); b != nil {
		return b
	}
	// Fallback to JSON files when nothing has been configured yet.
	b := NewJsonBackend(s.RunPath)
	SetBackend(b)
	return b
}

// LoadTaskFromJsonFile / LoadClientFromJsonFile / LoadHostFromJsonFile
// retain their historical names but in practice always go through the
// active backend. The first call also seeds the id counters from the
// stored max ids. loadAll is idempotent thanks to loadOnce; if the
// load fails the error is exposed through LoadError() so the
// boot-time caller can fail loudly.
var (
	loadOnce sync.Once
	loadMu   sync.Mutex
	loadErr  error
)

func (s *JsonDb) loadAll() {
	loadOnce.Do(func() {
		mc, mt, mh, err := s.backend().LoadAll(&s.Clients, &s.Tasks, &s.Hosts)
		loadMu.Lock()
		defer loadMu.Unlock()
		if err != nil {
			loadErr = err
			return
		}
		if mc > s.ClientIncreaseId {
			s.ClientIncreaseId = mc
		}
		if mt > s.TaskIncreaseId {
			s.TaskIncreaseId = mt
		}
		if mh > s.HostIncreaseId {
			s.HostIncreaseId = mh
		}
	})
}

func (s *JsonDb) LoadTaskFromJsonFile()   { s.loadAll() }
func (s *JsonDb) LoadClientFromJsonFile() { s.loadAll() }
func (s *JsonDb) LoadHostFromJsonFile()   { s.loadAll() }

// LoadError returns the error from the initial load, if any.
func (s *JsonDb) LoadError() error {
	loadMu.Lock()
	defer loadMu.Unlock()
	return loadErr
}

func (s *JsonDb) GetClient(id int) (c *Client, err error) {
	if v, ok := s.Clients.Load(id); ok {
		c = v.(*Client)
		return
	}
	err = errors.New("client not found")
	return
}

var hostLock sync.Mutex
var taskLock sync.Mutex
var clientLock sync.Mutex

// StoreHostToJsonFile flushes every host to the active backend.
func (s *JsonDb) StoreHostToJsonFile() {
	hostLock.Lock()
	defer hostLock.Unlock()
	if err := s.backend().FlushHosts(&s.Hosts); err != nil {
		log("flush hosts: %s", err)
	}
}

func (s *JsonDb) StoreTasksToJsonFile() {
	taskLock.Lock()
	defer taskLock.Unlock()
	if err := s.backend().FlushTasks(&s.Tasks); err != nil {
		log("flush tasks: %s", err)
	}
}

func (s *JsonDb) StoreClientsToJsonFile() {
	clientLock.Lock()
	defer clientLock.Unlock()
	if err := s.backend().FlushClients(&s.Clients); err != nil {
		log("flush clients: %s", err)
	}
}

func (s *JsonDb) GetClientId() int32 { return atomic.AddInt32(&s.ClientIncreaseId, 1) }
func (s *JsonDb) GetTaskId() int32   { return atomic.AddInt32(&s.TaskIncreaseId, 1) }
func (s *JsonDb) GetHostId() int32   { return atomic.AddInt32(&s.HostIncreaseId, 1) }

// log avoids importing beego logs (which would be a circular import)
// and falls back to the standard log package if no logger is set.
func log(format string, args ...interface{}) {
	common.Log("[file] "+format, args...)
}
