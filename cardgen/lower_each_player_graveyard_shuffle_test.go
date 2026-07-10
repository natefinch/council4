package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerEachPlayerGraveyardShuffle(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Mnemonic Nexus",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Each player shuffles their graveyard into their library.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 1 {
		t.Fatalf("mode = %#v", mode)
	}
	shuffle, ok := mode.Sequence[0].Primitive.(game.ShuffleGraveyardIntoLibrary)
	if !ok || shuffle.PlayerGroup != game.AllPlayersReference() {
		t.Fatalf("primitive = %#v, want all-player graveyard shuffle", mode.Sequence[0].Primitive)
	}
}
