package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerWhipOfErebosLeaveBattlefieldExileReplacement verifies that Whip of
// Erebos's "If it would leave the battlefield, exile it instead of putting it
// anywhere else." sub-effect lowers, after the reanimate/haste/delayed-exile
// chain, to a CreateReplacement bound to the reanimated permanent's linked
// object that redirects any battlefield departure to exile.
func TestLowerWhipOfErebosLeaveBattlefieldExileReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Whip of Erebos",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{2}{B}{B}",
		OracleText: "Creatures you control have lifelink.\n{2}{B}{B}, {T}: Return target creature card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of the next end step. If it would leave the battlefield, exile it instead of putting it anywhere else. Activate only as a sorcery.",
	})
	var sequence []game.Instruction
	for i := range face.ActivatedAbilities {
		for _, mode := range face.ActivatedAbilities[i].Content.Modes {
			if len(mode.Sequence) == 4 {
				sequence = mode.Sequence
			}
		}
	}
	if sequence == nil {
		t.Fatalf("no 4-instruction activated ability sequence found in %#v", face.ActivatedAbilities)
	}
	create, ok := sequence[3].Primitive.(game.CreateReplacement)
	if !ok {
		t.Fatalf("sequence[3] = %#v, want CreateReplacement", sequence[3].Primitive)
	}
	if create.Object.Kind() != game.ObjectReferenceLinkedObject || create.Object.LinkID() == "" {
		t.Fatalf("CreateReplacement.Object = %#v, want a linked object", create.Object)
	}
	if create.Replacement == nil {
		t.Fatal("CreateReplacement.Replacement is nil")
	}
	replacement := create.Replacement
	if replacement.MatchEvent != game.EventZoneChanged ||
		!replacement.MatchFromZone ||
		replacement.FromZone != zone.Battlefield ||
		replacement.ReplaceToZone != zone.Exile {
		t.Fatalf("replacement = %#v, want battlefield-departure redirect to exile", replacement)
	}
	if replacement.MatchToZone {
		t.Fatalf("replacement = %#v, want any destination matched (MatchToZone false)", replacement)
	}
}
