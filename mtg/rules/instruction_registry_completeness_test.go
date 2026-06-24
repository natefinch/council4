package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestPrimitiveRegistryCompleteness asserts that every non-unknown
// PrimitiveKind has a registered handler in the global registry the engine
// actually uses. If a future PrimitiveKind is added without a matching
// registerPrimitiveHandler call, this test fails immediately and names the
// missing kind, rather than only failing when a card reaches that path.
func TestPrimitiveRegistryCompleteness(t *testing.T) {
	reg := globalPrimitiveRegistry()
	for kind := game.PrimitiveUnknown + 1; int(kind) < game.PrimitiveKindCount; kind++ {
		if reg.handlers[kind] == nil {
			t.Errorf("primitive kind %d has no registered handler", int(kind))
		}
	}
}
