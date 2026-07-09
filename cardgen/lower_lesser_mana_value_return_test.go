package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func orahSkyclaveHierophantCard() *ScryfallCard {
	power, toughness := "3", "3"
	return &ScryfallCard{
		Name:     "Orah, Skyclave Hierophant",
		Layout:   "normal",
		ManaCost: "{2}{W}{B}",
		TypeLine: "Legendary Creature — Kor Cleric",
		OracleText: "Lifelink\n" +
			"Whenever Orah, Skyclave Hierophant or another Cleric you control dies, return target Cleric card with lesser mana value from your graveyard to the battlefield.",
		Power:     &power,
		Toughness: &toughness,
	}
}

// TestLowerOrahReturnCarriesEventRelativeManaValueBound proves the self-or-subtype
// dies trigger lowers to a graveyard-return target whose Selection carries the
// event-relative ManaValueLessThanEventPermanent bound alongside the Cleric
// subtype filter and graveyard zone, returning the target to the battlefield.
func TestLowerOrahReturnCarriesEventRelativeManaValueBound(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, orahSkyclaveHierophantCard())
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventPermanentDied ||
		!trigger.Trigger.Pattern.SubjectSelectionOrSelf {
		t.Fatalf("trigger pattern = %#v, want self-or-subject dies", trigger.Trigger.Pattern)
	}
	mode := trigger.Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	target := mode.Targets[0]
	sel := target.Selection.Val
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard {
		t.Fatalf("target zone/allow = %#v, want graveyard card", target)
	}
	if !sel.ManaValueLessThanEventPermanent {
		t.Fatal("target selection must carry ManaValueLessThanEventPermanent")
	}
	if !slices.Equal(sel.SubtypesAny, []types.Sub{types.Sub("Cleric")}) {
		t.Fatalf("subtypes = %v, want [Cleric]", sel.SubtypesAny)
	}
	if sel.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", sel.Controller)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.PutOnBattlefield); !ok {
		t.Fatalf("primitive = %#v, want PutOnBattlefield", mode.Sequence[0].Primitive)
	}
}

// TestGenerateOrahRendersEventRelativeManaValueBound proves the generated source
// renders the ManaValueLessThanEventPermanent field, exercising the compiler
// field copy and the render_target projection end to end.
func TestGenerateOrahRendersEventRelativeManaValueBound(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(orahSkyclaveHierophantCard(), "o")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "ManaValueLessThanEventPermanent: true") {
		t.Fatalf("generated source missing ManaValueLessThanEventPermanent:\n%s", source)
	}
}

// TestLowerEqualOrLesserManaValueReturnFailsClosed proves the ≤ wording "with
// equal or lesser mana value", which the strict event-relative bound must not
// absorb, stays unsupported rather than silently lowering to the strict bound.
func TestLowerEqualOrLesserManaValueReturnFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Equal Or Lesser Return",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target creature card with equal or lesser mana value from your graveyard to the battlefield.",
	})
}
