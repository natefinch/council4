package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const worldgorgerOracleID = "a628186d-b7d9-40a5-9ae2-fbc9d2a14c7c"
const worldgorgerExiles game.LinkedProducerKey = worldgorgerOracleID + "/enter/exiles"

var worldgorgerEnter = game.ProgramReference{ID: worldgorgerOracleID + "/enter/0", Version: "1"}
var worldgorgerLeave = game.ProgramReference{ID: worldgorgerOracleID + "/leave/0", Version: "1"}

// WorldgorgerDragon uses CR 607 links, not an exile-until duration.
var WorldgorgerDragon = cardset.Entry{
	OracleID: worldgorgerOracleID, Name: "Worldgorger Dragon", New: newWorldgorgerDragon,
	Verification: "mtg/rules/authored_lifetime_test.go: TestWorldgorgerProductionSimultaneousExileOwnerReturnAndClone; TestWorldgorgerLeavesBeforeEnterIsNotAnUntilEffect",
	Programs: []cardprogram.Program{
		{Reference: worldgorgerEnter, Resolve: resolveWorldgorgerEnter},
		{Reference: worldgorgerLeave, Resolve: resolveWorldgorgerLeave},
	},
}

func newWorldgorgerDragon() *game.CardDef {
	enter := game.Mode{Text: "Exile other controlled permanents linked to this incarnation", Program: worldgorgerEnter}.Ability()
	enter.LinkedProducers = []game.LinkedProducerKey{worldgorgerExiles}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Worldgorger Dragon", ManaCost: opt.Val(cost.Mana{cost.O(3), cost.R, cost.R, cost.R}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Nightmare, types.Dragon},
			Colors: []color.Color{color.Red}, Power: opt.Val(game.PT{Value: 7}), Toughness: opt.Val(game.PT{Value: 7}),
			OracleText:      "Flying, trample\nWhen this creature enters, exile all other permanents you control.\nWhen this creature leaves the battlefield, return the exiled cards to the battlefield under their owners' control.",
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody, game.TrampleStaticBody},
			TriggeredAbilities: []game.TriggeredAbility{
				{Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
					Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
				}}, Content: enter},
				{Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
					Event: game.EventZoneChanged, Source: game.TriggerSourceSelf,
					MatchFromZone: true, FromZone: zone.Battlefield,
				}}, Content: game.Mode{Text: "Return linked exiled cards under owner control", Program: worldgorgerLeave}.Ability()},
			},
		},
	}
}

func resolveWorldgorgerEnter(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.MovePermanent{
		Group:       game.BattlefieldGroupExcluding(game.Selection{Controller: game.ControllerYou}, game.SourcePermanentReference()),
		Destination: zone.Exile, PublishLinkedProducer: worldgorgerExiles,
	})
	return err
}

func resolveWorldgorgerLeave(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.PutOnBattlefield{
		Source: game.OriginalSourceLinkedBattlefieldSource(worldgorgerExiles),
	})
	return err
}
