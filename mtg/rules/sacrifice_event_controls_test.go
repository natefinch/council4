package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledSacrificeEventReturnRefusesUnavailablePayload(t *testing.T) {
	for _, corruption := range []string{"missing card", "missing version", "wrong card", "wrong version", "stale reentry"} {
		t.Run(corruption, func(t *testing.T) {
			def := compiledSacrificeCardReturn(t, " under your control")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCombatPermanent(g, game.Player2, def)
			victim := addCombatPermanent(g, game.Player3, vanillaCreature("Same Name", 2, 3))
			victim.Controller = game.Player1
			decoy := addCardToGraveyard(g, game.Player3, vanillaCreature("Same Name", 2, 3))
			if !sacrificePermanent(g, victim) {
				t.Fatal("sacrifice failed")
			}
			engine := NewEngine(nil)
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("natural trigger not queued")
			}
			obj, _ := g.Stack.Peek()
			switch corruption {
			case "missing card":
				obj.TriggerEvent.CardID = 0
			case "missing version":
				obj.TriggerEvent.CardZoneVersion = 0
			case "wrong card":
				obj.TriggerEvent.CardID = decoy
			case "wrong version":
				obj.TriggerEvent.CardZoneVersion++
			case "stale reentry":
				moveCardBetweenZones(g, game.Player3, victim.CardInstanceID, zone.Graveyard, zone.Exile)
				moveCardBetweenZones(g, game.Player3, victim.CardInstanceID, zone.Exile, zone.Graveyard)
			default:
				t.Fatal("unknown corruption")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if _, ok := findPermanentByCardID(g, victim.CardInstanceID); ok {
				t.Fatal("unavailable event-card incarnation returned")
			}
			if !g.Players[game.Player3].Graveyard.Contains(victim.CardInstanceID) ||
				!g.Players[game.Player3].Graveyard.Contains(decoy) {
				t.Fatal("failed exact reference consumed victim or same-name decoy")
			}
		})
	}
}

func TestCompiledSacrificeEventOpponentNontokenAndEventGates(t *testing.T) {
	for _, scenario := range []string{"opponent", "own controller", "token", "wrong event"} {
		t.Run(scenario, func(t *testing.T) {
			def := compiledSacrificeCardReturn(t, " under your control")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player2, def)
			victim := addCombatPermanent(g, game.Player3, vanillaCreature("Victim", 2, 3))
			victim.Controller = game.Player1
			if scenario == "own controller" {
				victim.Controller = game.Player2
			}
			if scenario == "token" {
				victim = &game.Permanent{ObjectID: g.IDGen.Next(), Owner: game.Player1,
					Controller: game.Player1, Token: true, TokenDef: vanillaCreature("Token", 1, 1)}
				g.Battlefield = append(g.Battlefield, victim)
			}
			if !sacrificePermanent(g, victim) {
				t.Fatal("actual sacrifice failed")
			}
			event := sacrificedCardEvent(t, g, victim)
			if scenario == "wrong event" {
				event.Kind = game.EventPermanentDied
				if triggerMatchesEvent(g, source, &def.TriggeredAbilities[0].Trigger.Pattern, event) {
					t.Fatal("death event acquired opponent-sacrifice ability")
				}
				return
			}
			engine := NewEngine(nil)
			queued := engine.putTriggeredAbilitiesOnStack(g)
			if queued != (scenario == "opponent") || g.Stack.Size() != map[bool]int{true: 1, false: 0}[queued] {
				t.Fatal("nontoken/opponent predicate broadened or duplicated")
			}
			if scenario == "token" && (event.CardID != 0 || event.CardZoneVersion != 0) {
				t.Fatal("token acquired invented physical card/version")
			}
		})
	}
}

func TestSacrificeEventBatchCaptureIncludesDepartingWatchers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := compiledSacrificeCardReturn(t, " under your control")
	source := addCombatPermanent(g, game.Player2, def)
	victim := addCombatPermanent(g, game.Player3, vanillaCreature("Victim", 2, 3))
	victim.Controller = game.Player1
	if !sacrificePermanentsSimultaneously(g, []*game.Permanent{source, victim}) {
		t.Fatal("batch sacrifice failed")
	}
	engine := NewEngine(nil)
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatal("simultaneously departing watcher lost opponent trigger or matched its own sacrifice")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	returned, ok := findPermanentByCardID(g, victim.CardInstanceID)
	if !ok || returned.Controller != game.Player2 || returned.Owner != game.Player3 ||
		!g.Players[game.Player2].Graveyard.Contains(source.CardInstanceID) {
		t.Fatal("departed watcher did not retain original body/controller or returned wrong card")
	}
}

