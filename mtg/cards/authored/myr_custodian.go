package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var MyrCustodian = cardset.Entry{
	OracleID: "43923f24-7e95-4934-9a5c-8d9a5e06ddea",
	Name:     "Myr Custodian",
	Verification: "mtg/rules/myr_custodian_test.go: " +
		"TestMyrCustodianPaidCastAndPrivateScry, " +
		"TestMyrCustodianControllerAndAPNAP, " +
		"TestMyrCustodianSourceDepartureCopiedTriggerAndClone, " +
		"TestMyrCustodianEmptyLibrariesAndDepartedOpponent",
	New: newMyrCustodian,
}

func newMyrCustodian() *game.CardDef {
	text := "When this creature enters, scry 2. Then each opponent may scry 1. (To scry X, a player looks at the top X cards of their library, then puts any number of them on the bottom and the rest on top in any order.)"
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Myr Custodian", OracleText: text,
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
		Types:    []types.Card{types.Artifact, types.Creature},
		Subtypes: []types.Sub{types.Myr},
		Power:    opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 3}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: text,
			Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
				Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
			}},
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.Scry{Player: game.ControllerReference(), Amount: game.Fixed(2)}},
				{
					Primitive: game.Scry{Player: game.GroupOfferMemberReference(), Amount: game.Fixed(1)},
					Optional:  true, OptionalActorGroup: opt.Val(game.OpponentsReference()),
				},
			}}.Ability(),
		}},
	}}
}
