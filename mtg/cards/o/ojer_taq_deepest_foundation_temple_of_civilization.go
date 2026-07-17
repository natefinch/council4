package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OjerTaqDeepestFoundation is the card definition for Ojer Taq, Deepest Foundation // Temple of Civilization.
//
// Type: Legendary Creature — God // Land
// Face: Temple of Civilization — Land
//
// Oracle text:
//
//	Vigilance
//	If one or more creature tokens would be created under your control, three times that many of those tokens are created instead.
//	When Ojer Taq dies, return it to the battlefield tapped and transformed under its owner's control.
var OjerTaqDeepestFoundation = newOjerTaqDeepestFoundation

func newOjerTaqDeepestFoundation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ojer Taq, Deepest Foundation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.God},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
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
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source:           game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
									EntryTapped:      true,
									EntryTransformed: true,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.TokenCreationReplacementFiltered("If one or more creature tokens would be created under your control, three times that many of those tokens are created instead.", &game.TokenCreationReplacementSpec{Multiplier: 3, Types: []types.Card{types.Creature}, Filter: game.TriggerControllerYou}),
			},
			OracleText: `
			Vigilance
			If one or more creature tokens would be created under your control, three times that many of those tokens are created instead.
			When Ojer Taq dies, return it to the battlefield tapped and transformed under its owner's control.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:  "Temple of Civilization",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{2}{W}, {T}: Transform this land. Activate only if you attacked with three or more creatures this turn and only as a sorcery.",
					ManaCost:        opt.Val(cost.Mana{cost.O(2), cost.W}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Timing:          game.SorceryOnly,
					ActivationCondition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Controller: game.TriggerControllerYou,
						}, Window: game.EventHistoryCurrentTurn, MinCount: 3}),
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			OracleText: `
			(Transforms from Ojer Taq, Deepest Foundation.)
			{T}: Add {W}.
			{2}{W}, {T}: Transform this land. Activate only if you attacked with three or more creatures this turn and only as a sorcery.
		`,
		}),
	}
}
