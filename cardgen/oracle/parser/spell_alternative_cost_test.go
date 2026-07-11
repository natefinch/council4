package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestParseCommanderControlledAlternativeSpellCost(t *testing.T) {
	t.Parallel()
	for _, determiner := range []string{"a", "your"} {
		source := "If you control " + determiner + " commander, you may cast this spell without paying its mana cost.\nCounter target noncreature spell."
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", determiner, diagnostics)
		}

		if len(document.Abilities) != 2 {
			t.Fatalf("%q abilities = %d, want 2", determiner, len(document.Abilities))
		}
		ability := document.Abilities[0]
		if ability.Kind != AbilitySpellAlternativeCost || ability.AlternativeCost == nil {
			t.Fatalf("%q ability = %#v, want typed alternative spell cost", determiner, ability)
		}
		if ability.AlternativeCost.Condition != SpellAlternativeCostConditionControlsCommander ||
			!ability.AlternativeCost.WithoutPayingManaCost {
			t.Fatalf("%q alternative cost = %#v", determiner, ability.AlternativeCost)
		}
		if len(ability.Sentences) != 0 || ability.Optional {
			t.Fatalf("%q alternative cost parsed as resolving content: %#v", determiner, ability)
		}
		if len(ability.ConditionClauses) != 0 {
			t.Fatalf("%q alternative cost emitted redundant condition clauses: %#v", determiner, ability.ConditionClauses)
		}
	}
}

func TestParseCommanderControlledCreatureExileIsComplete(t *testing.T) {
	t.Parallel()
	alternativeText := "If you control a commander, you may cast this spell without paying its mana cost."
	source := alternativeText + "\nExile target creature."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if coverage := DocumentCoverage(document); !coverage.Complete {
		t.Fatalf("coverage = %#v, want complete", coverage)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(document.Abilities))
	}
	alternative := document.Abilities[0].AlternativeCost
	if alternative == nil ||
		alternative.Kind != SpellAlternativeCostCommander ||
		alternative.Condition != SpellAlternativeCostConditionControlsCommander ||
		!alternative.WithoutPayingManaCost {
		t.Fatalf("alternative cost = %#v", alternative)
	}
	if got := source[alternative.Span.Start.Offset:alternative.Span.End.Offset]; got != alternativeText {
		t.Fatalf("alternative span text = %q, want %q", got, alternativeText)
	}
	sentences := document.Abilities[1].Sentences
	if len(sentences) != 1 || len(sentences[0].Effects) != 1 || len(sentences[0].Targets) != 1 {
		t.Fatalf("exile syntax = %#v", document.Abilities[1])
	}
	effect, target := sentences[0].Effects[0], sentences[0].Targets[0]
	if effect.Kind != EffectExile || !effect.Exact {
		t.Fatalf("effect = %#v, want exact exile", effect)
	}
	if !target.Exact ||
		target.Cardinality != (TargetCardinalitySyntax{Min: 1, Max: 1}) ||
		target.Selection.Kind != SelectionCreature {
		t.Fatalf("target = %#v, want exact one creature", target)
	}
}

func TestParseFlashbackAlternativeSpellCost(t *testing.T) {
	t.Parallel()
	source := "Return target creature card from your graveyard to the battlefield.\nFlashback—Sacrifice three creatures. (You may cast this card from your graveyard for its flashback cost. Then exile it.)"
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if coverage := DocumentCoverage(document); !coverage.Complete {
		t.Fatalf("coverage = %#v, want complete", coverage)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(document.Abilities))
	}
	ability := document.Abilities[1]
	if ability.Kind != AbilitySpellAlternativeCost || ability.AlternativeCost == nil {
		t.Fatalf("ability = %#v, want typed alternative spell cost", ability)
	}
	if ability.AlternativeCost.Kind != SpellAlternativeCostFlashback {
		t.Fatalf("alternative cost kind = %v, want flashback", ability.AlternativeCost.Kind)
	}
	if ability.AbilityWord != nil {
		t.Fatalf("flashback label parsed as a generic ability word: %#v", ability.AbilityWord)
	}
	if ability.CostSyntax == nil || len(ability.CostSyntax.Components) != 1 ||
		ability.CostSyntax.Components[0].Kind != CostComponentSacrifice ||
		ability.CostSyntax.Components[0].AmountValue != 3 {
		t.Fatalf("flashback cost syntax = %#v", ability.CostSyntax)
	}
	if len(ability.Sentences) != 0 {
		t.Fatalf("flashback cost parsed as resolving content: %#v", ability.Sentences)
	}
}

