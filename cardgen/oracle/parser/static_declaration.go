package parser

import (
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// StaticDeclarationKind identifies the static-declaration family the parser
// recognized for one composable clause.
type StaticDeclarationKind string

// Static declaration families recognized by the parser.
const (
	StaticDeclarationUnknown                              StaticDeclarationKind = ""
	StaticDeclarationContinuousPowerToughness             StaticDeclarationKind = "StaticDeclarationContinuousPowerToughness"
	StaticDeclarationContinuousBasePowerToughness         StaticDeclarationKind = "StaticDeclarationContinuousBasePowerToughness"
	StaticDeclarationContinuousCharacteristic             StaticDeclarationKind = "StaticDeclarationContinuousCharacteristic"
	StaticDeclarationContinuousEntryChoiceSubtype         StaticDeclarationKind = "StaticDeclarationContinuousEntryChoiceSubtype"
	StaticDeclarationChosenCreatureTypeTriggerMultiplier  StaticDeclarationKind = "StaticDeclarationChosenCreatureTypeTriggerMultiplier"
	StaticDeclarationKeywordGrant                         StaticDeclarationKind = "StaticDeclarationKeywordGrant"
	StaticDeclarationKeywordLoss                          StaticDeclarationKind = "StaticDeclarationKeywordLoss"
	StaticDeclarationRule                                 StaticDeclarationKind = "StaticDeclarationRule"
	StaticDeclarationCostModifier                         StaticDeclarationKind = "StaticDeclarationCostModifier"
	StaticDeclarationCardAbilityGrant                     StaticDeclarationKind = "StaticDeclarationCardAbilityGrant"
	StaticDeclarationPermanentAbilityGrant                StaticDeclarationKind = "StaticDeclarationPermanentAbilityGrant"
	StaticDeclarationControlGrant                         StaticDeclarationKind = "StaticDeclarationControlGrant"
	StaticDeclarationMonarchControlGrant                  StaticDeclarationKind = "StaticDeclarationMonarchControlGrant"
	StaticDeclarationPlayerRule                           StaticDeclarationKind = "StaticDeclarationPlayerRule"
	StaticDeclarationLoseAbilitiesBecome                  StaticDeclarationKind = "StaticDeclarationLoseAbilitiesBecome"
	StaticDeclarationOpponentActionRestriction            StaticDeclarationKind = "StaticDeclarationOpponentActionRestriction"
	StaticDeclarationSpellUncounterable                   StaticDeclarationKind = "StaticDeclarationSpellUncounterable"
	StaticDeclarationEnteringTriggerMultiplier            StaticDeclarationKind = "StaticDeclarationEnteringTriggerMultiplier"
	StaticDeclarationControlledTriggerMultiplier          StaticDeclarationKind = "StaticDeclarationControlledTriggerMultiplier"
	StaticDeclarationUntapDuringOtherUntapStep            StaticDeclarationKind = "StaticDeclarationUntapDuringOtherUntapStep"
	StaticDeclarationCharacteristicDefiningPowerToughness StaticDeclarationKind = "StaticDeclarationCharacteristicDefiningPowerToughness"
	StaticDeclarationCastAsThoughFlash                    StaticDeclarationKind = "StaticDeclarationCastAsThoughFlash"
	StaticDeclarationEnchantedTypeChange                  StaticDeclarationKind = "StaticDeclarationEnchantedTypeChange"
	StaticDeclarationEnterBattlefieldRestriction          StaticDeclarationKind = "StaticDeclarationEnterBattlefieldRestriction"
	StaticDeclarationContinuousQuotedAbilityGrant         StaticDeclarationKind = "StaticDeclarationContinuousQuotedAbilityGrant"
	StaticDeclarationAbilityCostSet                       StaticDeclarationKind = "StaticDeclarationAbilityCostSet"
	StaticDeclarationGraveyardCardKeywordGrant            StaticDeclarationKind = "StaticDeclarationGraveyardCardKeywordGrant"
	StaticDeclarationDrawLimit                            StaticDeclarationKind = "StaticDeclarationDrawLimit"
	StaticDeclarationCastLimit                            StaticDeclarationKind = "StaticDeclarationCastLimit"
	StaticDeclarationOpeningHandPlay                      StaticDeclarationKind = "StaticDeclarationOpeningHandPlay"
	StaticDeclarationOpponentEnteringTriggerSuppression   StaticDeclarationKind = "StaticDeclarationOpponentEnteringTriggerSuppression"
	StaticDeclarationCreatureAttackTax                    StaticDeclarationKind = "StaticDeclarationCreatureAttackTax"
	StaticDeclarationManaProductionMultiplier             StaticDeclarationKind = "StaticDeclarationManaProductionMultiplier"
	StaticDeclarationCombatDamagePrevention               StaticDeclarationKind = "StaticDeclarationCombatDamagePrevention"
)

// StaticAttackTaxAmountKind identifies how a per-creature attack-tax declaration
// (Baird, Archon of Absolution, Sphere of Safety, Collective Restraint) derives
// its per-attacker generic amount.
type StaticAttackTaxAmountKind string

// Per-creature attack-tax amount kinds recognized by the parser.
const (
	// StaticAttackTaxAmountFixed is a fixed generic amount printed as the cost
	// symbol ("pays {1} for each of those creatures.", Baird, Archon of
	// Absolution).
	StaticAttackTaxAmountFixed StaticAttackTaxAmountKind = "StaticAttackTaxAmountFixed"

	// StaticAttackTaxAmountEnchantments is the number of enchantments the
	// controller controls ("where X is the number of enchantments you control.",
	// Sphere of Safety).
	StaticAttackTaxAmountEnchantments StaticAttackTaxAmountKind = "StaticAttackTaxAmountEnchantments"

	// StaticAttackTaxAmountDomain is the number of basic land types among lands
	// the controller controls ("where X is the number of basic land types among
	// lands you control.", Collective Restraint, domain).
	StaticAttackTaxAmountDomain StaticAttackTaxAmountKind = "StaticAttackTaxAmountDomain"
)

// StaticDeclarationDynamicValueKind identifies the rules-derived count a
// characteristic-defining power/toughness declaration sets the source object's
// power and toughness equal to ("equal to the number of cards in your hand").
type StaticDeclarationDynamicValueKind string

// Static declaration characteristic-defining count kinds recognized by the
// parser. Each maps onto one runtime dynamic-value kind.
const (
	StaticDeclarationDynamicValueNone                         StaticDeclarationDynamicValueKind = ""
	StaticDeclarationDynamicValueControllerHandSize           StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerHandSize"
	StaticDeclarationDynamicValueControllerGraveyardSize      StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerGraveyardSize"
	StaticDeclarationDynamicValueControllerCreatureCount      StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerCreatureCount"
	StaticDeclarationDynamicValueControllerLandCount          StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerLandCount"
	StaticDeclarationDynamicValueControllerArtifactCount      StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerArtifactCount"
	StaticDeclarationDynamicValueAllBattlefieldCreatureCount  StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueAllBattlefieldCreatureCount"
	StaticDeclarationDynamicValueAllGraveyardsSize            StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueAllGraveyardsSize"
	StaticDeclarationDynamicValueCreatureCardsInAllGraveyards StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueCreatureCardsInAllGraveyards"
	StaticDeclarationDynamicValueCardTypesAmongAllGraveyards  StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueCardTypesAmongAllGraveyards"

	StaticDeclarationDynamicValueControllerCreatureCardsInGraveyard         StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerCreatureCardsInGraveyard"
	StaticDeclarationDynamicValueControllerInstantOrSorceryCardsInGraveyard StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerInstantOrSorceryCardsInGraveyard"
	StaticDeclarationDynamicValueControllerLandCardsInGraveyard             StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerLandCardsInGraveyard"
	StaticDeclarationDynamicValueControllerCardTypesInGraveyard             StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerCardTypesInGraveyard"
	StaticDeclarationDynamicValueControllerPermanentCardsInGraveyard        StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerPermanentCardsInGraveyard"
	StaticDeclarationDynamicValueControllerSubtypeCount                     StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerSubtypeCount"
	StaticDeclarationDynamicValueControllerBasicLandTypeCount               StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerBasicLandTypeCount"
	StaticDeclarationDynamicValueControllerLifeTotal                        StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerLifeTotal"
	StaticDeclarationDynamicValueAllPlayersHandSize                         StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueAllPlayersHandSize"
	StaticDeclarationDynamicValueControllerColorPermanentCount              StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerColorPermanentCount"
	StaticDeclarationDynamicValueControllerCardsDrawnThisTurn               StaticDeclarationDynamicValueKind = "StaticDeclarationDynamicValueControllerCardsDrawnThisTurn"
)

// StaticDeclarationSubjectKind identifies the affected group named by a typed
// static declaration. Group subjects carry their typed effect-subject value.
type StaticDeclarationSubjectKind string

// Static declaration subjects recognized by the parser.
const (
	StaticDeclarationSubjectUnknown        StaticDeclarationSubjectKind = ""
	StaticDeclarationSubjectSourceCreature StaticDeclarationSubjectKind = "StaticDeclarationSubjectSourceCreature"
	StaticDeclarationSubjectSourceSpell    StaticDeclarationSubjectKind = "StaticDeclarationSubjectSourceSpell"
	StaticDeclarationSubjectSourceNamed    StaticDeclarationSubjectKind = "StaticDeclarationSubjectSourceNamed"
	StaticDeclarationSubjectGroup          StaticDeclarationSubjectKind = "StaticDeclarationSubjectGroup"
	StaticDeclarationSubjectControllerHand StaticDeclarationSubjectKind = "StaticDeclarationSubjectControllerHand"
	StaticDeclarationSubjectController     StaticDeclarationSubjectKind = "StaticDeclarationSubjectController"
	StaticDeclarationSubjectEachPlayer     StaticDeclarationSubjectKind = "StaticDeclarationSubjectEachPlayer"
	// StaticDeclarationSubjectControllerGraveyard names the set of the
	// controller's graveyard cards a "[During your turn,] <filter> cards in your
	// graveyard have <keyword>." keyword-grant declaration affects (Six, Wrenn
	// and Six Emblem). CardFilter constrains the affected cards by card type.
	StaticDeclarationSubjectControllerGraveyard StaticDeclarationSubjectKind = "StaticDeclarationSubjectControllerGraveyard"
)

// StaticDeclarationPlayerRuleKind identifies the closed player-scoped rule a
// typed static declaration carries.
type StaticDeclarationPlayerRuleKind string

// Static declaration player rules recognized by the parser.
const (
	StaticDeclarationPlayerRuleUnknown           StaticDeclarationPlayerRuleKind = ""
	StaticDeclarationPlayerRuleNoMaximumHandSize StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleNoMaximumHandSize"
	// StaticDeclarationPlayerRuleSkipDrawStep makes the controller skip their draw
	// step ("Skip your draw step.", Necropotence, Yawgmoth's Bargain).
	StaticDeclarationPlayerRuleSkipDrawStep        StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleSkipDrawStep"
	StaticDeclarationPlayerRuleAttackTax           StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleAttackTax"
	StaticDeclarationPlayerRuleAdditionalLandPlays StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleAdditionalLandPlays"
	// StaticDeclarationPlayerRulePlayLandsFromGraveyard grants the controller a
	// continuous permission to play land cards from their graveyard ("You may
	// play lands from your graveyard.", Ramunap Excavator, Crucible of Worlds).
	StaticDeclarationPlayerRulePlayLandsFromGraveyard StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRulePlayLandsFromGraveyard"
	// StaticDeclarationPlayerRulePlayLandsFromLibraryTop grants the controller a
	// continuous permission to play land cards from the top of their library ("You
	// may play lands from the top of your library.", Oracle of Mul Daya, Courser of
	// Kruphix).
	StaticDeclarationPlayerRulePlayLandsFromLibraryTop StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRulePlayLandsFromLibraryTop"
	// StaticDeclarationPlayerRulePlayWithTopCardRevealed makes the controller play
	// with the top card of their library revealed ("Play with the top card of your
	// library revealed.", Oracle of Mul Daya, Courser of Kruphix, Future Sight).
	StaticDeclarationPlayerRulePlayWithTopCardRevealed StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRulePlayWithTopCardRevealed"
	// StaticDeclarationPlayerRuleCastSpellsFromLibraryTop grants the controller a
	// continuous permission to cast spells from the top of their library ("You may
	// cast spells from the top of your library.", Bolas's Citadel; "You may play
	// lands and cast spells from the top of your library.", Future Sight). The
	// optional CastSpellTypes filter restricts the castable spells by card type;
	// AlsoPlayLands additionally grants the land-play permission of the combined
	// "play lands and cast spells" wording.
	StaticDeclarationPlayerRuleCastSpellsFromLibraryTop StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleCastSpellsFromLibraryTop"
	// StaticDeclarationPlayerRuleCastThisFromGraveyard grants the controller a
	// continuous permission to cast the source card itself from their graveyard
	// ("You may cast this card from your graveyard.", Hogaak; "You may cast this
	// card from your graveyard as long as you control a Zombie.", Gravecrawler).
	// The permission is self-scoped to the source card and may carry an optional
	// "as long as <condition>" gate.
	StaticDeclarationPlayerRuleCastThisFromGraveyard StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleCastThisFromGraveyard"
	// StaticDeclarationPlayerRuleCastThisFromExile grants the controller a
	// continuous permission to cast the source card itself from exile ("You may
	// cast this card from exile.", Misthollow Griffin, Eternal Scourge). The
	// permission is self-scoped to the source card and may carry an optional "as
	// long as <condition>" gate.
	StaticDeclarationPlayerRuleCastThisFromExile StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleCastThisFromExile"
	// StaticDeclarationPlayerRulePlayAndCastFromExileWithCounter grants the
	// controller a continuous permission to play land cards and cast spells from
	// among the cards they own in exile that carry a named marker counter ("You
	// may play lands and cast spells from among cards you own in exile with croak
	// counters on them.", Grolnok, the Omnivore). ExileCounter names the required
	// counter so the runtime restricts the permission to exiled cards bearing it.
	StaticDeclarationPlayerRulePlayAndCastFromExileWithCounter StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRulePlayAndCastFromExileWithCounter"
	// StaticDeclarationPlayerRuleLookAtTopCardAnyTime lets the controller look at
	// the top card of their library at any time ("You may look at the top card of
	// your library any time.", Bolas's Citadel, Vizier of the Menagerie, Sphinx of
	// Jwar Isle). It is a private-visibility static: only the controller may see
	// the card.
	StaticDeclarationPlayerRuleLookAtTopCardAnyTime StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleLookAtTopCardAnyTime"
	// StaticDeclarationPlayerRuleLifeForColoredMana lets the controller pay 2 life
	// rather than a mana of ManaColor for each such colored symbol in a cost ("For
	// each {B} in a cost, you may pay 2 life rather than pay that mana.", K'rrik,
	// Son of Yawgmoth).
	StaticDeclarationPlayerRuleLifeForColoredMana StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleLifeForColoredMana"
	// StaticDeclarationPlayerRuleLifeForCommanderTax lets the controller pay 2
	// life rather than each {2} of the command-zone commander tax when casting the
	// source card itself ("Rather than pay {2} for each previous time you've cast
	// this spell from the command zone this game, pay 2 life that many times.",
	// Liesa, Shroud of Dusk).
	StaticDeclarationPlayerRuleLifeForCommanderTax StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleLifeForCommanderTax"
	// StaticDeclarationPlayerRuleHexproof grants the controller hexproof, the
	// player-scoped protection static "You have hexproof." (Aegis of the Gods,
	// Leyline of Sanctity, Spirit of the Hearth): the player can't be the target
	// of spells or abilities opponents control.
	StaticDeclarationPlayerRuleHexproof StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleHexproof"
	// StaticDeclarationPlayerRuleShroud grants the controller shroud, the
	// player-scoped protection static "You have shroud." (Ivory Mask, True
	// Believer): the player can't be the target of spells or abilities at all.
	StaticDeclarationPlayerRuleShroud StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleShroud"
	// StaticDeclarationPlayerRuleDamageDoesntCauseLifeLoss stops damage dealt to
	// the controller from reducing their life total ("damage doesn't cause you to
	// lose life", Archon of Coronation, usually gated "As long as you're the
	// monarch, ..."). The damage is still dealt; only the life-loss step is
	// skipped.
	StaticDeclarationPlayerRuleDamageDoesntCauseLifeLoss StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleDamageDoesntCauseLifeLoss"
	// StaticDeclarationPlayerRuleRedirectDamageToSource redirects all damage that
	// would be dealt to the controller to the ability's source permanent instead
	// ("All damage that would be dealt to you is dealt to this creature instead.",
	// Protector of the Crown). The redirect target is the source permanent, which
	// the runtime resolves from the rule effect's source.
	StaticDeclarationPlayerRuleRedirectDamageToSource StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleRedirectDamageToSource"
	// StaticDeclarationPlayerRuleActivateAbilitiesAsThoughHaste lets the controller
	// activate abilities of creatures they control as though those creatures had
	// haste ("You may activate abilities of creatures you control as though those
	// creatures had haste.", Thousand-Year Elixir, Shang-Chi, Tyvar). It removes
	// the summoning-sickness restriction on {T} and {Q} activation costs for the
	// controller's creatures; it is an activation permission only and does not let
	// a summoning-sick creature attack.
	StaticDeclarationPlayerRuleActivateAbilitiesAsThoughHaste StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleActivateAbilitiesAsThoughHaste"
)

// StaticDeclarationCardFilterKind identifies the closed card filter that a
// controller-hand subject constrains.
type StaticDeclarationCardFilterKind string

// Static declaration card filters recognized by the parser.
const (
	StaticDeclarationCardFilterNone     StaticDeclarationCardFilterKind = ""
	StaticDeclarationCardFilterLand     StaticDeclarationCardFilterKind = "StaticDeclarationCardFilterLand"
	StaticDeclarationCardFilterCreature StaticDeclarationCardFilterKind = "StaticDeclarationCardFilterCreature"
	StaticDeclarationCardFilterHistoric StaticDeclarationCardFilterKind = "StaticDeclarationCardFilterHistoric"
	// StaticDeclarationCardFilterNonlandPermanent selects nonland permanent cards
	// ("nonland permanent cards in your graveyard", Six): cards whose type line
	// includes a permanent card type other than Land.
	StaticDeclarationCardFilterNonlandPermanent StaticDeclarationCardFilterKind = "StaticDeclarationCardFilterNonlandPermanent"
	// StaticDeclarationCardFilterPermanent selects permanent cards ("permanent
	// cards in your graveyard"): cards whose type line includes any permanent
	// card type.
	StaticDeclarationCardFilterPermanent StaticDeclarationCardFilterKind = "StaticDeclarationCardFilterPermanent"
	// StaticDeclarationCardFilterInstantOrSorcery selects instant and sorcery
	// cards ("instant and sorcery cards in your graveyard", Wrenn and Six
	// Emblem).
	StaticDeclarationCardFilterInstantOrSorcery StaticDeclarationCardFilterKind = "StaticDeclarationCardFilterInstantOrSorcery"
)

// StaticDeclarationCostModifierKind identifies the closed cost-modifier shape a
// typed static declaration carries.
type StaticDeclarationCostModifierKind string

// Static declaration cost-modifier shapes recognized by the parser.
const (
	StaticDeclarationCostModifierUnknown          StaticDeclarationCostModifierKind = ""
	StaticDeclarationCostModifierAbilityReduction StaticDeclarationCostModifierKind = "StaticDeclarationCostModifierAbilityReduction"
	StaticDeclarationCostModifierReplaceCost      StaticDeclarationCostModifierKind = "StaticDeclarationCostModifierReplaceCost"
	StaticDeclarationCostModifierReplaceFirstCost StaticDeclarationCostModifierKind = "StaticDeclarationCostModifierReplaceFirstCost"
	StaticDeclarationCostModifierSpellReduction   StaticDeclarationCostModifierKind = "StaticDeclarationCostModifierSpellReduction"
	StaticDeclarationCostModifierSpellIncrease    StaticDeclarationCostModifierKind = "StaticDeclarationCostModifierSpellIncrease"
	// StaticDeclarationCostModifierSpellSharedExiledTypeReduction is the dynamic
	// controller cast-cost discount that scales with the card types the spell
	// shares with the cards exiled with the source permanent ("Spells you cast
	// cost {N} less to cast for each card type they share with cards exiled with
	// this creature.", Cemetery Prowler). CostReductionAmount carries the per-
	// shared-type amount.
	StaticDeclarationCostModifierSpellSharedExiledTypeReduction StaticDeclarationCostModifierKind = "StaticDeclarationCostModifierSpellSharedExiledTypeReduction"
	// StaticDeclarationCostModifierSpellPerObjectReduction is the dynamic group
	// cast-cost discount that scales with a countable battlefield permanent the
	// controller controls ("[<filter>] spells you cast cost {N} less to cast for
	// each <permanent> you control[ with power M or greater]."). Temur
	// Battlecrier ("During your turn, spells you cast cost {1} less to cast for
	// each creature you control with power 4 or greater.") and Hamza, Guardian
	// of Arashin ("Creature spells you cast cost {1} less to cast for each
	// creature you control with a +1/+1 counter on it.") are the targets.
	// CostReductionAmount carries the per-permanent amount and
	// PerObjectCountSelection carries the typed battlefield count subject.
	StaticDeclarationCostModifierSpellPerObjectReduction StaticDeclarationCostModifierKind = "StaticDeclarationCostModifierSpellPerObjectReduction"
)

// StaticDeclarationSpellTypeKind identifies the closed spell-type filter a
// controller cast-cost modifier constrains.
type StaticDeclarationSpellTypeKind string

// Static declaration spell-type filters recognized by the parser.
const (
	StaticDeclarationSpellTypeAll              StaticDeclarationSpellTypeKind = ""
	StaticDeclarationSpellTypeArtifact         StaticDeclarationSpellTypeKind = "StaticDeclarationSpellTypeArtifact"
	StaticDeclarationSpellTypeCreature         StaticDeclarationSpellTypeKind = "StaticDeclarationSpellTypeCreature"
	StaticDeclarationSpellTypeEnchantment      StaticDeclarationSpellTypeKind = "StaticDeclarationSpellTypeEnchantment"
	StaticDeclarationSpellTypeInstant          StaticDeclarationSpellTypeKind = "StaticDeclarationSpellTypeInstant"
	StaticDeclarationSpellTypeSorcery          StaticDeclarationSpellTypeKind = "StaticDeclarationSpellTypeSorcery"
	StaticDeclarationSpellTypeInstantOrSorcery StaticDeclarationSpellTypeKind = "StaticDeclarationSpellTypeInstantOrSorcery"
)

// StaticDeclarationSpellColorKind identifies the closed single-color filter a
// controller cast-cost modifier constrains. It is mutually exclusive with the
// spell-type filter: a declaration carries at most one of the two.
type StaticDeclarationSpellColorKind string

// Static declaration spell-color filters recognized by the parser.
const (
	StaticDeclarationSpellColorNone      StaticDeclarationSpellColorKind = ""
	StaticDeclarationSpellColorWhite     StaticDeclarationSpellColorKind = "StaticDeclarationSpellColorWhite"
	StaticDeclarationSpellColorBlue      StaticDeclarationSpellColorKind = "StaticDeclarationSpellColorBlue"
	StaticDeclarationSpellColorBlack     StaticDeclarationSpellColorKind = "StaticDeclarationSpellColorBlack"
	StaticDeclarationSpellColorRed       StaticDeclarationSpellColorKind = "StaticDeclarationSpellColorRed"
	StaticDeclarationSpellColorGreen     StaticDeclarationSpellColorKind = "StaticDeclarationSpellColorGreen"
	StaticDeclarationSpellColorColorless StaticDeclarationSpellColorKind = "StaticDeclarationSpellColorColorless"
)

// StaticDeclarationSpellCasterKind identifies which players' spells a cast-cost
// modifier affects ("Spells you cast ..." vs "Spells your opponents cast ..."
// vs "Spells that ..."). The empty kind is the default controller scope.
type StaticDeclarationSpellCasterKind string

// Static declaration spell-caster filters recognized by the parser.
const (
	StaticDeclarationSpellCasterController StaticDeclarationSpellCasterKind = ""
	StaticDeclarationSpellCasterOpponents  StaticDeclarationSpellCasterKind = "StaticDeclarationSpellCasterOpponents"
	StaticDeclarationSpellCasterAny        StaticDeclarationSpellCasterKind = "StaticDeclarationSpellCasterAny"
)

// StaticDeclarationCastZoneKind identifies a non-hand zone that a cast-zone
// restriction forbids the affected players from casting spells out of, or that a
// cast-cost modifier scopes its discount to.
type StaticDeclarationCastZoneKind string

// Static declaration cast-zone restriction zones recognized by the parser.
const (
	StaticDeclarationCastZoneGraveyard StaticDeclarationCastZoneKind = "StaticDeclarationCastZoneGraveyard"
	StaticDeclarationCastZoneLibrary   StaticDeclarationCastZoneKind = "StaticDeclarationCastZoneLibrary"
	StaticDeclarationCastZoneExile     StaticDeclarationCastZoneKind = "StaticDeclarationCastZoneExile"
	StaticDeclarationCastZoneCommand   StaticDeclarationCastZoneKind = "StaticDeclarationCastZoneCommand"
	// StaticDeclarationCastZoneNonHand scopes a cast-cost modifier to spells cast
	// from any zone other than the caster's hand ("Spells you cast from anywhere
	// other than your hand cost {N} less to cast.", Sage of the Beyond). Lowering
	// expands it to the concrete non-hand cast zones (graveyard, exile, library,
	// and command).
	StaticDeclarationCastZoneNonHand StaticDeclarationCastZoneKind = "StaticDeclarationCastZoneNonHand"
)

// StaticDeclarationEnterFilterKind identifies which entering cards an
// enter-the-battlefield zone restriction ("<filter> cards in graveyards can't
// enter the battlefield.") affects.
type StaticDeclarationEnterFilterKind string

// Static declaration enter-restriction card filters recognized by the parser.
const (
	StaticDeclarationEnterFilterCreature         StaticDeclarationEnterFilterKind = "StaticDeclarationEnterFilterCreature"
	StaticDeclarationEnterFilterPermanent        StaticDeclarationEnterFilterKind = "StaticDeclarationEnterFilterPermanent"
	StaticDeclarationEnterFilterNonlandPermanent StaticDeclarationEnterFilterKind = "StaticDeclarationEnterFilterNonlandPermanent"
)

// StaticDeclarationSubject is a source-spanned typed affected group.
type StaticDeclarationSubject struct {
	Kind       StaticDeclarationSubjectKind    `json:",omitempty"`
	Span       shared.Span                     `json:"-"`
	Group      EffectStaticSubjectSyntax       `json:",omitzero"`
	CardFilter StaticDeclarationCardFilterKind `json:",omitempty"`
	// GroupSelection carries a full parsed selection for a controlled-permanent
	// group whose shape the enumerated Group vocabulary cannot express, most
	// notably a union of alternatives ("nontoken artifact creatures and Vehicles
	// you control"). When set, the subject Kind is StaticDeclarationSubjectGroup
	// and downstream layers project this selection through the shared
	// CompiledSelector machinery instead of the enumerated Group. It is
	// parser-internal, so it is excluded from serialization.
	GroupSelection *SelectionSyntax `json:"-"`
}

// StaticGrantedManaAbilitySyntax is one typed activated mana ability quoted by
// a static permanent-ability grant.
type StaticGrantedManaAbilitySyntax struct {
	Span     shared.Span `json:"-"`
	TapCost  bool        `json:",omitempty"`
	Amount   int         `json:",omitempty"`
	AnyColor bool        `json:",omitempty"`
	// Text is the exact quoted ability source text without its surrounding
	// quotes, carried so downstream layers reproduce the granted ability's
	// printed wording without re-deriving it from typed fields.
	Text string `json:",omitempty"`
	// Sacrifice marks the "Sacrifice this artifact" additional cost carried by
	// the Treasure-style granted ability.
	Sacrifice bool `json:",omitempty"`
	// AnyOneColor marks the "Add <N> mana of any one color" output, where the
	// controller chooses one color and adds Amount mana of it (Amount >= 2).
	// It is mutually exclusive with AnyColor.
	AnyOneColor bool `json:",omitempty"`
	// Colorless marks the bare "{T}: Add {C}" ability that adds one colorless
	// mana. It is mutually exclusive with AnyColor and AnyOneColor.
	Colorless bool `json:",omitempty"`
}

// StaticGrantedAbilitySyntax is one full quoted ability (a triggered or
// activated ability) a static declaration grants to its subject ("Equipped
// creature has '<quoted ability>'."). The parser parses the quoted body once
// through the same pipeline so downstream layers lower the granted ability from
// the typed inner document rather than re-parsing its Oracle wording.
type StaticGrantedAbilitySyntax struct {
	Span shared.Span `json:"-"`
	// Text is the exact quoted ability source text without its surrounding
	// quotes, carried so downstream layers reproduce the granted ability's
	// printed wording without re-deriving it from typed fields.
	Text        string `json:",omitempty"`
	document    Document
	diagnostics []shared.Diagnostic
}

// Inner returns the parsed inner document of the granted ability together with
// the diagnostics its inner parse produced. Consumers lower the typed inner
// document instead of re-parsing the granted ability's Oracle text.
func (s *StaticGrantedAbilitySyntax) Inner() (Document, []shared.Diagnostic) {
	return s.document, s.diagnostics
}

// StaticDeclarationSyntax is one composable typed static declaration. The
// compiler maps these onto its semantic vocabulary mechanically; it inspects no
// Oracle source text to derive meaning.
type StaticDeclarationSyntax struct {
	Kind          StaticDeclarationKind    `json:",omitempty"`
	Span          shared.Span              `json:"-"`
	OperationSpan shared.Span              `json:"-"`
	Subject       StaticDeclarationSubject `json:",omitzero"`

	// HasCondition records whether a single supported-shaped condition clause
	// applies to this declaration; ConditionSpan links to that clause.
	HasCondition  bool        `json:",omitempty"`
	ConditionSpan shared.Span `json:"-"`

	// Continuous power/toughness payload.
	PowerDelta     SignedAmountSyntax `json:",omitzero"`
	ToughnessDelta SignedAmountSyntax `json:",omitzero"`
	Dynamic        bool               `json:",omitempty"`

	// Continuous base power/toughness (characteristic-setting) payload.
	BasePower     int  `json:",omitempty"`
	BaseToughness int  `json:",omitempty"`
	BasePTSet     bool `json:",omitempty"`

	// Characteristic-defining power/toughness payload: the rules-derived count
	// a "<source>'s power and toughness are each equal to <count>" declaration
	// sets the source object's power and toughness equal to.
	DynamicValue StaticDeclarationDynamicValueKind `json:",omitempty"`

	// DynamicValueSubtype carries the subtype counted by a
	// StaticDeclarationDynamicValueControllerSubtypeCount declaration ("the
	// number of Swamps you control", "the number of Goblins you control"). It is
	// empty for every other count kind.
	DynamicValueSubtype types.Sub `json:"-"`

	// DynamicValueColor carries the color counted by a
	// StaticDeclarationDynamicValueControllerColorPermanentCount declaration
	// ("the number of red permanents you control"). It is empty for every other
	// count kind.
	DynamicValueColor Color `json:"-"`

	// DynamicSetsPower and DynamicSetsToughness record which characteristics a
	// characteristic-defining power/toughness declaration sets. "power and
	// toughness are each equal to" sets both; "power is equal to" sets power
	// only (the printed toughness stands); the Tarmogoyf form "power is equal to
	// <count> and its toughness is equal to that number plus N" sets both with a
	// toughness offset.
	DynamicSetsPower       bool `json:",omitempty"`
	DynamicSetsToughness   bool `json:",omitempty"`
	DynamicToughnessOffset int  `json:",omitempty"`

	// LoseAllAbilities marks a StaticDeclarationLoseAbilitiesBecome declaration
	// whose affected object loses all abilities ("loses all abilities"). For that
	// kind Colors, CardTypes, and Subtypes are SET (replacing the object's
	// existing colors, card types, and creature types) rather than added.
	LoseAllAbilities bool `json:",omitempty"`

	// Continuous characteristic addition payload: the colors, card types, and
	// subtypes a "<group> is/are ... in addition to ..." declaration grants. A
	// bare "<group> is/are <color>" with no "in addition" tail sets colors and
	// leaves ColorsAdd false; an explicit "in addition to its other colors" tail
	// sets ColorsAdd. Card types and subtypes are always additive.
	Colors    []Color     `json:"-"`
	CardTypes []CardType  `json:"-"`
	Subtypes  []types.Sub `json:"-"`
	ColorsAdd bool        `json:",omitempty"`
	// EveryCreatureType marks a "<group> is/are every creature type" continuous
	// characteristic declaration (Maskwood Nexus, Mistform Ultimus). It adds
	// every creature subtype at the type layer (CR 702.73) and is mutually
	// exclusive with the enumerated Colors/CardTypes/Subtypes payload.
	EveryCreatureType bool `json:",omitempty"`

	// EveryBasicLandType marks a "<group> is/are every basic land type [in
	// addition to their other types]" declaration (Dryad of the Ilysian Grove,
	// Prismatic Omen). The compiler expands it to a LayerType continuous effect
	// that adds all five basic land subtypes rather than enumerating them in the
	// characteristic list.
	EveryBasicLandType bool `json:",omitempty"`

	// Keyword-grant and card-ability-grant payload: the spans of the granted
	// keyword atoms in source order.
	KeywordSpans []shared.Span `json:"-"`

	// Permanent-ability-grant payload.
	GrantedManaAbility *StaticGrantedManaAbilitySyntax `json:",omitempty"`

	// Rule payload.
	Rule StaticRuleSyntax `json:",omitzero"`

	// Cost-modifier payload.
	CostModifier        StaticDeclarationCostModifierKind `json:",omitempty"`
	CostReductionAmount int                               `json:",omitempty"`
	// CostIncreaseColors lists the colored mana symbols a cast-cost increase
	// adds on top of any generic amount ("Black spells you cast cost {B} more to
	// cast.", Derelor and the mono-color Leech cycle). Each entry is one basic
	// colored mana symbol the parser read from the "cost {C} more to cast"
	// clause; it is set only with StaticDeclarationCostModifierSpellIncrease. An
	// empty slice means the increase adds no colored mana.
	CostIncreaseColors []mana.Color `json:",omitempty"`
	// CostIncreaseLife carries the life a cast-cost tax adds as an additional
	// cost ("Spells your opponents cast that target this creature cost an
	// additional 3 life to cast.", Terror of the Peaks). It is a positive count
	// of life set only with StaticDeclarationCostModifierSpellIncrease; the
	// caster pays that much life on top of the spell's mana cost. A zero value
	// adds no life tax.
	CostIncreaseLife int    `json:",omitempty"`
	CostReplacement  string `json:",omitempty"`
	// AbilityCostKeyword names the activated-ability keyword whose cost a
	// StaticDeclarationAbilityCostSet declaration sets ("Equipment you control
	// have equip {0}." sets the Equip ability cost). CostReplacement carries the
	// canonical replacement mana cost (an empty string is the free {0} cost).
	AbilityCostKeyword KeywordKind                     `json:",omitempty"`
	SpellType          StaticDeclarationSpellTypeKind  `json:",omitempty"`
	SpellColor         StaticDeclarationSpellColorKind `json:",omitempty"`
	ChosenCreatureType bool                            `json:",omitempty"`

	// SpellCastZone scopes a cast-cost modifier to spells cast from a single
	// non-hand zone ("Spells you cast from your graveyard cost {N} less to
	// cast."). The empty kind applies no zone filter, so the modifier affects the
	// controller's spells cast from any zone.
	SpellCastZone StaticDeclarationCastZoneKind `json:",omitempty"`

	// SpellPowerAtLeast carries the base-power threshold of a cast-cost modifier
	// filtered by power ("Creature spells you cast with power 4 or greater cost
	// {2} less to cast.", Goreclaw): a spell matches only when its printed power
	// is greater than or equal to this value. MatchSpellPowerAtLeast marks the
	// threshold present so a zero threshold stays expressible. It combines with
	// the spell-type, color, subtype, and zone filters.
	SpellPowerAtLeast      int  `json:",omitempty"`
	MatchSpellPowerAtLeast bool `json:",omitempty"`

	// SpellManaValueAtLeast carries the mana-value threshold of a cast-cost
	// modifier filtered by mana value ("Creature spells you cast with mana value
	// 6 or greater cost {2} less to cast.", Krosan Drover): a spell matches only
	// when its mana value is greater than or equal to this value.
	// MatchSpellManaValueAtLeast marks the threshold present so a zero threshold
	// stays expressible. It is mutually exclusive with the power threshold and
	// combines with the spell-type, color, subtype, and zone filters.
	SpellManaValueAtLeast      int  `json:",omitempty"`
	MatchSpellManaValueAtLeast bool `json:",omitempty"`

	// PerObjectCountSelection carries the typed battlefield count subject of a
	// StaticDeclarationCostModifierSpellPerObjectReduction modifier ("... for
	// each creature you control with power 4 or greater."): the controller's
	// matching battlefield permanents are counted at cost time, each worth
	// CostReductionAmount generic mana. It is nil for every other modifier shape.
	PerObjectCountSelection *SelectionSyntax `json:"-"`

	// SpellCaster scopes a cast-cost modifier to a set of casting players
	// ("Spells you cast ..." vs "Spells your opponents cast ..." vs "Spells
	// that ..."). The empty kind is the default controller scope.
	// SpellTargetsSource marks the "that target <source>" predicate, restricting
	// the modifier to spells that target the source permanent ("Spells your
	// opponents cast that target this creature cost {2} more to cast.").
	SpellCaster        StaticDeclarationSpellCasterKind `json:",omitempty"`
	SpellTargetsSource bool                             `json:",omitempty"`

	// SpellColors lists the colors of a cast-cost modifier's color disjunction
	// ("Each spell you cast that's red or green ..." / "Blue spells and red
	// spells you cast ..."): a spell matches when it has any one of these
	// colors. It carries two or more real colors and is mutually exclusive with
	// SpellColor and the spell-type filter.
	SpellColors []StaticDeclarationSpellColorKind `json:"-"`

	// SpellSubtypes lists the subtype filter of a cast-cost modifier ("Aura and
	// Equipment spells you cast ..."): a spell matches when it has any one of
	// these subtypes. It combines with SpellColor (an optional leading color
	// word) and is mutually exclusive with the SpellType single-card-type filter
	// and the SpellColors color disjunction.
	SpellSubtypes []types.Sub `json:"-"`

	// SpellExcludedTypes lists card types a cast-cost modifier exempts, set by a
	// "non"-prefixed card-type filter word ("Noncreature spells cost {1} more to
	// cast.", Thalia, Guardian of Thraben; "Nonartifact spells cost {1} more to
	// cast.", Lodestone Golem): a spell matches only when it carries none of
	// these card types. It is mutually exclusive with the SpellType single-type
	// filter, the SpellSubtypes subtype filter, and the color filters.
	SpellExcludedTypes []CardType `json:"-"`

	// SpellRequiredTypes lists the card types of a cast-cost modifier whose
	// card-type filter is a coordinated list of two or more types joined by
	// "and" ("Artifact and enchantment spells your opponents cast cost {2} more
	// to cast.", Aura of Silence; "Artifact, instant, and sorcery spells your
	// opponents cast cost {1} more to cast.", Dovin, Hand of Control): a spell
	// matches when it carries any one of the listed types. It carries two or more
	// real card types and is mutually exclusive with the SpellType single-type
	// filter, the SpellSubtypes subtype filter, the SpellExcludedTypes exclusion,
	// and the color filters.
	SpellRequiredTypes []CardType `json:"-"`

	// Player-rule payload: the closed player-scoped rule this declaration grants
	// to the static ability's controller.
	PlayerRule          StaticDeclarationPlayerRuleKind `json:",omitempty"`
	AttackTaxGeneric    int                             `json:",omitempty"`
	AdditionalLandPlays int                             `json:",omitempty"`

	// ExileCounter names the marker counter a
	// StaticDeclarationPlayerRulePlayAndCastFromExileWithCounter declaration
	// requires on the exiled cards its play/cast permission covers ("... in exile
	// with croak counters on them.", Grolnok, the Omnivore). It is the zero value
	// for every other declaration kind.
	ExileCounter counter.Kind `json:"-"`

	// ExilePlayOncePerTurn, ExilePlayExiledByControlledAbility, and
	// ExilePlaySpendAnyColorMana carry the optional riders of a
	// StaticDeclarationPlayerRulePlayAndCastFromExileWithCounter declaration
	// spelled as "play a card from exile" ("Once each turn, you may play a card
	// from exile with a collection counter on it if it was exiled by an ability
	// you controlled, and you may spend mana as though it were mana of any color to
	// cast it.", Evelyn, the Covetous): a once-per-turn use cap, a provenance
	// filter restricting the permission to cards the controller's own ability
	// exiled, and an any-color mana permission for casting. Each is independent, so
	// wordings that carry only some of the riders lower correctly. They are false
	// for every other declaration kind.
	ExilePlayOncePerTurn               bool `json:",omitempty"`
	ExilePlayExiledByControlledAbility bool `json:",omitempty"`
	ExilePlaySpendAnyColorMana         bool `json:",omitempty"`

	// Per-creature attack-tax payload (Baird, Archon of Absolution, Sphere of
	// Safety, Collective Restraint): AttackTaxAmountKind selects how the
	// per-attacker generic amount is derived (a fixed AttackTaxGeneric, the
	// controller's enchantment count, or domain). AttackTaxIncludesPlaneswalkers
	// records the "or planeswalkers you control" scope. They are unset for every
	// other declaration kind.
	AttackTaxAmountKind            StaticAttackTaxAmountKind `json:",omitempty"`
	AttackTaxIncludesPlaneswalkers bool                      `json:",omitempty"`

	// ManaMultiplier carries the factor of a
	// StaticDeclarationManaProductionMultiplier declaration ("If you tap a
	// permanent for mana, it produces twice as much of that mana instead.", Mana
	// Reflection, 2; "... three times as much ...", Nyxbloom Ancient, 3). It is at
	// least 2 for that kind and zero for every other declaration kind.
	ManaMultiplier int `json:",omitempty"`

	// ManaColor carries the colored mana symbol of a
	// StaticDeclarationPlayerRuleLifeForColoredMana declaration ("For each {B} in
	// a cost, ..."). It is empty for every other player rule.
	ManaColor mana.Color `json:"-"`

	// Opponent action-restriction payload: a continuous prohibition stopping the
	// affected players from casting spells and/or activating abilities of
	// permanents whose type is in RestrictActivateTypes. RestrictAffectsAllPlayers
	// selects every player ("Players can't ...") rather than only opponents;
	// RestrictDuringControllerTurn scopes the prohibition to the controller's turn.
	RestrictCastSpells           bool       `json:",omitempty"`
	RestrictActivateTypes        []CardType `json:"-"`
	RestrictAffectsAllPlayers    bool       `json:",omitempty"`
	RestrictDuringControllerTurn bool       `json:",omitempty"`

	// Cast-zone restriction payload: when the cast prohibition is scoped to a set
	// of source zones ("... can't cast spells from graveyards or libraries.") the
	// zones are listed in RestrictCastFromZones. RestrictCastOnlyFromHand records
	// the complementary "... can't cast spells from anywhere other than their
	// hands." form (Drannith Magistrate), which forbids every non-hand cast zone.
	RestrictCastFromZones    []StaticDeclarationCastZoneKind `json:"-"`
	RestrictCastOnlyFromHand bool                            `json:",omitempty"`

	// Draw-limit payload: a continuous per-turn draw cap stopping the affected
	// players from drawing more than DrawLimit cards each turn ("Each opponent
	// can't draw more than one card each turn.", Narset, Parter of Veils).
	// DrawLimitAffectsAllPlayers selects every player ("Each player can't ...",
	// Spirit of the Labyrinth); DrawLimitAffectsController selects only the
	// controller ("You can't ..."). With neither flag set the cap affects only the
	// controller's opponents.
	DrawLimit                  int  `json:",omitempty"`
	DrawLimitAffectsAllPlayers bool `json:",omitempty"`
	DrawLimitAffectsController bool `json:",omitempty"`

	// Cast-limit payload: a continuous per-turn spell cap stopping the affected
	// players from casting more than CastLimit spells each turn ("Each player
	// can't cast more than one spell each turn.", Rule of Law, Eidolon of
	// Rhetoric, Arcane Laboratory). CastLimitAffectsAllPlayers selects every
	// player ("Each player"/"Players"); CastLimitAffectsController selects only
	// the controller ("You"). With neither flag set the cap affects only the
	// controller's opponents.
	CastLimit                  int  `json:",omitempty"`
	CastLimitAffectsAllPlayers bool `json:",omitempty"`
	CastLimitAffectsController bool `json:",omitempty"`

	// Entering-trigger-multiplier payload: the entering permanent's card-type
	// filter for an "If <filter> entering causes a triggered ability of a
	// permanent you control to trigger, that ability triggers an additional
	// time." declaration. An empty EnteringFilterTypes matches any entering
	// permanent ("a permanent").
	EnteringFilterTypes []CardType `json:"-"`

	// Controlled-trigger-multiplier payload: the source permanent's filter for an
	// "If a triggered ability of <filter> you control triggers, that ability
	// triggers an additional time." declaration (Annie Joins Up, Katara, the
	// Fearless, Splinter, Radical Rat). Each branch carries a conjunctive
	// type/supertype filter and a disjunctive subtype filter; a leading "another"
	// sets ExcludeSelf. Multiple branches model an "or"-joined filter ("a Shaman
	// or another Wizard", Harmonic Prodigy). At least one branch is present and
	// every branch names at least one subtype or card type.
	ControlledFilterBranches []ControlledTriggerSourceFilter `json:"-"`

	// Untap-during-other-players'-untap-step payload: the filtered set of the
	// controller's permanents that gain an extra untap during each other
	// player's (or opponent's) untap step.
	UntapGroup StaticUntapGroupKind `json:",omitempty"`

	// Cast-from-library-top payload: the card-type filter restricting which
	// spells the controller may cast from the top of their library ("You may cast
	// creature spells from the top of your library."). An empty CastSpellTypes
	// permits casting any spell. CastColorless additionally permits casting
	// colorless spells, recording the "colorless spells" clause that may stand
	// alone ("You may cast colorless spells from the top of your library.") or
	// combine with a card-type clause ("You may cast artifact spells and colorless
	// spells from the top of your library.", Mystic Forge); a spell qualifies when
	// it matches the card-type filter or is colorless. AlsoPlayLands records the
	// combined "play lands and cast spells from the top of your library." wording,
	// which additionally grants the land-play permission. CastChosenCreatureType
	// records a trailing "of the chosen type" qualifier, narrowing the permission
	// to spells sharing the creature subtype the source permanent chose as it
	// entered ("creature spells of the chosen type from the top of your library.",
	// Realmwalker).
	CastSpellTypes         []CardType `json:"-"`
	CastColorless          bool       `json:",omitempty"`
	AlsoPlayLands          bool       `json:",omitempty"`
	CastChosenCreatureType bool       `json:",omitempty"`
	// CastPayLifeManaValue records a trailing "If you cast a spell this way, pay
	// life equal to its mana value rather than pay its mana cost." rider on a
	// cast-from-library-top declaration ("You may play lands and cast spells from
	// the top of your library. If you cast a spell this way, pay life equal to its
	// mana value rather than pay its mana cost.", Bolas's Citadel), so lowering
	// makes spells cast this way pay life equal to their mana value instead of
	// their mana cost.
	CastPayLifeManaValue bool `json:",omitempty"`

	// FlashSpellType and FlashSpellSubtypes carry the optional spell filter of a
	// StaticDeclarationCastAsThoughFlash declaration ("You may cast sorcery
	// spells as though they had flash.", "You may cast Aura and Equipment spells
	// as though they had flash."). An empty FlashSpellType and FlashSpellSubtypes
	// grant the permission for every spell ("You may cast spells as though they
	// had flash.").
	FlashSpellType     StaticDeclarationSpellTypeKind `json:",omitempty"`
	FlashSpellSubtypes []types.Sub                    `json:"-"`
	// UncounterableSpellSubtypes carries the optional subtype filter of a
	// StaticDeclarationSpellUncounterable declaration ("Human spells you control
	// can't be countered."). An empty slice with SpellType affects spells by card
	// type only ("Creature spells you control can't be countered.").
	UncounterableSpellSubtypes []types.Sub `json:"-"`
	// PreventionRecipient carries the recipient group of a
	// StaticDeclarationCombatDamagePrevention declaration ("Prevent all combat
	// damage that would be dealt to attacking Humans you control."). It is the
	// parsed selection describing which permanents the prevention protects.
	PreventionRecipient SelectionSyntax `json:"-"`
	// Enchanted-type-change payload: a removal Aura whose continuous effect sets
	// the enchanted permanent's card types and creature subtypes (CardTypes,
	// Subtypes, SET), optionally makes it colorless (BecomeColorless), optionally
	// grants a single mana ability (GrantedManaAbility), and optionally strips its
	// other abilities (LoseAllAbilities). Backs "Enchanted permanent is a
	// colorless Forest land." (Song of the Dryads) and "Enchanted permanent is a
	// colorless land with '{T}: Add {C}' and loses all other card types and
	// abilities." (Imprisoned in the Moon).
	BecomeColorless bool `json:",omitempty"`
	// Enter-the-battlefield zone-restriction payload: an
	// EnterRestrictFilter-filtered set of cards cannot enter the battlefield out
	// of the zones in EnterRestrictFromZones ("Creature cards in graveyards and
	// libraries can't enter the battlefield."). The restriction is global.
	EnterRestrictFilter    StaticDeclarationEnterFilterKind `json:",omitempty"`
	EnterRestrictFromZones []StaticDeclarationCastZoneKind  `json:"-"`

	// GrantedAbility carries the quoted triggered/activated ability body a
	// StaticDeclarationContinuousQuotedAbilityGrant confers on its subject
	// ("Equipped creature has '<quoted ability>'."). The parser parses the
	// quoted text once so downstream layers lower the granted ability from the
	// typed inner document instead of re-parsing its Oracle wording.
	GrantedAbility *StaticGrantedAbilitySyntax `json:",omitempty"`
}

// StaticUntapGroupKind identifies the closed group of the controller's
// permanents an "Untap <group> you control during each other player's untap
// step." declaration untaps.
type StaticUntapGroupKind string

// Static untap-step group filters recognized by the parser.
const (
	StaticUntapGroupNone       StaticUntapGroupKind = ""
	StaticUntapGroupSelf       StaticUntapGroupKind = "StaticUntapGroupSelf"
	StaticUntapGroupPermanents StaticUntapGroupKind = "StaticUntapGroupPermanents"
	StaticUntapGroupCreatures  StaticUntapGroupKind = "StaticUntapGroupCreatures"
	StaticUntapGroupArtifacts  StaticUntapGroupKind = "StaticUntapGroupArtifacts"
	StaticUntapGroupLands      StaticUntapGroupKind = "StaticUntapGroupLands"
)

func emitStaticDeclarations(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if ability.Modal != nil {
			continue
		}
		body := staticDeclarationBodyTokens(ability)
		if len(body) == 0 {
			continue
		}
		declarations := parseStaticDeclarations(body, ability.Quoted, ability.Atoms, ability.ConditionClauses)
		if len(declarations) > 0 {
			ability.StaticDeclarations = declarations
			dropControllerTurnConditionsNativelyConsumed(ability, declarations)
			dropExilePlayProvenanceConditionNativelyConsumed(ability, declarations)
			foldStaticCastFromTopPayLifeRider(ability, declarations)
			foldStaticSkipDrawStep(ability, declarations)
			foldStaticDamageDoesntCauseLifeLoss(ability, declarations)
			foldStaticRedirectDamageToSource(ability, declarations)
			foldStaticRemoveAuraRider(ability, declarations)
		}
	}
}

