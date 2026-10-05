// Package hotplug 管理输入设备的上下线与绑定迁移。
package hotplug

import (
	"errors"
	"sort"
)

// ErrUnknownDevice 表示给没上线过的设备下线。
var ErrUnknownDevice = errors.New("设备从未上线")

// ErrDeviceOffline 表示设备已经下线。
var ErrDeviceOffline = errors.New("设备不在线")

// Binding 是一条按键绑定。
type Binding struct {
	Action string
	Device string
	Key    string
	Orphan bool
}

// Manager 是设备与绑定台账。
type Manager struct {
	devices  []string            // 台账：同一 id 只保留一条，按首次上线顺序
	kinds    map[string]string   // 设备 id -> 类型
	online   map[string]bool     // 设备 id -> 是否在线
	byKind   map[string][]string // 类型 -> 在线设备 id，升序
	bindings []Binding
	byDevice map[string][]int // 设备 id -> 绑定下标，迁移只扫自己那一份
	scanned  int
}

// New 建台账。
func New() *Manager {
	return &Manager{
		kinds:    map[string]string{},
		online:   map[string]bool{},
		byKind:   map[string][]string{},
		byDevice: map[string][]int{},
	}
}

// AddDevice 让设备上线。设备已在线时返回 false，台账不重复登记。
func (m *Manager) AddDevice(id, kind string) bool {
	if m.online[id] {
		return false
	}
	if _, known := m.kinds[id]; !known {
		m.devices = append(m.devices, id)
		m.kinds[id] = kind
	}
	m.online[id] = true
	ids := m.byKind[m.kinds[id]]
	pos := sort.SearchStrings(ids, id)
	m.byKind[m.kinds[id]] = append(ids, "")
	copy(m.byKind[m.kinds[id]][pos+1:], ids[pos:])
	m.byKind[m.kinds[id]][pos] = id
	return true
}

// AddBinding 登记一条绑定。
func (m *Manager) AddBinding(action, device, key string) {
	m.byDevice[device] = append(m.byDevice[device], len(m.bindings))
	m.bindings = append(m.bindings, Binding{Action: action, Device: device, Key: key})
}

// RemoveDevice 让设备下线并迁移绑定，返回迁移条数。
// 设备从未上线或已下线时返回 0；需要错误信息时用 Offline。
func (m *Manager) RemoveDevice(id string) int {
	moved, _ := m.Offline(id)
	return moved
}

// Offline 让设备下线并把它的绑定迁移到同类型 id 最小的在线设备，
// 返回迁移条数；没有同类型在线设备时绑定标为孤儿。设备从未上线或已下线时返回错误。
func (m *Manager) Offline(id string) (int, error) {
	kind, known := m.kinds[id]
	if !known {
		return 0, ErrUnknownDevice
	}
	if !m.online[id] {
		return 0, ErrDeviceOffline
	}
	m.online[id] = false
	ids := m.byKind[kind]
	pos := sort.SearchStrings(ids, id)
	if pos < len(ids) && ids[pos] == id {
		m.byKind[kind] = append(ids[:pos], ids[pos+1:]...)
	}
	target := ""
	if rest := m.byKind[kind]; len(rest) > 0 {
		target = rest[0]
	}
	indexes := m.byDevice[id]
	delete(m.byDevice, id)
	m.scanned += len(indexes)
	moved := 0
	for _, index := range indexes {
		if target == "" {
			m.bindings[index].Device = ""
			m.bindings[index].Orphan = true
			continue
		}
		m.bindings[index].Device = target
		m.byDevice[target] = append(m.byDevice[target], index)
		moved++
	}
	return moved, nil
}

// Bindings 返回绑定快照。
func (m *Manager) Bindings() []Binding {
	return append([]Binding{}, m.bindings...)
}

// Devices 返回设备台账快照，同一 id 只出现一次。
func (m *Manager) Devices() []string {
	return append([]string{}, m.devices...)
}

// Scanned 是累计扫描的绑定数（规模观测）。
func (m *Manager) Scanned() int { return m.scanned }
