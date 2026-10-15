package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Response is the card definition for Response // Resurgence.
//
// Type: Instant // Sorcery
// Cost: {R/W}{R/W} // {3}{R}{W}
// Face: Resurgence — Sorcery ({3}{R}{W})
//
// Oracle text:
//
//	Response deals 5 damage to target attacking or blocking creature.
var Response = newResponse

func newResponse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Response",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
			}),
			Colors: []color.Color{color.White, color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target attacking or blocking creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttackingOrBlocking}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Response deals 5 damage to target attacking or blocking creature.
		`,
		},
		Layout: game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{
			Name: "Resurgence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
			}),
			Colors: []color.Color{color.White, color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:    game.DurationPermanent,
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									Layer:       game.LayerAbility,
									AddKeywords: []game.Keyword{game.FirstStrike, game.Vigilance},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.AddExtraPhases{
							Combat: true,
							Main:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Creatures you control gain first strike and vigilance until end of turn. After this main phase, there is an additional combat phase followed by an additional main phase.
		`,
		}),
	}
}
