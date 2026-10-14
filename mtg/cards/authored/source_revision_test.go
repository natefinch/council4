package authored

import (
	"testing"

	"github.com/natefinch/council4/internal/sourcebinding"
)

func TestExecutableSourceBindings(t *testing.T) {
	for _, check := range []struct {
		want string
		dirs []string
	}{
		{SourceRevision, []string{"mtg/cards/authored"}},
		{RulesRevision, []string{"mtg/rules", "mtg/game", "mtg/cardprogram", "opt"}},
	} {
		got, err := sourcebinding.Digest("../../..", check.dirs...)
		if err != nil || got != check.want {
			t.Fatalf("stale executable source binding: %s != %s (%v); run go generate ./mtg/cards/authored", got, check.want, err)
		}
	}
}
