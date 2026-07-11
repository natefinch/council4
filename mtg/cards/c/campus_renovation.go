package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CampusRenovation is the card definition for Campus Renovation.
//
// Type: Sorcery
// Cost: {3}{R}{W}
//
// Oracle text:
//
//	Return up to one target artifact or enchantment card from your graveyard to the battlefield. Exile the top two cards of your library. Until the end of your next turn, you may play those cards.
var CampusRenovation = newCampusRenovation

func newCampusRenovation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Campus Renovation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
			}),
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 1,
						Constraint: "up to one target artifact or enchantment card from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
						},
					},
					{
						Primitive: game.ImpulseExile{
							Player:   game.ControllerReference(),
							Amount:   game.Fixed(2),
							Duration: game.DurationUntilEndOfYourNextTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return up to one target artifact or enchantment card from your graveyard to the battlefield. Exile the top two cards of your library. Until the end of your next turn, you may play those cards.
		`,
		},
	}
}
