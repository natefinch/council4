package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func TestLowerChishiroModifiedCreatureCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Chishiro, the Shattered Blade",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Snake Samurai",
		ManaCost:   "{2}{R}{G}",
		Power:      new("4"),
		Toughness:  new("4"),
		OracleText: "Whenever an Aura or Equipment you control enters, create a 2/2 red Spirit creature token with menace.\nAt the beginning of your end step, put a +1/+1 counter on each modified creature you control. (Equipment, Auras you control, and counters are modifications.)",
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want 2", len(face.TriggeredAbilities))
	}
	add, ok := face.TriggeredAbilities[1].Content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
	if !ok ||
		add.CounterKind != counter.PlusOnePlusOne ||
		!add.Group.Valid() ||
		!add.Group.Selection().MatchModified ||
		add.Group.Selection().Controller != game.ControllerYou {
		t.Fatalf("counter primitive = %#v", add)
	}
}
