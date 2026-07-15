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

	// ChosenSubtypeFrom, when non-empty, requires the matched card to carry the
	// creature subtype published under this resolution-choice key (the source
	// permanent's entry-time "choose a creature type", seeded into the resolving
	// object's ResolutionChoices). The chosen value must be a known creature
	// subtype (KnownSubtypeForType) or the predicate fails closed. It backs the
	// "creature card of the chosen type" gate of chosen-type library-top triggers
	// (Herald's Horn). Unlike SubtypeChoice it names an explicit key and applies
	// the known-subtype guard, so it is matched only against cards in a
	// non-battlefield zone; other subjects fail it closed.
	ChosenSubtypeFrom ChoiceKey
	// ChosenCardTypeFrom requires the subject to have the card type published
	// under this resolution choice key.
	ChosenCardTypeFrom ChoiceKey

	// SharesCardTypeFromLinked requires the subject to share at least one card
	// type with the object recorded under this linked key by an earlier
	// instruction in the same resolution (its last-known information once it has
	// left the battlefield). It backs "a permanent ... that shares a card type
	// with it", where "it" is a just-sacrificed permanent whose types are read
	// from the linked object the sacrifice published (Braids, Arisen Nightmare).
	// A subject that shares no card type with that object, or whose linked object
	// is absent, never matches. The linked object is resolved against the
	// resolving stack object, so it reads the sacrificed permanent's types
	// through last-known information after it has left the battlefield.
	SharesCardTypeFromLinked LinkedKey

	// ColorsAny matches when any listed color is present. ExcludedColors must
	// all be absent. Colorless requires no colors; Multicolored requires at
	// least two colors. Colored requires one or more colors, i.e. the
	// permanent is not colorless ("permanents ... that are one or more colors",
	// All Is Dust). It is the complement of Colorless.
	ColorsAny      []color.Color
	ExcludedColors []color.Color
	Colorless      bool
	Multicolored   bool
	Colored        bool

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

	// MatchNoCounters, when true, requires the matched permanent to carry no
	// counters of any kind ("all creatures with no counters on them"). It is the
	// kind-agnostic negation of MatchAnyCounter. A non-battlefield subject, which
	// has no counters to inspect, never matches. Placed beside MatchAnyCounter to
	// pack into the bool cluster.
	MatchNoCounters bool

	// MatchExcludedCounter, when true, requires the matched permanent to carry no
	// counter of ExcludedCounter's kind ("each creature without a +1/+1 counter
	// on it"). Unlike MatchNoCounters it is kind-specific: counters of other
	// kinds do not disqualify the permanent. A non-battlefield subject, which has
	// no counters to inspect, never matches. Placed beside MatchNoCounters to
	// pack into the bool cluster.
	MatchExcludedCounter bool

	// MatchModified, when true, requires the matched permanent to be modified: it
	// carries one or more counters, or has one or more Auras or Equipment
	// attached to it ("modified creatures you control"). A non-battlefield
	// subject never matches.
	MatchModified bool

	// MatchCommander, when true, requires the matched permanent to be a commander
	// ("commander creatures you control"). A non-battlefield subject never
	// matches. Placed beside MatchModified to pack into the bool cluster.
	MatchCommander bool

	// MatchGoaded, when true, requires the matched permanent to be goaded right
	// now ("Whenever a goaded creature attacks", Vengeful Ancestor; CR 701.38).
	// A non-battlefield subject never matches. Placed beside MatchModified to
	// pack into the bool cluster.
	MatchGoaded bool

	// MatchEnchanted requires the matched permanent to have one or more Auras
	// attached to it ("as long as this creature is enchanted"); MatchEquipped
	// requires one or more Equipment attached to it ("as long as this creature
	// is equipped"). A non-battlefield subject never matches. Placed beside
	// MatchModified to pack into the bool cluster.
	MatchEnchanted bool
	MatchEquipped  bool

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

	// ManaValueDynamic, when set, bounds the matched card's mana value by a value
	// computed as the predicate is evaluated rather than by a fixed number,
	// modeling "with mana value less than or equal to the amount of life you
	// (lost|gained) this turn" (Betor, Ancestor's Voice). The comparison is
	// always less-than-or-equal and the bound is controller-relative: the "you"
	// names the viewing player. Only the turn-event life totals
	// (DynamicAmountLifeLostThisTurn, DynamicAmountLifeGainedThisTurn) are
	// modeled; any other dynamic amount is rejected by Validate.
	ManaValueDynamic opt.V[ManaValueDynamicBound]

	// RequiredCounter names the counter kind required when MatchCounter is set.
	RequiredCounter counter.Kind

	// ExcludedCounter names the counter kind forbidden when MatchExcludedCounter
	// is set ("without a +1/+1 counter on it").
	ExcludedCounter counter.Kind

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

	// PowerAboveBase requires the matched permanent's current power to be
	// strictly greater than its base power (CR 208.3): the power after
	// characteristic-defining and set effects but before counters and other
	// modifiers ("creatures you control ... with power greater than its base
	// power", Kutzil, Malamet Exemplar). Only a battlefield permanent can carry
	// the counters, Auras, Equipment, or temporary modifiers that raise power
	// above base, so a card or last-known snapshot never matches. Placed at the
	// end so the bool joins no existing cluster's packing.
	PowerAboveBase bool

	// Name, when non-empty, requires the matched object's name to equal it,
	// modeling a "card named <Name>" filter (Daru Cavalier, Trustworthy Scout's
	// library searches). It composes with the other fields but in practice stands
	// alone on a plain "card named X" effect. A subject whose name is unavailable
	// never matches. Placed at the end so the field joins no existing cluster.
	Name string

	// RequirePermanentCard requires the matched card to be a permanent card (one
	// with at least one permanent card type), the "if it's a permanent card" gate
	// of reveal-and-put effects (Chaos Warp). It is a card-zone predicate matched
	// against a card's printed types; a subject that is not a card fails it
	// closed. Placed at the end so the bool joins no existing cluster's packing.
	RequirePermanentCard bool

	// NameUniqueAmongControlled requires the matched permanent's name to differ
	// from every other permanent its controller controls ("target enchantment
	// you control that doesn't have the same name as another permanent you
	// control", Yenna, Redtooth Regent). A subject whose name is unavailable, or
	// that is not an on-battlefield permanent, fails it closed. Placed at the end
	// so the bool joins no existing cluster's packing.
	NameUniqueAmongControlled bool

	// SharesCreatureTypeWithSource requires the matched card to share at least
	// one creature type (subtype that is a creature type) with the predicate's
	// source permanent ("if it shares a creature type with this creature", the
	// Kinship ability word). It reads the source permanent's effective creature
	// subtypes, so a subject with no source, a source that is not a permanent, or
	// a source with no creature types never matches. Placed at the end so the
	// bool joins no existing cluster's packing.
	SharesCreatureTypeWithSource bool

	// DealtDamageThisTurn requires the matched permanent to have been dealt
	// damage during the current turn ("target creature that was dealt damage this
	// turn", Fatal Blow). The rules layer scans the current turn's damage events
	// for one whose damaged permanent is this object (CR 120). A non-battlefield
	// subject, which receives no damage, never matches. Placed at the end so the
	// bool joins no existing cluster's packing.
	DealtDamageThisTurn bool

	// OwnerNotController requires the matched permanent's owner to differ from
	// its controller ("creatures you control but don't own", Garland, Royal
	// Kidnapper). It is the permanent-only ownership predicate that distinguishes
	// permanents a player controls without owning from those they both own and
	// control. A non-battlefield subject, which has no distinct controller, never
	// matches. Placed at the end so the bool joins no existing cluster's packing.
	OwnerNotController bool

	// ControlledByEventPlayer requires the matched permanent to be controlled by
	// the player of the triggering event ("target creature that player controls",
	// Garland, Royal Kidnapper, where "that player" is the opponent who just
	// became the monarch). The rules layer resolves the event player from the
	// resolving ability's triggering event and compares it to the permanent's
	// controller. Outside a triggered resolution there is no event player, so the
	// predicate matches nothing. Placed at the end so the bool joins no existing
	// cluster's packing.
	ControlledByEventPlayer bool

	// ControlledByDefendingPlayer requires the matched permanent to be controlled
	// by the defending player of the triggering attack ("destroy target tapped
	// nonland permanent that player controls", The Spear of Bashenga, where "that
	// player" is the monarch whose attack was declared against them). It differs
	// from ControlledByEventPlayer because an attack event records the attacker in
	// its event player and the defending player separately, so "that player" on an
	// attack trigger names the attacked player, not the attacker. The rules layer
	// resolves the defending player from the resolving ability's triggering attack
	// event and compares it to the permanent's controller. Outside a triggered
	// attack resolution there is no defending player, so the predicate matches
	// nothing. Placed at the end so the bool joins no existing cluster's packing.
	ControlledByDefendingPlayer bool

	// ManaValueLessThanEventPermanent requires the matched card's mana value to
	// be strictly less than the mana value of the triggering event's permanent —
	// the creature that died ("return target Cleric card with lesser mana value
	// from your graveyard to the battlefield", Orah, Skyclave Hierophant) or the
	// artifact put into a graveyard ("return to your hand target artifact card in
	// your graveyard with lesser mana value", Scrap Trawler). It is the
	// mana-value, event-relative sibling of PowerLessThanSource: the bound reads
	// the triggering event's permanent through last-known information (CR
	// 608.2h), not the ability's source, so a card returned by a self-or-subtype
	// dies trigger is compared to whichever permanent died. Outside a triggered
	// resolution, or when the event names no permanent, the bound matches
	// nothing. Placed at the end so the bool joins no existing cluster's packing.
	ManaValueLessThanEventPermanent bool

	// Owner constrains the matched permanent by its OWNER relative to the viewing
	// player, independent of who controls it ("Commander creatures you own",
	// Dungeon Delver and other Backgrounds). Unlike Controller, which follows
	// control changes, Owner keys off the permanent's fixed owner, so a commander
	// owned by the viewing player but controlled by an opponent (a stolen
	// commander) still matches OwnerYou, while a commander controlled by the
	// viewing player but owned by an opponent does not. A non-battlefield subject,
	// which the runtime evaluates without an owner, never matches a non-Any
	// relation. Placed at the end so the field joins no existing cluster's
	// packing.
	Owner OwnerRelation
}

