package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// exileUntilMonarchCardDef mirrors the two abilities the cardgen lowering
// produces for Palace Jailer: an enters trigger that exiles a target under the
// monarch link key, and a synthesized "when an opponent becomes the monarch"
// trigger that returns the linked card.
func exileUntilMonarchCardDef() *game.CardDef {
	key := game.LinkedKey("exile-until-opponent-monarch")
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Warden of the Crown",
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
						Event:  game.EventBecameMonarch,
						Player: game.TriggerPlayerOpponent,
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
					Source: game.LinkedBattlefieldSource(key),
				}}}}.Ability(),
			},
		},
	}}
}

// TestExileUntilOpponentBecomesMonarchReturnsWhenOpponentTakesCrown models Palace
// Jailer: a creature exiled "until an opponent becomes the monarch" stays exiled
// while its controller holds the crown and returns to its owner's control once an
// opponent becomes the monarch.
func TestExileUntilOpponentBecomesMonarchReturnsWhenOpponentTakesCrown(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, exileUntilMonarchCardDef())
	setMonarch(g, game.Player1)

	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(victim.ObjectID)}
	resolveInstruction(engine, g, obj, game.Exile{
		Object:         game.TargetPermanentReference(0),
		ExileLinkedKey: game.LinkedKey("exile-until-opponent-monarch"),
	}, nil)

	if permanentByCardID(g, victim.CardInstanceID) != nil {
		t.Fatal("victim remained on the battlefield after exile-until-monarch")
	}
	if !g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatal("victim did not reach its owner's exile zone")
	}

	// The controller staying the monarch must not return the card.
	setMonarch(g, game.Player1)
	if engine.putTriggeredAbilitiesOnStack(g) {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	if permanentByCardID(g, victim.CardInstanceID) != nil {
		t.Fatal("victim returned before an opponent became the monarch")
	}

	// An opponent taking the crown returns the exiled card to its owner.
	setMonarch(g, game.Player2)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("become-monarch return trigger did not fire when an opponent took the crown")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	returned := permanentByCardID(g, victim.CardInstanceID)
	if returned == nil || returned.Controller != game.Player2 {
		t.Fatalf("returned permanent = %+v, want victim back under owner Player2 control", returned)
	}
	if g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatal("victim remained in exile after an opponent became the monarch")
	}
}
