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
