package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

// DynamicAmountKind identifies a rules-derived integer for effect resolution.
// Variable values such as X and "equal to" quantities are determined as the
// resolving instruction applies unless the card text says otherwise
// (CR 107.3, CR 608.2c).
type DynamicAmountKind int

// Dynamic amount kind values identify supported effect-resolution quantities.
const (
	DynamicAmountNone DynamicAmountKind = iota
	DynamicAmountConstant
	DynamicAmountX
	DynamicAmountTargetPower
	DynamicAmountTargetToughness
	DynamicAmountTargetManaValue
	DynamicAmountTargetCounters
	DynamicAmountControllerLife
	DynamicAmountControllerHandSize
	DynamicAmountControllerGraveyardSize
	DynamicAmountControllerBasicLandTypeCount
	DynamicAmountCountSelector
	DynamicAmountCountCardsInZone
	DynamicAmountPreviousEffectResult
	DynamicAmountOpponentCount
	DynamicAmountEventDamage
	DynamicAmountPreviousEffectExcessDamage
	DynamicAmountObjectPower
	// DynamicAmountEventCardCount is the number of cards drawn or discarded in
	// the triggering event batch (CR 122, CR 700.4). It scales draw and discard
	// triggers such as "for each card discarded this way" by reading the
	// simultaneous batch recorded on the resolving ability's TriggerEvent.
	DynamicAmountEventCardCount
	// DynamicAmountObjectToughness is the toughness of a referenced object,
	// read from the live permanent or, once it has left the battlefield, from
	// its last-known snapshot (CR 608.2h). It backs "gain/lose life equal to
	// its toughness" riders whose subject was exiled, destroyed, or bounced by
	// an earlier effect in the same resolution.
	DynamicAmountObjectToughness
	// DynamicAmountObjectManaValue is the mana value of a referenced object,
	// read from the live permanent's printed mana cost or, once it has left the
	// battlefield, from its last-known snapshot (CR 202.3, CR 608.2h). Unlike
	// DynamicAmountTargetManaValue, which reads only the live target permanent,
	// this kind falls back to last-known information so a "gain/lose life equal
	// to its mana value" rider reads the same object after a zone move. It backs
	// destroy-then-life staples (Feed the Swarm, Divine Offering) and linked
	// graveyard-return riders (Reanimate). Added last so existing kinds keep
	// their wire values.
	DynamicAmountObjectManaValue
	// DynamicAmountChosenNumber reads a prior ResolutionChoiceNumber result
	// published under ResultKey.
	DynamicAmountChosenNumber
	// DynamicAmountObjectCounters is the number of counters of CounterKind on a
	// referenced object, using last-known information after it leaves play.
	DynamicAmountObjectCounters
	// DynamicAmountCapturedTargetManaValue reads the mana value captured for an
	// enclosing effect's stack-object target when a delayed trigger was created.
	DynamicAmountCapturedTargetManaValue
	// DynamicAmountGreatestPowerInGroup is the greatest power among the
	// permanents of Group, evaluated as the effect resolves (CR 608.2c). It
	// backs "equal to the greatest power among <group>" amounts (Garruk, Primal
	// Hunter; Fungal Sprouting); an empty group yields zero. Added last so
	// existing kinds keep their wire values.
	DynamicAmountGreatestPowerInGroup
	// DynamicAmountGreatestToughnessInGroup is the greatest toughness among the
	// permanents of Group, the toughness sibling of
	// DynamicAmountGreatestPowerInGroup. An empty group yields zero.
	DynamicAmountGreatestToughnessInGroup
	// DynamicAmountGreatestManaValueInGroup is the greatest mana value among the
	// permanents of Group, the mana-value sibling of
	// DynamicAmountGreatestPowerInGroup. An empty group yields zero.
	DynamicAmountGreatestManaValueInGroup
	// DynamicAmountDevotion is the controller's devotion to Colors, the number
	// of mana symbols of those colors among the mana costs of permanents the
	// controller controls (CR 700.5). A hybrid or Phyrexian symbol counts toward
	// every color it contains, and a symbol counts once for a multi-color
	// devotion when it matches any listed color. Added last so existing kinds
	// keep their wire values.
	DynamicAmountDevotion
	// DynamicAmountSpellsCastThisTurn is the number of spells the controller has
	// cast this turn, counted from the turn's spell-cast events (CR 608.2c). It
	// backs the storm-counter family such as Aetherflux Reservoir's "you gain 1
	// life for each spell you've cast this turn." Added last so existing kinds
	// keep their wire values.
	DynamicAmountSpellsCastThisTurn
	// DynamicAmountEventLifeChange is the amount of life gained or lost by the
	// event that triggered the resolving ability ("that much life" in "Whenever
	// you gain life, target opponent loses that much life."). It reads the
	// triggering event's life quantity from the resolving ability's
	// TriggerEvent, backing the life-drain mirror family (Sanguine Bond,
	// Exquisite Blood). Added last so existing kinds keep their wire values.
	DynamicAmountEventLifeChange
	// DynamicAmountTotalPowerInGroup is the sum of power among the permanents of
	// Group, evaluated as the effect resolves (CR 608.2c). It backs "the total
	// power of <group>" amounts (Ghalta, Primal Hunger's cost reduction); an
	// empty group yields zero. DynamicAmountTotalToughnessInGroup is the
	// toughness sibling. Added last so existing kinds keep their wire values.
	DynamicAmountTotalPowerInGroup
	DynamicAmountTotalToughnessInGroup
	// DynamicAmountColorCountInGroup is the number of distinct colors among the
	// permanents of Group, evaluated as the effect resolves (CR 608.2c). A
	// permanent contributes each of its colors (CR 105.2, CR 202.2); colorless
	// permanents contribute none. It backs "the number of colors among <group>"
	// amounts such as Faeburrow Elder's "+1/+1 for each color among permanents
	// you control"; an empty or fully colorless group yields zero. Added last so
	// existing kinds keep their wire values.
	DynamicAmountColorCountInGroup
	// DynamicAmountSharedCreatureTypeCountInGroup is the number of permanents of
	// Group, other than the permanent the amount is evaluated for, that share at
	// least one creature type with it (CR 700.4, CR 608.2c). Unlike every other
	// group amount it yields a different value per affected permanent, so it is
	// evaluated against the permanent a continuous power/toughness modification
	// applies to rather than the effect's source. A Changeling has every creature
	// type, so it shares with any other creature that has at least one creature
	// type. It backs the shared-creature-type anthem family (Coat of Arms: "Each
	// creature gets +1/+1 for each other creature on the battlefield that shares a
	// creature type with it"). Added last so existing kinds keep their wire
	// values.
	DynamicAmountSharedCreatureTypeCountInGroup
	// DynamicAmountSourceCardPower is the power of the resolving ability's source
	// card, read from the card instance in whatever zone it occupies (CR 702.94).
	// It backs Scavenge, whose exile-from-graveyard cost moves the source card to
	// exile before the ability resolves, so the count reads the card's printed
	// power directly rather than a battlefield permanent. Added last so existing
	// kinds keep their wire values.
	DynamicAmountSourceCardPower
	// DynamicAmountBlockingCreaturesBeyondFirst is the number of creatures
	// blocking the resolving ability's source permanent beyond the first, read
	// from the current combat's block declarations as the ability resolves
	// (CR 509.1, CR 702.23). It is zero when the source is blocked by one or no
	// creatures, or outside combat. It backs Rampage N, whose Multiplier scales
	// the count by the printed N. Added last so existing kinds keep their wire
	// values.
	DynamicAmountBlockingCreaturesBeyondFirst
	// DynamicAmountLifeLostThisTurn is the total life Player has lost so far this
	// turn, summed from the turn's EventLifeLost amounts (CR 608.2c). Damage to
	// the player counts, because dealing damage to a player causes that player to
	// lose that much life (CR 120.3), which the rules emit as a life-loss event.
	// It backs Children of Korlis's "gain life equal to the life you've lost this
	// turn." DynamicAmountLifeGainedThisTurn is the life-gained sibling, summed
	// from the turn's EventLifeGained amounts. Both read the resolving ability's
	// controller. Added last so existing kinds keep their wire values.
	DynamicAmountLifeLostThisTurn
	DynamicAmountLifeGainedThisTurn
	// DynamicAmountMaxOf is the greatest value among Operands, each evaluated as
	// the effect resolves (CR 608.2c). It backs the "whichever is greater"
	// combinator that picks the larger of two rules-derived amounts ("equal to
	// the amount of life you gained this turn or the amount of life you lost
	// this turn, whichever is greater." — Willowdusk, Essence Seer). Each
	// operand is itself a DynamicAmount, so the combinator composes any two
	// supported amounts; an empty operand list yields zero. Added last so
	// existing kinds keep their wire values.
	DynamicAmountMaxOf
	// DynamicAmountEventCounterCount is the number of counters added in the
	// triggering EventCountersAdded batch (CR 122, CR 700.4). It scales
	// counter-placement triggers such as "draw that many cards" by reading the
	// counter count recorded on the resolving ability's TriggerEvent.
	// Added last so existing kinds keep their wire values.
	DynamicAmountEventCounterCount
	// DynamicAmountColorsOfManaSpentToCast is the number of distinct colors of
	// mana spent to cast the spell that became the resolving ability's source
	// (CR 202.2, CR 702.76). It backs the Converge count ("for each color of
	// mana spent to cast it"), most commonly the enters-with-counters quantity
	// on a creature such as Crystalline Crawler. The count is recorded as the
	// spell's costs are paid and read from the evaluating stack object, which
	// carries it forward to the entering permanent's replacement; colorless mana
	// contributes no color, so a spell paid entirely with colorless or
	// generic-from-colorless mana yields zero. Added last so existing kinds keep
	// their wire values.
	DynamicAmountColorsOfManaSpentToCast
	// DynamicAmountTotalManaValueInGroup is the sum of mana value among the
	// permanents of Group, evaluated as the effect resolves (CR 608.2c). It backs
	// "the total mana value of <group>" amounts (Metalwork Colossus, Earthquake
	// Dragon, Excalibur, Sword of Eden cost reductions); an empty group yields
	// zero. Added last so existing kinds keep their wire values.
	DynamicAmountTotalManaValueInGroup
	// DynamicAmountTimesKicked is the number of times the spell that became the
	// resolving object was kicked (its Multikicker count, CR 702.32). It is read
	// from the evaluating stack object's KickerCount, which the casting machinery
	// records as the additional cost is paid and which the entering permanent's
	// replacement carries forward. It backs "for each time it was kicked" amounts:
	// the enters-with-counters quantity on Multikicker permanents (Everflowing
	// Chalice, Gnarlid Pack) and the count payoff on Multikicker spells (Wolfbriar
	// Elemental's Wolf tokens). A spell cast without Multikicker, or a permanent
	// that did not enter from a kicked cast, yields zero. Added last so existing
	// kinds keep their wire values.
	DynamicAmountTimesKicked
	// DynamicAmountOpponentsAttackedThisCombat is the number of the resolving
	// ability controller's opponents being attacked this combat by creatures the
	// controller controls, read from the current combat's attack declarations as
	// the ability resolves (CR 506.2, CR 702.72). Each distinct opponent counts
	// once; it is zero outside combat. It backs Melee N, whose attack trigger
	// pumps the source "+1/+1 until end of turn for each opponent you attacked
	// this combat." Added last so existing kinds keep their wire values.
	DynamicAmountOpponentsAttackedThisCombat
	// DynamicAmountControllerSpeed is the resolving ability controller's current
	// speed (the Start your engines! subsystem, CR 702.179), read from the
	// player's speed as the effect resolves. A player with no speed has a value
	// of zero, and speed is capped at 4. It backs "your speed" amounts such as
	// The Speed Demon's "you draw X cards and lose X life, where X is your
	// speed." Added last so existing kinds keep their wire values.
	DynamicAmountControllerSpeed
	// DynamicAmountOpponentControllingCount is the number of the resolving ability
	// controller's opponents who control at least one permanent matching Group
	// ("the number of opponents who control a creature with power 4 or greater",
	// Summon: Yojimbo chapter IV). Group's selection is evaluated relative to each
	// opponent (its controller relation is "you control" read from that
	// opponent's perspective); each qualifying opponent counts once. It is a
	// player count, not a board count. Added last so existing kinds keep their
	// wire values.
	DynamicAmountOpponentControllingCount
	// DynamicAmountCardsDrawnThisTurn is the number of cards Player has drawn so
	// far this turn, counted from the turn's EventCardDrawn events for that
	// player (CR 608.2c). The triggering or just-resolved draw counts, because
	// its draw event precedes the resolving ability. It reads the resolving
	// ability's controller and backs the draw-payoff family: Thundering Djinn's
	// attack-trigger damage "equal to the number of cards you've drawn this turn"
	// and the characteristic-defining power sibling DynamicValueControllerCardsDrawnThisTurn.
	// Added last so existing kinds keep their wire values.
	DynamicAmountCardsDrawnThisTurn
	// DynamicAmountCardsNamedSourceInGraveyards is the number of cards in every
	// player's graveyard whose name matches the resolving ability's source card
	// name ("for each card named Rite of Flame in each graveyard", Rite of
	// Flame). Every graveyard is counted, not only the controller's; a card
	// counts when its name equals the source's name as the effect resolves
	// (CR 201.2, CR 608.2c). It reads the source name from the resolving stack
	// object and backs the self-named graveyard-count ritual payoff. Added last
	// so existing kinds keep their wire values.
	DynamicAmountCardsNamedSourceInGraveyards
	// DynamicAmountCardsNamedSourceInControllerGraveyard is the number of cards
	// in the resolving ability's controller's graveyard whose name matches the
	// resolving ability's source card name ("for each card named <this card> in
	// your graveyard", Compound Fracture, Growth Cycle). Only the controller's
	// graveyard is counted, unlike DynamicAmountCardsNamedSourceInGraveyards,
	// which counts every graveyard; a card counts when its name equals the
	// source's name as the effect resolves (CR 201.2, CR 608.2c). It reads the
	// source name from the resolving stack object. Added last so existing kinds
	// keep their wire values.
	DynamicAmountCardsNamedSourceInControllerGraveyard
	// DynamicAmountCommanderCastCount is the number of times the resolving
	// ability controller has cast their commander from the command zone this game
	// ("for each time you've cast your commander from the command zone this
	// game"), read from the controller's CommanderCastCount as the effect
	// resolves (CR 903.8, CR 608.2c). It backs the command-zone-cast anthem family
	// (Commander's Insignia; Vanguard of the Restless). It is controller-scoped
	// and zero for a player with no commander. Added last so existing kinds keep
	// their wire values.
	DynamicAmountCommanderCastCount
	// DynamicAmountBlockingCreatures is the number of creatures blocking the
	// permanent named by Object, read from the current combat's block
	// declarations as the effect resolves (CR 509.1, CR 608.2c). Unlike
	// DynamicAmountBlockingCreaturesBeyondFirst, which reads the resolving
	// ability's source and drops the first blocker for Rampage, this counts every
	// blocker of the pumped permanent (Object), so it is correct whether the
	// pumped creature is the source itself ("Whenever this creature becomes
	// blocked, it gets +2/+2 … for each creature blocking it" — Rabid Elephant)
	// or the triggering permanent ("Whenever a Beast becomes blocked, it gets …"
	// — Berserk Murlodont). It is zero when combat is not active or the permanent
	// is unblocked. Added last so existing kinds keep their wire values.
	DynamicAmountBlockingCreatures
	// DynamicAmountPlayerLife is the current life total of the player named by
	// Player, read as the effect resolves (CR 608.2c). Divisor and RoundUp halve
	// it for the "loses half their life, rounded up" family (Quietus Spike, Virtus
	// the Veiled, Scytheclaw): lowering sets Player to the losing player, Divisor
	// to 2, and RoundUp per the printed rounding. It backs a player-life amount,
	// unlike the controller-only DynamicAmountControllerLife. Added last so
	// existing kinds keep their wire values.
	DynamicAmountPlayerLife
	// DynamicAmountSpellTargetCount is the number of targets of the triggering
	// spell-cast event's spell that match Selection, counted as the ability
	// resolves. It backs the "that many" anaphor of a "Whenever you cast a spell
	// that targets one or more <selection>, put that many +1/+1 counters on
	// <self>" trigger (Arcee, Acrobatic Coupe): Selection is the same permanent
	// pattern the trigger's SpellTargetPattern matched, so the count is the
	// number of the spell's targets the trigger keyed on. It is zero outside a
	// spell-cast trigger or when the cast spell has left the stack. Added last so
	// existing kinds keep their wire values.
	DynamicAmountSpellTargetCount
	// DynamicAmountPartySize is the controller's maximum filled party roles
	// among Cleric, Rogue, Warrior, and Wizard creatures they control, with each
	// creature filling at most one role (CR 700.8).
	DynamicAmountPartySize
)

