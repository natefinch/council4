package rules

import (
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledTargetingEventPlayerSurvivesCauseDepartureAndCopy(t *testing.T) {
	def := legacyCard(t, carddata.ScryfallCard{
		Name: "Ashenmoor Liege", Layout: "normal", TypeLine: "Creature", Power: new("4"), Toughness: new("1"),
		OracleText: "Other black creatures you control get +1/+1.\nOther red creatures you control get +1/+1.\n" +
			"Whenever this creature becomes the target of a spell or ability an opponent controls, that player loses 4 life.",
	})
	for _, controller := range []game.PlayerID{game.Player1, game.Player3} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player2, def)
		source.Controller = game.Player1
		cause := addStackSpell(g, controller, "Causing spell", []types.Card{types.Instant})
		cause.Targets = []game.Target{game.PermanentTarget(source.ObjectID)}
		emitTargetEvents(g, cause)
		triggered := engine.putTriggeredAbilitiesOnStack(g)
		if controller == game.Player1 {
			if triggered {
				t.Fatal("source controller's own spell passed the opponent causing-controller qualifier")
			}
			continue
		}
		if !triggered {
			t.Fatal("natural target event did not trigger")
		}
		original, _ := g.Stack.Peek()
		if original.TriggerEvent.Controller != controller {
			t.Fatal("target event failed to capture causing controller")
		}
		resolveInstruction(engine, g, &game.StackObject{Controller: game.Player4, Targets: []game.Target{game.StackObjectTarget(original.ID)}},
			game.CopyStackObject{Object: game.TargetStackObjectReference(0)}, nil)
		copied, _ := g.Stack.Peek()
		if !copied.Copy || copied.Controller != game.Player4 || copied.TriggerEvent.Controller != controller {
			t.Fatal("real stack copy lost the original cause controller")
		}
		g.Stack.RemoveByID(cause.ID)
		resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID},
			game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Exile}, nil)
		engine.resolveTopOfStack(g, &TurnLog{})
		engine.resolveTopOfStack(g, &TurnLog{})
		for player, state := range g.Players {
			want := 40
			if game.PlayerID(player) == controller {
				want = 32
			}
			if state.Life != want {
				t.Fatalf("player%d life=%d, want%d; cause3 must differ from target controller1/owner2/copy4", player, state.Life, want)
			}
		}
	}
	if _, ok := triggeringEventPlayer(game.Event{Kind: game.EventPermanentEnteredBattlefield, Controller: game.Player3}); ok {
		t.Fatal("unrelated permanent-entry event was widened to Event.Controller")
	}
}

func TestUnblockedLifeLossDoesNotSuppressCombatAssignment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 6)
	g.CardInstances[attacker.CardInstanceID].Def.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
			Event: game.EventAttackerBecameUnblocked, Source: game.TriggerSourceSelf,
		}},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.LoseLife{
			Amount: game.Fixed(2), Player: game.DefendingPlayerReference(),
		}}}}.Ability(),
	}}
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	engine := NewEngine(nil)
	emitUnblockedAttackerEvents(g)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("natural unblocked-attack event did not trigger life loss")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Life != 38 {
		t.Fatal("unblocked-attack life loss did not happen")
	}
	t.Log("ordinary unblocked LoseLife resolves 40 -> 38")
	log := &TurnLog{}
	engine.resolveCombatDamage(g, log)
	if g.Players[game.Player2].Life != 32 || len(log.CombatDamage) != 1 {
		t.Fatal("ordinary life loss was incorrectly treated as a no-combat-assignment replacement")
	}
	t.Log("without assignment suppression, six combat damage resolves 38 -> 32")
}
