package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SelectionDim names one optional CompiledSelector predicate dimension that a
// calling context may need to suppress. SelectionForSelector honors every
// dimension by default (the canonical superset); a SelectionMask removes the
// dimensions a particular runtime context cannot express. The constants are bit
// flags so a mask can name several dimensions at once.
type SelectionDim uint32

// SelectionDim flags. Each names a dimension that SelectionForSelector maps onto
// a game.Selection field when honored. The dimensions a context omits compose
// into a SelectionMask through Ignoring / Rejecting.
const (
	// DimExcludeSource is the "another" / "other" self-exclusion that drops the
	// predicate's own source object (Selection.ExcludeSource).
	DimExcludeSource SelectionDim = 1 << iota

	// DimExcludedSupertype is the single excluded supertype ("nonbasic",
	// "nonlegendary") carried by Selection.ExcludedSupertype.
	DimExcludedSupertype

	// DimExcludedSubtype is the single excluded creature subtype ("non-Human")
	// carried by Selection.ExcludedSubtype.
	DimExcludedSubtype

	// DimMatchAnyCounter is the kind-agnostic "with a counter on it" qualifier
	// (Selection.MatchAnyCounter).
	DimMatchAnyCounter

	// DimSubtypeChoiceExcluded is the "aren't of the chosen type" resolution
	// restriction (Selection.SubtypeChoice == SubtypeChoiceResolutionExcluded).
	// The positive entry-choice and chosen-type sources are always honored.
	DimSubtypeChoiceExcluded

	// DimConjunctiveTypes routes a multi-type set through the conjunctive
	// RequiredTypes filter instead of the default RequiredTypesAny union.
	DimConjunctiveTypes

	// DimNonToken is the "nontoken" qualifier (Selection.NonToken).
	DimNonToken

	// DimTokenOnly is the "token" qualifier (Selection.TokenOnly).
	DimTokenOnly

	// DimHistoric is the historic disjunction (artifact, legendary, or Saga)
	// lowered to a Selection.AnyOf of those three alternatives.
	DimHistoric

	// DimPowerVsSource is the source-relative "with lesser/greater power"
	// comparison (Selection.PowerLessThanSource / PowerGreaterThanSource).
	DimPowerVsSource

	// DimRequiredName is the "named <Name>" filter that requires the matched
	// object's card name to equal a verbatim name (Selection.Name). Most
	// contexts cannot represent it and reject it; the battlefield group
	// recipient honors it ("each other creature you control named Charmed
	// Stray").
	DimRequiredName

	// DimNameUniqueAmongControlled is the "that doesn't have the same name as
	// another permanent you control" restriction (Yenna, Redtooth Regent),
	// carried by Selection.NameUniqueAmongControlled.
	DimNameUniqueAmongControlled

	// DimMatchNoCounters is the negated "with no counters on it/them" qualifier
	// (Selection.MatchNoCounters), matching a permanent that carries no counters.
	DimMatchNoCounters

	// DimMatchExcludedCounter is the kind-specific negated "without a <kind>
	// counter on it/them" qualifier (Selection.MatchExcludedCounter), matching a
	// permanent that carries no counter of Selection.ExcludedCounter's kind.
	DimMatchExcludedCounter
)

// SelectionMask records the optional CompiledSelector dimensions a calling
// context cannot honor. SelectionForSelectorMasked drops every dimension in
// Ignore that is active in the selector and fails closed on every dimension in
// Reject that is active; all other dimensions are honored as by the canonical
// superset. A dimension named in neither set is honored.
//
// Ignore reproduces a projector that historically dropped an unsupported
// qualifier (the qualifier never reached that context, so dropping it is
// behavior-preserving). Reject reproduces a projector that failed closed on the
// qualifier, the conservative text-blind default for a dimension a context
// genuinely cannot represent.
type SelectionMask struct {
	ignore SelectionDim
	reject SelectionDim
}

// Ignoring returns a copy of the mask that additionally drops the named
// dimensions when they are active in the selector.
func (m SelectionMask) Ignoring(dims ...SelectionDim) SelectionMask {
	for _, d := range dims {
		m.ignore |= d
	}
	return m
}

