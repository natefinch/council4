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
		"Counter target enchantment, instant, or sorcery spell an opponent controls.",
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
		// A single subtype adjective may also qualify a card-type noun ("Zombie
		// creature card", "Hero creature cards"); the subtype precedes the type
		// noun and lowering restricts on both the subtype and the type.
		"Return target Zombie creature card from your graveyard to the battlefield.",
		"Return target Hero creature card from your graveyard to your hand.",
		"Return target Zombie creature card from your graveyard to the battlefield tapped.",
		"Return up to three target Hero creature cards from your graveyard to the battlefield.",
		"Return target creature card with mana value 3 or less from your graveyard to your hand.",
		// A power or toughness numeric qualifier renders in canonical order and
		// lowering restricts the graveyard-card selection on the bound.
		"Return target creature card with power 2 or less from your graveyard to your hand.",
		"Return target creature card with toughness 3 or greater from your graveyard to the battlefield.",
		// A single supertype adjective ("basic", "legendary", "snow") precedes
		// the card-type noun and lowering restricts on the supertype.
		"Return target basic land card from your graveyard to the battlefield.",
		"Return target legendary creature card from your graveyard to your hand.",
		// "Return X target <type> cards" consumes the X into the effect amount,
		// rendering "Return X " before the plural noun phrase.
		"Return X target creature cards from your graveyard to your hand.",
		"Return X target creature cards from your graveyard to the battlefield.",
		"Return another target creature card from your graveyard to your hand.",
		"Return two target creature cards from your graveyard to your hand.",
		"Return up to two target creature cards from your graveyard to your hand.",
		"Return up to three target permanent cards from your graveyard to your hand.",
		"Return up to two target Zombie cards from your graveyard to the battlefield.",
		"Return up to two target cards with cycling from your graveyard to your hand.",
		// Three-or-more type unions render with the serial-comma "X, Y, or Z"
		// form for a single target and the "X, Y, and/or Z" form for plural
		// multi-target counts.
		"Return target artifact, creature, or enchantment card from your graveyard to the battlefield.",
		"Return target instant, sorcery, or creature card from your graveyard to your hand.",
		"Return up to two target artifact, creature, and/or enchantment cards from your graveyard to your hand.",
		// Plural multi-target type unions join with "and/or" rather than "or".
		"Return up to two target instant and/or sorcery cards from your graveyard to your hand.",
		// "one or two"/"one, two, or three" target counts are exact cardinalities.
		"Return one or two target creature cards from your graveyard to your hand.",
		"Return one, two, or three target creature cards from your graveyard to your hand.",
		// Owner-relative hand destinations lower identically to "to your hand"
		// because returned cards always move to their owner's hand.
		"Return target creature card from your graveyard to its owner's hand.",
		"Return up to two target creature cards from your graveyard to their hand.",
		"Return up to three target creature cards from your graveyard to their owners' hands.",
		// An excluded card type ("nonland permanent") renders as a "non<type>"
		// prefix on the card noun and round-trips through ExcludedTypes.
		"Return target nonland permanent card from your graveyard to the battlefield.",
		"Return target noncreature, nonland card from your graveyard to your hand.",
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
		// A subtype-qualified type noun is exact, but a supertype exclusion on
		// that noun ("nonlegendary creature card") is still unrendered, so it must
		// keep failing closed rather than dropping the exclusion.
		"Return target nonlegendary creature card from your graveyard to the battlefield.",
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
		// A single supertype adjective precedes the card-type noun and lowering
		// restricts on the supertype.
		"Return a basic land card from your graveyard to your hand.",
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

