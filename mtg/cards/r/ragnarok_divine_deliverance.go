package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RagnarokDivineDeliverance is the card definition for Ragnarok, Divine Deliverance.
//
// Type: Legendary Creature — Beast Avatar
//
// Oracle text:
//
//	Vigilance, menace, trample, reach, haste
//	When Ragnarok dies, destroy target permanent and return target nonlegendary permanent card from your graveyard to the battlefield.
var RagnarokDivineDeliverance = newRagnarokDivineDeliverance()

func newRagnarokDivineDeliverance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:       "Ragnarok, Divine Deliverance",
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Beast, types.Avatar},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.MenaceStaticBody,
				game.TrampleStaticBody,
				game.ReachStaticBody,
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target permanent",
								Allow:      game.TargetAllowPermanent,
							},
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target nonlegendary permanent card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, ExcludedSupertype: types.Legendary, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance, menace, trample, reach, haste
			When Ragnarok dies, destroy target permanent and return target nonlegendary permanent card from your graveyard to the battlefield.
		`,
		},
		Layout: game.LayoutMeld,
	}
}
