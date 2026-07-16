package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerWhitePlumeAdventurer(t *testing.T) {
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "White Plume Adventurer",
		Layout:     "normal",
		ManaCost:   "{2}{W}",
		TypeLine:   "Creature — Orc Cleric",
		OracleText: "When this creature enters, you take the initiative.\nAt the beginning of each opponent's upkeep, untap a creature you control. If you've completed a dungeon, untap all creatures you control instead.",
		Power:      new("3"),
		Toughness:  new("3"),
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %#v", face.TriggeredAbilities)
	}
	sequence := face.TriggeredAbilities[1].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("untap sequence = %#v", sequence)
	}
	one, ok := sequence[0].Primitive.(game.Untap)
	if !ok || !one.ChooseOne || one.ChooseUpTo || one.Amount.Value() != 0 {
		t.Fatalf("base untap = %#v", sequence[0].Primitive)
	}
	all, ok := sequence[1].Primitive.(game.Untap)
	if !ok || all.Group.Domain() == 0 {
		t.Fatalf("replacement untap = %#v", sequence[1].Primitive)
	}
}