func TestParseOverloadAlternativeSpellCost(t *testing.T) {
	t.Parallel()
	source := `Destroy target artifact you don't control.
Overload {4}{R} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")`
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(document.Abilities))
	}
	ability := document.Abilities[1]
	if ability.Kind != AbilitySpellAlternativeCost || ability.AlternativeCost == nil {
		t.Fatalf("ability = %#v, want typed overload cost", ability)
	}
	if ability.AlternativeCost.Kind != SpellAlternativeCostOverload ||
		!ability.AlternativeCost.ReplaceTargetWithEach ||
		!slices.Equal(ability.AlternativeCost.ManaCost, cost.Mana{cost.O(4), cost.R}) {
		t.Fatalf("alternative cost = %#v", ability.AlternativeCost)
	}
	if len(ability.Sentences) != 0 {
		t.Fatalf("overload parsed as resolving content: %#v", ability)
	}
}

func TestParseOverloadFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`Overload {4}{R} (Change "target" to "each.")`,
		`Overload {4}{R} (You may cast this spell for its overload cost. If you do, change "target" in its text to "all.")`,
		`Overload {4}{R}.`,
		`Overload`,
		`Overload {4}{R} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.") extra`,
	} {
		document, _ := Parse(source, Context{InstantOrSorcery: true})
		if len(document.Abilities) != 1 {
			t.Fatalf("%q abilities = %d, want 1", source, len(document.Abilities))
		}
		if document.Abilities[0].AlternativeCost != nil {
			t.Fatalf("%q unexpectedly recognized as overload", source)
		}
	}
}

func TestParseCommanderControlledAlternativeSpellCostFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"If you own a commander, you may cast this spell without paying its mana cost.",
		"If you control a commander, cast this spell without paying its mana cost.",
		"If you control a commander, you may cast this spell without paying its mana cost from exile.",
		"If you control a commander, you may cast this spell without paying its mana cost",
		"If you control a commander. You may cast this spell without paying its mana cost.",
		"If you control a commander; you may cast this spell without paying its mana cost.",
		"If you control a commander you may cast this spell without paying its mana cost.",
		"If you control a commander,, you may cast this spell without paying its mana cost.",
		"If you control a commander, you may cast this spell without paying its mana cost..",
		"If you control a commander, you may cast this spell; without paying its mana cost.",
	} {
		document, _ := Parse(source, Context{InstantOrSorcery: true})
		if len(document.Abilities) != 1 {
			t.Fatalf("%q abilities = %d, want 1", source, len(document.Abilities))
		}
		if document.Abilities[0].AlternativeCost != nil {
			t.Fatalf("%q unexpectedly recognized as alternative cost", source)
		}
		if coverage := DocumentCoverage(document); coverage.Complete {
			t.Fatalf("%q coverage unexpectedly complete", source)
		}
	}
}

func TestParsePitchAlternativeSpellCost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		source        string
		wantColor     Color
		wantCount     int
		wantLife      int
		wantCondition SpellAlternativeCostCondition
	}{
		{
			name:      "force of will",
			source:    "You may pay 1 life and exile a blue card from your hand rather than pay this spell's mana cost.\nCounter target spell.",
			wantColor: ColorBlue,
			wantCount: 1,
			wantLife:  1,
		},
		{
			name:          "force of negation",
			source:        "If it's not your turn, you may exile a blue card from your hand rather than pay this spell's mana cost.\nCounter target noncreature spell.",
			wantColor:     ColorBlue,
			wantCount:     1,
			wantCondition: SpellAlternativeCostConditionNotYourTurn,
		},
		{
			name:      "misdirection no extra cost",
			source:    "You may exile a blue card from your hand rather than pay this spell's mana cost.\nCounter target spell.",
			wantColor: ColorBlue,
			wantCount: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := document.Abilities[0]
			alternative := ability.AlternativeCost
			if alternative == nil || alternative.Kind != SpellAlternativeCostPitch {
				t.Fatalf("alternative cost = %#v, want pitch", alternative)
			}
			if alternative.Condition != test.wantCondition {
				t.Fatalf("pitch condition = %v, want %v", alternative.Condition, test.wantCondition)
			}
			if ability.CostSyntax == nil {
				t.Fatal("pitch cost syntax = nil, want components")
			}
			var gotLife int
			var exile *CostComponent
			for i := range ability.CostSyntax.Components {
				component := &ability.CostSyntax.Components[i]
				switch component.Kind {
				case CostComponentPayLife:
					gotLife = component.AmountValue
				case CostComponentExile:
					exile = component
				default:
				}
			}
			if gotLife != test.wantLife {
				t.Fatalf("pitch life = %d, want %d", gotLife, test.wantLife)
			}
			if exile == nil {
				t.Fatalf("pitch cost syntax = %#v, want exile component", ability.CostSyntax)
			}
			if exile.ObjectColor != test.wantColor ||
				!exile.ObjectColorKnown ||
				exile.AmountValue != test.wantCount ||
				exile.SourceZone != zone.Hand {
				t.Fatalf("pitch exile component = %#v", exile)
			}
		})
	}
}

