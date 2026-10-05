// Package hotplug 管理输入设备的上下线与绑定迁移。
package hotplug

import (
	"container/heap"
	"errors"
	"fmt"
)

// ErrDeviceNotOnline 表示给没在线的设备下线。
var ErrDeviceNotOnline = errors.New("hotplug: 设备不在线")

// Binding 是一条按键绑定。
type Binding struct {
	Action string
	Device string
	Key    string
	Orphan bool
}

// idHeap 是按 id 升序的最小堆，离线条目惰性删除。
type idHeap []string

func (h idHeap) Len() int           { return len(h) }
func (h idHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h idHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *idHeap) Push(x any) {
	*h = append(*h, x.(string))
}

func (h *idHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = ""
	*h = old[:n-1]
	return item
}

// Manager 是设备与绑定台账。
type Manager struct {
	devices  []string
	kinds    map[string]string
	online   map[string]bool
	byKind   map[string]*idHeap
	bindings []Binding
	byDevice map[string][]int
	scanned  int
	err      error
}

// New 建台账。
func New() *Manager {
	return &Manager{
		kinds:    map[string]string{},
		online:   map[string]bool{},
		byKind:   map[string]*idHeap{},
		byDevice: map[string][]int{},
	}
}

// AddDevice 让设备上线。同一个 id 只保留一条，重复上线返回 false。
func (m *Manager) AddDevice(id, kind string) bool {
	if m.online[id] {
		return false
	}
	if _, seen := m.kinds[id]; !seen {
		m.devices = append(m.devices, id)
		m.kinds[id] = kind
	}
	m.online[id] = true
	onlineIDs := m.byKind[m.kinds[id]]
	if onlineIDs == nil {
		onlineIDs = &idHeap{}
		m.byKind[m.kinds[id]] = onlineIDs
	}
	heap.Push(onlineIDs, id)
	return true
}

// AddBinding 登记一条绑定。
func (m *Manager) AddBinding(action, device, key string) {
	m.byDevice[device] = append(m.byDevice[device], len(m.bindings))
	m.bindings = append(m.bindings, Binding{Action: action, Device: device, Key: key})
}

// RemoveDevice 让设备下线并把它的绑定迁移到同类型 id 最小的在线设备，
// 返回迁移条数。同类型没有在线设备时绑定保留并把 Orphan 置真。
// 给没在线过的设备下线返回 0，错误信息可通过 Err 获取。
func (m *Manager) RemoveDevice(id string) int {
	if !m.online[id] {
		m.err = fmt.Errorf("%w: %s", ErrDeviceNotOnline, id)
		return 0
	}
	m.err = nil
	m.online[id] = false
	target := m.minOnline(m.kinds[id])
	moved := 0
	for _, index := range m.byDevice[id] {
		m.scanned++
		if target == "" {
			m.bindings[index].Device = ""
			m.bindings[index].Orphan = true
			continue
		}
		m.bindings[index].Device = target
		m.byDevice[target] = append(m.byDevice[target], index)
		moved++
	}
	delete(m.byDevice, id)
	return moved
}

// Err 返回最近一次下线操作的错误，没有错误时为 nil。
func (m *Manager) Err() error { return m.err }

// LastError 返回最近一次下线操作的错误，是 Err 的别名。
func (m *Manager) LastError() error { return m.err }

// minOnline 返回该类型 id 最小的在线设备，没有在线设备时返回空串。
func (m *Manager) minOnline(kind string) string {
	onlineIDs := m.byKind[kind]
	for onlineIDs != nil && onlineIDs.Len() > 0 {
		top := (*onlineIDs)[0]
		if m.online[top] {
			return top
		}
		heap.Pop(onlineIDs)
	}
	return ""
}

// Bindings 返回绑定快照。
func (m *Manager) Bindings() []Binding {
	return append([]Binding{}, m.bindings...)
}

// Devices 返回设备快照。
func (m *Manager) Devices() []string {
	return append([]string{}, m.devices...)
}

// Scanned 是累计扫描的绑定数（规模观测）。
func (m *Manager) Scanned() int { return m.scanned }
