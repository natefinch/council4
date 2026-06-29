package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// championCardDef mirrors the cardgen Champion lowering: an enters trigger that
// exiles another creature the controller owns under the exile-until-leaves key,
// plus the synthesized leaves-the-battlefield return.
func championCardDef() *game.CardDef {
	key := game.LinkedKey("exile-until-leaves")
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Champion Caller",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Trigger: game.TriggerCondition{
					Type:    game.TriggerWhen,
					Pattern: game.TriggerPattern{Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf},
				},
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.ChampionExile{
					Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true},
					LinkedKey: key,
				}}}}.Ability(),
			},
			{
				Trigger: game.TriggerCondition{
					Type:    game.TriggerWhen,
					Pattern: game.TriggerPattern{Event: game.EventZoneChanged, Source: game.TriggerSourceSelf, MatchFromZone: true, FromZone: zone.Battlefield},
				},
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
					Source: game.LinkedBattlefieldSource(key),
				}}}}.Ability(),
			},
		},
	}}
}

func TestChampionExilesAnotherCreatureAndReturnsOnLeave(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	championed := addCombatCreaturePermanent(g, game.Player1)
	source := addCombatPermanent(g, game.Player1, championCardDef())

	obj := linkedSourceObject(source)
	resolveInstruction(engine, g, obj, game.ChampionExile{
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true},
		LinkedKey: game.LinkedKey("exile-until-leaves"),
	}, nil)

	if permanentByCardID(g, championed.CardInstanceID) != nil {
		t.Fatal("championed creature stayed on the battlefield")
	}
	if !g.Players[game.Player1].Exile.Contains(championed.CardInstanceID) {
		t.Fatal("championed creature did not reach exile")
	}

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("leaves-the-battlefield return trigger did not fire")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if permanentByCardID(g, championed.CardInstanceID) == nil {
		t.Fatal("championed creature did not return when the source left")
	}
}

func TestChampionSacrificesSourceWhenNoEligibleCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, championCardDef())

	obj := linkedSourceObject(source)
	resolveInstruction(engine, g, obj, game.ChampionExile{
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true},
		LinkedKey: game.LinkedKey("exile-until-leaves"),
	}, nil)

	if permanentByCardID(g, source.CardInstanceID) != nil {
		t.Fatal("source remained despite no creature to champion")
	}
}
