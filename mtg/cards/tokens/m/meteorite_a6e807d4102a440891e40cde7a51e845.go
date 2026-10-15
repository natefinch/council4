package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Meteorite is the card definition for Meteorite.
//
// Type: Token Artifact
//
// Oracle text:
//   When Meteorite enters the battlefield, it deals 2 damage to any target.
//   {T}: Add one mana of any color.

// MeteoriteTokena6e807d4102a440891e40cde7a51e845 is the card definition for Meteorite.
var MeteoriteTokena6e807d4102a440891e40cde7a51e845 = newMeteoriteTokena6e807d4102a440891e40cde7a51e845()

func newMeteoriteTokena6e807d4102a440891e40cde7a51e845() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Meteorite",
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Meteorite enters the battlefield, it deals 2 damage to any target.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "any target",
										Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(2),
											Recipient:    game.AnyTargetDamageRecipient(0),
											DamageSource: opt.Val(game.EventPermanentReference()),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			When Meteorite enters the battlefield, it deals 2 damage to any target.
			{T}: Add one mana of any color.
		`,
		},
	}
}
