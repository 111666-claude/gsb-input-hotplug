package hotplug

import "testing"

func TestAddBindingThenSnapshot(t *testing.T) {
	manager := New()
	manager.AddDevice("pad-a", "pad")
	manager.AddBinding("jump", "pad-a", "space")
	if len(manager.Bindings()) != 1 {
		t.Fatal("应该有一条绑定")
	}
}

func TestRemoveUnknownDeviceMovesNothing(t *testing.T) {
	manager := New()
	manager.AddBinding("jump", "pad-a", "space")
	if got := manager.RemoveDevice("pad-z"); got != 0 {
		t.Fatalf("给没上线过的设备下线不该迁移任何绑定，得到 %d", got)
	}
}

func TestScannedStartsAtZero(t *testing.T) {
	if New().Scanned() != 0 {
		t.Fatal("Scanned 初始应该是 0")
	}
}

func TestDuplicateAddDeviceKeepsOne(t *testing.T) {
	manager := New()
	if !manager.AddDevice("pad-a", "pad") {
		t.Fatal("首次上线应该返回 true")
	}
	if manager.AddDevice("pad-a", "pad") {
		t.Fatal("重复上线应该返回 false")
	}
	if len(manager.Devices()) != 1 {
		t.Fatalf("同一个 id 只保留一条，得到 %d", len(manager.Devices()))
	}
}

func TestRemoveUnknownDeviceReportsError(t *testing.T) {
	manager := New()
	if got := manager.RemoveDevice("pad-z"); got != 0 {
		t.Fatalf("没上线过的设备不该迁移绑定，得到 %d", got)
	}
	if manager.Err() == nil {
		t.Fatal("给没上线过的设备下线应该记录错误")
	}
	manager.AddDevice("pad-a", "pad")
	manager.RemoveDevice("pad-a")
	if manager.Err() != nil {
		t.Fatal("正常下线后错误应该被清除")
	}
}

func TestMigrationTargetsSmallestOnlineID(t *testing.T) {
	manager := New()
	manager.AddDevice("pad-b", "pad")
	manager.AddDevice("pad-a", "pad")
	manager.AddDevice("pad-c", "pad")
	manager.AddBinding("jump", "pad-c", "space")
	if moved := manager.RemoveDevice("pad-c"); moved != 1 {
		t.Fatalf("应该迁移 1 条，得到 %d", moved)
	}
	if got := manager.Bindings()[0].Device; got != "pad-a" {
		t.Fatalf("应该迁到 id 最小的 pad-a，得到 %s", got)
	}
}

func TestOrphanBindingKeptWhenNoOnlineKind(t *testing.T) {
	manager := New()
	manager.AddDevice("wheel-a", "wheel")
	manager.AddBinding("attack", "wheel-a", "btn0")
	manager.RemoveDevice("wheel-a")
	bindings := manager.Bindings()
	if len(bindings) != 1 {
		t.Fatalf("孤儿绑定不许丢弃，剩 %d 条", len(bindings))
	}
	if !bindings[0].Orphan || bindings[0].Device != "" {
		t.Fatalf("孤儿绑定应该 Orphan=true 且设备为空，得到 %+v", bindings[0])
	}
}

func TestScannedDoesNotMultiplyWithMigrations(t *testing.T) {
	manager := New()
	manager.AddDevice("pad-a", "pad")
	for index := 0; index < 3000; index++ {
		manager.AddBinding("a", "pad-a", "k")
	}
	for index := 0; index < 3000; index++ {
		manager.AddDevice("pad-c", "pad")
		manager.RemoveDevice("pad-c")
	}
	if manager.Scanned() > 6000 {
		t.Fatalf("Scanned 不该随迁移次数膨胀，得到 %d", manager.Scanned())
	}
}
