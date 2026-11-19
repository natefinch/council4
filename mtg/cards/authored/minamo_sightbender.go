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

var MinamoSightbender = cardset.Entry{
	OracleID: "c4f1d8c8-29e5-43fe-8281-7d42d0173dc5",
	Name:     "Minamo Sightbender",
	New:      newMinamoSightbender,
	Verification: "mtg/rules/target_mana_value_x_test.go: TestMinamoSightbenderPowerAtMostX; " +
		"printing a598d08d-c9d4-4de5-bef5-998dd115b8b6; CR 2026-09-25 " +
		"107.3a, 208.1, 509.1a-b, 514.2, 602.2b, 602.5a, 608.2b, 611.2a, 613.4a-d",
}

func newMinamoSightbender() *game.CardDef {
	const text = "{X}, {T}: Target creature with power X or less can't be blocked this turn."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Minamo Sightbender",
			OracleText: text,
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.U}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Wizard},
			Colors:     []color.Color{color.Blue},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:            text,
				ZoneOfFunction:  zone.Battlefield,
				ManaCost:        opt.Val(cost.Mana{cost.X}),
				AdditionalCosts: cost.Tap,
				Content: game.Mode{
					Targets: []game.TargetSpec{{
						MinTargets:   1,
						MaxTargets:   1,
						Constraint:   "target creature with power X or less",
						Allow:        game.TargetAllowPermanent,
						Selection:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						PowerAtMostX: true,
					}},
					Sequence: []game.Instruction{{
						Primitive: game.ApplyRule{
							Object:      opt.Val(game.TargetPermanentReference(0)),
							RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectCantBeBlocked}},
							Duration:    game.DurationUntilEndOfTurn,
						},
					}},
				}.Ability(),
			}},
		},
	}
}
