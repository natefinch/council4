package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Colors returns positive color atoms in source order.
func (a Atoms) Colors() []ColorAtom {
	var result []ColorAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomColor {
			result = append(result, ColorAtom{Color: atom.Color, Span: atom.Span})
		}
	}
	return result
}

// ExcludedColors returns non-color atoms in source order.
func (a Atoms) ExcludedColors() []ColorAtom {
	var result []ColorAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomExcludedColor {
			result = append(result, ColorAtom{Color: atom.Color, Span: atom.Span})
		}
	}
	return result
}

// ColorQualifiers returns color-family qualifier atoms in source order.
func (a Atoms) ColorQualifiers() []ColorQualifierAtom {
	var result []ColorQualifierAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomColorQualifier {
			result = append(result, ColorQualifierAtom{Qualifier: atom.Qualifier, Span: atom.Span})
		}
	}
	return result
}

// CardTypes returns card-type atoms in source order.
func (a Atoms) CardTypes() []CardTypeAtom {
	var result []CardTypeAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomCardType {
			result = append(result, CardTypeAtom{Type: atom.CardType, Span: atom.Span})
		}
	}
	return result
}

// ExcludedTypes returns non-card-type atoms in source order.
func (a Atoms) ExcludedTypes() []CardTypeAtom {
	var result []CardTypeAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomExcludedType {
			result = append(result, CardTypeAtom{Type: atom.CardType, Span: atom.Span})
		}
	}
	return result
}

// Supertypes returns supertype atoms in source order.
func (a Atoms) Supertypes() []SupertypeAtom {
	var result []SupertypeAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomSupertype {
			result = append(result, SupertypeAtom{Supertype: atom.Supertype, Span: atom.Span})
		}
	}
	return result
}

// Subtypes returns subtype atoms in source order.
func (a Atoms) Subtypes() []SubtypeAtom {
	var result []SubtypeAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomSubtype {
			result = append(result, SubtypeAtom{Identity: atom.Subtype, Span: atom.Span})
		}
	}
	return result
}

// ObjectNouns returns object-noun atoms in source order.
func (a Atoms) ObjectNouns() []ObjectNounAtom {
	var result []ObjectNounAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomObjectNoun {
			result = append(result, ObjectNounAtom{Noun: atom.ObjectNoun, Span: atom.Span})
		}
	}
	return result
}

// Zones returns zone atoms in source order.
func (a Atoms) Zones() []ZoneAtom {
	var result []ZoneAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomZone {
			result = append(result, ZoneAtom{Zone: atom.Zone, Role: atom.ZoneRole, Span: atom.Span})
		}
	}
	return result
}

// Counters returns counter atoms in source order.
func (a Atoms) Counters() []CounterAtom {
	var result []CounterAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomCounter {
			result = append(result, CounterAtom{Kind: atom.Counter, Span: atom.Span})
		}
	}
	return result
}

// Cardinals returns cardinal atoms in source order.
func (a Atoms) Cardinals() []CardinalAtom {
	var result []CardinalAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomCardinal {
			result = append(result, CardinalAtom{Value: atom.Value, Span: atom.Span})
		}
	}
	return result
}

// Ordinals returns ordinal atoms in source order.
func (a Atoms) Ordinals() []OrdinalAtom {
	var result []OrdinalAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomOrdinal {
			result = append(result, OrdinalAtom{Value: atom.Value, Span: atom.Span})
		}
	}
	return result
}

// SelectionFlags returns selector-flag atoms in source order.
func (a Atoms) SelectionFlags() []SelectionFlagAtom {
	var result []SelectionFlagAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomSelectionFlag {
			result = append(result, SelectionFlagAtom{Flag: atom.Flag, Span: atom.Span})
		}
	}
	return result
}

// Controllers returns controller-relation atoms in source order.
func (a Atoms) Controllers() []ControllerRelationAtom {
	var result []ControllerRelationAtom
	for _, atom := range a.semantic {
		if atom.Kind == atomController {
			result = append(result, ControllerRelationAtom{Relation: atom.Controller, Span: atom.Span})
		}
	}
	return result
}

// References returns explicit references in source order.
func (a Atoms) References() []Reference {
	return append([]Reference(nil), a.references...)
}

// Keywords returns recognized keywords in source order.
func (a Atoms) Keywords() []Keyword {
	return append([]Keyword(nil), a.keywords...)
}

