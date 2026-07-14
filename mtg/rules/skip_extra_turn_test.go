package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestSkipExtraTurnsRuleAffectsOnlyOpponents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Extra Turn Suppressor",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectSkipExtraTurns,
			AffectedPlayer: game.PlayerOpponent,
		}}}},
	}})
	g.Turn.ExtraTurns = []game.PlayerID{game.Player1, game.Player2}
	next, ok := popExtraTurn(g)
	if !ok || next != game.Player1 {
		t.Fatalf("popExtraTurn = (%v, %v), want Player1 extra turn after opponent turn skipped", next, ok)
	}

	if len(g.Turn.ExtraTurns) != 0 {
		t.Fatalf("extra turns = %#v, want empty", g.Turn.ExtraTurns)
	}
}

func TestOpponentSecondDrawTriggerUsesEventPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Trouble Source",
		Types: []types.Card{types.Enchantment},
	}})
	pattern := &game.TriggerPattern{
		Event:                      game.EventCardDrawn,
		Player:                     game.TriggerPlayerOpponent,
		PlayerEventOrdinalThisTurn: 2,
	}
	if !triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:                       game.EventCardDrawn,
		Player:                     game.Player2,
		PlayerEventOrdinalThisTurn: 2,
	}) {
		t.Fatal("opponent second draw did not match")
	}
	if triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:                       game.EventCardDrawn,
		Player:                     game.Player1,
		PlayerEventOrdinalThisTurn: 2,
	}) {
		t.Fatal("controller second draw matched opponent trigger")
	}
}