// dropControllerTurnConditionsNativelyConsumed removes the controller-turn
// condition ("During your turn,") the generic scanner emits when a static
// declaration on the same ability already encodes the controller's-turn scope
// itself: the graveyard-card keyword grant and the opponent action restriction
// both strip the "During your turn," prefix into RestrictDuringControllerTurn.
// Both the condition boundary (which feeds the compiler's condition segments)
// and the matching clause are dropped, so the prefix surfaces exactly once,
// through the consuming declaration, restoring those cards' prior behavior.
func dropControllerTurnConditionsNativelyConsumed(ability *Ability, declarations []StaticDeclarationSyntax) {
	var consuming []shared.Span
	for i := range declarations {
		if declarations[i].RestrictDuringControllerTurn {
			consuming = append(consuming, declarations[i].Span)
		}
	}
	if len(consuming) == 0 {
		return
	}
	keptBoundaries := ability.ConditionBoundaries[:0]
	for _, boundary := range ability.ConditionBoundaries {
		if boundary.Kind == ConditionIntroAsLongAs &&
			positionWithinAnySpan(boundary.Start, consuming) &&
			tokenWordAtPosition(ability.Tokens, boundary.Start, "during") {
			continue
		}
		keptBoundaries = append(keptBoundaries, boundary)
	}
	ability.ConditionBoundaries = keptBoundaries
	keptClauses := ability.ConditionClauses[:0]
	for _, clause := range ability.ConditionClauses {
		if clause.Predicate == ConditionPredicateControllerTurn && spanCoveredByAny(clause.Span, consuming) {
			continue
		}
		keptClauses = append(keptClauses, clause)
	}
	ability.ConditionClauses = keptClauses
}

