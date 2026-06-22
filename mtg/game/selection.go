package game

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SubtypeChoiceSource identifies where a Selection's required creature subtype
// is chosen during play, for predicates that match permanents of a subtype
// decided in-game rather than printed in the predicate. The zero value imposes
// no chosen-subtype restriction. The two sources are mutually exclusive, so a
// single field captures both at one byte, packing into Selection's bool cluster.
type SubtypeChoiceSource uint8

// SubtypeChoiceSource values name the supported in-game subtype sources.
const (
	// SubtypeChoiceNone imposes no chosen-subtype restriction.
	SubtypeChoiceNone SubtypeChoiceSource = iota

	// SubtypeChoiceSourceEntry requires the matched permanent to share the
	// creature subtype the predicate's source permanent chose as it entered (its
	// EntryChoices[EntryTypeChoiceKey]), the "of the chosen type" restriction of
	// chosen-type anthems. A missing source, choice, or subtype matches nothing.
	SubtypeChoiceSourceEntry

	// SubtypeChoiceResolution requires the matched permanent to share the creature
	// subtype published under SpellChosenTypeChoiceKey by an earlier Choose
	// instruction in the same resolution (its StackObject.ResolutionChoices), the
	// "of that type" restriction of "Choose a creature type. ... of that type."
	// spells (Distant Melody). A missing or non-subtype choice matches nothing.
	SubtypeChoiceResolution

	// SubtypeChoiceResolutionExcluded requires the matched permanent to NOT share
	// the creature subtype published under SpellChosenTypeChoiceKey by an earlier
	// Choose instruction in the same resolution, the "aren't of the chosen type"
	// restriction of "Choose a creature type. Destroy all creatures that aren't of
	// the chosen type." spells (Kindred Dominance). A missing or non-subtype choice
	// matches nothing, failing closed like its positive sibling.
	SubtypeChoiceResolutionExcluded
)

// ColorChoiceSource identifies where a Selection's required color is chosen
// during play, for predicates that match permanents of a color decided in-game
// rather than printed in the predicate. The zero value imposes no chosen-color
// restriction. A single byte captures the source, packing into Selection's bool
// cluster.
type ColorChoiceSource uint8

// ColorChoiceSource values name the supported in-game color sources.
const (
	// ColorChoiceNone imposes no chosen-color restriction.
	ColorChoiceNone ColorChoiceSource = iota

	// ColorChoiceSourceEntry requires the matched permanent to share the color
	// the predicate's source permanent chose as it entered (its
	// EntryChoices[EntryColorChoiceKey]), the "of the chosen color" restriction
	// of chosen-color anthems (Heraldic Banner). A missing source, choice, or
	// color matches nothing.
	ColorChoiceSourceEntry
)

// SubtypeChoiceWithoutEntry returns choice with the source-entry variant cleared
// to SubtypeChoiceNone, leaving any other chosen-subtype source intact. Trigger
// subject validators use it to permit the entry-choice predicate (whose subtype
// the entering source supplies) while keeping other in-game subtype sources
// failing closed in contexts where they are unavailable.
func SubtypeChoiceWithoutEntry(choice SubtypeChoiceSource) SubtypeChoiceSource {
	if choice == SubtypeChoiceSourceEntry {
		return SubtypeChoiceNone
	}
	return choice
}

