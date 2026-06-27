package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SkarrgGuildmage is the card definition for Skarrg Guildmage.
//
// Type: Creature — Human Shaman
// Cost: {R}{G}
//
// Oracle text:
//
//	{R}{G}: Creatures you control gain trample until end of turn.
//	{1}{R}{G}: Target land you control becomes a 4/4 Elemental creature until end of turn. It's still a land.
var SkarrgGuildmage = newSkarrgGuildmage()

func newSkarrgGuildmage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Skarrg Guildmage",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{R}{G}: Creatures you control gain trample until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.R, cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Trample,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:           "{1}{R}{G}: Target land you control becomes a 4/4 Elemental creature until end of turn. It's still a land.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.R, cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target land you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:       game.LayerType,
											AddTypes:    []types.Card{types.Creature},
											AddSubtypes: []types.Sub{types.Elemental},
										},
										game.ContinuousEffect{
											Layer:        game.LayerPowerToughnessSet,
											SetPower:     opt.Val(game.PT{Value: 4}),
											SetToughness: opt.Val(game.PT{Value: 4}),
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{R}{G}: Creatures you control gain trample until end of turn.
			{1}{R}{G}: Target land you control becomes a 4/4 Elemental creature until end of turn. It's still a land.
		`,
		},
	}
}
