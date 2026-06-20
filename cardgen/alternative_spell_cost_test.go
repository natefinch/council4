package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

const commanderAlternativeCostText = "If you control a commander, you may cast this spell without paying its mana cost."

func TestLowerFierceGuardianship(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Fierce Guardianship",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{U}",
		OracleText: commanderAlternativeCostText + "\nCounter target noncreature spell.",
	})
	if len(face.AlternativeCosts) != 1 ||
		face.AlternativeCosts[0].Condition != cost.AlternativeConditionControlsCommander ||
		face.AlternativeCosts[0].ManaCost.Exists {
		t.Fatalf("alternative costs = %#v", face.AlternativeCosts)
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 ||
		!slices.Equal(mode.Targets[0].Predicate.ExcludedSpellCardTypes, []types.Card{types.Creature}) {
		t.Fatalf("targets = %#v, want target noncreature spell", mode.Targets)
	}
	counter, ok := mode.Sequence[0].Primitive.(game.CounterObject)
	if !ok || counter.Object != game.TargetStackObjectReference(0) {
		t.Fatalf("primitive = %#v, want counter target stack object", mode.Sequence[0].Primitive)
	}
}

func TestLowerCommanderAlternativeCostIsTextBlind(t *testing.T) {
	t.Parallel()
	lowered, diagnostic := lowerSpellAlternativeCost(compiler.CompiledAbility{
		Kind: compiler.AbilitySpellAlternativeCost,
		Text: "not Oracle wording",
		AlternativeCost: &compiler.CompiledAlternativeCost{
			Condition:             compiler.AlternativeCostConditionControlsCommander,
			WithoutPayingManaCost: true,
		},
	})
	if diagnostic != nil {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	if len(lowered.alternativeCosts) != 1 ||
		lowered.alternativeCosts[0].Condition != cost.AlternativeConditionControlsCommander {
		t.Fatalf("alternative costs = %#v", lowered.alternativeCosts)
	}
}

func TestGenerateFierceGuardianshipSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Fierce Guardianship",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{U}",
		OracleText: commanderAlternativeCostText + "\nCounter target noncreature spell.",
	}, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"AlternativeCosts: []cost.Alternative{",
		"Condition: cost.AlternativeConditionControlsCommander",
		"ExcludedSpellCardTypes: []types.Card{types.Creature}",
		"Primitive: game.CounterObject",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestCommanderAlternativeCostSiblings(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "Deadly Rollick", body: "Exile target creature."},
		{name: "Flawless Maneuver", body: "Creatures you control gain indestructible until end of turn."},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				TypeLine:   "Instant",
				ManaCost:   "{3}{B}",
				OracleText: commanderAlternativeCostText + "\n" + test.body,
			})
			if len(face.AlternativeCosts) != 1 {
				t.Fatalf("alternative costs = %#v", face.AlternativeCosts)
			}
		})
	}
}

func TestCommanderAlternativeCostDoesNotHideUnsupportedBody(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Deflecting Swat",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{R}",
		OracleText: commanderAlternativeCostText + "\nYou may choose new targets for target spell or ability.",
	})
	if !face.empty() {
		t.Fatalf("partially lowered unsupported card: %#v", face)
	}
}
