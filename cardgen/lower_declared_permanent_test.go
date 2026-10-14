package cardgen

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestDeclaredPermanentReadbackCardDef(t *testing.T) {
	for _, action := range []string{
		"Return target creature to its owner's hand.",
		"Destroy target creature.",
		"Tap target creature.",
	} {
		for _, rider := range []string{"", "You gain 1 life. "} {
			t.Run(action+rider, func(t *testing.T) {
				prefix := "Exile target card from a graveyard. "
				before := 1
				if action == "Return target creature to its owner's hand." {
					// Retain the independent controller-rider composition control.
					prefix += "You gain 1 life. "
					before++
				}
				card := &ScryfallCard{Name: "Declared Target Probe", Layout: "normal", TypeLine: "Sorcery",
					OracleText: prefix + action + " " + rider +
						"Its controller loses 2 life. Its owner gains 3 life."}
				index := before + 1
				if rider != "" {
					index++
				}
				assertCardPaths(t, card,
					fmt.Sprintf("Sequence[%d].Primitive.(game.LoseLife).Player.kind = game.PlayerReferenceObjectController", index),
					fmt.Sprintf("Sequence[%d].Primitive.(game.LoseLife).Player.object.Val.kind = game.ObjectReferenceTargetPermanent", index),
					fmt.Sprintf("Sequence[%d].Primitive.(game.LoseLife).Player.object.Val.targetIndex = 1", index),
					fmt.Sprintf("Sequence[%d].Primitive.(game.GainLife).Player.kind = game.PlayerReferenceObjectOwner", index+1),
					fmt.Sprintf("Sequence[%d].Primitive.(game.GainLife).Player.object.Val.kind = game.ObjectReferenceTargetPermanent", index+1),
					fmt.Sprintf("Sequence[%d].Primitive.(game.GainLife).Player.object.Val.targetIndex = 1", index+1),
				)
			})
		}
	}
}

func TestDeclaredPermanentSameClauseRawComposition(t *testing.T) {
	for _, noun := range []string{"creature", "creature or Vehicle"} {
		t.Run(noun, func(t *testing.T) {
			card := &ScryfallCard{Name: "Raw Declared Composition", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{G}",
				OracleText: "Exile target card from a graveyard. Return target " + noun +
					" to its owner's hand. Its controller loses 2 life. Its owner gains 3 life."}
			defs, diagnostics, err := CompileCardDefs(card)
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("raw compilation: err=%v diagnostics=%v definitions=%d", err, diagnostics, len(defs))
			}
			if problems := game.ValidateCardDef(defs[0]); len(problems) != 0 {
				t.Fatalf("whole CardDef invalid: %v", problems)
			}
			if !defs[0].SpellAbility.Exists || len(defs[0].SpellAbility.Val.Modes) != 1 {
				t.Fatal("raw composition lost its single complete spell mode")
			}
			mode := defs[0].SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 2 || len(mode.Sequence) != 4 {
				t.Fatalf("targets=%d instructions=%d, want 2/4 without isolation riders", len(mode.Targets), len(mode.Sequence))
			}
			exile := captureTestPrimitive[game.MoveCard](t, mode.Sequence[0].Primitive)
			bounce := captureTestPrimitive[game.MovePermanent](t, mode.Sequence[1].Primitive)
			lose := captureTestPrimitive[game.LoseLife](t, mode.Sequence[2].Primitive)
			gain := captureTestPrimitive[game.GainLife](t, mode.Sequence[3].Primitive)
			if exile.Card.Kind != game.CardReferenceTarget || exile.Card.TargetIndex != 0 ||
				exile.FromZone != zone.Graveyard || exile.Destination != zone.Exile ||
				bounce.Object != game.TargetPermanentReference(1) || bounce.Destination != zone.Hand ||
				lose.Player != game.ObjectControllerReference(game.TargetPermanentReference(1)) || lose.Amount.Value() != 2 ||
				gain.Player != game.ObjectOwnerReference(game.TargetPermanentReference(1)) || gain.Amount.Value() != 3 {
				t.Fatal("raw primitives borrowed the independent card, source, or wrong permanent occurrence")
			}
			if noun == "creature or Vehicle" {
				assertCardPaths(t, card,
					"Modes[0].Targets[1].Selection.Val.AnyOf[0].RequiredTypesAny[0] = types.Creature",
					"Modes[0].Targets[1].Selection.Val.AnyOf[1].SubtypesAny[0] = types.Vehicle")
			}
		})
	}
}

func TestDeclaredPermanentReadbackRefusesAmbiguity(t *testing.T) {
	for _, text := range []string{
		"Exile target card from a graveyard. Destroy target creature. Tap target artifact. Its owner gains 3 life.",
		"Exile target card from a graveyard. Tap two target creatures. Its owner gains 3 life.",
		"Exile target card from a graveyard. Tap up to one target creature. Its owner gains 3 life.",
		"Exile target card from a graveyard. Tap any number of target creatures. Its owner gains 3 life.",
		"Exile target card from a graveyard. Tap target creature. Their owner gains 3 life.",
		"Exile target card from a graveyard. Tap target creature. Bolster 1. That creature's owner gains 3 life.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{Name: "Ambiguous Declared Probe", Layout: "normal", TypeLine: "Sorcery", OracleText: text})
		})
	}
}

func TestDeclaredPermanentExplicitPossessiveAndPrefixDomains(t *testing.T) {
	for _, prefix := range []string{"Exile target card from a graveyard. ", "Target player gains 1 life. ", "Counter target spell. "} {
		for _, subject := range []string{"Its", "That creature's"} {
			t.Run(prefix+subject, func(t *testing.T) {
				card := &ScryfallCard{Name: "Explicit Declared Probe", Layout: "normal", TypeLine: "Sorcery",
					OracleText: prefix + "Tap target creature. " + subject + " owner gains 3 life."}
				assertCardPaths(t, card,
					"Sequence[2].Primitive.(game.GainLife).Player.kind = game.PlayerReferenceObjectOwner",
					"Sequence[2].Primitive.(game.GainLife).Player.object.Val.kind = game.ObjectReferenceTargetPermanent",
					"Sequence[2].Primitive.(game.GainLife).Player.object.Val.targetIndex = 1")
			})
		}
	}
}
