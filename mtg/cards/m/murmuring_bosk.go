package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MurmuringBosk is the card definition for Murmuring Bosk.
//
// Type: Land — Forest
//
// Oracle text:
//
//	({T}: Add {G}.)
//	As this land enters, you may reveal a Treefolk card from your hand. If you don't, this land enters tapped.
//	{T}: Add {W} or {B}. This land deals 1 damage to you.
var MurmuringBosk = newMurmuringBosk

func newMurmuringBosk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name:     "Murmuring Bosk",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
				game.ManaAbility{
					Text: "{T}: Add {W} or {B}. This land deals 1 damage to you.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Choose{
											Choice: game.ResolutionChoice{
												Kind:   game.ResolutionChoiceMana,
												Prompt: "Choose a color",
												Colors: []mana.Color{mana.W, mana.B},
											},
											PublishChoice: game.ChoiceKey("oracle-mana-color"),
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:     game.Fixed(1),
											ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
										},
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(1),
											Recipient:    game.PlayerDamageRecipient(game.ControllerReference()),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may reveal a Treefolk card from your hand. If you don't, this land enters tapped.", game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReveal,
							Amount:      1,
							Source:      zone.Hand,
							SubtypesAny: cost.SubtypeSet{types.Treefolk},
						},
					},
				}),
			},
			OracleText: `
			({T}: Add {G}.)
			As this land enters, you may reveal a Treefolk card from your hand. If you don't, this land enters tapped.
			{T}: Add {W} or {B}. This land deals 1 damage to you.
		`,
		},
	}
}
