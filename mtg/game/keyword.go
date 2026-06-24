package game

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// KeywordAbility is a sealed data-only variant for one printed keyword ability.
type KeywordAbility interface {
	isKeywordAbility()
	keyword() Keyword
	cloneKeywordAbility() KeywordAbility
}

// SimpleKeyword is a non-parameterized keyword ability such as flying or haste.
type SimpleKeyword struct {
	Kind Keyword
}

// WardKeyword parameterizes Ward for mana-valued ward costs. AdditionalCosts
// carries the non-mana components of a composite or non-mana ward cost
// ("Ward—Pay 2 life.", "Ward—{2}, Pay 2 life.", "Ward—Sacrifice a creature.").
// An opponent must pay every component of Cost and AdditionalCosts together, or
// the spell or ability that targeted this permanent is countered (CR 702.21).
type WardKeyword struct {
	Cost            cost.Mana
	AdditionalCosts []cost.Additional
}

// CumulativeUpkeepKeyword parameterizes cumulative upkeep for fixed mana costs.
type CumulativeUpkeepKeyword struct {
	Cost cost.Mana
}

// EquipKeyword parameterizes Equip activation costs.
type EquipKeyword struct {
	Cost cost.Mana
}

// ReconfigureKeyword parameterizes Reconfigure activation costs (CR 702.151).
// It carries the Reconfigure keyword identity inside the activated ability built
// by ReconfigureActivatedAbility so the rules layer can dispatch the attachment
// like Equip.
type ReconfigureKeyword struct {
	Cost cost.Mana
}

// EnchantKeyword parameterizes Enchant attachment legality.
type EnchantKeyword struct {
	Target TargetSpec
}

// CyclingKeyword parameterizes Cycling activation costs.
type CyclingKeyword struct {
	Cost cost.Mana
}

// ScavengeKeyword parameterizes Scavenge for its graveyard-activation mana cost.
type ScavengeKeyword struct {
	Cost cost.Mana
}

// UnearthKeyword parameterizes Unearth for its graveyard-activation mana cost
// (CR 702.83). While this card is in its owner's graveyard, paying Cost at
// sorcery speed returns it to the battlefield with haste until it is exiled at
// the next end step.
type UnearthKeyword struct {
	Cost cost.Mana
}

// NinjutsuKeyword parameterizes Ninjutsu activation costs.
type NinjutsuKeyword struct {
	Cost cost.Mana
}

// OutlastKeyword parameterizes Outlast activation costs.
type OutlastKeyword struct {
	Cost cost.Mana
}

// SaddleKeyword parameterizes the Saddle N keyword (CR 702.166). Power is the
// total power of other creatures the controller must tap to saddle the Mount.
type SaddleKeyword struct {
	Power int
}

// CrewKeyword parameterizes the Crew N keyword (CR 702.122). Power is the total
// power of creatures the controller must tap to make the Vehicle become an
// artifact creature until end of turn.
type CrewKeyword struct {
	Power int
}

// MutateKeyword parameterizes Mutate alternative casting costs.
type MutateKeyword struct {
	Cost cost.Mana
}

// KickerKeyword parameterizes Kicker additional costs and bonus instructions.
type KickerKeyword struct {
	Cost         cost.Mana
	BonusContent AbilityContent
	// Multi marks the keyword as Multikicker (CR 702.32): the additional cost may
	// be paid any number of times as the spell is cast, and the number of times
	// it was paid (the kick count) scales "for each time it was kicked" payoffs.
	// It is false for ordinary Kicker, which may be paid at most once.
	Multi bool
}

// MadnessKeyword parameterizes Madness alternative costs.
type MadnessKeyword struct {
	Cost cost.Mana
}

// FlashbackKeyword parameterizes Flashback alternative casting costs.
type FlashbackKeyword struct {
	Cost cost.Mana
}

// MorphKeyword parameterizes Morph turn-face-up costs.
type MorphKeyword struct {
	Cost cost.Mana
}

// DisguiseKeyword parameterizes Disguise turn-face-up costs.
type DisguiseKeyword struct {
	Cost cost.Mana
}

