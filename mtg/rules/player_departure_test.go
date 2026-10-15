package rules

import (
	"testing"

	cardc "github.com/natefinch/council4/mtg/cards/c"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestPlayerDepartureNonownedDefaultControlIsExiled(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Controller = game.Player2
	branch := g.Clone()
	eliminatePlayer(branch, game.Player2)
	if _, present := permanentByObjectID(branch, creature.ObjectID); present ||
		!branch.Players[game.Player1].Exile.Contains(creature.CardInstanceID) {
		t.Fatal("departing default controller's nonowned creature was not exiled")
	}
	if _, present := permanentByObjectID(g, creature.ObjectID); !present || g.Players[game.Player2].Eliminated {
		t.Fatal("departure in a clone mutated the original game")
	}
}

func TestPlayerDepartureControlEffectsRevealSurvivingController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Controller = game.Player3
	for _, controller := range []game.PlayerID{game.Player4, game.Player2} {
		effectID := g.IDGen.Next()
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID: effectID, Timestamp: game.Timestamp(effectID), Layer: game.LayerControl,
			AffectedObjectID: creature.ObjectID, NewController: opt.Val(controller),
		})
	}
	eliminatePlayer(g, game.Player2)
	if effectiveController(g, creature) != game.Player4 || creature.Controller != game.Player3 {
		t.Fatal("departure reset default control instead of revealing the surviving control effect")
	}
	eliminatePlayer(g, game.Player4)
	if effectiveController(g, creature) != game.Player3 {
		t.Fatal("the final theft did not return the creature to its default controller")
	}
}

func TestPlayerDepartureOwnedObjectsLeaveWithoutOrdinaryExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	owned := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	owned.Controller = game.Player1
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Owner: game.Player2, Def: vanillaCreature("Departing hand", 2, 2)}
	g.Players[game.Player2].Hand.Add(cardID)
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), SourceID: cardID, Controller: game.Player1})
	eliminatePlayer(g, game.Player2)
	if _, present := permanentByObjectID(g, owned.ObjectID); present {
		t.Fatal("owned permanent controlled by a survivor stayed in play")
	}
	if g.Players[game.Player2].Hand.Size() != 0 || g.Players[game.Player2].Exile.Size() != 0 || !g.Stack.IsEmpty() {
		t.Fatal("departing player's owned objects remained in live zones or were ordinarily exiled")
	}
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && (event.CardID == owned.CardInstanceID || event.CardID == cardID) {
			t.Fatal("leaving the game fabricated an ordinary zone change")
		}
	}
}

func TestPlayerDepartureForeignCardSpellIsExiledButCopyCeases(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Owner: game.Player1, Def: vanillaCreature("Foreign spell", 2, 2)}
	for _, copySpell := range []bool{false, true} {
		g.Stack.Push(&game.StackObject{
			ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: cardID,
			Controller: game.Player2, Copy: copySpell,
		})
	}
	g.Stack.Push(&game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, SourceCardID: cardID, Controller: game.Player2,
	})
	eliminatePlayer(g, game.Player2)
	if !g.Stack.IsEmpty() || !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("foreign card spell was dropped instead of exiled after non-card stack objects ceased")
	}
	count := 0
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.CardID == cardID &&
			event.FromZone == zone.Stack && event.ToZone == zone.Exile {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("foreign spell exile facts = %d, want exactly one", count)
	}
}

func TestPlayerDepartureDefaultControllerAlreadyLeftExilesOnExpiry(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Controller = game.Player2
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Layer: game.LayerControl, AffectedObjectID: creature.ObjectID,
		NewController: opt.Val(game.Player3), Duration: game.DurationUntilEndOfTurn,
	})
	eliminatePlayer(g, game.Player2)
	if _, present := permanentByObjectID(g, creature.ObjectID); !present {
		t.Fatal("departure exiled an object still controlled by a surviving player")
	}
	expireCleanupDurations(g)
	if _, present := permanentByObjectID(g, creature.ObjectID); present ||
		!g.Players[game.Player1].Exile.Contains(creature.CardInstanceID) {
		t.Fatal("control expiry did not immediately exile the departed-default object")
	}
}

