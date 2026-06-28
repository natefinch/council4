package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// sacrificedThisWayKey mirrors the published count key used by count-scaled
// sacrifice sequences ("sacrifice ... then <reward> that many/much").
const sacrificedThisWayKey = game.ResultKey("sacrificed-this-way")

// TestSacrificeAllPublishesCountForScaledCreateToken models Hellion Eruption:
// sacrificing every creature publishes the count so the follow-up creates
// exactly that many tokens.
func TestSacrificeAllPublishesCountForScaledCreateToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for _, name := range []string{"Alpha", "Beta", "Gamma"} {
		addBattlefieldPermanent(g, game.Player1, name, []types.Card{types.Creature})
	}
	token := &game.CardDef{CardFace: game.CardFace{
		Name:      "Hellion",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	}}
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{
			Primitive: game.SacrificePermanents{
				All:       true,
				Player:    game.ControllerReference(),
				Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			PublishResult: sacrificedThisWayKey,
		},
		{Primitive: game.CreateToken{
			Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountPreviousEffectResult, ResultKey: sacrificedThisWayKey}),
			Source: game.TokenDef(token),
		}},
	}, nil)
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	var tokens, originals int
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokens++
			continue
		}
		originals++
	}
	if tokens != 3 {
		t.Fatalf("tokens created = %d, want 3", tokens)
	}
	if originals != 0 {
		t.Fatalf("original creatures remaining = %d, want 0", originals)
	}
}

// TestSacrificeAnyNumberPublishesCountForScaledAddMana models Mana Seism:
// sacrificing a player-chosen number of lands publishes the count so the
// follow-up adds exactly that much mana.
func TestSacrificeAnyNumberPublishesCountForScaledAddMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for _, name := range []string{"Forest", "Mountain", "Island"} {
		addBattlefieldPermanent(g, game.Player1, name, []types.Card{types.Land})
	}
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{
			Primitive: game.SacrificePermanents{
				AnyNumber: true,
				Player:    game.ControllerReference(),
				Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
			},
			PublishResult: sacrificedThisWayKey,
		},
		{Primitive: game.AddMana{
			Amount:    game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountPreviousEffectResult, ResultKey: sacrificedThisWayKey}),
			ManaColor: mana.C,
		}},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0, 1}}},
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 2 {
		t.Fatalf("colorless mana added = %d, want 2", got)
	}
	var lands int
	for range g.Battlefield {
		lands++
	}
	if lands != 1 {
		t.Fatalf("lands remaining = %d, want 1", lands)
	}
}