// dropExilePlayProvenanceConditionNativelyConsumed removes the "if it was exiled
// by an ability you controlled" condition the generic scanner emits when a
// play-a-card-from-exile-with-counter declaration already captured that
// provenance clause as its ExilePlayExiledByControlledAbility rider (Evelyn, the
// Covetous). The declaration owns the whole sentence text-blind, so dropping the
// boundary and its segment keeps the provenance clause from surfacing as a
// static-wide "if" gate the compiler would preserve as an unlowered condition.
func dropExilePlayProvenanceConditionNativelyConsumed(ability *Ability, declarations []StaticDeclarationSyntax) {
	var consuming []shared.Span
	for i := range declarations {
		if declarations[i].ExilePlayExiledByControlledAbility {
			consuming = append(consuming, declarations[i].Span)
		}
	}
	if len(consuming) == 0 {
		return
	}
	keptBoundaries := ability.ConditionBoundaries[:0]
	for _, boundary := range ability.ConditionBoundaries {
		if boundary.Kind == ConditionIntroIf && positionWithinAnySpan(boundary.Start, consuming) {
			continue
		}
		keptBoundaries = append(keptBoundaries, boundary)
	}
	ability.ConditionBoundaries = keptBoundaries
	keptSegments := ability.ConditionSegments[:0]
	for _, segment := range ability.ConditionSegments {
		if segment.Kind == ConditionIntroIf && spanCoveredByAny(segment.Span, consuming) {
			continue
		}
		keptSegments = append(keptSegments, segment)
	}
	ability.ConditionSegments = keptSegments
}

func positionWithinAnySpan(pos shared.Position, spans []shared.Span) bool {
	for _, span := range spans {
		if span.Start.Offset <= pos.Offset && pos.Offset < span.End.Offset {
			return true
		}
	}
	return false
}

func tokenWordAtPosition(tokens []shared.Token, pos shared.Position, word string) bool {
	for i := range tokens {
		if tokens[i].Span.Start.Offset == pos.Offset {
			return equalWord(tokens[i], word)
		}
	}
	return false
}

func spanCoveredByAny(inner shared.Span, outers []shared.Span) bool {
	for _, outer := range outers {
		if spanCovers(outer, inner) {
			return true
		}
	}
	return false
}

// foldStaticCastFromTopPayLifeRider clears the effects of the "If you cast a
// spell this way, pay life equal to its mana value rather than pay its mana
// cost." rider sentence when the ability recognized a cast-from-library-top
// static permission carrying that rider (Bolas's Citadel). The rider's life
// payment is already captured on the static declaration, so clearing the
// sentence's effects keeps the static permission free of a stray standalone
// effect and lets reference and coverage scans credit the rider text.
func foldStaticCastFromTopPayLifeRider(ability *Ability, declarations []StaticDeclarationSyntax) {
	credited := false
	for i := range declarations {
		if declarations[i].PlayerRule == StaticDeclarationPlayerRuleCastSpellsFromLibraryTop &&
			declarations[i].CastPayLifeManaValue {
			credited = true
			break
		}
	}
	if !credited {
		return
	}
	for i := range ability.Sentences {
		if len(ability.Sentences[i].Effects) == 0 && !ability.Sentences[i].LegacyEffects {
			continue
		}
		if !isPlayFromTopPayLifeRiderTokens(semanticEffectTokens(ability.Sentences[i].Tokens)) {
			continue
		}
		ability.Sentences[i].Effects = nil
		ability.Sentences[i].LegacyEffects = false
		ability.Sentences[i].PlayFromTopPayLifeRider = true
		return
	}
}

// foldStaticRemoveAuraRider credits the inert "This effect doesn't remove this
// Aura." sentence that the protection-Aura cycle appends to its "Enchanted
// creature has protection from <color>" keyword grant. The grant's declaration
// already covers the protection; the rider only overrides paper Magic's
// state-based Aura falloff, which this engine never performs, so marking the
// sentence keeps it from surfacing as an unrecognized sibling that would block
// the keyword grant from lowering. The sentence's own effects, targets, and
// references are cleared so the static ability lowers through the typed
// keyword-grant path with a clean otherwise-empty shell.
func foldStaticRemoveAuraRider(ability *Ability, declarations []StaticDeclarationSyntax) {
	credited := false
	for i := range declarations {
		if declarations[i].Kind == StaticDeclarationKeywordGrant {
			credited = true
			break
		}
	}
	if !credited {
		return
	}
	for i := range ability.Sentences {
		sentence := &ability.Sentences[i]
		if sentence.RemoveAuraRider {
			continue
		}
		if !isRemoveAuraRider(semanticEffectTokens(sentence.Tokens)) {
			continue
		}
		sentence.RemoveAuraRider = true
		sentence.Effects = nil
		sentence.Targets = nil
		sentence.LegacyEffects = false
	}
}

// foldStaticSkipDrawStep clears the legacy effect that the imperative "Skip your
// draw step." sentence produces once the ability has recognized that text as the
// controller-scoped skip-draw-step turn-structure rule (Necropotence, Yawgmoth's
// Bargain). The static declaration already captures the turn-structure rule, so
// dropping the sentence's standalone effect keeps the static ability free of a
// stray imperative effect and lets it lower through the typed player-rule path
// instead of the unsupported non-keyword static fallback.
func foldStaticSkipDrawStep(ability *Ability, declarations []StaticDeclarationSyntax) {
	credited := false
	for i := range declarations {
		if declarations[i].PlayerRule == StaticDeclarationPlayerRuleSkipDrawStep {
			credited = true
			break
		}
	}
	if !credited {
		return
	}
	for i := range ability.Sentences {
		if len(ability.Sentences[i].Effects) == 0 && !ability.Sentences[i].LegacyEffects {
			continue
		}
		if !staticWordsAt(semanticEffectTokens(ability.Sentences[i].Tokens), 0, "skip", "your", "draw", "step") {
			continue
		}
		ability.Sentences[i].Effects = nil
		ability.Sentences[i].LegacyEffects = false
		return
	}
}

// foldStaticRedirectDamageToSource drops the spurious "is dealt to this creature"
// damage effect the generic effect scanner emits for the redirect static "All
// damage that would be dealt to you is dealt to this creature instead." once the
// static-declaration recognizer has credited it, so the sentence is fully
// consumed by the static rule rather than surfacing as an unlowered effect.
func foldStaticRedirectDamageToSource(ability *Ability, declarations []StaticDeclarationSyntax) {
	credited := false
	for i := range declarations {
		if declarations[i].PlayerRule == StaticDeclarationPlayerRuleRedirectDamageToSource {
			credited = true
			break
		}
	}
	if !credited {
		return
	}
	phrase := []string{"all", "damage", "that", "would", "be", "dealt", "to", "you"}
	for i := range ability.Sentences {
		if len(ability.Sentences[i].Effects) == 0 && !ability.Sentences[i].LegacyEffects {
			continue
		}
		tokens := semanticEffectTokens(ability.Sentences[i].Tokens)
		if !tokenRunContains(tokens, phrase) {
			continue
		}
		ability.Sentences[i].Effects = nil
		ability.Sentences[i].LegacyEffects = false
		return
	}
}

// foldStaticDamageDoesntCauseLifeLoss drops the spurious "lose life" effect the
// generic classifier emits for "damage doesn't cause you to lose life." (Archon
// of Coronation) once that sentence is credited to the typed player-rule static
// declaration. The phrase may carry a leading "As long as <condition>," clause,
// so the "damage doesn't cause you to lose life" run is matched wherever it
// begins. Dropping the standalone effect lets the ability lower through the typed
// player-rule path instead of the unsupported non-keyword static fallback.
func foldStaticDamageDoesntCauseLifeLoss(ability *Ability, declarations []StaticDeclarationSyntax) {
	credited := false
	for i := range declarations {
		if declarations[i].PlayerRule == StaticDeclarationPlayerRuleDamageDoesntCauseLifeLoss {
			credited = true
			break
		}
	}
	if !credited {
		return
	}
	phrase := []string{"damage", "doesn't", "cause", "you", "to", "lose", "life"}
	for i := range ability.Sentences {
		if len(ability.Sentences[i].Effects) == 0 && !ability.Sentences[i].LegacyEffects {
			continue
		}
		tokens := semanticEffectTokens(ability.Sentences[i].Tokens)
		if !tokenRunContains(tokens, phrase) {
			continue
		}
		ability.Sentences[i].Effects = nil
		ability.Sentences[i].LegacyEffects = false
		return
	}
}

// tokenRunContains reports whether words appears as a consecutive run anywhere in
// tokens, matched case-insensitively word by word.
func tokenRunContains(tokens []shared.Token, words []string) bool {
	if len(words) == 0 || len(tokens) < len(words) {
		return false
	}
	for start := 0; start+len(words) <= len(tokens); start++ {
		if staticWordsAt(tokens, start, words...) {
			return true
		}
	}
	return false
}

// whose subject is the card's own printed name ("Toski attacks each combat if
// able.") instead of a "this creature"/"this permanent" marker. Sentence
// splitting runs before atoms are recognized, so the self-name form is resolved
// here once the source-name aliases are available, letting it lower through the
// same typed static-rule path as the marker form.
func emitSelfNameStaticRules(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		for j := range ability.Sentences {
			sentence := &ability.Sentences[j]
			if sentence.StaticRule != nil {
				continue
			}
			if rule, ok := parseSelfNameStaticRuleSyntax(sentence.Tokens, ability.Atoms); ok {
				sentence.StaticRule = rule
			}
		}
	}
}

// staticDeclarationBodyTokens returns the ability's semantic tokens with reminder
// and quoted text removed, and any ability-word label and its em dash dropped.
func staticDeclarationBodyTokens(ability *Ability) []shared.Token {
	tokens := eventHistorySemanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
	tokens = stripNonBattlefieldScopeRider(tokens)
	tokens = stripRemoveAuraRider(tokens)
	if ability.AbilityWord == nil {
		return tokens
	}
	for i := range tokens {
		if tokens[i].Kind == shared.EmDash {
			return tokens[i+1:]
		}
	}
	return tokens
}

// stripNonBattlefieldScopeRider removes a trailing "The same is true for <spells
// you control and cards you own> that aren't on the battlefield." sentence from a
// static ability's body. That rider extends a "<group> is/are <characteristic>"
// continuous declaration (Maskwood Nexus, Arcane Adaptation, Encroaching
// Mycosynth, Conspiracy, Celestial Dawn, Biotransference) to objects in
// non-battlefield zones — spells on the stack and cards in hand/library/
// graveyard. Continuous effects in this engine apply only to battlefield
// permanents, so the rider has no representable effect; dropping it lets the
// leading battlefield declaration lower while leaving battlefield simulation
// outcomes unchanged. The rider is distinguished by its "the same is true for"
// opening and "on the battlefield" close, which the keyword-copying "same is
// true for <keyword list>" riders (Odric, Cairn Wanderer) never share.
func stripNonBattlefieldScopeRider(tokens []shared.Token) []shared.Token {
	end := len(tokens)
	if end == 0 || tokens[end-1].Kind != shared.Period {
		return tokens
	}
	sentenceStart := 0
	for i := end - 2; i >= 0; i-- {
		if tokens[i].Kind == shared.Period {
			sentenceStart = i + 1
			break
		}
	}
	if sentenceStart == 0 {
		return tokens
	}
	if !isNonBattlefieldScopeRider(tokens[sentenceStart : end-1]) {
		return tokens
	}
	return tokens[:sentenceStart]
}

// isNonBattlefieldScopeRider reports whether sentence (the tokens of one
// sentence, excluding its terminating period) is the non-battlefield-zone scope
// extension "the same is true for ... on the battlefield".
func isNonBattlefieldScopeRider(sentence []shared.Token) bool {
	if len(sentence) < 8 {
		return false
	}
	return staticWordsAt(sentence, 0, "the", "same", "is", "true", "for") &&
		staticWordsAt(sentence, len(sentence)-3, "on", "the", "battlefield")
}

// stripRemoveAuraRider removes a trailing "This effect doesn't remove this Aura."
// reminder sentence from a static ability's body. The Ward cycle of protection
// Auras (Cho-Manno's Blessing, Flickering Ward, Pentarch Ward, the
// White/Blue/Black/Red/Green Ward fixed cycle) and several chosen-color Auras
// append this clarification to their "Enchanted creature has protection from
// <color>" grant. Protection from a color makes an Aura attached to that
// creature an illegal attachment in paper Magic, which would normally put the
// Aura into the graveyard as a state-based action; the rider overrides that so
// the Aura stays attached. This engine never falls Auras off for color
// protection, so the rider has no representable effect and dropping it lets the
// leading protection grant lower while leaving simulation outcomes unchanged.
func stripRemoveAuraRider(tokens []shared.Token) []shared.Token {
	end := len(tokens)
	if end == 0 || tokens[end-1].Kind != shared.Period {
		return tokens
	}
	sentenceStart := 0
	for i := end - 2; i >= 0; i-- {
		if tokens[i].Kind == shared.Period {
			sentenceStart = i + 1
			break
		}
	}
	if sentenceStart == 0 {
		return tokens
	}
	if !isRemoveAuraRider(tokens[sentenceStart : end-1]) {
		return tokens
	}
	return tokens[:sentenceStart]
}

// isRemoveAuraRider reports whether sentence (the tokens of one sentence,
// excluding its terminating period) is an inert "This effect doesn't remove
// <attachments>." clarification. The recognized variants are "this effect
// doesn't remove this Aura", "this effect doesn't remove Auras", and "this
// effect doesn't remove Auras and Equipment you control that are already
// attached to it".
func isRemoveAuraRider(sentence []shared.Token) bool {
	if len(sentence) > 0 && sentence[len(sentence)-1].Kind == shared.Period {
		sentence = sentence[:len(sentence)-1]
	}
	if len(sentence) < 5 {
		return false
	}
	if !staticWordsAt(sentence, 0, "this", "effect", "doesn't", "remove") {
		return false
	}
	rest := sentence[4:]
	switch len(rest) {
	case 2:
		return staticWordsAt(rest, 0, "this", "aura")
	case 1:
		return equalWord(rest[0], "auras")
	default:
		return staticWordsAt(rest, 0, "auras", "and", "equipment") &&
			staticWordsAt(rest, len(rest)-3, "attached", "to", "it")
	}
}

