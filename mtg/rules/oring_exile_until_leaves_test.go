package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// oringCardDef mirrors the two triggered abilities the cardgen O-Ring lowering
// produces: an enters trigger that exiles a target under a constant linked key,
// and a leaves-the-battlefield trigger that returns the linked card.
func oringCardDef() *game.CardDef {
	key := game.LinkedKey("exile-until-leaves")
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Prison Warden",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentEnteredBattlefield,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.Mode{
					Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent}},
					Sequence: []game.Instruction{{Primitive: game.Exile{
						Object:         game.TargetPermanentReference(0),
						ExileLinkedKey: key,
					}}},
				}.Ability(),
			},
			{
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:         game.EventZoneChanged,
						Source:        game.TriggerSourceSelf,
						MatchFromZone: true,
						FromZone:      zone.Battlefield,
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
					Source: game.LinkedBattlefieldSource(key),
				}}}}.Ability(),
			},
		},
	}}
}

func TestORingExileUntilLeavesReturnsOnSourceLeaving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, oringCardDef())

	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(victim.ObjectID)}
	resolveInstruction(engine, g, obj, game.Exile{
		Object:         game.TargetPermanentReference(0),
		ExileLinkedKey: game.LinkedKey("exile-until-leaves"),
	}, nil)

	if permanentByCardID(g, victim.CardInstanceID) != nil {
		t.Fatal("victim remained on the battlefield after exile-until-leaves")
	}
	if !g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatal("victim did not reach its owner's exile zone")
	}

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("leaves-the-battlefield return trigger did not fire")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	returned := permanentByCardID(g, victim.CardInstanceID)
	if returned == nil || returned.Controller != game.Player2 {
		t.Fatalf("returned permanent = %+v, want victim back under owner Player2 control", returned)
	}
	if g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatal("victim remained in exile after the source left the battlefield")
	}
}

func TestORingExileDoesNotReturnWhileSourceRemains(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, oringCardDef())

	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(victim.ObjectID)}
	resolveInstruction(engine, g, obj, game.Exile{
		Object:         game.TargetPermanentReference(0),
		ExileLinkedKey: game.LinkedKey("exile-until-leaves"),
	}, nil)

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("return trigger fired while the source was still on the battlefield")
	}
	if !g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatal("victim left exile without the source leaving the battlefield")
	}
}
