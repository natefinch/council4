package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerLaezelVlaakithsChampion(t *testing.T) {
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Lae'zel, Vlaakith's Champion",
		Layout:     "normal",
		ManaCost:   "{2}{W}",
		TypeLine:   "Legendary Creature — Gith Warrior",
		OracleText: "If you would put one or more counters on a creature or planeswalker you control or on yourself, put that many plus one of each of those kinds of counters on that permanent or player instead.\nChoose a Background",
		Power:      new("3"),
		Toughness:  new("3"),
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("replacement abilities = %#v", face.ReplacementAbilities)
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if replacement.CounterAddend != 1 ||
		!replacement.CounterRecipientControllerPlayer ||
		!slices.Equal(replacement.CounterRecipientSelection.RequiredTypesAny, []types.Card{types.Creature, types.Planeswalker}) ||
		replacement.ControllerFilter != game.TriggerControllerYou {
		t.Fatalf("replacement = %#v", replacement)
	}
}
