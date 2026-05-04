package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"ehang.io/nps/lib/common"
)

// InitBackend chooses the persistence backend based on the supplied
// configuration values. It is intended to be called once at process
// start, before GetDb() is first invoked.
//
// driver: "json" (default) or one of "postgres", "postgresql", "pg".
// dsn:    Postgres DSN, only used when the postgres driver is selected.
// runPath: directory containing the conf/ folder used by the JSON
// backend.
//
// When postgres is selected and the conf/clients.json file still
// exists, the function imports the JSON snapshot into Postgres on
// first start and renames the JSON files to .imported so a second
// start does not re-import.
func InitBackend(driver, dsn, runPath string) error {
	driver = strings.ToLower(strings.TrimSpace(driver))
	switch driver {
	case "", "json", "file":
		SetBackend(NewJsonBackend(runPath))
		return nil
	case "postgres", "postgresql", "pg":
		pg, err := NewPgBackend(dsn)
		if err != nil {
			return err
		}
		SetBackend(pg)
		if err := importJsonIntoBackend(runPath, pg); err != nil {
			common.Log("[file] postgres import skipped: %s", err.Error())
		}
		return nil
	default:
		return errors.New("unknown db_driver: " + driver)
	}
}

// importJsonIntoBackend reads conf/{clients,tasks,hosts}.json once
// and pushes every record into the active backend, then renames the
// files to .imported.
func importJsonIntoBackend(runPath string, b Backend) error {
	cdir := filepath.Join(runPath, "conf")
	clientFile := filepath.Join(cdir, "clients.json")
	taskFile := filepath.Join(cdir, "tasks.json")
	hostFile := filepath.Join(cdir, "hosts.json")

	if !exists(clientFile) && !exists(taskFile) && !exists(hostFile) {
		return nil
	}

	tmp := NewJsonBackend(runPath)
	clients, tasks, hosts := newSyncMaps()
	if _, _, _, err := tmp.LoadAll(clients, tasks, hosts); err != nil {
		return err
	}
	clients.Range(func(_, v interface{}) bool {
		_ = b.UpsertClient(v.(*Client))
		return true
	})
	tasks.Range(func(_, v interface{}) bool {
		_ = b.UpsertTask(v.(*Tunnel))
		return true
	})
	hosts.Range(func(_, v interface{}) bool {
		_ = b.UpsertHost(v.(*Host))
		return true
	})
	for _, p := range []string{clientFile, taskFile, hostFile} {
		if exists(p) {
			_ = os.Rename(p, p+".imported")
		}
	}
	return nil
}

func exists(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil
}