// SuspendKeyword parameterizes Suspend costs and time counters.
type SuspendKeyword struct {
	Cost         cost.Mana
	TimeCounters int
}

// ProtectionKeyword parameterizes Protection effects.
type ProtectionKeyword struct {
	FromColors   []color.Color // protection from specific colors
	FromTypes    []types.Card  // protection from card types (artifact, creature, …)
	FromSubtypes []types.Sub   // protection from creature/land subtypes (Dragon, Human, …)
	Multicolored bool          // protection from sources with ≥2 colors
	Monocolored  bool          // protection from sources with exactly 1 color
	Everything   bool          // protection from all sources
	EachColor    bool          // protection from sources with any color (all five)
	// ChosenColor marks a grant that resolves to protection from a single color
	// chosen as the granting ability resolves ("protection from the color of
	// your choice"). The rules resolve the choice and rewrite this into
	// FromColors before the continuous effect is stored, so protection checks
	// never observe ChosenColor.
	ChosenColor bool
}

// HideawayKeyword parameterizes the Hideaway N keyword (CR 702.75). Amount is
// the number of cards looked at from the top of the library when the permanent
// enters; one of them is exiled face down and linked to the source, and the
// rest are put on the bottom of the library in a random order.
type HideawayKeyword struct {
	Amount int
}

// ToxicKeyword parameterizes the number of poison counters given after combat
// damage to a player.
type ToxicKeyword struct {
	Amount int
}

// FabricateKeyword parameterizes Fabricate for its counter-or-token count.
type FabricateKeyword struct {
	Count int
}

// RampageKeyword parameterizes Rampage for its per-extra-blocker bonus
// (CR 702.23). Count is the printed N: the creature gets +N/+N until end of turn
// for each creature blocking it beyond the first.
type RampageKeyword struct {
	Count int
}

// SoulshiftKeyword parameterizes Soulshift for its mana-value bound on the
// Spirit card returned when this creature dies (CR 702.46).
type SoulshiftKeyword struct {
	Count int
}

// DredgeKeyword parameterizes Dredge for its mill count (CR 702.52). While this
// card is in its owner's graveyard, if that player would draw a card they may
// instead mill Count cards and return this card from the graveyard to their
// hand. Count is the printed N and must be positive.
type DredgeKeyword struct {
	Count int
}

// LandwalkKeyword parameterizes the landwalk evasion family (CR 702.14). A
// creature with landwalk can't be blocked as long as the defending player
// controls a land matching this filter: a land with Subtype (Forest, Island,
// Swamp, Mountain, Plains, Desert) for the typed variants, any land when AnyLand
// is true (generic "landwalk"), or a nonbasic land (a land without the Basic
// supertype) when Nonbasic is true.
type LandwalkKeyword struct {
	Subtype  types.Sub
	AnyLand  bool
	Nonbasic bool
}

func (SimpleKeyword) isKeywordAbility()           {}
func (WardKeyword) isKeywordAbility()             {}
func (CumulativeUpkeepKeyword) isKeywordAbility() {}
func (EquipKeyword) isKeywordAbility()            {}
func (ReconfigureKeyword) isKeywordAbility()      {}
func (EnchantKeyword) isKeywordAbility()          {}
func (CyclingKeyword) isKeywordAbility()          {}
func (NinjutsuKeyword) isKeywordAbility()         {}
func (OutlastKeyword) isKeywordAbility()          {}
func (MutateKeyword) isKeywordAbility()           {}
func (KickerKeyword) isKeywordAbility()           {}
func (MadnessKeyword) isKeywordAbility()          {}
func (FlashbackKeyword) isKeywordAbility()        {}
func (MorphKeyword) isKeywordAbility()            {}
func (DisguiseKeyword) isKeywordAbility()         {}
func (SuspendKeyword) isKeywordAbility()          {}
func (ProtectionKeyword) isKeywordAbility()       {}
func (ToxicKeyword) isKeywordAbility()            {}
func (HideawayKeyword) isKeywordAbility()         {}
func (ScavengeKeyword) isKeywordAbility()         {}
func (UnearthKeyword) isKeywordAbility()          {}
func (FabricateKeyword) isKeywordAbility()        {}
func (RampageKeyword) isKeywordAbility()          {}
func (SoulshiftKeyword) isKeywordAbility()        {}
func (DredgeKeyword) isKeywordAbility()           {}
func (LandwalkKeyword) isKeywordAbility()         {}
func (SaddleKeyword) isKeywordAbility()           {}
func (CrewKeyword) isKeywordAbility()             {}

