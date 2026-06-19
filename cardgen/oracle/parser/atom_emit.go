package parser

import (
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
type ZoneRole string

// Zone roles recognized by the parser.
const (
	ZoneRolePlain ZoneRole = ""
	ZoneRoleFrom  ZoneRole = "ZoneRoleFrom"
	ZoneRoleTo    ZoneRole = "ZoneRoleTo"
)

// ColorAtom is a source-spanned typed Oracle color.
type ColorAtom struct {
	Color Color       `json:",omitempty"`
	Span  shared.Span `json:"-"`
}

// ColorQualifierAtom is a source-spanned color-family qualifier.
type ColorQualifierAtom struct {
	Qualifier ColorQualifier `json:",omitempty"`
	Span      shared.Span    `json:"-"`
}

// CardTypeAtom is a source-spanned typed Oracle card type.
type CardTypeAtom struct {
	Type CardType    `json:",omitempty"`
	Span shared.Span `json:"-"`
}

// SupertypeAtom is a source-spanned typed Oracle supertype.
type SupertypeAtom struct {
	Supertype Supertype   `json:",omitempty"`
	Span      shared.Span `json:"-"`
}

// SubtypeAtom is a source-spanned creature or land subtype identity. Identity is
// the canonical types.Sub value; the parser owns the spelling and plural
// normalization that resolve it.
type SubtypeAtom struct {
	Identity types.Sub   `json:",omitempty"`
	Span     shared.Span `json:"-"`
}

// ObjectNounAtom is a source-spanned typed Oracle object noun.
type ObjectNounAtom struct {
	Noun ObjectNoun  `json:",omitempty"`
	Span shared.Span `json:"-"`
}

// ZoneAtom is a source-spanned game zone with the role its introducing wording
// gives it.
type ZoneAtom struct {
	Zone zone.Type   `json:",omitempty"`
	Role ZoneRole    `json:",omitempty"`
	Span shared.Span `json:"-"`
}

// CounterAtom is a source-spanned counter kind. Span covers the counter-kind
// name tokens preceding the "counter(s)" noun.
type CounterAtom struct {
	Kind counter.Kind `json:",omitempty"`
	Span shared.Span  `json:"-"`
}

// CardinalAtom is a source-spanned small-cardinal number word and its value.
type CardinalAtom struct {
	Value int         `json:",omitempty"`
	Span  shared.Span `json:"-"`
}

// OrdinalAtom is a source-spanned ordinal number word and its value.
type OrdinalAtom struct {
	Value int         `json:",omitempty"`
	Span  shared.Span `json:"-"`
}

// SelectionFlagAtom is a source-spanned selector modifier.
type SelectionFlagAtom struct {
	Flag SelectionFlag `json:",omitempty"`
	Span shared.Span   `json:"-"`
}

// ControllerRelationAtom is a source-spanned control/ownership relation.
type ControllerRelationAtom struct {
	Relation ControllerRelation `json:",omitempty"`
	Span     shared.Span        `json:"-"`
}

type atomKind string

const (
	atomColor             atomKind = "atomColor"
	atomExcludedColor     atomKind = "atomExcludedColor"
	atomColorQualifier    atomKind = "atomColorQualifier"
	atomCardType          atomKind = "atomCardType"
	atomExcludedType      atomKind = "atomExcludedType"
	atomSupertype         atomKind = "atomSupertype"
	atomExcludedSupertype atomKind = "atomExcludedSupertype"
	atomSubtype           atomKind = "atomSubtype"
	atomObjectNoun        atomKind = "atomObjectNoun"
	atomZone              atomKind = "atomZone"
	atomCounter           atomKind = "atomCounter"
	atomCardinal          atomKind = "atomCardinal"
	atomOrdinal           atomKind = "atomOrdinal"
	atomSelectionFlag     atomKind = "atomSelectionFlag"
	atomController        atomKind = "atomController"
)

type semanticAtom struct {
	Span       shared.Span        `json:"-"`
	Kind       atomKind           `json:",omitempty"`
	Color      Color              `json:",omitempty"`
	Qualifier  ColorQualifier     `json:",omitempty"`
	CardType   CardType           `json:",omitempty"`
	Supertype  Supertype          `json:",omitempty"`
	Subtype    types.Sub          `json:",omitempty"`
	ObjectNoun ObjectNoun         `json:",omitempty"`
	Zone       zone.Type          `json:",omitempty"`
	ZoneRole   ZoneRole           `json:",omitempty"`
	Counter    counter.Kind       `json:",omitempty"`
	Value      int                `json:",omitempty"`
	Flag       SelectionFlag      `json:",omitempty"`
	Controller ControllerRelation `json:",omitempty"`
}

// Atoms is the collection of source-spanned typed atoms recognized within one
// syntax node. Downstream stages look atoms up by the span of the tokens they
// are examining rather than by re-recognizing spelling. The zero value contains
// no atoms and therefore fails closed.
type Atoms struct {
	semantic         []semanticAtom
	references       []Reference
	keywords         []Keyword
	keywordSelectors []KeywordSelector
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

// WithKeywords adds typed keywords to NewAtoms.
func WithKeywords(keywords ...Keyword) AtomOption {
	return func(atoms *Atoms) {
		atoms.keywords = append(atoms.keywords, keywords...)
	}
}

// WithKeywordSelectors adds typed keyword selectors to NewAtoms.
func WithKeywordSelectors(selectors ...KeywordSelector) AtomOption {
	return func(atoms *Atoms) {
		atoms.keywordSelectors = append(atoms.keywordSelectors, selectors...)
	}
}
