package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VisageOfBolas is the card definition for Visage of Bolas.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	When this artifact enters, you may search your library and/or graveyard for a card named Nicol Bolas, the Deceiver, reveal it, and put it into your hand. If you search your library this way, shuffle.
//	{T}: Add {U}, {B}, or {R}.
var VisageOfBolas = newVisageOfBolas()

func newVisageOfBolas() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Visage of Bolas",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B, mana.R),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:    zone.Library,
										Destination:   zone.Hand,
										Name:          "Nicol Bolas, the Deceiver",
										Reveal:        true,
										AlsoGraveyard: true,
									},
									Amount: game.Fixed(1),
								},
								Optional:      true,
								OptionalActor: opt.Val(game.ControllerReference()),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this artifact enters, you may search your library and/or graveyard for a card named Nicol Bolas, the Deceiver, reveal it, and put it into your hand. If you search your library this way, shuffle.
			{T}: Add {U}, {B}, or {R}.
		`,
		},
	}
}