func TestExactDestroyMassChosenTypeAcceptsAndRecordsField(t *testing.T) {
	t.Parallel()
	// "Choose a creature type. Destroy all creatures [that aren't] of the chosen
	// type." (Kindred Dominance and its positive sibling) round-trips to an exact
	// group destroy whose selection records the resolution chosen-type filter.
	cases := []struct {
		phrase   string
		excluded bool
	}{
		{"Destroy all creatures that aren't of the chosen type.", true},
		{"Destroy all creatures of the chosen type.", false},
	}
	for _, tc := range cases {
		document, diagnostics := Parse(tc.phrase, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", tc.phrase, diagnostics)
		}
		effects := document.Abilities[0].Sentences[0].Effects
		if len(effects) != 1 || effects[0].Kind != EffectDestroy {
			t.Fatalf("Parse(%q) effects = %#v", tc.phrase, effects)
		}
		if !effects[0].Exact {
			t.Errorf("destroyEffectExact(%q) = false, want true", tc.phrase)
		}
		if got := effects[0].Selection.SubtypeFromChosenTypeExcluded; got != tc.excluded {
			t.Errorf("%q SubtypeFromChosenTypeExcluded = %v, want %v", tc.phrase, got, tc.excluded)
		}
		if got := effects[0].Selection.SubtypeFromChosenType; got != !tc.excluded {
			t.Errorf("%q SubtypeFromChosenType = %v, want %v", tc.phrase, got, !tc.excluded)
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
	// A mana-value qualifier on a comma-free two-member union ("creature or
	// planeswalker ... with mana value 3 or less") applies to the whole union in
	// a word order the round-trip does not reconstruct; the qualified Oxford-list
	// disjunction (handled elsewhere) is a distinct, supported shape.
	rejected := []string{
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
// faithfully reconstruct outside the exact envelope: a union that mixes a card
// type with a subtype. A per-member keyword or power qualifier on an
// Oxford-comma list is a distinct, supported shape handled by the qualified
// disjunctive permanent target.
func TestExactOxfordUnionFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
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
		// A power or toughness numeric qualifier renders in canonical order and
		// lowering restricts the graveyard-card selection on the bound, the same
		// as the graveyard return and put paths.
		// "from a single graveyard" restricts every chosen card to one shared
		// graveyard; it round-trips on the any-graveyard owner relation.
		"Exile up to three target cards from a single graveyard.",
		"Exile up to two target cards from a single graveyard.",
		// The plural "from graveyards" owner relation lets each chosen card lie
		// in a different graveyard; it round-trips with no same-graveyard flag.
		"Exile up to two target cards from graveyards.",
		"Exile up to three target cards from graveyards.",
	}
	for _, source := range accepted {
		if !exileEffectExact(t, source) {
			t.Errorf("exileEffectExact(%q) = false, want true", source)
		}
	}
}

// TestExactGraveyardExileFailsClosed keeps graveyard-exile wordings the canonical
// owner-suffix reconstruction cannot render outside the exact envelope: a "single"
// qualifier on an owner-named graveyard ("your single graveyard"), which names one
// graveyard already and has no canonical "from a single graveyard" rendering.
func TestExactGraveyardExileFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Exile target card from your single graveyard.",
	}
	for _, source := range rejected {
		if exileEffectExact(t, source) {
			t.Errorf("exileEffectExact(%q) = true, want false (fail closed)", source)
		}
	}
}

// TestParseSingleGraveyardSelectionFlag proves the parser captures the "from a
// single graveyard" qualifier as the typed SingleGraveyard flag on an
// any-controller graveyard card selection, so the compiler and lowering can carry
// the same-graveyard restriction without inspecting wording.
func TestParseSingleGraveyardSelectionFlag(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Exile up to three target cards from a single graveyard.", Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	targets := document.Abilities[0].Sentences[0].Targets
	if len(targets) != 1 {
		t.Fatalf("targets = %#v, want one", targets)
	}
	selection := targets[0].Selection
	if selection.Zone != zone.Graveyard {
		t.Fatalf("zone = %v, want Graveyard", selection.Zone)
	}
	if selection.Controller != SelectionControllerAny {
		t.Fatalf("controller = %v, want any", selection.Controller)
	}
	if !selection.SingleGraveyard {
		t.Fatal("SingleGraveyard = false, want true")
	}
}
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
		{"The Mending of Dominaria", "Shuffle your graveyard into your library."},
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
	}
	for _, tc := range cases {
		if shuffleSelfIntoLibraryExact(t, tc.cardName, tc.source) {
			t.Errorf("shuffleSelfIntoLibraryExact(%q, %q) = true, want false", tc.cardName, tc.source)
		}
	}
}