func TestParsePitchAlternativeCostFailsClosed(t *testing.T) {
	t.Parallel()
	// "with mana value X" selectors are not modeled and must not parse as a
	// plain colored-card pitch.
	source := "You may exile a blue card with mana value X from your hand rather than pay this spell's mana cost.\nCounter target spell."
	document, _ := Parse(source, Context{InstantOrSorcery: true})
	if alternative := document.Abilities[0].AlternativeCost; alternative != nil &&
		alternative.Kind == SpellAlternativeCostPitch {
		t.Fatalf("unexpectedly parsed mana-value pitch as plain pitch: %#v", alternative)
	}
}

func TestParseFreeAlternativeSpellCost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		source        string
		wantCondition SpellAlternativeCostCondition
		wantSubtype   types.Sub
		wantComponent CostComponentKind
		wantAmount    int
	}{
		{
			name:          "snuff out pays life gated by a swamp",
			source:        "If you control a Swamp, you may pay 4 life rather than pay this spell's mana cost.\nDestroy target nonblack creature. It can't be regenerated.",
			wantCondition: SpellAlternativeCostConditionControlsSubtype,
			wantSubtype:   types.Swamp,
			wantComponent: CostComponentPayLife,
			wantAmount:    4,
		},
		{
			name:          "fireblast sacrifices two lands unconditionally",
			source:        "You may sacrifice two Mountains rather than pay this spell's mana cost.\nFireblast deals 4 damage to any target.",
			wantComponent: CostComponentSacrifice,
			wantAmount:    2,
		},
		{
			name:          "mine collapse gated by your turn",
			source:        "If it's your turn, you may sacrifice a Mountain rather than pay this spell's mana cost.\nMine Collapse deals 5 damage to target creature or planeswalker.",
			wantCondition: SpellAlternativeCostConditionYourTurn,
			wantComponent: CostComponentSacrifice,
			wantAmount:    1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := document.Abilities[0]
			alternative := ability.AlternativeCost
			if alternative == nil || alternative.Kind != SpellAlternativeCostFree {
				t.Fatalf("alternative cost = %#v, want free", alternative)
			}
			if alternative.Condition != test.wantCondition {
				t.Fatalf("free condition = %v, want %v", alternative.Condition, test.wantCondition)
			}
			if alternative.ConditionSubtype != test.wantSubtype {
				t.Fatalf("free condition subtype = %q, want %q", alternative.ConditionSubtype, test.wantSubtype)
			}
			if ability.CostSyntax == nil || len(ability.CostSyntax.Components) != 1 {
				t.Fatalf("free cost syntax = %#v, want exactly one component", ability.CostSyntax)
			}
			component := ability.CostSyntax.Components[0]
			if component.Kind != test.wantComponent || component.AmountValue != test.wantAmount {
				t.Fatalf("free cost component = %#v, want %v amount %d", component, test.wantComponent, test.wantAmount)
			}
		})
	}
}

func TestParseFreeAlternativeCostFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
	}{
		{
			// A compound "and" payment splits into multiple components the free
			// lowering does not model, so it must not parse as a free spell.
			name:   "compound and payment",
			source: "You may pay 3 life and sacrifice a creature rather than pay this spell's mana cost.\nDraw a card.",
		},
		{
			// Only basic land control conditions are modeled; anything else must
			// fail closed rather than silently drop the gate.
			name:   "unmodeled control condition",
			source: "If you control an artifact, you may pay 2 life rather than pay this spell's mana cost.\nDraw a card.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			if alternative := document.Abilities[0].AlternativeCost; alternative != nil &&
				alternative.Kind == SpellAlternativeCostFree {
				t.Fatalf("unexpectedly parsed as free alternative cost: %#v", alternative)
			}
		})
	}
}