func (ability SimpleKeyword) keyword() Keyword { return ability.Kind }
func (WardKeyword) keyword() Keyword           { return Ward }
func (CumulativeUpkeepKeyword) keyword() Keyword {
	return CumulativeUpkeep
}
func (EquipKeyword) keyword() Keyword { return Equip }
func (ReconfigureKeyword) keyword() Keyword {
	return Reconfigure
}
func (EnchantKeyword) keyword() Keyword    { return Enchant }
func (CyclingKeyword) keyword() Keyword    { return Cycling }
func (NinjutsuKeyword) keyword() Keyword   { return Ninjutsu }
func (OutlastKeyword) keyword() Keyword    { return Outlast }
func (MutateKeyword) keyword() Keyword     { return Mutate }
func (KickerKeyword) keyword() Keyword     { return Kicker }
func (MadnessKeyword) keyword() Keyword    { return Madness }
func (FlashbackKeyword) keyword() Keyword  { return Flashback }
func (MorphKeyword) keyword() Keyword      { return Morph }
func (DisguiseKeyword) keyword() Keyword   { return Disguise }
func (SuspendKeyword) keyword() Keyword    { return Suspend }
func (ProtectionKeyword) keyword() Keyword { return Protection }
func (ToxicKeyword) keyword() Keyword      { return Toxic }
func (HideawayKeyword) keyword() Keyword   { return Hideaway }
func (ScavengeKeyword) keyword() Keyword   { return Scavenge }
func (UnearthKeyword) keyword() Keyword    { return Unearth }
func (FabricateKeyword) keyword() Keyword  { return Fabricate }
func (RampageKeyword) keyword() Keyword    { return Rampage }
func (SoulshiftKeyword) keyword() Keyword  { return Soulshift }
func (DredgeKeyword) keyword() Keyword     { return Dredge }
func (LandwalkKeyword) keyword() Keyword   { return Landwalk }
func (SaddleKeyword) keyword() Keyword     { return Saddle }
func (CrewKeyword) keyword() Keyword       { return Crew }

func (ability SimpleKeyword) cloneKeywordAbility() KeywordAbility { return ability }
func (ability WardKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	ability.AdditionalCosts = slices.Clone(ability.AdditionalCosts)
	return ability
}
func (ability CumulativeUpkeepKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}
func (ability EquipKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}

func (ability ReconfigureKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}

func (ability EnchantKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Target = cloneTargetSpecs([]TargetSpec{ability.Target})[0]
	return ability
}
func (ability CyclingKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}
func (ability NinjutsuKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}
func (ability OutlastKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}
func (ability MutateKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}
func (ability KickerKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	ability.BonusContent = cloneAbilityContent(ability.BonusContent)
	return ability
}
func (ability MadnessKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}
func (ability FlashbackKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}
func (ability MorphKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}
func (ability DisguiseKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}
func (ability SuspendKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}
func (ability ProtectionKeyword) cloneKeywordAbility() KeywordAbility {
	ability.FromColors = append([]color.Color(nil), ability.FromColors...)
	ability.FromTypes = append([]types.Card(nil), ability.FromTypes...)
	ability.FromSubtypes = append([]types.Sub(nil), ability.FromSubtypes...)
	return ability
}
func (ability ToxicKeyword) cloneKeywordAbility() KeywordAbility    { return ability }
func (ability HideawayKeyword) cloneKeywordAbility() KeywordAbility { return ability }
func (ability ScavengeKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}
func (ability UnearthKeyword) cloneKeywordAbility() KeywordAbility {
	ability.Cost = append(cost.Mana(nil), ability.Cost...)
	return ability
}
func (ability FabricateKeyword) cloneKeywordAbility() KeywordAbility { return ability }
func (ability RampageKeyword) cloneKeywordAbility() KeywordAbility   { return ability }
func (ability SoulshiftKeyword) cloneKeywordAbility() KeywordAbility { return ability }
func (ability DredgeKeyword) cloneKeywordAbility() KeywordAbility    { return ability }
func (ability LandwalkKeyword) cloneKeywordAbility() KeywordAbility  { return ability }
func (ability SaddleKeyword) cloneKeywordAbility() KeywordAbility    { return ability }
func (ability CrewKeyword) cloneKeywordAbility() KeywordAbility      { return ability }

