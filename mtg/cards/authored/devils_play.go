package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const devilsPlayOracleID = "ee9f0b29-8a54-4cb8-8e2c-7bd67c2184ba"
const devilsPlayOracleText = "Devil's Play deals X damage to any target.\nFlashback {X}{R}{R}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)"

var DevilsPlay = cardset.Entry{
	OracleID: devilsPlayOracleID,
	Name:     "Devil's Play",
	New:      newDevilsPlay,
	Verification: "mtg/rules/authored_existing_paragraph_cards_test.go: " +
		"TestDevilsPlayScenario",
}

func newDevilsPlay() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:     "Devil's Play",
			ManaCost: opt.Val(cost.Mana{cost.X, cost.R}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{{
				Text: "Flashback {X}{R}{R}{R}",
				KeywordAbilities: []game.KeywordAbility{
					game.FlashbackKeyword{Cost: cost.Mana{cost.X, cost.R, cost.R, cost.R}},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "any target",
					Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle},
					}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.Damage{
						Amount: game.Dynamic(game.DynamicAmount{
							Kind: game.DynamicAmountX,
						}),
						Recipient: game.AnyTargetDamageRecipient(0),
					},
				}},
			}.Ability()),
			OracleText: devilsPlayOracleText,
		},
	}
}
