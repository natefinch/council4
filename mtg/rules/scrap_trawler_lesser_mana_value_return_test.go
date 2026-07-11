package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lesserManaValueArtifactDef(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
		Types:    []types.Card{types.Artifact},
	}}
}

// TestScrapTrawlerReturnTargetsLesserManaValueArtifact covers the event-relative
// "with lesser mana value" graveyard-return target filter that Scrap Trawler's
// self-or-another-artifact death trigger uses: Selection.
// ManaValueLessThanEventPermanent requires the candidate graveyard card's mana
// value to be strictly less than the mana value of the artifact named by the
// triggering event (the artifact that was put into the graveyard). A died
// artifact of mana value 3 lets the return target a lesser-mana-value artifact
// card (2) but not one of equal (3) or greater (4) mana value, the artifact
// filter still excludes a lesser non-artifact, and without a triggering event
// permanent the bound fails closed.
func TestScrapTrawlerReturnTargetsLesserManaValueArtifact(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	diedArtifact := addCardInstance(g, game.Player1, lesserManaValueArtifactDef("Died Artifact", 3))
	lesser := addCardInstance(g, game.Player1, lesserManaValueArtifactDef("Lesser Artifact", 2))
	equal := addCardInstance(g, game.Player1, lesserManaValueArtifactDef("Equal Artifact", 3))
	greater := addCardInstance(g, game.Player1, lesserManaValueArtifactDef("Greater Artifact", 4))
	lesserNonArtifact := addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Lesser Soldier",
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Soldier},
	}})
	for _, cardID := range []game.ObjectID{diedArtifact, lesser, equal, greater, lesserNonArtifact} {
		g.Players[game.Player1].Graveyard.Add(cardID)
	}

	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection: opt.Val(game.Selection{
			RequiredTypes:                   []types.Card{types.Artifact},
			Controller:                      game.ControllerYou,
			ManaValueLessThanEventPermanent: true,
		}),
	}

	// The died artifact (mana value 3) is the permanent named by the trigger event.
	event := game.Event{Kind: game.EventZoneChanged, FromZone: zone.Battlefield, ToZone: zone.Graveyard, CardID: diedArtifact, PermanentID: g.IDGen.Next()}

	if !targetMatchesSpec(g, game.Player1, 0, event, &spec, game.CardTarget(lesser)) {
		t.Fatal("an artifact card with lesser mana value than the died artifact should be a legal target")
	}
	if targetMatchesSpec(g, game.Player1, 0, event, &spec, game.CardTarget(equal)) {
		t.Fatal("an artifact card with equal mana value must not match a lesser-mana-value filter")
	}
	if targetMatchesSpec(g, game.Player1, 0, event, &spec, game.CardTarget(greater)) {
		t.Fatal("an artifact card with greater mana value must not match a lesser-mana-value filter")
	}
	if targetMatchesSpec(g, game.Player1, 0, event, &spec, game.CardTarget(lesserNonArtifact)) {
		t.Fatal("a lesser-mana-value non-artifact card must not match the artifact filter")
	}
	if targetMatchesSpec(g, game.Player1, 0, game.Event{}, &spec, game.CardTarget(lesser)) {
		t.Fatal("without a triggering event permanent the mana-value bound must fail closed")
	}
}
