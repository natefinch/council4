package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// eachPlayerChooseDestroyDef is a minimal stand-in for Druid of Purification: a
// permanent whose enters trigger drives the each-player choose-and-destroy.
func eachPlayerChooseDestroyDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Druid of Purification",
		Types: []types.Card{types.Creature},
	}}
}

// TestEachPlayerChooseDestroyDestroysChosenNotControlledPermanents verifies the
// Druid of Purification interaction: each player, starting with the controller,
// chooses up to one permanent from the shared "you don't control" pool, and every
// chosen permanent is destroyed while the controller's own matching permanent —
// never in the pool — survives.
func TestEachPlayerChooseDestroyDestroysChosenNotControlledPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	theirArtifact := addBattlefieldPermanent(g, game.Player2, "Their Rock", []types.Card{types.Artifact})
	theirEnchant := addBattlefieldPermanent(g, game.Player2, "Their Aura", []types.Card{types.Enchantment})
	myArtifact := addBattlefieldPermanent(g, game.Player1, "My Rock", []types.Card{types.Artifact})
	source := addCombatPermanent(g, game.Player1, eachPlayerChooseDestroyDef())
	obj := linkedSourceObject(source)

	// Player1 (controller) picks the pool's first candidate (their artifact);
	// Player2 picks the pool's second candidate (their enchantment); the other
	// players decline. The pool is identical for every chooser because it is
	// evaluated relative to the controller ("you don't control").
	agents := [game.NumPlayers]PlayerAgent{
		&choiceOnlyAgent{choices: [][]int{{0}}},
		&choiceOnlyAgent{choices: [][]int{{1}}},
		&choiceOnlyAgent{},
		&choiceOnlyAgent{},
	}

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.EachPlayerChooseDestroy{
		Selection: game.Selection{
			RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment},
			Controller:       game.ControllerNotYou,
		},
		Optional: true,
	}}, agents, &TurnLog{})

	if permanentByCardID(g, theirArtifact.CardInstanceID) != nil {
		t.Fatal("opponent's chosen artifact survived the destroy")
	}
	if permanentByCardID(g, theirEnchant.CardInstanceID) != nil {
		t.Fatal("opponent's chosen enchantment survived the destroy")
	}
	if permanentByCardID(g, myArtifact.CardInstanceID) == nil {
		t.Fatal("controller's own artifact was destroyed but is not in the 'you don't control' pool")
	}
	if !g.Players[game.Player2].Graveyard.Contains(theirArtifact.CardInstanceID) ||
		!g.Players[game.Player2].Graveyard.Contains(theirEnchant.CardInstanceID) {
		t.Fatal("destroyed permanents did not reach their owner's graveyard")
	}
}

// TestEachPlayerChooseDestroyDedupesSharedChoice verifies that when more than one
// player chooses the same permanent from the shared pool it is destroyed once and
// a declining player removes nothing.
func TestEachPlayerChooseDestroyDedupesSharedChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	theirArtifact := addBattlefieldPermanent(g, game.Player2, "Their Rock", []types.Card{types.Artifact})
	source := addCombatPermanent(g, game.Player1, eachPlayerChooseDestroyDef())
	obj := linkedSourceObject(source)

	// Both the controller and Player2 pick the single pool candidate; the rest
	// decline. The permanent must be destroyed exactly once without error.
	agents := [game.NumPlayers]PlayerAgent{
		&choiceOnlyAgent{choices: [][]int{{0}}},
		&choiceOnlyAgent{choices: [][]int{{0}}},
		&choiceOnlyAgent{},
		&choiceOnlyAgent{},
	}

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.EachPlayerChooseDestroy{
		Selection: game.Selection{
			RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment},
			Controller:       game.ControllerNotYou,
		},
		Optional: true,
	}}, agents, &TurnLog{})

	if permanentByCardID(g, theirArtifact.CardInstanceID) != nil {
		t.Fatal("the shared chosen permanent survived the destroy")
	}
	if !g.Players[game.Player2].Graveyard.Contains(theirArtifact.CardInstanceID) {
		t.Fatal("the destroyed permanent did not reach its owner's graveyard")
	}
}

// TestEachPlayerChooseDestroyMandatoryRequiresChoice verifies that a
// non-optional instance (Optional:false) makes every chooser pick from a
// non-empty pool: with a scripted decline the mandatory chooser still resolves
// to its default (first) candidate, so the pool candidate is destroyed.
func TestEachPlayerChooseDestroyMandatoryRequiresChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	theirArtifact := addBattlefieldPermanent(g, game.Player2, "Their Rock", []types.Card{types.Artifact})
	source := addCombatPermanent(g, game.Player1, eachPlayerChooseDestroyDef())
	obj := linkedSourceObject(source)

	// Every player attempts to decline (empty scripted choice). A mandatory
	// choice over a non-empty pool cannot be declined, so the choice falls back
	// to its default first candidate for the controller.
	agents := [game.NumPlayers]PlayerAgent{
		&choiceOnlyAgent{choices: [][]int{nil}},
		&choiceOnlyAgent{choices: [][]int{nil}},
		&choiceOnlyAgent{choices: [][]int{nil}},
		&choiceOnlyAgent{choices: [][]int{nil}},
	}

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.EachPlayerChooseDestroy{
		Selection: game.Selection{
			RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment},
			Controller:       game.ControllerNotYou,
		},
	}}, agents, &TurnLog{})

	if permanentByCardID(g, theirArtifact.CardInstanceID) != nil {
		t.Fatal("mandatory each-player choose left the pool candidate on the battlefield")
	}
}