func parseStaticDeclarations(tokens []shared.Token, quoted []Delimited, atoms Atoms, conditions []ConditionClause) []StaticDeclarationSyntax {
	if declaration, ok := parseStaticOpeningHandPlayDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticOpponentEnteringTriggerSuppressionDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticCreatureAttackTaxDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticManaProductionMultiplierDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseChosenCreatureTypeTriggerMultiplierDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseEnteringTriggerMultiplierDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseControlledTriggerMultiplierDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticCostModifierDeclaration(tokens, atoms, conditions); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticAbilityCostSetDeclaration(tokens, conditions); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticSpellCostModifierDeclaration(tokens, atoms); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticCastAsThoughFlashDeclaration(tokens, atoms); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticSpellUncounterableDeclaration(tokens, atoms); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticCombatDamagePreventionDeclaration(tokens, atoms); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticUntapDuringOtherUntapStepDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticCardAbilityGrantDeclaration(tokens, atoms); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticGraveyardCardKeywordGrantDeclaration(tokens, atoms); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticPermanentAbilityGrantDeclaration(tokens, quoted, conditions); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticControlGrantDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticMonarchControlGrantDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticCastThisFromGraveyardDeclaration(tokens, conditions); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticCastThisFromExileDeclaration(tokens, conditions); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticPlayerRuleDeclaration(tokens, conditions); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticOpponentActionRestrictionDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticDrawLimitDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticCastLimitDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticEnchantedTypeChangeDeclaration(tokens, quoted, atoms); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticEnterBattlefieldRestrictionDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticLoseAbilitiesBecomeDeclaration(tokens, atoms); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseCharacteristicDefiningPowerToughnessDeclaration(tokens, atoms); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declarations, ok := parseStaticQuotedAbilityGrantDeclarations(tokens, quoted, atoms, conditions); ok {
		return declarations
	}
	if declarations, ok := parseStaticSubjectDeclarations(tokens, atoms, conditions); ok {
		return declarations
	}
	return nil
}

func parseChosenCreatureTypeTriggerMultiplierDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 21 ||
		tokens[14].Kind != shared.Comma ||
		tokens[20].Kind != shared.Period ||
		!staticWordsAt(tokens, 0,
			"if", "a", "triggered", "ability", "of", "another", "creature", "you", "control",
			"of", "the", "chosen", "type", "triggers") ||
		!staticWordsAt(tokens, 15, "it", "triggers", "an", "additional", "time") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationChosenCreatureTypeTriggerMultiplier,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens),
	}, true
}

// parseEnteringTriggerMultiplierDeclaration recognizes the "triggers an
// additional time" replacement family "If <filter> entering causes a triggered
// ability of a permanent you control to trigger, that ability triggers an
// additional time." (Panharmonicon, Yarok, Ancient Greenwarden). <filter> is "a
// permanent" (matching any entering permanent) or an article followed by an
// "or"-joined card-type list ("an artifact or creature", "a land"). The entering
// permanent's type filter is captured in EnteringFilterTypes; an empty list
// matches any permanent. Any deviation leaves the clause unconsumed.
func parseEnteringTriggerMultiplierDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	const suffixLen = 20
	if len(tokens) < suffixLen+3 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "if") {
		return StaticDeclarationSyntax{}, false
	}
	enter := len(tokens) - suffixLen
	if !staticWordsAt(tokens, enter,
		"entering", "causes", "a", "triggered", "ability", "of", "a", "permanent",
		"you", "control", "to", "trigger") ||
		tokens[enter+12].Kind != shared.Comma ||
		!staticWordsAt(tokens, enter+13, "that", "ability", "triggers", "an", "additional", "time") {
		return StaticDeclarationSyntax{}, false
	}
	filterTypes, ok := parseEnteringFilter(tokens, 1, enter)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                StaticDeclarationEnteringTriggerMultiplier,
		Span:                shared.SpanOf(tokens),
		OperationSpan:       shared.SpanOf(tokens),
		EnteringFilterTypes: filterTypes,
	}, true
}

// parseEnteringFilter consumes the entering-permanent filter "a permanent" or an
// article followed by an "or"-joined card-type list. It returns an empty slice
// for "a permanent" (any entering permanent) and the listed card types
// otherwise, failing closed when the region is not exactly one such filter.
func parseEnteringFilter(tokens []shared.Token, index, end int) ([]CardType, bool) {
	if index >= end || (!equalWord(tokens[index], "a") && !equalWord(tokens[index], "an")) {
		return nil, false
	}
	index++
	if index >= end {
		return nil, false
	}
	if index+1 == end && equalWord(tokens[index], "permanent") {
		return nil, true
	}
	cardTypes, next, ok := parseStaticCardTypeList(tokens, index, end)
	if !ok || next != end {
		return nil, false
	}
	return cardTypes, true
}

// parseControlledTriggerMultiplierDeclaration recognizes the "triggers an
// additional time" family scoped to a filtered controlled permanent: "If a
// triggered ability of <filter> you control triggers, that ability triggers an
// additional time." (Annie Joins Up — "a legendary creature"; Katara, the
// Fearless — "an Ally"; Splinter, Radical Rat — "a Ninja creature"). The closing
// clause may instead read "it triggers an additional time." (Twinflame
// Travelers). <filter> is one or more "or"-joined branches, each an article
// ("a", "an", or "another") followed by an optional supertype, optional subtype,
// and optional card type ("a Shaman or another Wizard", Harmonic Prodigy); a
// branch led by "another" excludes the doubler itself. Any deviation leaves the
// clause unconsumed. The entering-permanent and "of the chosen type" forms are
// owned by their own parsers and do not reach here.
func parseControlledTriggerMultiplierDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	const prefixLen = 5
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "if", "a", "triggered", "ability", "of") {
		return StaticDeclarationSyntax{}, false
	}
	tail, ok := controlledTriggerMultiplierFilterEnd(tokens)
	if !ok || tail <= prefixLen {
		return StaticDeclarationSyntax{}, false
	}
	branches, ok := parseControlledTriggerBranches(tokens, prefixLen, tail)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                     StaticDeclarationControlledTriggerMultiplier,
		Span:                     shared.SpanOf(tokens),
		OperationSpan:            shared.SpanOf(tokens),
		ControlledFilterBranches: branches,
	}, true
}

// controlledTriggerMultiplierFilterEnd locates the index just past the source
// filter, i.e. the start of the "you control triggers, <that ability|it>
// triggers an additional time." suffix. It returns false when the suffix is not
// the exact controlled-trigger multiplier closing clause.
func controlledTriggerMultiplierFilterEnd(tokens []shared.Token) (int, bool) {
	for _, suffix := range [][]string{
		{"you", "control", "triggers", ",", "that", "ability", "triggers", "an", "additional", "time", "."},
		{"you", "control", "triggers", ",", "it", "triggers", "an", "additional", "time", "."},
	} {
		start := len(tokens) - len(suffix)
		if start < 0 {
			continue
		}
		if controlledTriggerSuffixMatches(tokens, start, suffix) {
			return start, true
		}
	}
	return 0, false
}

// controlledTriggerSuffixMatches reports whether the tokens from start match the
// given closing-clause words, treating "," and "." as their punctuation kinds
// and every other element as a word.
func controlledTriggerSuffixMatches(tokens []shared.Token, start int, suffix []string) bool {
	for offset, word := range suffix {
		token := tokens[start+offset]
		switch word {
		case ",":
			if token.Kind != shared.Comma {
				return false
			}
		case ".":
			if token.Kind != shared.Period {
				return false
			}
		default:
			if !equalWord(token, word) {
				return false
			}
		}
	}
	return true
}

// ControlledTriggerSourceFilter is one "or"-joined branch of a
// controlled-trigger multiplier's source-permanent filter. Supertypes and
// CardTypes are conjunctive; Subtypes is disjunctive. ExcludeSelf records a
// leading "another", dropping the doubler's own source from the branch.
type ControlledTriggerSourceFilter struct {
	Supertypes  []Supertype
	Subtypes    []types.Sub
	CardTypes   []CardType
	ExcludeSelf bool
}

// parseControlledTriggerBranches consumes the source-permanent filter between the
// "of" and "you control" anchors as one or more "or"-joined branches. Each branch
// is "a/an/another [supertype] [subtype] [card type]" naming at least one subtype
// or card type; a bare supertype is not a permanent filter. The region must be
// consumed exactly.
func parseControlledTriggerBranches(tokens []shared.Token, index, end int) ([]ControlledTriggerSourceFilter, bool) {
	var branches []ControlledTriggerSourceFilter
	for index < end {
		branch, next, ok := parseControlledTriggerBranch(tokens, index, end)
		if !ok {
			return nil, false
		}
		branches = append(branches, branch)
		if next == end {
			return branches, true
		}
		if !equalWord(tokens[next], "or") {
			return nil, false
		}
		index = next + 1
	}
	return nil, false
}

// parseControlledTriggerBranch consumes a single branch starting at index,
// returning the branch and the index of the following token (the "or" separator
// or end). The branch begins with an article; "another" sets ExcludeSelf.
func parseControlledTriggerBranch(tokens []shared.Token, index, end int) (ControlledTriggerSourceFilter, int, bool) {
	if index >= end || tokens[index].Kind != shared.Word {
		return ControlledTriggerSourceFilter{}, 0, false
	}
	var branch ControlledTriggerSourceFilter
	switch {
	case equalWord(tokens[index], "a"), equalWord(tokens[index], "an"):
	case equalWord(tokens[index], "another"):
		branch.ExcludeSelf = true
	default:
		return ControlledTriggerSourceFilter{}, 0, false
	}
	index++
	for index < end && tokens[index].Kind == shared.Word {
		word := tokens[index].Text
		if equalWord(tokens[index], "or") {
			break
		}
		if supertype, ok := recognizeSupertypeWord(word); ok {
			branch.Supertypes = append(branch.Supertypes, supertype)
			index++
			continue
		}
		if cardType, ok := recognizeCardTypeWord(word); ok {
			branch.CardTypes = append(branch.CardTypes, cardType)
			index++
			continue
		}
		if subtype, ok := recognizeSubtypePhrase(word); ok {
			branch.Subtypes = append(branch.Subtypes, subtype)
			index++
			continue
		}
		return ControlledTriggerSourceFilter{}, 0, false
	}
	if len(branch.Subtypes) == 0 && len(branch.CardTypes) == 0 {
		return ControlledTriggerSourceFilter{}, 0, false
	}
	return branch, index, true
}

func parseStaticPermanentAbilityGrantDeclaration(
	tokens []shared.Token,
	quoted []Delimited,
	conditions []ConditionClause,
) (StaticDeclarationSyntax, bool) {
	if len(conditions) != 0 ||
		len(quoted) != 1 ||
		len(tokens) != 4 ||
		!staticWordsAt(tokens, 1, "you", "control", "have") {
		return StaticDeclarationSyntax{}, false
	}
	subject, ok := staticPermanentGrantSubject(tokens[0], shared.SpanOf(tokens[:3]))
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	ability, ok := parseStaticGrantedManaAbility(quoted[0])
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:               StaticDeclarationPermanentAbilityGrant,
		Span:               shared.Span{Start: tokens[0].Span.Start, End: quoted[0].Span.End},
		OperationSpan:      quoted[0].Span,
		Subject:            subject,
		GrantedManaAbility: &ability,
	}, true
}

// staticPermanentGrantSubject maps the leading "<group> you control" noun of a
// permanent-ability grant onto a typed group subject. It recognizes the
// controlled land, creature, and artifact groups, plus the Treasure artifact
// subtype, and fails closed for any other group noun.
func staticPermanentGrantSubject(noun shared.Token, span shared.Span) (StaticDeclarationSubject, bool) {
	group := EffectStaticSubjectSyntax{Span: span}
	switch {
	case equalWord(noun, "lands"):
		group.Kind = EffectStaticSubjectControlledLands
	case equalWord(noun, "creatures"):
		group.Kind = EffectStaticSubjectControlledCreatures
	case equalWord(noun, "artifacts"):
		group.Kind = EffectStaticSubjectControlledArtifacts
	case equalWord(noun, "treasures"):
		group.Kind = EffectStaticSubjectControlledArtifacts
		group.Subtype = types.Treasure
		group.SubtypeText = string(types.Treasure)
		group.SubtypeKnown = true
	default:
		return StaticDeclarationSubject{}, false
	}
	return StaticDeclarationSubject{
		Kind:  StaticDeclarationSubjectGroup,
		Span:  span,
		Group: group,
	}, true
}

// parseStaticGrantedManaAbility recognizes one of two quoted activated mana
// abilities a permanent-ability grant may confer: the bare tap form
// "{T}: Add one mana of any color." and the Treasure-style sacrifice form
// "{T}, Sacrifice this artifact: Add <N> mana of any one color." (N >= 2).
func parseStaticGrantedManaAbility(quoted Delimited) (StaticGrantedManaAbilitySyntax, bool) {
	if ability, ok := parseStaticGrantedAnyColorManaAbility(quoted); ok {
		return ability, true
	}
	if ability, ok := parseStaticGrantedColorlessManaAbility(quoted); ok {
		return ability, true
	}
	if ability, ok := parseStaticGrantedSacrificeAnyColorManaAbility(quoted); ok {
		return ability, true
	}
	return parseStaticGrantedSacrificeManaAbility(quoted)
}

// parseStaticGrantedColorlessManaAbility recognizes the bare quoted ability
// "{T}: Add {C}." that adds one colorless mana, granted by removal Auras such as
// Imprisoned in the Moon.
func parseStaticGrantedColorlessManaAbility(quoted Delimited) (StaticGrantedManaAbilitySyntax, bool) {
	tokens := quoted.Tokens
	if len(tokens) < 6 ||
		tokens[0].Kind != shared.Quote ||
		tokens[1].Kind != shared.Symbol ||
		tokens[1].Text != "{T}" ||
		tokens[2].Kind != shared.Colon ||
		!staticWordsAt(tokens, 3, "add") ||
		tokens[4].Kind != shared.Symbol ||
		tokens[4].Text != "{C}" {
		return StaticGrantedManaAbilitySyntax{}, false
	}
	rest := tokens[5:]
	validTail := (len(rest) == 2 && rest[0].Kind == shared.Period && rest[1].Kind == shared.Quote) ||
		(len(rest) == 1 && rest[0].Kind == shared.Quote)
	if !validTail {
		return StaticGrantedManaAbilitySyntax{}, false
	}
	return StaticGrantedManaAbilitySyntax{
		Span:      shared.SpanOf(tokens[1:5]),
		Text:      staticGrantedAbilityText(quoted),
		TapCost:   true,
		Amount:    1,
		Colorless: true,
	}, true
}

func parseStaticGrantedAnyColorManaAbility(quoted Delimited) (StaticGrantedManaAbilitySyntax, bool) {
	tokens := quoted.Tokens
	if len(tokens) != 11 ||
		tokens[0].Kind != shared.Quote ||
		tokens[1].Kind != shared.Symbol ||
		tokens[1].Text != "{T}" ||
		tokens[2].Kind != shared.Colon ||
		!staticWordsAt(tokens, 3, "add", "one", "mana", "of", "any", "color") ||
		tokens[9].Kind != shared.Period ||
		tokens[10].Kind != shared.Quote {
		return StaticGrantedManaAbilitySyntax{}, false
	}
	return StaticGrantedManaAbilitySyntax{
		Span:     shared.SpanOf(tokens[1:10]),
		Text:     staticGrantedAbilityText(quoted),
		TapCost:  true,
		Amount:   1,
		AnyColor: true,
	}, true
}

// parseStaticGrantedSacrificeAnyColorManaAbility recognizes the count-1 quoted
// ability "{T}, Sacrifice this artifact: Add one mana of any color." (Ninja
// Pizza): tap and sacrifice the host artifact to add one mana of any color the
// controller chooses. It is the any-color, count-1 counterpart of the
// Treasure-style "Add <N> mana of any one color." sacrifice ability.
func parseStaticGrantedSacrificeAnyColorManaAbility(quoted Delimited) (StaticGrantedManaAbilitySyntax, bool) {
	tokens := quoted.Tokens
	if len(tokens) != 15 ||
		tokens[0].Kind != shared.Quote ||
		tokens[1].Kind != shared.Symbol ||
		tokens[1].Text != "{T}" ||
		tokens[2].Kind != shared.Comma ||
		!staticWordsAt(tokens, 3, "sacrifice", "this", "artifact") ||
		tokens[6].Kind != shared.Colon ||
		!staticWordsAt(tokens, 7, "add", "one", "mana", "of", "any", "color") ||
		tokens[13].Kind != shared.Period ||
		tokens[14].Kind != shared.Quote {
		return StaticGrantedManaAbilitySyntax{}, false
	}
	return StaticGrantedManaAbilitySyntax{
		Span:      shared.SpanOf(tokens[1:14]),
		Text:      staticGrantedAbilityText(quoted),
		TapCost:   true,
		Amount:    1,
		AnyColor:  true,
		Sacrifice: true,
	}, true
}

func parseStaticGrantedSacrificeManaAbility(quoted Delimited) (StaticGrantedManaAbilitySyntax, bool) {
	tokens := quoted.Tokens
	if len(tokens) != 16 ||
		tokens[0].Kind != shared.Quote ||
		tokens[1].Kind != shared.Symbol ||
		tokens[1].Text != "{T}" ||
		tokens[2].Kind != shared.Comma ||
		!staticWordsAt(tokens, 3, "sacrifice", "this", "artifact") ||
		tokens[6].Kind != shared.Colon ||
		!staticWordsAt(tokens, 7, "add") ||
		!staticWordsAt(tokens, 9, "mana", "of", "any", "one", "color") ||
		tokens[14].Kind != shared.Period ||
		tokens[15].Kind != shared.Quote {
		return StaticGrantedManaAbilitySyntax{}, false
	}
	count, ok := manaAnyOneColorCount(tokens[8])
	if !ok {
		return StaticGrantedManaAbilitySyntax{}, false
	}
	return StaticGrantedManaAbilitySyntax{
		Span:        shared.SpanOf(tokens[1:15]),
		Text:        staticGrantedAbilityText(quoted),
		TapCost:     true,
		Amount:      count,
		Sacrifice:   true,
		AnyOneColor: true,
	}, true
}

// staticGrantedAbilityText returns the quoted ability's source text with its
// surrounding double quotes removed.
func staticGrantedAbilityText(quoted Delimited) string {
	return strings.TrimSuffix(strings.TrimPrefix(quoted.Text, `"`), `"`)
}

// parseStaticControlGrantDeclaration recognizes the static source-tied control
// grant printed on control Auras: "You control enchanted creature." or "You
// control enchanted permanent." The affected group is the attached object; the
// new controller is the static ability's controller (you).
func parseStaticControlGrantDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 5 || tokens[4].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "control", "enchanted") {
		return StaticDeclarationSyntax{}, false
	}
	if !equalWord(tokens[3], "creature") && !equalWord(tokens[3], "permanent") {
		return StaticDeclarationSyntax{}, false
	}
	objectSpan := shared.SpanOf(tokens[2:4])
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationControlGrant,
		Span:          shared.SpanOf(tokens),
		OperationSpan: tokens[1].Span,
		Subject: StaticDeclarationSubject{
			Kind:  StaticDeclarationSubjectGroup,
			Span:  objectSpan,
			Group: EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttachedObject, Span: objectSpan},
		},
	}, true
}

// parseStaticMonarchControlGrantDeclaration recognizes "The monarch controls
// enchanted creature." (Fealty to the Realm): the enchanted permanent is
// controlled by whoever currently holds the monarch designation. Unlike the
// fixed "You control enchanted creature." grant, the controller follows the
// crown, so the compiler lowers it to a monarch-bound control effect. Any other
// wording leaves the clause unconsumed and fails closed.
func parseStaticMonarchControlGrantDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 6 || tokens[5].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "the", "monarch", "controls", "enchanted") {
		return StaticDeclarationSyntax{}, false
	}
	if !equalWord(tokens[4], "creature") && !equalWord(tokens[4], "permanent") {
		return StaticDeclarationSyntax{}, false
	}
	objectSpan := shared.SpanOf(tokens[3:5])
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationMonarchControlGrant,
		Span:          shared.SpanOf(tokens),
		OperationSpan: tokens[2].Span,
		Subject: StaticDeclarationSubject{
			Kind:  StaticDeclarationSubjectGroup,
			Span:  objectSpan,
			Group: EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttachedObject, Span: objectSpan},
		},
	}, true
}

type staticPlayerRuleParser func([]shared.Token) (StaticDeclarationSyntax, bool)

// parseStaticOpponentActionRestrictionDeclaration recognizes the continuous
// action prohibition family "[During your turn,] <players> can't cast spells [or
// activate abilities of <types>] [during your turn]." (Grand Abolisher, Teferi).
// <players> is "your opponents", "each opponent", or "players"; the trailing
// type list (e.g. "artifacts, creatures, or enchantments") scopes the activation
// prohibition. The passive "spells can't be cast [during your turn]." is also
// recognized. Any deviation leaves the clause unconsumed and fails closed.
func parseStaticOpponentActionRestrictionDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 4 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	end := len(tokens) - 1
	index := 0
	duringControllerTurn := false
	if staticWordsAt(tokens, index, "during", "your", "turn") {
		duringControllerTurn = true
		index += 3
		if index < end && tokens[index].Kind == shared.Comma {
			index++
		}
	}
	if declaration, ok := parseStaticPassiveCastProhibition(tokens, index, end, duringControllerTurn); ok {
		return declaration, true
	}
	affectsAll := false
	switch {
	case staticWordsAt(tokens, index, "your", "opponents"):
		index += 2
	case staticWordsAt(tokens, index, "each", "opponent"):
		index += 2
	case staticWordsAt(tokens, index, "players"):
		affectsAll = true
		index++
	default:
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, index, "can't") && !staticWordsAt(tokens, index, "cannot") {
		return StaticDeclarationSyntax{}, false
	}
	index++
	actions, index, ok := parseStaticRestrictedActions(tokens, index, end)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	if staticWordsAt(tokens, index, "during", "your", "turn") {
		duringControllerTurn = true
		index += 3
	}
	if index != end {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                         StaticDeclarationOpponentActionRestriction,
		Span:                         shared.SpanOf(tokens),
		OperationSpan:                shared.SpanOf(tokens[:end]),
		RestrictCastSpells:           actions.cast,
		RestrictActivateTypes:        actions.activateTypes,
		RestrictCastFromZones:        actions.castFromZones,
		RestrictCastOnlyFromHand:     actions.castOnlyFromHand,
		RestrictAffectsAllPlayers:    affectsAll,
		RestrictDuringControllerTurn: duringControllerTurn,
	}, true
}

// staticRestrictedActions holds the parsed actions a static prohibition forbids:
// casting spells and/or activating abilities of the listed permanent types. When
// the cast prohibition is scoped to a set of source zones the zones are listed in
// castFromZones; castOnlyFromHand records the "anywhere other than their hands"
// form that forbids every non-hand cast zone.
type staticRestrictedActions struct {
	cast             bool
	activateTypes    []CardType
	castFromZones    []StaticDeclarationCastZoneKind
	castOnlyFromHand bool
}

// parseStaticPassiveCastProhibition recognizes the passive "spells can't be cast
// [during your turn]." form, which forbids every player from casting spells.
func parseStaticPassiveCastProhibition(tokens []shared.Token, index, end int, duringControllerTurn bool) (StaticDeclarationSyntax, bool) {
	if !staticWordsAt(tokens, index, "spells", "can't", "be", "cast") &&
		!staticWordsAt(tokens, index, "spells", "cannot", "be", "cast") {
		return StaticDeclarationSyntax{}, false
	}
	index += 4
	if staticWordsAt(tokens, index, "during", "your", "turn") {
		duringControllerTurn = true
		index += 3
	}
	if index != end {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                         StaticDeclarationOpponentActionRestriction,
		Span:                         shared.SpanOf(tokens),
		OperationSpan:                shared.SpanOf(tokens[:end]),
		RestrictCastSpells:           true,
		RestrictAffectsAllPlayers:    true,
		RestrictDuringControllerTurn: duringControllerTurn,
	}, true
}

