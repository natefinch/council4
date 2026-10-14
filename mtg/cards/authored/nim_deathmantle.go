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

const nimDeathmantleOracleID = "66d41377-626d-4ae6-ba86-17bf0c8b3362"

var nimDeathmantleReturn = game.ProgramReference{ID: nimDeathmantleOracleID + "/death/0", Version: "1"}

// NimDeathmantle keeps payment and attachment in the shared Engine.
var NimDeathmantle = cardset.Entry{
	OracleID: nimDeathmantleOracleID, Name: "Nim Deathmantle", New: newNimDeathmantle,
	Programs: []cardprogram.Program{{Reference: nimDeathmantleReturn, Resolve: resolveNimDeathmantle}},
}

func newNimDeathmantle() *game.CardDef {
	attached := game.AttachedObjectGroup(game.SourcePermanentReference())
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Nim Deathmantle", ManaCost: opt.Val(cost.Mana{cost.O(2)}),
		Types: []types.Card{types.Artifact}, Subtypes: []types.Sub{types.Equipment},
		OracleText: "Equipped creature gets +2/+2, has intimidate, and is a black Zombie. (A creature with intimidate can't be blocked except by artifact creatures and/or creatures that share a color with it.)\nWhenever a nontoken creature is put into your graveyard from the battlefield, you may pay {4}. If you do, return that card to the battlefield and attach this Equipment to it.\nEquip {4}",
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{
			{Layer: game.LayerPowerToughnessModify, Group: attached, PowerDelta: 2, ToughnessDelta: 2},
			{Layer: game.LayerAbility, Group: attached, AddKeywords: []game.Keyword{game.Intimidate}},
			{Layer: game.LayerColor, Group: attached, SetColors: []color.Color{color.Black}},
			{Layer: game.LayerType, Group: attached, RemoveSubtypes: types.SubtypesForType(types.Creature), AddSubtypes: []types.Sub{types.Zombie}},
		}}},
		ActivatedAbilities: []game.ActivatedAbility{game.EquipActivatedAbility(cost.Mana{cost.O(4)})},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
				Event: game.EventZoneChanged, Player: game.TriggerPlayerYou,
				MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Graveyard,
				RequirePermanentTypes: []types.Card{types.Creature}, RequireNonToken: true,
			}},
			Content: game.Mode{Text: "Optional payment to return and attach to the exact departed creature", Program: nimDeathmantleReturn}.Ability(),
		}},
	}}
}

func resolveNimDeathmantle(ctx cardprogram.Context) error {
	paid, err := ctx.Execute(game.Pay{Payment: game.ResolutionPayment{
		Prompt:   "Pay {4} to return the creature and attach Nim Deathmantle?",
		ManaCost: opt.Val(cost.Mana{cost.O(4)}),
	}})
	if err != nil || !paid.Succeeded {
		return err
	}
	const arrived game.LinkedKey = nimDeathmantleOracleID + "/returned"
	returned, err := ctx.Execute(game.PutOnBattlefield{
		Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
		PublishLinked: arrived,
	})
	if err != nil || !returned.Succeeded {
		return err
	}
	_, err = ctx.Execute(game.Attach{
		Attachment: game.SourcePermanentReference(), Target: game.LinkedObjectReference(string(arrived)),
	})
	return err
}
