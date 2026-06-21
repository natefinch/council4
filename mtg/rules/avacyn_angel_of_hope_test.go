package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestAvacynOtherControlledPermanentsGainIndestructible proves that the
// "Other permanents you control have indestructible." static (Avacyn, Angel of
// Hope) grants indestructible to every other permanent its controller owns —
// creatures and non-creatures alike — while excluding the source itself and
// permanents controlled by opponents.
func TestAvacynOtherControlledPermanentsGainIndestructible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Avacyn, Angel of Hope",
		Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 8}), Toughness: opt.Val(game.PT{Value: 8}),
	}})
	otherCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	otherArtifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Controlled Artifact",
		Types: []types.Card{types.Artifact},
	}})
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: source.ObjectID,
		Layer:          game.LayerAbility,
		Group: game.ObjectControlledGroupExcluding(
			game.SourcePermanentReference(),
			game.Selection{},
			game.SourcePermanentReference(),
		),
		AddKeywords: []game.Keyword{game.Indestructible},
	})

	if !hasKeyword(g, otherCreature, game.Indestructible) {
		t.Fatal("another creature you control should gain indestructible")
	}
	if !hasKeyword(g, otherArtifact, game.Indestructible) {
		t.Fatal("a non-creature permanent you control should gain indestructible")
	}
	if hasKeyword(g, source, game.Indestructible) {
		t.Fatal("the source itself must be excluded by the \"Other\" qualifier")
	}
	if hasKeyword(g, opponentCreature, game.Indestructible) {
		t.Fatal("an opponent's permanent must not gain indestructible")
	}
}