// parseStaticDrawLimitDeclaration recognizes the continuous per-turn draw cap
// "<players> can't draw more than <N> card[s] each turn." (Narset, Parter of
// Veils; Spirit of the Labyrinth; Leovold). <players> is "each opponent"/"your
// opponents" (opponents), "each player"/"players" (every player), or "you" (the
// controller). Any deviation leaves the clause unconsumed and fails closed.
func parseStaticDrawLimitDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 4 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	end := len(tokens) - 1
	index := 0
	affectsAll := false
	affectsController := false
	switch {
	case staticWordsAt(tokens, index, "each", "opponent"):
		index += 2
	case staticWordsAt(tokens, index, "your", "opponents"):
		index += 2
	case staticWordsAt(tokens, index, "each", "player"):
		affectsAll = true
		index += 2
	case staticWordsAt(tokens, index, "players"):
		affectsAll = true
		index++
	case staticWordsAt(tokens, index, "you"):
		affectsController = true
		index++
	default:
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, index, "can't") && !staticWordsAt(tokens, index, "cannot") {
		return StaticDeclarationSyntax{}, false
	}
	index++
	if !staticWordsAt(tokens, index, "draw", "more", "than") {
		return StaticDeclarationSyntax{}, false
	}
	index += 3
	if index >= end || tokens[index].Kind != shared.Word {
		return StaticDeclarationSyntax{}, false
	}
	limit, ok := CardinalWordValue(tokens[index].Text)
	if !ok || limit < 1 {
		return StaticDeclarationSyntax{}, false
	}
	index++
	if !staticWordsAt(tokens, index, "card") && !staticWordsAt(tokens, index, "cards") {
		return StaticDeclarationSyntax{}, false
	}
	index++
	if !staticWordsAt(tokens, index, "each", "turn") {
		return StaticDeclarationSyntax{}, false
	}
	index += 2
	if index != end {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                       StaticDeclarationDrawLimit,
		Span:                       shared.SpanOf(tokens),
		OperationSpan:              shared.SpanOf(tokens[:end]),
		DrawLimit:                  limit,
		DrawLimitAffectsAllPlayers: affectsAll,
		DrawLimitAffectsController: affectsController,
	}, true
}

// parseStaticCastLimitDeclaration recognizes the continuous per-turn spell cap
// "<players> can't cast more than <N> spell[s] each turn." (Rule of Law, Eidolon
// of Rhetoric, Arcane Laboratory; Moderation). <players> is "each
// opponent"/"your opponents" (opponents), "each player"/"players" (every
// player), or "you" (the controller). Only the unqualified "spell" object is
// recognized; a type-scoped object ("noncreature spell") or any other deviation
// leaves the clause unconsumed and fails closed.
func parseStaticCastLimitDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 4 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	end := len(tokens) - 1
	index := 0
	affectsAll := false
	affectsController := false
	switch {
	case staticWordsAt(tokens, index, "each", "opponent"):
		index += 2
	case staticWordsAt(tokens, index, "your", "opponents"):
		index += 2
	case staticWordsAt(tokens, index, "each", "player"):
		affectsAll = true
		index += 2
	case staticWordsAt(tokens, index, "players"):
		affectsAll = true
		index++
	case staticWordsAt(tokens, index, "you"):
		affectsController = true
		index++
	default:
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, index, "can't") && !staticWordsAt(tokens, index, "cannot") {
		return StaticDeclarationSyntax{}, false
	}
	index++
	if !staticWordsAt(tokens, index, "cast", "more", "than") {
		return StaticDeclarationSyntax{}, false
	}
	index += 3
	if index >= end || tokens[index].Kind != shared.Word {
		return StaticDeclarationSyntax{}, false
	}
	limit, ok := CardinalWordValue(tokens[index].Text)
	if !ok || limit < 1 {
		return StaticDeclarationSyntax{}, false
	}
	index++
	if !staticWordsAt(tokens, index, "spell") && !staticWordsAt(tokens, index, "spells") {
		return StaticDeclarationSyntax{}, false
	}
	index++
	if !staticWordsAt(tokens, index, "each", "turn") {
		return StaticDeclarationSyntax{}, false
	}
	index += 2
	if index != end {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                       StaticDeclarationCastLimit,
		Span:                       shared.SpanOf(tokens),
		OperationSpan:              shared.SpanOf(tokens[:end]),
		CastLimit:                  limit,
		CastLimitAffectsAllPlayers: affectsAll,
		CastLimitAffectsController: affectsController,
	}, true
}
func parseStaticRestrictedActions(tokens []shared.Token, index, end int) (staticRestrictedActions, int, bool) {
	var actions staticRestrictedActions
	for {
		switch {
		case staticWordsAt(tokens, index, "cast", "spells"):
			if actions.cast {
				return staticRestrictedActions{}, 0, false
			}
			actions.cast = true
			index += 2
			if staticWordsAt(tokens, index, "from") {
				spec, ok := parseStaticCastZoneSpec(tokens, index+1, end)
				if !ok {
					return staticRestrictedActions{}, 0, false
				}
				actions.castOnlyFromHand = spec.onlyFromHand
				actions.castFromZones = spec.zones
				index = spec.next
			}
		case staticWordsAt(tokens, index, "activate", "abilities", "of"):
			if len(actions.activateTypes) != 0 {
				return staticRestrictedActions{}, 0, false
			}
			cardTypes, next, ok := parseStaticCardTypeList(tokens, index+3, end)
			if !ok {
				return staticRestrictedActions{}, 0, false
			}
			actions.activateTypes = cardTypes
			index = next
		default:
			return staticRestrictedActions{}, 0, false
		}
		if !staticWordsAt(tokens, index, "or") {
			break
		}
		index++
	}
	if !actions.cast && len(actions.activateTypes) == 0 {
		return staticRestrictedActions{}, 0, false
	}
	return actions, index, true
}

// staticCastZoneSpec captures the parsed zone scope of a cast prohibition: the
// complementary "anywhere other than their hands" form (onlyFromHand) or an
// explicit list of non-hand cast zones, plus the index after the consumed spec.
type staticCastZoneSpec struct {
	onlyFromHand bool
	zones        []StaticDeclarationCastZoneKind
	next         int
}

// parseStaticCastZoneSpec recognizes the zone scope of a cast prohibition that
// follows "cast spells from": either the complementary "anywhere other than
// their/your hand[s]" form (returning onlyFromHand) or a list of non-hand cast
// zones joined by commas and/or "or" ("graveyards", "graveyards or libraries",
// "graveyards, libraries, or exile"). It fails closed on any other wording.
func parseStaticCastZoneSpec(tokens []shared.Token, index, end int) (staticCastZoneSpec, bool) {
	if staticWordsAt(tokens, index, "anywhere", "other", "than") {
		index += 3
		if staticWordsAt(tokens, index, "their") || staticWordsAt(tokens, index, "your") {
			index++
		}
		if !staticWordsAt(tokens, index, "hands") && !staticWordsAt(tokens, index, "hand") {
			return staticCastZoneSpec{}, false
		}
		index++
		return staticCastZoneSpec{onlyFromHand: true, next: index}, true
	}
	var zones []StaticDeclarationCastZoneKind
	for index < end {
		castZone, consumed, zok := staticCastZoneWord(tokens, index)
		if !zok {
			break
		}
		zones = append(zones, castZone)
		index += consumed
		separated := false
		if index < end && tokens[index].Kind == shared.Comma {
			index++
			separated = true
		}
		if index < end && equalWord(tokens[index], "or") {
			index++
			separated = true
		}
		if !separated {
			break
		}
	}
	if len(zones) == 0 {
		return staticCastZoneSpec{}, false
	}
	return staticCastZoneSpec{zones: zones, next: index}, true
}

// staticCastZoneWord maps a singular or plural zone noun onto its closed
// cast-zone kind, returning the number of tokens consumed.
func staticCastZoneWord(tokens []shared.Token, index int) (StaticDeclarationCastZoneKind, int, bool) {
	switch {
	case equalWord(tokens[index], "graveyards") || equalWord(tokens[index], "graveyard"):
		return StaticDeclarationCastZoneGraveyard, 1, true
	case equalWord(tokens[index], "libraries") || equalWord(tokens[index], "library"):
		return StaticDeclarationCastZoneLibrary, 1, true
	case equalWord(tokens[index], "exile"):
		return StaticDeclarationCastZoneExile, 1, true
	case staticWordsAt(tokens, index, "command", "zone") || staticWordsAt(tokens, index, "command", "zones"):
		return StaticDeclarationCastZoneCommand, 2, true
	default:
		return "", 0, false
	}
}

// parseStaticEnterBattlefieldRestrictionDeclaration recognizes the continuous
// entry restriction family "<filter> cards in <zones> can't enter the
// battlefield." (Grafdigger's Cage, Soulless Jailer, Weathered Runestone,
// Kunoros). <filter> is "creature", "permanent", or "nonland permanent"; <zones>
// is a comma/"and"/"or"-joined list of "graveyards" and/or "libraries". The
// restriction is global. Any deviation leaves the clause unconsumed and fails
// closed.
func parseStaticEnterBattlefieldRestrictionDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 8 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	end := len(tokens) - 1
	index := 0
	filter, ok := parseStaticEnterRestrictionFilter(tokens, &index)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, index, "cards", "in") {
		return StaticDeclarationSyntax{}, false
	}
	index += 2
	zones, next, ok := parseStaticEnterRestrictionZones(tokens, index, end)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	index = next
	if !staticWordsAt(tokens, index, "can't", "enter", "the", "battlefield") &&
		!staticWordsAt(tokens, index, "cannot", "enter", "the", "battlefield") {
		return StaticDeclarationSyntax{}, false
	}
	index += 4
	if index != end {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                   StaticDeclarationEnterBattlefieldRestriction,
		Span:                   shared.SpanOf(tokens),
		OperationSpan:          shared.SpanOf(tokens[:end]),
		EnterRestrictFilter:    filter,
		EnterRestrictFromZones: zones,
	}, true
}

// parseStaticEnterRestrictionFilter recognizes the leading card filter of an
// entry restriction ("creature", "permanent", or "nonland permanent"), advancing
// index past the consumed filter words.
func parseStaticEnterRestrictionFilter(tokens []shared.Token, index *int) (StaticDeclarationEnterFilterKind, bool) {
	switch {
	case staticWordsAt(tokens, *index, "nonland", "permanent"):
		*index += 2
		return StaticDeclarationEnterFilterNonlandPermanent, true
	case staticWordsAt(tokens, *index, "permanent"):
		*index++
		return StaticDeclarationEnterFilterPermanent, true
	case staticWordsAt(tokens, *index, "creature"):
		*index++
		return StaticDeclarationEnterFilterCreature, true
	default:
		return "", false
	}
}

// parseStaticEnterRestrictionZones consumes a comma-, "and"-, and/or "or"-joined
// list of "graveyards" and "libraries", returning the recognized zones and the
// index after the list. It fails closed on an empty or unrecognized list.
func parseStaticEnterRestrictionZones(tokens []shared.Token, index, end int) ([]StaticDeclarationCastZoneKind, int, bool) {
	var zones []StaticDeclarationCastZoneKind
	for index < end {
		castZone, consumed, zok := staticCastZoneWord(tokens, index)
		if !zok || (castZone != StaticDeclarationCastZoneGraveyard && castZone != StaticDeclarationCastZoneLibrary) {
			break
		}
		zones = append(zones, castZone)
		index += consumed
		separated := false
		if index < end && tokens[index].Kind == shared.Comma {
			index++
			separated = true
		}
		if index < end && (equalWord(tokens[index], "and") || equalWord(tokens[index], "or")) {
			index++
			separated = true
		}
		if !separated {
			break
		}
	}
	if len(zones) == 0 {
		return nil, 0, false
	}
	return zones, index, true
}

// parseStaticCardTypeList consumes a comma- and/or "or"/"and"-separated list of
// pluralized card-type words ("artifacts, creatures, or enchantments") into typed
// card types, returning the index after the list.
func parseStaticCardTypeList(tokens []shared.Token, index, end int) ([]CardType, int, bool) {
	var cardTypes []CardType
	for index < end {
		if tokens[index].Kind != shared.Word {
			break
		}
		cardType, ok := recognizeCardTypeWord(tokens[index].Text)
		if !ok {
			break
		}
		cardTypes = append(cardTypes, cardType)
		index++
		separated := false
		if index < end && tokens[index].Kind == shared.Comma {
			index++
			separated = true
		}
		if index < end && (equalWord(tokens[index], "or") || equalWord(tokens[index], "and")) {
			index++
			separated = true
		}
		if !separated {
			break
		}
	}
	if len(cardTypes) == 0 {
		return nil, 0, false
	}
	return cardTypes, index, true
}

var staticPlayerRuleParsers = []staticPlayerRuleParser{
	parseStaticNoMaximumHandSizeDeclaration,
	parseStaticSkipDrawStepDeclaration,
	parseStaticAttackTaxDeclaration,
	parseStaticAdditionalLandPlaysDeclaration,
	parseStaticEachPlayerAdditionalLandPlaysDeclaration,
	parseStaticPlayLandsFromGraveyardDeclaration,
	parseStaticPlayAndCastFromExileWithCounterDeclaration,
	parseStaticPlayCardFromExileWithCounterDeclaration,
	parseStaticPlayLandsFromLibraryTopDeclaration,
	parseStaticPlayWithTopCardRevealedDeclaration,
	parseStaticCastSpellsFromLibraryTopDeclaration,
	parseStaticLookAtTopCardAnyTimeDeclaration,
	parseStaticLifeForColoredManaDeclaration,
	parseStaticLifeForCommanderTaxDeclaration,
	parseStaticPlayerHexproofDeclaration,
	parseStaticPlayerShroudDeclaration,
	parseStaticDamageDoesntCauseLifeLossDeclaration,
	parseStaticRedirectDamageToSourceDeclaration,
	parseStaticActivateAbilitiesAsThoughHasteDeclaration,
}

func parseStaticPlayerRuleDeclaration(tokens []shared.Token, conditions []ConditionClause) (StaticDeclarationSyntax, bool) {
	// Try the exact body recognizers first, so a rule that does not accept a gate
	// keeps its existing unconditional-only behavior.
	for _, parse := range staticPlayerRuleParsers {
		if declaration, ok := parse(tokens); ok {
			return declaration, true
		}
	}
	// A player-scoped rule that accepts an "as long as <condition>" gate ("As long
	// as you're the monarch, damage doesn't cause you to lose life.", Archon of
	// Coronation) is retried after stripping the leading condition-covered tokens,
	// recording the gate so lowering keeps it as the static ability's condition.
	// Only rules that support a gate are accepted with a condition, so a
	// conditional near-miss of an unconditional rule still fails closed.
	opTokens, condition, hasCondition := staticOperationTokens(tokens, conditions)
	if !hasCondition {
		return StaticDeclarationSyntax{}, false
	}
	for _, parse := range staticPlayerRuleParsers {
		if declaration, ok := parse(opTokens); ok {
			if !staticPlayerRuleAcceptsCondition(declaration.PlayerRule) {
				return StaticDeclarationSyntax{}, false
			}
			declaration.Span = shared.SpanOf(tokens)
			declaration.HasCondition = true
			declaration.ConditionSpan = condition.Span
			return declaration, true
		}
	}
	return StaticDeclarationSyntax{}, false
}

// staticPlayerRuleAcceptsCondition reports whether a player-scoped static rule may
// carry a leading "as long as <condition>" designation gate. Only the
// damage-doesn't-cause-life-loss rule does today (Archon of Coronation); every
// other player rule stays unconditional so a conditional near-miss fails closed.
func staticPlayerRuleAcceptsCondition(kind StaticDeclarationPlayerRuleKind) bool {
	return kind == StaticDeclarationPlayerRuleDamageDoesntCauseLifeLoss
}

// parseStaticNoMaximumHandSizeDeclaration recognizes the exact controller-scoped
// no-maximum-hand-size rule.
func parseStaticNoMaximumHandSizeDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 7 || tokens[6].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "have", "no", "maximum", "hand", "size") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[1:6]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[0].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleNoMaximumHandSize,
	}, true
}

// parseStaticSkipDrawStepDeclaration recognizes the exact controller-scoped
// "Skip your draw step." turn-structure rule (Necropotence, Yawgmoth's Bargain).
func parseStaticSkipDrawStepDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 5 || tokens[4].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "skip", "your", "draw", "step") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[0:4]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[1].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleSkipDrawStep,
	}, true
}

// parseStaticActivateAbilitiesAsThoughHasteDeclaration recognizes the exact
// controller-scoped activation permission "You may activate abilities of
// creatures you control as though those creatures had haste." (Thousand-Year
// Elixir, Shang-Chi, Tyvar). It removes the summoning-sickness restriction on
// {T} and {Q} activation costs for the controller's creatures.
func parseStaticActivateAbilitiesAsThoughHasteDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 15 || tokens[14].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0,
		"you", "may", "activate", "abilities", "of", "creatures", "you", "control",
		"as", "though", "those", "creatures", "had", "haste") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[1:14]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[0].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleActivateAbilitiesAsThoughHaste,
	}, true
}

// parseStaticRedirectDamageToSourceDeclaration recognizes the exact
// controller-scoped damage-redirection static "All damage that would be dealt to
// you is dealt to this creature/permanent instead." (Protector of the Crown). The
// redirect target is the ability's own source permanent, which the runtime
// resolves from the rule effect's source, so no object is captured here.
func parseStaticRedirectDamageToSourceDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 15 || tokens[14].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "all", "damage", "that", "would", "be", "dealt", "to", "you", "is", "dealt", "to", "this") {
		return StaticDeclarationSyntax{}, false
	}
	if !equalWord(tokens[12], "creature") && !equalWord(tokens[12], "permanent") {
		return StaticDeclarationSyntax{}, false
	}
	if !equalWord(tokens[13], "instead") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[0:14]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[7].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleRedirectDamageToSource,
	}, true
}

// parseStaticDamageDoesntCauseLifeLossDeclaration recognizes the exact
// controller-scoped static "damage doesn't cause you to lose life." (Archon of
// Coronation, usually gated by a leading "As long as you're the monarch,"
// condition the static-declaration machinery captures separately).
func parseStaticDamageDoesntCauseLifeLossDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 8 || tokens[7].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "damage", "doesn't", "cause", "you", "to", "lose", "life") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[0:7]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[3].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleDamageDoesntCauseLifeLoss,
	}, true
}

// parseStaticPlayerHexproofDeclaration recognizes the exact controller-scoped
// player protection static "You have hexproof." (Aegis of the Gods, Leyline of
// Sanctity, Spirit of the Hearth).
func parseStaticPlayerHexproofDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 4 || tokens[3].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "have", "hexproof") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[0:3]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[0].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleHexproof,
	}, true
}

// parseStaticPlayerShroudDeclaration recognizes the exact controller-scoped
// player protection static "You have shroud." (Ivory Mask, True Believer).
func parseStaticPlayerShroudDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 4 || tokens[3].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "have", "shroud") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[0:3]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[0].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleShroud,
	}, true
}

// parseStaticAttackTaxDeclaration recognizes the exact fixed-generic attack tax
// "Creatures can't attack you unless their controller pays {N} for each creature
// they control that's attacking you." The affected player is the static ability's
// controller; the cost is paid independently for each declared attacker.
func parseStaticAttackTaxDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 18 ||
		tokens[8].Kind != shared.Symbol ||
		tokens[17].Kind != shared.Period ||
		!staticWordsAt(tokens, 0, "creatures", "can't", "attack", "you", "unless", "their", "controller", "pays") ||
		!staticWordsAt(tokens, 9, "for", "each", "creature", "they", "control", "that's", "attacking", "you") {
		return StaticDeclarationSyntax{}, false
	}
	amount, ok := staticGenericSymbolValue(tokens[8].Text)
	if !ok || amount <= 0 {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[1:17]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[3].Span,
		},
		PlayerRule:       StaticDeclarationPlayerRuleAttackTax,
		AttackTaxGeneric: amount,
	}, true
}

// parseStaticAdditionalLandPlaysDeclaration recognizes the controller-scoped
// static grant of one or more extra land plays every turn: "You may play an
// additional land on each of your turns." and the multi-land "... two additional
// lands ..." variant. The "you may" permission is folded into an unconditional
// allowance; the controller still chooses whether to play the extra land.
func parseStaticAdditionalLandPlaysDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 12 || tokens[11].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "may", "play") {
		return StaticDeclarationSyntax{}, false
	}
	count, ok := additionalLandCountWord(tokens[3])
	if !ok || count <= 0 || !equalWord(tokens[4], "additional") {
		return StaticDeclarationSyntax{}, false
	}
	landWord := "land"
	if count != 1 {
		landWord = "lands"
	}
	if !equalWord(tokens[5], landWord) ||
		!staticWordsAt(tokens, 6, "on", "each", "of", "your", "turns") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[1:11]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[0].Span,
		},
		PlayerRule:          StaticDeclarationPlayerRuleAdditionalLandPlays,
		AdditionalLandPlays: count,
	}, true
}

// parseStaticEachPlayerAdditionalLandPlaysDeclaration recognizes the symmetric
// all-players grant of one or more extra land plays every turn: "Each player may
// play an additional land on each of their turns." and the multi-land "... two
// additional lands ..." variant (Rites of Flourishing, Ghirapur Orrery). The
// each-player subject distinguishes it from the controller-scoped form so the
// allowance is granted to every player rather than only the source's controller.
func parseStaticEachPlayerAdditionalLandPlaysDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 13 || tokens[12].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "each", "player", "may", "play") {
		return StaticDeclarationSyntax{}, false
	}
	count, ok := additionalLandCountWord(tokens[4])
	if !ok || count <= 0 || !equalWord(tokens[5], "additional") {
		return StaticDeclarationSyntax{}, false
	}
	landWord := "land"
	if count != 1 {
		landWord = "lands"
	}
	if !equalWord(tokens[6], landWord) ||
		!staticWordsAt(tokens, 7, "on", "each", "of", "their", "turns") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[2:12]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectEachPlayer,
			Span: shared.SpanOf(tokens[0:2]),
		},
		PlayerRule:          StaticDeclarationPlayerRuleAdditionalLandPlays,
		AdditionalLandPlays: count,
	}, true
}

// parseStaticPlayLandsFromGraveyardDeclaration recognizes the controller-scoped
// continuous permission to play land cards from the controller's graveyard ("You
// may play lands from your graveyard.", Ramunap Excavator, Crucible of Worlds).
// The "you may" permission is folded into an unconditional allowance; the
// controller still chooses whether to play such a land.
func parseStaticPlayLandsFromGraveyardDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 8 || tokens[7].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "may", "play", "lands", "from", "your", "graveyard") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[1:7]),
		Subject: StaticDeclarationSubject{
			Kind:       StaticDeclarationSubjectController,
			Span:       tokens[0].Span,
			CardFilter: StaticDeclarationCardFilterLand,
		},
		PlayerRule: StaticDeclarationPlayerRulePlayLandsFromGraveyard,
	}, true
}

