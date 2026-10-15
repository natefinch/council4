package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ThrashingWumpus is the card definition for Thrashing Wumpus.
//
// Type: Creature — Beast
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	{B}: This creature deals 1 damage to each creature and each player.
var ThrashingWumpus = newThrashingWumpus

func newThrashingWumpus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Thrashing Wumpus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{B}: This creature deals 1 damage to each creature and each player.",
					ManaCost: opt.Val(cost.Mana{
						cost.B,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(1),
											Recipient:    game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(1),
											Recipient:    game.PlayerGroupDamageRecipient(game.AllPlayersReference()),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			{B}: This creature deals 1 damage to each creature and each player.
		`,
		},
	}
}
