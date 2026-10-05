// Package hotplug 管理输入设备的上下线与绑定迁移。
// 缺陷：重复上线不去重、迁移目标按插入顺序取而不是设备 id 最小、没有同类型在线设备时绑定被丢掉、
// 每次迁移全扫全部绑定。
package hotplug

// Binding 是一条按键绑定。
type Binding struct {
	Action string
	Device string
	Key    string
	Orphan bool
}

// Manager 是设备与绑定台账。
type Manager struct {
	devices  []string
	kinds    map[string]string
	online   map[string]bool
	bindings []Binding
	scanned  int
}

// New 建台账。
func New() *Manager {
	return &Manager{kinds: map[string]string{}, online: map[string]bool{}}
}

// AddDevice 让设备上线。缺陷：同一个 id 重复上线也照收。
func (m *Manager) AddDevice(id, kind string) {
	m.devices = append(m.devices, id)
	m.kinds[id] = kind
	m.online[id] = true
}

// AddBinding 登记一条绑定。
func (m *Manager) AddBinding(action, device, key string) {
	m.bindings = append(m.bindings, Binding{Action: action, Device: device, Key: key})
}

// RemoveDevice 让设备下线并把它的绑定迁移出去，返回迁移条数。
// 缺陷：目标按插入顺序取、没有同类型在线设备时直接丢掉、迁移时全扫绑定。
func (m *Manager) RemoveDevice(id string) int {
	m.online[id] = false
	m.scanned += len(m.bindings)
	moved := 0
	for index, item := range m.bindings {
		if item.Device != id {
			continue
		}
		target := ""
		for _, other := range m.devices {
			if other != id && m.online[other] && m.kinds[other] == m.kinds[id] {
				target = other
				break
			}
		}
		if target == "" {
			continue
		}
		m.bindings[index].Device = target
		moved++
	}
	return moved
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
