package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerKrenkoCreateTokensEqualToPower verifies the "Whenever ~ attacks, put
// a +1/+1 counter on it, then create a number of 1/1 red Goblin creature tokens
// equal to ~'s power." family (Krenko, Tin Street Kingpin) lowers without
// diagnostics into an ordered sequence whose counter placement targets the
// triggering permanent and whose token creation reads the source permanent's
// power.
func TestLowerKrenkoCreateTokensEqualToPower(t *testing.T) {
	t.Parallel()
	oracle := "Whenever Krenko, Tin Street Kingpin attacks, put a +1/+1 counter on it, then create a number of 1/1 red Goblin creature tokens equal to Krenko's power."
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Krenko, Tin Street Kingpin",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Goblin",
		OracleText: oracle,
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	var sequence []game.Instruction
	for fi := range faces {
		for ti := range faces[fi].TriggeredAbilities {
			modes := faces[fi].TriggeredAbilities[ti].Content.Modes
			if len(modes) == 1 {
				sequence = modes[0].Sequence
			}
		}
	}
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2 (counter + token)", len(sequence))
	}
	if _, ok := sequence[0].Primitive.(game.AddCounter); !ok {
		t.Fatalf("instruction[0] = %#v, want AddCounter", sequence[0].Primitive)
	}
	token, ok := sequence[1].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("instruction[1] = %#v, want CreateToken", sequence[1].Primitive)
	}
	dynamic := token.Amount.DynamicAmount()
	if !dynamic.Exists {
		t.Fatalf("token amount = %#v, want dynamic", token.Amount)
	}
	if dynamic.Val.Kind != game.DynamicAmountObjectPower {
		t.Fatalf("token amount kind = %v, want DynamicAmountObjectPower", dynamic.Val.Kind)
	}
}
