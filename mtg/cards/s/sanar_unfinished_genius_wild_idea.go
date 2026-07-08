package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SanarUnfinishedGenius is the card definition for Sanar, Unfinished Genius // Wild Idea.
//
// Type: Legendary Creature — Goblin Sorcerer // Sorcery
// Cost: {U}{R} // {3}{U}{R}
// Face: Wild Idea — Sorcery ({3}{U}{R})
//
// Oracle text:
//
//	Sanar enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
//	{T}: Create a Treasure token. Activate only if you've cast an instant or sorcery spell this turn.
var SanarUnfinishedGenius = newSanarUnfinishedGenius

func newSanarUnfinishedGenius() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Sanar, Unfinished Genius",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.R,
			}),
			Colors:         []color.Color{color.Blue, color.Red},
			EntersPrepared: true,
			Supertypes:     []types.Super{types.Legendary},
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Goblin, types.Sorcerer},
			Power:          opt.Val(game.PT{Value: 0}),
			Toughness:      opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Create a Treasure token. Activate only if you've cast an instant or sorcery spell this turn.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
						}, Window: game.EventHistoryCurrentTurn}),
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(sanarUnfinishedGeniusToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Sanar enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
			{T}: Create a Treasure token. Activate only if you've cast an instant or sorcery spell this turn.
		`,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name: "Wild Idea",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.R,
			}),
			Colors: []color.Color{color.Blue, color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Hand,
								Filter:      game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
								Reveal:      true,
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for an instant or sorcery card, reveal it, put it into your hand, then shuffle.
		`,
		}),
	}
}

var sanarUnfinishedGeniusToken = newSanarUnfinishedGeniusToken()

func newSanarUnfinishedGeniusToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Treasure",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Treasure},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Choose{
									Choice: game.ResolutionChoice{
										Kind:   game.ResolutionChoiceMana,
										Prompt: "Choose a color",
										Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
									},
									PublishChoice: game.ChoiceKey("oracle-mana-color"),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:     game.Fixed(1),
									ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