// SimpleKeywords returns sealed keyword variants for non-parameterized keywords.
func SimpleKeywords(keywords ...Keyword) []KeywordAbility {
	abilities := make([]KeywordAbility, 0, len(keywords))
	for _, keyword := range keywords {
		abilities = append(abilities, SimpleKeyword{Kind: keyword})
	}
	return abilities
}

// BodyToxicAmount returns the Toxic value carried by an ability body.
func BodyToxicAmount(body Ability) (int, bool) {
	ability, ok := BodyKeywordAbility(body, Toxic)
	if !ok {
		return 0, false
	}
	toxic, ok := ability.(ToxicKeyword)
	if !ok || toxic.Amount <= 0 {
		return 0, false
	}
	return toxic.Amount, true
}

// KeywordAbilityKind returns the Keyword represented by a sealed keyword variant.
func KeywordAbilityKind(ability KeywordAbility) Keyword {
	if ability == nil {
		panic("game: nil KeywordAbility")
	}
	return ability.keyword()
}

// BodyKeywordAbilities returns the keyword abilities carried by a sealed body.
func BodyKeywordAbilities(body Ability) []KeywordAbility {
	switch b := body.(type) {
	case *StaticAbility:
		if b == nil {
			return nil
		}
		return b.KeywordAbilities
	case *ActivatedAbility:
		if b == nil {
			return nil
		}
		return b.KeywordAbilities
	case *TriggeredAbility:
		if b == nil {
			return nil
		}
		return b.KeywordAbilities
	default:
		return nil
	}
}

// BodyHasKeyword reports whether a sealed body carries the given keyword.
func BodyHasKeyword(body Ability, kw Keyword) bool {
	_, ok := BodyKeywordAbility(body, kw)
	return ok
}

// BodyKeywordAbility returns the first sealed keyword variant matching kw on body.
func BodyKeywordAbility(body Ability, kw Keyword) (KeywordAbility, bool) {
	for _, ka := range BodyKeywordAbilities(body) {
		if KeywordAbilityKind(ka) == kw {
			return ka, true
		}
	}
	return nil, false
}

// BodyAddKeywordKindsTo adds every keyword represented by body to m.
func BodyAddKeywordKindsTo(body Ability, m map[Keyword]bool) {
	for _, ka := range BodyKeywordAbilities(body) {
		m[KeywordAbilityKind(ka)] = true
	}
}

// BodyWardCost returns the Ward cost from a TriggeredAbilityBody's keywords.
func BodyWardCost(body *TriggeredAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Ward)
	if !ok {
		return nil, false
	}
	ward, ok := ka.(WardKeyword)
	if !ok {
		return nil, false
	}
	return ward.Cost, true
}

// BodyWardKeyword returns the full Ward keyword (mana plus any non-mana
// additional cost components) from a triggered ability's keywords.
func BodyWardKeyword(body *TriggeredAbility) (WardKeyword, bool) {
	ka, ok := BodyKeywordAbility(body, Ward)
	if !ok {
		return WardKeyword{}, false
	}
	ward, ok := ka.(WardKeyword)
	if !ok {
		return WardKeyword{}, false
	}
	return ward, true
}

