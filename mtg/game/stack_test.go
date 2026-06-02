package game

import "testing"

func TestStackObjectCarriesPlayerTarget(t *testing.T) {
	target := PlayerTarget(Player3)
	obj := &StackObject{
		Targets: []Target{target},
	}

	if len(obj.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(obj.Targets))
	}
	if obj.Targets[0] != target {
		t.Fatalf("target = %+v, want %+v", obj.Targets[0], target)
	}
}
