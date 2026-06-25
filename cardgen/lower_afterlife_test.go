package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerAfterlifeCreatesSpiritTokensOnDeath verifies that the parser-owned
// "Afterlife N" expansion lowers end-to-end into a dies-triggered ability that
// creates N 1/1 white and black Spirit creature tokens with flying (CR 702.135).
func TestLowerAfterlifeCreatesSpiritTokensOnDeath(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Afterlife Creature",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "Afterlife 2 (When this creature dies, create two 1/1 white and black Spirit creature tokens with flying.)",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventPermanentDied {
		t.Fatalf("trigger event = %v, want EventPermanentDied", trigger.Trigger.Pattern.Event)
	}
	create, ok := trigger.Content.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", trigger.Content.Modes[0].Sequence[0].Primitive)
	}
	if create.Amount.Value() != 2 {
		t.Fatalf("token amount = %d, want 2", create.Amount.Value())
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if len(def.Colors) != 2 || def.Colors[0] != color.White || def.Colors[1] != color.Black {
		t.Fatalf("token colors = %v, want [White Black]", def.Colors)
	}
	if len(def.Subtypes) != 1 || def.Subtypes[0] != types.Spirit {
		t.Fatalf("token subtypes = %v, want [Spirit]", def.Subtypes)
	}
	if !def.Power.Exists || def.Power.Val.Value != 1 || !def.Toughness.Exists || def.Toughness.Val.Value != 1 {
		t.Fatalf("token P/T = %v/%v, want 1/1", def.Power, def.Toughness)
	}
	if len(def.StaticAbilities) != 1 || !reflect.DeepEqual(def.StaticAbilities[0], game.FlyingStaticBody) {
		t.Fatalf("token static abilities = %v, want [flying]", def.StaticAbilities)
	}
}

// TestLowerAfterlifeOneSingularToken verifies the singular "Afterlife 1" form
// creates a single Spirit token.
func TestLowerAfterlifeOneSingularToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Afterlife One",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Afterlife 1",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	create, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
	if create.Amount.Value() != 1 {
		t.Fatalf("token amount = %d, want 1", create.Amount.Value())
	}
}
