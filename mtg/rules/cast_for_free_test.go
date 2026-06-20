package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func castForFreeInstruction() *game.Instruction {
	return &game.Instruction{
		Primitive: game.CastForFree{
			Player: game.ControllerReference(),
			Selection: game.Selection{
				ExcludedTypes: []types.Card{types.Land},
				ManaValue:     opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 5}),
			},
			Zone: zone.Hand,
		},
	}
}

// TestCastForFreeCastsChosenSpell verifies the controller's chosen matching
// spell leaves hand for the stack while a land and an over-costed spell stay.
func TestCastForFreeCastsChosenSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	spell := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Cheap Bolt",
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
	}})
	land := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Forest",
		Types: []types.Card{types.Land},
	}})
	expensive := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Big Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(7)}),
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, castForFreeInstruction(), agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(spell) {
		t.Fatal("chosen spell still in hand")
	}
	if !g.Players[game.Player1].Hand.Contains(land) {
		t.Fatal("land was removed from hand")
	}
	if !g.Players[game.Player1].Hand.Contains(expensive) {
		t.Fatal("over-costed spell was removed from hand")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 (the free-cast spell)", g.Stack.Size())
	}
}

// TestCastForFreeWithNoEligibleSpellDoesNothing verifies that with only an
// ineligible (over-costed) spell in hand, nothing is cast.
func TestCastForFreeWithNoEligibleSpellDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	expensive := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Big Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(7)}),
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, castForFreeInstruction(), agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(expensive) {
		t.Fatal("over-costed spell was removed from hand")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
}
