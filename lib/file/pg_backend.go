package file

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"ehang.io/nps/lib/rate"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgBackend stores clients, tasks and hosts in a PostgreSQL database.
// The whole record is serialised to JSONB so we don't need a wide
// schema migration every time the in-memory structs evolve. We keep
// a small set of generated columns that are useful for indexing /
// listing / search:
//
//	clients(id, remark, verify_key, socks5_user, socks5_pass, no_store,
//	        no_display, status, addr, data jsonb)
//	tasks(id, client_id, no_store, mode, port, password, data jsonb)
//	hosts(id, client_id, no_store, host, location, scheme, data jsonb)
type PgBackend struct {
	pool *pgxpool.Pool
}

const pgSchema = `
CREATE TABLE IF NOT EXISTS clients (
    id           INTEGER PRIMARY KEY,
    remark       TEXT       NOT NULL DEFAULT '',
    verify_key   TEXT       NOT NULL DEFAULT '',
    socks5_user  TEXT       NOT NULL DEFAULT '',
    socks5_pass  TEXT       NOT NULL DEFAULT '',
    no_store     BOOLEAN    NOT NULL DEFAULT FALSE,
    no_display   BOOLEAN    NOT NULL DEFAULT FALSE,
    status       BOOLEAN    NOT NULL DEFAULT TRUE,
    addr         TEXT       NOT NULL DEFAULT '',
    data         JSONB      NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_clients_socks5 ON clients (socks5_user, socks5_pass);
CREATE INDEX IF NOT EXISTS idx_clients_verify_key ON clients (verify_key);

CREATE TABLE IF NOT EXISTS tasks (
    id        INTEGER PRIMARY KEY,
    client_id INTEGER,
    no_store  BOOLEAN NOT NULL DEFAULT FALSE,
    mode      TEXT    NOT NULL DEFAULT '',
    port      INTEGER NOT NULL DEFAULT 0,
    password  TEXT    NOT NULL DEFAULT '',
    data      JSONB   NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_tasks_client_id ON tasks (client_id);

CREATE TABLE IF NOT EXISTS hosts (
    id        INTEGER PRIMARY KEY,
    client_id INTEGER,
    no_store  BOOLEAN NOT NULL DEFAULT FALSE,
    host      TEXT    NOT NULL DEFAULT '',
    location  TEXT    NOT NULL DEFAULT '/',
    scheme    TEXT    NOT NULL DEFAULT 'all',
    data      JSONB   NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_hosts_client_id ON hosts (client_id);
`

// NewPgBackend opens a connection pool against the given DSN and
// applies the schema if it has not been created yet.
func NewPgBackend(dsn string) (*PgBackend, error) {
	if dsn == "" {
		return nil, errors.New("empty postgres dsn")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	if cfg.MaxConns < 4 {
		cfg.MaxConns = 16
	}
	if cfg.MaxConnLifetime == 0 {
		cfg.MaxConnLifetime = time.Hour
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if _, err := pool.Exec(ctx, pgSchema); err != nil {
		pool.Close()
		return nil, err
	}
	return &PgBackend{pool: pool}, nil
}

func (b *PgBackend) Close() error {
	if b == nil || b.pool == nil {
		return nil
	}
	b.pool.Close()
	return nil
}

func (b *PgBackend) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func (b *PgBackend) LoadAll(clients, tasks, hosts *sync.Map) (mc, mt, mh int32, err error) {
	ctx, cancel := b.ctx()
	defer cancel()

	rows, err := b.pool.Query(ctx, `SELECT data FROM clients`)
	if err != nil {
		return
	}
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			rows.Close()
			return
		}
		c := new(Client)
		if json.Unmarshal(raw, c) != nil {
			continue
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
	}
	rows.Close()

	rows, err = b.pool.Query(ctx, `SELECT data FROM tasks`)
	if err != nil {
		return
	}
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			rows.Close()
			return
		}
		t := new(Tunnel)
		if json.Unmarshal(raw, t) != nil {
			continue
		}
		if t.Client != nil {
			if v, ok := clients.Load(t.Client.Id); ok {
				t.Client = v.(*Client)
			} else {
				continue
			}
		}
		tasks.Store(t.Id, t)
		if int32(t.Id) > mt {
			mt = int32(t.Id)
		}
	}
	rows.Close()

	rows, err = b.pool.Query(ctx, `SELECT data FROM hosts`)
	if err != nil {
		return
	}
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			rows.Close()
			return
		}
		h := new(Host)
		if json.Unmarshal(raw, h) != nil {
			continue
		}
		if h.Client != nil {
			if v, ok := clients.Load(h.Client.Id); ok {
				h.Client = v.(*Client)
			} else {
				continue
			}
		}
		hosts.Store(h.Id, h)
		if int32(h.Id) > mh {
			mh = int32(h.Id)
		}
	}
	rows.Close()

	err = nil
	return
}

