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
)

// DynamicAmount describes an effect amount determined as the effect resolves
// (CR 608.2c), separate from characteristic-defining P/T values in layers.
type DynamicAmount struct {
	Kind DynamicAmountKind

	Constant   int
	Multiplier int

	CounterKind counter.Kind
	Group       GroupReference
	Object      ObjectReference
	Player      *PlayerReference
	CardZone    zone.Type
	Selection   *Selection
	ResultKey   ResultKey
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
}