// KeywordSelectors returns recognized keyword selectors in source order.
func (a Atoms) KeywordSelectors() []KeywordSelector {
	return append([]KeywordSelector(nil), a.keywordSelectors...)
}

// KeywordsWithin returns keywords whose name starts at one of the supplied
// tokens. This supports semantic token subsets without re-recognizing spelling.
// A keyword name that is the parameter of a "with <keyword>"/"without <keyword>"
// selector qualifier (e.g. "target creature with flying") is part of that
// selector, not a keyword ability granted by the card, so it is excluded.
func (a Atoms) KeywordsWithin(tokens []shared.Token) []Keyword {
	starts := make(map[int]struct{}, len(tokens))
	for _, token := range tokens {
		starts[token.Span.Start.Offset] = struct{}{}
	}
	var result []Keyword
	for i := range a.keywords {
		if _, ok := starts[a.keywords[i].NameSpan.Start.Offset]; !ok {
			continue
		}
		if a.keywordSelectorCoversName(a.keywords[i].NameSpan) {
			continue
		}
		result = append(result, a.keywords[i])
	}
	return result
}

// keywordSelectorCoversName reports whether a keyword-selector qualifier span
// (the "with <keyword>"/"without <keyword>" clause of a selection) covers the
// given keyword name span, marking the keyword as a selector parameter rather
// than a keyword ability.
func (a Atoms) keywordSelectorCoversName(nameSpan shared.Span) bool {
	for i := range a.keywordSelectors {
		if spanCovers(a.keywordSelectors[i].Span, nameSpan) {
			return true
		}
	}
	return false
}

// KeywordSelectorIn returns the first keyword selector contained by span with
// the requested inclusion relation.
func (a Atoms) KeywordSelectorIn(span shared.Span, excluded bool) (KeywordSelector, bool) {
	for _, selector := range a.keywordSelectors {
		if selector.Excluded == excluded && spanCovers(span, selector.Span) {
			return selector, true
		}
	}
	return KeywordSelector{}, false
}

// KeywordSelectorAt returns the keyword selector whose full span equals span.
func (a Atoms) KeywordSelectorAt(span shared.Span) (KeywordSelector, bool) {
	for _, selector := range a.keywordSelectors {
		if spanEquals(span, selector.Span) {
			return selector, true
		}
	}
	return KeywordSelector{}, false
}

// KeywordSelectorStartingAt returns the keyword selector beginning at span.
func (a Atoms) KeywordSelectorStartingAt(span shared.Span) (KeywordSelector, bool) {
	for _, selector := range a.keywordSelectors {
		if spanStartsAt(selector.Span, span) {
			return selector, true
		}
	}
	return KeywordSelector{}, false
}

// SelfNameSpans returns self-name spans in source order.
func (a Atoms) SelfNameSpans() []shared.Span {
	return append([]shared.Span(nil), a.selfNameSpans...)
}

// SourceNameSpans returns source-name alias spans in source order.
func (a Atoms) SourceNameSpans() []shared.Span {
	return append([]shared.Span(nil), a.sourceNameSpans...)
}

// SourceMarkerSpans returns source-marker spans in source order.
func (a Atoms) SourceMarkerSpans() []shared.Span {
	return append([]shared.Span(nil), a.sourceMarkerSpans...)
}

func appendAtomColor(a *Atoms, color Color, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomColor, Color: color, Span: span})
}

func appendAtomExcludedColor(a *Atoms, color Color, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomExcludedColor, Color: color, Span: span})
}

func appendAtomColorQualifier(a *Atoms, qualifier ColorQualifier, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomColorQualifier, Qualifier: qualifier, Span: span})
}

func appendAtomCardType(a *Atoms, cardType CardType, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomCardType, CardType: cardType, Span: span})
}

func appendAtomExcludedType(a *Atoms, cardType CardType, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomExcludedType, CardType: cardType, Span: span})
}

func appendAtomSupertype(a *Atoms, supertype Supertype, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomSupertype, Supertype: supertype, Span: span})
}

func appendAtomSubtype(a *Atoms, identity types.Sub, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomSubtype, Subtype: identity, Span: span})
}

func appendAtomObjectNoun(a *Atoms, noun ObjectNoun, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomObjectNoun, ObjectNoun: noun, Span: span})
}

func appendAtomZone(a *Atoms, zoneValue zone.Type, role ZoneRole, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomZone, Zone: zoneValue, ZoneRole: role, Span: span})
}

func appendAtomCounter(a *Atoms, kind counter.Kind, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomCounter, Counter: kind, Span: span})
}