func (b *PgBackend) UpsertClient(c *Client) error {
	if c == nil || c.NoStore {
		return nil
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	user := ""
	pass := ""
	if c.Cnf != nil {
		user = c.Cnf.U
		pass = c.Cnf.P
	}
	ctx, cancel := b.ctx()
	defer cancel()
	_, err = b.pool.Exec(ctx, `
		INSERT INTO clients (id, remark, verify_key, socks5_user, socks5_pass,
		                    no_store, no_display, status, addr, data)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET
		    remark      = EXCLUDED.remark,
		    verify_key  = EXCLUDED.verify_key,
		    socks5_user = EXCLUDED.socks5_user,
		    socks5_pass = EXCLUDED.socks5_pass,
		    no_store    = EXCLUDED.no_store,
		    no_display  = EXCLUDED.no_display,
		    status      = EXCLUDED.status,
		    addr        = EXCLUDED.addr,
		    data        = EXCLUDED.data,
		    updated_at  = NOW()`,
		c.Id, c.Remark, c.VerifyKey, user, pass,
		c.NoStore, c.NoDisplay, c.Status, c.Addr, data)
	return err
}

func (b *PgBackend) DeleteClient(id int) error {
	ctx, cancel := b.ctx()
	defer cancel()
	_, err := b.pool.Exec(ctx, `DELETE FROM clients WHERE id=$1`, id)
	return err
}

func (b *PgBackend) UpsertTask(t *Tunnel) error {
	if t == nil || t.NoStore {
		return nil
	}
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	clientId := 0
	if t.Client != nil {
		clientId = t.Client.Id
	}
	ctx, cancel := b.ctx()
	defer cancel()
	_, err = b.pool.Exec(ctx, `
		INSERT INTO tasks (id, client_id, no_store, mode, port, password, data)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
		    client_id  = EXCLUDED.client_id,
		    no_store   = EXCLUDED.no_store,
		    mode       = EXCLUDED.mode,
		    port       = EXCLUDED.port,
		    password   = EXCLUDED.password,
		    data       = EXCLUDED.data,
		    updated_at = NOW()`,
		t.Id, clientId, t.NoStore, t.Mode, t.Port, t.Password, data)
	return err
}

func (b *PgBackend) DeleteTask(id int) error {
	ctx, cancel := b.ctx()
	defer cancel()
	_, err := b.pool.Exec(ctx, `DELETE FROM tasks WHERE id=$1`, id)
	return err
}

func (b *PgBackend) UpsertHost(h *Host) error {
	if h == nil || h.NoStore {
		return nil
	}
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}
	clientId := 0
	if h.Client != nil {
		clientId = h.Client.Id
	}
	ctx, cancel := b.ctx()
	defer cancel()
	_, err = b.pool.Exec(ctx, `
		INSERT INTO hosts (id, client_id, no_store, host, location, scheme, data)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
		    client_id  = EXCLUDED.client_id,
		    no_store   = EXCLUDED.no_store,
		    host       = EXCLUDED.host,
		    location   = EXCLUDED.location,
		    scheme     = EXCLUDED.scheme,
		    data       = EXCLUDED.data,
		    updated_at = NOW()`,
		h.Id, clientId, h.NoStore, h.Host, h.Location, h.Scheme, data)
	return err
}

func (b *PgBackend) DeleteHost(id int) error {
	ctx, cancel := b.ctx()
	defer cancel()
	_, err := b.pool.Exec(ctx, `DELETE FROM hosts WHERE id=$1`, id)
	return err
}

// FlushClients walks the in-memory map and upserts every record. It
// is called periodically to checkpoint flow counters; for ad-hoc
// changes UpsertClient is invoked directly.
func (b *PgBackend) FlushClients(m *sync.Map) error {
	var firstErr error
	m.Range(func(_, value interface{}) bool {
		c := value.(*Client)
		if err := b.UpsertClient(c); err != nil && firstErr == nil {
			firstErr = err
		}
		return true
	})
	return firstErr
}

func (b *PgBackend) FlushTasks(m *sync.Map) error {
	var firstErr error
	m.Range(func(_, value interface{}) bool {
		t := value.(*Tunnel)
		if err := b.UpsertTask(t); err != nil && firstErr == nil {
			firstErr = err
		}
		return true
	})
	return firstErr
}

func (b *PgBackend) FlushHosts(m *sync.Map) error {
	var firstErr error
	m.Range(func(_, value interface{}) bool {
		h := value.(*Host)
		if err := b.UpsertHost(h); err != nil && firstErr == nil {
			firstErr = err
		}
		return true
	})
	return firstErr
}
