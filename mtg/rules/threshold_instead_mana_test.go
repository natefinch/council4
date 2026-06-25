package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestThresholdInsteadManaResolution verifies the Cabal Ritual conditional-mana
// shape at runtime: three black productions gated on NOT(seven+ graveyard
// cards) and five gated on the threshold, so the caster floats {B}{B}{B} below
// threshold and {B}{B}{B}{B}{B} once seven cards sit in the graveyard.
func TestThresholdInsteadManaResolution(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		graveyardCards int
		wantBlack      int
	}{
		{name: "below threshold", graveyardCards: 6, wantBlack: 3},
		{name: "at threshold", graveyardCards: 7, wantBlack: 5},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			for range test.graveyardCards {
				addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name:  "Filler",
					Types: []types.Card{types.Instant},
				}})
			}
			threshold := game.Condition{Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 7}}}
			belowThreshold := threshold
			belowThreshold.Negate = true
			var seq []game.Instruction
			for range 3 {
				seq = append(seq, game.Instruction{
					Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.B},
					Condition: opt.Val(game.EffectCondition{Condition: opt.Val(belowThreshold)}),
				})
			}
			for range 5 {
				seq = append(seq, game.Instruction{
					Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.B},
					Condition: opt.Val(game.EffectCondition{Condition: opt.Val(threshold)}),
				})
			}
			addInstructionSpellToStack(g, seq)

			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

			if got := g.Players[game.Player1].ManaPool.Amount(mana.B); got != test.wantBlack {
				t.Fatalf("floated black = %d, want %d", got, test.wantBlack)
			}
		})
	}
}
