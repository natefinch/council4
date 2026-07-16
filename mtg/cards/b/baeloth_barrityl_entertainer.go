package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BaelothBarritylEntertainer is the card definition for Baeloth Barrityl, Entertainer.
//
// Type: Legendary Creature — Elf Shaman
// Cost: {4}{R}
//
// Oracle text:
//
//	Creatures your opponents control with power less than Baeloth Barrityl's power are goaded. (They attack each combat if able and attack a player other than you if able.)
//	Whenever a goaded attacking or blocking creature dies, you create a Treasure token.
//	Choose a Background (You can have a Background as a second commander.)
var BaelothBarritylEntertainer = newBaelothBarritylEntertainer

func newBaelothBarritylEntertainer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Baeloth Barrityl, Entertainer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Shaman},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectGoaded,
							PermanentTypes:    []types.Card{types.Creature},
							AffectedSelection: game.Selection{Controller: game.ControllerOpponent, PowerLessThanSource: true},
						},
					},
				},
				game.ChooseABackgroundStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttackingOrBlocking, MatchGoaded: true},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(baelothBarritylEntertainerToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Creatures your opponents control with power less than Baeloth Barrityl's power are goaded. (They attack each combat if able and attack a player other than you if able.)
			Whenever a goaded attacking or blocking creature dies, you create a Treasure token.
			Choose a Background (You can have a Background as a second commander.)
		`,
		},
	}
}

var baelothBarritylEntertainerToken = newBaelothBarritylEntertainerToken()

func newBaelothBarritylEntertainerToken() *game.CardDef {
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
