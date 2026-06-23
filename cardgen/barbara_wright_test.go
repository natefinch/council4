package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerBarbaraWrightReadAheadAnthem proves the "History Teacher — Sagas you
// control have read ahead." anthem lowers to a static ability that grants the
// Read ahead keyword to the controlled-Saga group. The leading "History Teacher"
// flavor ability word (rule 207.2c) does not block the otherwise-empty static
// shell, and the second line's "Doctor's companion" keyword lowers alongside it.
func TestLowerBarbaraWrightReadAheadAnthem(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Barbara Wright",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Advisor",
		ManaCost:   "{1}{W}",
		Power:      new("1"),
		Toughness:  new("3"),
		OracleText: "History Teacher — Sagas you control have read ahead. (As a Saga enters, choose a chapter and start with that many lore counters. Skipped chapters don't trigger.)\nDoctor's companion (You can have two commanders if the other is the Doctor.)",
	})

	grant := false
	for i := range face.StaticAbilities {
		body := face.StaticAbilities[i].Body
		for _, effect := range body.ContinuousEffects {
			if effect.Layer != game.LayerAbility ||
				!slices.Contains(effect.AddKeywords, game.ReadAhead) {
				continue
			}
			if !slices.Equal(effect.Group.Selection().SubtypesAny, []types.Sub{types.Saga}) {
				t.Fatalf("granted group = %#v, want controlled Sagas", effect.Group)
			}
			grant = true
		}
	}
	if !grant {
		t.Fatalf("no Read ahead grant to controlled Sagas in %#v", face.StaticAbilities)
	}
}

// TestLowerSagasYouControlAnthemKeyword proves the controlled-Saga anthem subject
// ("Sagas you control have <keyword>") lowers to a controlled-group ability grant
// even without the leading flavor word, mirroring the existing controlled-creature
// and controlled-artifact anthem subjects.
func TestLowerSagasYouControlAnthemKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Saga Anthem",
		Layout:     "normal",
		TypeLine:   "Creature — Advisor",
		ManaCost:   "{1}{W}",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "Sagas you control have read ahead.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want one Saga anthem", len(face.StaticAbilities))
	}
	effects := face.StaticAbilities[0].Body.ContinuousEffects
	if len(effects) != 1 {
		t.Fatalf("continuous effects = %#v, want one keyword grant", effects)
	}
	effect := effects[0]
	if effect.Layer != game.LayerAbility ||
		!slices.Equal(effect.AddKeywords, []game.Keyword{game.ReadAhead}) ||
		!slices.Equal(effect.Group.Selection().SubtypesAny, []types.Sub{types.Saga}) {
		t.Fatalf("grant = %#v, want Read ahead to controlled Sagas", effect)
	}
}
