package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BelligerentBrontodon is the card definition for Belligerent Brontodon.
//
// Type: Creature — Dinosaur
// Cost: {5}{G}{W}
//
// Oracle text:
//
//	Each creature you control assigns combat damage equal to its toughness rather than its power.
var BelligerentBrontodon = newBelligerentBrontodon

func newBelligerentBrontodon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Belligerent Brontodon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectAssignCombatDamageUsingToughness,
							AffectedController: game.ControllerYou,
							PermanentTypes:     []types.Card{types.Creature},
						},
					},
				},
			},
			OracleText: `
			Each creature you control assigns combat damage equal to its toughness rather than its power.
		`,
		},
	}
}
