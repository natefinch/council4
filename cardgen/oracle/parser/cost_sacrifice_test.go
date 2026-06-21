package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// soleCostComponent parses source as a single-ability document and returns its
// only typed cost component, failing the test if the structure is unexpected.
func soleCostComponent(t *testing.T, source string) CostComponent {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	syntax := document.Abilities[0].CostSyntax
	if syntax == nil || len(syntax.Components) != 1 {
		t.Fatalf("cost = %#v, want exactly one component", syntax)
	}
	return syntax.Components[0]
}

func TestParseSacrificeSubtypeCostObject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		amount  int
		subtype types.Sub
	}{
		{"singular goblin", "Sacrifice a Goblin: Draw a card.", 1, types.Goblin},
		{"plural treasures", "Sacrifice three Treasures: Draw a card.", 3, types.Treasure},
		{"basic land subtype", "Sacrifice a Forest: Draw a card.", 1, types.Forest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentSacrifice {
				t.Fatalf("kind = %v, want sacrifice", component.Kind)
			}
			if !component.AmountKnown || component.AmountValue != test.amount {
				t.Fatalf("amount = (%d, %v), want %d", component.AmountValue, component.AmountKnown, test.amount)
			}
			if len(component.SubtypesAny) != 1 || component.SubtypesAny[0] != test.subtype {
				t.Fatalf("subtypes = %v, want [%s]", component.SubtypesAny, test.subtype)
			}
			if component.ExcludeSource || component.SourceSelf {
				t.Fatalf("component = %#v, want neither exclude-source nor self", component)
			}
		})
	}
}

func TestParseSacrificeColorCostObject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		color   Color
		noun    ObjectNoun
		exclude bool
	}{
		{"another black creature", "Sacrifice another black creature: Draw a card.", ColorBlack, ObjectNounCreature, true},
		{"a blue creature", "Sacrifice a blue creature: Draw a card.", ColorBlue, ObjectNounCreature, false},
		{"a green permanent you control", "Sacrifice a green permanent you control: Draw a card.", ColorGreen, ObjectNounPermanent, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentSacrifice {
				t.Fatalf("kind = %v, want sacrifice", component.Kind)
			}
			if !component.ObjectColorKnown || component.ObjectColor != test.color {
				t.Fatalf("color = (%v, %v), want %v", component.ObjectColor, component.ObjectColorKnown, test.color)
			}
			if component.ObjectNoun != test.noun {
				t.Fatalf("noun = %v, want %v", component.ObjectNoun, test.noun)
			}
			if component.ExcludeSource != test.exclude {
				t.Fatalf("ExcludeSource = %v, want %v", component.ExcludeSource, test.exclude)
			}
			if !component.AmountKnown || component.AmountValue != 1 {
				t.Fatalf("amount = (%d, %v), want 1", component.AmountValue, component.AmountKnown)
			}
		})
	}
}

func TestParseSacrificeAnotherExcludesSource(t *testing.T) {
	t.Parallel()
	component := soleCostComponent(t, "Sacrifice another creature: Draw a card.")
	if component.Kind != CostComponentSacrifice {
		t.Fatalf("kind = %v, want sacrifice", component.Kind)
	}
	if !component.AmountKnown || component.AmountValue != 1 {
		t.Fatalf("amount = (%d, %v), want 1", component.AmountValue, component.AmountKnown)
	}
	if !component.ExcludeSource {
		t.Fatal("ExcludeSource = false, want true")
	}
	if component.ObjectNoun != ObjectNounCreature {
		t.Fatalf("noun = %v, want creature", component.ObjectNoun)
	}
}

func TestParseSacrificeThisSubtypeIsSelf(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Sacrifice this Aura: Draw a card.",
		"Sacrifice this Equipment: Draw a card.",
	} {
		component := soleCostComponent(t, source)
		if component.Kind != CostComponentSacrifice {
			t.Fatalf("kind = %v, want sacrifice", component.Kind)
		}
		if !component.SourceSelf {
			t.Fatalf("%q: SourceSelf = false, want true", source)
		}
		if component.ExcludeSource || len(component.SubtypesAny) != 0 {
			t.Fatalf("%q: component = %#v, want pure self reference", source, component)
		}
	}
}

