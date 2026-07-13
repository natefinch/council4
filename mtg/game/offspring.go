package game

import (
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

// OffspringStaticAbility builds the static ability that carries the Offspring
// keyword (CR 702.171, Bloomburrow) with its fixed additional mana cost. While
// the card is a spell on the stack the rules layer reads this keyword to offer
// paying the offspring cost in addition to the spell's other costs; paying it
// records the resulting permanent as offspring-paid. The payoff lives on the
// separate canonical enter trigger built by OffspringEnterTriggeredAbility.
func OffspringStaticAbility(offspringCost cost.Mana) StaticAbility {
	return StaticAbility{
		Text: "Offspring " + offspringCost.String(),
		KeywordAbilities: []KeywordAbility{
			OffspringKeyword{Cost: append(cost.Mana(nil), offspringCost...)},
		},
	}
}

// OffspringEnterTriggeredAbility is the canonical enters-the-battlefield
// triggered ability for the Offspring keyword (CR 702.171b): "When this creature
// enters, if the offspring cost was paid, create a 1/1 token copy of it." The
// intervening-if reads the entering permanent event's captured offspring-paid
// state, so the token is created only when the spell was cast with its offspring
// cost paid. The token is a copy of the entering permanent, except its base
// power and toughness are each 1, under the entering permanent's controller. A
// token copy was not itself cast with offspring paid, so its own copy of this
// trigger creates no further token (CR 707.10). Treat this value as immutable.
func OffspringEnterTriggeredAbility() TriggeredAbility {
	return TriggeredAbility{
		Text:             "Offspring",
		KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Offspring}},
		Trigger: TriggerCondition{
			Type: TriggerWhen,
			Pattern: TriggerPattern{
				Event:  EventPermanentEnteredBattlefield,
				Source: TriggerSourceSelf,
			},
			InterveningIfEventPermanentWasOffspring: true,
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: CreateToken{
				Amount: Fixed(1),
				Source: TokenCopyOf(TokenCopySpec{
					Source:       TokenCopySourceObject,
					Object:       SourcePermanentReference(),
					SetPower:     opt.Val(PT{Value: 1}),
					SetToughness: opt.Val(PT{Value: 1}),
				}),
			},
		}}}.Ability(),
	}
}
