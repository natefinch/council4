package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InvasionOfSegovia is the card definition for Invasion of Segovia // Caetus, Sea Tyrant of Segovia.
//
// Type: Battle — Siege // Legendary Creature — Serpent
// Face: Caetus, Sea Tyrant of Segovia — Legendary Creature — Serpent
//
// Oracle text:
//
//	(As a Siege enters, choose an opponent to protect it. You and others can attack it. When it's defeated, exile it, then cast it transformed.)
//	When this Siege enters, create two 1/1 blue Kraken creature tokens with trample.
var InvasionOfSegovia = newInvasionOfSegovia

func newInvasionOfSegovia() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Invasion of Segovia",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Battle},
			Subtypes: []types.Sub{types.Siege},
			Defense:  opt.Val(4),
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
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(invasionOfSegoviaToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			(As a Siege enters, choose an opponent to protect it. You and others can attack it. When it's defeated, exile it, then cast it transformed.)
			When this Siege enters, create two 1/1 blue Kraken creature tokens with trample.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:       "Caetus, Sea Tyrant of Segovia",
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Serpent},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectGrantSpellKeyword,
							AffectedController: game.ControllerYou,
							CardSelection:      game.Selection{ExcludedTypes: []types.Card{types.Creature}},
							GrantedKeyword:     game.Convoke,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 4,
								Constraint: "up to four target creatures",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(0),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(1),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(2),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(3),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Noncreature spells you cast have convoke. (Your creatures can help cast those spells. Each creature you tap while casting a noncreature spell pays for {1} or one mana of that creature's color.)
			At the beginning of your end step, untap up to four target creatures.
		`,
		}),
	}
}

var invasionOfSegoviaToken = newInvasionOfSegoviaToken()

func newInvasionOfSegoviaToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Kraken",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kraken},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
		},
	}
}
