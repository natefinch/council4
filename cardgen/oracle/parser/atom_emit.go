package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// This file owns parser emission of the reusable, source-spanned semantic atoms
// shared across grammar families. Parse records every recognized atom with the
// exact source span it covers; downstream stages consume the typed values by
// span and never reinspect Oracle spelling to recover the meanings recognized
// here. Genuine literal identities (game zones, counter kinds, and creature or
// land subtypes) are carried as their canonical typed values.

// ZoneRole records how a recognized zone is introduced: as a movement origin
// ("from"), a movement destination ("to"/"onto"/"into"/"on top of"), or plainly.
type ZoneRole uint8

// Zone roles recognized by the parser.
const (
	ZoneRolePlain ZoneRole = iota
	ZoneRoleFrom
	ZoneRoleTo
)

// ColorAtom is a source-spanned typed Oracle color.
type ColorAtom struct {
	Color Color
	Span  shared.Span
}

// ColorQualifierAtom is a source-spanned color-family qualifier.
type ColorQualifierAtom struct {
	Qualifier ColorQualifier
	Span      shared.Span
}

// CardTypeAtom is a source-spanned typed Oracle card type.
type CardTypeAtom struct {
	Type CardType
	Span shared.Span
}

// SupertypeAtom is a source-spanned typed Oracle supertype.
type SupertypeAtom struct {
	Supertype Supertype
	Span      shared.Span
}

// SubtypeAtom is a source-spanned creature or land subtype identity. Identity is
// the canonical types.Sub value; the parser owns the spelling and plural
// normalization that resolve it.
type SubtypeAtom struct {
	Identity types.Sub
	Span     shared.Span
}

// ObjectNounAtom is a source-spanned typed Oracle object noun.
type ObjectNounAtom struct {
	Noun ObjectNoun
	Span shared.Span
}

// ZoneAtom is a source-spanned game zone with the role its introducing wording
// gives it.
type ZoneAtom struct {
	Zone zone.Type
	Role ZoneRole
	Span shared.Span
}

// CounterAtom is a source-spanned counter kind. Span covers the counter-kind
// name tokens preceding the "counter(s)" noun.
type CounterAtom struct {
	Kind counter.Kind
	Span shared.Span
}

// CardinalAtom is a source-spanned small-cardinal number word and its value.
type CardinalAtom struct {
	Value int
	Span  shared.Span
}

// OrdinalAtom is a source-spanned ordinal number word and its value.
type OrdinalAtom struct {
	Value int
	Span  shared.Span
}

// SelectionFlagAtom is a source-spanned selector modifier.
type SelectionFlagAtom struct {
	Flag SelectionFlag
	Span shared.Span
}

// ControllerRelationAtom is a source-spanned control/ownership relation.
type ControllerRelationAtom struct {
	Relation ControllerRelation
	Span     shared.Span
}

type atomKind uint8

const (
	atomColor atomKind = iota + 1
	atomExcludedColor
	atomColorQualifier
	atomCardType
	atomExcludedType
	atomSupertype
	atomSubtype
	atomObjectNoun
	atomZone
	atomCounter
	atomCardinal
	atomOrdinal
	atomSelectionFlag
	atomController
)

type semanticAtom struct {
	Span       shared.Span
	Kind       atomKind
	Color      Color
	Qualifier  ColorQualifier
	CardType   CardType
	Supertype  Supertype
	Subtype    types.Sub
	ObjectNoun ObjectNoun
	Zone       zone.Type
	ZoneRole   ZoneRole
	Counter    counter.Kind
	Value      int
	Flag       SelectionFlag
	Controller ControllerRelation
}

// Atoms is the collection of source-spanned typed atoms recognized within one
// syntax node. Downstream stages look atoms up by the span of the tokens they
// are examining rather than by re-recognizing spelling. The zero value contains
// no atoms and therefore fails closed.
type Atoms struct {
	semantic   []semanticAtom
	references []Reference
	// SelfNameSpans records every occurrence of the card's own name, including
	// possessive forms. It is the parser-owned source of self-name recognition
	// so the compiler need not inspect name spelling.
	selfNameSpans []shared.Span
	// SourceNameSpans records source-subject aliases accepted by trigger grammar
	// (full name, short comma name, DFC front name, and first-word legend names).
	// They are not explicit references and are consumed only by exact trigger
	// subject productions.
	sourceNameSpans []shared.Span
	// SourceMarkerSpans records "this <source marker>" subject phrases accepted
	// only by exact trigger subject productions. They are distinct from explicit
	// references so unsupported effect and cost references still fail closed.
	sourceMarkerSpans []shared.Span
}

