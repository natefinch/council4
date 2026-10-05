package parser

import "testing"

func TestParseSingleLibraryCardProducer(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text string
		kind EffectKind
	}{
		{"Look at the top card of your library.", EffectLookAtLibraryTop},
		{"You may look at the top card of your library.", EffectLookAtLibraryTop},
		{"Look at the top card of target opponent's library.", EffectLookAtLibraryTop},
		{"Reveal the top card of your library.", EffectReveal},
		{"You may reveal the top card of your library.", EffectReveal},
		{"Reveal the top card of target player's library.", EffectReveal},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tt.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 || len(document.Abilities) != 1 {
				t.Fatalf("diagnostics=%#v abilities=%d", diagnostics, len(document.Abilities))
			}
			effect := document.Abilities[0].Sentences[0].Effects[0]
			if effect.Kind != tt.kind || !effect.Exact || effect.CardSource != EffectCardSourceTopOfPlayerLibrary ||
				!effect.Amount.Known || effect.Amount.Value != 1 {
				t.Fatalf("producer kind=%v exact=%v source=%v amount=%#v", effect.Kind, effect.Exact, effect.CardSource, effect.Amount)
			}
		})
	}
}

func TestParseSingleLibraryCardNearMisses(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Reveal the top two cards of your library.",
		"Reveal the top card of your library face down.",
		"Reveal the top card of each player's library.",
		"Look at the top card of your library any time.",
		"Look at the bottom card of your library.",
		"Reveal cards from the top of your library until you reveal a land card.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(text, Context{InstantOrSorcery: true})
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					for _, effect := range sentence.Effects {
						if effect.CardSource == EffectCardSourceTopOfPlayerLibrary && effect.Exact &&
							(effect.Kind == EffectReveal || effect.Kind == EffectLookAtLibraryTop) {
							t.Fatalf("near miss acquired a single-card producer: %#v", effect)
						}
					}
				}
			}
		})
	}
}

func TestLibraryCardReferencesOwnExactProducerClause(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Reveal the top card of your library. Draw a card. If it's a land card, put it onto the battlefield tapped. Reveal the top card of your library. Scry 1. If it's a creature card, put the revealed card into your hand.",
		"Look at the top card of your library. Scry 1. If it's a land card, put it into your hand. Look at the top card of your library. Draw a card. If it's a creature card, put the looked-at card into your graveyard.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}

			ability := document.Abilities[0]
			want := []int{1, 1, 4, 4}
			if len(ability.SemanticReferences) != len(want) {
				t.Fatalf("references=%#v", ability.SemanticReferences)
			}
			for i, reference := range ability.SemanticReferences {
				if reference.ProducerClauseID != want[i] {
					t.Errorf("reference[%d] %q producer=%d, want %d", reference.NodeID, reference.Text, reference.ProducerClauseID, want[i])
				}
			}
			for _, index := range []int{2, 5} {
				effect := ability.Sentences[index].Effects[0]
				if !effect.Exact || effect.CardSource != EffectCardSourcePriorInstructionResult {
					t.Errorf("library-card consumer clause %d: exact=%v source=%v refs=%#v", effect.ClauseID, effect.Exact, effect.CardSource, effect.References)
				}
			}
		})
	}
}

func TestLibraryCardIndependentConditionsDoNotAcquireCreatedToken(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Reveal the top card of your library. If it's a creature card, create a 1/1 green Saproling creature token. If it's a land card, put it onto the battlefield under your control.",
		Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	references := document.Abilities[0].SemanticReferences
	if len(references) != 3 {
		t.Fatalf("references=%#v", references)
	}
	for _, reference := range references {
		if reference.ProducerClauseID != 1 {
			t.Errorf("reference %d %q acquired producer %d rather than the observed card", reference.NodeID, reference.Text, reference.ProducerClauseID)
		}
	}
}

func TestLibraryCardConditionSubjectUsesSharedSelectionGrammar(t *testing.T) {
	t.Parallel()
	for _, noun := range []string{
		"land card",
		"snow card",
		"legendary card",
		"nonland card",
		"noncreature nonland card",
		"creature or land card",
		"colorless card",
	} {
		t.Run(noun, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse("Look at the top card of your library. If it's a "+noun+", you gain 2 life.",
				Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			references := document.Abilities[0].SemanticReferences
			if len(references) != 1 || references[0].Kind != ReferencePronoun ||
				references[0].Pronoun != PronounIt || references[0].NodeID < 0 ||
				references[0].ProducerClauseID != 1 {
				t.Fatalf("typed observed-card condition subject = %#v", references)
			}
		})
	}
	for _, noun := range []string{
		"card", "beautiful card", "face-down card", "card named Island",
		"snow beautiful card", "snow card with mana value 2", "snow and legendary card",
		"colorless beautiful card", "colorless card and creature", "blue and green card",
	} {
		t.Run("unsupported/"+noun, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse("Look at the top card of your library. If it's a "+noun+", you gain 2 life.",
				Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			if references := document.Abilities[0].SemanticReferences; len(references) != 0 {
				t.Fatalf("unsupported card qualifier acquired an observation: %#v", references)
			}
		})
	}
}

func TestLibraryCardOwnershipRejectsUnmodeledOrCompetingProducers(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Reveal the top two cards of your library. If it's a land card, put it into your hand.",
		"Reveal the top card of each player's library. If it's a land card, put it into your hand.",
		"Look at the top card of your library face down. If it's a land card, put it into your hand.",
		"Reveal the top card of your library. Exile target creature. If it's a creature card, put it into your hand.",
		"Look at the top card of your library. Create a 1/1 green Saproling creature token. Put it into your hand.",
		"Reveal the top card of your library. This creature gains flying until end of turn. If it's a creature card, put it into your hand.",
		"Reveal the top card of your library. That creature gains flying until end of turn.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			for _, reference := range document.Abilities[0].SemanticReferences {
				if reference.ProducerClauseID != 0 {
					t.Errorf("competing or unmodeled subject acquired library producer: %#v", reference)
				}
			}
		})
	}
}

func TestLibraryCardOwnershipDoesNotHijackNewBattlefieldSubject(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		text string
		want []int
	}{
		{"Look at the top card of your library. Put it onto the battlefield. It gains haste until end of turn.", []int{1, 0}},
		{"Reveal the top card of your library. If it's a creature card, put it onto the battlefield. It gains haste until end of turn.", []int{1, 1, 0}},
		{"Reveal the top card of your library. If it's a creature card, put it onto the battlefield, and it gains haste until end of turn.", []int{1, 1, 0}},
		{"Reveal the top card of your library. If it's a land card, put it onto the battlefield. Otherwise, put it into your hand.", []int{1, 1, 1}},
		{"Reveal the top card of your library. If it's a land card, put it onto the battlefield. If it's a snow card, you gain 2 life.", []int{1, 1, 1}},
		{"Look at the top card of your library. Put it onto the battlefield. Put it into your hand.", []int{1, 0}},
	} {
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			references := document.Abilities[0].SemanticReferences
			if len(references) != len(tc.want) {
				t.Fatalf("references=%#v, want producer clauses %v", references, tc.want)
			}
			for i, reference := range references {
				if reference.ProducerClauseID != tc.want[i] {
					t.Errorf("reference %d %q has producer %d, want %d", reference.NodeID, reference.Text, reference.ProducerClauseID, tc.want[i])
				}
			}
		})
	}
}
