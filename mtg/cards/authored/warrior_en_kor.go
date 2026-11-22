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

const warriorEnKorOracleID = "1245b165-e26f-4839-8468-2e95148fc6f6"

const warriorEnKorText = "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead."

var WarriorEnKor = cardset.Entry{
	OracleID: warriorEnKorOracleID,
	Name:     "Warrior en-Kor",
	New:      newWarriorEnKor,
	Verification: "mtg/rules/warrior_en_kor_test.go: TestWarriorEnKorRedirectsOneDamageToTargetCreature; " +
		"printing afcf1e51-5e9a-4c86-8a14-97d9e919c320; Scryfall rulings URI https://api.scryfall.com/cards/afcf1e51-5e9a-4c86-8a14-97d9e919c320/rulings; " +
		"CR 2026-11-06 115.1c, 602.1a, 602.2a-b, 614.9",
}

func newWarriorEnKor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Warrior en-Kor",
			ManaCost:   opt.Val(cost.Mana{cost.W, cost.W}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Kor, types.Warrior, types.Knight},
			Colors:     []color.Color{color.White},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			OracleText: warriorEnKorText,
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:           warriorEnKorText,
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
