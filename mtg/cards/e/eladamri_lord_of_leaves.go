package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EladamriLordOfLeaves is the card definition for Eladamri, Lord of Leaves.
//
// Type: Legendary Creature — Elf Warrior
// Cost: {G}{G}
//
// Oracle text:
//
//	Other Elf creatures have forestwalk. (They can't be blocked as long as defending player controls a Forest.)
//	Other Elves have shroud. (They can't be the targets of spells or abilities.)
var EladamriLordOfLeaves = newEladamriLordOfLeaves

func newEladamriLordOfLeaves() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Eladamri, Lord of Leaves",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Warrior},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Other Elf creatures have forestwalk. (They can't be blocked as long as defending player controls a Forest.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.BattlefieldGroupExcluding(game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}}, game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.StaticAbility{
									Text: "Forestwalk",
									KeywordAbilities: []game.KeywordAbility{
										game.LandwalkKeyword{
											Subtype: types.Forest,
										},
									},
								},
							},
						},
					},
				},
				game.StaticAbility{
					Text: "Other Elves have shroud. (They can't be the targets of spells or abilities.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.BattlefieldGroupExcluding(game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Shroud},
						},
					},
				},
			},
			OracleText: `
			Other Elf creatures have forestwalk. (They can't be blocked as long as defending player controls a Forest.)
			Other Elves have shroud. (They can't be the targets of spells or abilities.)
		`,
		},
	}
}
