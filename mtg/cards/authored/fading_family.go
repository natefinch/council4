package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	skyshroudRidgebackOracleID  = "34a160bc-5d7e-483c-b24c-ab30b7d313b0"
	cloudskateOracleID          = "648e4072-b991-46fd-b057-1e807091dcdf"
	skyshroudBehemothOracleID   = "77bee041-6f96-44e3-a18c-b52e508b9bdf"
	rejuvenationChamberOracleID = "c1ebe189-db71-47cd-a763-8cb8c305383a"
	blastodermOracleID          = "d6ee8bd6-0207-46ee-9892-c1d1b23b934e"

	skyshroudRidgebackFadingText  = "Fading 2 (This creature enters with two fade counters on it. At the beginning of your upkeep, remove a fade counter from it. If you can't, sacrifice it.)"
	cloudskateFadingText          = "Fading 3 (This creature enters with three fade counters on it. At the beginning of your upkeep, remove a fade counter from it. If you can't, sacrifice it.)"
	skyshroudBehemothFadingText   = "Fading 2 (This creature enters with two fade counters on it. At the beginning of your upkeep, remove a fade counter from it. If you can't, sacrifice it.)"
	rejuvenationChamberFadingText = "Fading 2 (This artifact enters with two fade counters on it. At the beginning of your upkeep, remove a fade counter from it. If you can't, sacrifice it.)"
	blastodermFadingText          = "Fading 3 (This creature enters with three fade counters on it. At the beginning of your upkeep, remove a fade counter from it. If you can't, sacrifice it.)"

	skyshroudRidgebackText  = skyshroudRidgebackFadingText
	cloudskateText          = "Flying\n" + cloudskateFadingText
	skyshroudBehemothText   = skyshroudBehemothFadingText + "\nThis creature enters tapped."
	rejuvenationChamberText = rejuvenationChamberFadingText + "\n{T}: You gain 2 life."
	blastodermText          = "Shroud (This creature can't be the target of spells or abilities.)\n" + blastodermFadingText

	fadingCounterRemovalResult game.ResultKey = "fading-upkeep/remove-counter"
)

var SkyshroudRidgeback = cardset.Entry{
	OracleID: skyshroudRidgebackOracleID,
	Name:     "Skyshroud Ridgeback",
	New:      newSkyshroudRidgeback,
	Verification: "mtg/rules/fading_family_test.go: " +
		"TestSkyshroudRidgebackUpkeepRemovesFadeThenSacrifices",
}

var Cloudskate = cardset.Entry{
	OracleID: cloudskateOracleID,
	Name:     "Cloudskate",
	New:      newCloudskate,
	Verification: "mtg/rules/fading_family_test.go: " +
		"TestCloudskateFadingAndFlying",
}

var SkyshroudBehemoth = cardset.Entry{
	OracleID: skyshroudBehemothOracleID,
	Name:     "Skyshroud Behemoth",
	New:      newSkyshroudBehemoth,
	Verification: "mtg/rules/fading_family_test.go: " +
		"TestSkyshroudBehemothEntersTappedAndFades",
}

var RejuvenationChamber = cardset.Entry{
	OracleID: rejuvenationChamberOracleID,
	Name:     "Rejuvenation Chamber",
	New:      newRejuvenationChamber,
	Verification: "mtg/rules/fading_family_test.go: " +
		"TestRejuvenationChamberActivatesForLifeAndFades",
}

var Blastoderm = cardset.Entry{
	OracleID: blastodermOracleID,
	Name:     "Blastoderm",
	New:      newBlastoderm,
	Verification: "mtg/rules/fading_family_test.go: " +
		"TestBlastodermShroudAndFading",
}

func newSkyshroudRidgeback() *game.CardDef {
	replacement, trigger := fadingAbilities(skyshroudRidgebackText, 2)
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:                 "Skyshroud Ridgeback",
			ManaCost:             opt.Val(cost.Mana{cost.G}),
			Colors:               []color.Color{color.Green},
			Types:                []types.Card{types.Creature},
			Subtypes:             []types.Sub{types.Beast},
			Power:                opt.Val(game.PT{Value: 2}),
			Toughness:            opt.Val(game.PT{Value: 3}),
			OracleText:           skyshroudRidgebackText,
			ReplacementAbilities: []game.ReplacementAbility{replacement},
			TriggeredAbilities:   []game.TriggeredAbility{trigger},
		},
	}
}