func TestSacrificeEventIndependentWatchersDoNotReuseCardIncarnation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := compiledSacrificeCardReturn(t, " under your control")
	addCombatPermanent(g, game.Player2, def)
	addCombatPermanent(g, game.Player4, def)
	victim := addCombatPermanent(g, game.Player3, vanillaCreature("Victim", 2, 3))
	victim.Controller = game.Player1
	if !sacrificePermanent(g, victim) {
		t.Fatal("sacrifice failed")
	}
	engine := NewEngine(nil)
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
		t.Fatal("two independent controller-owned watchers did not each queue exactly once")
	}
	first, _ := g.Stack.Peek()
	winner := first.Controller
	engine.resolveTopOfStack(g, &TurnLog{})
	firstReturned := permanentForCard(g, victim.CardInstanceID)
	if firstReturned == nil || firstReturned.Controller != winner {
		t.Fatal("first exact trigger did not control actual returned card")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if returned := permanentForCard(g, victim.CardInstanceID); returned != firstReturned {
		t.Fatal("second stale event returned/rebound new physical incarnation")
	}
}

func TestPreventedSacrificeHasNoEventOrReturn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player2, compiledSacrificeCardReturn(t, " under your control"))
	victim := addCombatPermanent(g, game.Player1, vanillaCreature("Victim", 2, 3))
	cantSacrificeControlNotOwnEnchantment(g, game.Player1)
	victim.Owner = game.Player3
	g.CardInstances[victim.CardInstanceID].Owner = game.Player3
	r := newEffectResolver(NewEngine(nil), g, &game.StackObject{Controller: game.Player1},
		[game.NumPlayers]PlayerAgent{}, &TurnLog{})
	r.resolveInstruction(&game.Instruction{Primitive: game.Sacrifice{
		Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
	}})
	if _, live := permanentByObjectID(g, victim.ObjectID); !live {
		t.Fatal("protected permanent was sacrificed")
	}
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentSacrificed && event.PermanentID == victim.ObjectID {
			t.Fatal("prevented sacrifice invented a completed event")
		}
	}
	if NewEngine(nil).putTriggeredAbilitiesOnStack(g) {
		t.Fatal("prevented sacrifice acquired return ability")
	}
}

func TestSacrificeEventLookbackSourceRetainsPhysicalOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player3, compiledSacrificeCardReturn(t, " under your control"))
	source.Controller = game.Player1
	if !sacrificePermanent(g, source) {
		t.Fatal("source sacrifice failed")
	}
	event := sacrificedCardEvent(t, g, source)
	lookback, ok := leftBattlefieldTriggerSource(g, event)
	if !ok || lookback.Owner != game.Player3 || lookback.Controller != game.Player1 ||
		lookback.CardInstanceID != source.CardInstanceID || lookback.ObjectID != source.ObjectID {
		t.Fatal("sacrificing event player replaced departed source's physical owner/controller/identity")
	}
}

func TestCompiledSacrificeEventReturnReportsActualSuccess(t *testing.T) {
	for _, scenario := range []string{"entered", "stale", "replaced entry"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			def := compiledSacrificeCardReturn(t, " under your control")
			addCombatPermanent(g, game.Player2, def)
			victim := addCombatPermanent(g, game.Player3, vanillaCreature("Victim", 2, 3))
			victim.Controller = game.Player1
			if !sacrificePermanent(g, victim) {
				t.Fatal("natural sacrifice failed")
			}
			engine := NewEngine(nil)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("natural compiled trigger missing")
			}
			obj, _ := g.Stack.Peek()
			if scenario == "stale" {
				moveCardBetweenZones(g, game.Player3, victim.CardInstanceID, zone.Graveyard, zone.Exile)
			}
			if scenario == "replaced entry" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: zone.Graveyard,
					MatchToZone: true, ToZone: zone.Battlefield, ReplaceToZone: zone.Exile,
				})
			}
			put, ok := def.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
			if !ok {
				t.Fatal("compiled body lost battlefield primitive")
			}
			r := newEffectResolver(engine, g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			outcome := handlePutOnBattlefield(r, put)
			if !outcome.accepted || outcome.succeeded != (scenario == "entered") {
				t.Fatal("receipt conflated accepted instruction with actual successful entry")
			}
			entered := 0
			for _, event := range g.Events {
				if event.Kind == game.EventPermanentEnteredBattlefield && event.CardID == victim.CardInstanceID {
					entered++
				}
			}
			if entered != map[bool]int{true: 1, false: 0}[outcome.succeeded] {
				t.Fatal("failed/stale/replaced return invented entered subject")
			}
		})
	}
}
