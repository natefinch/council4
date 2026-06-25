package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// distributiveDestroySagaDef is a minimal stand-in for The Curse of Fenric: a
// legendary Saga enchantment whose chapter drives the distributive destroy and
// the linked per-controller token payoff.
func distributiveDestroySagaDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "The Curse of Fenric",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Enchantment},
		Subtypes:   []types.Sub{types.Sub("Saga")},
	}}
}

// mutantTokenDef is the 3/3 green Mutant token the Fenric payoff mints.
func mutantTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Mutant",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sub("Mutant")},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}}
}

func mutantTokenCount(g *game.Game, controller game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.Token && permanent.Controller == controller &&
			permanent.TokenDef != nil && permanent.TokenDef.Name == "Mutant" {
			count++
		}
	}
	return count
}

// TestDestroyForEachPlayerDestroysOnePerPlayerUnderLink verifies the distributive
// Saga destroy: each player's one matching creature is destroyed under the
// destroyed-for-each-player link, with no prompt when a player controls a single
// eligible creature.
func TestDestroyForEachPlayerDestroysOnePerPlayerUnderLink(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := addCombatCreaturePermanent(g, game.Player1)
	theirs := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, distributiveDestroySagaDef())
	obj := linkedSourceObject(source)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.DestroyForEachPlayer{
		Chooser:   game.ControllerReference(),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		LinkedKey: game.LinkedKey("destroyed-for-each-player"),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if permanentByCardID(g, mine.CardInstanceID) != nil || permanentByCardID(g, theirs.CardInstanceID) != nil {
		t.Fatal("a chosen creature remained on the battlefield after distributive destroy")
	}
	if !g.Players[game.Player1].Graveyard.Contains(mine.CardInstanceID) {
		t.Fatal("Player1's creature did not reach its owner's graveyard")
	}
	if !g.Players[game.Player2].Graveyard.Contains(theirs.CardInstanceID) {
		t.Fatal("Player2's creature did not reach its owner's graveyard")
	}
	key := linkedObjectSourceKey(g, obj, "destroyed-for-each-player")
	if got := len(linkedObjects(g, key)); got != 2 {
		t.Fatalf("linked destroyed objects = %d, want 2 (one per player)", got)
	}
}

// TestCreateTokenForEachDestroyedMintsPerController verifies the chapter payoff:
// each destroyed creature's last-known controller creates one token, and the
// link is cleared so the payoff fires exactly once.
func TestCreateTokenForEachDestroyedMintsPerController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatCreaturePermanent(g, game.Player1)
	addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, distributiveDestroySagaDef())
	obj := linkedSourceObject(source)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.DestroyForEachPlayer{
		Chooser:   game.ControllerReference(),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		LinkedKey: game.LinkedKey("destroyed-for-each-player"),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.CreateTokenForEachDestroyed{
		Source:    game.TokenDef(mutantTokenDef()),
		LinkedKey: game.LinkedKey("destroyed-for-each-player"),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := mutantTokenCount(g, game.Player1); got != 1 {
		t.Fatalf("Player1 Mutant tokens = %d, want 1 (its destroyed creature's controller)", got)
	}
	if got := mutantTokenCount(g, game.Player2); got != 1 {
		t.Fatalf("Player2 Mutant tokens = %d, want 1 (its destroyed creature's controller)", got)
	}
	key := linkedObjectSourceKey(g, obj, "destroyed-for-each-player")
	if got := len(linkedObjects(g, key)); got != 0 {
		t.Fatalf("linked destroyed objects after payoff = %d, want 0 (link cleared)", got)
	}
}