func newCloudskate() *game.CardDef {
	replacement, trigger := fadingAbilities(cloudskateFadingText, 3)
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:                 "Cloudskate",
			ManaCost:             opt.Val(cost.Mana{cost.O(1), cost.U}),
			Colors:               []color.Color{color.Blue},
			Types:                []types.Card{types.Creature},
			Subtypes:             []types.Sub{types.Illusion},
			Power:                opt.Val(game.PT{Value: 2}),
			Toughness:            opt.Val(game.PT{Value: 2}),
			OracleText:           cloudskateText,
			StaticAbilities:      []game.StaticAbility{game.FlyingStaticBody},
			ReplacementAbilities: []game.ReplacementAbility{replacement},
			TriggeredAbilities:   []game.TriggeredAbility{trigger},
		},
	}
}

func newSkyshroudBehemoth() *game.CardDef {
	replacement, trigger := fadingAbilities(skyshroudBehemothFadingText, 2)
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Skyshroud Behemoth",
			ManaCost:   opt.Val(cost.Mana{cost.O(5), cost.G, cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Beast},
			Power:      opt.Val(game.PT{Value: 10}),
			Toughness:  opt.Val(game.PT{Value: 10}),
			OracleText: skyshroudBehemothText,
			ReplacementAbilities: []game.ReplacementAbility{
				replacement,
				game.EntersTappedReplacement("This creature enters tapped."),
			},
			TriggeredAbilities: []game.TriggeredAbility{trigger},
		},
	}
}

func newRejuvenationChamber() *game.CardDef {
	replacement, trigger := fadingAbilities(rejuvenationChamberFadingText, 2)
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:                 "Rejuvenation Chamber",
			ManaCost:             opt.Val(cost.Mana{cost.O(3)}),
			Types:                []types.Card{types.Artifact},
			OracleText:           rejuvenationChamberText,
			ReplacementAbilities: []game.ReplacementAbility{replacement},
			TriggeredAbilities:   []game.TriggeredAbility{trigger},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:            "{T}: You gain 2 life.",
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.GainLife{
						Amount: game.Fixed(2),
						Player: game.ControllerReference(),
					},
				}}}.Ability(),
			}},
		},
	}
}

func newBlastoderm() *game.CardDef {
	replacement, trigger := fadingAbilities(blastodermFadingText, 3)
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:                 "Blastoderm",
			ManaCost:             opt.Val(cost.Mana{cost.O(2), cost.G, cost.G}),
			Colors:               []color.Color{color.Green},
			Types:                []types.Card{types.Creature},
			Subtypes:             []types.Sub{types.Beast},
			Power:                opt.Val(game.PT{Value: 5}),
			Toughness:            opt.Val(game.PT{Value: 5}),
			OracleText:           blastodermText,
			StaticAbilities:      []game.StaticAbility{game.ShroudStaticBody},
			ReplacementAbilities: []game.ReplacementAbility{replacement},
			TriggeredAbilities:   []game.TriggeredAbility{trigger},
		},
	}
}

func fadingAbilities(text string, amount int) (game.ReplacementAbility, game.TriggeredAbility) {
	replacement := game.EntersWithCountersReplacement(text, game.CounterPlacement{
		Kind: counter.Fade, Amount: amount,
	})
	trigger := game.TriggeredAbility{
		Text: text,
		Trigger: game.TriggerCondition{
			Type: game.TriggerAt,
			Pattern: game.TriggerPattern{
				Event:      game.EventBeginningOfStep,
				Controller: game.TriggerControllerYou,
				Step:       game.StepUpkeep,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{
			{
				Primitive: game.RemoveCounter{
					Amount: game.Fixed(1), Object: game.SourcePermanentReference(),
					CounterKind: counter.Fade,
				},
				PublishResult: fadingCounterRemovalResult,
			},
			{
				Primitive: game.Sacrifice{Object: game.SourcePermanentReference()},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key: fadingCounterRemovalResult, Succeeded: game.TriFalse,
				}),
			},
		}}.Ability(),
	}
	return replacement, trigger
}
