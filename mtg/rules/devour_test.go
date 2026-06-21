package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func devourCreatureDef(multiplier int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Devourer",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
		ReplacementAbilities: []game.ReplacementAbility{
			game.DevourReplacement("As this creature enters, you may sacrifice any number of creatures, then it enters with 2 +1/+1 counters on it for each creature sacrificed.", multiplier),
		},
	}}
}

func fodderCreatureDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

// TestDevourSacrificesCreaturesForCounters verifies that a Devour 2 creature
// entering while its controller chooses to sacrifice two creatures enters with
// 2 counters per sacrifice (four), and that the chosen creatures are sacrificed.
func TestDevourSacrificesCreaturesForCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	fodder1 := addCombatPermanent(g, game.Player1, fodderCreatureDef("Fodder One"))
	fodder2 := addCombatPermanent(g, game.Player1, fodderCreatureDef("Fodder Two"))
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0, 1}}}}
	permanent := enterRiotCreature(t, g, engine, devourCreatureDef(2), agents)
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("devour counters = %d, want 4", got)
	}
	if _, ok := permanentByObjectID(g, fodder1.ObjectID); ok {
		t.Fatal("first sacrificed creature is still on the battlefield")
	}
	if _, ok := permanentByObjectID(g, fodder2.ObjectID); ok {
		t.Fatal("second sacrificed creature is still on the battlefield")
	}
}

// TestDevourDeclineSacrifice verifies that declining to sacrifice (the default)
// leaves the entering creature with no counters and the other creatures intact.
func TestDevourDeclineSacrifice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	fodder := addCombatPermanent(g, game.Player1, fodderCreatureDef("Fodder One"))
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{}}}}
	permanent := enterRiotCreature(t, g, engine, devourCreatureDef(2), agents)
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("devour counters = %d, want 0", got)
	}
	if _, ok := permanentByObjectID(g, fodder.ObjectID); !ok {
		t.Fatal("creature was sacrificed despite declining devour")
	}
}