// ManaValueDynamicBound bounds a card's mana value by a controller-relative
// value computed as the predicate is evaluated rather than a fixed number. The
// comparison is always less-than-or-equal; Multiplier and Addend scale and
// shift the evaluated amount (CR 608.2c). It backs Selection.ManaValueDynamic.
// Two amount families are modeled: the turn-event life totals
// (DynamicAmountLifeLostThisTurn, DynamicAmountLifeGainedThisTurn — Betor,
// Ancestor's Voice) which need no group, and a battlefield permanent count
// (DynamicAmountCountSelector — "the number of lands you control", Beseech the
// Queen) whose Group narrows which permanents are counted.
type ManaValueDynamicBound struct {
	Kind       DynamicAmountKind
	Multiplier int
	Addend     int
	// Group narrows the permanents counted by a DynamicAmountCountSelector
	// bound ("with mana value less than or equal to the number of lands you
	// control", Beseech the Queen). It is nil for the turn-event life totals,
	// which need no group. Held by pointer to keep Selection a finite type.
	Group *GroupReference
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
		!s.Colored &&
		s.Controller == ControllerAny &&
		s.Player == PlayerAny &&
		s.Owner == OwnerAny &&
		s.Tapped == TriAny &&
		s.CombatState == CombatStateAny &&
		s.Keyword == KeywordNone &&
		s.ExcludedKeyword == KeywordNone &&
		!s.ManaValue.Exists &&
		!s.Power.Exists &&
		!s.Toughness.Exists &&
		!s.ManaValueDynamic.Exists &&
		!s.MatchCounter &&
		!s.MatchAnyCounter &&
		!s.MatchNoCounters &&
		!s.MatchExcludedCounter &&
		!s.RequiredCounterCount.Exists &&
		!s.EnteredThisTurn &&
		!s.MatchModified &&
		!s.MatchCommander &&
		!s.MatchGoaded &&
		!s.MatchEnchanted &&
		!s.MatchEquipped &&
		s.ColorChoice == ColorChoiceNone &&
		!s.ExcludeSource &&
		!s.NonToken &&
		!s.TokenOnly &&
		!s.PowerLessThanSource &&
		!s.PowerGreaterThanSource &&
		!s.PowerAboveBase &&
		s.Name == "" &&
		s.ChosenSubtypeFrom == "" &&
		s.ChosenCardTypeFrom == "" &&
		s.SharesCardTypeFromLinked == "" &&
		!s.RequirePermanentCard &&
		!s.NameUniqueAmongControlled &&
		!s.SharesCreatureTypeWithSource &&
		!s.DealtDamageThisTurn &&
		!s.OwnerNotController &&
		!s.ControlledByEventPlayer &&
		!s.ControlledByDefendingPlayer &&
		!s.ManaValueLessThanEventPermanent
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
	if s.Colorless && s.Colored {
		problems = append(problems, "selection cannot require both colorless and colored")
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
	if s.MatchNoCounters && (s.MatchAnyCounter || s.MatchCounter || s.RequiredCounterCount.Exists) {
		problems = append(problems, "selection cannot require both no counters and a counter")
	}
	if s.MatchExcludedCounter && s.MatchCounter && s.ExcludedCounter == s.RequiredCounter {
		problems = append(problems, "selection cannot both require and exclude the same counter kind")
	}
	if s.ManaValueDynamic.Exists {
		switch s.ManaValueDynamic.Val.Kind {
		case DynamicAmountLifeLostThisTurn, DynamicAmountLifeGainedThisTurn:
			if s.ManaValueDynamic.Val.Group != nil {
				problems = append(problems, "dynamic mana-value life-total bound must not set a group")
			}
		case DynamicAmountCountSelector:
			if s.ManaValueDynamic.Val.Group == nil || s.ManaValueDynamic.Val.Group.Empty() {
				problems = append(problems, "dynamic mana-value count bound requires a group")
			} else {
				problems = append(problems, s.ManaValueDynamic.Val.Group.Validate()...)
			}
		default:
			problems = append(problems, fmt.Sprintf("dynamic mana-value bound uses unsupported amount %v", s.ManaValueDynamic.Val.Kind))
		}
	}
	return problems
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