func spanCovers(outer, inner shared.Span) bool {
	return inner.Start.Offset >= outer.Start.Offset && inner.End.Offset <= outer.End.Offset
}

func spanStartsAt(atom, target shared.Span) bool {
	return atom.Start.Offset == target.Start.Offset
}

func spanEquals(left, right shared.Span) bool {
	return left == right
}

// AtomOption adds typed parser atoms to a manually constructed Atoms value.
type AtomOption func(*Atoms)

// NewAtoms constructs an Atoms value from typed parser atoms. Omitted atom
// families stay empty, preserving the zero value's fail-closed behavior.
func NewAtoms(options ...AtomOption) Atoms {
	var atoms Atoms
	for _, option := range options {
		option(&atoms)
	}
	return atoms
}

// WithZones adds zone atoms to NewAtoms.
func WithZones(zones ...ZoneAtom) AtomOption {
	return func(atoms *Atoms) {
		for _, atom := range zones {
			appendAtomZone(atoms, atom.Zone, atom.Role, atom.Span)
		}
	}
}

// WithCounters adds counter atoms to NewAtoms.
func WithCounters(counters ...CounterAtom) AtomOption {
	return func(atoms *Atoms) {
		for _, atom := range counters {
			appendAtomCounter(atoms, atom.Kind, atom.Span)
		}
	}
}

// WithCardinals adds cardinal atoms to NewAtoms.
func WithCardinals(cardinals ...CardinalAtom) AtomOption {
	return func(atoms *Atoms) {
		for _, atom := range cardinals {
			appendAtomCardinal(atoms, atom.Value, atom.Span)
		}
	}
}

// WithSubtypes adds subtype atoms to NewAtoms.
func WithSubtypes(subtypes ...SubtypeAtom) AtomOption {
	return func(atoms *Atoms) {
		for _, atom := range subtypes {
			appendAtomSubtype(atoms, atom.Identity, atom.Span)
		}
	}
}

// WithObjectNouns adds object-noun atoms to NewAtoms.
func WithObjectNouns(nouns ...ObjectNounAtom) AtomOption {
	return func(atoms *Atoms) {
		for _, atom := range nouns {
			appendAtomObjectNoun(atoms, atom.Noun, atom.Span)
		}
	}
}

// WithSelectionFlags adds selection-flag atoms to NewAtoms.
func WithSelectionFlags(flags ...SelectionFlagAtom) AtomOption {
	return func(atoms *Atoms) {
		for _, atom := range flags {
			appendAtomSelectionFlag(atoms, atom.Flag, atom.Span)
		}
	}
}

// WithReferences adds explicit reference atoms to NewAtoms.
func WithReferences(references ...Reference) AtomOption {
	return func(atoms *Atoms) {
		atoms.references = append(atoms.references, references...)
	}
}

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