// parseStaticPlayAndCastFromExileWithCounterDeclaration recognizes the
// controller-scoped continuous permission to play land cards and cast spells
// from among the cards the controller owns in exile that carry a named marker
// counter ("You may play lands and cast spells from among cards you own in exile
// with croak counters on them.", Grolnok, the Omnivore). The counter name is read
// text-blind between "with" and the "counters" noun; any named marker counter is
// accepted, so the permission generalizes across cards. The "you may" permission
// is folded into an unconditional allowance; the controller still chooses whether
// to play or cast such a card.
func parseStaticPlayAndCastFromExileWithCounterDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 20 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "may", "play", "lands", "and", "cast", "spells",
		"from", "among", "cards", "you", "own", "in", "exile", "with") {
		return StaticDeclarationSyntax{}, false
	}
	counterIndex := -1
	for i := 16; i < len(tokens); i++ {
		if equalWord(tokens[i], "counters") {
			counterIndex = i
			break
		}
	}
	if counterIndex < 0 {
		return StaticDeclarationSyntax{}, false
	}
	kind, span, ok := counterNameBefore(tokens, counterIndex)
	if !ok || span.Start.Offset != tokens[15].Span.Start.Offset {
		return StaticDeclarationSyntax{}, false
	}
	if counterIndex+4 != len(tokens) || !staticWordsAt(tokens, counterIndex+1, "on", "them") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[1 : counterIndex+3]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[0].Span,
		},
		PlayerRule:   StaticDeclarationPlayerRulePlayAndCastFromExileWithCounter,
		ExileCounter: kind,
	}, true
}

// parseStaticPlayCardFromExileWithCounterDeclaration recognizes the
// controller-scoped continuous permission to play a card from exile that carries
// a named marker counter, spelled "you may play a card from exile with a
// <counter> counter on it" ("Once each turn, you may play a card from exile with
// a collection counter on it if it was exiled by an ability you controlled, and
// you may spend mana as though it were mana of any color to cast it.", Evelyn, the
// Covetous). The counter name is read text-blind between the "with a/an" article
// and the "counter" noun, so any named marker counter is accepted. Three optional
// riders are captured as independent typed flags: a leading "Once each turn,"
// (ExilePlayOncePerTurn), an "if it was exiled by an ability you controlled"
// provenance clause (ExilePlayExiledByControlledAbility), and a trailing "and you
// may spend mana as though it were mana of any color to cast it" permission
// (ExilePlaySpendAnyColorMana). Because each rider maps to its own flag, wordings
// that carry only some of them lower correctly. The declaration reuses the
// play-and-cast-from-exile-with-counter player rule; the "you may" permission is
// folded into an allowance the controller may exercise.
func parseStaticPlayCardFromExileWithCounterDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	coreStart := 0
	oncePerTurn := false
	if staticWordsAt(tokens, 0, "once", "each", "turn") && len(tokens) > 3 && tokens[3].Kind == shared.Comma {
		oncePerTurn = true
		coreStart = 4
	}
	if !staticWordsAt(tokens, coreStart, "you", "may", "play", "a", "card", "from", "exile", "with") {
		return StaticDeclarationSyntax{}, false
	}
	articleIndex := coreStart + 8
	if articleIndex >= len(tokens) || (!equalWord(tokens[articleIndex], "a") && !equalWord(tokens[articleIndex], "an")) {
		return StaticDeclarationSyntax{}, false
	}
	nameStart := articleIndex + 1
	counterIndex := -1
	for i := nameStart; i < len(tokens); i++ {
		if equalWord(tokens[i], "counter") || equalWord(tokens[i], "counters") {
			counterIndex = i
			break
		}
	}
	if counterIndex <= nameStart {
		return StaticDeclarationSyntax{}, false
	}
	kind, span, ok := counterNameBefore(tokens, counterIndex)
	if !ok || span.Start.Offset != tokens[nameStart].Span.Start.Offset {
		return StaticDeclarationSyntax{}, false
	}
	if counterIndex+2 >= len(tokens) || !equalWord(tokens[counterIndex+1], "on") ||
		(!equalWord(tokens[counterIndex+2], "it") && !equalWord(tokens[counterIndex+2], "them")) {
		return StaticDeclarationSyntax{}, false
	}
	pos := counterIndex + 3
	exiledByControlled := false
	if staticWordsAt(tokens, pos, "if", "it", "was", "exiled", "by", "an", "ability", "you", "controlled") {
		exiledByControlled = true
		pos += 9
	}
	spendAnyColor := false
	anyColorStart := pos
	if pos < len(tokens) && tokens[pos].Kind == shared.Comma && equalWord(tokens[pos+1], "and") {
		anyColorStart = pos + 2
	}
	if staticWordsAt(tokens, anyColorStart, "you", "may", "spend", "mana", "as", "though",
		"it", "were", "mana", "of", "any", "color", "to", "cast", "it") {
		spendAnyColor = true
		pos = anyColorStart + 15
	}
	if pos != len(tokens)-1 || tokens[pos].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[coreStart : counterIndex+3]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[coreStart].Span,
		},
		PlayerRule:                         StaticDeclarationPlayerRulePlayAndCastFromExileWithCounter,
		ExileCounter:                       kind,
		ExilePlayOncePerTurn:               oncePerTurn,
		ExilePlayExiledByControlledAbility: exiledByControlled,
		ExilePlaySpendAnyColorMana:         spendAnyColor,
	}, true
}

// parseStaticCastThisFromGraveyardDeclaration recognizes the controller-scoped
// continuous permission to cast the source card itself from the controller's
// graveyard ("You may cast this card from your graveyard.", Hogaak). An optional
// "as long as <condition>" clause gates the permission ("... as long as you
// control a Zombie.", Gravecrawler); the gate is captured as the declaration's
// condition. The "you may" permission is folded into an allowance; the controller
// still chooses whether to cast the card.
func parseStaticCastThisFromGraveyardDeclaration(tokens []shared.Token, conditions []ConditionClause) (StaticDeclarationSyntax, bool) {
	span := shared.SpanOf(tokens)
	opTokens := tokens
	condition, hasCondition := staticDeclarationCondition(tokens, conditions)
	if hasCondition {
		opTokens = tokensOutsideCondition(tokens, condition.Span)
	}
	if len(opTokens) != 9 || opTokens[8].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(opTokens, 0, "you", "may", "cast", "this", "card", "from", "your", "graveyard") {
		return StaticDeclarationSyntax{}, false
	}
	declaration := StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          span,
		OperationSpan: shared.SpanOf(opTokens[1:8]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: opTokens[0].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleCastThisFromGraveyard,
	}
	if hasCondition {
		declaration.HasCondition = true
		declaration.ConditionSpan = condition.Span
	}
	return declaration, true
}

// parseStaticCastThisFromExileDeclaration recognizes the controller-scoped
// continuous permission to cast the source card itself from exile ("You may cast
// this card from exile.", Misthollow Griffin, Eternal Scourge). An optional "as
// long as <condition>" clause gates the permission; the gate is captured as the
// declaration's condition. The "you may" permission is folded into an allowance;
// the controller still chooses whether to cast the card.
func parseStaticCastThisFromExileDeclaration(tokens []shared.Token, conditions []ConditionClause) (StaticDeclarationSyntax, bool) {
	span := shared.SpanOf(tokens)
	opTokens := tokens
	condition, hasCondition := staticDeclarationCondition(tokens, conditions)
	if hasCondition {
		opTokens = tokensOutsideCondition(tokens, condition.Span)
	}
	if len(opTokens) != 8 || opTokens[7].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(opTokens, 0, "you", "may", "cast", "this", "card", "from", "exile") {
		return StaticDeclarationSyntax{}, false
	}
	declaration := StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          span,
		OperationSpan: shared.SpanOf(opTokens[1:7]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: opTokens[0].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleCastThisFromExile,
	}
	if hasCondition {
		declaration.HasCondition = true
		declaration.ConditionSpan = condition.Span
	}
	return declaration, true
}

// continuous permission to play land cards from the top of the controller's
// library ("You may play lands from the top of your library.", Oracle of Mul
// Daya, Courser of Kruphix).
func parseStaticPlayLandsFromLibraryTopDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 11 || tokens[10].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "may", "play", "lands", "from", "the", "top", "of", "your", "library") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[1:10]),
		Subject: StaticDeclarationSubject{
			Kind:       StaticDeclarationSubjectController,
			Span:       tokens[0].Span,
			CardFilter: StaticDeclarationCardFilterLand,
		},
		PlayerRule: StaticDeclarationPlayerRulePlayLandsFromLibraryTop,
	}, true
}

// parseStaticPlayWithTopCardRevealedDeclaration recognizes the controller-scoped
// visibility static that reveals the top card of the controller's library ("Play
// with the top card of your library revealed.", Oracle of Mul Daya, Courser of
// Kruphix, Future Sight).
func parseStaticPlayWithTopCardRevealedDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 10 || tokens[9].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "play", "with", "the", "top", "card", "of", "your", "library", "revealed") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[0:9]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[6].Span,
		},
		PlayerRule: StaticDeclarationPlayerRulePlayWithTopCardRevealed,
	}, true
}

// parseStaticLookAtTopCardAnyTimeDeclaration recognizes the controller-scoped
// private-visibility static "You may look at the top card of your library any
// time." (Bolas's Citadel, Vizier of the Menagerie, Sphinx of Jwar Isle).
func parseStaticLookAtTopCardAnyTimeDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 13 || tokens[12].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "may", "look", "at", "the", "top", "card", "of", "your", "library", "any", "time") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[1:12]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[0].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleLookAtTopCardAnyTime,
	}, true
}

// parseStaticLifeForColoredManaDeclaration recognizes the exact controller-scoped
// life-for-mana substitution "For each {C} in a cost, you may pay 2 life rather
// than pay that mana." (K'rrik, Son of Yawgmoth), where {C} is a single colored
// mana symbol. It records the colored mana symbol so lowering can scope the
// payment permission to that color.
func parseStaticLifeForColoredManaDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 18 ||
		tokens[2].Kind != shared.Symbol ||
		tokens[6].Kind != shared.Comma ||
		tokens[10].Kind != shared.Integer || tokens[10].Text != "2" ||
		tokens[17].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "for", "each") ||
		!staticWordsAt(tokens, 3, "in", "a", "cost") ||
		!staticWordsAt(tokens, 7, "you", "may", "pay") ||
		!staticWordsAt(tokens, 11, "life", "rather", "than", "pay", "that", "mana") {
		return StaticDeclarationSyntax{}, false
	}
	inner, ok := strings.CutPrefix(tokens[2].Text, "{")
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	inner, ok = strings.CutSuffix(inner, "}")
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	color, ok := keywordManaColor(inner)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[7].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleLifeForColoredMana,
		ManaColor:  color,
	}, true
}

// parseStaticLifeForCommanderTaxDeclaration recognizes the exact self-scoped
// command-zone tax substitution "Rather than pay {2} for each previous time
// you've cast this spell from the command zone this game, pay 2 life that many
// times." (Liesa, Shroud of Dusk), which lets the caster pay 2 life rather than
// each {2} of the commander tax.
func parseStaticLifeForCommanderTaxDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 26 ||
		tokens[3].Kind != shared.Symbol || tokens[3].Text != "{2}" ||
		tokens[18].Kind != shared.Comma ||
		tokens[20].Kind != shared.Integer || tokens[20].Text != "2" ||
		tokens[25].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "rather", "than", "pay") ||
		!staticWordsAt(tokens, 4, "for", "each", "previous", "time", "you've", "cast", "this", "spell", "from", "the", "command", "zone", "this", "game") ||
		!staticWordsAt(tokens, 19, "pay") ||
		!staticWordsAt(tokens, 21, "life", "that", "many", "times") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[8].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleLifeForCommanderTax,
	}, true
}

// parseStaticCastSpellsFromLibraryTopDeclaration recognizes the controller-scoped
// continuous permission to cast spells from the top of the controller's library:
// "You may cast spells from the top of your library." (Bolas's Citadel), the
// typed "You may cast <types> spells from the top of your library." (Vizier of
// the Menagerie, Precognition Field), and the combined "You may play lands and
// cast spells from the top of your library." (Future Sight). The optional card
// type list is reconstructed into typed card types; the combined wording records
// AlsoPlayLands so lowering also grants the land-play permission. A trailing "of
// the chosen type" qualifier ("You may cast creature spells of the chosen type
// from the top of your library.", Realmwalker) records CastChosenCreatureType so
// lowering narrows the permission to the source permanent's entry-chosen creature
// subtype.
func parseStaticCastSpellsFromLibraryTopDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 11 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	core, payLife, ok := splitStaticCastFromTopPayLifeRider(tokens)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	fullSpan := shared.SpanOf(tokens)
	tokens = core
	if len(tokens) < 11 {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "may") {
		return StaticDeclarationSyntax{}, false
	}
	index := 2
	alsoPlayLands := false
	switch {
	case staticWordsAt(tokens, index, "play", "lands", "and", "cast"):
		alsoPlayLands = true
		index += 4
	case index < len(tokens) && equalWord(tokens[index], "cast"):
		index++
	default:
		return StaticDeclarationSyntax{}, false
	}
	filter, ok := parseCastSpellTypeList(tokens, index, len(tokens)-1)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	index = filter.next
	chosenCreatureType := false
	if staticWordsAt(tokens, index, "of", "the", "chosen", "type") {
		chosenCreatureType = true
		index += 4
	}
	if index != len(tokens)-7 ||
		!staticWordsAt(tokens, index, "from", "the", "top", "of", "your", "library") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          fullSpan,
		OperationSpan: shared.SpanOf(tokens[1 : len(tokens)-1]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[0].Span,
		},
		PlayerRule:             StaticDeclarationPlayerRuleCastSpellsFromLibraryTop,
		CastSpellTypes:         filter.cardTypes,
		CastColorless:          filter.colorless,
		AlsoPlayLands:          alsoPlayLands,
		CastChosenCreatureType: chosenCreatureType,
		CastPayLifeManaValue:   payLife,
	}, true
}

// splitStaticCastFromTopPayLifeRider separates the leading cast-from-library-top
// sentence of a static cast permission from an optional trailing "If you cast a
// spell this way, pay life equal to its mana value rather than pay its mana
// cost." rider (Bolas's Citadel). The leading sentence carries no internal
// period, so it ends at the first period; any remaining tokens must form exactly
// the pay-life rider sentence. It returns the leading sentence's tokens, whether
// the rider was present, and false when trailing tokens exist that are not the
// recognized rider so the declaration fails closed rather than dropping text.
func splitStaticCastFromTopPayLifeRider(tokens []shared.Token) (core []shared.Token, payLife, ok bool) {
	first := -1
	for i := range tokens {
		if tokens[i].Kind == shared.Period {
			first = i
			break
		}
	}
	if first < 0 {
		return nil, false, false
	}
	core = tokens[:first+1]
	rest := tokens[first+1:]
	if len(rest) == 0 {
		return core, false, true
	}
	if isPlayFromTopPayLifeRiderTokens(rest) {
		return core, true, true
	}
	return nil, false, false
}

// castSpellFilter is the parsed card-type and color filter of a
// cast-from-library-top declaration, plus the token index after the final
// "spells".
type castSpellFilter struct {
	cardTypes []CardType
	colorless bool
	next      int
}

// parseCastSpellTypeList consumes the card-type and color filter that precedes
// "spells" in a cast-from-library-top declaration: the bare "spells" (no filter),
// "<type> spells", "<t1> and <t2> spells" (shared trailing "spells"), the
// colorless filter "colorless spells", or a combined "<type> spells and colorless
// spells" with a repeated "spells". It returns the recognized card types, whether
// a colorless filter was present, and the index after the final "spells", failing
// closed on any word that is neither a card type nor "colorless" and on a missing
// "spells". It also returns at a trailing "of" (the "of the chosen type"
// qualifier the caller handles separately) once "spells" has been matched.
func parseCastSpellTypeList(tokens []shared.Token, index, end int) (castSpellFilter, bool) {
	var cardTypes []CardType
	colorless := false
	matchedSpells := false
	for index < end {
		token := tokens[index]
		switch {
		case equalWord(token, "from"):
			if !matchedSpells {
				return castSpellFilter{}, false
			}
			return castSpellFilter{cardTypes: cardTypes, colorless: colorless, next: index}, true
		case equalWord(token, "of"):
			if !matchedSpells {
				return castSpellFilter{}, false
			}
			return castSpellFilter{cardTypes: cardTypes, colorless: colorless, next: index}, true
		case equalWord(token, "spells"):
			matchedSpells = true
		case equalWord(token, "and") || equalWord(token, "or"):
		case token.Kind == shared.Comma:
		case equalWord(token, "colorless"):
			colorless = true
		default:
			cardType, ok := recognizeCardTypeWord(token.Text)
			if !ok {
				return castSpellFilter{}, false
			}
			cardTypes = append(cardTypes, cardType)
		}
		index++
	}
	return castSpellFilter{}, false
}

// staticDeclarationCondition returns the single condition clause that lies within
// the declaration body, if exactly one is present.
func staticDeclarationCondition(tokens []shared.Token, conditions []ConditionClause) (ConditionClause, bool) {
	body := shared.SpanOf(tokens)
	matched := -1
	for i := range conditions {
		if spanCovers(body, conditions[i].Span) {
			if matched >= 0 {
				return ConditionClause{}, false
			}
			matched = i
		}
	}
	if matched < 0 {
		return ConditionClause{}, false
	}
	return conditions[matched], true
}

// tokensOutsideCondition removes a condition clause's tokens from the body and
// drops a comma left dangling by a leading condition.
func tokensOutsideCondition(tokens []shared.Token, span shared.Span) []shared.Token {
	result := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if !spanCovers(span, token.Span) {
			result = append(result, token)
		}
	}
	if len(result) > 0 && result[0].Kind == shared.Comma {
		result = result[1:]
	}
	return result
}

func staticOperationTokens(tokens []shared.Token, conditions []ConditionClause) ([]shared.Token, ConditionClause, bool) {
	condition, ok := staticDeclarationCondition(tokens, conditions)
	if !ok {
		return tokens, ConditionClause{}, false
	}
	return tokensOutsideCondition(tokens, condition.Span), condition, true
}

