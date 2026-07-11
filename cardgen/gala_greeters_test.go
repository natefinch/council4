package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func TestLowerGalaGreetersModesUniquePerTurn(t *testing.T) {
	t.Parallel()

	power, toughness := "1", "1"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Gala Greeters",
		Layout:     "normal",
		TypeLine:   "Creature — Elf Druid",
		OracleText: "Alliance — Whenever another creature you control enters, choose one that hasn't been chosen this turn —\n• Put a +1/+1 counter on this creature.\n• Create a tapped Treasure token.\n• You gain 2 life.",
		Power:      &power,
		Toughness:  &toughness,
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	content := face.TriggeredAbilities[0].Content
	if !content.ModesUniquePerTurn || content.MinModes != 1 ||
		content.MaxModes != 1 || len(content.Modes) != 3 {
		t.Fatalf("content = %#v, want three modes unique per turn", content)
	}
	add, ok := content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
	if !ok || add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("first mode = %#v, want +1/+1 counter", content.Modes[0])
	}
	if _, ok := content.Modes[1].Sequence[0].Primitive.(game.CreateToken); !ok {
		t.Fatalf("second mode = %#v, want Treasure token", content.Modes[1])
	}
	if _, ok := content.Modes[2].Sequence[0].Primitive.(game.GainLife); !ok {
		t.Fatalf("third mode = %#v, want gain life", content.Modes[2])
	}
}
