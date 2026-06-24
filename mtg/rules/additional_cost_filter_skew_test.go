package rules

import (
	"maps"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// The choice layer (candidateSacrificePermanents / candidateAdditionalCostCards)
// and the payment planner (additionalCostMatchesPermanent / ...Card) historically
// filtered additional-cost objects with two divergent matchers. These tests pin
// that both layers now derive their eligible set from one canonical conversion
// (payment.SelectionForAdditionalCost) and one matcher (matchSelection), so the
// presented candidates and the planner-accepted objects are always identical.

func sacrificeChoicePermanentIDs(g *game.Game, playerID game.PlayerID, addCost cost.Additional) map[id.ID]bool {
	set := map[id.ID]bool{}
	for _, permanent := range candidateSacrificePermanents(g, playerID, addCost, nil) {
		set[permanent.ObjectID] = true
	}
	return set
}

// sacrificePlannerPermanentIDs mirrors the planner's per-permanent acceptance
// gate for a sacrifice cost: controlled, on the battlefield, and accepted by the
// canonical additional-cost matcher reached through payment.State.
func sacrificePlannerPermanentIDs(g *game.Game, playerID game.PlayerID, addCost cost.Additional) map[id.ID]bool {
	state := &rulesPaymentState{g: g}
	sel, ok := payment.SelectionForAdditionalCost(addCost)
	set := map[id.ID]bool{}
	if !ok {
		return set
	}
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) || permanent.Controller != playerID {
			continue
		}
		if state.PermanentMatchesSelection(permanent, sel) {
			set[permanent.ObjectID] = true
		}
	}
	return set
}

func exileChoiceCardIDs(g *game.Game, playerID game.PlayerID, addCost cost.Additional) map[id.ID]bool {
	set := map[id.ID]bool{}
	for _, cardID := range candidateAdditionalCostCards(g, playerID, addCost) {
		set[cardID] = true
	}
	return set
}

// exilePlannerCardIDs mirrors the planner's per-card acceptance gate for a card
// cost: the card is in the cost's source zone and accepted by the canonical
// matcher reached through payment.State.
func exilePlannerCardIDs(g *game.Game, playerID game.PlayerID, addCost cost.Additional) map[id.ID]bool {
	state := &rulesPaymentState{g: g}
	sel, ok := payment.SelectionForAdditionalCost(addCost)
	set := map[id.ID]bool{}
	if !ok {
		return set
	}
	for _, cardID := range g.Players[playerID].Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		if state.CardMatchesSelection(cardFaceOrDefault(card, game.FaceFront), sel) {
			set[cardID] = true
		}
	}
	return set
}

func filterTestCardDef(name string, cardTypes []types.Card, supertypes []types.Super, subtypes []types.Sub, colors []color.Color) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Types:      cardTypes,
		Supertypes: supertypes,
		Subtypes:   subtypes,
		Colors:     colors,
	}}
}