func TestParseSacrificeUnrecognizedSubtypeFailsClosed(t *testing.T) {
	t.Parallel()
	// "Wall" is not a permanent subtype the cost grammar recognizes here; the
	// component must stay bare so lowering rejects it.
	component := soleCostComponent(t, "Sacrifice a Bogus: Draw a card.")
	if component.AmountKnown || len(component.SubtypesAny) != 0 || component.SourceSelf || component.ExcludeSource {
		t.Fatalf("component = %#v, want bare unrecognized sacrifice", component)
	}
}

func TestParseSacrificeSubtypeWithNoun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		amount  int
		subtype types.Sub
	}{
		{"creature subtype noun", "Sacrifice a Goblin creature: Draw a card.", 1, types.Goblin},
		{"cleric creature", "Sacrifice a Cleric creature: Draw a card.", 1, types.Cleric},
		{"plural token noun", "Sacrifice two Blood tokens: Draw a card.", 2, types.Blood},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentSacrifice {
				t.Fatalf("kind = %v, want sacrifice", component.Kind)
			}
			if !component.AmountKnown || component.AmountValue != test.amount {
				t.Fatalf("amount = (%d, %v), want %d", component.AmountValue, component.AmountKnown, test.amount)
			}
			if len(component.SubtypesAny) != 1 || component.SubtypesAny[0] != test.subtype {
				t.Fatalf("subtypes = %v, want [%s]", component.SubtypesAny, test.subtype)
			}
		})
	}
}

func TestParseSacrificeTwoSubtypeUnion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		first   types.Sub
		second  types.Sub
		exclude bool
	}{
		{"land subtypes", "Sacrifice a Forest or Plains: Draw a card.", types.Forest, types.Plains, false},
		{"another creature subtypes", "Sacrifice another Orc or Goblin: Draw a card.", types.Orc, types.Goblin, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentSacrifice {
				t.Fatalf("kind = %v, want sacrifice", component.Kind)
			}
			if !component.AmountKnown || component.AmountValue != 1 {
				t.Fatalf("amount = (%d, %v), want 1", component.AmountValue, component.AmountKnown)
			}
			if len(component.SubtypesAny) != 2 ||
				component.SubtypesAny[0] != test.first || component.SubtypesAny[1] != test.second {
				t.Fatalf("subtypes = %v, want [%s %s]", component.SubtypesAny, test.first, test.second)
			}
			if component.ExcludeSource != test.exclude {
				t.Fatalf("excludeSource = %v, want %v", component.ExcludeSource, test.exclude)
			}
		})
	}
}

func TestParseSacrificeNumberedOtherExcludesSource(t *testing.T) {
	t.Parallel()
	component := soleCostComponent(t, "Sacrifice two other creatures: Draw a card.")
	if component.Kind != CostComponentSacrifice {
		t.Fatalf("kind = %v, want sacrifice", component.Kind)
	}
	if !component.AmountKnown || component.AmountValue != 2 {
		t.Fatalf("amount = (%d, %v), want 2", component.AmountValue, component.AmountKnown)
	}
	if !component.ExcludeSource {
		t.Fatal("ExcludeSource = false, want true")
	}
	if component.ObjectNoun != ObjectNounCreature {
		t.Fatalf("noun = %v, want creature", component.ObjectNoun)
	}
}

func TestParseSacrificeSubtypeNounFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		// "artifact creature" is a two-type predicate, not a subtype.
		"Sacrifice an artifact creature: Draw a card.",
		// "Scions" is not a permanent-type noun, so the leading subtype is bare.
		"Sacrifice two Eldrazi Scions: Draw a card.",
		// A subtype that does not belong to the named type fails closed.
		"Sacrifice a Goblin land: Draw a card.",
	} {
		component := soleCostComponent(t, source)
		if component.AmountKnown || len(component.SubtypesAny) != 0 ||
			component.ObjectNoun != ObjectNounUnknown {
			t.Fatalf("%q: component = %#v, want bare unrecognized sacrifice", source, component)
		}
	}
}

