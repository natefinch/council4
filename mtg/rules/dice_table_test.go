package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// dieRollTableInstructions builds the "Roll a d20." outcome table used by the
// gated-row tests: a RollDie that publishes its rolled value followed by three
// Draw instructions, each gated on a distinct inclusive result interval and
// drawing a distinct number of cards so the resolved hand size identifies which
// row resolved. Exactly one row's interval contains any rolled value, so at most
// one Draw applies.
func dieRollTableInstructions() []game.Instruction {
	const resultKey = game.ResultKey("die-roll-result")
	gatedDraw := func(amount, lo, hi int) game.Instruction {
		return game.Instruction{
			Primitive: game.Draw{Amount: game.Fixed(amount), Player: game.ControllerReference()},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:         resultKey,
				AmountRange: opt.Val(game.IntRange{Min: lo, Max: hi}),
			}),
		}
	}
	return []game.Instruction{
		{Primitive: game.RollDie{Sides: 20}, PublishResult: resultKey},
		gatedDraw(1, 1, 9),
		gatedDraw(2, 10, 19),
		gatedDraw(3, 20, 20),
	}
}

// TestDiceTableResolvesMatchingRow verifies the d20 outcome-table mechanic
// (Bag of Tricks et al.): a RollDie publishes its rolled value and each row's
// instruction is gated on the value falling within the row's inclusive interval,
// so exactly the matching row resolves. Three seeds force a low (1—9), middle
// (10—19), and maximum (20) roll, and the resulting hand size proves only that
// row's draw applied.
func TestDiceTableResolvesMatchingRow(t *testing.T) {
	cases := []struct {
		name      string
		seed1     uint64
		seed2     uint64
		wantRoll  int
		wantDrawn int
	}{
		{"low row 1-9 draws 1", 4, 5, 5, 1},
		{"middle row 10-19 draws 2", 1, 2, 16, 2},
		{"max row 20 draws 3", 40, 41, 20, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rand.New(rand.NewPCG(tc.seed1, tc.seed2)).IntN(20) + 1; got != tc.wantRoll {
				t.Fatalf("seed (%d,%d) rolls %d, want %d", tc.seed1, tc.seed2, got, tc.wantRoll)
			}
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(rand.New(rand.NewPCG(tc.seed1, tc.seed2)))
			card := &game.CardDef{CardFace: game.CardFace{
				Name:  "Filler",
				Types: []types.Card{types.Artifact},
			}}
			for range 10 {
				addCardToLibrary(g, game.Player1, card)
			}
			before := g.Players[game.Player1].Hand.Size()
			addInstructionSpellToStack(g, dieRollTableInstructions())
			engine.resolveTopOfStack(g, &TurnLog{})

			if drawn := g.Players[game.Player1].Hand.Size() - before; drawn != tc.wantDrawn {
				t.Fatalf("drew %d cards on a roll of %d, want %d (only the matching row resolves)", drawn, tc.wantRoll, tc.wantDrawn)
			}
		})
	}
}
