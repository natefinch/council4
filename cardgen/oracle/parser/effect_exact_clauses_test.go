package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

func counterEffectExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectCounter {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

func TestExactCounterSpellTypeUnion(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Counter target artifact or enchantment spell.",
		"Counter target enchantment, instant, or sorcery spell.",
	} {
		if !counterEffectExact(t, source) {
			t.Errorf("counterEffectExact(%q) = false, want true", source)
		}
	}
}

func TestExactTargetColorRider(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		kind   EffectKind
		color  TriggerColor
	}{
		{"Counter target spell if it's blue.", EffectCounter, TriggerColorBlue},
		{"Destroy target permanent if it's red.", EffectDestroy, TriggerColorRed},
		{"Counter target spell if it is green.", EffectCounter, TriggerColorGreen},
	}
	for _, tc := range cases {
		document, diagnostics := Parse(tc.source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", tc.source, diagnostics)
		}
		ability := document.Abilities[0]
		effects := ability.Sentences[0].Effects
		if len(effects) != 1 || effects[0].Kind != tc.kind {
			t.Fatalf("Parse(%q) effects = %#v", tc.source, effects)
		}
		if !effects[0].Exact {
			t.Errorf("Parse(%q) effect Exact = false, want true", tc.source)
		}
		if len(effects[0].Targets) != 1 || !effects[0].Targets[0].Exact {
			t.Errorf("Parse(%q) target not exact: %#v", tc.source, effects[0].Targets)
		}
		if len(ability.ConditionClauses) != 1 ||
			ability.ConditionClauses[0].Predicate != ConditionPredicateTargetColor {
			t.Fatalf("Parse(%q) condition clauses = %#v", tc.source, ability.ConditionClauses)
		}
		colors := ability.ConditionClauses[0].Selection.ColorsAny
		if len(colors) != 1 || colors[0] != tc.color {
			t.Errorf("Parse(%q) colors = %#v, want %v", tc.source, colors, tc.color)
		}
	}
}

func TestExactCounterSpellTypeUnionFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Counter target blue enchantment, instant, or sorcery spell.",
		"Counter target enchantment, instant, or sorcery spell an opponent controls.",
		"Counter target enchantment, instant, or sorcery spell with mana value 3 or less.",
		"Counter target enchantment, enchantment, or sorcery spell.",
		"Counter target enchantment and instant spell.",
	} {
		if counterEffectExact(t, source) {
			t.Errorf("counterEffectExact(%q) = true, want false", source)
		}
	}
}

// graveyardReturnExact parses a single graveyard-return sentence and reports
// whether its resolving effect round-tripped to an exact, lowerable production.
func graveyardReturnExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectReturn {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

func TestExactChosenCardsBattlefieldReturn(t *testing.T) {
	t.Parallel()
	source := "Return the chosen cards to the battlefield tapped."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || !effects[0].Exact {
		t.Fatalf("effects = %#v; want one exact chosen-cards return", effects)
	}
	if len(effects[0].References) != 1 ||
		effects[0].References[0].Kind != ReferenceChosenCards {
		t.Fatalf("effect references = %#v; want chosen-cards reference", effects[0].References)
	}
}

func TestExactChosenCardsBattlefieldReturnFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Return the selected cards to the battlefield tapped.",
		"Return the chosen card to the battlefield tapped.",
		"Return the chosen cards to the battlefield.",
		"Return the chosen cards to the battlefield tapped under your control.",
	} {
		if graveyardReturnExact(t, source) {
			t.Errorf("graveyardReturnExact(%q) = true, want false", source)
		}
	}
}

func TestExactChosenCreatureCardsInYourGraveyardTarget(t *testing.T) {
	t.Parallel()
	source := "Choose two target creature cards in your graveyard."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	targets := document.Abilities[0].Sentences[0].Targets
	if len(targets) != 1 {
		t.Fatalf("targets = %#v; want one target group", targets)
	}
	target := targets[0]
	if !target.Exact ||
		target.Cardinality != (TargetCardinalitySyntax{Min: 2, Max: 2}) ||
		target.Selection.Kind != SelectionCreature ||
		target.Selection.Controller != SelectionControllerYou ||
		target.Selection.Zone != zone.Graveyard {
		t.Fatalf("target = %#v; want exact two creature cards in your graveyard", target)
	}
	if got := shared.SliceSpan(source, target.ChoiceSpan); got != "Choose" {
		t.Fatalf("choice span = %q; want %q", got, "Choose")
	}
}

func TestExactGraveyardCardTargetAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Return target card from your graveyard to your hand.",
		"Return target creature card from your graveyard to your hand.",
		"Return target artifact card from an opponent's graveyard to your hand.",
		"Return target land card from a graveyard to your hand.",
		"Return target creature or enchantment card from your graveyard to your hand.",
		"Return target artifact or creature card from your graveyard to the battlefield.",
		"Return target instant or sorcery card from your graveyard to your hand.",
		// Single instant/sorcery types have no permanent selection Kind but are
		// retained in RequiredTypesAny, so lowering restricts to the type.
		"Return target sorcery card from your graveyard to your hand.",
		"Return target instant card from your graveyard to your hand.",
		"Return target permanent card from your graveyard to the battlefield.",
		"Return target green card from your graveyard to your hand.",
		"Return target multicolored card from your graveyard to your hand.",
		"Return target colorless card from your graveyard to your hand.",
		// Color + type/subtype combinations render in canonical order; lowering
		// restricts on both the color and the type.
		"Return target blue creature card from your graveyard to your hand.",
		"Return target red sorcery card from your graveyard to your hand.",
		"Return target colorless artifact card from your graveyard to the battlefield.",
		"Return target Zombie card from your graveyard to the battlefield.",
		"Return target creature card with mana value 3 or less from your graveyard to your hand.",
		"Return another target creature card from your graveyard to your hand.",
		"Return two target creature cards from your graveyard to your hand.",
		"Return up to two target creature cards from your graveyard to your hand.",
		"Return up to three target permanent cards from your graveyard to your hand.",
		"Return up to two target Zombie cards from your graveyard to the battlefield.",
		"Return up to two target cards with cycling from your graveyard to your hand.",
	}
	for _, source := range accepted {
		if !graveyardReturnExact(t, source) {
			t.Errorf("graveyardReturnExact(%q) = false, want true", source)
		}
	}
}

func TestExactGraveyardCardTargetFailsClosed(t *testing.T) {
	t.Parallel()
	// Each of these carries a qualifier the canonical graveyard-card phrasing
	// cannot faithfully reconstruct, so the round-trip must fail closed and the
	// card must keep failing rather than lower to a wrong predicate.
	rejected := []string{
		// Supertype and excluded-type combinations are unrendered.
		"Return target basic land card from your graveyard to the battlefield.",
		"Return target nonland permanent card from your graveyard to the battlefield.",
		// "and/or" unions are not the canonical " or " join.
		"Return up to two target instant and/or sorcery cards from your graveyard to your hand.",
	}
	for _, source := range rejected {
		if graveyardReturnExact(t, source) {
			t.Errorf("graveyardReturnExact(%q) = true, want false (fail closed)", source)
		}
	}
}

// TestExactChosenGraveyardReturnAccepts covers the non-target "Return a
// <filter> card from your graveyard to your hand" recursion wording, chosen at
// resolution rather than targeted. It reuses the same canonical noun-phrase
// reconstruction as the targeted path, so the same card filters round-trip.
func TestExactChosenGraveyardReturnAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Return a card from your graveyard to your hand.",
		"Return a creature card from your graveyard to your hand.",
		"Return a creature or planeswalker card from your graveyard to your hand.",
		"Return an artifact card from your graveyard to your hand.",
		"Return a permanent card from your graveyard to your hand.",
		"Return a green card from your graveyard to your hand.",
		// A single instant/sorcery type is retained in RequiredTypesAny.
		"Return a sorcery card from your graveyard to your hand.",
		"Return an instant card from your graveyard to your hand.",
		// Color + type combinations render in canonical order.
		"Return a blue creature card from your graveyard to your hand.",
		"Return a red sorcery card from your graveyard to your hand.",
		"Return a creature card with mana value 3 or less from your graveyard to your hand.",
	}
	for _, source := range accepted {
		if !graveyardReturnExact(t, source) {
			t.Errorf("graveyardReturnExact(%q) = false, want true", source)
		}
	}
}