func TestPlayerDepartureReanimationThroughLiveAndSimulatorPriority(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	if !movePermanentToZone(g, original, zone.Graveyard) {
		t.Fatal("could not prepare reanimation target")
	}
	addReanimateSpell(g, currentCardTarget(t, g, original.CardInstanceID), false)
	engine.resolveTopOfStack(g, &TurnLog{})
	reanimated, ok := reanimatedPermanent(g, original.CardInstanceID)
	if !ok || reanimated.Controller != game.Player1 {
		t.Fatal("real reanimation did not establish a nonowner default controller")
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Players[game.Player1].Life = 0
	policies := simPassPolicies()
	simulated := engine.Simulator().ResolvePriority(g, policies)
	live := g.Clone()
	engine.runPriorityLoop(live, policies, &TurnLog{})
	for _, resolved := range []*game.Game{live, simulated} {
		if _, present := permanentByObjectID(resolved, reanimated.ObjectID); present ||
			!resolved.Players[game.Player2].Exile.Contains(original.CardInstanceID) {
			t.Fatal("reanimation departure disagreed through live/Simulator priority")
		}
		if resolved.Turn.ActivePlayer != game.Player1 || resolved.Turn.Phase != game.PhasePrecombatMain {
			t.Fatal("active-player departure prematurely ended the turn")
		}
		if resolved.Turn.PriorityPlayer == game.Player1 {
			t.Fatal("departed player retained priority")
		}
	}
	if g.Players[game.Player1].Eliminated {
		t.Fatal("Simulator mutated its input")
	}
}

func TestPlayerDepartureControlSourceRemovalExilesImmediately(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Controller = game.Player2
	source := addCombatPermanent(g, game.Player3, &game.CardDef{})
	if !applySourceTiedControlEffect(g, game.Player3, source, creature, game.DurationForAsLongAsSourceOnBattlefield) {
		t.Fatal("could not establish surviving control")
	}
	eliminatePlayer(g, game.Player2)
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("could not remove the surviving control source")
	}
	if _, present := permanentByObjectID(g, creature.ObjectID); present ||
		!g.Players[game.Player1].Exile.Contains(creature.CardInstanceID) {
		t.Fatal("source departure waited for an SBA to exile the departed-default object")
	}
}

func TestPlayerDepartureLeavesBattlefieldButDoesNotDieOrEnterExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Battlefield,
		SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	owned := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	owned.Controller = game.Player1
	eliminatePlayer(g, game.Player2)
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatal("owned permanent leaving the game did not trigger the surviving LTB watcher once")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != 41 {
		t.Fatal("leaving-game trigger did not resolve normally")
	}
	for _, event := range g.Events {
		if event.PermanentID == owned.ObjectID &&
			(event.Kind == game.EventPermanentDied || event.ToZone == zone.Exile || event.Kind == game.EventZoneChanged) {
			t.Fatal("owned permanent leaving the game fabricated death/exile/ordinary movement")
		}
	}
}

func TestPlayerDepartureRealExileHonorsDestinationReplacement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Controller = game.Player2
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
		MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Exile,
		ReplaceToZone: zone.Hand,
	})
	eliminatePlayer(g, game.Player2)
	if !g.Players[game.Player1].Hand.Contains(creature.CardInstanceID) ||
		g.Players[game.Player1].Exile.Contains(creature.CardInstanceID) {
		t.Fatal("actual departure exile bypassed its normal replacement path")
	}
}

