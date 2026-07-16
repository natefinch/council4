package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DelneyStreetwiseLookout is the card definition for Delney, Streetwise Lookout.
//
// Type: Legendary Creature — Human Scout
// Cost: {2}{W}
//
// Oracle text:
//
//	Creatures you control with power 2 or less can't be blocked by creatures with power 3 or greater.
//	If a triggered ability of a creature you control with power 2 or less triggers, that ability triggers an additional time.
var DelneyStreetwiseLookout = newDelneyStreetwiseLookout

func newDelneyStreetwiseLookout() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Delney, Streetwise Lookout",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Scout},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedController: game.ControllerYou,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerGreaterOrEqual,
								Power: 3,
							},
							PermanentTypes:    []types.Card{types.Creature},
							AffectedSelection: game.Selection{Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})},
						},
					},
				},
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectAdditionalTriggerForControlledPermanent,
							AffectedSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})},
						},
					},
				},
			},
			OracleText: `
			Creatures you control with power 2 or less can't be blocked by creatures with power 3 or greater.
			If a triggered ability of a creature you control with power 2 or less triggers, that ability triggers an additional time.
		`,
		},
	}
}