// Selection is pure rules data describing WHICH game objects share a predicate.
// It is the single, valence-agnostic matcher description that subsumes the
// characteristic fields formerly duplicated across TargetPredicate, the
// condition controls-matching filter, the permanent/card filters of
// TriggerPattern, and the historical mass-effect selector constants.
//
// Selection describes WHAT matches, never where candidates come from. Counting,
// total power, and candidate-domain concerns (controlled, defending, equipped,
// all permanents) stay outside Selection and belong with the valence-specific
// references that will later own runtime binding. The zero value is a wildcard
// that matches anything.
type Selection struct {
	// AnyOf requires the subject to match at least one alternative selection.
	// Fields on this selection remain common conjunctive requirements.
	AnyOf []Selection

	// RequiredTypes lists card types that must all be present (conjunctive, an
	// "artifact creature" type line). RequiredTypesAny matches when any listed
	// card type is present (disjunctive, "creature or artifact"). ExcludedTypes
	// lists card types that must all be absent.
	RequiredTypes    []types.Card
	RequiredTypesAny []types.Card
	ExcludedTypes    []types.Card

	// Supertypes must all be present. ExcludedSupertype, when non-empty, names a
	// single supertype that must be absent (a "nonbasic" / "nonlegendary" /
	// "nonsnow" filter). One scalar suffices because no canonical Oracle card
	// excludes more than one supertype. SubtypesAny matches when any listed
	// subtype is present.
	Supertypes        []types.Super
	ExcludedSupertype types.Super
	SubtypesAny       []types.Sub

	// ExcludedSubtype names a creature subtype that must be absent, the
	// "non-<subtype>" filter ("non-Human creatures you control"). It parallels
	// ExcludedSupertype: a matched object carrying this subtype fails the
	// selection. The empty value means no exclusion.
	ExcludedSubtype types.Sub

	// ColorsAny matches when any listed color is present. ExcludedColors must
	// all be absent. Colorless requires no colors; Multicolored requires at
	// least two colors.
	ColorsAny      []color.Color
	ExcludedColors []color.Color
	Colorless      bool
	Multicolored   bool

	// ExcludeSource drops the predicate's own source object from the match, for
	// "another" target restrictions and "other ..." mass effects.
	ExcludeSource bool

	// NonToken requires the matched object to not be a token. TokenOnly requires
	// the matched object to be a token.
	NonToken  bool
	TokenOnly bool

	// MatchCounter, when true, requires the matched permanent to carry at least
	// one counter of RequiredCounter's kind ("creature you control with a +1/+1
	// counter on it"). A non-battlefield subject (a card or spell, which has no
	// counters) never matches. A bool flag distinguishes "no counter requirement"
	// from "requires a +1/+1 counter" because counter.Kind's zero value names the
	// +1/+1 counter.
	MatchCounter bool

	// MatchAnyCounter, when true, requires the matched permanent to carry at
	// least one counter of any kind ("if this permanent has counters on it").
	// Unlike MatchCounter it is kind-agnostic. A non-battlefield subject never
	// matches. Placed beside MatchCounter to pack into the bool cluster.
	MatchAnyCounter bool

	// MatchModified, when true, requires the matched permanent to be modified: it
	// carries one or more counters, or has one or more Auras or Equipment
	// attached to it ("modified creatures you control"). A non-battlefield
	// subject never matches.
	MatchModified bool

	// MatchCommander, when true, requires the matched permanent to be a commander
	// ("commander creatures you control"). A non-battlefield subject never
	// matches. Placed beside MatchModified to pack into the bool cluster.
	MatchCommander bool

	// SubtypeChoice constrains the matched permanent to a creature subtype chosen
	// during play; see SubtypeChoiceSource. The zero value imposes no restriction.
	SubtypeChoice SubtypeChoiceSource

	// Controller constrains a permanent by its controller relative to the
	// viewing player. Player constrains a player relative to the viewing player.
	Controller ControllerRelation
	Player     PlayerRelation

	// Tapped constrains tapped state; CombatState constrains combat involvement.
	Tapped      TriState
	CombatState CombatStateFilter

	// Keyword must be present; ExcludedKeyword must be absent.
	Keyword         Keyword
	ExcludedKeyword Keyword

	// ManaValue, Power, and Toughness compare numeric characteristics.
	ManaValue opt.V[compare.Int]
	Power     opt.V[compare.Int]
	Toughness opt.V[compare.Int]

	// RequiredCounter names the counter kind required when MatchCounter is set.
	RequiredCounter counter.Kind

	// RequiredCounterCount compares the number of counters of RequiredCounter's
	// kind the matched permanent carries ("as long as ~ has seven or more quest
	// counters on it"). When present it imposes the comparison independently of
	// MatchCounter and names its kind through RequiredCounter. A non-battlefield
	// subject, which has no counters, never matches.
	RequiredCounterCount opt.V[compare.Int]

	// EnteredThisTurn requires the matched permanent to have entered the
	// battlefield this turn ("each green creature that entered this turn"). A
	// non-battlefield subject never matches. Placed at the end of the struct so
	// the bool joins no existing cluster's documented packing.
	EnteredThisTurn bool

	// ColorChoice constrains the matched permanent to a color chosen during
	// play; see ColorChoiceSource. The zero value imposes no restriction. It
	// backs the "of the chosen color" group filter of chosen-color anthems
	// (Heraldic Banner), reading the source permanent's entry-time color choice.
	ColorChoice ColorChoiceSource

	// PowerLessThanSource and PowerGreaterThanSource require the matched
	// permanent's power to be strictly less / greater than the predicate's source
	// permanent's power ("target attacking creature with lesser power", Mentor).
	// They are source-relative, so a subject with no source or no power never
	// matches. Placed at the end so the bools join no existing cluster's packing.
	PowerLessThanSource    bool
	PowerGreaterThanSource bool

	// Name, when non-empty, requires the matched object's name to equal it,
	// modeling a "card named <Name>" filter (Daru Cavalier, Trustworthy Scout's
	// library searches). It composes with the other fields but in practice stands
	// alone on a plain "card named X" effect. A subject whose name is unavailable
	// never matches. Placed at the end so the field joins no existing cluster.
	Name string
}