func TestPlayerDepartureOwnedStackCopyAndSurvivingAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	copyID, abilityID := g.IDGen.Next(), g.IDGen.Next()
	g.Stack.Push(&game.StackObject{
		ID: copyID, Kind: game.StackSpell, Controller: game.Player3,
		Copy: true, CopyOwner: opt.Val(game.Player2), SourceID: source.CardInstanceID,
	})
	g.Stack.Push(&game.StackObject{
		ID: abilityID, Kind: game.StackActivatedAbility, Controller: game.Player3,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
	})
	eliminatePlayer(g, game.Player2)
	top, ok := g.Stack.Peek()
	if !ok || g.Stack.Size() != 1 || top.ID != abilityID {
		t.Fatal("departure confused a copy's owner with a surviving ability's source owner")
	}
	if _, ok := g.GetCardInstance(source.CardInstanceID); !ok {
		t.Fatal("departure discarded historical source definition needed by surviving abilities")
	}
}

func TestPlayerDepartureNextTurnDurationsEndAtSkippedSeat(t *testing.T) {
	for _, extra := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.ActivePlayer = game.Player4
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID: g.IDGen.Next(), Layer: game.LayerPowerToughnessModify, AffectedObjectID: creature.ObjectID,
			PowerDelta: 1, Duration: game.DurationUntilYourNextTurn, ExpiresFor: game.Player2,
		})
		g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
			Kind: game.RuleEffectNoMaximumHandSize, Controller: game.Player1,
			Duration: game.DurationUntilEndOfYourNextTurn, ExpiresFor: game.Player2,
		})
		eliminatePlayer(g, game.Player2)
		if effectivePower(g, creature) != 3 || len(g.RuleEffects) != 1 {
			t.Fatal("departure expired a next-turn duration immediately")
		}
		engine := NewEngine(nil)
		if extra {
			g.Turn.ExtraTurns = []game.PlayerID{game.Player2}
		} else {
			engine.advanceToNextTurn(g)
			if effectivePower(g, creature) != 3 {
				t.Fatal("duration ended before the departed seat's next turn")
			}
		}
		engine.advanceToNextTurn(g)
		if effectivePower(g, creature) != 2 || len(g.RuleEffects) != 0 ||
			g.Turn.ActivePlayer == game.Player2 {
			t.Fatalf("duration survived a departed player's would-be turn (extra=%v)", extra)
		}
	}
}

func TestPlayerDepartureCannotReceiveObjectsOrActions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	eliminatePlayer(g, game.Player2)
	cardID := addCardToHand(g, game.Player1, vanillaCreature("Cannot enter", 2, 2))
	card, _ := g.GetCardInstance(cardID)
	obj := &game.StackObject{Controller: game.Player3}
	resolver := &effectResolver{engine: engine, game: g, obj: obj}
	if _, entered := resolver.putResolvedCardOnBattlefieldValue(card, zone.Hand, game.Player2, nil, permanentCreationOptions{}); entered ||
		!g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("attempted entry under a departed controller moved the card")
	}
	if token, created := createTokenPermanent(g, game.Player2, vanillaCreature("Cannot create", 1, 1)); created || token != nil {
		t.Fatal("departed player received a token")
	}
	if engine.applyAction(g, game.Player2, action.Pass()) {
		t.Fatal("departed player performed an action")
	}
}

type departedDecisionProbe struct{ calls int }

func (p *departedDecisionProbe) ChooseAction(PlayerObservation, []action.Action) action.Action {
	p.calls++
	return action.Pass()
}

func (p *departedDecisionProbe) ChooseChoice(PlayerObservation, game.ChoiceRequest) []int {
	p.calls++
	return nil
}

func TestPlayerDepartureActiveTurnFinishesWithoutDepartedDecisions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep, Step: game.StepEnd,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()}}}, nil)
	eliminatePlayer(g, game.Player2)
	probe := &departedDecisionProbe{}
	policies := simPassPolicies()
	policies[game.Player2] = probe
	engine.finishCurrentTurn(g, policies)
	if g.Players[game.Player1].Life != 42 || g.Turn.ActivePlayer != game.Player3 {
		t.Fatal("the departed active player's remaining end-step window did not finish")
	}
	if probe.calls != 0 {
		t.Fatal("the departed active player received a decision")
	}
}