func appendAtomCardinal(a *Atoms, value int, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomCardinal, Value: value, Span: span})
}

func appendAtomOrdinal(a *Atoms, value int, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomOrdinal, Value: value, Span: span})
}

func appendAtomSelectionFlag(a *Atoms, flag SelectionFlag, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomSelectionFlag, Flag: flag, Span: span})
}

func appendAtomController(a *Atoms, relation ControllerRelation, span shared.Span) {
	a.semantic = append(a.semantic, semanticAtom{Kind: atomController, Controller: relation, Span: span})
}

// ColorAt returns the color atom that begins at span, when present.
func (a Atoms) ColorAt(span shared.Span) (Color, bool) {
	for _, atom := range a.Colors() {
		if spanEquals(atom.Span, span) {
			return atom.Color, true
		}
	}
	return ColorUnknown, false
}

// ExcludedColorAt returns the non-color atom at span, when present.
func (a Atoms) ExcludedColorAt(span shared.Span) (Color, bool) {
	for _, atom := range a.ExcludedColors() {
		if spanEquals(atom.Span, span) {
			return atom.Color, true
		}
	}
	return ColorUnknown, false
}

// ColorQualifierAt returns the color qualifier atom that begins at span.
func (a Atoms) ColorQualifierAt(span shared.Span) (ColorQualifier, bool) {
	for _, atom := range a.ColorQualifiers() {
		if spanEquals(atom.Span, span) {
			return atom.Qualifier, true
		}
	}
	return ColorQualifierUnknown, false
}

// CardTypeAt returns the card-type atom that begins at span, when present.
func (a Atoms) CardTypeAt(span shared.Span) (CardType, bool) {
	for _, atom := range a.CardTypes() {
		if spanEquals(atom.Span, span) {
			return atom.Type, true
		}
	}
	return CardTypeUnknown, false
}

// ExcludedCardTypeAt returns the non-card-type atom at span, when present.
func (a Atoms) ExcludedCardTypeAt(span shared.Span) (CardType, bool) {
	for _, atom := range a.ExcludedTypes() {
		if spanEquals(atom.Span, span) {
			return atom.Type, true
		}
	}
	return CardTypeUnknown, false
}

// SupertypeAt returns the supertype atom that begins at span, when present.
func (a Atoms) SupertypeAt(span shared.Span) (Supertype, bool) {
	for _, atom := range a.Supertypes() {
		if spanEquals(atom.Span, span) {
			return atom.Supertype, true
		}
	}
	return SupertypeUnknown, false
}

// SubtypeAt returns the subtype identity atom that begins at span, when present.
func (a Atoms) SubtypeAt(span shared.Span) (types.Sub, bool) {
	for _, atom := range a.Subtypes() {
		if spanEquals(atom.Span, span) {
			return atom.Identity, true
		}
	}
	return "", false
}

// SubtypesIn returns subtype identity atoms whose spans fall within span.
func (a Atoms) SubtypesIn(span shared.Span) []types.Sub {
	var result []types.Sub
	for _, atom := range a.Subtypes() {
		if spanCovers(span, atom.Span) && !slices.Contains(result, atom.Identity) {
			result = append(result, atom.Identity)
		}
	}
	return result
}

// ObjectNounAt returns the object-noun atom that begins at span, when present.
func (a Atoms) ObjectNounAt(span shared.Span) (ObjectNoun, bool) {
	for _, atom := range a.ObjectNouns() {
		if spanEquals(atom.Span, span) {
			return atom.Noun, true
		}
	}
	return ObjectNounUnknown, false
}

// CardinalAt returns the cardinal atom that begins at span, when present.
func (a Atoms) CardinalAt(span shared.Span) (int, bool) {
	for _, atom := range a.Cardinals() {
		if spanEquals(atom.Span, span) {
			return atom.Value, true
		}
	}
	return 0, false
}

// OrdinalAt returns the ordinal atom that begins at span, when present.
func (a Atoms) OrdinalAt(span shared.Span) (int, bool) {
	for _, atom := range a.Ordinals() {
		if spanEquals(atom.Span, span) {
			return atom.Value, true
		}
	}
	return 0, false
}

// SelectionFlagIn reports whether a selector flag exists within span.
func (a Atoms) SelectionFlagIn(span shared.Span, flag SelectionFlag) bool {
	for _, atom := range a.SelectionFlags() {
		if atom.Flag == flag && spanCovers(span, atom.Span) {
			return true
		}
	}
	return false
}

