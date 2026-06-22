package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestBecomeTypeAddsCardTypeUntilEndOfTurn proves the lowered shape of
// Liquimetal Torque's "Target permanent becomes an artifact in addition to its
// other types until end of turn.": an ApplyContinuous at LayerType that adds the
// artifact card type to the target permanent while leaving its existing types
// intact, and which expires at end of turn.
func TestBecomeTypeAddsCardTypeUntilEndOfTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Liquimetal Torque",
		Types: []types.Card{types.Artifact},
	}})
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Forest",
		Types: []types.Card{types.Land},
	}})

	if permanentHasType(g, target, types.Artifact) {
		t.Fatal("target is an artifact before the effect, want only land")
	}

	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}
	if !applyTypedContinuousEffects(g, obj, target, []game.ContinuousEffect{{
		Layer:    game.LayerType,
		AddTypes: []types.Card{types.Artifact},
	}}, game.DurationUntilEndOfTurn) {
		t.Fatal("applyTypedContinuousEffects returned false")
	}

	if !permanentHasType(g, target, types.Artifact) {
		t.Fatal("target did not gain the artifact type")
	}
	if !permanentHasType(g, target, types.Land) {
		t.Fatal("target lost its land type, want it retained in addition")
	}

	expireCleanupDurations(g)

	if permanentHasType(g, target, types.Artifact) {
		t.Fatal("target retained the artifact type after end of turn, want it expired")
	}
	if !permanentHasType(g, target, types.Land) {
		t.Fatal("target lost its land type after the effect expired")
	}
}
