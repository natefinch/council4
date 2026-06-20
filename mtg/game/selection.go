package game

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Selection is pure rules data describing WHICH game objects share a predicate.
// It is the single, valence-agnostic matcher description that subsumes the
// characteristic fields formerly duplicated across TargetPredicate,
// PermanentFilter, the permanent/card filters of TriggerPattern, and the
// historical mass-effect selector constants.
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
		!s.ExcludeSource &&
		!s.NonToken &&
		!s.TokenOnly
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
	return Selection{
		RequiredTypesAny:  p.PermanentTypes,
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
	}
}

// Selection returns the characteristic-matching portion of a PermanentFilter as
// a Selection. The filter's count and total-power concerns stay outside
// Selection. The result shares the filter's backing slices.
func (f PermanentFilter) Selection() Selection {
	return Selection{
		RequiredTypes:  f.Types,
		Supertypes:     f.Supertypes,
		SubtypesAny:    f.SubtypesAny,
		ColorsAny:      f.ColorsAny,
		ExcludedColors: f.ExcludedColors,
		Power:          f.Power,
		Toughness:      f.Toughness,
		ExcludeSource:  f.ExcludeSource,
	}
}

// SelectionCount pairs a Selection with the count and total-power thresholds
// that a "controls matching" condition needs but that Selection deliberately
// excludes. MinCount defaults to 1 when the Selection is non-empty.
type SelectionCount struct {
	Selection  Selection
	MinCount   int
	TotalPower opt.V[compare.Int]
}

// Empty reports whether the SelectionCount carries no active predicate.
func (c SelectionCount) Empty() bool {
	return c.Selection.Empty() && c.MinCount == 0 && !c.TotalPower.Exists
}
