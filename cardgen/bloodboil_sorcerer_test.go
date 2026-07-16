package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

func TestLowerBloodboilSorcerer(t *testing.T) {
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Bloodboil Sorcerer",
		Layout:     "normal",
		ManaCost:   "{3}{R}",
		TypeLine:   "Creature — Human Shaman Sorcerer",
		OracleText: "When this creature enters, you take the initiative.\nCrown of Madness — {1}{R}, Sacrifice an artifact or creature: Goad target creature.",
		Power:      new("3"),
		Toughness:  new("3"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v", face.TriggeredAbilities)
	}
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v", face.ActivatedAbilities)
	}
	ability := face.ActivatedAbilities[0]
	if len(ability.AdditionalCosts) != 1 ||
		ability.AdditionalCosts[0].Kind != cost.AdditionalSacrifice {
		t.Fatalf("additional costs = %#v", ability.AdditionalCosts)
	}
	if _, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Goad); !ok {
		t.Fatalf("activated content = %#v", ability.Content)
	}
}