// TestExactChosenGraveyardReturnFailsClosed verifies that non-target graveyard
// returns whose filter, zone, or destination the round-trip cannot faithfully
// reconstruct keep failing rather than lowering to a wrong predicate.
func TestExactChosenGraveyardReturnFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		// Supertype combinations are unrendered.
		"Return a basic land card from your graveyard to your hand.",
		// Another player's graveyard has no non-target "your" phrasing here.
		"Return a creature card from an opponent's graveyard to your hand.",
	}
	for _, source := range rejected {
		if graveyardReturnExact(t, source) {
			t.Errorf("graveyardReturnExact(%q) = true, want false (fail closed)", source)
		}
	}
}

// destroyEffectExact parses a single destroy sentence and reports whether its
// resolving effect round-tripped to an exact, lowerable production.
func destroyEffectExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectDestroy {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

func TestExactDestroyMassSubtypeAndNumericAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		// Bare-subtype mass groups select any permanent with the subtype.
		"Destroy all Islands.",
		"Destroy all Goblins.",
		"Destroy all Auras.",
		"Destroy all Plains.",
		"Destroy all Walls.",
		// A subtype before an explicit card type keeps both constraints.
		"Destroy all Dragon creatures.",
		// "untapped" mirrors the existing "tapped" mass prefix.
		"Destroy all untapped creatures.",
		// Numeric mass groups now extend past creatures to other nouns and an
		// optional excluded-type prefix; mana value applies to every permanent.
		"Destroy all artifacts with mana value 3 or less.",
		"Destroy all nonland permanents with mana value 1 or less.",
		"Destroy all permanents with mana value 2 or greater.",
		// Nonbasic lands are the canonical excluded-supertype mass group.
		"Destroy all nonbasic lands.",
	}
	for _, source := range accepted {
		if !destroyEffectExact(t, source) {
			t.Errorf("destroyEffectExact(%q) = false, want true", source)
		}
	}
}

func TestExactDestroyMassEachAccepts(t *testing.T) {
	t.Parallel()
	// The singular "each" mass form selects every matching permanent exactly as
	// the plural "all" form does, so it must round-trip to an exact group destroy.
	accepted := []string{
		"Destroy each creature.",
		"Destroy each artifact.",
		"Destroy each enchantment.",
		"Destroy each permanent.",
		"Destroy each nonland permanent.",
		"Destroy each nonland permanent with mana value 2 or less.",
		"Destroy each creature with power 3 or greater.",
		"Destroy each tapped creature.",
		"Destroy each creature you control.",
		"Destroy each nonbasic land.",
	}
	for _, source := range accepted {
		if !destroyEffectExact(t, source) {
			t.Errorf("destroyEffectExact(%q) = false, want true", source)
		}
	}
}

func TestExactDestroyMassEachFailsClosed(t *testing.T) {
	t.Parallel()
	// "each" mass groups inherit the same fail-closed shapes as the plural form;
	// power/toughness exist only on creatures and multi-qualifier subtypes have
	// no canonical singular round-trip.
	rejected := []string{
		"Destroy each artifact with power 3 or less.",
	}
	for _, source := range rejected {
		if destroyEffectExact(t, source) {
			t.Errorf("destroyEffectExact(%q) = true, want false (fail closed)", source)
		}
	}
}

func TestExactDestroyMassFailsClosed(t *testing.T) {
	t.Parallel()
	// Each carries a shape the canonical mass phrasing cannot faithfully
	// reconstruct, so the round-trip must fail closed rather than lower wrong.
	rejected := []string{
		// Power and toughness exist only on creatures; a non-creature numeric
		// mass group must not silently apply them.
		"Destroy all artifacts with power 3 or less.",
		// A subtype paired with another qualifier is unrendered.
		"Destroy all tapped Goblins.",
		// Multi-subtype mass groups are not the single-subtype shape.
		"Destroy all Auras and Equipment.",
	}
	for _, source := range rejected {
		if destroyEffectExact(t, source) {
			t.Errorf("destroyEffectExact(%q) = true, want false (fail closed)", source)
		}
	}
}

func TestExactDestroyTypeUnionManaValueAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Destroy target creature or planeswalker with mana value 3 or less.",
		"Destroy target artifact or enchantment with mana value 4 or less.",
		"Destroy target artifact or land with mana value 2 or greater.",
	}
	for _, source := range accepted {
		if !destroyEffectExact(t, source) {
			t.Errorf("destroyEffectExact(%q) = false, want true", source)
		}
	}
}

func TestExactDestroyTypeUnionManaValueFailsClosed(t *testing.T) {
	t.Parallel()
	// Power and toughness on a type union would silently drop the union members
	// that never carry that characteristic, and a controller clause combined
	// with a mana-value qualifier is an unreconstructed word order.
	rejected := []string{
		"Destroy target artifact, enchantment, or creature with power 4 or greater.",
		"Destroy target creature or planeswalker you control with mana value 3 or less.",
	}
	for _, source := range rejected {
		if destroyEffectExact(t, source) {
			t.Errorf("destroyEffectExact(%q) = true, want false (fail closed)", source)
		}
	}
}

func TestExactPermanentTypeUnionRejectsSpellOnlyTypes(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Destroy target creature or instant.",
		"Destroy target land or sorcery.",
		"Destroy target artifact, creature, or instant.",
	} {
		if destroyEffectExact(t, source) {
			t.Errorf("destroyEffectExact(%q) = true, want false", source)
		}
	}
}

// exileEffectExact parses a single exile sentence and reports whether its
// resolving effect round-tripped to an exact, lowerable production.
func exileEffectExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectExile {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

// TestExactOxfordTypeUnionAccepts proves a three-or-more-member card-type union
// written as an Oxford-comma list round-trips to an exact production, the same
// way the two-member "X or Y" union already does.
func TestExactOxfordTypeUnionAccepts(t *testing.T) {
	t.Parallel()
	exileSources := []string{
		"Exile target artifact, creature, or enchantment.",
		"Exile target artifact, creature, or land.",
		"Exile target artifact, creature, or enchantment an opponent controls.",
	}
	for _, source := range exileSources {
		if !exileEffectExact(t, source) {
			t.Errorf("exileEffectExact(%q) = false, want true", source)
		}
	}
	destroySources := []string{
		"Destroy target artifact, creature, or planeswalker.",
		"Destroy target artifact, enchantment, or planeswalker.",
	}
	for _, source := range destroySources {
		if !destroyEffectExact(t, source) {
			t.Errorf("destroyEffectExact(%q) = false, want true", source)
		}
	}
}

// TestExactNoncreatureTypeUnionAccepts proves a card-type union qualified by a
// single excluded type ("noncreature artifact or noncreature enchantment")
// round-trips exact in both Oracle renderings — the qualifier repeated on every
// member and printed once before the union ("noncreature artifact or
// enchantment"). Both describe the same selection. Haywire Mite and Guerrilla
// Gorilla print the repeated form; Hulk's Thunderclap prints the once-front form.
func TestExactNoncreatureTypeUnionAccepts(t *testing.T) {
	t.Parallel()
	exileSources := []string{
		"Exile target noncreature artifact or noncreature enchantment.",
		"Exile target noncreature artifact or enchantment.",
	}
	for _, source := range exileSources {
		if !exileEffectExact(t, source) {
			t.Errorf("exileEffectExact(%q) = false, want true", source)
		}
	}
	destroySources := []string{
		"Destroy target noncreature artifact or noncreature enchantment.",
		"Destroy target noncreature artifact or enchantment.",
	}
	for _, source := range destroySources {
		if !destroyEffectExact(t, source) {
			t.Errorf("destroyEffectExact(%q) = false, want true", source)
		}
	}
}

// TestExactNoncreatureTypeUnionFailsClosed keeps qualified unions the round-trip
// cannot faithfully reconstruct outside the exact envelope: a member-specific
// excluded type ("noncreature artifact or nonland enchantment") and more than one
// excluded type on the union.
func TestExactNoncreatureTypeUnionFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Exile target noncreature artifact or nonland enchantment.",
		"Exile target noncreature nonland artifact or enchantment.",
	}
	for _, source := range rejected {
		if exileEffectExact(t, source) {
			t.Errorf("exileEffectExact(%q) = true, want false (fail closed)", source)
		}
	}
}