// Rejecting returns a copy of the mask that additionally fails closed on the
// named dimensions when they are active in the selector.
func (m SelectionMask) Rejecting(dims ...SelectionDim) SelectionMask {
	for _, d := range dims {
		m.reject |= d
	}
	return m
}

// dimension reports how the mask treats a dimension that is active in the
// selector: honor maps it, drop skips it, and ok is false when the dimension is
// rejected and the projection must fail closed. An inactive dimension yields
// honor=false, drop=false, ok=true so callers can guard with a single check.
func (m SelectionMask) dimension(active bool, dim SelectionDim) (honor, ok bool) {
	if !active {
		return false, true
	}
	if m.reject&dim != 0 {
		return false, false
	}
	if m.ignore&dim != 0 {
		return false, true
	}
	return true, true
}

// SelectionForSelector projects a CompiledSelector onto the canonical
// game.Selection that matches the same objects. It is the single superset
// projector: it maps every selector predicate dimension that the executable
// backend can represent onto its Selection field and fails closed on the
// dimensions no Selection field can carry. Callers that run in a context unable
// to honor a particular dimension narrow the projection with
// SelectionForSelectorMasked instead of hand-writing a bespoke projector.
//
// It returns false when the selector carries a predicate the backend cannot
// represent (a zone, basic-land-type, source-type, or total-mana-value filter,
// a player-or-planeswalker target, an inclusive one-of-each set, a chosen-{X}
// mana-value bound, a contradictory tapped state, more than one excluded
// supertype or subtype, or a structurally invalid combination) so unsupported
// wordings stay unsupported rather than silently dropping a constraint. The
// named-card filter (Selection.Name) is honored as a dimension; a context that
// cannot represent it rejects DimRequiredName through its mask.
func SelectionForSelector(selector compiler.CompiledSelector) (game.Selection, bool) {
	return SelectionForSelectorMasked(selector, SelectionMask{})
}

