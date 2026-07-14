package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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

// TestCounterRecipientSelectionExcludesSource proves that an ExcludeSource
// recipient filter ("another creature you control", Benevolent Hydra) modifies
// counters placed on other controlled creatures while leaving the source's own
// counters unmodified.
func TestCounterRecipientSelectionExcludesSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	hydra := &game.CardDef{CardFace: game.CardFace{
		Name:  "Benevolent Hydra",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.ControlledPermanentSelectionCounterKindPlacementReplacement(
				"If one or more +1/+1 counters would be put on another creature you control, that many plus one +1/+1 counters are put on it instead.",
				1, 1,
				counter.PlusOnePlusOne,
				game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludeSource: true},
				game.TriggerControllerYou,
			),
		},
	}}
	hydraPermanent := addReplacementPermanent(t, g, game.Player1, hydra)

	otherCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Other Creature",
		Types: []types.Card{types.Creature},
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
		{"another controlled creature (match)", otherCreature, 3},
		{"source creature (excluded)", hydraPermanent, 2},
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

func TestEntersUntappedGroupOverridesTappedLandEntry(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spelunking := &game.CardDef{CardFace: game.CardFace{
		Name:  "Spelunking",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersUntappedGroupReplacement(
				"Lands you control enter untapped.",
				game.TriggerControllerYou,
				types.Land,
			),
		},
	}}
	addReplacementPermanent(t, g, game.Player1, spelunking)
	tapland := &game.CardDef{CardFace: game.CardFace{
		Name:                 "Tapland",
		Types:                []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{game.EntersTappedReplacement("This land enters tapped.")},
	}}
	permanent := addReplacementPermanent(t, g, game.Player1, tapland)
	if permanent.Tapped {
		t.Fatal("land entered tapped despite Spelunking")
	}
	tappedCreature := &game.CardDef{CardFace: game.CardFace{
		Name:                 "Tapped Creature",
		Types:                []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{game.EntersTappedReplacement("This creature enters tapped.")},
	}}
	creature := addReplacementPermanent(t, g, game.Player1, tappedCreature)
	if !creature.Tapped {
		t.Fatal("nonland permanent entered untapped under land-only replacement")
	}
	forcedID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Forced Tapped Land",
		Types: []types.Card{types.Land},
	}})
	forcedCard, ok := g.GetCardInstance(forcedID)
	if !ok {
		t.Fatal("forced-tapped land card not found")
	}
	prepared, ok := prepareCardPermanentFaceForSimultaneousEntry(
		NewEngine(nil),
		g,
		forcedCard,
		game.Player1,
		zone.Hand,
		game.FaceFront,
		nil,
		permanentCreationOptions{ForceTapped: true},
		[game.NumPlayers]PlayerAgent{},
		&TurnLog{},
	)
	if !ok || prepared.permanent.Tapped {
		t.Fatalf("forced-tapped land prepared as %#v, want untapped", prepared.permanent)
	}
}
