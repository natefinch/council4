package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KRrikSonOfYawgmoth is the card definition for K'rrik, Son of Yawgmoth.
//
// Type: Legendary Creature — Phyrexian Horror Minion
// Cost: {4}{B/P}{B/P}{B/P}
//
// Oracle text:
//
//	({B/P} can be paid with either {B} or 2 life.)
//	Lifelink
//	For each {B} in a cost, you may pay 2 life rather than pay that mana.
//	Whenever you cast a black spell, put a +1/+1 counter on K'rrik.
var KRrikSonOfYawgmoth = newKRrikSonOfYawgmoth

func newKRrikSonOfYawgmoth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "K'rrik, Son of Yawgmoth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.PhyrexianMana(mana.B),
				cost.PhyrexianMana(mana.B),
				cost.PhyrexianMana(mana.B),
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Phyrexian, types.Horror, types.Minion},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "For each {B} in a cost, you may pay 2 life rather than pay that mana.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectPayLifeForColoredMana,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							ManaColor:      mana.B,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a black spell, put a +1/+1 counter on K'rrik.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{ColorsAny: []color.Color{color.Black}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
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
			({B/P} can be paid with either {B} or 2 life.)
			Lifelink
			For each {B} in a cost, you may pay 2 life rather than pay that mana.
			Whenever you cast a black spell, put a +1/+1 counter on K'rrik.
		`,
		},
	}
}