// SelectionForSelectorMasked is SelectionForSelector restricted by a mask. A
// dimension the mask ignores is dropped from the projection (the calling
// context guarantees it never appears); a dimension the mask rejects fails the
// projection closed. The empty mask reproduces SelectionForSelector.
func SelectionForSelectorMasked(selector compiler.CompiledSelector, mask SelectionMask) (game.Selection, bool) {
	// Predicate dimensions with no Selection field: every context fails closed
	// when one is active so no projection silently drops the constraint.
	if selector.Zone != zone.None ||
		selector.BasicLandType ||
		selector.PlayerOrPlaneswalker ||
		selector.MatchTotalManaValue ||
		selector.ManaValueDynamic != compiler.DynamicAmountNone ||
		selector.ManaValueLessThanEventPermanent ||
		selector.ManaValueDynamicCount != nil ||
		selector.InclusiveOneOfEach ||
		len(selector.SourceTypes()) != 0 ||
		(selector.Tapped && selector.Untapped) {
		return game.Selection{}, false
	}

	selection := game.Selection{
		RequiredTypesAny:    slices.Clone(selector.RequiredTypesAny()),
		ExcludedTypes:       slices.Clone(selector.ExcludedTypes()),
		Supertypes:          slices.Clone(selector.Supertypes()),
		SubtypesAny:         slices.Clone(selector.SubtypesAny()),
		ColorsAny:           slices.Clone(selector.ColorsAny()),
		ExcludedColors:      slices.Clone(selector.ExcludedColors()),
		Colorless:           selector.Colorless,
		Multicolored:        selector.Multicolored,
		EnteredThisTurn:     selector.EnteredThisTurn,
		DealtDamageThisTurn: selector.DealtDamageThisTurn,
	}

	if honor, ok := mask.dimension(selector.Another || selector.Other, DimExcludeSource); !ok {
		return game.Selection{}, false
	} else if honor {
		selection.ExcludeSource = true
	}

	if honor, ok := mask.dimension(selector.RequiredName != "", DimRequiredName); !ok {
		return game.Selection{}, false
	} else if honor {
		selection.Name = selector.RequiredName
	}

	if excludedSupertypes := selector.ExcludedSupertypes(); len(excludedSupertypes) > 0 {
		honor, ok := mask.dimension(true, DimExcludedSupertype)
		if !ok {
			return game.Selection{}, false
		}
		if honor {
			if len(excludedSupertypes) > 1 {
				return game.Selection{}, false
			}
			selection.ExcludedSupertype = excludedSupertypes[0]
		}
	}

	for _, alternative := range selector.Alternatives {
		lowered, ok := SelectionForSelectorMasked(alternative, mask)
		if !ok {
			return game.Selection{}, false
		}
		selection.AnyOf = append(selection.AnyOf, lowered)
	}

	if len(selection.RequiredTypesAny) == 0 {
		if requiredType, ok := massGroupRequiredType(selector.Kind); ok {
			selection.RequiredTypes = []types.Card{requiredType}
		} else if selector.Kind == compiler.SelectorUnknown {
			// A bare subtype noun ("Destroy all Islands.") selects any permanent
			// carrying that subtype with no card-type restriction; the subtype
			// filter supplies the constraint. Without one, an unrecognized noun
			// has no representable predicate and fails closed.
			if len(selection.SubtypesAny) == 0 {
				return game.Selection{}, false
			}
		} else if selector.Kind != compiler.SelectorPermanent {
			return game.Selection{}, false
		}
	}

	switch selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		selection.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		selection.Controller = game.ControllerOpponent
	case compiler.ControllerNotYou:
		selection.Controller = game.ControllerNotYou
	case compiler.ControllerDefendingPlayer:
		selection.ControlledByDefendingPlayer = true
	default:
		return game.Selection{}, false
	}

	switch {
	case selector.Attacking && selector.Blocking:
		selection.CombatState = game.CombatStateAttackingOrBlocking
	case selector.Attacking:
		selection.CombatState = game.CombatStateAttacking
	case selector.Blocking:
		selection.CombatState = game.CombatStateBlocking
	default:
	}

	switch {
	case selector.Tapped:
		selection.Tapped = game.TriTrue
	case selector.Untapped:
		selection.Tapped = game.TriFalse
	default:
	}

	if selector.MatchManaValue {
		// game.Selection's mana-value bound is a fixed comparison; it cannot
		// express the spell's chosen {X} ("with mana value X or less"), so an
		// X-derived bound fails closed rather than lowering to a wrong fixed bound.
		if selector.ManaValueX {
			return game.Selection{}, false
		}
		selection.ManaValue = opt.Val(selector.ManaValue)
	}
	if selector.MatchPower {
		selection.Power = opt.Val(selector.Power)
	}
	if selector.MatchToughness {
		selection.Toughness = opt.Val(selector.Toughness)
	}

	if selector.Keyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selector.Keyword)
		if !ok {
			return game.Selection{}, false
		}
		selection.Keyword = keyword
	}
	if selector.ExcludedKeyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selector.ExcludedKeyword)
		if !ok {
			return game.Selection{}, false
		}
		selection.ExcludedKeyword = keyword
	}

	if selector.MatchCounter {
		selection.MatchCounter = true
		selection.RequiredCounter = selector.RequiredCounter
	}
	if honor, ok := mask.dimension(selector.MatchAnyCounter, DimMatchAnyCounter); !ok {
		return game.Selection{}, false
	} else if honor {
		selection.MatchAnyCounter = true
	}
	if honor, ok := mask.dimension(selector.MatchNoCounters, DimMatchNoCounters); !ok {
		return game.Selection{}, false
	} else if honor {
		selection.MatchNoCounters = true
	}
	if honor, ok := mask.dimension(selector.MatchExcludedCounter, DimMatchExcludedCounter); !ok {
		return game.Selection{}, false
	} else if honor {
		selection.MatchExcludedCounter = true
		selection.ExcludedCounter = selector.ExcludedCounter
	}

	// The positive attachment/modification qualifiers ("target modified
	// creature", "target enchanted permanent", "target equipped creature") map
	// onto the runtime battlefield predicates. A non-battlefield subject never
	// satisfies them, matching the runtime's own guard, so they are mapped
	// unconditionally like the counter requirement above.
	if selector.Modified {
		selection.MatchModified = true
	}
	if selector.Enchanted {
		selection.MatchEnchanted = true
	}
	if selector.Equipped {
		selection.MatchEquipped = true
	}

	choice, choiceOK := selectorSubtypeChoice(selector, mask)
	if !choiceOK {
		return game.Selection{}, false
	}
	selection.SubtypeChoice = choice

	if selector.ColorFromEntryChoice {
		selection.ColorChoice = game.ColorChoiceSourceEntry
	}

	if honor, ok := mask.dimension(selector.ConjunctiveTypes, DimConjunctiveTypes); !ok {
		return game.Selection{}, false
	} else if honor {
		// "artifact creature" names two card types the permanent must carry at
		// once, so its type set lowers to the conjunctive RequiredTypes (all-of)
		// filter rather than the default any-of RequiredTypesAny union.
		selection.RequiredTypes = selection.RequiredTypesAny
		selection.RequiredTypesAny = nil
	}

	if excludedSubtypes := selector.ExcludedSubtypes(); len(excludedSubtypes) > 0 {
		honor, ok := mask.dimension(true, DimExcludedSubtype)
		if !ok {
			return game.Selection{}, false
		}
		if honor {
			if len(excludedSubtypes) > 1 {
				return game.Selection{}, false
			}
			selection.ExcludedSubtype = excludedSubtypes[0]
		}
	}

	if honor, ok := mask.dimension(selector.NonToken, DimNonToken); !ok {
		return game.Selection{}, false
	} else if honor {
		selection.NonToken = true
	}
	if honor, ok := mask.dimension(selector.TokenOnly, DimTokenOnly); !ok {
		return game.Selection{}, false
	} else if honor {
		selection.TokenOnly = true
	}

	if honor, ok := mask.dimension(selector.Historic, DimHistoric); !ok {
		return game.Selection{}, false
	} else if honor {
		// A historic card is an artifact, a legendary, or a Saga (CR 702.61b).
		// That spans a card type, a supertype, and a subtype, which the flat
		// type/supertype/subtype fields cannot OR together, so it lowers to an
		// AnyOf disjunction conjunctive with the selection's other fields.
		selection.AnyOf = append(selection.AnyOf, historicSelectionAlternatives()...)
	}

	if honor, ok := mask.dimension(selector.PowerLessThanSource || selector.PowerGreaterThanSource, DimPowerVsSource); !ok {
		return game.Selection{}, false
	} else if honor {
		selection.PowerLessThanSource = selector.PowerLessThanSource
		selection.PowerGreaterThanSource = selector.PowerGreaterThanSource
	}

	if honor, ok := mask.dimension(selector.NameUniqueAmongControlled, DimNameUniqueAmongControlled); !ok {
		return game.Selection{}, false
	} else if honor {
		selection.NameUniqueAmongControlled = true
	}

	if len(selection.Validate()) != 0 {
		return game.Selection{}, false
	}
	return selection, true
}