// collectAtoms recognizes every reusable atom within the semantic tokens of a
// syntax node and returns them with their source spans. Reminder and quoted
// spans are excluded so that recognized meaning matches the semantic tokens the
// compiler consumes.
func collectAtoms(tokens []shared.Token, reminders, quoted []Delimited, cardName string) Atoms {
	tokens = atomSemanticTokens(tokens, reminders, quoted)
	atoms := Atoms{
		references:        collectReferences(tokens, cardName),
		selfNameSpans:     collectSelfNameSpans(tokens, cardName),
		sourceNameSpans:   collectSourceNameSpans(tokens, cardName),
		sourceMarkerSpans: collectSourceMarkerSpans(tokens),
	}
	for _, token := range tokens {
		if token.Kind != shared.Word {
			continue
		}
		if color, ok := recognizeColorWord(token.Text); ok {
			appendAtomColor(&atoms, color, token.Span)
		}
		if rest, ok := strings.CutPrefix(strings.ToLower(token.Text), "non"); ok {
			if color, colorOK := recognizeColorWord(rest); colorOK {
				appendAtomExcludedColor(&atoms, color, token.Span)
			}
			if cardType, typeOK := recognizeCardTypeWord(rest); typeOK {
				appendAtomExcludedType(&atoms, cardType, token.Span)
			}
		}
		if qualifier, ok := recognizeColorQualifierWord(token.Text); ok {
			appendAtomColorQualifier(&atoms, qualifier, token.Span)
		}
		if cardType, ok := recognizeCardTypeWord(token.Text); ok {
			appendAtomCardType(&atoms, cardType, token.Span)
		}
		if supertype, ok := recognizeSupertypeWord(token.Text); ok {
			appendAtomSupertype(&atoms, supertype, token.Span)
		}
		if noun, ok := recognizeObjectNoun(token); ok {
			appendAtomObjectNoun(&atoms, noun, token.Span)
		}
		if value, ok := CardinalWordValue(token.Text); ok {
			appendAtomCardinal(&atoms, value, token.Span)
		}
		if value, ok := OrdinalWordValue(token.Text); ok {
			appendAtomOrdinal(&atoms, value, token.Span)
		}
		if flag, ok := recognizeSelectionFlag(token.Text); ok {
			appendAtomSelectionFlag(&atoms, flag, token.Span)
		}
	}
	for _, atom := range scanSubtypes(tokens) {
		appendAtomSubtype(&atoms, atom.Identity, atom.Span)
	}
	for _, atom := range scanControllerRelations(tokens) {
		appendAtomController(&atoms, atom.Relation, atom.Span)
	}
	for _, atom := range scanZones(tokens) {
		appendAtomZone(&atoms, atom.Zone, atom.Role, atom.Span)
	}
	for _, atom := range scanCounters(tokens) {
		appendAtomCounter(&atoms, atom.Kind, atom.Span)
	}
	return atoms
}

func atomSemanticTokens(tokens []shared.Token, reminders, quoted []Delimited) []shared.Token {
	if len(reminders) == 0 && len(quoted) == 0 {
		return tokens
	}
	excluded := append(append([]Delimited(nil), reminders...), quoted...)
	result := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		skip := false
		for _, delimiter := range excluded {
			if token.Span.Start.Offset >= delimiter.Span.Start.Offset &&
				token.Span.End.Offset <= delimiter.Span.End.Offset {
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, token)
		}
	}
	return result
}

// scanZones recognizes origin and destination zone phrases and emits a zone atom
// for each occurrence in source order.
func scanZones(tokens []shared.Token) []ZoneAtom {
	var atoms []ZoneAtom
	for i := range tokens {
		switch {
		case equalWord(tokens[i], "from") && i+1 < len(tokens):
			if zoneValue, ok := zonePhrase(tokens[i+1:]); ok {
				atoms = appendZone(atoms, tokens, i, zoneValue, ZoneRoleFrom)
			}
		case (equalWord(tokens[i], "to") || equalWord(tokens[i], "into") || equalWord(tokens[i], "onto")) && i+1 < len(tokens):
			if zoneValue, ok := zonePhrase(tokens[i+1:]); ok {
				atoms = appendZone(atoms, tokens, i, zoneValue, ZoneRoleTo)
			}
		case equalWord(tokens[i], "on") && i+3 < len(tokens) &&
			(equalWord(tokens[i+1], "top") || equalWord(tokens[i+1], "bottom")) &&
			equalWord(tokens[i+2], "of"):
			if zoneValue, ok := zonePhrase(tokens[i+3:]); ok {
				atoms = appendZone(atoms, tokens, i, zoneValue, ZoneRoleTo)
			}
		case equalWord(tokens[i], "on") && i+4 < len(tokens) &&
			equalWord(tokens[i+1], "the") &&
			(equalWord(tokens[i+2], "top") || equalWord(tokens[i+2], "bottom")) &&
			equalWord(tokens[i+3], "of"):
			if zoneValue, ok := zonePhrase(tokens[i+4:]); ok {
				atoms = appendZone(atoms, tokens, i, zoneValue, ZoneRoleTo)
			}
		default:
		}
	}
	return atoms
}

func zonePhrase(tokens []shared.Token) (zone.Type, bool) {
	switch {
	case graveyardZonePhrase(tokens):
		return zone.Graveyard, true
	case battlefieldZonePhrase(tokens):
		return zone.Battlefield, true
	case handZonePhrase(tokens):
		return zone.Hand, true
	case libraryZonePhrase(tokens):
		return zone.Library, true
	case exileZonePhrase(tokens):
		return zone.Exile, true
	default:
		return zone.None, false
	}
}