// ControllerIn returns the first controller relation inside span.
func (a Atoms) ControllerIn(span shared.Span) (ControllerRelation, bool) {
	for _, atom := range a.Controllers() {
		if spanCovers(span, atom.Span) {
			return atom.Relation, true
		}
	}
	return ControllerRelationUnknown, false
}

// ZoneIn returns the first zone atom with the given role whose span falls within
// span, in source order.
func (a Atoms) ZoneIn(span shared.Span, role ZoneRole) (zone.Type, bool) {
	for _, atom := range a.Zones() {
		if atom.Role == role && spanCovers(span, atom.Span) {
			return atom.Zone, true
		}
	}
	return zone.None, false
}

// CounterIn returns the counter atom whose name span falls within span, with the
// span it covers, when present.
func (a Atoms) CounterIn(span shared.Span) (counter.Kind, shared.Span, bool) {
	for _, atom := range a.Counters() {
		if spanCovers(span, atom.Span) {
			return atom.Kind, atom.Span, true
		}
	}
	return 0, shared.Span{}, false
}

// ReferencesIn returns the explicit references whose span falls within span.
func (a Atoms) ReferencesIn(span shared.Span) []Reference {
	var result []Reference
	for _, reference := range a.references {
		if spanCovers(span, reference.Span) {
			result = append(result, reference)
		}
	}
	return result
}

// ReferenceIDAt returns the NodeID of the reference whose span exactly equals
// span, or -1 when no reference fills that span. It lets the parser record a
// typed identity link to a reference so downstream stages need not match the
// reference by comparing source spans.
func (a Atoms) ReferenceIDAt(span shared.Span) int {
	for _, reference := range a.references {
		if reference.Span == span {
			return reference.NodeID
		}
	}
	return -1
}

// PronounAt returns the typed pronoun reference beginning at span.
func (a Atoms) PronounAt(span shared.Span) (PronounKind, bool) {
	for _, reference := range a.references {
		if reference.Kind == ReferencePronoun && spanStartsAt(reference.Span, span) {
			return reference.Pronoun, true
		}
	}
	return PronounUnknown, false
}

// SelfNameAt reports whether any occurrence of the card's own name covers the
// token at the given span.
func (a Atoms) SelfNameAt(span shared.Span) bool {
	for _, name := range a.selfNameSpans {
		if spanCovers(name, span) {
			return true
		}
	}
	return false
}

// SelfNameStartingAt reports whether an occurrence of the card's own name begins
// at span.
func (a Atoms) SelfNameStartingAt(span shared.Span) bool {
	_, ok := a.SelfNameSpanStartingAt(span)
	return ok
}

// SelfNameSpanStartingAt returns the span of the card-name occurrence that
// begins at span, when present.
func (a Atoms) SelfNameSpanStartingAt(span shared.Span) (shared.Span, bool) {
	for _, name := range a.selfNameSpans {
		if spanStartsAt(name, span) {
			return name, true
		}
	}
	return shared.Span{}, false
}

// SourceNameSpanStartingAt returns the trigger-source alias span beginning at
// span, when present.
func (a Atoms) SourceNameSpanStartingAt(span shared.Span) (shared.Span, bool) {
	for _, name := range a.sourceNameSpans {
		if spanStartsAt(name, span) {
			return name, true
		}
	}
	return shared.Span{}, false
}

// SourceMarkerSpanStartingAt returns the trigger-source marker span beginning at
// span, when present.
func (a Atoms) SourceMarkerSpanStartingAt(span shared.Span) (shared.Span, bool) {
	for _, marker := range a.sourceMarkerSpans {
		if spanStartsAt(marker, span) {
			return marker, true
		}
	}
	return shared.Span{}, false
}

// ReferencesWithin returns the explicit references whose first token begins at
// the start of one of the supplied tokens, preserving recognition order. It lets
// callers consume references over a non-contiguous token selection (for example
// after an activation-timing clause has been removed) without re-recognizing
// spelling.
func (a Atoms) ReferencesWithin(tokens []shared.Token) []Reference {
	starts := make(map[int]struct{}, len(tokens))
	for _, token := range tokens {
		starts[token.Span.Start.Offset] = struct{}{}
	}
	var result []Reference
	for _, reference := range a.references {
		if _, ok := starts[reference.Span.Start.Offset]; ok {
			result = append(result, reference)
		}
	}
	return result
}