// StaticBodyWardCost returns the Ward cost from a static ability.
func StaticBodyWardCost(body *StaticAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Ward)
	if !ok {
		return nil, false
	}
	ward, ok := ka.(WardKeyword)
	if !ok {
		return nil, false
	}
	return ward.Cost, true
}

// StaticBodyWardCosts returns the mana and additional payment of a static
// ability's Ward keyword, or (nil, nil, false) when the body has no Ward
// keyword. It carries the composite and non-mana Ward payment ("Ward—{2}, Pay 2
// life.", "Ward—Sacrifice a creature.") that StaticBodyWardCost cannot express
// with mana alone.
func StaticBodyWardCosts(body *StaticAbility) (cost.Mana, []cost.Additional, bool) {
	ka, ok := BodyKeywordAbility(body, Ward)
	if !ok {
		return nil, nil, false
	}
	ward, ok := ka.(WardKeyword)
	if !ok {
		return nil, nil, false
	}
	return ward.Cost, ward.AdditionalCosts, true
}

// StaticBodyDredgeCount returns the Dredge mill count carried by a static
// ability, or (0, false) when the body has no Dredge keyword.
func StaticBodyDredgeCount(body *StaticAbility) (int, bool) {
	ka, ok := BodyKeywordAbility(body, Dredge)
	if !ok {
		return 0, false
	}
	dredge, ok := ka.(DredgeKeyword)
	if !ok || dredge.Count <= 0 {
		return 0, false
	}
	return dredge.Count, true
}

// ActivatedBodyCyclingCost returns the Cycling cost from an activated ability.
func ActivatedBodyCyclingCost(body *ActivatedAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Cycling)
	if !ok {
		return nil, false
	}
	cycling, ok := ka.(CyclingKeyword)
	if !ok {
		return nil, false
	}
	return cycling.Cost, true
}

// ActivatedBodyScavengeCost returns the Scavenge cost from an activated ability.
func ActivatedBodyScavengeCost(body *ActivatedAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Scavenge)
	if !ok {
		return nil, false
	}
	scavenge, ok := ka.(ScavengeKeyword)
	if !ok {
		return nil, false
	}
	return scavenge.Cost, true
}

// ActivatedBodyUnearthCost returns the Unearth cost from an activated ability.
func ActivatedBodyUnearthCost(body *ActivatedAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Unearth)
	if !ok {
		return nil, false
	}
	unearth, ok := ka.(UnearthKeyword)
	if !ok {
		return nil, false
	}
	return unearth.Cost, true
}

// ActivatedBodySaddlePower returns the Saddle N power threshold from an
// activated ability, used to recognize and render the Saddle template.
func ActivatedBodySaddlePower(body *ActivatedAbility) (int, bool) {
	ka, ok := BodyKeywordAbility(body, Saddle)
	if !ok {
		return 0, false
	}
	saddle, ok := ka.(SaddleKeyword)
	if !ok {
		return 0, false
	}
	return saddle.Power, true
}

// ActivatedBodyCrewPower returns the Crew N power threshold from an activated
// ability, used to recognize and render the Crew template.
func ActivatedBodyCrewPower(body *ActivatedAbility) (int, bool) {
	ka, ok := BodyKeywordAbility(body, Crew)
	if !ok {
		return 0, false
	}
	crew, ok := ka.(CrewKeyword)
	if !ok {
		return 0, false
	}
	return crew.Power, true
}

// ActivatedBodyNinjutsuCost returns the Ninjutsu cost from an activated ability.
func ActivatedBodyNinjutsuCost(body *ActivatedAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Ninjutsu)
	if !ok {
		return nil, false
	}
	ninjutsu, ok := ka.(NinjutsuKeyword)
	if !ok {
		return nil, false
	}
	return append(cost.Mana(nil), ninjutsu.Cost...), true
}

// StaticBodyMutateCost returns the Mutate cost from a static ability.
func StaticBodyMutateCost(body *StaticAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Mutate)
	if !ok {
		return nil, false
	}
	mutate, ok := ka.(MutateKeyword)
	if !ok {
		return nil, false
	}
	return append(cost.Mana(nil), mutate.Cost...), true
}

