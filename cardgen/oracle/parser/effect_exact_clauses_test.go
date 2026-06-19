package parser

import "testing"

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
		"Return target permanent card from your graveyard to the battlefield.",
		"Return target green card from your graveyard to your hand.",
		"Return target multicolored card from your graveyard to your hand.",
		"Return target colorless card from your graveyard to your hand.",
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
		// Single instant/sorcery types are not retained by the compiler's
		// single-type path, so they would lower to an unrestricted card.
		"Return target sorcery card from your graveyard to your hand.",
		"Return target instant card from your graveyard to your hand.",
		// Supertype, excluded type, and color+type combinations are unrendered.
		"Return target basic land card from your graveyard to the battlefield.",
		"Return target nonland permanent card from your graveyard to the battlefield.",
		"Return target blue creature card from your graveyard to your hand.",
		// "and/or" unions are not the canonical " or " join.
		"Return up to two target instant and/or sorcery cards from your graveyard to your hand.",
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
