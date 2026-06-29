package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LeylineDowser is the card definition for Leyline Dowser.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	{1}, {T}: Mill a card. You may put an instant or sorcery card milled this way into your hand.
//	Tap an untapped legendary creature you control: Untap this artifact.
var LeylineDowser = newLeylineDowser()

func newLeylineDowser() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Leyline Dowser",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{1}, {T}: Mill a card. You may put an instant or sorcery card milled this way into your hand.",
					ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:   game.ControllerReference(),
									Look:     game.Fixed(1),
									Take:     game.Fixed(1),
									Filter:   opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}}),
									TakeUpTo: true,
									Reveal:   true,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text: "Tap an untapped legendary creature you control: Untap this artifact.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalTapPermanents,
							Text:               "Tap an untapped legendary creature you control",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							RequireSupertype:   types.Legendary,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{1}, {T}: Mill a card. You may put an instant or sorcery card milled this way into your hand.
			Tap an untapped legendary creature you control: Untap this artifact.
		`,
		},
	}
}
