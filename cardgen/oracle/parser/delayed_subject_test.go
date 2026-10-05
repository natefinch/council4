package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestDelayedSubjectOwnership(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		text       string
		kind       DelayedSubjectKind
		occurrence int
		cardZone   zone.Type
	}{
		{"Whenever this creature blocks or becomes blocked by a creature, destroy that creature at end of combat.", DelayedSubjectEventRelated, 0, zone.None},
		{"Whenever this creature deals combat damage to a creature, destroy that creature at end of combat.", DelayedSubjectEvent, 0, zone.None},
		{"Create two 1/1 green Insect creature tokens. You gain 2 life. Exile them at end of combat.", DelayedSubjectProduct, 0, zone.None},
		{"Tap target creature. Untap target creature. Draw a card. Exile it at end of combat.", DelayedSubjectTarget, 1, zone.None},
		{"Tap target creature. Untap target land. Sacrifice that creature at end of combat.", DelayedSubjectTarget, 0, zone.None},
		{"Destroy target blocking creature at end of combat.", DelayedSubjectTarget, 0, zone.None},
		{"When this creature dies, return it to its owner's hand at the beginning of the next end step.", DelayedSubjectSource, 0, zone.Graveyard},
		{"Whenever another creature you control dies, return that card to its owner's hand at the beginning of the next end step.", DelayedSubjectEvent, 0, zone.Graveyard},
		{"Whenever you cast a creature spell, exile it at end of combat.", DelayedSubjectUnsupported, 0, zone.None},
		{"Whenever a creature enters, sacrifice that land at end of combat.", DelayedSubjectUnsupported, 0, zone.None},
	} {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.text, Context{})
			found := false
			for _, sentence := range document.Abilities[0].Sentences {
				for _, effect := range sentence.Effects {
					if effect.DelayedTiming == DelayedTimingNone {
						continue
					}
					found = true
					subject := effect.DelayedSubject
					if subject.Kind != test.kind || subject.TargetOccurrence != test.occurrence || subject.CardZone != test.cardZone {
						t.Fatalf("ownership = %+v, want kind=%v occurrence=%d zone=%v", subject, test.kind, test.occurrence, test.cardZone)
					}
					if test.kind != DelayedSubjectUnsupported && !effect.Exact {
						t.Fatal("modeled delayed action is not exact")
					}
					if test.kind == DelayedSubjectProduct && subject.ProducerClauseID != sentenceProducerClauseID(document.Abilities[0]) {
						t.Fatal("product ownership lost its exact producer clause")
					}
				}
			}
			if !found {
				t.Fatal("no modeled delayed effect")
			}
		})
	}
}

func sentenceProducerClauseID(ability Ability) int {
	for _, sentence := range ability.Sentences {
		for _, effect := range sentence.Effects {
			if effect.Kind == EffectCreate {
				return effect.ClauseID
			}
		}
	}
	return -1
}

func TestDelayedConditionEvaluationTiming(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		text   string
		future bool
	}{
		{"Tap target creature. If it's an Elf, destroy it at the beginning of the next end step.", false},
		{"Tap target creature. At the beginning of the next end step, destroy it if it's an Elf.", true},
		{"Tap target creature. At the beginning of the next end step, if it's an Elf, destroy it.", true},
	} {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.text, Context{InstantOrSorcery: true})
			segments := document.Abilities[0].ConditionSegments
			if len(segments) != 1 || segments[0].Ownership.DelayedBody != test.future {
				t.Fatalf("condition ownership = %+v, want future=%t", segments, test.future)
			}
		})
	}
}
