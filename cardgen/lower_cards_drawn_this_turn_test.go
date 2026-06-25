package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerThunderingDjinn proves "Whenever this creature attacks, it deals
// damage to any target equal to the number of cards you've drawn this turn."
// lowers to an attack-triggered ability whose damage amount is the controller's
// cards drawn this turn.
func TestLowerThunderingDjinn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Thundering Djinn",
		Layout:     "normal",
		TypeLine:   "Creature — Djinn",
		ManaCost:   "{3}{U}{R}",
		Power:      new("3"),
		Toughness:  new("4"),
		OracleText: "Flying\nWhenever this creature attacks, it deals damage to any target equal to the number of cards you've drawn this turn.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	prim := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive
	damage, ok := prim.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", prim)
	}
	dyn := damage.Amount.DynamicAmount()
	if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountCardsDrawnThisTurn {
		t.Fatalf("amount = %#v, want DynamicAmountCardsDrawnThisTurn", damage.Amount)
	}
}

// TestLowerDuelistOfTheMindPower proves the characteristic-defining power line
// "Duelist of the Mind's power is equal to the number of cards you've drawn this
// turn." lowers to a DynamicPower of DynamicValueControllerCardsDrawnThisTurn.
func TestLowerDuelistOfTheMindPower(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Duelist of the Mind",
		Layout:     "normal",
		TypeLine:   "Creature — Human Advisor",
		ManaCost:   "{1}{U}",
		Power:      new("*"),
		Toughness:  new("3"),
		OracleText: "Flying, vigilance\nDuelist of the Mind's power is equal to the number of cards you've drawn this turn.",
	})
	if !face.DynamicPower.Exists {
		t.Fatalf("dynamic power = %#v, want present", face.DynamicPower)
	}
	if face.DynamicPower.Val.Kind != game.DynamicValueControllerCardsDrawnThisTurn {
		t.Fatalf("dynamic power kind = %v, want DynamicValueControllerCardsDrawnThisTurn", face.DynamicPower.Val.Kind)
	}
}
