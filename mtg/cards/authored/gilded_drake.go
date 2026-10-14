package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const gildedDrakeOracleID = "7f06c098-6482-4bf3-a9a1-110d6d5b5703"

var gildedDrakeExchangeProgram = game.ProgramReference{ID: gildedDrakeOracleID + "/enter/0", Version: "1"}

// GildedDrake declares its resolution exception rather than a resolver name check.
var GildedDrake = cardset.Entry{
	OracleID: gildedDrakeOracleID, Name: "Gilded Drake", New: newGildedDrake,
	Programs: []cardprogram.Program{{Reference: gildedDrakeExchangeProgram, Resolve: resolveGildedDrakeExchange}},
}

func newGildedDrake() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Gilded Drake", ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Drake}, Colors: []color.Color{color.Blue},
			Power: opt.Val(game.PT{Value: 3}), Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
			OracleText:      "Flying\nWhen this creature enters, exchange control of this creature and up to one target creature an opponent controls. If you don't or can't make an exchange, sacrifice this creature. This ability still resolves if its target becomes illegal.",
			TriggeredAbilities: []game.TriggeredAbility{{
				Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
					Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
				}},
				ResolveWithIllegalTargets: true,
				Content: game.Mode{
					Targets: []game.TargetSpec{{
						MinTargets: 0, MaxTargets: 1, Allow: game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
					}},
					Program: gildedDrakeExchangeProgram,
				}.Ability(),
			}},
		},
	}
}

func resolveGildedDrakeExchange(ctx cardprogram.Context) error {
	outcome, err := ctx.Exchange(game.SourcePermanentReference(), game.TargetPermanentReference(0))
	if err != nil {
		return err
	}
	if outcome.Exchanged {
		return nil
	}
	_, err = ctx.Execute(game.Sacrifice{Object: game.SourcePermanentReference()})
	return err
}
