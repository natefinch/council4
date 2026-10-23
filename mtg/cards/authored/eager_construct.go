package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const eagerConstructOracleID = "39363190-7f79-4fae-bcfc-b17ec4e07bfe"

var EagerConstruct = cardset.Entry{
	OracleID: eagerConstructOracleID,
	Name:     "Eager Construct",
	New:      newEagerConstruct,
	Verification: "mtg/rules/scry_interaction_test.go: " +
		"TestEagerConstructDefaultRegistryPaidCastAndAPNAP; " +
		"TestEagerConstructEmptyLibrariesAndDepartedPlayers; " +
		"TestEagerConstructSourceDepartureCopyCloneAndPrivateReplay. " +
		"Sources: pinned oracle-cards.json SHA256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; " +
		"2XM 250 printing 5b23ca61-157b-4616-bbf4-0c1298ddbbb4; full " +
		"https://api.scryfall.com/cards/5b23ca61-157b-4616-bbf4-0c1298ddbbb4/rulings " +
		"retrieved 2026-10-22 and retained in the rules test source record. Official CR " +
		"https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt, " +
		"effective 2026-09-25, SHA256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca: " +
		"101.4a-c and 701.22a-d require APNAP choices, private same-time looks, and simultaneous placements; " +
		"401.2-4 preserve hidden identities and permit public counts; 603.2/3a and 113.7a keep the ETB trigger " +
		"and its controller after source departure; 707.10 makes a copy independent without copying later choices.",
}

func newEagerConstruct() *game.CardDef {
	const text = "When this creature enters, each player may scry 1. (To scry 1, look at the top card of your library, then you may put that card on the bottom.)"
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Eager Construct",
		OracleText: text,
		ManaCost:   opt.Val(cost.Mana{cost.O(2)}),
		Types:      []types.Card{types.Artifact, types.Creature},
		Subtypes:   []types.Sub{types.Construct},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: text,
			Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
				Event:  game.EventPermanentEnteredBattlefield,
				Source: game.TriggerSourceSelf,
			}},
			Content: game.Mode{Text: text, Sequence: []game.Instruction{{
				Primitive: game.Scry{
					Player: game.GroupOfferMemberReference(),
					Amount: game.Fixed(1),
				},
				Optional:           true,
				OptionalActorGroup: opt.Val(game.AllPlayersReference()),
			}}}.Ability(),
		}},
	}}
}