// selectorSubtypeChoice resolves the chosen-creature-type restriction a selector
// imposes, honoring the entry-choice and resolution-chosen-type sources and
// masking the "aren't of the chosen type" exclusion per the mask.
func selectorSubtypeChoice(selector compiler.CompiledSelector, mask SelectionMask) (game.SubtypeChoiceSource, bool) {
	if selector.SubtypeFromChosenTypeExcluded {
		honor, ok := mask.dimension(true, DimSubtypeChoiceExcluded)
		if !ok {
			return game.SubtypeChoiceNone, false
		}
		if honor {
			return game.SubtypeChoiceResolutionExcluded, true
		}
	}
	switch {
	case selector.SubtypeFromChosenType:
		return game.SubtypeChoiceResolution, true
	case selector.SubtypeFromEntryChoice:
		return game.SubtypeChoiceSourceEntry, true
	default:
		return game.SubtypeChoiceNone, true
	}
}

// historicSelectionAlternatives returns the three AnyOf alternatives that define
// a historic card (an artifact, a legendary, or a Saga; CR 702.61b). A historic
// qualifier spans a card type, a supertype, and a subtype, which the flat
// type/supertype/subtype fields cannot OR together, so callers append these as a
// Selection.AnyOf disjunction conjunctive with the selection's other fields.
func historicSelectionAlternatives() []game.Selection {
	return []game.Selection{
		{RequiredTypes: []types.Card{types.Artifact}},
		{Supertypes: []types.Super{types.Legendary}},
		{SubtypesAny: []types.Sub{types.Saga}},
	}
}