func parseStaticCostModifierDeclaration(
	tokens []shared.Token,
	atoms Atoms,
	conditions []ConditionClause,
) (StaticDeclarationSyntax, bool) {
	span := shared.SpanOf(tokens)
	opTokens, condition, hasCondition := staticOperationTokens(tokens, conditions)
	if len(opTokens) == 0 || opTokens[len(opTokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	keyword, ok := staticSoleBareCyclingKeyword(opTokens, atoms)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	declaration := StaticDeclarationSyntax{
		Kind: StaticDeclarationCostModifier,
		Span: span,
	}
	if hasCondition {
		declaration.HasCondition = true
		declaration.ConditionSpan = condition.Span
	}
	if reduction, ok := parseStaticAbilityReduction(opTokens, keyword); ok {
		declaration.CostModifier = StaticDeclarationCostModifierAbilityReduction
		declaration.CostReductionAmount = reduction
		declaration.OperationSpan = keyword.Span
		return declaration, true
	}
	if replacement, ok := parseStaticReplaceCyclingCost(opTokens, keyword); ok {
		declaration.CostModifier = StaticDeclarationCostModifierReplaceCost
		declaration.CostReplacement = replacement
		declaration.OperationSpan = keyword.Span
		return declaration, true
	}
	if replacement, ok := parseStaticReplaceFirstCyclingCost(opTokens, keyword); ok {
		declaration.CostModifier = StaticDeclarationCostModifierReplaceFirstCost
		declaration.CostReplacement = replacement
		declaration.OperationSpan = keyword.Span
		return declaration, true
	}
	return StaticDeclarationSyntax{}, false
}

// parseStaticAbilityCostSetDeclaration recognizes the static ability-cost setting
// "Equipment you control have equip {N}." that fixes the Equip activation cost of
// the controller's Equipment to {N} (commonly {0}). An optional "as long as ..."
// condition clause gates the static and is split off before the operation tokens.
// CostReplacement carries the canonical replacement mana cost; the free {0} cost
// reduces to an empty string.
func parseStaticAbilityCostSetDeclaration(
	tokens []shared.Token,
	conditions []ConditionClause,
) (StaticDeclarationSyntax, bool) {
	span := shared.SpanOf(tokens)
	opTokens, condition, hasCondition := staticOperationTokens(tokens, conditions)
	if len(opTokens) != 7 ||
		opTokens[6].Kind != shared.Period ||
		opTokens[5].Kind != shared.Symbol ||
		!staticWordsAt(opTokens, 0, "equipment", "you", "control", "have", "equip") {
		return StaticDeclarationSyntax{}, false
	}
	replacement, ok := staticReplacementCost(opTokens[5].Text)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	declaration := StaticDeclarationSyntax{
		Kind:               StaticDeclarationAbilityCostSet,
		Span:               span,
		OperationSpan:      shared.SpanOf(opTokens[:6]),
		AbilityCostKeyword: KeywordEquip,
		CostReplacement:    replacement,
	}
	if hasCondition {
		declaration.HasCondition = true
		declaration.ConditionSpan = condition.Span
	}
	return declaration, true
}

// parseStaticSpellCostModifierDeclaration recognizes the static cast-cost
// modifier "[<filter>] spells you cast [from <zone>] cost {N} less/more to cast."
// where the optional leading filter constrains the affected spells. The filter
// combines an optional leading color word (one of the five colors or colorless)
// with an optional single card-type word, an "instant and sorcery" pair, or a
// subtype list joined by "and" ("Aura and Equipment"). A color word may precede
// a card type ("Black creature spells"). An optional cast-zone scope narrows the
// discount to spells cast "from your graveyard" or "from anywhere other than your
// hand"; any other zone wording fails closed. The affected group is always the
// static ability's controller's spells.
func parseStaticSpellCostModifierDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if declaration, ok := parseChosenCreatureTypeSpellCostReduction(tokens); ok {
		return declaration, true
	}
	if declaration, ok := parseStaticSpellSharedExiledTypeCostReduction(tokens); ok {
		return declaration, true
	}
	if declaration, ok := parseStaticSpellPerObjectCostReduction(tokens, atoms); ok {
		return declaration, true
	}
	if declaration, ok := parseStaticSpellColorDisjunctionCostModifier(tokens); ok {
		return declaration, true
	}
	if declaration, ok := parseStaticSpellColorPairCostModifier(tokens); ok {
		return declaration, true
	}
	if declaration, ok := parseStaticSpellColorTypePairCostModifier(tokens); ok {
		return declaration, true
	}
	if declaration, ok := parseStaticSpellTargetsSourceCostModifier(tokens, atoms); ok {
		return declaration, true
	}
	rest := tokens
	spellColor := staticSpellColorWord(rest[0])
	if spellColor != StaticDeclarationSpellColorNone {
		rest = rest[1:]
	}
	spellType := StaticDeclarationSpellTypeAll
	var subtypes []types.Sub
	var excludedTypes []CardType
	var requiredTypes []CardType
	if subs, next, ok := staticSpellSubtypeFilter(rest, atoms); ok {
		subtypes = subs
		rest = next
	} else if excluded, next, ok := staticSpellExcludedTypeFilter(rest); ok {
		excludedTypes = excluded
		rest = next
	} else if filterType, next, ok := staticSpellTypeFilter(rest); ok {
		spellType = filterType
		rest = next
	} else if list, next, ok := staticSpellCardTypeList(rest); ok && spellColor == StaticDeclarationSpellColorNone {
		requiredTypes = list
		rest = next
	} else {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(rest, 0, "spells") {
		return StaticDeclarationSyntax{}, false
	}
	rest = rest[1:]
	var caster StaticDeclarationSpellCasterKind
	switch {
	case staticWordsAt(rest, 0, "you", "cast"):
		caster = StaticDeclarationSpellCasterController
		rest = rest[2:]
	case staticWordsAt(rest, 0, "your", "opponents", "cast"):
		caster = StaticDeclarationSpellCasterOpponents
		rest = rest[3:]
	default:
		caster = StaticDeclarationSpellCasterAny
	}
	var castZone StaticDeclarationCastZoneKind
	switch {
	case staticWordsAt(rest, 0, "from", "your", "graveyard"):
		castZone = StaticDeclarationCastZoneGraveyard
		rest = rest[3:]
	case staticWordsAt(rest, 0, "from", "anywhere", "other", "than"):
		// "from anywhere other than your hand" scopes the discount to every
		// non-hand cast zone. Any other complement fails closed rather than
		// silently dropping the zone scope.
		index := 4
		if index < len(rest) && (equalWord(rest[index], "your") || equalWord(rest[index], "their")) {
			index++
		}
		if index >= len(rest) || (!equalWord(rest[index], "hand") && !equalWord(rest[index], "hands")) {
			return StaticDeclarationSyntax{}, false
		}
		castZone = StaticDeclarationCastZoneNonHand
		rest = rest[index+1:]
	default:
	}
	powerAtLeast, next, ok := staticSpellPowerThreshold(rest)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	rest = next
	var manaValueAtLeast int
	if powerAtLeast == 0 {
		manaValueAtLeast, next, ok = staticSpellManaValueThreshold(rest)
		if !ok {
			return StaticDeclarationSyntax{}, false
		}
		rest = next
	}
	tail, ok := staticSpellCostModifierTail(rest)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                       StaticDeclarationCostModifier,
		Span:                       shared.SpanOf(tokens),
		OperationSpan:              tail.OperationSpan,
		CostModifier:               tail.Kind,
		CostReductionAmount:        tail.Amount,
		CostIncreaseColors:         tail.IncreaseColors,
		CostIncreaseLife:           tail.LifeIncrease,
		SpellType:                  spellType,
		SpellColor:                 spellColor,
		SpellSubtypes:              subtypes,
		SpellExcludedTypes:         excludedTypes,
		SpellRequiredTypes:         requiredTypes,
		SpellCaster:                caster,
		SpellCastZone:              castZone,
		SpellPowerAtLeast:          powerAtLeast,
		MatchSpellPowerAtLeast:     powerAtLeast > 0,
		SpellManaValueAtLeast:      manaValueAtLeast,
		MatchSpellManaValueAtLeast: manaValueAtLeast > 0,
	}, true
}

// parseStaticSpellTargetsSourceCostModifier recognizes the static cast-cost
// modifier "Spells [your opponents cast | you cast] that target <source> cost
// {N} less/more to cast.", a defensive tax (or a controller discount) on spells
// that target the source permanent ("Spells your opponents cast that target
// this creature cost {2} more to cast.", Boreal Elemental; "Spells you cast that
// target this creature cost {2} less to cast.", Elderwood Scion). <source> is a
// "this creature"/"this permanent" marker or the card's own printed name (Charix,
// Syr Elenora); any other target scope (a controlled group such as "creatures
// you control") fails closed. An absent caster phrase ("Spells that target ...")
// applies to every player's spells.
func parseStaticSpellTargetsSourceCostModifier(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if !staticWordsAt(tokens, 0, "spells") {
		return StaticDeclarationSyntax{}, false
	}
	rest := tokens[1:]
	caster := StaticDeclarationSpellCasterAny
	switch {
	case staticWordsAt(rest, 0, "your", "opponents", "cast"):
		caster = StaticDeclarationSpellCasterOpponents
		rest = rest[3:]
	case staticWordsAt(rest, 0, "you", "cast"):
		caster = StaticDeclarationSpellCasterController
		rest = rest[2:]
	default:
	}
	if !staticWordsAt(rest, 0, "that", "target") {
		return StaticDeclarationSyntax{}, false
	}
	rest = rest[2:]
	width, ok := sourceNameSpanWidthAt(rest, 0, atoms)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	rest = rest[width:]
	tail, ok := staticSpellCostModifierTail(rest)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                StaticDeclarationCostModifier,
		Span:                shared.SpanOf(tokens),
		OperationSpan:       tail.OperationSpan,
		CostModifier:        tail.Kind,
		CostReductionAmount: tail.Amount,
		CostIncreaseColors:  tail.IncreaseColors,
		CostIncreaseLife:    tail.LifeIncrease,
		SpellType:           StaticDeclarationSpellTypeAll,
		SpellCaster:         caster,
		SpellTargetsSource:  true,
	}, true
}

// staticSpellPowerThreshold reads an optional "with power <n> or greater"
// qualifier scoping a cast-cost modifier to spells whose base printed power
// meets the threshold ("Creature spells you cast with power 4 or greater cost
// {2} less to cast.", Goreclaw). It returns the threshold (zero when the
// qualifier is absent) and the remaining tokens. A malformed qualifier fails.
func staticSpellPowerThreshold(tokens []shared.Token) (int, []shared.Token, bool) {
	if !staticWordsAt(tokens, 0, "with", "power") {
		return 0, tokens, true
	}
	if len(tokens) < 5 {
		return 0, nil, false
	}
	value, ok := conditionNumberValue(tokens[2])
	if !ok || value <= 0 {
		return 0, nil, false
	}
	if !staticWordsAt(tokens, 3, "or", "greater") {
		return 0, nil, false
	}
	return value, tokens[5:], true
}

// staticSpellManaValueThreshold reads an optional "with mana value <n> or
// greater" qualifier scoping a cast-cost modifier to spells whose mana value
// meets the threshold ("Creature spells you cast with mana value 6 or greater
// cost {2} less to cast.", Krosan Drover). It returns the threshold (zero when
// the qualifier is absent) and the remaining tokens. A malformed qualifier
// fails.
func staticSpellManaValueThreshold(tokens []shared.Token) (int, []shared.Token, bool) {
	if !staticWordsAt(tokens, 0, "with", "mana", "value") {
		return 0, tokens, true
	}
	if len(tokens) < 6 {
		return 0, nil, false
	}
	value, ok := conditionNumberValue(tokens[3])
	if !ok || value <= 0 {
		return 0, nil, false
	}
	if !staticWordsAt(tokens, 4, "or", "greater") {
		return 0, nil, false
	}
	return value, tokens[6:], true
}

// parseStaticSpellSharedExiledTypeCostReduction recognizes the dynamic
// controller cast-cost discount whose amount scales with the card types the
// spell shares with the cards exiled with the source permanent:
//
//	"Spells you cast cost {N} less to cast for each card type they share with cards exiled with this creature." (Cemetery Prowler)
//
// The affected spells are all the controller's spells; the trailing "for each
// card type they share with cards exiled with this creature" marks the per-
// shared-type scaling. Any other tail falls through to the fixed-amount forms.
func parseStaticSpellSharedExiledTypeCostReduction(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if !staticWordsAt(tokens, 0, "spells", "you", "cast", "cost") {
		return StaticDeclarationSyntax{}, false
	}
	rest := tokens[4:]
	if len(rest) != 17 ||
		rest[0].Kind != shared.Symbol ||
		!staticWordsAt(rest, 1, "less", "to", "cast", "for", "each", "card", "type", "they", "share", "with", "cards", "exiled", "with", "this", "creature") ||
		rest[16].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	amount, ok := staticGenericSymbolValue(rest[0].Text)
	if !ok || amount <= 0 {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                StaticDeclarationCostModifier,
		Span:                shared.SpanOf(tokens),
		OperationSpan:       shared.SpanOf(rest[0:4]),
		CostModifier:        StaticDeclarationCostModifierSpellSharedExiledTypeReduction,
		CostReductionAmount: amount,
		SpellType:           StaticDeclarationSpellTypeAll,
	}, true
}

// parseStaticSpellPerObjectCostReduction recognizes the dynamic group cast-cost
// discount whose amount scales with a countable battlefield permanent the
// controller controls:
//
//	"During your turn, spells you cast cost {1} less to cast for each creature you control with power 4 or greater." (Temur Battlecrier)
//	"Creature spells you cast cost {1} less to cast for each creature you control with a +1/+1 counter on it." (Hamza, Guardian of Arashin)
//
// The optional leading "During your turn," gate scopes the discount to the
// controller's turn; an optional single card-type word ("Creature spells")
// constrains the affected spells. The trailing "for each <permanent> you
// control[ with ...]" count subject must resolve to a battlefield permanent
// count (zone.None) so the runtime can total the controller's matching
// permanents at cost time. Card-zone counts, dynamic non-count forms, or any
// other tail fall through to the fixed-amount and shared-exiled forms.
func parseStaticSpellPerObjectCostReduction(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	idx := 0
	duringControllerTurn := false
	if staticWordsAt(tokens, idx, "during", "your", "turn") {
		duringControllerTurn = true
		idx += 3
		if idx < len(tokens) && tokens[idx].Kind == shared.Comma {
			idx++
		}
	}
	spellType := StaticDeclarationSpellTypeAll
	if filterType, rest, ok := staticSpellTypeFilter(tokens[idx:]); ok {
		spellType = filterType
		idx = len(tokens) - len(rest)
	}
	if !staticWordsAt(tokens, idx, "spells", "you", "cast", "cost") {
		return StaticDeclarationSyntax{}, false
	}
	idx += 4
	if idx >= len(tokens) || tokens[idx].Kind != shared.Symbol {
		return StaticDeclarationSyntax{}, false
	}
	amount, ok := staticGenericSymbolValue(tokens[idx].Text)
	if !ok || amount <= 0 {
		return StaticDeclarationSyntax{}, false
	}
	opSpan := tokens[idx].Span
	idx++
	if !staticWordsAt(tokens, idx, "less", "to", "cast", "for", "each") {
		return StaticDeclarationSyntax{}, false
	}
	idx += 5
	if idx >= len(tokens)-1 {
		return StaticDeclarationSyntax{}, false
	}
	subject, ok := parseDynamicCountSubject(tokens, idx, atoms)
	if !ok || !subject.count {
		return StaticDeclarationSyntax{}, false
	}
	if subject.end != len(tokens)-1 {
		return StaticDeclarationSyntax{}, false
	}
	if subject.amount.DynamicKind != EffectDynamicAmountCount ||
		subject.amount.Selection == nil ||
		subject.amount.Selection.Zone != zone.None {
		return StaticDeclarationSyntax{}, false
	}
	selection := *subject.amount.Selection
	return StaticDeclarationSyntax{
		Kind:                         StaticDeclarationCostModifier,
		Span:                         shared.SpanOf(tokens),
		OperationSpan:                opSpan,
		CostModifier:                 StaticDeclarationCostModifierSpellPerObjectReduction,
		CostReductionAmount:          amount,
		SpellType:                    spellType,
		PerObjectCountSelection:      &selection,
		RestrictDuringControllerTurn: duringControllerTurn,
	}, true
}

// parseChosenCreatureTypeSpellCostReduction recognizes the static cast-cost
// reducer filtered by the source's entry-time chosen creature type:
//
//	"Creature spells of the chosen type cost {N} less to cast." (Urza's Incubator)
//	"Creature spells you cast of the chosen type cost {N} less to cast." (Herald's Horn)
//
// The optional "you cast" qualifier does not change the affected group: the
// modifier always applies to the controller's creature spells of the chosen
// type. The trailing "cost {N} less to cast" carries the reduction amount.
func parseChosenCreatureTypeSpellCostReduction(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if !staticWordsAt(tokens, 0, "creature", "spells") {
		return StaticDeclarationSyntax{}, false
	}
	rest := tokens[2:]
	if staticWordsAt(rest, 0, "you", "cast") {
		rest = rest[2:]
	}
	if len(rest) != 10 ||
		!staticWordsAt(rest, 0, "of", "the", "chosen", "type", "cost") ||
		rest[5].Kind != shared.Symbol ||
		!staticWordsAt(rest, 6, "less", "to", "cast") ||
		rest[9].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	amount, ok := staticGenericSymbolValue(rest[5].Text)
	if !ok || amount <= 0 {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                StaticDeclarationCostModifier,
		Span:                shared.SpanOf(tokens),
		OperationSpan:       shared.SpanOf(rest[4:9]),
		CostModifier:        StaticDeclarationCostModifierSpellReduction,
		CostReductionAmount: amount,
		SpellType:           StaticDeclarationSpellTypeCreature,
		ChosenCreatureType:  true,
	}, true
}

// parseStaticSpellColorDisjunctionCostModifier recognizes the static cast-cost
// modifier whose color filter is a disjunction expressed with "or":
//
//	"Each spell you cast that's red or green costs {N} less to cast." (Goblin Anarchomancer)
//
// The affected spells are the controller's spells that carry any one of the
// listed colors. Two or more real colors are required; a single color falls
// through to the "<color> spells you cast ..." form.
func parseStaticSpellColorDisjunctionCostModifier(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if !staticWordsAt(tokens, 0, "each", "spell", "you", "cast", "that's") {
		return StaticDeclarationSyntax{}, false
	}
	colors, next, ok := staticSpellColorDisjunction(tokens[5:])
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	tail, ok := staticSpellCostModifierTail(tokens[5+next:])
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                StaticDeclarationCostModifier,
		Span:                shared.SpanOf(tokens),
		OperationSpan:       tail.OperationSpan,
		CostModifier:        tail.Kind,
		CostReductionAmount: tail.Amount,
		CostIncreaseColors:  tail.IncreaseColors,
		CostIncreaseLife:    tail.LifeIncrease,
		SpellType:           StaticDeclarationSpellTypeAll,
		SpellColors:         colors,
	}, true
}

// parseStaticSpellColorPairCostModifier recognizes the static cast-cost modifier
// whose color filter is a disjunction expressed as two "<color> spells" phrases
// joined by "and":
//
//	"Blue spells and red spells you cast cost {N} less to cast." (Nightscape Familiar and the other Familiars)
//
// The affected spells are the controller's spells that carry either color.
func parseStaticSpellColorPairCostModifier(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 5 {
		return StaticDeclarationSyntax{}, false
	}
	first := staticSpellColorWord(tokens[0])
	if first == StaticDeclarationSpellColorNone || first == StaticDeclarationSpellColorColorless {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 1, "spells", "and") {
		return StaticDeclarationSyntax{}, false
	}
	second := staticSpellColorWord(tokens[3])
	if second == StaticDeclarationSpellColorNone || second == StaticDeclarationSpellColorColorless {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 4, "spells", "you", "cast") {
		return StaticDeclarationSyntax{}, false
	}
	tail, ok := staticSpellCostModifierTail(tokens[7:])
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                StaticDeclarationCostModifier,
		Span:                shared.SpanOf(tokens),
		OperationSpan:       tail.OperationSpan,
		CostModifier:        tail.Kind,
		CostReductionAmount: tail.Amount,
		CostIncreaseColors:  tail.IncreaseColors,
		CostIncreaseLife:    tail.LifeIncrease,
		SpellType:           StaticDeclarationSpellTypeAll,
		SpellColors:         []StaticDeclarationSpellColorKind{first, second},
	}, true
}

// parseStaticSpellColorTypePairCostModifier recognizes the static cast-cost
// modifier whose filter is a disjunction of two or more "<color> <type> spells"
// phrases that share one card type, joined by "and":
//
//	"Red creature spells and green creature spells cost {1} more to cast." (High Seas)
//	"Green enchantment spells and white enchantment spells cost {2} more to cast." (Irini Sengir)
//
// A spell matches when it carries any one of the listed colors and the shared
// card type. The caster phrase ("you cast", "your opponents cast", or absent)
// scopes the affected players. A disjunction whose elements name different card
// types, or fewer than two colors, fails closed.
func parseStaticSpellColorTypePairCostModifier(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	match, ok := staticSpellColorTypeDisjunction(tokens)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	caster, rest := staticSpellCasterPhrase(match.rest)
	tail, ok := staticSpellCostModifierTail(rest)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                StaticDeclarationCostModifier,
		Span:                shared.SpanOf(tokens),
		OperationSpan:       tail.OperationSpan,
		CostModifier:        tail.Kind,
		CostReductionAmount: tail.Amount,
		CostIncreaseColors:  tail.IncreaseColors,
		CostIncreaseLife:    tail.LifeIncrease,
		SpellType:           match.sharedType,
		SpellColors:         match.colors,
		SpellCaster:         caster,
	}, true
}

// staticColorTypeDisjunctionMatch is the parsed result of a "<color> <type>
// spells and <color> <type> spells" disjunction: the colors in source order, the
// single card type all phrases share, and the tokens remaining past the final
// "spells" noun.
type staticColorTypeDisjunctionMatch struct {
	colors     []StaticDeclarationSpellColorKind
	sharedType StaticDeclarationSpellTypeKind
	rest       []shared.Token
}

// staticSpellColorTypeDisjunction reads a run of "<color> <type> spells" phrases
// joined by "and" ("Red creature spells and green creature spells"), returning
// the colors in source order, the single card type the phrases share, and the
// remaining tokens beginning just past the final "spells" noun. Each phrase must
// open with a real color word (colorless is not admitted) followed by the same
// single card-type word; a phrase naming a different type, or a list shorter
// than two colors, fails closed by returning false.
func staticSpellColorTypeDisjunction(tokens []shared.Token) (staticColorTypeDisjunctionMatch, bool) {
	var colors []StaticDeclarationSpellColorKind
	sharedType := StaticDeclarationSpellTypeAll
	rest := tokens
	for {
		if len(rest) == 0 {
			return staticColorTypeDisjunctionMatch{}, false
		}
		color := staticSpellColorWord(rest[0])
		if color == StaticDeclarationSpellColorNone || color == StaticDeclarationSpellColorColorless {
			return staticColorTypeDisjunctionMatch{}, false
		}
		spellType, next, ok := staticSpellTypeFilter(rest[1:])
		if !ok || spellType == StaticDeclarationSpellTypeAll {
			return staticColorTypeDisjunctionMatch{}, false
		}
		if len(colors) == 0 {
			sharedType = spellType
		} else if spellType != sharedType {
			return staticColorTypeDisjunctionMatch{}, false
		}
		colors = append(colors, color)
		if !staticWordsAt(next, 0, "spells") {
			return staticColorTypeDisjunctionMatch{}, false
		}
		rest = next[1:]
		// Continue only when "and" introduces another colored phrase; a trailing
		// "and" before the cost clause or caster phrase ends the disjunction.
		if len(rest) >= 2 && equalWord(rest[0], "and") {
			nextColor := staticSpellColorWord(rest[1])
			if nextColor != StaticDeclarationSpellColorNone && nextColor != StaticDeclarationSpellColorColorless {
				rest = rest[1:]
				continue
			}
		}
		break
	}
	if len(colors) < 2 {
		return staticColorTypeDisjunctionMatch{}, false
	}
	return staticColorTypeDisjunctionMatch{colors: colors, sharedType: sharedType, rest: rest}, true
}

// staticSpellCardTypeList reads a coordinated list of two or more card-type
// words joined by "and" with optional Oxford commas ("Artifact and enchantment",
// "Artifact, instant, and sorcery"), returning the card types in source order
// and the remaining tokens beginning at the following "spells" noun. A list of
// fewer than two recognized card types, or a word that is not a card type, fails
// closed by returning false so the single-type and instant-and-sorcery filters
// keep their meaning.
func staticSpellCardTypeList(tokens []shared.Token) ([]CardType, []shared.Token, bool) {
	var list []CardType
	rest := tokens
	for {
		if len(rest) == 0 {
			return nil, nil, false
		}
		cardType, ok := recognizeCardTypeWord(rest[0].Text)
		if !ok {
			return nil, nil, false
		}
		list = append(list, cardType)
		rest = rest[1:]
		if len(rest) > 0 && rest[0].Kind == shared.Comma {
			rest = rest[1:]
			if len(rest) > 0 && equalWord(rest[0], "and") {
				rest = rest[1:]
			}
			continue
		}
		if len(rest) > 0 && equalWord(rest[0], "and") {
			rest = rest[1:]
			continue
		}
		break
	}
	if len(list) < 2 {
		return nil, nil, false
	}
	return list, rest, true
}

// staticSpellCasterPhrase consumes an optional cast-cost modifier caster phrase
// ("you cast" for the controller, "your opponents cast" for opponents), returning
// the caster filter and the remaining tokens. An absent phrase yields the
// any-player filter with the tokens unchanged.
func staticSpellCasterPhrase(tokens []shared.Token) (StaticDeclarationSpellCasterKind, []shared.Token) {
	switch {
	case staticWordsAt(tokens, 0, "you", "cast"):
		return StaticDeclarationSpellCasterController, tokens[2:]
	case staticWordsAt(tokens, 0, "your", "opponents", "cast"):
		return StaticDeclarationSpellCasterOpponents, tokens[3:]
	default:
		return StaticDeclarationSpellCasterAny, tokens
	}
}

// staticSpellColorDisjunction reads a run of color words joined by "or"
// ("red or green", "white or blue or black"), returning the colors in source
// order and the number of tokens consumed. It succeeds only for two or more
// real colors; colorless is not admitted in a disjunction.
func staticSpellColorDisjunction(tokens []shared.Token) ([]StaticDeclarationSpellColorKind, int, bool) {
	var colors []StaticDeclarationSpellColorKind
	index := 0
	for {
		if index >= len(tokens) {
			return nil, 0, false
		}
		color := staticSpellColorWord(tokens[index])
		if color == StaticDeclarationSpellColorNone || color == StaticDeclarationSpellColorColorless {
			return nil, 0, false
		}
		colors = append(colors, color)
		index++
		if index < len(tokens) && equalWord(tokens[index], "or") {
			index++
			continue
		}
		break
	}
	if len(colors) < 2 {
		return nil, 0, false
	}
	return colors, index, true
}

