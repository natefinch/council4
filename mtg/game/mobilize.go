package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MobilizeDynamicKind identifies a rules-derived Mobilize count (CR 702.169).
// The zero value is the fixed count, in which case MobilizeAmount.Fixed is the
// printed N; a non-zero value names a count evaluated as the attack trigger
// resolves.
type MobilizeDynamicKind int

// Mobilize dynamic count kinds.
const (
	// MobilizeDynamicNone is a fixed printed count ("Mobilize 2").
	MobilizeDynamicNone MobilizeDynamicKind = iota
	// MobilizeDynamicCreatureCardsInGraveyard is the number of creature cards in
	// the ability controller's graveyard ("Mobilize X, where X is the number of
	// creature cards in your graveyard", Avenger of the Fallen).
	MobilizeDynamicCreatureCardsInGraveyard
)

// MobilizeAmount is the typed count of tokens a Mobilize keyword creates. It is
// a fixed N when Dynamic is MobilizeDynamicNone, otherwise a rules-derived count
// named by Dynamic (Fixed is then ignored). Both fixed and dynamic forms compose
// through the single Quantity the body's CreateToken uses.
type MobilizeAmount struct {
	Fixed   int
	Dynamic MobilizeDynamicKind
}

// Quantity returns the token count as the shared effect Quantity, so the fixed
// and dynamic forms feed the same CreateToken.Amount.
func (a MobilizeAmount) Quantity() Quantity {
	switch a.Dynamic {
	case MobilizeDynamicCreatureCardsInGraveyard:
		controller := ControllerReference()
		return Dynamic(DynamicAmount{
			Kind:       DynamicAmountCountCardsInZone,
			Multiplier: 1,
			Player:     &controller,
			CardZone:   zone.Graveyard,
			Selection:  &Selection{RequiredTypes: []types.Card{types.Creature}},
		})
	default:
		return Fixed(a.Fixed)
	}
}

// mobilizeWarriorToken is the canonical 1/1 red Warrior creature token created by
// the Mobilize keyword (CR 702.169).
var mobilizeWarriorToken = &CardDef{
	CardFace: CardFace{
		Name:      "Warrior",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Warrior},
		Colors:    []color.Color{color.Red},
		Power:     opt.Val(PT{Value: 1}),
		Toughness: opt.Val(PT{Value: 1}),
	},
}

// mobilizeLinkKey links each Mobilize resolution's created tokens to the delayed
// next-end-step trigger that sacrifices exactly that set, so unrelated Warriors
// and tokens that leave and re-enter (new objects) survive, and each Mobilize
// trigger's tokens are disposed independently.
const mobilizeLinkKey = LinkedKey("mobilize-tokens")

// MobilizeTriggeredBody is the canonical triggered ability for the Mobilize
// keyword (CR 702.169): "Whenever this creature attacks, create N tapped and
// attacking 1/1 red Warrior creature tokens. Sacrifice them at the beginning of
// the next end step." The attack trigger creates the tokens under the source's
// controller attacking the same player or planeswalker as the source
// (AttackSameAsSource), publishes them under mobilizeLinkKey, and schedules a
// next-end-step delayed trigger that sacrifices that captured set. amount carries
// the fixed or dynamic token count. The ability carries the Mobilize keyword so
// HasKeyword(Mobilize) reports true.
func MobilizeTriggeredBody(amount MobilizeAmount) TriggeredAbility {
	return TriggeredAbility{
		Text:             "Mobilize",
		KeywordAbilities: []KeywordAbility{MobilizeKeyword{Amount: amount}},
		Trigger: TriggerCondition{
			Type: TriggerWhenever,
			Pattern: TriggerPattern{
				Event:  EventAttackerDeclared,
				Source: TriggerSourceSelf,
			},
		},
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: CreateToken{
					Amount:             amount.Quantity(),
					Source:             TokenDef(mobilizeWarriorToken),
					Recipient:          opt.Val(ControllerReference()),
					EntryTapped:        true,
					AttackSameAsSource: true,
					PublishLinked:      mobilizeLinkKey,
				},
			},
			{
				Primitive: CreateDelayedTrigger{
					Trigger: DelayedTriggerDef{
						Timing:              DelayedAtBeginningOfNextEndStep,
						CapturedObjectGroup: opt.Val(LinkedObjectReference(string(mobilizeLinkKey))),
						Content: Mode{
							Sequence: []Instruction{
								{
									Primitive: Sacrifice{
										Group: CapturedObjectsGroup(),
									},
								},
							},
						}.Ability(),
					},
				},
			},
		}}.Ability(),
	}
}
