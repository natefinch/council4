package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

func cardInfo(name string, manaValue int, cardTypes ...types.Card) opt.V[game.ChoiceCardInfo] {
	return opt.Val(game.ChoiceCardInfo{Name: name, ManaValue: manaValue, Types: cardTypes})
}

func scryRequest(subject opt.V[game.ChoiceCardInfo]) game.ChoiceRequest {
	return game.ChoiceRequest{
		Kind:       game.ChoiceScry,
		Options:    []game.ChoiceOption{{Index: 0, Label: "top"}, {Index: 1, Label: "bottom"}},
		MinChoices: 1,
		MaxChoices: 1,
		Subject:    subject,
	}
}

func observationWithLands(t *testing.T, lands int) rules.PlayerObservation {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range lands {
		addObservedPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:  "Forest",
			Types: []types.Card{types.Land},
		}})
	}
	return rules.NewObservation(g, game.Player1)
}

func TestScryKeepsLandWhenShortBottomsWhenFlooded(t *testing.T) {
	strategy := GenericStrategy{}
	land := cardInfo("Forest", 0, types.Land)

	short := strategy.ChooseChoice(observationWithLands(t, 2), scryRequest(land))
	if len(short) != 1 || short[0] != placementKeepTop {
		t.Errorf("scry with few lands = %v, want keep on top", short)
	}

	flooded := strategy.ChooseChoice(observationWithLands(t, 7), scryRequest(land))
	if len(flooded) != 1 || flooded[0] != placementElsewhere {
		t.Errorf("scry while flooded = %v, want bottom", flooded)
	}
}

func TestScryKeepsCastableSpellBottomsExpensive(t *testing.T) {
	strategy := GenericStrategy{}
	obs := observationWithLands(t, 3)

	castable := strategy.ChooseChoice(obs, scryRequest(cardInfo("Bear", 2, types.Creature)))
	if len(castable) != 1 || castable[0] != placementKeepTop {
		t.Errorf("scry a castable spell = %v, want keep on top", castable)
	}

	expensive := strategy.ChooseChoice(obs, scryRequest(cardInfo("Dragon", 7, types.Creature)))
	if len(expensive) != 1 || expensive[0] != placementElsewhere {
		t.Errorf("scry an uncastable spell = %v, want bottom", expensive)
	}
}

func TestScryUnknownSubjectKeepsTop(t *testing.T) {
	strategy := GenericStrategy{}
	got := strategy.ChooseChoice(observationWithLands(t, 3), scryRequest(opt.V[game.ChoiceCardInfo]{}))
	if len(got) != 1 || got[0] != placementKeepTop {
		t.Errorf("scry with unknown subject = %v, want keep on top (conservative)", got)
	}
}

func TestSurveilGraveyardsExpensiveCard(t *testing.T) {
	strategy := GenericStrategy{}
	request := game.ChoiceRequest{
		Kind:       game.ChoiceSurveil,
		Options:    []game.ChoiceOption{{Index: 0, Label: "top"}, {Index: 1, Label: "graveyard"}},
		MinChoices: 1,
		MaxChoices: 1,
		Subject:    cardInfo("Dragon", 8, types.Creature),
	}
	got := strategy.ChooseChoice(observationWithLands(t, 2), request)
	if len(got) != 1 || got[0] != placementElsewhere {
		t.Errorf("surveil an uncastable card = %v, want graveyard", got)
	}
}

func TestPaymentLosesLeastValuableCard(t *testing.T) {
	strategy := GenericStrategy{}
	request := game.ChoiceRequest{
		Kind: game.ChoicePayment,
		Options: []game.ChoiceOption{
			{Index: 0, Label: "Bomb", Card: cardInfo("Bomb", 6, types.Creature)},
			{Index: 1, Label: "Token", Card: cardInfo("Spirit", 1, types.Creature)},
			{Index: 2, Label: "Midrange", Card: cardInfo("Bear", 3, types.Creature)},
		},
		MinChoices: 1,
		MaxChoices: 1,
	}
	got := strategy.ChooseChoice(rules.PlayerObservation{}, request)
	if len(got) != 1 || got[0] != 1 {
		t.Errorf("payment selection = %v, want the cheapest option (index 1, MV 1)", got)
	}
}

func TestPaymentLosesCheapestTwoWhenForced(t *testing.T) {
	strategy := GenericStrategy{}
	request := game.ChoiceRequest{
		Kind: game.ChoicePayment,
		Options: []game.ChoiceOption{
			{Index: 0, Label: "A", Card: cardInfo("A", 5, types.Creature)},
			{Index: 1, Label: "B", Card: cardInfo("B", 1, types.Creature)},
			{Index: 2, Label: "C", Card: cardInfo("C", 2, types.Creature)},
		},
		MinChoices: 2,
		MaxChoices: 2,
	}
	got := strategy.ChooseChoice(rules.PlayerObservation{}, request)
	if len(got) != 2 {
		t.Fatalf("payment selection = %v, want 2 options", got)
	}
	chosen := map[int]bool{got[0]: true, got[1]: true}
	if !chosen[1] || !chosen[2] {
		t.Errorf("forced to lose 2, selection %v should be the two cheapest (indices 1 and 2)", got)
	}
}

func TestPaymentNonCardOptionsFallBack(t *testing.T) {
	strategy := GenericStrategy{}
	// A hybrid mana-or-life payment carries no card info; the heuristic must not
	// apply and the baseline default is used.
	request := game.ChoiceRequest{
		Kind: game.ChoicePayment,
		Options: []game.ChoiceOption{
			{Index: 0, Label: "Pay mana"},
			{Index: 1, Label: "Pay 2 life"},
		},
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
	got := strategy.ChooseChoice(rules.PlayerObservation{}, request)
	if len(got) != 1 || got[0] != 0 {
		t.Errorf("non-card payment = %v, want the baseline default [0]", got)
	}
}