// DynamicAmount describes an effect amount determined as the effect resolves
// (CR 608.2c), separate from characteristic-defining P/T values in layers.
type DynamicAmount struct {
	Kind DynamicAmountKind

	Constant   int
	Multiplier int
	// Addend is a fixed integer added to the amount after the multiplier is
	// applied, so the value is amount*Multiplier + Addend. It backs the "plus N"
	// rider on a counted amount ("the number of cards in your hand plus one.",
	// Sea Gate Restoration). It is zero for amounts with no such rider.
	Addend int

	CounterKind counter.Kind
	Group       GroupReference
	Object      ObjectReference
	Player      *PlayerReference
	CardZone    zone.Type
	Selection   *Selection
	ResultKey   ResultKey
	// Divisor, when greater than one, divides the amount after the multiplier and
	// addend are applied, rounding down unless RoundUp is set. It backs the "half
	// their library, rounded up/down" mill amounts (Traumatize, Fleet Swallower),
	// where the counted library size is halved as the effect resolves (CR 107.4).
	// A Divisor of zero or one leaves the value unchanged.
	Divisor int
	// RoundUp rounds a Divisor division up instead of down ("rounded up" versus
	// "rounded down"). It is meaningful only alongside a Divisor greater than one
	// and is ignored otherwise.
	RoundUp bool
	// Colors lists the colors counted by a DynamicAmountDevotion amount; empty
	// for every other kind.
	Colors []color.Color
	// ColorFrom, when set on a DynamicAmountDevotion amount, names a published
	// ResolutionChoiceMana result whose chosen color is the single devotion
	// color, overriding Colors. It backs "Add an amount of mana of that color
	// equal to your devotion to that color." (Nykthos, Shrine to Nyx), where the
	// devotion color is the color chosen as the ability resolves rather than a
	// fixed color printed in the amount.
	ColorFrom ChoiceKey
	// Operands lists the sub-amounts of a DynamicAmountMaxOf combinator, each
	// evaluated independently; the amount's value is the greatest among them. It
	// is empty for every other kind.
	Operands []DynamicAmount
}
