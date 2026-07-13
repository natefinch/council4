package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ZulAshurLichLord is the card definition for Zul Ashur, Lich Lord.
//
// Type: Legendary Creature — Zombie Warlock
// Cost: {1}{B}
//
// Oracle text:
//
//	Ward—Pay 2 life. (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays 2 life.)
//	{T}: You may cast target Zombie creature card from your graveyard this turn.
var ZulAshurLichLord = newZulAshurLichLord

func newZulAshurLichLord() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Zul Ashur, Lich Lord",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Zombie, types.Warlock},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{
					{
						Kind:   cost.AdditionalPayLife,
						Text:   "Pay 2 life",
						Amount: 2,
					},
				}),
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: You may cast target Zombie creature card from your graveyard this turn.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Zombie creature card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Zombie")}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.GrantCastPermission{
									Card:     game.CardReference{Kind: game.CardReferenceTarget},
									FromZone: zone.Graveyard,
									Face:     game.FaceFront,
									Duration: game.DurationUntilEndOfTurn,
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Ward—Pay 2 life. (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays 2 life.)
			{T}: You may cast target Zombie creature card from your graveyard this turn.
		`,
		},
	}
}
