package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestExileLibraryUntilNonlandCastDigsAndCasts exiles top library cards until a
// nonland card, leaves the dug lands in exile, and casts the nonland for free.
func TestExileLibraryUntilNonlandCastDigsAndCasts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	keep := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bottom Card",
		Types: []types.Card{types.Sorcery},
	}})
	spell := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Cheap Bolt",
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
	}})
	land := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Forest",
		Types: []types.Card{types.Land},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	instr := &game.Instruction{Primitive: game.ExileLibraryUntilNonlandCast{Player: game.ControllerReference()}}
	engine.resolveInstructionWithChoices(g, obj, instr, agents, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(land) {
		t.Fatal("land not exiled")
	}
	if g.Players[game.Player1].Hand.Contains(spell) || g.Players[game.Player1].Library.Contains(spell) {
		t.Fatal("nonland spell was not removed from library to be cast")
	}
	if !g.Players[game.Player1].Library.Contains(keep) {
		t.Fatal("card past the first nonland was disturbed")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 free-cast spell", g.Stack.Size())
	}
}
