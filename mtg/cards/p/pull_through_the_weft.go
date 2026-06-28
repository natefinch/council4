package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PullThroughTheWeft is the card definition for Pull Through the Weft.
//
// Type: Sorcery
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Return up to two target nonland permanent cards from your graveyard to your hand, then return up to two target land cards from your graveyard to the battlefield tapped.
var PullThroughTheWeft = newPullThroughTheWeft()

func newPullThroughTheWeft() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Pull Through the Weft",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 2,
						Constraint: "up to two target nonland permanent cards from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
					},
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 2,
						Constraint: "up to two target land cards from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1}),
							EntryTapped: true,
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 2}),
							EntryTapped: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return up to two target nonland permanent cards from your graveyard to your hand, then return up to two target land cards from your graveyard to the battlefield tapped.
		`,
		},
	}
}
