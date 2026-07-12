package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestExchangeLifeTotalWithSourceToughnessUsesEffectiveValue(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 7
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Tree of Redemption",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 13}),
	}})
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		Layer:            game.LayerPowerToughnessModify,
		AffectedObjectID: source.ObjectID,
		ToughnessDelta:   2,
	})
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		Controller:   game.Player1,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
	}
	resolver := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}
	result := handleExchangeLifeTotalWithSourceCharacteristic(
		resolver,
		game.ExchangeLifeTotalWithSourceCharacteristic{
			Player:         game.ControllerReference(),
			Characteristic: game.SourceToughness,
		},
	)
	if !result.succeeded || g.Players[game.Player1].Life != 15 {
		t.Fatalf("result = %#v, life = %d, want successful exchange to 15", result, g.Players[game.Player1].Life)
	}
	if toughness, ok := effectiveToughness(g, source); !ok || toughness != 9 {
		t.Fatalf("effective toughness = %d (ok=%v), want former life 7 plus modifier 2", toughness, ok)
	}
}

func TestExchangeTargetOpponentLifeWithSourceToughness(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player2].Life = 40
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Tree of Perdition",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 13}),
	}})
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		Controller:   game.Player1,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Targets:      []game.Target{{Kind: game.TargetPlayer, PlayerID: game.Player2}},
	}
	resolver := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}
	result := handleExchangeLifeTotalWithSourceCharacteristic(
		resolver,
		game.ExchangeLifeTotalWithSourceCharacteristic{
			Player:         game.TargetPlayerReference(0),
			Characteristic: game.SourceToughness,
		},
	)
	if !result.succeeded || g.Players[game.Player2].Life != 13 {
		t.Fatalf("result = %#v, opponent life = %d", result, g.Players[game.Player2].Life)
	}
	if toughness, ok := effectiveToughness(g, source); !ok || toughness != 40 {
		t.Fatalf("effective toughness = %d (ok=%v), want 40", toughness, ok)
	}
}

func TestExchangeFailsAtomicallyWhenPlayerCantGainLife(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 7
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Tree of Redemption",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 13}),
	}})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectCantGainLife,
		AffectedPlayer: game.PlayerYou,
		Controller:     game.Player1,
	})
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		Controller:   game.Player1,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
	}
	resolver := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}
	result := handleExchangeLifeTotalWithSourceCharacteristic(
		resolver,
		game.ExchangeLifeTotalWithSourceCharacteristic{
			Player:         game.ControllerReference(),
			Characteristic: game.SourceToughness,
		},
	)
	if result.succeeded || g.Players[game.Player1].Life != 7 || len(g.ContinuousEffects) != 0 {
		t.Fatalf("result = %#v, life = %d, effects = %d; want atomic no-op",
			result, g.Players[game.Player1].Life, len(g.ContinuousEffects))
	}
}
