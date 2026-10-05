package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

func TestCompiledCounteredSpellUsesActualSecondOccurrence(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Second Counter", Layout: "normal", TypeLine: "Instant",
		OracleText: "Counter target spell. Counter target spell. You gain 1 life. If that spell's mana value was 3 or less, you gain 2 life."})
	for _, value := range []int{2, 3, 4} {
		t.Run(strconv.Itoa(value), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			decoy := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(2)})
			target := spellWithManaValue(g, game.Player3, cost.Mana{cost.O(value)})
			addInstructionSpellToStackForController(g, game.Player1, def.SpellAbility.Val.Modes[0].Sequence,
				[]game.Target{game.StackObjectTarget(decoy.ID), game.StackObjectTarget(target.ID)})
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			want := 41
			if value <= 3 {
				want += 2
			}
			if g.Players[game.Player1].Life != want {
				t.Fatalf("life=%d, want %d; slot zero must not satisfy slot one", g.Players[game.Player1].Life, want)
			}
			if g.Players[game.Player2].Life != 40 || g.Players[game.Player3].Life != 40 {
				t.Fatal("payoff leaked to target controller")
			}
		})
	}
}
