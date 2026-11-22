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

const spiritEnKorOracleID = "f066174a-a959-4365-920f-c04506d94a5a"

const spiritEnKorRedirectText = "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead."

var SpiritEnKor = cardset.Entry{
	OracleID: spiritEnKorOracleID,
	Name:     "Spirit en-Kor",
	New:      newSpiritEnKor,
	Verification: "mtg/rules/spirit_en_kor_test.go: TestSpiritEnKorFlyingAndRedirectsOneCombatDamage; " +
		"printing 8c57ca87-75b0-4058-acd8-5394545d0846; Scryfall rulings URI https://api.scryfall.com/cards/8c57ca87-75b0-4058-acd8-5394545d0846/rulings; " +
		"CR 2026-11-06 115.1c, 601.2c, 602.1a, 602.2a-b, 120.1, 120.3e, 614.1a, 614.5, 614.9, 616.1, 702.9b",
}

func newSpiritEnKor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Spirit en-Kor",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.W}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Kor, types.Spirit},
			Colors:     []color.Color{color.White},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			OracleText: "Flying\n" + spiritEnKorRedirectText,
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:           spiritEnKorRedirectText,
				ZoneOfFunction: zone.Battlefield,
				Content: game.Mode{
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypes: []types.Card{types.Creature},
							Controller:    game.ControllerYou,
						}),
						Chooser: game.TargetChooserController,
					}},
					Sequence: []game.Instruction{{
						Primitive: game.PreventDamage{
							Amount:         game.Fixed(1),
							Object:         game.SourcePermanentReference(),
							Redirect:       game.DamageRedirectToObject,
							RedirectObject: game.TargetPermanentReference(0),
						},
					}},
				}.Ability(),
			}},
		},
	}
}