func appendZone(atoms []ZoneAtom, tokens []shared.Token, i int, value zone.Type, role ZoneRole) []ZoneAtom {
	return append(atoms, ZoneAtom{
		Zone: value,
		Role: role,
		Span: tokens[i].Span,
	})
}

// counterKindNames lists the counter kinds the parser recognizes by name, in the
// priority order the compiler historically matched them.
var counterKindNames = []counter.Kind{
	counter.PlusOnePlusOne,
	counter.MinusOneMinusOne,
	counter.Loyalty,
	counter.Charge,
	counter.Time,
	counter.Defense,
	counter.Poison,
	counter.Lore,
	counter.Verse,
	counter.Shield,
	counter.Stun,
	counter.Finality,
	counter.Brick,
	counter.Page,
	counter.Enlightened,
	counter.Oil,
	counter.Blood,
	counter.Indestructible,
	counter.Deathtouch,
	counter.Flying,
	counter.FirstStrike,
	counter.Hexproof,
	counter.Lifelink,
	counter.Menace,
	counter.Reach,
	counter.Trample,
	counter.Vigilance,
	counter.Energy,
	counter.Experience,
}

// scanCounters emits a counter atom for each "<kind> counter(s)" phrase, spanning
// the kind-name tokens that immediately precede the counter noun.
func scanCounters(tokens []shared.Token) []CounterAtom {
	var atoms []CounterAtom
	for i := range tokens {
		if !equalWord(tokens[i], "counter") && !equalWord(tokens[i], "counters") {
			continue
		}
		if kind, span, ok := counterNameBefore(tokens, i); ok {
			atoms = append(atoms, CounterAtom{Kind: kind, Span: span})
		}
	}
	return atoms
}

func counterNameBefore(tokens []shared.Token, counterIndex int) (counter.Kind, shared.Span, bool) {
	for start := counterIndex - 1; start >= 0; start-- {
		name := strings.ToLower(joinTokens(tokens[start:counterIndex]))
		if kind, ok := counterKindAlias(name); ok {
			return kind, shared.SpanOf(tokens[start:counterIndex]), true
		}
		for _, kind := range counterKindNames {
			if name == kind.String() {
				return kind, shared.SpanOf(tokens[start:counterIndex]), true
			}
		}
	}
	return 0, shared.Span{}, false
}

func counterKindAlias(name string) (counter.Kind, bool) {
	switch name {
	case "storage", "fuse":
		return counter.Charge, true
	default:
		return 0, false
	}
}

func scanControllerRelations(tokens []shared.Token) []ControllerRelationAtom {
	patterns := []struct {
		words    []string
		relation ControllerRelation
	}{
		{[]string{"you", "control"}, ControllerRelationYouControl},
		{[]string{"you", "don't", "control"}, ControllerRelationYouDontControl},
		{[]string{"an", "opponent", "controls"}, ControllerRelationOpponentControls},
		{[]string{"your", "opponents", "control"}, ControllerRelationOpponentControls},
		{[]string{"you", "own"}, ControllerRelationYouOwn},
		{[]string{"an", "opponent", "owns"}, ControllerRelationOpponentOwns},
	}
	var atoms []ControllerRelationAtom
	for i := range tokens {
		for _, pattern := range patterns {
			if atomWordsAt(tokens, i, pattern.words...) {
				atoms = append(atoms, ControllerRelationAtom{Relation: pattern.relation, Span: shared.SpanOf(tokens[i : i+len(pattern.words)])})
			}
		}
	}
	return atoms
}

func allWordTokens(tokens []shared.Token) bool {
	for _, token := range tokens {
		if token.Kind != shared.Word {
			return false
		}
	}
	return len(tokens) > 0
}

func atomWordsAt(tokens []shared.Token, start int, words ...string) bool {
	if start < 0 || start+len(words) > len(tokens) {
		return false
	}
	for i, word := range words {
		if !equalWord(tokens[start+i], word) {
			return false
		}
	}
	return true
}

