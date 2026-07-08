package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// QuestForTheGravelord is the card definition for Quest for the Gravelord.
//
// Type: Enchantment
// Cost: {B}
//
// Oracle text:
//
//	Whenever a creature dies, you may put a quest counter on this enchantment.
//	Remove three quest counters from this enchantment and sacrifice it: Create a 5/5 black Zombie Giant creature token.
var QuestForTheGravelord = newQuestForTheGravelord

func newQuestForTheGravelord() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Quest for the Gravelord",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Remove three quest counters from this enchantment and sacrifice it: Create a 5/5 black Zombie Giant creature token.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove three quest counters from this enchantment",
							Amount:      3,
							CounterKind: counter.Quest,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "sacrifice it",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(questForTheGravelordToken),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Quest,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a creature dies, you may put a quest counter on this enchantment.
			Remove three quest counters from this enchantment and sacrifice it: Create a 5/5 black Zombie Giant creature token.
		`,
		},
	}
}

var questForTheGravelordToken = newQuestForTheGravelordToken()

func newQuestForTheGravelordToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie Giant",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Giant},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
		},
	}
}
