package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func enterRiotCreature(t *testing.T, g *game.Game, engine *Engine, def *game.CardDef, agents [game.NumPlayers]PlayerAgent) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)
	permanent, ok := createCardPermanentWithChoices(engine, g, card, game.Player1, zone.Hand, agents, &TurnLog{})
	if !ok {
		t.Fatal("createCardPermanentWithChoices() = false, want true")
	}
	return permanent
}

func riotCreatureDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:            "Riot Bear",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{game.RiotStaticBody},
	}}
}

// TestRiotIntrinsicCounterChoice verifies that a creature with the riot keyword
// enters with a +1/+1 counter when its controller chooses the counter mode.
func TestRiotIntrinsicCounterChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
	permanent := enterRiotCreature(t, g, engine, riotCreatureDef(), agents)
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("riot counter choice: +1/+1 counters = %d, want 1", got)
	}
	if !permanent.SummoningSick {
		t.Fatal("riot counter choice must not clear summoning sickness")
	}
}

// TestRiotIntrinsicHasteChoice verifies that the haste mode clears summoning
// sickness and adds no counter.
func TestRiotIntrinsicHasteChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	permanent := enterRiotCreature(t, g, engine, riotCreatureDef(), agents)
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("riot haste choice: +1/+1 counters = %d, want 0", got)
	}
	if permanent.SummoningSick {
		t.Fatal("riot haste choice must clear summoning sickness")
	}
}

// TestRiotGroupGrantedCounterChoice verifies that riot granted to a group
// ("Nontoken creatures you control have riot.") applies the riot entry choice
// to a creature that enters while the granting permanent is in play.
func TestRiotGroupGrantedCounterChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Riot Granter",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
				),
				AddKeywords: []game.Keyword{game.Riot},
			}},
		}},
	}})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
	plain := &game.CardDef{CardFace: game.CardFace{
		Name:      "Plain Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
	permanent := enterRiotCreature(t, g, engine, plain, agents)
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("group-granted riot counter choice: +1/+1 counters = %d, want 1", got)
	}
}