func scanSubtypes(tokens []shared.Token) []SubtypeAtom {
	var atoms []SubtypeAtom
	used := make([]bool, len(tokens))
	for width := 3; width >= 1; width-- {
		for i := 0; i+width <= len(tokens); i++ {
			if slices.Contains(used[i:i+width], true) || !allWordTokens(tokens[i:i+width]) {
				continue
			}
			if identity, ok := recognizeSubtypePhrase(joinTokens(tokens[i : i+width])); ok {
				atoms = append(atoms, SubtypeAtom{Identity: identity, Span: shared.SpanOf(tokens[i : i+width])})
				for j := i; j < i+width; j++ {
					used[j] = true
				}
			}
		}
	}
	slices.SortFunc(atoms, func(a, b SubtypeAtom) int {
		return a.Span.Start.Offset - b.Span.Start.Offset
	})
	return atoms
}

var subtypeCardFamilies = []types.Card{
	types.Artifact,
	types.Battle,
	types.Creature,
	types.Enchantment,
	types.Instant,
	types.Kindred,
	types.Land,
	types.Planeswalker,
	types.Sorcery,
	types.Plane,
	types.Dungeon,
}

// recognizeSubtypePhrase resolves an Oracle subtype phrase to its canonical
// typed identity, owning capitalization, multiword, and plural normalization.
func recognizeSubtypePhrase(phrase string) (types.Sub, bool) {
	phrase = strings.TrimSpace(phrase)
	if phrase == "" {
		return "", false
	}
	candidates := subtypeIdentityCandidates(phrase)
	for _, candidate := range candidates {
		sub := types.Sub(candidate)
		for _, cardType := range subtypeCardFamilies {
			if types.KnownSubtypeForType(cardType, sub) {
				return sub, true
			}
		}
	}
	return "", false
}

func subtypeIdentityCandidates(phrase string) []string {
	lower := strings.ToLower(phrase)
	switch lower {
	case "children":
		return []string{string(types.Child)}
	case "mice":
		return []string{string(types.Mouse)}
	}
	seen := map[string]struct{}{}
	var candidates []string
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if _, ok := seen[candidate]; ok {
			return
		}
		seen[candidate] = struct{}{}
		candidates = append(candidates, candidate)
	}
	add(phrase)
	add(titleCaseWord(phrase))
	if strings.Contains(phrase, " ") {
		hyphenated := strings.ReplaceAll(phrase, " ", "-")
		add(hyphenated)
		add(titleCaseWord(hyphenated))
	}
	words := strings.Fields(phrase)
	if len(words) > 0 {
		last := words[len(words)-1]
		for _, singular := range SingularNounForms(last) {
			if singular == last {
				continue
			}
			candidateWords := append([]string(nil), words...)
			candidateWords[len(candidateWords)-1] = singular
			candidate := strings.Join(candidateWords, " ")
			add(candidate)
			add(titleCaseWord(candidate))
			if strings.Contains(candidate, " ") {
				hyphenated := strings.ReplaceAll(candidate, " ", "-")
				add(hyphenated)
				add(titleCaseWord(hyphenated))
			}
		}
	}
	for _, singular := range SingularNounForms(phrase) {
		if singular != phrase {
			add(singular)
			add(titleCaseWord(singular))
		}
	}
	return candidates
}

func titleCaseWord(word string) string {
	if word == "" {
		return ""
	}
	parts := strings.Fields(word)
	if len(parts) > 1 {
		for i := range parts {
			parts[i] = titleCaseWord(parts[i])
		}
		return strings.Join(parts, " ")
	}
	return strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
}

func joinTokens(tokens []shared.Token) string {
	if len(tokens) == 0 {
		return ""
	}
	var builder strings.Builder
	for i, token := range tokens {
		if i > 0 && atomNeedsSpace(tokens[i-1], token) {
			_ = builder.WriteByte(' ')
		}
		_, _ = builder.WriteString(token.Text)
	}
	return builder.String()
}

// atomNeedsSpace mirrors the compiler's needsSemanticSpace so that the joined
// counter-name text matches the spelling the compiler historically compared.
func atomNeedsSpace(previous, current shared.Token) bool {
	if current.Kind == shared.Comma || current.Kind == shared.Period || current.Kind == shared.Colon ||
		current.Kind == shared.Semicolon || current.Kind == shared.RightParen ||
		previous.Kind == shared.LeftParen || previous.Kind == shared.Quote || current.Kind == shared.Quote {
		return false
	}
	if previous.Kind == shared.Plus || previous.Kind == shared.Minus || previous.Kind == shared.Slash ||
		current.Kind == shared.Slash {
		return false
	}
	return previous.Kind != shared.Symbol && current.Kind != shared.Symbol
}
