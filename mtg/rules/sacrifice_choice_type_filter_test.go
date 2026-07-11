package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestSacrificePermanentsEachPlayerPlaneswalkerSelection proves the single-type
// planeswalker sacrifice selection (Sheoldred's Edict's planeswalker mode,
// Angrath's Rampage) restricts each player's eligible set to planeswalkers: over
// the all-players group every player sacrifices a planeswalker they control
// while their creatures survive, and with exactly one eligible planeswalker each
// no player is asked to choose.
func TestSacrificePermanentsEachPlayerPlaneswalkerSelection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	pw1 := addColoredPermanent(g, game.Player1, "Walker One", nil, []types.Card{types.Planeswalker}, nil)
	creature1 := addCreaturePermanent(g, game.Player1)
	pw2 := addColoredPermanent(g, game.Player2, "Walker Two", nil, []types.Card{types.Planeswalker}, nil)
	creature2 := addCreaturePermanent(g, game.Player2)

	addEffectSpellToStack(g, game.Player1, game.SacrificePermanents{
		PlayerGroup: game.AllPlayersReference(),
		Amount:      game.Fixed(1),
		Selection:   game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
	}, nil)

	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want none (one eligible planeswalker per player)", log.Choices)
	}
	for _, sacrificed := range []*game.Permanent{pw1, pw2} {
		if _, ok := permanentByObjectID(g, sacrificed.ObjectID); ok {
			t.Fatalf("planeswalker %v survived, want sacrificed", sacrificed.ObjectID)
		}
	}
	for _, survivor := range []*game.Permanent{creature1, creature2} {
		if _, ok := permanentByObjectID(g, survivor.ObjectID); !ok {
			t.Fatalf("creature %v was sacrificed, want survived (planeswalker-only edict)", survivor.ObjectID)
		}
	}
}

// TestSacrificePermanentsCreatureTokenSelection proves the card-type token
// sacrifice selection (Sheoldred's Edict's creature-token mode, Gaius van
// Baelsar) restricts the eligible set to token creatures: a player controlling
// both a token creature and a nontoken creature sacrifices the token one, and
// with a single eligible token no choice is required.
func TestSacrificePermanentsCreatureTokenSelection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	tokenDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Zombie Token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
	tokenCreature := addCombatPermanent(g, game.Player2, tokenDef)
	tokenCreature.Token = true
	tokenCreature.TokenDef = tokenDef
	nontokenCreature := addCreaturePermanent(g, game.Player2)

	addEffectSpellToStack(g, game.Player1, game.SacrificePermanents{
		Player:    game.TargetPlayerReference(0),
		Amount:    game.Fixed(1),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true},
	}, []game.Target{game.PlayerTarget(game.Player2)})

	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want none (one eligible token creature)", log.Choices)
	}
	if _, ok := permanentByObjectID(g, tokenCreature.ObjectID); ok {
		t.Fatal("token creature survived, want sacrificed")
	}
	if _, ok := permanentByObjectID(g, nontokenCreature.ObjectID); !ok {
		t.Fatal("nontoken creature was sacrificed, want survived (token-only edict)")
	}
}
