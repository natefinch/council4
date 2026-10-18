package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func toothCollectorTestGame(t *testing.T) (*game.Game, *Engine, *game.CardDef) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(authored.ToothCollector.OracleID)
	if err != nil {
		t.Fatal(err)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	return g, NewEngineWithPrograms(nil, registry.Programs()), def
}

func toothCollectorTestCreature(name string, power, toughness int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: power}), Toughness: opt.Val(game.PT{Value: toughness}),
	}}
}

func addToothCollectorDeliriumCards(g *game.Game, includeSorcery bool) id.ID {
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Relic", Supertypes: []types.Super{types.Legendary},
		Types: []types.Card{types.Artifact, types.Creature}, Subtypes: []types.Sub{types.Human},
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Instant One", Types: []types.Card{types.Instant},
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Instant Two", Types: []types.Card{types.Instant},
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{
		Layout: game.LayoutTransform,
		CardFace: game.CardFace{
			Name: "Front-Face Creature", Types: []types.Card{types.Creature},
		},
		Back: opt.Val(game.CardFace{
			Name: "Back-Face Enchantment", Types: []types.Card{types.Enchantment},
		}),
	})
	if !includeSorcery {
		return 0
	}
	return addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Sorcery", Types: []types.Card{types.Sorcery},
	}})
}

func emitToothCollectorOpponentUpkeep(g *game.Game) {
	g.Turn.ActivePlayer = game.Player2
	emitEvent(g, game.Event{
		Kind: game.EventBeginningOfStep, Step: game.StepUpkeep,
		Player: game.Player2, Controller: game.Player2,
	})
}

func TestToothCollectorEnterTriggerAppliesMinusOneMinusOneToOpponentCreature(t *testing.T) {
	g, engine, def := toothCollectorTestGame(t)
	ownCreature := addCombatPermanent(g, game.Player1, toothCollectorTestCreature("Own Creature", 2, 2))
	opponentCreature := addCombatPermanent(g, game.Player2, toothCollectorTestCreature("Opponent Creature", 1, 1))
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("Tooth Collector could not be cast")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	source := permanentByCardID(g, cardID)
	if source == nil {
		t.Fatal("Tooth Collector did not enter")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Tooth Collector's enter trigger did not reach the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || len(trigger.Targets) != 1 || trigger.Targets[0].PermanentID != opponentCreature.ObjectID {
		t.Fatalf("enter-trigger targets = %+v, want the opponent's creature", trigger)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyStateBasedActions(g)

	if _, ok := permanentByObjectID(g, opponentCreature.ObjectID); ok {
		t.Fatal("opponent's 1/1 survived the -1/-1 effect and state-based actions")
	}
	if _, ok := permanentByObjectID(g, ownCreature.ObjectID); !ok {
		t.Fatal("enter trigger affected its controller's creature")
	}
}

func TestToothCollectorDeliriumRequiresFourDistinctCardTypes(t *testing.T) {
	t.Run("three types do not trigger despite extra cards and subtypes", func(t *testing.T) {
		g, engine, def := toothCollectorTestGame(t)
		addCombatPermanent(g, game.Player1, def)
		addToothCollectorDeliriumCards(g, false)
		addCombatPermanent(g, game.Player2, toothCollectorTestCreature("Opponent Creature", 1, 1))
		emitToothCollectorOpponentUpkeep(g)

		if engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("delirium triggered with only artifact, creature, and instant card types")
		}
	})

	t.Run("four types trigger and resolve", func(t *testing.T) {
		g, engine, def := toothCollectorTestGame(t)
		addCombatPermanent(g, game.Player1, def)
		addToothCollectorDeliriumCards(g, true)
		target := addCombatPermanent(g, game.Player2, toothCollectorTestCreature("Opponent Creature", 1, 1))
		emitToothCollectorOpponentUpkeep(g)

		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("delirium did not trigger with four distinct card types")
		}
		trigger, ok := g.Stack.Peek()
		if !ok || len(trigger.Targets) != 1 || trigger.Targets[0].PermanentID != target.ObjectID {
			t.Fatalf("delirium-trigger targets = %+v, want the opponent's creature", trigger)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		engine.applyStateBasedActions(g)
		if _, ok := permanentByObjectID(g, target.ObjectID); ok {
			t.Fatal("four-type delirium effect did not give its 1/1 target -1/-1")
		}
	})
}

func TestToothCollectorDeliriumRechecksAtResolution(t *testing.T) {
	g, engine, def := toothCollectorTestGame(t)
	addCombatPermanent(g, game.Player1, def)
	sorceryID := addToothCollectorDeliriumCards(g, true)
	target := addCombatPermanent(g, game.Player2, toothCollectorTestCreature("Opponent Creature", 1, 1))
	emitToothCollectorOpponentUpkeep(g)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("delirium did not trigger with four distinct card types")
	}
	moveCardBetweenZones(g, game.Player1, sorceryID, zone.Graveyard, zone.Hand)
	log := &TurnLog{}
	engine.resolveTopOfStack(g, log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "intervening if false" {
		t.Fatalf("delirium resolve = %+v, want intervening-if false", log.Resolves)
	}
	if _, ok := permanentByObjectID(g, target.ObjectID); !ok {
		t.Fatal("delirium effect resolved after its intervening-if condition became false")
	}
}

