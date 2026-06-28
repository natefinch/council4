package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ElusiveOtter is the card definition for Elusive Otter.
//
// Type: Creature — Otter // Sorcery — Adventure
// Cost: {U} // {X}{G}
// Face: Grove's Bounty — Sorcery — Adventure ({X}{G})
//
// Oracle text:
//
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	Creatures with power less than this creature's power can't block it.
var ElusiveOtter = newElusiveOtter()

func newElusiveOtter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Elusive Otter",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Otter},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ProwessStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectCantBlock,
							PermanentTypes:    []types.Card{types.Creature},
							AffectedSelection: game.Selection{PowerLessThanSource: true},
							BlockedSource:     true,
						},
					},
				},
			},
			OracleText: `
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
			Creatures with power less than this creature's power can't block it.
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Grove's Bounty",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 99,
						Constraint: "any number of target creatures you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Object:      game.AllTargetPermanentsReference(0),
							CounterKind: counter.PlusOnePlusOne,
							Distribute:  true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Distribute X +1/+1 counters among any number of target creatures you control. (Then exile this card. You may cast the creature later from exile.)
		`,
		}),
	}
}
