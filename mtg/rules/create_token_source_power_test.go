package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func goblinTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Goblin",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Goblin},
	}}
}

// TestCreateTokenSourcePowerCountReadsPostCounterPower verifies the "Whenever ~
// attacks, put a +1/+1 counter on it, then create a number of 1/1 ... tokens
// equal to ~'s power." family (Krenko, Tin Street Kingpin): the token count is
// the source creature's power read after the +1/+1 counter is applied, so a
// base-3 creature creates 4 tokens.
func TestCreateTokenSourcePowerCountReadsPostCounterPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	krenko := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	obj := &game.StackObject{Controller: game.Player1, SourceID: krenko.ObjectID}

	resolveInstruction(engine, g, obj, game.AddCounter{
		Amount:      game.Fixed(1),
		Object:      game.SourcePermanentReference(),
		CounterKind: counter.PlusOnePlusOne,
	}, &TurnLog{})
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountObjectPower,
			Multiplier: 1,
			Object:     game.SourcePermanentReference(),
		}),
		Source: game.TokenDef(goblinTokenDef()),
	}, &TurnLog{})

	if got := countTokenPermanentsNamed(g, "Goblin"); got != 4 {
		t.Fatalf("Goblin tokens = %d, want 4 (base power 3 + 1 counter)", got)
	}
}

// TestCreateTokenSourcePowerCountReadsBasePower verifies that without the
// counter the token count equals the source's base power, confirming the count
// tracks the live source power rather than a fixed value.
func TestCreateTokenSourcePowerCountReadsBasePower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	krenko := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	obj := &game.StackObject{Controller: game.Player1, SourceID: krenko.ObjectID}

	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountObjectPower,
			Multiplier: 1,
			Object:     game.SourcePermanentReference(),
		}),
		Source: game.TokenDef(goblinTokenDef()),
	}, &TurnLog{})

	if got := countTokenPermanentsNamed(g, "Goblin"); got != 2 {
		t.Fatalf("Goblin tokens = %d, want 2 (base power 2, no counter)", got)
	}
}