func TestPlayerDepartureNestedSearchControlUnwinds(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	endOuter := beginSearchControl(g, searchControl{searcher: game.Player3, decisionMaker: game.Player4})
	endInner := beginSearchControl(g, searchControl{searcher: game.Player1, decisionMaker: game.Player2})
	eliminatePlayer(g, game.Player2)
	if g.DecisionMaker(game.Player1, game.DecisionRules) != game.Player1 ||
		g.DecisionMaker(game.Player3, game.DecisionRules) != game.Player4 {
		t.Fatal("departure failed to end exactly the affected search override")
	}
	endInner()
	endOuter()
	if len(g.SearchControlScopes) != 0 {
		t.Fatal("nested search-control scopes did not unwind")
	}
}

func TestPlayerDepartureSimultaneousLossesEndAllControlGrantsFirst(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Layer: game.LayerControl, AffectedObjectID: creature.ObjectID,
		NewController: opt.Val(game.Player2), Duration: game.DurationUntilEndOfTurn,
	})
	g.Players[game.Player1].Life = 0
	g.Players[game.Player2].Life = 0
	NewEngine(nil).applyStateBasedActions(g)
	if _, present := permanentByObjectID(g, creature.ObjectID); !present ||
		effectiveController(g, creature) != game.Player3 {
		t.Fatal("simultaneous losses exiled an object before every departing control grant ended")
	}
}

func TestPlayerDepartureForeignControlAuraEndsGrantBeforeExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	aura := addCombatPermanent(g, game.Player1, cardc.ControlMagic())
	aura.Controller = game.Player2
	if !attachPermanent(g, aura, creature) {
		t.Fatal("could not attach Control Magic")
	}
	if effectiveController(g, creature) != game.Player2 {
		t.Fatal("fixture did not establish the foreign Aura's control grant")
	}
	eliminatePlayer(g, game.Player2)
	if _, present := permanentByObjectID(g, aura.ObjectID); present ||
		!g.Players[game.Player1].Exile.Contains(aura.CardInstanceID) {
		t.Fatal("departed-default Aura was not exiled")
	}
	if _, present := permanentByObjectID(g, creature.ObjectID); !present ||
		effectiveController(g, creature) != game.Player3 {
		t.Fatal("the ended static grant incorrectly exiled the enchanted creature")
	}
}

func TestPlayerDepartureCleanupExileSeesAllDurationsExpired(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Controller = game.Player2
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Layer: game.LayerControl, AffectedObjectID: creature.ObjectID,
		NewController: opt.Val(game.Player3), Duration: game.DurationUntilEndOfTurn,
	})
	eliminatePlayer(g, game.Player2)
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged, Duration: game.DurationUntilEndOfTurn,
		MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Exile,
		ReplaceToZone: zone.Hand,
	})
	NewEngine(nil).runCleanupStep(g, simPassPolicies(), &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(creature.CardInstanceID) {
		t.Fatal("cleanup exile used a replacement expiring at that same boundary")
	}
}

func TestPlayerDepartureAbilityOwnershipSurvivesControlChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	ability := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, Controller: game.Player2,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
	}
	g.Stack.Push(ability)
	addEffectSpellToStack(g, game.Player3, game.ChangeStackObjectController{
		Object: game.TargetStackObjectReference(0), Controller: game.ControllerReference(),
	}, []game.Target{game.StackObjectTarget(ability.ID)})
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if ability.Controller != game.Player3 {
		t.Fatal("the real stack-control effect did not resolve")
	}
	branch := g.Clone()
	eliminatePlayer(branch, game.Player2)
	if !branch.Stack.IsEmpty() || g.Stack.Size() != 1 {
		t.Fatal("ability control change lost its original owner or clone isolation")
	}
}
