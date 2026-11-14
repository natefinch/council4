package authored

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	squeakingPieSneakOracleID  = "7243b935-73fc-47d2-ac65-ab390ed94f3f"
	sadisticSkymarcherOracleID = "9e1fe5ec-1669-4587-8e42-b189fd6d39e0"
	flamekinBladewhirlOracleID = "c37770aa-4770-4198-b6a6-f9ff86dbb06b"
	silvergillAdeptOracleID    = "cb66c5ec-e7a4-4bd2-b825-6c9b846ba40d"
	wrensRunVanquisherOracleID = "fe9cdd15-a390-4a1e-bae9-474ce8d355b8"
)

var (
	SqueakingPieSneak = cardset.Entry{
		OracleID: squeakingPieSneakOracleID,
		Name:     "Squeaking Pie Sneak",
		Verification: "mtg/rules/reveal_or_pay_creatures_test.go: " +
			"TestSqueakingPieSneakPaysThreeAndHasFear",
		New: newSqueakingPieSneak,
	}
	SadisticSkymarcher = cardset.Entry{
		OracleID: sadisticSkymarcherOracleID,
		Name:     "Sadistic Skymarcher",
		Verification: "mtg/rules/reveal_or_pay_creatures_test.go: " +
			"TestSadisticSkymarcherPaysOneAndHasFlyingLifelink",
		New: newSadisticSkymarcher,
	}
	FlamekinBladewhirl = cardset.Entry{
		OracleID: flamekinBladewhirlOracleID,
		Name:     "Flamekin Bladewhirl",
		Verification: "mtg/rules/reveal_or_pay_creatures_test.go: " +
			"TestFlamekinBladewhirlRevealsElemental",
		New: newFlamekinBladewhirl,
	}
	SilvergillAdept = cardset.Entry{
		OracleID: silvergillAdeptOracleID,
		Name:     "Silvergill Adept",
		Verification: "mtg/rules/reveal_or_pay_creatures_test.go: " +
			"TestSilvergillAdeptRevealsMerfolkAndDrawsOnEntry",
		New: newSilvergillAdept,
	}
	WrensRunVanquisher = cardset.Entry{
		OracleID: wrensRunVanquisherOracleID,
		Name:     "Wren's Run Vanquisher",
		Verification: "mtg/rules/reveal_or_pay_creatures_test.go: " +
			"TestWrensRunVanquisherRevealsElfAndHasDeathtouch",
		New: newWrensRunVanquisher,
	}
)

func revealOrPayCost(subtype types.Sub, manaAmount int) cost.AdditionalChoice {
	article := "a"
	if strings.Contains("AEIOU", strings.ToUpper(string(subtype[0]))) {
		article = "an"
	}
	revealText := fmt.Sprintf("reveal %s %s card from your hand", article, subtype)
	return cost.AdditionalChoice{Options: []cost.AdditionalChoiceOption{
		{
			Label: strings.ToUpper(revealText[:1]) + revealText[1:],
			Costs: []cost.Additional{{
				Kind:        cost.AdditionalReveal,
				Text:        revealText,
				Amount:      1,
				Source:      zone.Hand,
				SubtypesAny: cost.SubtypeSet{subtype},
			}},
		},
		{
			Label: fmt.Sprintf("Pay {%d}", manaAmount),
			Mana:  cost.Mana{cost.O(manaAmount)},
		},
	}}
}

func newSqueakingPieSneak() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:      "Squeaking Pie Sneak",
			ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.B}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FearStaticBody,
			},
			AdditionalCostChoices: []cost.AdditionalChoice{
				revealOrPayCost(types.Goblin, 3),
			},
			OracleText: "As an additional cost to cast this spell, reveal a Goblin card from your hand or pay {3}.\n" +
				"Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)",
		},
	}
}

func newSadisticSkymarcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:      "Sadistic Skymarcher",
			ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.B}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
			},
			AdditionalCostChoices: []cost.AdditionalChoice{
				revealOrPayCost(types.Vampire, 1),
			},
			OracleText: "As an additional cost to cast this spell, reveal a Vampire card from your hand or pay {1}.\n" +
				"Flying, lifelink",
		},
	}
}

func newFlamekinBladewhirl() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:      "Flamekin Bladewhirl",
			ManaCost:  opt.Val(cost.Mana{cost.R}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			AdditionalCostChoices: []cost.AdditionalChoice{
				revealOrPayCost(types.Elemental, 3),
			},
			OracleText: "As an additional cost to cast this spell, reveal an Elemental card from your hand or pay {3}.",
		},
	}
}

func newSilvergillAdept() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Silvergill Adept",
			ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.U}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			AdditionalCostChoices: []cost.AdditionalChoice{
				revealOrPayCost(types.Merfolk, 3),
			},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: "When this creature enters, draw a card.",
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentEnteredBattlefield,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.Draw{
						Player: game.ControllerReference(),
						Amount: game.Fixed(1),
					},
				}}}.Ability(),
			}},
			OracleText: "As an additional cost to cast this spell, reveal a Merfolk card from your hand or pay {3}.\n" +
				"When this creature enters, draw a card.",
		},
	}
}

func newWrensRunVanquisher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:      "Wren's Run Vanquisher",
			ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.G}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
			},
			AdditionalCostChoices: []cost.AdditionalChoice{
				revealOrPayCost(types.Elf, 3),
			},
			OracleText: "As an additional cost to cast this spell, reveal an Elf card from your hand or pay {3}.\n" +
				"Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)",
		},
	}
}
