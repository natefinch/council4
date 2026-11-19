package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	deepwoodLegateOracleID  = "414eb407-dd39-41d2-bf25-a7a77ff35cb4"
	kyrenLegateOracleID     = "41b46aa5-6091-4bc4-8d17-d2c6b5fe1c5e"
	choArrimLegateOracleID  = "780c9cb2-9d98-4789-87e0-cb81db477204"
	saprazzanLegateOracleID = "d20b62fc-e248-4c84-87ac-7319b92531d4"
	rushwoodLegateOracleID  = "f3623759-bf3a-4846-9ef7-1b732ad4355f"
)

var (
	DeepwoodLegate = cardset.Entry{
		OracleID: deepwoodLegateOracleID,
		Name:     "Deepwood Legate",
		New:      newDeepwoodLegate,
		Verification: "mtg/rules/legate_cycle_test.go: " +
			"TestDeepwoodLegateConditionalFreeCastAndPump",
	}
	KyrenLegate = cardset.Entry{
		OracleID: kyrenLegateOracleID,
		Name:     "Kyren Legate",
		New:      newKyrenLegate,
		Verification: "mtg/rules/legate_cycle_test.go: " +
			"TestKyrenLegateConditionalFreeCastAndHaste",
	}
	ChoArrimLegate = cardset.Entry{
		OracleID: choArrimLegateOracleID,
		Name:     "Cho-Arrim Legate",
		New:      newChoArrimLegate,
		Verification: "mtg/rules/legate_cycle_test.go: " +
			"TestChoArrimLegateConditionalFreeCastAndProtectionFromBlack",
	}
	SaprazzanLegate = cardset.Entry{
		OracleID: saprazzanLegateOracleID,
		Name:     "Saprazzan Legate",
		New:      newSaprazzanLegate,
		Verification: "mtg/rules/legate_cycle_test.go: " +
			"TestSaprazzanLegateConditionalFreeCastAndFlying",
	}
	RushwoodLegate = cardset.Entry{
		OracleID: rushwoodLegateOracleID,
		Name:     "Rushwood Legate",
		New:      newRushwoodLegate,
		Verification: "mtg/rules/legate_cycle_test.go: " +
			"TestRushwoodLegateConditionalFreeCast",
	}
)

func newDeepwoodLegate() *game.CardDef {
	const pumpText = "{B}: This creature gets +1/+1 until end of turn."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Deepwood Legate",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.B}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Shade},
			Colors:     []color.Color{color.Black},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			OracleText: "If an opponent controls a Forest and you control a Swamp, you may cast this spell without paying its mana cost.\n" + pumpText,
			AlternativeCosts: []cost.Alternative{{
				Label:    "Cast without paying its mana cost",
				ManaCost: opt.Val(cost.Mana{}),
				Conditions: []cost.AlternativePredicate{
					{Kind: cost.AlternativeConditionControlsPermanentSubtype, Subtype: types.Swamp},
					{Kind: cost.AlternativeConditionOpponentControlsPermanentSubtype, Subtype: types.Forest},
				},
			}},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:           pumpText,
				ManaCost:       opt.Val(cost.Mana{cost.B}),
				ZoneOfFunction: zone.Battlefield,
				Content: game.Mode{
					Text: pumpText,
					Sequence: []game.Instruction{{
						Primitive: game.ModifyPT{
							Object:         game.SourcePermanentReference(),
							PowerDelta:     game.Fixed(1),
							ToughnessDelta: game.Fixed(1),
							Duration:       game.DurationUntilEndOfTurn,
						},
					}},
				}.Ability(),
			}},
		},
	}
}

func newKyrenLegate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Kyren Legate",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.R}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Goblin},
			Colors:     []color.Color{color.Red},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			OracleText: "If an opponent controls a Plains and you control a Mountain, you may cast this spell without paying its mana cost.\nHaste",
			AlternativeCosts: []cost.Alternative{{
				Label:    "Cast without paying its mana cost",
				ManaCost: opt.Val(cost.Mana{}),
				Conditions: []cost.AlternativePredicate{
					{Kind: cost.AlternativeConditionControlsPermanentSubtype, Subtype: types.Mountain},
					{Kind: cost.AlternativeConditionOpponentControlsPermanentSubtype, Subtype: types.Plains},
				},
			}},
			StaticAbilities: []game.StaticAbility{game.HasteStaticBody},
		},
	}
}

func newChoArrimLegate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Cho-Arrim Legate",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.W}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Colors:     []color.Color{color.White},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			OracleText: "If an opponent controls a Swamp and you control a Plains, you may cast this spell without paying its mana cost.\nProtection from black",
			AlternativeCosts: []cost.Alternative{{
				Label:    "Cast without paying its mana cost",
				ManaCost: opt.Val(cost.Mana{}),
				Conditions: []cost.AlternativePredicate{
					{Kind: cost.AlternativeConditionControlsPermanentSubtype, Subtype: types.Plains},
					{Kind: cost.AlternativeConditionOpponentControlsPermanentSubtype, Subtype: types.Swamp},
				},
			}},
			StaticAbilities: []game.StaticAbility{game.ProtectionFromColorsStaticAbility(color.Black)},
		},
	}
}

func newSaprazzanLegate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Saprazzan Legate",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.U}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Merfolk, types.Soldier},
			Colors:     []color.Color{color.Blue},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			OracleText: "If an opponent controls a Mountain and you control an Island, you may cast this spell without paying its mana cost.\nFlying",
			AlternativeCosts: []cost.Alternative{{
				Label:    "Cast without paying its mana cost",
				ManaCost: opt.Val(cost.Mana{}),
				Conditions: []cost.AlternativePredicate{
					{Kind: cost.AlternativeConditionControlsPermanentSubtype, Subtype: types.Island},
					{Kind: cost.AlternativeConditionOpponentControlsPermanentSubtype, Subtype: types.Mountain},
				},
			}},
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
		},
	}
}

func newRushwoodLegate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Rushwood Legate",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.G}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dryad},
			Colors:     []color.Color{color.Green},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			OracleText: "If an opponent controls an Island and you control a Forest, you may cast this spell without paying its mana cost.",
			AlternativeCosts: []cost.Alternative{{
				Label:    "Cast without paying its mana cost",
				ManaCost: opt.Val(cost.Mana{}),
				Conditions: []cost.AlternativePredicate{
					{Kind: cost.AlternativeConditionControlsPermanentSubtype, Subtype: types.Forest},
					{Kind: cost.AlternativeConditionOpponentControlsPermanentSubtype, Subtype: types.Island},
				},
			}},
		},
	}
}
