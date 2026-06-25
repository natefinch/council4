package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// explicitORingCardDef mirrors the two explicit triggered abilities the cardgen
// Shape A lowering produces for Oblivion Ring, Journey to Nowhere, and Fiend
// Hunter: a written-out enters trigger that exiles a target under the constant
// exile-until-leaves key, and a separate written-out leaves-the-battlefield
// trigger that returns the linked card under its owner's control. The link
// coordination publishes the same key on both halves, so the runtime binding is
// identical to the single-ability Shape B form.
func explicitORingCardDef() *game.CardDef {
	key := game.LinkedKey("exile-until-leaves")
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Banishing Cage",
		Types: []types.Card{types.Enchantment},
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

// noncreaturePermanentDef is an enchantment victim mirroring Oblivion Ring's
// "another target nonland permanent" reach beyond creatures.
func noncreaturePermanentDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Captured Enchantment",
		Types: []types.Card{types.Enchantment},
	}}
}

func TestExplicitORingReturnReleasesNonCreatureOnSourceLeaving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatPermanent(g, game.Player2, noncreaturePermanentDef())
	source := addCombatPermanent(g, game.Player1, explicitORingCardDef())

	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(victim.ObjectID)}
	resolveInstruction(engine, g, obj, game.Exile{
		Object:         game.TargetPermanentReference(0),
		ExileLinkedKey: game.LinkedKey("exile-until-leaves"),
	}, nil)

	if permanentByCardID(g, victim.CardInstanceID) != nil {
		t.Fatal("victim remained on the battlefield after the explicit enters exile")
	}
	if !g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatal("victim did not reach its owner's exile zone")
	}

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("explicit leaves-the-battlefield return trigger did not fire")
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
