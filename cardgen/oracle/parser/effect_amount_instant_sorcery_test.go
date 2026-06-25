package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

// instantSorceryCountSelection parses a single-effect ability and returns the
// dynamic count selection of its first effect, used by the "instant and sorcery
// card[s] in your <zone>" count-subject tests.
func instantSorceryCountSelection(t *testing.T, source string, context Context) *SelectionSyntax {
	t.Helper()
	document, _ := Parse(source, context)
	for ai := range document.Abilities {
		ability := &document.Abilities[ai]
		for si := range ability.Sentences {
			sentence := &ability.Sentences[si]
			for ei := range sentence.Effects {
				if selection := sentence.Effects[ei].Amount.Selection; selection != nil {
					return selection
				}
			}
		}
	}
	return nil
}

func TestParseInstantSorceryGraveyardCountSubject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		zone   zone.Type
	}{
		{
			name:   "for each in your graveyard",
			source: "This spell costs {1} less to cast for each instant and sorcery card in your graveyard.",
			zone:   zone.Graveyard,
		},
		{
			name:   "number of in your graveyard",
			source: "Draw cards equal to the number of instant and sorcery cards in your graveyard.",
			zone:   zone.Graveyard,
		},
		{
			name:   "for each in your hand",
			source: "This spell costs {1} less to cast for each instant and sorcery card in your hand.",
			zone:   zone.Hand,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			selection := instantSorceryCountSelection(t, test.source, Context{InstantOrSorcery: true})
			if selection == nil {
				t.Fatalf("source %q yielded no dynamic count selection", test.source)
			}
			want := []CardType{CardTypeInstant, CardTypeSorcery}
			got := slices.Clone(selection.RequiredTypesAny)
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Fatalf("RequiredTypesAny = %v, want %v", selection.RequiredTypesAny, want)
			}
			if selection.ConjunctiveTypes {
				t.Fatal("instant and sorcery card union must stay disjunctive, got ConjunctiveTypes=true")
			}
			if selection.Zone != test.zone {
				t.Fatalf("Zone = %v, want %v", selection.Zone, test.zone)
			}
			if selection.Controller != SelectionControllerYou {
				t.Fatalf("Controller = %v, want SelectionControllerYou", selection.Controller)
			}
		})
	}
}

func TestParseInstantSorceryCountSubjectFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "library zone",
			source: "This spell costs {1} less to cast for each instant and sorcery card in your library.",
		},
		{
			name:   "non spell-type pair",
			source: "This spell costs {1} less to cast for each creature and artifact card in your graveyard.",
		},
		{
			name:   "single instant type",
			source: "This spell costs {1} less to cast for each instant card in your library.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if effect := sourceSpellReductionEffect(t, test.source, Context{InstantOrSorcery: true}); effect != nil {
				t.Fatalf("source %q was unexpectedly recognized as a cost reduction", test.source)
			}
		})
	}
}
