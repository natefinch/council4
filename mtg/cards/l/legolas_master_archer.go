package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LegolasMasterArcher is the card definition for Legolas, Master Archer.
//
// Type: Legendary Creature — Elf Archer
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Reach
//	Whenever you cast a spell that targets Legolas, put a +1/+1 counter on Legolas.
//	Whenever you cast a spell that targets a creature you don't control, Legolas deals damage equal to its power to up to one target creature.
var LegolasMasterArcher = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Legolas, Master Archer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Archer},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			OracleText: `
				Reach
				Whenever you cast a spell that targets Legolas, put a +1/+1 counter on Legolas.
				Whenever you cast a spell that targets a creature you don't control, Legolas deals damage equal to its power to up to one target creature.
			`,
		},
	}

	card.StaticAbilities = append(card.StaticAbilities,
		game.ReachStaticBody,
	)

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbilityBody{
			Text: `
				Whenever you cast a spell that targets Legolas, put a +1/+1 counter on Legolas.
			`,
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:              game.EventSpellCast,
					Controller:         game.TriggerControllerYou,
					SpellTargetsSource: true,
				},
			},
			Content: game.PlainAbilityContent{
				Sequence: []game.Effect{
					{
						Type:        game.EffectAddCounter,
						Amount:      1,
						CounterKind: counter.PlusOnePlusOne,
						TargetIndex: game.TargetIndexSourcePermanent,
					},
				},
			},
		},
	)

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbilityBody{
			Text: `
				Whenever you cast a spell that targets a creature you don't control, Legolas deals damage equal to its power to up to one target creature.
			`,
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:            game.EventSpellCast,
					Controller:       game.TriggerControllerYou,
					SpellTargetAllow: game.TargetAllowPermanent,
					SpellTargetPattern: opt.Val(game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerNotYou,
					}),
				},
			},
			Optional: true,
			Content: game.PlainAbilityContent{
				Targets: []game.TargetSpec{
					{
						MinTargets: 0,
						MaxTargets: 1,
						Constraint: "creature",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{types.Creature},
						},
					},
				},
				Sequence: []game.Effect{
					{
						Type:        game.EffectDamage,
						TargetIndex: 0,
						DynamicAmount: opt.Val(game.DynamicAmount{
							Kind:   game.DynamicAmountObjectPower,
							Object: game.ObjectReference{Kind: game.ObjectReferenceSourcePermanent},
						}),
					},
				},
			},
		},
	)
	return card
}()
