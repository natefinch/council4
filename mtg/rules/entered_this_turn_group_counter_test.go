package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestGroupCounterPlacementHonorsEnteredThisTurnFilter resolves Oran-Rief's
// "{T}: Put a +1/+1 counter on each green creature that entered this turn."
// activated ability as a group AddCounter and asserts the counter lands only on
// the green creature that entered this turn, skipping a green creature already in
// play before this turn and a non-green creature that entered this turn.
func TestGroupCounterPlacementHonorsEnteredThisTurnFilter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	greenNew := addCombatPermanent(g, game.Player1, greenCreatureDef("Green Newcomer"))
	greenOld := addCombatPermanent(g, game.Player1, greenCreatureDef("Green Veteran"))
	redNew := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Red Newcomer",
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Red},
	}})

	emitEvent(g, game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: greenNew.ObjectID})
	emitEvent(g, game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: redNew.ObjectID})

	addEffectSpellToStack(g, game.Player1, game.AddCounter{
		Amount:      game.Fixed(1),
		CounterKind: counter.PlusOnePlusOne,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes:   []types.Card{types.Creature},
			ColorsAny:       []color.Color{color.Green},
			EnteredThisTurn: true,
		}),
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := greenNew.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("green creature that entered this turn got %d counters, want 1", got)
	}
	if got := greenOld.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("green creature already in play got %d counters, want 0", got)
	}
	if got := redNew.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("non-green creature that entered this turn got %d counters, want 0", got)
	}
}

func greenCreatureDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:   name,
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Green},
	}}
}
