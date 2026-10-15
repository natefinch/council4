package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DeadRevels is the card definition for Dead Revels.
//
// Type: Sorcery
// Cost: {3}{B}
//
// Oracle text:
//
//	Spectacle {1}{B} (You may cast this spell for its spectacle cost rather than its mana cost if an opponent lost life this turn.)
//	Return up to two target creature cards from your graveyard to your hand.
var DeadRevels = newDeadRevels

func newDeadRevels() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dead Revels",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:     "Spectacle",
					ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.B}),
					Condition: cost.AlternativeConditionOpponentLostLifeThisTurn,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 2,
						Constraint: "up to two target creature cards from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind: game.CardReferenceTarget,
							},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind:        game.CardReferenceTarget,
								TargetIndex: 1,
							},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Spectacle {1}{B} (You may cast this spell for its spectacle cost rather than its mana cost if an opponent lost life this turn.)
			Return up to two target creature cards from your graveyard to your hand.
		`,
		},
	}
}
