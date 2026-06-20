package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

const esperSentinelOracle = "Whenever an opponent casts their first noncreature spell each turn, draw a card unless that player pays {X}, where X is this creature's power."

func TestGenerateExecutableCardSourceEsperSentinel(t *testing.T) {
	t.Parallel()
	power, toughness := "1", "1"
	card := &ScryfallCard{
		Name:       "Esper Sentinel",
		Layout:     "normal",
		ManaCost:   "{W}",
		TypeLine:   "Artifact Creature — Human Soldier",
		OracleText: esperSentinelOracle,
		Power:      &power,
		Toughness:  &toughness,
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.TriggerControllerOpponent",
		"PlayerEventOrdinalThisTurn: 1",
		"CardSelection:              game.Selection{ExcludedTypes: []types.Card{types.Creature}}",
		"game.EventPlayerReference()",
		"DynamicGenericManaCost: opt.Val(game.DynamicAmount{",
		"Kind:       game.DynamicAmountObjectPower",
		"Object:     game.SourcePermanentReference()",
		"Succeeded: game.TriFalse",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}

	face := lowerSingleFace(t, card)
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Pattern.PlayerEventOrdinalThisTurn != 1 ||
		len(ability.Trigger.Pattern.CardSelection.ExcludedTypes) != 1 ||
		ability.Trigger.Pattern.CardSelection.ExcludedTypes[0] != types.Creature {
		t.Fatalf("trigger pattern = %#v", ability.Trigger.Pattern)
	}
	sequence := ability.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(sequence))
	}
	pay, ok := sequence[0].Primitive.(game.Pay)
	if !ok || !pay.Payment.DynamicGenericManaCost.Exists {
		t.Fatalf("payment = %#v, want dynamic generic payment", sequence[0].Primitive)
	}
	dynamic := pay.Payment.DynamicGenericManaCost.Val
	if dynamic.Kind != game.DynamicAmountObjectPower || dynamic.Object != game.SourcePermanentReference() {
		t.Fatalf("dynamic payment = %#v", dynamic)
	}
	if sequence[1].Optional {
		t.Fatal("draw is optional, want mandatory draw unless paid")
	}
}

func TestEsperSentinelFailClosedNearMisses(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Whenever an opponent casts their second noncreature spell each turn, draw a card unless that player pays {X}, where X is this creature's power.",
		"Whenever an opponent casts their first spell each turn, draw a card unless that player pays {X}, where X is this creature's power.",
		"Whenever an opponent copies their first noncreature spell each turn, draw a card unless that player pays {X}, where X is this creature's power.",
		"Whenever an opponent casts their first noncreature spell each turn, gain 1 life unless that player pays {X}, where X is this creature's power.",
		"Whenever an opponent casts their first noncreature spell each turn, draw a card unless that player pays {X}, where X is this creature's toughness.",
		"Whenever an opponent casts their first noncreature spell each turn, draw a card unless that player pays {X}, where X is that spell's mana value.",
	} {
		oracle := oracle
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			power, toughness := "1", "1"
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Near Miss",
				Layout:     "normal",
				TypeLine:   "Artifact Creature — Construct",
				OracleText: oracle,
				Power:      &power,
				Toughness:  &toughness,
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected unsupported diagnostic for %q", oracle)
			}
			if len(faces) != 0 && len(faces[0].TriggeredAbilities) != 0 {
				t.Fatalf("unexpected executable trigger for %q", oracle)
			}
		})
	}
}
