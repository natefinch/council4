package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestKickerAdditionalCostsComposeWithEveryCostOption(t *testing.T) {
	t.Parallel()
	sacrifice := cost.Additional{
		Kind:               cost.AdditionalSacrifice,
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Creature,
	}
	card := &game.CardDef{CardFace: game.CardFace{
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.G}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{
				AdditionalCosts: []cost.Additional{sacrifice},
			}},
		}},
		AlternativeCosts: []cost.Alternative{{
			Label:    "Alternative",
			ManaCost: opt.Val(cost.Mana{cost.G}),
		}},
	}}
	options := spellCostOptionsForZoneAndKicker(
		manaAlternativeState{},
		game.Player1,
		card,
		zone.Hand,
		true,
		0,
		false,
		nil,
	)
	if len(options) != 2 {
		t.Fatalf("options = %#v, want normal and alternative", options)
	}
	for _, option := range options {
		if len(option.additionalCosts) != 1 || option.additionalCosts[0].Kind != cost.AdditionalSacrifice {
			t.Fatalf("%s additional costs = %#v, want Kicker sacrifice", option.label, option.additionalCosts)
		}
	}
}

func TestMultikickerRepeatsAdditionalCosts(t *testing.T) {
	t.Parallel()
	kicker := game.KickerKeyword{
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 1}},
		Multi:           true,
	}
	got, ok := appendKickerAdditionalCosts(nil, kicker, true, true, 3)
	if !ok || len(got) != 3 {
		t.Fatalf("additional costs = %#v, want three repeated payments", got)
	}
}

func TestKickerAdditionalChoiceGroupsStayIndependent(t *testing.T) {
	t.Parallel()
	existing := []cost.Additional{{Kind: cost.AdditionalDiscard, ChoiceGroup: 1}}
	kicker := game.KickerKeyword{
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalSacrifice, ChoiceGroup: 1},
			{Kind: cost.AdditionalPayLife, ChoiceGroup: 1},
		},
		Multi: true,
	}
	got, ok := appendKickerAdditionalCosts(existing, kicker, true, true, 2)
	if !ok || len(got) != 5 {
		t.Fatalf("additional costs = %#v, want existing plus two Kicker choices", got)
	}
	wantGroups := []uint8{1, 2, 2, 3, 3}
	for i, want := range wantGroups {
		if got[i].ChoiceGroup != want {
			t.Fatalf("choice groups = %#v, want %v", got, wantGroups)
		}
	}
}

func TestAlternativeRequestIncludesCompleteKickerCost(t *testing.T) {
	t.Parallel()
	card := &game.CardDef{CardFace: game.CardFace{
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{
				Cost:            cost.Mana{cost.PhyrexianMana(mana.W)},
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 1}},
			}},
		}},
	}}
	options := spellCostOptionsForRequest(manaAlternativeState{}, SpellRequest{
		PlayerID:    game.Player1,
		SourceZone:  zone.Hand,
		Card:        card,
		KickerPaid:  true,
		KickerCount: 1,
		Alternative: opt.Val(cost.Alternative{ManaCost: opt.Val(cost.Mana{})}),
	})
	if len(options) != 1 ||
		options[0].manaCost == nil ||
		len(*options[0].manaCost) != 1 ||
		len(options[0].additionalCosts) != 1 {
		t.Fatalf("alternative options = %#v, want complete mana and nonmana Kicker cost", options)
	}
}

func TestKickerPaymentCountIsBounded(t *testing.T) {
	t.Parallel()
	card := &game.CardDef{CardFace: game.CardFace{
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 1}},
				Multi:           true,
			}},
		}},
	}}
	if options := spellCostOptionsForZoneAndKicker(
		manaAlternativeState{},
		game.Player1,
		card,
		zone.Hand,
		true,
		maxKickerPaymentCount+1,
		false,
		nil,
	); len(options) != 0 {
		t.Fatalf("options = %#v, want excessive Multikicker count rejected", options)
	}
	if options := spellCostOptionsForZoneAndKicker(
		manaAlternativeState{},
		game.Player1,
		card,
		zone.Hand,
		false,
		-1,
		false,
		nil,
	); len(options) != 0 {
		t.Fatalf("options = %#v, want negative Multikicker count rejected", options)
	}
}