// Empty reports whether the Selection carries no active predicate and therefore
// matches anything.
func (s Selection) Empty() bool {
	return len(s.AnyOf) == 0 &&
		len(s.RequiredTypes) == 0 &&
		len(s.RequiredTypesAny) == 0 &&
		len(s.ExcludedTypes) == 0 &&
		len(s.Supertypes) == 0 &&
		s.ExcludedSupertype == "" &&
		len(s.SubtypesAny) == 0 &&
		s.ExcludedSubtype == "" &&
		s.SubtypeChoice == SubtypeChoiceNone &&
		len(s.ColorsAny) == 0 &&
		len(s.ExcludedColors) == 0 &&
		!s.Colorless &&
		!s.Multicolored &&
		s.Controller == ControllerAny &&
		s.Player == PlayerAny &&
		s.Tapped == TriAny &&
		s.CombatState == CombatStateAny &&
		s.Keyword == KeywordNone &&
		s.ExcludedKeyword == KeywordNone &&
		!s.ManaValue.Exists &&
		!s.Power.Exists &&
		!s.Toughness.Exists &&
		!s.MatchCounter &&
		!s.MatchAnyCounter &&
		!s.RequiredCounterCount.Exists &&
		!s.EnteredThisTurn &&
		!s.MatchModified &&
		!s.MatchCommander &&
		s.ColorChoice == ColorChoiceNone &&
		!s.ExcludeSource &&
		!s.NonToken &&
		!s.TokenOnly &&
		!s.PowerLessThanSource &&
		!s.PowerGreaterThanSource &&
		s.Name == ""
}

// Validate reports structural contradictions in the Selection that represent
// card-definition bugs rather than board-state outcomes. It returns one message
// per problem found and is consumed by ValidateCardDef.
func (s Selection) Validate() []string {
	var problems []string
	for i := range s.AnyOf {
		for _, problem := range s.AnyOf[i].Validate() {
			problems = append(problems, fmt.Sprintf("alternative %d: %s", i, problem))
		}
	}
	for _, t := range s.RequiredTypes {
		if slices.Contains(s.ExcludedTypes, t) {
			problems = append(problems, fmt.Sprintf("card type %v is both required and excluded", t))
		}
	}
	for _, t := range s.Supertypes {
		if t == s.ExcludedSupertype {
			problems = append(problems, fmt.Sprintf("supertype %v is both required and excluded", t))
		}
	}
	if len(s.RequiredTypesAny) > 0 && !slices.ContainsFunc(s.RequiredTypesAny, func(t types.Card) bool {
		return !slices.Contains(s.ExcludedTypes, t)
	}) {
		problems = append(problems, "every any-of card type is excluded")
	}
	if len(s.ColorsAny) > 0 && !slices.ContainsFunc(s.ColorsAny, func(c color.Color) bool {
		return !slices.Contains(s.ExcludedColors, c)
	}) {
		problems = append(problems, "every any-of color is excluded")
	}
	for _, sub := range s.SubtypesAny {
		if sub == s.ExcludedSubtype {
			problems = append(problems, fmt.Sprintf("subtype %v is both required and excluded", sub))
		}
	}
	if s.Colorless && s.Multicolored {
		problems = append(problems, "selection cannot require both colorless and multicolored")
	}
	if s.Colorless && len(s.ColorsAny) > 0 {
		problems = append(problems, "selection cannot require both colorless and any color")
	}
	if s.Keyword != KeywordNone && s.Keyword == s.ExcludedKeyword {
		problems = append(problems, fmt.Sprintf("keyword %v is both required and excluded", s.Keyword))
	}
	if s.NonToken && s.TokenOnly {
		problems = append(problems, "selection cannot require both token and non-token objects")
	}
	return problems
}

// Selection returns the Selection-equivalent of a TargetPredicate. It shares the
// predicate's backing slices rather than copying them, so callers must not
// mutate the result.
func (p TargetPredicate) Selection() Selection {
	selection := Selection{
		ExcludedTypes:     p.ExcludedTypes,
		Supertypes:        p.Supertypes,
		ExcludedSupertype: p.ExcludedSupertype,
		SubtypesAny:       p.Subtypes,
		ColorsAny:         p.Colors,
		ExcludedColors:    p.ExcludedColors,
		Controller:        p.Controller,
		Player:            p.Player,
		Tapped:            p.Tapped,
		CombatState:       p.CombatState,
		Keyword:           p.Keyword,
		ExcludedKeyword:   p.ExcludedKeyword,
		ManaValue:         p.ManaValue,
		Power:             p.Power,
		Toughness:         p.Toughness,
		ExcludeSource:     p.Another,

		PowerLessThanSource:    p.PowerLessThanSource,
		PowerGreaterThanSource: p.PowerGreaterThanSource,
		TokenOnly:              p.TokenOnly,
		NonToken:               p.NonToken,
	}
	if p.PermanentTypesConjunctive {
		selection.RequiredTypes = p.PermanentTypes
	} else {
		selection.RequiredTypesAny = p.PermanentTypes
	}
	return selection
}

// SelectionCount pairs a Selection with the count and total-power thresholds
// that a "controls matching" condition needs but that Selection deliberately
// excludes. MinCount defaults to 1 when the Selection is non-empty. DistinctNames
// constrains how many of the matched permanents must have distinct names (for
// "with different names" qualifiers).
type SelectionCount struct {
	Selection     Selection
	MinCount      int
	TotalPower    opt.V[compare.Int]
	DistinctNames opt.V[compare.Int]
}

// Empty reports whether the SelectionCount carries no active predicate.
func (c SelectionCount) Empty() bool {
	return c.Selection.Empty() && c.MinCount == 0 && !c.TotalPower.Exists && !c.DistinctNames.Exists
}
