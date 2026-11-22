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

const outriderEnKorOracleID = "644b7aba-a7b8-4861-9e83-73329cf85be2"

const outriderEnKorRedirectText = "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead."

var OutriderEnKor = cardset.Entry{
	OracleID: outriderEnKorOracleID,
	Name:     "Outrider en-Kor",
	New:      newOutriderEnKor,
	Verification: "mtg/rules/outrider_en_kor_test.go: TestOutriderEnKorFlankingAndRedirect; " +
		"printing de9b2506-ead4-49ff-a733-66f3f13dbe17; Scryfall rulings URI https://api.scryfall.com/cards/de9b2506-ead4-49ff-a733-66f3f13dbe17/rulings; " +
		"CR 2026-11-06 601.2c, 602.1a, 602.2a-b, 509.3d, 614.1a, 614.9, 702.25a",
}

func newOutriderEnKor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:      "Outrider en-Kor",
			ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.W}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kor, types.Rebel, types.Knight},
			Colors:    []color.Color{color.White},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			OracleText: "Flanking (Whenever a creature without flanking blocks this creature, the blocking creature gets -1/-1 until end of turn.)\n" +
				outriderEnKorRedirectText,
			TriggeredAbilities: []game.TriggeredAbility{
				game.FlankingTriggeredBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:           outriderEnKorRedirectText,
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
