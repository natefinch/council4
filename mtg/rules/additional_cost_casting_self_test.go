package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// demandAnswersLikeSpell is a zero-mana instant whose additional cost is "sacrifice
// an artifact or discard a card", mirroring Demand Answers. The zero mana cost
// isolates the additional-cost payability from mana so the test needs no lands.
func demandAnswersLikeSpell(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:         name,
		ManaCost:     opt.Val(cost.Mana{cost.O(0)}),
		Types:        []types.Card{types.Instant},
		SpellAbility: opt.Val(game.AbilityContent{}),
		AdditionalCosts: []cost.Additional{
			{
				Kind:               cost.AdditionalSacrifice,
				Text:               "sacrifice an artifact",
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Artifact,
				ChoiceGroup:        1,
			},
			{
				Kind:        cost.AdditionalDiscard,
				Text:        "discard a card",
				Amount:      1,
				Source:      zone.Hand,
				ChoiceGroup: 1,
			},
		},
	}}
}

func castActionForCard(actions []action.Action, cardID id.ID) bool {
	for _, a := range actions {
		if a.Kind != action.ActionCastSpell {
			continue
		}
		if cast, ok := a.CastSpellPayload(); ok && cast.CardID == cardID {
			return true
		}
	}
	return false
}

// TestSpellCannotPayCardCostWithItself covers the CR 601.2a timing rule: a spell
// moves to the stack before its costs are paid, so it cannot itself pay one of its
// own card costs. A "sacrifice an artifact or discard a card" spell that is the
// only card in hand (and whose controller has no artifact) is therefore not
// castable — legalActions must not offer it, or applyAction rejects a "legal"
// action and the priority loop panics.
func TestSpellCannotPayCardCostWithItself(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, demandAnswersLikeSpell("Demand Answers"))
	setMainPhasePriority(g, game.Player1)

	legal := e.Simulator().LegalActions(g, game.Player1)
	if castActionForCard(legal, spellID) {
		t.Fatal("spell offered for casting although its discard cost cannot be paid (it is the only card in hand)")
	}
}

// TestSpellCanPayCardCostWithAnotherCard is the positive control: with a second
// card in hand to discard, the same spell is castable.
func TestSpellCanPayCardCostWithAnotherCard(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, demandAnswersLikeSpell("Demand Answers"))
	addCardToHand(g, game.Player1, evidenceCard("Spare Card", 1))
	setMainPhasePriority(g, game.Player1)

	legal := e.Simulator().LegalActions(g, game.Player1)
	if !castActionForCard(legal, spellID) {
		t.Fatal("spell not offered for casting although a second card is available to discard")
	}
}
