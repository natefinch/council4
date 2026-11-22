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

const nomadsEnKorOracleID = "b8d395a3-0bfe-45c3-bb3a-820d4f235b88"

const nomadsEnKorRedirectText = "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead."

var NomadsEnKor = cardset.Entry{
	OracleID: nomadsEnKorOracleID,
	Name:     "Nomads en-Kor",
	New:      newNomadsEnKor,
	Verification: "mtg/rules/nomads_en_kor_test.go: TestNomadsEnKorRedirectsOneDamageToTargetCreature; " +
		"printing 48ed076a-909f-4ef0-b85e-a1935b1c2827; Scryfall rulings URI https://api.scryfall.com/cards/48ed076a-909f-4ef0-b85e-a1935b1c2827/rulings; " +
		"CR 2026-11-06 115.1c, 601.2c, 602.1a, 602.2a-b, 120.1, 614.1a, 614.5, 614.9, 616.1",
}

func newNomadsEnKor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Nomads en-Kor",
			ManaCost:   opt.Val(cost.Mana{cost.W}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Kor, types.Nomad, types.Soldier},
			Colors:     []color.Color{color.White},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			OracleText: nomadsEnKorRedirectText,
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:           nomadsEnKorRedirectText,
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