func TestParseSacrificeTwoTypeUnion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		first  ObjectNoun
		second ObjectNoun
	}{
		{"artifact or creature", "Sacrifice an artifact or creature: Draw a card.", ObjectNounArtifact, ObjectNounCreature},
		{"creature or planeswalker", "Sacrifice a creature or planeswalker: Draw a card.", ObjectNounCreature, ObjectNounPlaneswalker},
		{"creature or land", "Sacrifice a creature or land: Draw a card.", ObjectNounCreature, ObjectNounLand},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentSacrifice {
				t.Fatalf("kind = %v, want sacrifice", component.Kind)
			}
			if !component.AmountKnown || component.AmountValue != 1 {
				t.Fatalf("amount = (%d, %v), want 1", component.AmountValue, component.AmountKnown)
			}
			if component.ObjectNoun != test.first || component.SecondObjectNoun != test.second {
				t.Fatalf("nouns = (%v, %v), want (%v, %v)",
					component.ObjectNoun, component.SecondObjectNoun, test.first, test.second)
			}
		})
	}
}

func TestParseSacrificeTwoTypeUnionWithArticle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		first   ObjectNoun
		second  ObjectNoun
		exclude bool
	}{
		{"another with second article", "Sacrifice another creature or an enchantment: Draw a card.", ObjectNounCreature, ObjectNounEnchantment, true},
		{"another with second artifact", "Sacrifice another creature or an artifact: Draw a card.", ObjectNounCreature, ObjectNounArtifact, true},
		{"second article without another", "Sacrifice an artifact or a creature: Draw a card.", ObjectNounArtifact, ObjectNounCreature, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentSacrifice {
				t.Fatalf("kind = %v, want sacrifice", component.Kind)
			}
			if !component.AmountKnown || component.AmountValue != 1 {
				t.Fatalf("amount = (%d, %v), want 1", component.AmountValue, component.AmountKnown)
			}
			if component.ObjectNoun != test.first || component.SecondObjectNoun != test.second {
				t.Fatalf("nouns = (%v, %v), want (%v, %v)",
					component.ObjectNoun, component.SecondObjectNoun, test.first, test.second)
			}
			if component.ExcludeSource != test.exclude {
				t.Fatalf("excludeSource = %v, want %v", component.ExcludeSource, test.exclude)
			}
		})
	}
}

func TestParseSacrificeTwoTypeUnionFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		// A repeated type carries no union; treat it as unrecognized.
		"Sacrifice a creature or creature: Draw a card.",
		// "permanent" is not a constraining union member.
		"Sacrifice an artifact or permanent: Draw a card.",
		// A subtype is not a permanent-type union member here.
		"Sacrifice a creature or Goblin: Draw a card.",
		// Planeswalker is valid only as the second union member.
		"Sacrifice a planeswalker or creature: Draw a card.",
	} {
		component := soleCostComponent(t, source)
		if component.AmountKnown || component.SecondObjectNoun != ObjectNounUnknown ||
			component.ObjectNoun != ObjectNounUnknown {
			t.Fatalf("%q: component = %#v, want bare unrecognized sacrifice", source, component)
		}
	}
}

func TestParseExileThisCardFromGraveyardIsSelf(t *testing.T) {
	t.Parallel()
	component := soleCostComponent(t, "Exile this card from your graveyard: Draw a card.")
	if component.Kind != CostComponentExile {
		t.Fatalf("kind = %v, want exile", component.Kind)
	}
	if !component.SourceSelf {
		t.Fatal("SourceSelf = false, want true")
	}
	if component.SourceZone != zone.Graveyard {
		t.Fatalf("source zone = %v, want graveyard", component.SourceZone)
	}
	if component.ObjectIsCard {
		t.Fatal("ObjectIsCard = true, want false for a source self-exile")
	}
}

