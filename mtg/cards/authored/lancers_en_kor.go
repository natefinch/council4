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

const lancersEnKorOracleID = "ef49dc78-9fd9-4cf8-af10-5a6b8ef7fc57"

const lancersEnKorRedirectText = "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead."

var LancersEnKor = cardset.Entry{
	OracleID: lancersEnKorOracleID,
	Name:     "Lancers en-Kor",
	New:      newLancersEnKor,
	Verification: "mtg/rules/lancers_en_kor_test.go: TestLancersEnKorTramplesAndRedirectsOneCombatDamage; " +
		"printing e7cd99c2-6d4a-48e8-8848-6fdc0788525d; Scryfall rulings URI https://api.scryfall.com/cards/e7cd99c2-6d4a-48e8-8848-6fdc0788525d/rulings; " +
		"CR 2026-11-06 115.1c, 601.2c, 602.1a, 602.2a-b, 510.1a-d, 510.2, 120.1, 120.2a, 120.4b, 614.1a, 614.5, 614.9, 616.1, 702.19a-b",
}

func newLancersEnKor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Lancers en-Kor",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.W, cost.W}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Kor, types.Soldier},
			Colors:     []color.Color{color.White},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			OracleText: "Trample\n" + lancersEnKorRedirectText,
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:           lancersEnKorRedirectText,
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
