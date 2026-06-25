package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestCounterRecipientSelectionMatchesTypeUnion proves that routing a counter
// replacement's recipient filter through the canonical matchSelection preserves
// the "an artifact or creature you control" union recipient (Ozolith, the
// Shattered Spire): counters placed on a controlled artifact or creature are
// modified, while a controlled land, and any permanent an opponent controls,
// are unaffected.
func TestCounterRecipientSelectionMatchesTypeUnion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ozolith := &game.CardDef{CardFace: game.CardFace{
		Name:  "Ozolith, the Shattered Spire",
		Types: []types.Card{types.Artifact},
		ReplacementAbilities: []game.ReplacementAbility{
			game.ControlledPermanentTypesCounterPlacementReplacement(
				"If one or more +1/+1 counters would be put on an artifact or creature you control, that many plus one +1/+1 counters are put on it instead.",
				1, 1,
				[]types.Card{types.Artifact, types.Creature},
				game.TriggerControllerYou,
			),
		},
	}}
	addReplacementPermanent(t, g, game.Player1, ozolith)

	myArtifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "My Artifact",
		Types: []types.Card{types.Artifact},
	}})
	myCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "My Creature",
		Types: []types.Card{types.Creature},
	}})
	myLand := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "My Land",
		Types: []types.Card{types.Land},
	}})
	theirCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Their Creature",
		Types: []types.Card{types.Creature},
	}})

	cases := []struct {
		name      string
		permanent *game.Permanent
		want      int
	}{
		{"controlled artifact (union match)", myArtifact, 3},
		{"controlled creature (union match)", myCreature, 3},
		{"controlled land (no match)", myLand, 2},
		{"opponent creature (controller filter)", theirCreature, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !addCountersToPermanent(g, tc.permanent, counter.PlusOnePlusOne, 2) {
				t.Fatal("addCountersToPermanent() = false, want true")
			}
			if got := tc.permanent.Counters.Get(counter.PlusOnePlusOne); got != tc.want {
				t.Fatalf("+1/+1 counters = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestEntersTappedSelectionMatchesTypeUnion proves that routing a group
// enters-tapped replacement's filter through the canonical matchSelection
// preserves the "any of these card types" recipient union: an opponent's
// artifact or creature enters tapped, while an opponent's land and the
// controller's own permanents do not.
func TestEntersTappedSelectionMatchesTypeUnion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	suppressor := &game.CardDef{CardFace: game.CardFace{
		Name:  "Type Suppressor",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedGroupReplacement(
				"Artifacts and creatures your opponents control enter tapped.",
				game.TriggerControllerOpponent,
				types.Artifact, types.Creature,
			),
		},
	}}
	addReplacementPermanent(t, g, game.Player1, suppressor)

	withType := func(name string, cardType types.Card) *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{cardType}}}
	}

	cases := []struct {
		name       string
		controller game.PlayerID
		cardType   types.Card
		wantTapped bool
	}{
		{"opponent artifact", game.Player2, types.Artifact, true},
		{"opponent creature", game.Player2, types.Creature, true},
		{"opponent land", game.Player2, types.Land, false},
		{"own creature", game.Player1, types.Creature, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			permanent := addReplacementPermanent(t, g, tc.controller, withType(tc.name, tc.cardType))
			if permanent.Tapped != tc.wantTapped {
				t.Fatalf("entered tapped = %t, want %t", permanent.Tapped, tc.wantTapped)
			}
		})
	}
}
