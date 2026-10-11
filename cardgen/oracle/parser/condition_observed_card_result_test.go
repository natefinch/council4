package parser

import (
	"slices"
	"testing"
)

func TestObservedCardResultOwnsLookAcrossIndependentRider(t *testing.T) {
	for _, test := range []struct {
		name, text string
		producer   int
	}{
		{"adjacent", "You may look at the top card of target opponent's library. If you do, you may put that card on the bottom of that player's library.", 1},
		{"independent rider", "You may look at the top card of target opponent's library. You gain 1 life. If you do, you may put that card on the bottom of that player's library.", 1},
		{"nonzero target", "Target player gains 1 life. You may look at the top card of target opponent's library. You gain 1 life. If you do, you may put that card on the bottom of that player's library.", 2},
		{"latest actual observation", "You may look at the top card of target opponent's library. You may look at the top card of target opponent's library. You gain 1 life. If you do, you may put that card on the bottom of that player's library.", 2},
		{"explicit look before reveal", "You may look at the top card of target opponent's library. Reveal the top card of target opponent's library. You gain 1 life. If you do, you may put the looked-at card on the bottom of that player's library.", 1},
		{"unproved reveal result refuses", "You may look at the top card of target opponent's library. Reveal the top card of target opponent's library. You gain 1 life. If you do, you may put that card on the bottom of that player's library.", 0},
		{"explicit revealed card refuses", "You may look at the top card of target opponent's library. Reveal the top card of target opponent's library. You gain 1 life. If you do, you may put the revealed card on the bottom of that player's library.", 0},
		{"existing exact reveal result", "You may reveal the top card of your library. If you do, you may put that card into your hand.", 1},
		{"existing observation payment", "Look at the top card of target player's library. If it's a nonland card, you may pay 2 life. If you do, put it into that player's graveyard.", 2},
		{"distinct optional payment", "You may look at the top card of target player's library. You may pay 2 life. If you do, put it into that player's graveyard.", 2},
		{"distinct optional loss", "You may look at the top card of target player's library. You may lose 2 life. If you do, put it into that player's graveyard.", 2},
		{"ordinary gain remains gain", "You gain 1 life. If you do, you gain 2 life.", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			document, diagnostics := Parse(test.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			found := false
			for _, segment := range document.Abilities[0].ConditionSegments {
				if !resultConditionPredicate(document.Abilities[0].ConditionClauses[segment.ClauseIndex].Predicate) {
					continue
				}
				found = true
				if segment.Ownership.ResultProducerClauseID != test.producer {
					t.Fatalf("producer=%d want=%d", segment.Ownership.ResultProducerClauseID, test.producer)
				}
			}
			if !found {
				t.Fatal("missing actual-result condition")
			}
		})
	}
}

func TestObservedCardResultRefusesUnownedObservation(t *testing.T) {
	const text = "You may look at the top card of target opponent's library. You gain 1 life. If you do, you may put that card on the bottom of that player's library."
	for _, test := range []struct {
		name   string
		mutate func(*Ability)
	}{
		{"canonical node missing", func(ability *Ability) { ability.SemanticReferences[0].NodeID = 99 }},
		{"canonical node duplicated", func(ability *Ability) {
			ability.SemanticReferences = append(ability.SemanticReferences, ability.SemanticReferences[0])
		}},
		{"canonical producer drifted", func(ability *Ability) { ability.SemanticReferences[0].ProducerClauseID = 2 }},
		{"canonical order drifted", func(ability *Ability) { ability.SemanticReferences[0].Order.Start++ }},
		{"library owner instead of actual card", func(ability *Ability) {
			ability.SemanticReferences[0].Kind = ReferenceThatPlayer
			ability.SemanticReferences[0].SubjectNoun = ObjectNounPlayer
			ability.Sentences[2].Effects[0].References[0].Kind = ReferenceThatPlayer
			ability.Sentences[2].Effects[0].References[0].SubjectNoun = ObjectNounPlayer
		}},
		{"reference unavailable", func(ability *Ability) { ability.SemanticReferences[0].LibraryCardObservation = false }},
		{"canonical clause duplicated", func(ability *Ability) {
			ability.Sentences[1].Effects[0].ClauseID = ability.Sentences[0].Effects[0].ClauseID
		}},
		{"forward observation", func(ability *Ability) { ability.Sentences[0].Effects[0].Order = ability.Sentences[2].Effects[0].Order }},
		{"missing observation order", func(ability *Ability) { ability.Sentences[0].Effects[0].Order.End = 0 }},
		{"unmodeled plurality", func(ability *Ability) { ability.Sentences[0].Effects[0].Amount.Value = 2 }},
		{"foreign context", func(ability *Ability) { ability.Sentences[0].Effects[0].Context = EffectContextEventPlayer }},
	} {
		t.Run(test.name, func(t *testing.T) {
			document, diagnostics := Parse(text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}

			ability := document.Abilities[0]
			ability.SemanticReferences = slices.Clone(ability.SemanticReferences)
			ability.ConditionSegments[0].Ownership.ResultProducerClauseID = 2
			test.mutate(&ability)
			bindObservedCardResultOwnership(ability.Sentences, ability.ConditionSegments, ability.ConditionClauses, ability.SemanticReferences)
			if ability.ConditionSegments[0].Ownership.ResultProducerClauseID != 0 {
				t.Fatal("unowned observation acquired a result receipt")
			}
		})
	}
}

func TestObservedCardResultModeOwnershipStaysLocal(t *testing.T) {
	const body = "You may look at the top card of target opponent's library. You gain 1 life. If you do, you may put that card on the bottom of that player's library."
	document, diagnostics := Parse("Choose two \u2014\n\u2022 "+body+"\n\u2022 "+body, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	modes := document.Abilities[0].Modal.Options
	for _, mode := range modes {
		if mode.ConditionSegments[0].Ownership.ResultProducerClauseID != 1 {
			t.Fatal("mode outcome borrowed another body or its Gain rider")
		}
	}
	first := modes[0]
	first.ConditionSegments[0].Ownership.ResultProducerClauseID = 2
	first.SemanticReferences = slices.Clone(modes[1].SemanticReferences)
	bindObservedCardResultOwnership(first.Sentences, first.ConditionSegments, first.ConditionClauses, first.SemanticReferences)
	if first.ConditionSegments[0].Ownership.ResultProducerClauseID != 0 {
		t.Fatal("foreign mode reference acquired a canonical observation outcome")
	}
}