// ActivatedBodyEquipCost returns the Equip cost from an activated ability.
func ActivatedBodyEquipCost(body *ActivatedAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Equip)
	if !ok {
		return nil, false
	}
	equip, ok := ka.(EquipKeyword)
	if !ok {
		return nil, false
	}
	return equip.Cost, true
}

// ActivatedBodyReconfigureCost returns the Reconfigure cost from an activated
// ability.
func ActivatedBodyReconfigureCost(body *ActivatedAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Reconfigure)
	if !ok {
		return nil, false
	}
	reconfigure, ok := ka.(ReconfigureKeyword)
	if !ok {
		return nil, false
	}
	return reconfigure.Cost, true
}

// BodyMadnessCost returns the Madness cost from a TriggeredAbilityBody's keywords.
func BodyMadnessCost(body *TriggeredAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Madness)
	if !ok {
		return nil, false
	}
	madness, ok := ka.(MadnessKeyword)
	if !ok {
		return nil, false
	}
	return madness.Cost, true
}

// StaticBodyEnchantTarget returns the Enchant target spec for a static ability body.
func StaticBodyEnchantTarget(body *StaticAbility) (TargetSpec, bool) {
	ka, ok := BodyKeywordAbility(body, Enchant)
	if !ok {
		return TargetSpec{}, false
	}
	enchant, ok := ka.(EnchantKeyword)
	if !ok {
		return TargetSpec{}, false
	}
	return enchant.Target, true
}

// ActivatedBodySuspendInfo returns suspend info from an ActivatedAbilityBody.
func ActivatedBodySuspendInfo(body *ActivatedAbility) (cost.Mana, int, bool) {
	ka, ok := BodyKeywordAbility(body, Suspend)
	if !ok {
		return nil, 0, false
	}
	suspend, ok := ka.(SuspendKeyword)
	if !ok || suspend.TimeCounters <= 0 {
		return nil, 0, false
	}
	return suspend.Cost, suspend.TimeCounters, true
}

// StaticBodySuspendInfo returns suspend info from a StaticAbilityBody.
func StaticBodySuspendInfo(body *StaticAbility) (cost.Mana, int, bool) {
	ka, ok := BodyKeywordAbility(body, Suspend)
	if !ok {
		return nil, 0, false
	}
	suspend, ok := ka.(SuspendKeyword)
	if !ok || suspend.TimeCounters <= 0 {
		return nil, 0, false
	}
	return suspend.Cost, suspend.TimeCounters, true
}

// ActivatedBodyMorphCost returns the morph cost from an ActivatedAbilityBody.
func ActivatedBodyMorphCost(body *ActivatedAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Morph)
	if !ok {
		return nil, false
	}
	morph, ok := ka.(MorphKeyword)
	if !ok {
		return nil, false
	}
	return morph.Cost, true
}

// StaticBodyMorphCost returns the morph cost from a StaticAbilityBody.
func StaticBodyMorphCost(body *StaticAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Morph)
	if !ok {
		return nil, false
	}
	morph, ok := ka.(MorphKeyword)
	if !ok {
		return nil, false
	}
	return morph.Cost, true
}

// ActivatedBodyDisguiseCost returns the disguise cost from an ActivatedAbilityBody.
func ActivatedBodyDisguiseCost(body *ActivatedAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Disguise)
	if !ok {
		return nil, false
	}
	disguise, ok := ka.(DisguiseKeyword)
	if !ok {
		return nil, false
	}
	return disguise.Cost, true
}

// StaticBodyDisguiseCost returns the disguise cost from a StaticAbilityBody.
func StaticBodyDisguiseCost(body *StaticAbility) (cost.Mana, bool) {
	ka, ok := BodyKeywordAbility(body, Disguise)
	if !ok {
		return nil, false
	}
	disguise, ok := ka.(DisguiseKeyword)
	if !ok {
		return nil, false
	}
	return disguise.Cost, true
}

// ActivatedBodyEternalize reports whether the body is an eternalize ability.
func ActivatedBodyEternalize(body *ActivatedAbility) bool {
	return BodyHasKeyword(body, Eternalize)
}

