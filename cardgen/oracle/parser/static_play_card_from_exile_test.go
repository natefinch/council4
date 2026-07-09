package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// parseStaticPlayCardFromExileDeclaration parses a single static ability and
// returns the sole StaticDeclaration the parser emitted. It fails the test when
// the source did not produce exactly one ability carrying one static declaration.
func parseStaticPlayCardFromExileDeclaration(t *testing.T, source string) StaticDeclarationSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].StaticDeclarations) != 1 {
		t.Fatalf("abilities = %#v, want one ability with one static declaration", document.Abilities)
	}
	return document.Abilities[0].StaticDeclarations[0]
}

// TestParsePlayCardFromExileWithCounterRiders covers the play-a-card-from-exile
// recognizer and its three independent riders. The counter name is read
// text-blind, and the once-per-turn, provenance, and any-color riders each map to
// their own typed flag so wordings carrying any subset lower correctly.
func TestParsePlayCardFromExileWithCounterRiders(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source        string
		counter       counter.Kind
		oncePerTurn   bool
		exiledByYou   bool
		spendAnyColor bool
	}{
		"evelyn full wording": {
			source:        "Once each turn, you may play a card from exile with a collection counter on it if it was exiled by an ability you controlled, and you may spend mana as though it were mana of any color to cast it.",
			counter:       counter.Collection,
			oncePerTurn:   true,
			exiledByYou:   true,
			spendAnyColor: true,
		},
		"core permission with no riders": {
			source:  "You may play a card from exile with a collection counter on it.",
			counter: counter.Collection,
		},
		"once per turn only": {
			source:      "Once each turn, you may play a card from exile with a collection counter on it.",
			counter:     counter.Collection,
			oncePerTurn: true,
		},
		"provenance and any-color without once per turn": {
			source:        "You may play a card from exile with a collection counter on it if it was exiled by an ability you controlled, and you may spend mana as though it were mana of any color to cast it.",
			counter:       counter.Collection,
			exiledByYou:   true,
			spendAnyColor: true,
		},
		"different named counter": {
			source:      "Once each turn, you may play a card from exile with a croak counter on it.",
			counter:     counter.Croak,
			oncePerTurn: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declaration := parseStaticPlayCardFromExileDeclaration(t, tc.source)
			if declaration.PlayerRule != StaticDeclarationPlayerRulePlayAndCastFromExileWithCounter {
				t.Fatalf("PlayerRule = %v, want play-and-cast-from-exile-with-counter", declaration.PlayerRule)
			}
			if declaration.ExileCounter != tc.counter {
				t.Fatalf("ExileCounter = %v, want %v", declaration.ExileCounter, tc.counter)
			}
			if declaration.ExilePlayOncePerTurn != tc.oncePerTurn {
				t.Fatalf("ExilePlayOncePerTurn = %v, want %v", declaration.ExilePlayOncePerTurn, tc.oncePerTurn)
			}
			if declaration.ExilePlayExiledByControlledAbility != tc.exiledByYou {
				t.Fatalf("ExilePlayExiledByControlledAbility = %v, want %v", declaration.ExilePlayExiledByControlledAbility, tc.exiledByYou)
			}
			if declaration.ExilePlaySpendAnyColorMana != tc.spendAnyColor {
				t.Fatalf("ExilePlaySpendAnyColorMana = %v, want %v", declaration.ExilePlaySpendAnyColorMana, tc.spendAnyColor)
			}
		})
	}
}

// TestParsePlayCardFromExileWithCounterFailsClosed verifies the recognizer
// fail-closes on wordings that deviate from the play-a-card-from-exile template
// rather than dropping a rider it cannot represent, so a card is never
// mis-generated. Each source produces no static declaration.
func TestParsePlayCardFromExileWithCounterFailsClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"missing counter noun": "You may play a card from exile with a collection on it.",
		"no counter named":     "You may play a card from exile with a counter on it.",
		"wrong verb":           "You may exile a card from exile with a collection counter on it.",
		"missing on it close":  "You may play a card from exile with a collection counter.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{})
			for _, ability := range document.Abilities {
				for _, declaration := range ability.StaticDeclarations {
					if declaration.PlayerRule == StaticDeclarationPlayerRulePlayAndCastFromExileWithCounter {
						t.Fatalf("source %q was wrongly recognized as a play-a-card-from-exile permission", source)
					}
				}
			}
		})
	}
}