func TestParseExileThisFromHandIsSelf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
	}{
		{"this card", "Exile this card from your hand: Add {R}."},
		{"this creature", "Exile this creature from your hand: Add {G}."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentExile {
				t.Fatalf("kind = %v, want exile", component.Kind)
			}
			if !component.SourceSelf {
				t.Fatal("SourceSelf = false, want true")
			}
			if component.SourceZone != zone.Hand {
				t.Fatalf("source zone = %v, want hand", component.SourceZone)
			}
			if component.ObjectIsCard {
				t.Fatal("ObjectIsCard = true, want false for a source self-exile")
			}
		})
	}
}

func TestParseExileTypedCardAmounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		source      string
		amountValue int
		amountKnown bool
		amountFromX bool
		noun        ObjectNoun
	}{
		{"single creature", "Exile a creature card from your graveyard: Draw a card.", 1, true, false, ObjectNounCreature},
		{"two creatures", "Exile two creature cards from your graveyard: Draw a card.", 2, true, false, ObjectNounCreature},
		{"five creatures", "Exile five creature cards from your graveyard: Draw a card.", 5, true, false, ObjectNounCreature},
		{"x creatures", "Exile X creature cards from your graveyard: Draw a card.", 0, false, true, ObjectNounCreature},
		{"x cards", "Exile X cards from your graveyard: Draw a card.", 0, false, true, ObjectNounCard},
		{"seven cards", "Exile seven cards from your graveyard: Draw a card.", 7, true, false, ObjectNounCard},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentExile {
				t.Fatalf("kind = %v, want exile", component.Kind)
			}
			if component.AmountValue != test.amountValue || component.AmountKnown != test.amountKnown ||
				component.AmountFromX != test.amountFromX {
				t.Fatalf("amount = (%d, known=%v, fromX=%v), want (%d, %v, %v)",
					component.AmountValue, component.AmountKnown, component.AmountFromX,
					test.amountValue, test.amountKnown, test.amountFromX)
			}
			if component.ObjectNoun != test.noun {
				t.Fatalf("noun = %v, want %v", component.ObjectNoun, test.noun)
			}
			if component.SourceZone != zone.Graveyard {
				t.Fatalf("source zone = %v, want graveyard", component.SourceZone)
			}
		})
	}
}

func TestParseExileSubtypeCardFromGraveyard(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		source  string
		subtype types.Sub
	}{
		{"elf card", "Exile an Elf card from your graveyard: Draw a card.", types.Elf},
		{"assassin card", "Exile an Assassin card from your graveyard: Draw a card.", types.Assassin},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentExile {
				t.Fatalf("kind = %v, want exile", component.Kind)
			}
			if !component.AmountKnown || component.AmountValue != 1 {
				t.Fatalf("amount = (%d, %v), want 1", component.AmountValue, component.AmountKnown)
			}
			if !component.ObjectIsCard || component.ObjectNoun != ObjectNounCard {
				t.Fatalf("object = (isCard=%v, noun=%v), want card", component.ObjectIsCard, component.ObjectNoun)
			}
			if len(component.SubtypesAny) != 1 || component.SubtypesAny[0] != test.subtype {
				t.Fatalf("subtypes = %v, want [%s]", component.SubtypesAny, test.subtype)
			}
			if component.SourceZone != zone.Graveyard {
				t.Fatalf("source zone = %v, want graveyard", component.SourceZone)
			}
		})
	}
}

func TestParseExileGraveyardCardFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		// An unrecognized middle noun leaves the object bare.
		"Exile a Bogus card from your graveyard: Draw a card.",
		// "instant or sorcery" is a card-type union the cost grammar cannot carry.
		"Exile an instant or sorcery card from your graveyard: Draw a card.",
	} {
		component := soleCostComponent(t, source)
		if component.ObjectIsCard || len(component.SubtypesAny) != 0 {
			t.Fatalf("%q: component = %#v, want bare unrecognized exile object", source, component)
		}
	}
}