// TestExactSubtypeUnionAccepts proves a union of subtypes that stands in for the
// permanent noun ("target Skeleton, Vampire, or Zombie") round-trips exact.
func TestExactSubtypeUnionAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Exile target Skeleton, Vampire, or Zombie.",
		"Exile target Skeleton, Spirit, or Zombie.",
	}
	for _, source := range accepted {
		if !exileEffectExact(t, source) {
			t.Errorf("exileEffectExact(%q) = false, want true", source)
		}
	}
}

// TestExactOxfordUnionFailsClosed keeps unions the runtime predicate cannot
// faithfully reconstruct outside the exact envelope: a per-member keyword or
// power qualifier, and a union that mixes a card type with a subtype.
func TestExactOxfordUnionFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Exile target artifact, enchantment, or creature with flying.",
		"Exile target artifact, enchantment, or creature with power 4 or greater.",
		"Exile target creature or Spacecraft.",
	}
	for _, source := range rejected {
		if exileEffectExact(t, source) {
			t.Errorf("exileEffectExact(%q) = true, want false (fail closed)", source)
		}
	}
}

// TestExactGraveyardExileAccepts proves "Exile <target> from <owner> graveyard."
// round-trips to an exact production for the canonical owner suffixes, typed and
// plain card nouns, and "up to N" counts, the same graveyard-card target the
// return and put paths already accept.
func TestExactGraveyardExileAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Exile target card from a graveyard.",
		"Exile target card from your graveyard.",
		"Exile target card from an opponent's graveyard.",
		"Exile target creature card from a graveyard.",
		"Exile target artifact card from a graveyard.",
		"Exile up to one target card from a graveyard.",
	}
	for _, source := range accepted {
		if !exileEffectExact(t, source) {
			t.Errorf("exileEffectExact(%q) = false, want true", source)
		}
	}
}

// TestExactGraveyardExileFailsClosed keeps graveyard-exile wordings the canonical
// owner-suffix reconstruction cannot render outside the exact envelope: the "from
// a single graveyard" shared-graveyard constraint and a power qualifier that
// graveyard cards never carry in printed Oracle text.
func TestExactGraveyardExileFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Exile up to three target cards from a single graveyard.",
		"Exile target creature card with power 4 or greater from a graveyard.",
	}
	for _, source := range rejected {
		if exileEffectExact(t, source) {
			t.Errorf("exileEffectExact(%q) = true, want false (fail closed)", source)
		}
	}
}

// shuffleSelfIntoLibraryExact parses a single shuffle sentence with the given
// card name and reports whether the resulting effect is exact.
func shuffleSelfIntoLibraryExact(t *testing.T, cardName, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true, CardName: cardName})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectShuffle {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

func TestExactSourceSpellShuffleIntoLibraryAccepts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cardName string
		source   string
	}{
		{"Green Sun's Zenith", "Shuffle Green Sun's Zenith into its owner's library."},
		{"Beacon of Destruction", "Shuffle Beacon of Destruction into its owner's library."},
	}
	for _, tc := range cases {
		if !shuffleSelfIntoLibraryExact(t, tc.cardName, tc.source) {
			t.Errorf("shuffleSelfIntoLibraryExact(%q, %q) = false, want true", tc.cardName, tc.source)
		}
	}
}

func TestExactSourceSpellShuffleIntoLibraryFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cardName string
		source   string
	}{
		{"Foo", "Shuffle target creature into its owner's library."},
		{"Foo", "Shuffle your graveyard into your library."},
	}
	for _, tc := range cases {
		if shuffleSelfIntoLibraryExact(t, tc.cardName, tc.source) {
			t.Errorf("shuffleSelfIntoLibraryExact(%q, %q) = true, want false", tc.cardName, tc.source)
		}
	}
}