// staticSpellCostTail is the parsed trailing amount of a spell cast-cost
// modifier: the modifier kind, the generic amount, and the span covering
// "cost {N} less to cast".
type staticSpellCostTail struct {
	Kind           StaticDeclarationCostModifierKind
	Amount         int
	IncreaseColors []mana.Color
	LifeIncrease   int
	OperationSpan  shared.Span
}

// staticSpellCostModifierTail parses the trailing cost clause of a spell
// cast-cost modifier. It recognizes both the mana form "cost(s) {N} less/more to
// cast." and the life-tax form "cost(s) an additional N life to cast." (Terror
// of the Peaks). The cost verb is "cost" or "costs" so both the plural
// "<filter> spells ... cost" and the singular "Each spell ... costs" subjects
// fit. In the mana form the cost symbol is a generic {N}, or a single basic
// colored mana symbol ({W}, {U}, {B}, {R}, or {G}) when the operation is "more"
// ("Black spells you cast cost {B} more to cast.", Derelor): a colored symbol
// names a colored mana increase, which only makes sense as a tax, so a colored
// "less" reduction fails closed.
func staticSpellCostModifierTail(tokens []shared.Token) (staticSpellCostTail, bool) {
	if tail, ok := staticSpellCostLifeIncreaseTail(tokens); ok {
		return tail, true
	}
	if len(tokens) != 6 ||
		(!equalWord(tokens[0], "cost") && !equalWord(tokens[0], "costs")) ||
		tokens[1].Kind != shared.Symbol ||
		!staticWordsAt(tokens, 3, "to", "cast") ||
		tokens[5].Kind != shared.Period {
		return staticSpellCostTail{}, false
	}
	var kind StaticDeclarationCostModifierKind
	switch {
	case equalWord(tokens[2], "less"):
		kind = StaticDeclarationCostModifierSpellReduction
	case equalWord(tokens[2], "more"):
		kind = StaticDeclarationCostModifierSpellIncrease
	default:
		return staticSpellCostTail{}, false
	}
	span := shared.SpanOf(tokens[0:5])
	if amount, ok := staticGenericSymbolValue(tokens[1].Text); ok && amount > 0 {
		return staticSpellCostTail{Kind: kind, Amount: amount, OperationSpan: span}, true
	}
	if kind == StaticDeclarationCostModifierSpellIncrease {
		if color, ok := staticBasicColoredSymbol(tokens[1].Text); ok {
			return staticSpellCostTail{Kind: kind, IncreaseColors: []mana.Color{color}, OperationSpan: span}, true
		}
	}
	return staticSpellCostTail{}, false
}

// staticSpellCostLifeIncreaseTail parses the life-tax trailing clause "cost(s)
// an additional N life to cast." of a spell cast-cost modifier ("Spells your
// opponents cast that target this creature cost an additional 3 life to cast.",
// Terror of the Peaks). N is a positive integer count of life the caster pays as
// an additional cost. The cost verb is "cost" or "costs" so both the plural and
// singular spell subjects fit. Any other shape leaves the clause unconsumed.
func staticSpellCostLifeIncreaseTail(tokens []shared.Token) (staticSpellCostTail, bool) {
	if len(tokens) != 8 ||
		(!equalWord(tokens[0], "cost") && !equalWord(tokens[0], "costs")) ||
		!staticWordsAt(tokens, 1, "an", "additional") ||
		tokens[3].Kind != shared.Integer ||
		!staticWordsAt(tokens, 4, "life", "to", "cast") ||
		tokens[7].Kind != shared.Period {
		return staticSpellCostTail{}, false
	}
	amount, ok := conditionNumberValue(tokens[3])
	if !ok || amount <= 0 {
		return staticSpellCostTail{}, false
	}
	return staticSpellCostTail{
		Kind:          StaticDeclarationCostModifierSpellIncrease,
		LifeIncrease:  amount,
		OperationSpan: shared.SpanOf(tokens[0:7]),
	}, true
}

// parseStaticCastAsThoughFlashDeclaration recognizes the static timing
// permission "You may cast [<filter>] spells as though they had flash."
// (Vedalken Orrery, Leyline of Anticipation; Hypersonic Dragon's "sorcery
// spells"; Sigarda's Aid's "Aura and Equipment spells"). <filter> is an optional
// card-type filter ("creature", "sorcery", "instant and sorcery") or a subtype
// list ("Aura and Equipment"); an absent filter grants the permission for every
// spell. Any deviation leaves the clause unconsumed and fails closed.
func parseStaticCastAsThoughFlashDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "may", "cast") {
		return StaticDeclarationSyntax{}, false
	}
	rest := tokens[3:]
	spellType := StaticDeclarationSpellTypeAll
	var subtypes []types.Sub
	if subs, next, ok := staticSpellSubtypeFilter(rest, atoms); ok {
		subtypes = subs
		rest = next
	} else {
		var ok bool
		spellType, rest, ok = staticSpellTypeFilter(rest)
		if !ok {
			return StaticDeclarationSyntax{}, false
		}
	}
	if len(rest) != 7 ||
		!staticWordsAt(rest, 0, "spells", "as", "though", "they", "had", "flash") ||
		rest[6].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:               StaticDeclarationCastAsThoughFlash,
		Span:               shared.SpanOf(tokens),
		OperationSpan:      shared.SpanOf(rest[0:6]),
		FlashSpellType:     spellType,
		FlashSpellSubtypes: subtypes,
	}, true
}

// parseStaticSpellUncounterableDeclaration recognizes the static
// "[<filter>] spells you control can't be countered." (Rhythm of the
// Wild, Prowling Serpopard, Cavern-style grants). The optional leading filter
// constrains the affected spells to a single card type ("Creature") or a
// subtype list ("Human"); a bare "Spells you control ..." affects every spell
// the controller casts. Color filters and the instant-and-sorcery filter fail
// closed because the runtime counter check matches only the spell's card types
// and subtypes.
func parseStaticSpellUncounterableDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	spellType := StaticDeclarationSpellTypeAll
	rest := tokens
	var subtypes []types.Sub
	if subs, next, ok := staticSpellSubtypeFilter(tokens, atoms); ok {
		subtypes = subs
		rest = next
	} else {
		var ok bool
		spellType, rest, ok = staticSpellTypeFilter(tokens)
		if !ok || spellType == StaticDeclarationSpellTypeInstantOrSorcery {
			return StaticDeclarationSyntax{}, false
		}
	}
	if len(rest) != 7 ||
		!staticWordsAt(rest, 0, "spells", "you", "control") ||
		(!staticWordsAt(rest, 3, "can't") && !staticWordsAt(rest, 3, "cannot")) ||
		!staticWordsAt(rest, 4, "be", "countered") ||
		rest[6].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                       StaticDeclarationSpellUncounterable,
		Span:                       shared.SpanOf(tokens),
		OperationSpan:              shared.SpanOf(rest[3:6]),
		SpellType:                  spellType,
		UncounterableSpellSubtypes: subtypes,
	}, true
}

// parseStaticCombatDamagePreventionDeclaration recognizes the continuous static
// prevention "Prevent all combat damage that would be dealt to <group>."
// (Goldbug, Humanity's Ally; Dolmen Gate), where <group> is a controller-scoped
// permanent selection ("attacking Humans you control", "attacking creatures you
// control"). It models a forward, TO-only, continuous prevention aimed at a
// controlled group. To stay correct it fails closed for wordings this mechanic
// does not model: a trailing "this turn" (the one-shot combat-damage shield owned
// by the effect parser), a "dealt by" direction or any "by" source filter (the
// two-way and source-scoped shields, e.g. Fog Bank, Armored Transport), and
// back-reference or attached recipients that carry no controller relation ("this
// creature", "equipped creature", Guard Gomazoa, General's Kabuto). The required
// controller relation is the property that distinguishes a well-defined continuous
// group from those singular self/attached preventions; the recipient tokens are
// parsed with the shared selection parser so any permanent filter it recognizes
// works.
func parseStaticCombatDamagePreventionDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	body := tokens[:len(tokens)-1]
	prefix := []string{"prevent", "all", "combat", "damage", "that", "would", "be", "dealt", "to"}
	if len(body) <= len(prefix) {
		return StaticDeclarationSyntax{}, false
	}
	for i, want := range prefix {
		if !equalWord(body[i], want) {
			return StaticDeclarationSyntax{}, false
		}
	}
	recipientTokens := body[len(prefix):]
	if len(recipientTokens) >= 2 &&
		equalWord(recipientTokens[len(recipientTokens)-2], "this") &&
		equalWord(recipientTokens[len(recipientTokens)-1], "turn") {
		return StaticDeclarationSyntax{}, false
	}
	for _, token := range recipientTokens {
		if equalWord(token, "by") {
			return StaticDeclarationSyntax{}, false
		}
	}
	recipient := parseSelection(recipientTokens, atoms)
	if recipient.Controller == SelectionControllerAny {
		return StaticDeclarationSyntax{}, false
	}
	if conjunctiveTypeTarget(recipient) {
		recipient.ConjunctiveTypes = true
	}
	return StaticDeclarationSyntax{
		Kind:                StaticDeclarationCombatDamagePrevention,
		Span:                shared.SpanOf(tokens),
		OperationSpan:       shared.SpanOf(body[0:4]),
		PreventionRecipient: recipient,
	}, true
}

// parseStaticUntapDuringOtherUntapStepDeclaration recognizes the static "Untap
// <group> you control during each other player's untap step." (Seedborn Muse,
// Drumbellower) and the self form "Untap this <permanent> during each other
// player's untap step." (Unwinding Clock-style printings). The trailing timing
// also accepts the equivalent "during each opponent's untap step" wording. The
// group is one of a closed set of controller-scoped filters (every permanent,
// creatures, artifacts, or lands) or the source permanent itself; color,
// subtype, multi-type, and counter-filtered groups fail closed here because the
// runtime untap effect filters only by card type.
func parseStaticUntapDuringOtherUntapStepDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) == 0 ||
		tokens[len(tokens)-1].Kind != shared.Period ||
		!equalWord(tokens[0], "untap") {
		return StaticDeclarationSyntax{}, false
	}
	group, next, ok := staticUntapGroup(tokens)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, next, "during", "each") {
		return StaticDeclarationSyntax{}, false
	}
	timing := next + 2
	switch {
	case staticWordsAt(tokens, timing, "other", "player's", "untap", "step") &&
		timing+4 == len(tokens)-1:
	case staticWordsAt(tokens, timing, "opponent's", "untap", "step") &&
		timing+3 == len(tokens)-1:
	default:
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationUntapDuringOtherUntapStep,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[:1]),
		UntapGroup:    group,
	}, true
}

// staticUntapGroup strips the affected group from an untap-during-other-untap-
// step declaration and returns the closed group filter with the index of the
// first token after the group. It recognizes "this <permanent>" (the source
// itself) and "all <permanents|creatures|artifacts|lands> you control".
func staticUntapGroup(tokens []shared.Token) (StaticUntapGroupKind, int, bool) {
	if staticWordsAt(tokens, 1, "this") && len(tokens) > 2 {
		switch {
		case equalWord(tokens[2], "artifact"),
			equalWord(tokens[2], "creature"),
			equalWord(tokens[2], "permanent"),
			equalWord(tokens[2], "land"),
			equalWord(tokens[2], "enchantment"):
			return StaticUntapGroupSelf, 3, true
		default:
			return StaticUntapGroupNone, 0, false
		}
	}
	if !staticWordsAt(tokens, 1, "all") || len(tokens) < 6 ||
		!staticWordsAt(tokens, 3, "you", "control") {
		return StaticUntapGroupNone, 0, false
	}
	switch {
	case equalWord(tokens[2], "permanents"):
		return StaticUntapGroupPermanents, 5, true
	case equalWord(tokens[2], "creatures"):
		return StaticUntapGroupCreatures, 5, true
	case equalWord(tokens[2], "artifacts"):
		return StaticUntapGroupArtifacts, 5, true
	case equalWord(tokens[2], "lands"):
		return StaticUntapGroupLands, 5, true
	default:
		return StaticUntapGroupNone, 0, false
	}
}

// staticSpellSubtypeFilter strips an optional leading subtype filter from a
// "spells you cast cost ..." declaration. It accepts both the shared-noun
// conjunction, where the subtypes precede a single "spells" head noun ("Aura
// spells", "Aura and Equipment spells"), and the repeated-noun conjunction,
// where each subtype carries its own "spells" head noun joined by "and"
// ("Kithkin spells and Soldier spells", the Banneret cycle). It returns the
// recognized subtypes with the remaining tokens beginning at the final "spells"
// head noun, or false when the leading tokens are not a subtype list terminated
// by "spells". Either way every listed subtype is matched disjunctively, so a
// spell qualifies when it carries any one of them.
func staticSpellSubtypeFilter(tokens []shared.Token, atoms Atoms) ([]types.Sub, []shared.Token, bool) {
	var subtypes []types.Sub
	rest := tokens
	for {
		if len(rest) == 0 {
			return nil, nil, false
		}
		sub, ok := atoms.SubtypeAt(rest[0].Span)
		if !ok {
			return nil, nil, false
		}
		subtypes = append(subtypes, sub)
		rest = rest[1:]
		if len(rest) >= 1 && equalWord(rest[0], "spells") {
			// Repeated-noun conjunction: a following "and <Subtype> spells"
			// continues the list with another head noun ("Kithkin spells and
			// Soldier spells"). The lookahead requires the conjoined word to be
			// a recognized subtype so an unrelated trailing "and ..." clause is
			// left for the caller.
			if len(rest) >= 3 && equalWord(rest[1], "and") {
				if _, ok := atoms.SubtypeAt(rest[2].Span); ok {
					rest = rest[2:]
					continue
				}
			}
			return subtypes, rest, true
		}
		// Shared-noun conjunction: "and <Subtype>" before any head noun
		// continues the list under one trailing "spells" ("Aura and Equipment
		// spells").
		if len(rest) >= 1 && equalWord(rest[0], "and") {
			rest = rest[1:]
			continue
		}
		return nil, nil, false
	}
}

// staticSpellColorWord maps a single color word ("White", "Blue", "Black",
// "Red", "Green", or "Colorless") onto its closed color filter, returning
// StaticDeclarationSpellColorNone when the token is not a recognized color word.
func staticSpellColorWord(token shared.Token) StaticDeclarationSpellColorKind {
	switch {
	case equalWord(token, "white"):
		return StaticDeclarationSpellColorWhite
	case equalWord(token, "blue"):
		return StaticDeclarationSpellColorBlue
	case equalWord(token, "black"):
		return StaticDeclarationSpellColorBlack
	case equalWord(token, "red"):
		return StaticDeclarationSpellColorRed
	case equalWord(token, "green"):
		return StaticDeclarationSpellColorGreen
	case equalWord(token, "colorless"):
		return StaticDeclarationSpellColorColorless
	default:
		return StaticDeclarationSpellColorNone
	}
}

// staticSpellTypeFilter strips an optional leading spell-type filter from a
// "spells you cast cost ..." declaration and returns the closed filter kind with
// the remaining tokens. It returns false when a leading word is present that is
// not a recognized single-type or instant-and-sorcery filter.
func staticSpellTypeFilter(tokens []shared.Token) (StaticDeclarationSpellTypeKind, []shared.Token, bool) {
	if len(tokens) == 0 {
		return StaticDeclarationSpellTypeAll, nil, false
	}
	if equalWord(tokens[0], "spells") {
		return StaticDeclarationSpellTypeAll, tokens, true
	}
	if len(tokens) >= 4 &&
		equalWord(tokens[0], "instant") &&
		equalWord(tokens[1], "and") &&
		equalWord(tokens[2], "sorcery") &&
		equalWord(tokens[3], "spells") {
		return StaticDeclarationSpellTypeInstantOrSorcery, tokens[3:], true
	}
	if len(tokens) < 2 || !equalWord(tokens[1], "spells") {
		return StaticDeclarationSpellTypeAll, nil, false
	}
	switch {
	case equalWord(tokens[0], "artifact"):
		return StaticDeclarationSpellTypeArtifact, tokens[1:], true
	case equalWord(tokens[0], "creature"):
		return StaticDeclarationSpellTypeCreature, tokens[1:], true
	case equalWord(tokens[0], "enchantment"):
		return StaticDeclarationSpellTypeEnchantment, tokens[1:], true
	case equalWord(tokens[0], "instant"):
		return StaticDeclarationSpellTypeInstant, tokens[1:], true
	case equalWord(tokens[0], "sorcery"):
		return StaticDeclarationSpellTypeSorcery, tokens[1:], true
	default:
		return StaticDeclarationSpellTypeAll, nil, false
	}
}

// staticSpellExcludedTypeFilter strips a leading "non<type> spells" exclusion
// filter ("Noncreature spells ...", "Nonartifact spells ...") from a cast-cost
// modifier declaration, returning the excluded card types and the remaining
// tokens beginning at the "spells" noun. It recognizes a single "non"-prefixed
// card-type word immediately followed by "spells"; any other leading word fails
// closed so the bare and single-type filters keep their meaning.
func staticSpellExcludedTypeFilter(tokens []shared.Token) ([]CardType, []shared.Token, bool) {
	if len(tokens) < 2 || !equalWord(tokens[1], "spells") {
		return nil, nil, false
	}
	cardType, ok := recognizeExcludedCardTypeWord(tokens[0].Text)
	if !ok {
		return nil, nil, false
	}
	return []CardType{cardType}, tokens[1:], true
}

// parseStaticAbilityReduction recognizes "Cycling abilities you activate cost up
// to {N} less to activate." and returns the generic reduction N.
func parseStaticAbilityReduction(tokens []shared.Token, keyword Keyword) (int, bool) {
	if len(tokens) != 12 ||
		keyword.NameSpan.Start.Offset != tokens[0].Span.Start.Offset ||
		!staticWordsAt(tokens, 1, "abilities", "you", "activate", "cost", "up", "to") ||
		tokens[7].Kind != shared.Symbol ||
		!staticWordsAt(tokens, 8, "less", "to", "activate") {
		return 0, false
	}
	return staticGenericSymbolValue(tokens[7].Text)
}

// parseStaticReplaceCyclingCost recognizes "you may pay {N} rather than pay
// cycling costs." and returns the replacement cost text.
func parseStaticReplaceCyclingCost(tokens []shared.Token, keyword Keyword) (string, bool) {
	if len(tokens) != 10 ||
		!staticWordsAt(tokens, 0, "you", "may", "pay") ||
		tokens[3].Kind != shared.Symbol ||
		!staticWordsAt(tokens, 4, "rather", "than", "pay") ||
		keyword.NameSpan.Start.Offset != tokens[7].Span.Start.Offset ||
		!staticWordsAt(tokens, 8, "costs") {
		return "", false
	}
	return staticReplacementCost(tokens[3].Text)
}

// parseStaticReplaceFirstCyclingCost recognizes "You may pay {N} rather than pay
// the cycling cost of the first card you cycle each turn" and returns the
// replacement cost text.
func parseStaticReplaceFirstCyclingCost(tokens []shared.Token, keyword Keyword) (string, bool) {
	if len(tokens) != 19 ||
		!staticWordsAt(tokens, 0, "you", "may", "pay") ||
		tokens[3].Kind != shared.Symbol ||
		!staticWordsAt(tokens, 4, "rather", "than", "pay", "the") ||
		keyword.NameSpan.Start.Offset != tokens[8].Span.Start.Offset ||
		!staticWordsAt(tokens, 9, "cost", "of", "the", "first", "card", "you", "cycle", "each", "turn") {
		return "", false
	}
	return staticReplacementCost(tokens[3].Text)
}

func parseStaticCardAbilityGrantDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 9 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "each") {
		return StaticDeclarationSyntax{}, false
	}
	filter := staticHandCardFilter(tokens[1])
	if filter == StaticDeclarationCardFilterNone {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 2, "card", "in", "your", "hand", "has") {
		return StaticDeclarationSyntax{}, false
	}
	keyword, width, ok := staticKeywordAt(tokens, 7, len(tokens)-1, atoms)
	if !ok || keyword.Kind != KeywordCycling ||
		keyword.Parameter.Kind != KeywordParameterManaCost || 7+width != len(tokens)-1 {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationCardAbilityGrant,
		Span:          shared.SpanOf(tokens),
		OperationSpan: keyword.Span,
		Subject: StaticDeclarationSubject{
			Kind:       StaticDeclarationSubjectControllerHand,
			Span:       shared.SpanOf(tokens[:6]),
			CardFilter: filter,
		},
		KeywordSpans: []shared.Span{keyword.Span},
	}, true
}

func staticHandCardFilter(token shared.Token) StaticDeclarationCardFilterKind {
	switch {
	case equalWord(token, "land"):
		return StaticDeclarationCardFilterLand
	case equalWord(token, "creature"):
		return StaticDeclarationCardFilterCreature
	case equalWord(token, "historic"):
		return StaticDeclarationCardFilterHistoric
	default:
		return StaticDeclarationCardFilterNone
	}
}

// staticSoleBareCyclingKeyword returns the single cycling keyword atom in the
// body when it is the only keyword and carries no parameter.
func staticSoleBareCyclingKeyword(tokens []shared.Token, atoms Atoms) (Keyword, bool) {
	keywords := atoms.KeywordsWithin(tokens)
	if len(keywords) != 1 ||
		keywords[0].Kind != KeywordCycling ||
		keywords[0].Parameter.Kind != KeywordParameterNone {
		return Keyword{}, false
	}
	return keywords[0], true
}

// staticGenericSymbolValue returns the generic value of a single {N} symbol.
func staticGenericSymbolValue(text string) (int, bool) {
	symbol, ok := staticTrimSymbol(text)
	if !ok || symbol == "" || (len(symbol) > 1 && symbol[0] == '0') {
		return 0, false
	}
	for i := range symbol {
		if symbol[i] < '0' || symbol[i] > '9' {
			return 0, false
		}
	}
	value, err := strconv.Atoi(symbol)
	if err != nil {
		return 0, false
	}
	return value, true
}

// staticReplacementCost returns the canonical mana cost text for a single {N}
// generic symbol, where {0} renders as the empty string.
func staticReplacementCost(text string) (string, bool) {
	value, ok := staticGenericSymbolValue(text)
	if !ok {
		return "", false
	}
	if value == 0 {
		return "", true
	}
	return text, true
}

func staticTrimSymbol(text string) (string, bool) {
	symbol, ok := strings.CutPrefix(text, "{")
	if !ok {
		return "", false
	}
	return strings.CutSuffix(symbol, "}")
}

// staticBasicColoredSymbol returns the basic colored mana symbol named by a
// single mana-symbol token ("{B}" -> mana.B). It accepts only the five basic
// colored symbols W, U, B, R, and G; generic, colorless, hybrid, and every
// other symbol shape fail closed by returning false.
func staticBasicColoredSymbol(text string) (mana.Color, bool) {
	symbol, ok := staticTrimSymbol(text)
	if !ok {
		return "", false
	}
	switch symbol {
	case "W":
		return mana.W, true
	case "U":
		return mana.U, true
	case "B":
		return mana.B, true
	case "R":
		return mana.R, true
	case "G":
		return mana.G, true
	default:
		return "", false
	}
}
