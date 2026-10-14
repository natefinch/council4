package parser

import "testing"

func declaredPermanentBody(t *testing.T, text string) Ability {
	t.Helper()
	document, diagnostics := Parse(text, Context{CardName: "Declared Probe", InstantOrSorcery: true})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("parse: %v", diagnostics)
	}
	return document.Abilities[0]
}

func TestDeclaredPermanentGrammar(t *testing.T) {
	for _, action := range []string{
		"Return target creature to its owner's hand.",
		"Destroy target artifact.",
		"Tap target creature.",
		"Untap target land.",
	} {
		for _, prefix := range []string{"", "Exile target card from a graveyard. ", "Target player gains 1 life. ", "Counter target spell. "} {
			for _, rider := range []string{"", "You gain 1 life. "} {
				t.Run(prefix+action+rider, func(t *testing.T) {
					ability := declaredPermanentBody(t, prefix+action+" "+rider+"Its controller loses 2 life. Its owner gains 3 life.")
					for _, reference := range ability.SemanticReferences[len(ability.SemanticReferences)-2:] {
						if reference.DeclaredPermanent == nil || reference.DeclaredPermanent.ClauseID == 0 ||
							reference.ProducerClauseID != 0 {
							t.Fatalf("lost declared-original identity: %+v", reference)
						}
						copies := 0
						for _, sentence := range ability.Sentences {
							for _, effect := range sentence.Effects {
								for _, owned := range effect.References {
									if owned.NodeID == reference.NodeID {
										copies++
										if owned.DeclaredPermanent == nil || *owned.DeclaredPermanent != *reference.DeclaredPermanent {
											t.Fatal("authoritative identity not distributed to consumer")
										}
									}
								}
							}
						}
						if copies != 1 {
							t.Fatalf("consumer copies = %d", copies)
						}
					}
				})
			}
		}
	}
}

func TestDeclaredPermanentGrammarRefusesAmbiguity(t *testing.T) {
	for _, action := range []string{
		"Destroy target creature. Tap target artifact.",
		"Tap two target creatures.",
		"Tap up to one target creature.",
		"Tap any number of target creatures.",
	} {
		t.Run(action, func(t *testing.T) {
			ability := declaredPermanentBody(t, "Exile target card from a graveyard. "+action+" Its owner gains 3 life.")
			reference := ability.SemanticReferences[len(ability.SemanticReferences)-1]
			if reference.DeclaredPermanent == nil || reference.DeclaredPermanent.ClauseID != 0 {
				t.Fatalf("ambiguous declaration not explicitly refused: %+v", reference)
			}
		})
	}
}

func TestDeclaredPermanentGrammarLeavesOtherSubjectsAlone(t *testing.T) {
	for _, text := range []string{
		"Tap target creature. This creature gets +1/+1 until end of turn. Its controller gains 3 life.",
		"Tap target creature. Exile target card from a graveyard. Its owner gains 3 life.",
		"Tap target creature. Counter target spell. Its controller gains 3 life.",
		"Tap target creature. Target player gains 1 life. They draw a card.",
		"Return target creature card from your graveyard to the battlefield. Its controller gains 3 life.",
		"Reveal the top card of your library. Its owner gains 3 life.",
		"Exile target creature. Its controller gains life equal to its power.",
		"Tap target creature. That card's owner gains 3 life.",
		"Exile target card from a graveyard. Tap target creature. Bolster 1. That creature's owner gains 3 life.",
		"Exile target card from a graveyard. Tap target creature. Bolster 1. Its owner gains 3 life.",
		"Exile target card from a graveyard. Tap target creature. Manifest the top card of your library. Its owner gains 3 life.",
		"Exile target card from a graveyard. Tap target creature. Cloak the top card of your library. Its owner gains 3 life.",
		"Exile target card from a graveyard. Tap target creature. Populate. Its owner gains 3 life.",
	} {
		t.Run(text, func(t *testing.T) {
			ability := declaredPermanentBody(t, text)
			for _, reference := range ability.SemanticReferences {
				if reference.DeclaredPermanent != nil {
					t.Fatalf("other subject domain acquired declared-original contract: %+v", reference)
				}
			}
		})
	}
}
