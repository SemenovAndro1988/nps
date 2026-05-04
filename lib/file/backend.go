// Package file currently keeps every record in two layers:
//
//  1. An in-memory `sync.Map` keyed by record id. This is the hot
//     path: the bridge, the SOCKS5 entry and the panel all read from
//     here without blocking each other.
//  2. A persistence backend (defined here) that the in-memory layer
//     flushes to whenever something changes. Two backends are
//     supported:
//
//     - `jsonBackend` writes the maps to plain JSON files
//       (the historical behaviour).
//     - `pgBackend` writes them to Postgres tables with a connection
//       pool. This scales much better when there are thousands of
//       bots because every mutation only touches a single row instead
//       of rewriting the whole file.
//
// The choice is driven by the `db_driver` value in `nps.conf`: empty
// or `json` falls back to JSON files, `postgres`/`postgresql`/`pg`
// switches to Postgres.
package file

import (
	"sync"
)

// Backend defines the operations every persistence backend must
// implement. Mutations are always called from outside the in-memory
// map so the backend just needs to durably persist the current state
// of one record (or, for *All, dump the whole map).
type Backend interface {
	// LoadAll fills the provided sync.Maps with whatever the backend
	// currently knows about. Implementations also report the largest
	// id they observed so the in-memory id sequence can be advanced.
	LoadAll(clients, tasks, hosts *sync.Map) (maxClientId, maxTaskId, maxHostId int32, err error)

	UpsertClient(c *Client) error
	DeleteClient(id int) error

	UpsertTask(t *Tunnel) error
	DeleteTask(id int) error

	UpsertHost(h *Host) error
	DeleteHost(id int) error

	// FlushClients / FlushTasks / FlushHosts are called by the
	// periodic flow snapshot (server.flowSession) to checkpoint the
	// current flow counters. JSON backend does a full rewrite, the
	// Postgres backend updates rows that changed.
	FlushClients(m *sync.Map) error
	FlushTasks(m *sync.Map) error
	FlushHosts(m *sync.Map) error

	Close() error
}

// activeBackend is the persistence backend chosen at process start.
// Tests that need to inject a custom backend can call SetBackend
// before GetDb is invoked for the first time.
var (
	activeBackend Backend
	backendMu     sync.RWMutex
)

func SetBackend(b Backend) {
	backendMu.Lock()
	defer backendMu.Unlock()
	activeBackend = b
}

func GetBackend() Backend {
	backendMu.RLock()
	defer backendMu.RUnlock()
	return activeBackend
}