// ActivatedBodyEmbalm reports whether the body is an embalm ability.
func ActivatedBodyEmbalm(body *ActivatedAbility) bool {
	return BodyHasKeyword(body, Embalm)
}

// ActivatedBodyEternalizeParams returns the mana cost and the card's printed
// creature subtypes from an Eternalize activated body, reporting whether the
// body carries the Eternalize keyword and the structure those parameters
// produce. Callers confirm the exact body with EternalizeActivatedBody.
func ActivatedBodyEternalizeParams(body *ActivatedAbility) (cost.Mana, []types.Sub, bool) {
	return eternalizeFamilyParams(body, Eternalize)
}

// ActivatedBodyEmbalmParams returns the mana cost and the card's printed
// creature subtypes from an Embalm activated body, reporting whether the body
// carries the Embalm keyword. Callers confirm the exact body with
// EmbalmActivatedBody.
func ActivatedBodyEmbalmParams(body *ActivatedAbility) (cost.Mana, []types.Sub, bool) {
	return eternalizeFamilyParams(body, Embalm)
}

func eternalizeFamilyParams(body *ActivatedAbility, kw Keyword) (cost.Mana, []types.Sub, bool) {
	if !BodyHasKeyword(body, kw) || !body.ManaCost.Exists {
		return nil, nil, false
	}
	subtypes, ok := tokenCopyAddedCreatureSubtypes(body)
	if !ok {
		return nil, nil, false
	}
	return body.ManaCost.Val, subtypes, true
}

// tokenCopyAddedCreatureSubtypes recovers the card's printed creature subtypes
// from an Eternalize/Embalm token-copy body by dropping the leading Zombie type
// the builders prepend. It reports false when the body is not a single
// copy-token instruction in that exact shape.
func tokenCopyAddedCreatureSubtypes(body *ActivatedAbility) ([]types.Sub, bool) {
	if len(body.Content.Modes) != 1 || len(body.Content.Modes[0].Sequence) != 1 {
		return nil, false
	}
	create, ok := body.Content.Modes[0].Sequence[0].Primitive.(CreateToken)
	if !ok {
		return nil, false
	}
	spec, ok := create.Source.TokenCopy()
	if !ok {
		return nil, false
	}
	if len(spec.SetSubtypes) == 0 || spec.SetSubtypes[0] != types.Zombie {
		return nil, false
	}
	return spec.SetSubtypes[1:], true
}

// ActivatedBodyKicker returns the KickerKeyword from an ActivatedAbilityBody's keywords.
func ActivatedBodyKicker(body *ActivatedAbility) (KickerKeyword, bool) {
	ka, ok := BodyKeywordAbility(body, Kicker)
	if !ok {
		return KickerKeyword{}, false
	}
	kicker, ok := ka.(KickerKeyword)
	return kicker, ok
}

// StaticBodyProtectionColors returns the protection colors from a StaticAbilityBody.
func StaticBodyProtectionColors(body *StaticAbility) []color.Color {
	ka, ok := BodyKeywordAbility(body, Protection)
	if !ok {
		return nil
	}
	protection, ok := ka.(ProtectionKeyword)
	if !ok {
		return nil
	}
	return protection.FromColors
}

// StaticBodyProtectionKeyword returns the ProtectionKeyword from a StaticAbility body.
func StaticBodyProtectionKeyword(body *StaticAbility) (ProtectionKeyword, bool) {
	ka, ok := BodyKeywordAbility(body, Protection)
	if !ok {
		return ProtectionKeyword{}, false
	}
	protection, ok := ka.(ProtectionKeyword)
	return protection, ok
}

// StaticBodyLandwalkKeyword returns the LandwalkKeyword from a StaticAbility body.
func StaticBodyLandwalkKeyword(body *StaticAbility) (LandwalkKeyword, bool) {
	ka, ok := BodyKeywordAbility(body, Landwalk)
	if !ok {
		return LandwalkKeyword{}, false
	}
	landwalk, ok := ka.(LandwalkKeyword)
	return landwalk, ok
}