func TestToothCollectorDeliriumTargetsEventPlayerAndResolvesAfterSourceLeaves(t *testing.T) {
	g, engine, def := toothCollectorTestGame(t)
	source := addCombatPermanent(g, game.Player1, def)
	addToothCollectorDeliriumCards(g, true)
	eventPlayerCreature := addCombatPermanent(g, game.Player2, toothCollectorTestCreature("Upkeep Creature", 1, 1))
	bystanderCreature := addCombatPermanent(g, game.Player3, toothCollectorTestCreature("Bystander Creature", 2, 2))
	emitToothCollectorOpponentUpkeep(g)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("delirium did not trigger on the opponent's upkeep")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || len(trigger.Targets) != 1 || trigger.Targets[0].PermanentID != eventPlayerCreature.ObjectID {
		t.Fatalf("delirium-trigger targets = %+v, want only the upkeep player's creature", trigger)
	}
	movePermanentToZone(g, source, zone.Graveyard)
	log := &TurnLog{}
	engine.resolveTopOfStack(g, log)
	engine.applyStateBasedActions(g)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "resolved" {
		t.Fatalf("delirium resolve after source departure = %+v, want resolved", log.Resolves)
	}
	if _, ok := permanentByObjectID(g, eventPlayerCreature.ObjectID); ok {
		t.Fatal("triggered ability did not resolve after its source left")
	}
	if _, ok := permanentByObjectID(g, bystanderCreature.ObjectID); !ok {
		t.Fatal("delirium ability affected a creature not controlled by the upkeep player")
	}
}

func TestToothCollectorDeliriumTargetMustRemainControlledByEventPlayer(t *testing.T) {
	g, engine, def := toothCollectorTestGame(t)
	addCombatPermanent(g, game.Player1, def)
	addToothCollectorDeliriumCards(g, true)
	target := addCombatPermanent(g, game.Player2, toothCollectorTestCreature("Upkeep Creature", 1, 1))
	bystander := addCombatPermanent(g, game.Player3, toothCollectorTestCreature("Bystander Creature", 1, 1))
	emitToothCollectorOpponentUpkeep(g)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("delirium did not trigger on the opponent's upkeep")
	}

	trigger, ok := g.Stack.Peek()
	if !ok || len(trigger.Targets) != 1 || trigger.Targets[0].PermanentID != target.ObjectID {
		t.Fatalf("delirium-trigger targets = %+v, want the upkeep player's creature", trigger)
	}
	target.Controller = game.Player3
	log := &TurnLog{}
	engine.resolveTopOfStack(g, log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "countered by rules" {
		t.Fatalf("delirium resolve after control changed = %+v, want countered by rules", log.Resolves)
	}
	if _, ok := permanentByObjectID(g, target.ObjectID); !ok {
		t.Fatal("trigger affected a target that no longer met its event-player restriction")
	}
	if _, ok := permanentByObjectID(g, bystander.ObjectID); !ok {
		t.Fatal("trigger affected a non-target creature")
	}
}

func TestToothCollectorBothAbilitiesExpireAtCleanup(t *testing.T) {
	for _, test := range []struct {
		name   string
		upkeep bool
	}{
		{name: "enter"},
		{name: "upkeep", upkeep: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			g, engine, def := toothCollectorTestGame(t)
			target := addCombatPermanent(g, game.Player2, toothCollectorTestCreature("Surviving target", 3, 3))
			if test.upkeep {
				addCombatPermanent(g, game.Player1, def)
				addToothCollectorDeliriumCards(g, true)
				emitToothCollectorOpponentUpkeep(g)
			} else {
				cardID := addCardToHand(g, game.Player1, def)
				g.Players[game.Player1].ManaPool.Add(mana.C, 2)
				g.Players[game.Player1].ManaPool.Add(mana.B, 1)
				if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
					t.Fatal("Tooth Collector could not be cast")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
			}
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("Tooth Collector ability did not trigger")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if toughness, ok := effectiveToughness(g, target); !ok || toughness != 2 || effectivePower(g, target) != 2 {
				t.Fatal("target did not receive exactly -1/-1 before cleanup")
			}
			engine.runCleanupStep(g, simPassPolicies(), &TurnLog{})
			if toughness, ok := effectiveToughness(g, target); !ok || toughness != 3 || effectivePower(g, target) != 3 {
				t.Fatal("Tooth Collector's -1/-1 did not expire at cleanup")
			}
		})
	}
}
