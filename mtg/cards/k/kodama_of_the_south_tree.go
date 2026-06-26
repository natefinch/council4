package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KodamaOfTheSouthTree is the card definition for Kodama of the South Tree.
//
// Type: Legendary Creature — Spirit
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Whenever you cast a Spirit or Arcane spell, each other creature you control gets +1/+1 and gains trample until end of turn.
var KodamaOfTheSouthTree = newKodamaOfTheSouthTree()

func newKodamaOfTheSouthTree() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Kodama of the South Tree",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spirit},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Spirit"), types.Sub("Arcane")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}, game.SourcePermanentReference()),
											PowerDelta:     1,
											ToughnessDelta: 1,
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}, game.SourcePermanentReference()),
											AddKeywords: []game.Keyword{
												game.Trample,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you cast a Spirit or Arcane spell, each other creature you control gets +1/+1 and gains trample until end of turn.
		`,
		},
	}
}
