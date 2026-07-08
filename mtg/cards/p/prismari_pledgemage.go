package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrismariPledgemage is the card definition for Prismari Pledgemage.
//
// Type: Creature — Orc Wizard
// Cost: {U/R}{U/R}
//
// Oracle text:
//
//	Defender
//	Magecraft — Whenever you cast or copy an instant or sorcery spell, this creature can attack this turn as though it didn't have defender.
var PrismariPledgemage = newPrismariPledgemage

func newPrismariPledgemage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Prismari Pledgemage",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.U, mana.R),
				cost.HybridMana(mana.U, mana.R),
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:          game.EventSpellCast,
							Controller:     game.TriggerControllerYou,
							MatchSpellCopy: true,
							CardSelection:  game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.SourcePermanentReference()),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCanAttackAsThoughDefender,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Defender
			Magecraft — Whenever you cast or copy an instant or sorcery spell, this creature can attack this turn as though it didn't have defender.
		`,
		},
	}
}