func TestAdditionalCostPermanentFilterChoiceMatchesPlanner(t *testing.T) {
	tests := []struct {
		name     string
		addCost  cost.Additional
		defs     map[string]*game.CardDef
		eligible []string
	}{
		{
			name: "artifact or creature union",
			addCost: cost.Additional{
				Kind:               cost.AdditionalSacrifice,
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Artifact,
				PermanentTypeAlt:   types.Creature,
			},
			defs: map[string]*game.CardDef{
				"creature":         filterTestCardDef("Creature", []types.Card{types.Creature}, nil, nil, nil),
				"artifact":         filterTestCardDef("Artifact", []types.Card{types.Artifact}, nil, nil, nil),
				"artifactCreature": filterTestCardDef("Artifact Creature", []types.Card{types.Artifact, types.Creature}, nil, nil, nil),
				"land":             filterTestCardDef("Land", []types.Card{types.Land}, nil, nil, nil),
			},
			eligible: []string{"creature", "artifact", "artifactCreature"},
		},
		{
			name: "black creature",
			addCost: cost.Additional{
				Kind:               cost.AdditionalSacrifice,
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Creature,
				MatchCardColor:     true,
				CardColor:          color.Black,
			},
			defs: map[string]*game.CardDef{
				"blackCreature":      filterTestCardDef("Black Creature", []types.Card{types.Creature}, nil, nil, []color.Color{color.Black}),
				"whiteCreature":      filterTestCardDef("White Creature", []types.Card{types.Creature}, nil, nil, []color.Color{color.White}),
				"blackArtifact":      filterTestCardDef("Black Artifact", []types.Card{types.Artifact}, nil, nil, []color.Color{color.Black}),
				"blackGreenCreature": filterTestCardDef("Black-Green Creature", []types.Card{types.Creature}, nil, nil, []color.Color{color.Black, color.Green}),
			},
			eligible: []string{"blackCreature", "blackGreenCreature"},
		},
		{
			name: "historic permanent",
			addCost: cost.Additional{
				Kind:          cost.AdditionalSacrifice,
				Amount:        1,
				MatchHistoric: true,
			},
			defs: map[string]*game.CardDef{
				"artifact":  filterTestCardDef("Artifact", []types.Card{types.Artifact}, nil, nil, nil),
				"legendary": filterTestCardDef("Legendary Creature", []types.Card{types.Creature}, []types.Super{types.Legendary}, nil, nil),
				"saga":      filterTestCardDef("Saga", []types.Card{types.Enchantment}, nil, []types.Sub{types.Saga}, nil),
				"vanilla":   filterTestCardDef("Vanilla Creature", []types.Card{types.Creature}, nil, nil, nil),
			},
			eligible: []string{"artifact", "legendary", "saga"},
		},
		{
			name: "legendary supertype",
			addCost: cost.Additional{
				Kind:             cost.AdditionalSacrifice,
				Amount:           1,
				RequireSupertype: types.Legendary,
			},
			defs: map[string]*game.CardDef{
				"legendary":    filterTestCardDef("Legendary Creature", []types.Card{types.Creature}, []types.Super{types.Legendary}, nil, nil),
				"nonlegendary": filterTestCardDef("Nonlegendary Creature", []types.Card{types.Creature}, nil, nil, nil),
			},
			eligible: []string{"legendary"},
		},
		{
			name: "goblin or orc subtype",
			addCost: cost.Additional{
				Kind:        cost.AdditionalSacrifice,
				Amount:      1,
				SubtypesAny: cost.SubtypeSet{types.Goblin, types.Orc},
			},
			defs: map[string]*game.CardDef{
				"goblin": filterTestCardDef("Goblin", []types.Card{types.Creature}, nil, []types.Sub{types.Goblin}, nil),
				"orc":    filterTestCardDef("Orc", []types.Card{types.Creature}, nil, []types.Sub{types.Orc}, nil),
				"elf":    filterTestCardDef("Elf", []types.Card{types.Creature}, nil, []types.Sub{types.Elf}, nil),
			},
			eligible: []string{"goblin", "orc"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			ids := map[string]id.ID{}
			for key, def := range test.defs {
				ids[key] = addCombatPermanent(g, game.Player1, def).ObjectID
			}

			want := map[id.ID]bool{}
			for _, key := range test.eligible {
				want[ids[key]] = true
			}

			choice := sacrificeChoicePermanentIDs(g, game.Player1, test.addCost)
			planner := sacrificePlannerPermanentIDs(g, game.Player1, test.addCost)

			if !maps.Equal(choice, planner) {
				t.Fatalf("choice set %v != planner set %v", choice, planner)
			}
			if !maps.Equal(choice, want) {
				t.Fatalf("eligible set %v, want %v", choice, want)
			}
		})
	}
}

func TestAdditionalCostCardFilterChoiceMatchesPlanner(t *testing.T) {
	addCost := cost.Additional{Kind: cost.AdditionalExile, MatchHistoric: true}
	defs := map[string]*game.CardDef{
		"artifact":  filterTestCardDef("Artifact Card", []types.Card{types.Artifact}, nil, nil, nil),
		"legendary": filterTestCardDef("Legendary Card", []types.Card{types.Creature}, []types.Super{types.Legendary}, nil, nil),
		"saga":      filterTestCardDef("Saga Card", []types.Card{types.Enchantment}, nil, []types.Sub{types.Saga}, nil),
		"vanilla":   filterTestCardDef("Vanilla Card", []types.Card{types.Creature}, nil, nil, nil),
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ids := map[string]id.ID{}
	for key, def := range defs {
		ids[key] = addCardToGraveyard(g, game.Player1, def)
	}

	want := map[id.ID]bool{
		ids["artifact"]:  true,
		ids["legendary"]: true,
		ids["saga"]:      true,
	}

	choice := exileChoiceCardIDs(g, game.Player1, addCost)
	planner := exilePlannerCardIDs(g, game.Player1, addCost)

	if !maps.Equal(choice, planner) {
		t.Fatalf("choice set %v != planner set %v", choice, planner)
	}
	if !maps.Equal(choice, want) {
		t.Fatalf("eligible set %v, want %v", choice, want)
	}
}
