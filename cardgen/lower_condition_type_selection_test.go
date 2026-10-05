package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerConditionTypeSelections(t *testing.T) {
	t.Parallel()
	tests := []struct {
		noun string
		want game.Selection
	}{
		{"permanent card", game.Selection{RequiredTypesAny: []types.Card{
			types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker,
		}}},
		{"land card", game.Selection{RequiredTypes: []types.Card{types.Land}}},
		{"artifact card", game.Selection{RequiredTypes: []types.Card{types.Artifact}}},
		{"creature card", game.Selection{RequiredTypes: []types.Card{types.Creature}}},
		{"instant card", game.Selection{RequiredTypes: []types.Card{types.Instant}}},
		{"sorcery card", game.Selection{RequiredTypes: []types.Card{types.Sorcery}}},
		{"artifact creature card", game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}},
		{"artifact or creature card", game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}},
		{"noncreature card", game.Selection{ExcludedTypes: []types.Card{types.Creature}}},
		{"nonland card", game.Selection{ExcludedTypes: []types.Card{types.Land}}},
		{"noncreature artifact card", game.Selection{RequiredTypes: []types.Card{types.Artifact}, ExcludedTypes: []types.Card{types.Creature}}},
		{"nonland permanent card", game.Selection{
			RequiredTypesAny: []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker},
			ExcludedTypes:    []types.Card{types.Land},
		}},
		{"artifact creature or enchantment card", game.Selection{AnyOf: []game.Selection{
			{RequiredTypes: []types.Card{types.Artifact, types.Creature}},
			{RequiredTypes: []types.Card{types.Enchantment}},
		}}},
	}
	for _, test := range tests {
		t.Run(test.noun, func(t *testing.T) {
			t.Parallel()
			sequence := lowerSpellSequence(t, "Type Selection Probe",
				"Destroy target permanent. If it was a "+test.noun+", draw a card.")
			if len(sequence) != 2 {
				t.Fatalf("sequence = %#v, want destroy and gated draw", sequence)
			}
			gate := effectConditionMatch(t, sequence[1])
			if gate.Object.Val.Kind() != game.ObjectReferenceTargetPermanent || gate.Object.Val.TargetIndex() != 0 {
				t.Fatalf("object = %#v, want existing target-permanent binding", gate.Object)
			}
			assertConditionTypeSelection(t, gate.ObjectMatches.Val, test.want)
		})
	}
}

func assertConditionTypeSelection(t *testing.T, got, want game.Selection) {
	t.Helper()
	if !slices.Equal(got.RequiredTypes, want.RequiredTypes) ||
		!slices.Equal(got.RequiredTypesAny, want.RequiredTypesAny) ||
		!slices.Equal(got.ExcludedTypes, want.ExcludedTypes) ||
		len(got.AnyOf) != len(want.AnyOf) || got.RequirePermanentCard {
		t.Fatalf("selection = %#v, want %#v", got, want)
	}
	for i := range got.AnyOf {
		assertConditionTypeSelection(t, got.AnyOf[i], want.AnyOf[i])
	}
}

func TestLowerConditionTypeSelectionsInvalidValues(t *testing.T) {
	t.Parallel()
	for _, invalid := range []types.Card{"", "unexpected"} {
		for _, selection := range []compiler.ConditionSelection{
			{RequiredTypes: []types.Card{invalid}},
			{RequiredTypesAny: []types.Card{invalid}},
			{ExcludedTypes: []types.Card{invalid}},
			{AnyOf: []compiler.ConditionSelection{{ExcludedTypes: []types.Card{invalid}}}},
		} {
			if got, ok := lowerConditionSelection(selection); ok {
				t.Fatalf("accepted invalid selection %#v: %#v", selection, got)
			}
		}
	}
	for _, selection := range []compiler.ConditionSelection{
		{RequiredTypes: []types.Card{types.Creature}, ExcludedTypes: []types.Card{types.Creature}},
		{RequiredTypesAny: []types.Card{types.Creature, types.Land}, ExcludedTypes: []types.Card{types.Creature, types.Land}},
	} {
		if got, ok := lowerConditionSelection(selection); ok {
			t.Fatalf("accepted contradictory selection %#v: %#v", selection, got)
		}
	}
	if _, ok := conditionCardTypes([]types.Card{types.Instant}); ok {
		t.Fatal("unrelated permanent-only adapters must remain restricted")
	}
}

func TestConditionExpandedTypesNeedExecutableBindings(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Exile target card from a graveyard. If it was a permanent card, draw a card.",
		"Exile target card from a graveyard. If it was a noncreature card, draw a card.",
		"When this creature enters, exile target card from a graveyard. If it was a permanent card, draw a card.",
	} {
		card := &ScryfallCard{Name: "Binding Dependency", Layout: "normal", TypeLine: "Creature", OracleText: text, Power: new("1"), Toughness: new("1")}
		if text[0] != 'W' {
			card.TypeLine = "Instant"
		}
		assertCardPaths(t, card, "Object.Val.kind = game.ObjectReferenceTargetCard", "ObjectMatches.Exists = true")
	}
	card := &ScryfallCard{
		Name: "Lion Sash", Layout: "normal", TypeLine: "Artifact Creature — Equipment Cat",
		ManaCost: "{1}{W}", Power: new("1"), Toughness: new("1"),
		OracleText: "{W}: Exile target card from a graveyard. If it was a permanent card, put a +1/+1 counter on this creature.\nEquipped creature gets +1/+1 for each +1/+1 counter on Lion Sash.\nReconfigure {2}",
	}
	assertCardPaths(t, card,
		"ActivatedAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val.Object.Val.kind = game.ObjectReferenceTargetCard",
		"ActivatedAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val.ObjectMatches.Val.RequiredTypesAny[0] = types.Artifact")
}

func TestConditionTypeSelectionDoesNotConsumeCommaQualifierPrefix(t *testing.T) {
	t.Parallel()
	assertCardUnsupported(t, &ScryfallCard{
		Name: "Chrome Replicator", Layout: "normal", TypeLine: "Artifact Creature — Construct",
		ManaCost: "{5}", Power: new("4"), Toughness: new("4"),
		OracleText: "When this creature enters, if you control two or more nonland, nontoken permanents with the same name as one another, create a 4/4 colorless Construct artifact creature token.",
	})
}

func TestConditionTypeSelectionsCardDefPaths(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Type Union CardDef Probe", Layout: "normal", TypeLine: "Instant",
		OracleText: "Destroy target permanent. If it was an artifact or creature, draw a card.",
	}
	assertCardPaths(t, card,
		"SpellAbility.Val.Modes[0].Sequence[1].Condition.Val.Condition.Val.ObjectMatches.Val.RequiredTypesAny[0] = types.Artifact",
		"SpellAbility.Val.Modes[0].Sequence[1].Condition.Val.Condition.Val.ObjectMatches.Val.RequiredTypesAny[1] = types.Creature",
	)
}
