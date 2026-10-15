package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HamaPasharRuinSeeker is the card definition for Hama Pashar, Ruin Seeker.
//
// Type: Legendary Creature — Human Wizard
// Cost: {1}{W}{U}
//
// Oracle text:
//
//	Room abilities of dungeons you own trigger an additional time.
var HamaPasharRuinSeeker = newHamaPasharRuinSeeker

func newHamaPasharRuinSeeker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Hama Pashar, Ruin Seeker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Wizard},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Room abilities of dungeons you own trigger an additional time.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:     game.RuleEffectAdditionalTriggerForRoomAbility,
							Duration: game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			Room abilities of dungeons you own trigger an additional time.
		`,
		},
	}
}
