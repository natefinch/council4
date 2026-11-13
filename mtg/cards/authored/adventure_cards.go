package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	rosethornAcolyteOracleID = "639a8804-8513-4ba4-8683-ea1e2e5a9530"
	crystalDragonOracleID    = "56b0cb64-b886-413e-bb20-392bf12d3926"
	pollenShieldHareOracleID = "3ef02f2b-b611-40c3-a1d5-8b34989643cf"

	seasonalRitualText = "Add one mana of any color. (Then exile this card. You may cast the creature later from exile.)"
	robTheHoardText    = "Return target artifact, enchantment, or legendary card from your graveyard to your hand. (Then exile this card. You may cast the creature later from exile.)"
	hareRaisingText    = "Target creature you control gains vigilance and gets +X/+X until end of turn, where X is the number of creatures you control."
)

var (
	RosethornAcolyte = cardset.Entry{
		OracleID: rosethornAcolyteOracleID,
		Name:     "Rosethorn Acolyte",
		New:      newRosethornAcolyte,
		Verification: "mtg/rules/adventure_cards_test.go: " +
			"TestRosethornAcolyteSeasonalRitualScenario",
	}
	CrystalDragon = cardset.Entry{
		OracleID: crystalDragonOracleID,
		Name:     "Crystal Dragon",
		New:      newCrystalDragon,
		Verification: "mtg/rules/adventure_cards_test.go: " +
			"TestCrystalDragonRobTheHoardScenario",
	}
	PollenShieldHare = cardset.Entry{
		OracleID: pollenShieldHareOracleID,
		Name:     "Pollen-Shield Hare",
		New:      newPollenShieldHare,
		Verification: "mtg/rules/adventure_cards_test.go: " +
			"TestPollenShieldHareRaisingScenario",
	}
)

func newRosethornAcolyte() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		Layout:        game.LayoutAdventure,
		CardFace: game.CardFace{
			Name:          "Rosethorn Acolyte",
			ManaCost:      opt.Val(cost.Mana{cost.O(2), cost.G}),
			Colors:        []color.Color{color.Green},
			Types:         []types.Card{types.Creature},
			Subtypes:      []types.Sub{types.Elf, types.Druid},
			Power:         opt.Val(game.PT{Value: 2}),
			Toughness:     opt.Val(game.PT{Value: 3}),
			OracleText:    "{T}: Add one mana of any color.",
			ManaAbilities: []game.ManaAbility{game.TapAnyColorManaAbility()},
		},
		Alternate: opt.Val(game.CardFace{
			Name:       "Seasonal Ritual",
			ManaCost:   opt.Val(cost.Mana{cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Sorcery},
			Subtypes:   []types.Sub{types.Adventure},
			OracleText: seasonalRitualText,
			SpellAbility: opt.Val(game.Mode{
				Text: seasonalRitualText,
				Sequence: []game.Instruction{
					{Primitive: game.Choose{
						Choice: game.ResolutionChoice{
							Kind:   game.ResolutionChoiceMana,
							Prompt: "Choose a color",
							Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
						},
						PublishChoice: game.ChoiceKey("seasonal-ritual-color"),
					}},
					{Primitive: game.AddMana{
						Amount:     game.Fixed(1),
						ChoiceFrom: game.ChoiceKey("seasonal-ritual-color"),
					}},
				},
			}.Ability()),
		}),
	}
}

func newCrystalDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		Layout:        game.LayoutAdventure,
		CardFace: game.CardFace{
			Name:            "Crystal Dragon",
			ManaCost:        opt.Val(cost.Mana{cost.O(4), cost.W, cost.W}),
			Colors:          []color.Color{color.White},
			Types:           []types.Card{types.Creature},
			Subtypes:        []types.Sub{types.Dragon},
			Power:           opt.Val(game.PT{Value: 4}),
			Toughness:       opt.Val(game.PT{Value: 4}),
			OracleText:      "Flying, vigilance",
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody, game.VigilanceStaticBody},
		},
		Alternate: opt.Val(game.CardFace{
			Name:       "Rob the Hoard",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.W}),
			Colors:     []color.Color{color.White},
			Types:      []types.Card{types.Sorcery},
			Subtypes:   []types.Sub{types.Adventure},
			OracleText: robTheHoardText,
			SpellAbility: opt.Val(game.Mode{
				Text: robTheHoardText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target artifact, enchantment, or legendary card from your graveyard",
					Allow:      game.TargetAllowCard,
					TargetZone: zone.Graveyard,
					Selection: opt.Val(game.Selection{
						Controller: game.ControllerYou,
						AnyOf: []game.Selection{
							{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}},
							{Supertypes: []types.Super{types.Legendary}},
						},
					}),
				}},
				Sequence: []game.Instruction{{Primitive: game.MoveCard{
					Card:        game.CardReference{Kind: game.CardReferenceTarget},
					FromZone:    zone.Graveyard,
					Destination: zone.Hand,
				}}},
			}.Ability()),
		}),
	}
}

func newPollenShieldHare() *game.CardDef {
	creaturesYouControl := game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
	}
	creaturesYouControlGroup := game.BattlefieldGroup(creaturesYouControl)
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		Layout:        game.LayoutAdventure,
		CardFace: game.CardFace{
			Name:     "Pollen-Shield Hare",
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Rabbit},
			Power:    opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{
				Value: 2,
			}),
			OracleText: "Creature tokens you control get +1/+1.",
			StaticAbilities: []game.StaticAbility{{
				Text: "Creature tokens you control get +1/+1.",
				ContinuousEffects: []game.ContinuousEffect{{
					Duration: game.DurationPermanent,
					Layer:    game.LayerPowerToughnessModify,
					Group: game.BattlefieldGroup(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Controller:    game.ControllerYou,
						TokenOnly:     true,
					}),
					PowerDelta:     1,
					ToughnessDelta: 1,
				}},
			}},
		},
		Alternate: opt.Val(game.CardFace{
			Name:       "Hare Raising",
			ManaCost:   opt.Val(cost.Mana{cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Sorcery},
			Subtypes:   []types.Sub{types.Adventure},
			OracleText: hareRaisingText,
			SpellAbility: opt.Val(game.Mode{
				Text: hareRaisingText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature you control",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(creaturesYouControl),
				}},
				Sequence: []game.Instruction{
					{Primitive: game.ApplyContinuous{
						Object: opt.Val(game.TargetPermanentReference(0)),
						ContinuousEffects: []game.ContinuousEffect{{
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Vigilance},
						}},
						Duration: game.DurationUntilEndOfTurn,
					}},
					{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(0),
						PowerDelta: game.Dynamic(game.DynamicAmount{
							Kind:  game.DynamicAmountCountSelector,
							Group: creaturesYouControlGroup,
						}),
						ToughnessDelta: game.Dynamic(game.DynamicAmount{
							Kind:  game.DynamicAmountCountSelector,
							Group: creaturesYouControlGroup,
						}),
						Duration: game.DurationUntilEndOfTurn,
					}},
				},
			}.Ability()),
		}),
	}
}
