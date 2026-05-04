package file

import (
	"sort"
	"sync"
)

// Pair holds the tuple sortClientByKey works on. clientFlow is never
// nil; the populator allocates a placeholder when the source client
// has no Flow yet so the reflection in Less cannot panic.
type Pair struct {
	key        string //sort key
	cId        int
	order      string
	clientFlow *Flow
}

// PairList implements sort.Interface to sort by the chosen key.
type PairList []*Pair

func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int      { return len(p) }
func (p PairList) Less(i, j int) bool {
	a := flowFieldInt(p[i].clientFlow, p[i].key)
	b := flowFieldInt(p[j].clientFlow, p[j].key)
	if p[i].order == "desc" {
		return a < b
	}
	return a > b
}

// flowFieldInt returns the int value of the named Flow field, or 0
// when the field does not exist or the flow pointer is nil. The
// allowed field set matches what GetClientList passes as `sort` and
// is intentionally narrow: anything else returns 0 instead of
// panicking on reflection.
func flowFieldInt(f *Flow, name string) int64 {
	if f == nil || name == "" {
		return 0
	}
	f.RLock()
	defer f.RUnlock()
	switch name {
	case "InletFlow":
		return f.InletFlow
	case "ExportFlow":
		return f.ExportFlow
	case "FlowLimit":
		return f.FlowLimit
	}
	return 0
}

// sortClientByKey returns a list of client ids sorted by the chosen
// flow field. The caller is expected to pass a pointer to the live
// sync.Map (so we don't copy the lock value, which is forbidden on
// sync.Map after first use).
func sortClientByKey(m *sync.Map, sortKey, order string) (res []int) {
	p := make(PairList, 0)
	m.Range(func(key, value interface{}) bool {
		c, ok := value.(*Client)
		if !ok || c == nil {
			return true
		}
		flow := c.Flow
		if flow == nil {
			flow = new(Flow)
		}
		p = append(p, &Pair{sortKey, c.Id, order, flow})
		return true
	})
	sort.Sort(p)
	for _, v := range p {
		res = append(res, v.cId)
	}
	return
}
