package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HelpingHand is the card definition for Helping Hand.
//
// Type: Sorcery
// Cost: {W}
//
// Oracle text:
//
//	Return target creature card with mana value 3 or less from your graveyard to the battlefield tapped.
var HelpingHand = newHelpingHand

func newHelpingHand() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Helping Hand",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card with mana value 3 or less from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							EntryTapped: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target creature card with mana value 3 or less from your graveyard to the battlefield tapped.
		`,
		},
	}
}
