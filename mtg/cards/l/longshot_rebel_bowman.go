package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LongshotRebelBowman is the card definition for Longshot, Rebel Bowman.
//
// Type: Legendary Creature — Human Rebel Ally
// Cost: {3}{R}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Noncreature spells you cast cost {1} less to cast.
//	Whenever you cast a noncreature spell, Longshot deals 2 damage to each opponent.
var LongshotRebelBowman = newLongshotRebelBowman()

func newLongshotRebelBowman() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Longshot, Rebel Bowman",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Rebel, types.Ally},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{ExcludedTypes: []types.Card{types.Creature}},
								GenericReduction: 1,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(2),
									Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Noncreature spells you cast cost {1} less to cast.
			Whenever you cast a noncreature spell, Longshot deals 2 damage to each opponent.
		`,
		},
	}
}
