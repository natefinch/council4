package rules

import (
	"math/rand/v2"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const mindslaverID = "a806f4ee-f48a-46c3-8772-5aaabebe4c7e"

type controlScript struct {
	action   func(PlayerObservation, []action.Action) action.Action
	choice   func(PlayerObservation, game.ChoiceRequest) []int
	observed func(game.PlayerID, action.Action, PlayerObservation)
	actions  []game.PlayerID
	choices  []game.ChoiceRequest
}

func (a *controlScript) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	a.actions = append(a.actions, obs.Player)
	if a.action != nil {
		return a.action(obs, legal)
	}
	return legal[len(legal)-1]
}

func (a *controlScript) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	a.choices = append(a.choices, request)
	if a.choice != nil {
		return a.choice(obs, request)
	}
	return fallbackChoice(request)
}

func (a *controlScript) ObserveAction(actor game.PlayerID, act action.Action, obs PlayerObservation) {
	if a.observed != nil {
		a.observed(actor, act, obs)
	}
}

func mindslaverGame(t *testing.T) (*game.Game, *Engine, *game.CardDef) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def := lifetimeDefinition(t, registry, mindslaverID)
	g, _ := setupBestowMain(t)
	return g, NewEngineWithPrograms(nil, registry.Programs()), def
}

func activateMindslaver(t *testing.T, g *game.Game, e *Engine, def *game.CardDef, controller, target game.PlayerID) *game.Permanent {
	t.Helper()
	source := addCombatPermanent(g, controller, def)
	g.Turn.PriorityPlayer = controller
	g.Players[controller].ManaPool.Add(mana.C, 4)
	act := action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PlayerTarget(target)}, 0)
	if !containsAction(e.Simulator().LegalActions(g, controller), act) ||
		!e.applyAction(g, controller, act) {
		t.Fatal("Mindslaver activation was not legal/payable")
	}
	if !source.Tapped || !g.Players[controller].Graveyard.Contains(source.CardInstanceID) ||
		permanentByCardID(g, source.CardInstanceID) != nil {
		t.Fatal("activation did not tap and sacrifice the source")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	return source
}

func controlledMindslaverGame(t *testing.T) (*game.Game, *Engine, *game.CardDef) {
	t.Helper()
	g, e, def := mindslaverGame(t)
	activateMindslaver(t, g, e, def, game.Player1, game.Player2)
	e.advanceToNextTurn(g)
	if g.Turn.ActivePlayer != game.Player2 || g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player1 {
		t.Fatal("next turn did not start controlled")
	}
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	return g, e, def
}

func TestMindslaverCastingActivationCostsTargetsSourceAndCounter(t *testing.T) {
	g, e, def := mindslaverGame(t)
	card := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 5)
	cast := action.CastSpell(card, nil, 0, nil)
	if e.applyAction(g, game.Player1, cast) {
		t.Fatal("cast without six mana succeeded")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if !e.applyAction(g, game.Player1, cast) {
		t.Fatal("actual six-mana cast failed")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	source := permanentByCardID(g, card)
	act := action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PlayerTarget(game.Player2)}, 0)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	if e.applyAction(g, game.Player1, act) || source.Tapped || permanentByCardID(g, card) == nil {
		t.Fatal("unpayable activation changed the source")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	source.Tapped = true
	if e.applyAction(g, game.Player1, act) {
		t.Fatal("tapped source activated")
	}
	source.Tapped = false
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Player ward", Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind: game.RuleEffectPlayerHexproof, AffectedPlayer: game.PlayerYou,
		}}}},
	}})
	if e.applyAction(g, game.Player1, act) || e.applyAction(g, game.Player2, act) {
		t.Fatal("illegal target or foreign resources allowed activation")
	}
	self := action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PlayerTarget(game.Player1)}, 0)
	if !e.applyAction(g, game.Player1, self) {
		t.Fatal("self target should be legal")
	}
	obj, _ := g.Stack.Peek()
	if obj.Controller != game.Player1 || !source.Tapped ||
		!g.Players[game.Player1].Graveyard.Contains(card) || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("activation did not pay all costs independently of resolution")
	}
	if !counterStackObject(g, obj.ID) || len(g.PlayerControlEffects) != 0 {
		t.Fatal("countered activation scheduled control")
	}
}

func TestMindslaverTargetLeavingAndLostTargetLegality(t *testing.T) {
	for _, leave := range []bool{false, true} {
		t.Run(map[bool]string{false: "hexproof", true: "leave"}[leave], func(t *testing.T) {
			g, e, def := mindslaverGame(t)
			source := addCombatPermanent(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.C, 4)
			if !e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PlayerTarget(game.Player2)}, 0)) {
				t.Fatal("activation failed")
			}
			if leave {
				e.Concede(g, game.Player2)
			} else {
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
					ID: g.IDGen.Next(), Kind: game.RuleEffectPlayerHexproof, Controller: game.Player2,
					AffectedPlayer: game.PlayerYou, Duration: game.DurationPermanent,
				})
			}
			e.resolveTopOfStack(g, &TurnLog{})
			if len(g.PlayerControlEffects) != 0 || !g.Players[game.Player1].Graveyard.Contains(source.CardInstanceID) {
				t.Fatal("illegal resolution target scheduled control or undid sacrifice")
			}
		})
	}
}

func TestMindslaverWholeTurnActionsModesTargetsAndResolvingChoices(t *testing.T) {
	g, e, def := mindslaverGame(t)
	activateMindslaver(t, g, e, def, game.Player1, game.Player2)
	addCardToLibraryNamed(g, game.Player2, "Draw witness")
	victimSpell := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Modal choice witness", ManaCost: opt.Val(cost.Mana{}), Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{MinModes: 1, MaxModes: 1, Modes: []game.Mode{
			{Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1)}}}},
			{Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer}},
				Sequence: []game.Instruction{
					{Primitive: game.LoseLife{Player: game.TargetPlayerReference(0), Amount: game.Fixed(3)}},
					{Primitive: game.Scry{Amount: game.Fixed(1), Player: game.ControllerReference()}},
				}},
		}}),
	}})
	scry := addCardToLibraryNamed(g, game.Player2, "Scry witness")
	addCardToLibraryNamed(g, game.Player2, "Drawn first")
	g.Turn.ExtraTurns = []game.PlayerID{game.Player2}
	e.advanceToNextTurn(g)
	controller := &controlScript{}
	controller.action = func(obs PlayerObservation, legal []action.Action) action.Action {
		if obs.DecisionMaker != game.Player1 {
			t.Fatal("controlled decision attributed to the actor")
		}
		if obs.Player == game.Player2 && obs.Turn.Phase == game.PhasePrecombatMain {
			for _, act := range legal {
				payload, ok := act.CastSpellPayload()
				if ok && payload.CardID == victimSpell && slices.Equal(payload.ChosenModes, []int{1}) &&
					slices.Equal(payload.Targets, []game.Target{game.PlayerTarget(game.Player3)}) {
					return act
				}
			}
		}
		return legal[len(legal)-1]
	}
	controller.choice = func(obs PlayerObservation, req game.ChoiceRequest) []int {
		if req.Player != obs.Player || obs.DecisionMaker != game.Player1 {
			t.Fatal("resolving choice lost actor/decision maker separation")
		}
		if req.Kind == game.ChoiceScry {
			if !req.Subject.Exists || req.Subject.Val.CardID != scry {
				t.Fatal("controlling agent did not receive the private scry subject")
			}
			return []int{1}
		}
		return fallbackChoice(req)
	}
	victim := &controlScript{}
	log := e.runTurn(g, [game.NumPlayers]PlayerAgent{controller, victim, &controlScript{}, &controlScript{}})
	if g.Players[game.Player3].Life != 37 || !g.Players[game.Player2].Graveyard.Contains(victimSpell) ||
		g.Turn.ActivePlayer != game.Player3 || g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player2 ||
		len(victim.actions) != 0 || len(victim.choices) != 0 {
		t.Fatal("whole-turn control did not cast the actor's selected mode/target or expire")
	}
	found := false
	for _, choice := range log.Choices {
		if choice.Request.Kind == game.ChoiceScry {
			found = true
			if choice.DecisionMaker != game.Player1 || choice.Request.Player != game.Player2 ||
				choice.Request.Subject.Exists || !choice.Request.Private || choice.UsedFallback ||
				!slices.Equal(choice.Selected, []int{1}) {
				t.Fatal("Scry resolving callback did not use the control protocol")
			}
		}
	}
	if !found || !slices.Contains(controller.actions, game.Player1) || !slices.Contains(controller.actions, game.Player2) {
		t.Fatalf("controller did not retain their own decisions as well: scry=%v actors=%v choices=%+v", found, controller.actions, log.Choices)
	}
	for _, act := range log.Actions {
		if act.Player == game.Player2 && act.DecisionMaker != game.Player1 {
			t.Fatal("turn log lost the action's decision maker")
		}
	}
}

func TestMindslaverScriptedActivationSimulatorAndCleanupPriority(t *testing.T) {
	g, e, def := mindslaverGame(t)
	source := addCombatPermanent(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	activate := action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PlayerTarget(game.Player2)}, 0)
	branch, ok := e.Simulator().Apply(g, game.Player1, activate, [game.NumPlayers]PlayerAgent{})
	if !ok || source.Tapped || len(g.PlayerControlEffects) != 0 || permanentByCardID(g, source.CardInstanceID) == nil {
		t.Fatal("Simulator activation was not isolated")
	}
	branch = e.Simulator().ResolvePriority(branch, [game.NumPlayers]PlayerAgent{})
	if len(branch.PlayerControlEffects) != 1 || len(g.PlayerControlEffects) != 0 {
		t.Fatal("Simulator did not schedule the actual authored activation")
	}
	script := &controlScript{action: func(_ PlayerObservation, legal []action.Action) action.Action {
		if containsAction(legal, activate) {
			return activate
		}
		return legal[len(legal)-1]
	}}
	e.runPriorityLoop(g, [game.NumPlayers]PlayerAgent{script}, &TurnLog{})
	if len(g.PlayerControlEffects) != 1 || !g.Players[game.Player1].Graveyard.Contains(source.CardInstanceID) {
		t.Fatal("scripted activation/priority resolution did not schedule control")
	}
	e.advanceToNextTurn(g)
	for range 9 {
		addCardToHand(g, game.Player2, vanillaCreature("Cleanup card", 1, 1))
	}
	instant := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "End step instant", Types: []types.Card{types.Instant}, ManaCost: opt.Val(cost.Mana{}),
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}}.Ability()),
	}})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Discard trigger witness", Types: []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event: game.EventCardDiscarded, Player: game.TriggerPlayerYou,
			}},
			Content: game.Mode{
				Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer}},
				Sequence: []game.Instruction{{Primitive: game.LoseLife{Player: game.TargetPlayerReference(0), Amount: game.Fixed(1)}}},
			}.Ability(),
		}},
	}})
	cleanupPriority := false
	script.action = func(obs PlayerObservation, legal []action.Action) action.Action {
		if obs.Player == game.Player2 && obs.Turn.Step == game.StepCleanup {
			cleanupPriority = true
			if obs.DecisionMaker != game.Player1 {
				t.Fatal("control ended before cleanup priority")
			}
		}
		if obs.Player == game.Player2 && obs.Turn.Step == game.StepEnd {
			for _, act := range legal {
				if cast, ok := act.CastSpellPayload(); ok && cast.CardID == instant {
					return act
				}
			}
		}
		return legal[len(legal)-1]
	}
	script.choice = func(obs PlayerObservation, req game.ChoiceRequest) []int {
		if obs.DecisionMaker != game.Player1 || req.Player != game.Player2 {
			t.Fatal("cleanup choice was not controlled")
		}
		if req.Kind == game.ChoiceZoneSelection {
			return []int{0, 1}
		}
		for _, option := range req.Options {
			if slices.Equal(option.Targets, []game.Target{game.PlayerTarget(game.Player3)}) {
				return []int{option.Index}
			}
		}
		return fallbackChoice(req)
	}
	log := &TurnLog{}
	e.runEndingPhaseWithLog(g, [game.NumPlayers]PlayerAgent{script, &controlScript{}}, log)
	if !g.Players[game.Player2].Graveyard.Contains(instant) || g.Players[game.Player2].Life != 41 ||
		g.Players[game.Player2].Hand.Size() != 7 || g.Players[game.Player3].Life != 38 || !cleanupPriority {
		t.Fatalf("end/cleanup priority: cast=%v life=%d hand=%d opponent=%d cleanup=%v choices=%+v", g.Players[game.Player2].Graveyard.Contains(instant), g.Players[game.Player2].Life, g.Players[game.Player2].Hand.Size(), g.Players[game.Player3].Life, cleanupPriority, log.Choices)
	}
	if len(log.Steps) < 3 || log.Steps[len(log.Steps)-1].Step != game.StepCleanup ||
		g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player1 {
		t.Fatal("cleanup did not repeat or ended control too early")
	}
}

func TestMindslaverCombatAttackAndBlockDispatch(t *testing.T) {
	g, e, _ := controlledMindslaverGame(t)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	attacker.SummoningSick, blocker.SummoningSick = false, false
	controller := &controlScript{action: func(obs PlayerObservation, legal []action.Action) action.Action {
		for _, act := range legal {
			if attack, ok := act.DeclareAttackersPayload(); ok && len(attack.Attackers) == 1 &&
				attack.Attackers[0].Target.Player == game.Player3 {
				if obs.Player != game.Player2 || obs.DecisionMaker != game.Player1 {
					t.Fatal("attacker decision routed incorrectly")
				}
				return act
			}
		}
		return legal[len(legal)-1]
	}}
	defender := &controlScript{action: func(_ PlayerObservation, legal []action.Action) action.Action {
		for _, act := range legal {
			if block, ok := act.DeclareBlockersPayload(); ok && len(block.Blockers) == 1 {
				return act
			}
		}
		return legal[len(legal)-1]
	}}
	e.runCombatPhase(g, [game.NumPlayers]PlayerAgent{controller, &controlScript{}, defender}, &TurnLog{})
	if !g.Players[game.Player3].Graveyard.Contains(blocker.CardInstanceID) || !attacker.Tapped {
		t.Fatal("controlled attack was not executed through actual combat")
	}
	// Exercise the same shared protocol at a defender decision seam. Mindslaver
	// itself does not make an active player a defender; limited-duration control
	// also uses this dispatch, without exchanging the active or defending seats.
	g.Turn.ActivePlayer = game.Player3
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	controller.action = func(obs PlayerObservation, legal []action.Action) action.Action {
		for _, act := range legal {
			if block, ok := act.DeclareBlockersPayload(); ok && len(block.Blockers) == 1 &&
				block.Blockers[0].Blocker == second.ObjectID {
				if obs.Player != game.Player2 || obs.DecisionMaker != game.Player1 {
					t.Fatal("blocker decision routed incorrectly")
				}
				return act
			}
		}
		return legal[len(legal)-1]
	}
	combatEngine{e}.declareBlockers(g, [game.NumPlayers]PlayerAgent{controller, &controlScript{}}, &TurnLog{})
	if len(g.Combat.Blockers) != 1 || g.Combat.Blockers[0].Blocker != second.ObjectID {
		t.Fatal("blocker callback bypassed shared dispatch")
	}
}

func TestMindslaverManaAdditionalPaymentAndReplacementDecisions(t *testing.T) {
	g, e, _ := controlledMindslaverGame(t)
	manaSource := addCombatPermanent(g, game.Player2, combinationManaSource([]mana.Color{mana.R, mana.G}, 3))
	first := addCombatPermanent(g, game.Player2, permanentWitness(types.Artifact))
	second := addCombatPermanent(g, game.Player2, permanentWitness(types.Artifact))
	spell := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Payment witness", ManaCost: opt.Val(cost.Mana{cost.R}), Types: []types.Card{types.Instant},
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrifice, MatchPermanentType: true, PermanentType: types.Artifact}},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.Pay{Payment: game.ResolutionPayment{
			Prompt: "Pay another red?", ManaCost: opt.Val(cost.Mana{cost.R}),
		}}}}}.Ability()),
	}})
	controller := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		if obs.Player != game.Player2 || obs.DecisionMaker != game.Player1 || req.Player != game.Player2 {
			t.Fatal("payment was offered to the wrong seat")
		}
		if req.Kind == game.ChoiceManaCombination {
			return []int{0, 0, 0}
		}
		if req.Kind == game.ChoiceReplacement {
			return []int{len(req.Options) - 1}
		}
		if req.Kind == game.ChoicePayment {
			for _, option := range req.Options {
				if option.Card.Exists && option.Card.Val.CardID == second.CardInstanceID ||
					strings.Contains(option.Label, "Pay") {
					return []int{option.Index}
				}
			}
		}
		return fallbackChoice(req)
	}}
	victim := &controlScript{}
	policies := [game.NumPlayers]PlayerAgent{controller, victim}
	log := &TurnLog{}
	if !e.applyActionWithChoices(g, game.Player2, action.ActivateAbility(manaSource.ObjectID, 0, nil, 0), policies, log) ||
		g.Players[game.Player2].ManaPool.Amount(mana.R) != 3 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("mana choice/pool was not actor-owned")
	}
	if !e.applyActionWithChoices(g, game.Player2, action.CastSpell(spell, nil, 0, nil), policies, log) {
		t.Fatal("controlled spell/additional payment failed")
	}
	e.resolveTopOfStackWithChoices(g, policies, log)
	if !g.Players[game.Player2].Graveyard.Contains(second.CardInstanceID) ||
		permanentByCardID(g, first.CardInstanceID) == nil || g.Players[game.Player2].ManaPool.Amount(mana.R) != 1 {
		t.Fatal("controller's payment choice did not spend only victim resources")
	}
	addReplacementPermanent(t, g, game.Player1, damageAddendReplacementCardDef())
	addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	damageSource := addColoredSourceCard(g, game.Player1, color.Red)
	e.setReplacementChoiceContext(g, policies, log)
	dealt := dealPlayerDamage(g, damageSource, 0, game.Player1, game.Player2, 2, false)
	g.ClearChoiceContext()
	if dealt != 5 || len(g.ReplacementDecisions) != 1 || g.ReplacementDecisions[0].UsedFallback ||
		!slices.Equal(g.ReplacementDecisions[0].Selected, []int{1}) || len(victim.choices) != 0 {
		t.Fatal("replacement selection bypassed the decision controller")
	}
}

func TestMindslaverNextActualTurnLifetimeCompetitionCloneAndLeaving(t *testing.T) {
	for _, extra := range []bool{false, true} {
		t.Run(map[bool]string{false: "regular skip", true: "extra skip"}[extra], func(t *testing.T) {
			g, e, def := mindslaverGame(t)
			source := activateMindslaver(t, g, e, def, game.Player1, game.Player2)
			if source.ObjectID != g.PlayerControlEffects[0].SourceObjectID {
				t.Fatal("source departure lost captured provenance")
			}
			activateMindslaver(t, g, e, def, game.Player3, game.Player2)
			if g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player2 {
				t.Fatal("pending effect started in the resolving turn")
			}
			g.SkippedTurns = map[game.PlayerID]int{game.Player2: 1}
			if extra {
				g.Turn.ExtraTurns = []game.PlayerID{game.Player2}
			}
			branch := g.Clone()
			branch.PlayerControlEffects[0].Controller = game.Player4
			if g.PlayerControlEffects[0].Controller != game.Player3 {
				t.Fatal("pending identities were aliased or source departure lost provenance")
			}
			e.advanceToNextTurn(g)
			if extra {
				// A skipped extra turn does not advance the regular rotation.
				if g.Turn.ActivePlayer != game.Player2 {
					t.Fatal("skipped extra turn moved the normal turn cursor")
				}
			} else {
				if g.Turn.ActivePlayer != game.Player3 || len(g.PlayerControlEffects) != 1 {
					t.Fatal("skip consumed pending control")
				}
				for g.Turn.ActivePlayer != game.Player2 {
					e.advanceToNextTurn(g)
				}
			}
			if g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player3 || len(g.PlayerControlEffects) != 1 {
				t.Fatal("creation timestamp did not overwrite all competing effects")
			}
			e.Concede(g, game.Player3)
			if g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player2 || len(g.PlayerControlEffects) != 0 {
				t.Fatal("departing controller failed to restore victim or revived earlier control")
			}
		})
	}
	g, e, def := mindslaverGame(t)
	activateMindslaver(t, g, e, def, game.Player1, game.Player2)
	g.Turn.ExtraTurns = []game.PlayerID{game.Player2, game.Player2}
	e.advanceToNextTurn(g)
	if g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player1 {
		t.Fatal("first extra turn was not the next turn")
	}
	e.runEndingPhaseWithLog(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player1 {
		t.Fatal("control ended in cleanup instead of at the next turn")
	}
	e.advanceToNextTurn(g)
	if g.Turn.ActivePlayer != game.Player2 || g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player2 {
		t.Fatal("control leaked into consecutive extra turn")
	}
}

func TestMindslaverPersonalChoicesConcessionAndCleanup(t *testing.T) {
	g, e, _ := controlledMindslaverGame(t)
	for range 9 {
		addCardToHand(g, game.Player2, permanentWitness(types.Artifact))
	}
	controller := &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		if req.Kind != game.ChoiceZoneSelection || obs.Player != game.Player2 {
			t.Fatal("personal decision was redirected to controller")
		}
		return []int{0, 1}
	}}
	victim := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int { return []int{1} }}
	policies := [game.NumPlayers]PlayerAgent{controller, victim}
	first := g.Players[game.Player2].Hand.All()[:2]
	e.discardToMaximumHandSizeWithChoices(g, game.Player2, policies, &TurnLog{})
	for _, card := range first {
		if !g.Players[game.Player2].Graveyard.Contains(card) {
			t.Fatal("cleanup discard was not controlled")
		}
	}
	req := game.ChoiceRequest{Kind: game.ChoiceMay, Player: game.Player2, DecisionKind: game.DecisionPersonal,
		Options: []game.ChoiceOption{{Index: 0}, {Index: 1}}, MinChoices: 1, MaxChoices: 1}
	if !slices.Equal(e.chooseChoice(g, policies, req, &TurnLog{}), []int{1}) || len(victim.choices) != 1 ||
		g.DecisionMaker(game.Player2, game.DecisionPersonal) != game.Player2 {
		t.Fatal("out-of-game/personal decision was made by controlling player")
	}
	if !e.Concede(g, game.Player2) || !g.Players[game.Player2].Eliminated ||
		g.Players[game.Player1].Eliminated || len(g.PlayerControlEffects) != 0 {
		t.Fatal("victim could not concede normally or effects survived its departure")
	}
}

func TestMindslaverSimulatorSearchAndConcurrentStreams(t *testing.T) {
	g, e, _ := controlledMindslaverGame(t)
	hand := addCardToHand(g, game.Player2, permanentWitness(types.Artifact))
	own := addCardToHand(g, game.Player1, permanentWitness(types.Creature))
	for range 6 {
		addCardToLibraryNamed(g, game.Player1, "Unknown own library")
		addCardToLibraryNamed(g, game.Player2, "Unknown controlled library")
		addCardToLibraryNamed(g, game.Player3, "Unknown opponent library")
	}
	addCardToHand(g, game.Player3, permanentWitness(types.Enchantment))
	ctx := e.newSearchContext(g, game.Player2)
	if ctx.Player() != game.Player2 || ctx.DecisionMaker() != game.Player1 {
		t.Fatal("search context conflated actor and decision maker")
	}
	world := ctx.Determinize()
	if !world.Players[game.Player2].Hand.Contains(hand) || !world.Players[game.Player1].Hand.Contains(own) ||
		world.DecisionMaker(game.Player2, game.DecisionRules) != game.Player1 {
		t.Fatal("determinization changed known hands or decision policy")
	}
	act := action.CastSpell(hand, nil, 0, nil)
	branch, ok := ctx.Simulator().Apply(world, game.Player2, act, [game.NumPlayers]PlayerAgent{})
	if !ok || !g.Players[game.Player2].Hand.Contains(hand) || !world.Players[game.Player2].Hand.Contains(hand) {
		t.Fatal("Simulator application failed or mutated the live game")
	}
	branch = ctx.Simulator().ResolvePriority(branch, [game.NumPlayers]PlayerAgent{})
	if permanentByCardID(branch, hand) == nil || permanentByCardID(branch, hand).Controller != game.Player2 {
		t.Fatal("Simulator changed object control with decision control")
	}
	sim := ctx.Simulator()
	if err := sim.engine.ValidatePrograms([game.NumPlayers]game.PlayerConfig{{Deck: []*game.CardDef{authored.Mindslaver.New()}}}); err != nil {
		t.Fatal("search Engine lost immutable program catalog")
	}
	// Search must not consume any of the live rules RNG, even across clones.
	want := rand.New(rand.NewPCG(1, 2)).Uint64()
	if e.rng.Uint64() != want {
		t.Fatal("search advanced the live random stream")
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			local := g.Clone()
			engine := NewEngineWithPrograms(nil, sim.engine.programs)
			local.PlayerControlEffects[0].Controller = game.Player4
			if local.DecisionMaker(game.Player2, game.DecisionRules) != game.Player4 {
				t.Error("parallel clone lost policy")
			}
			engine.advanceToNextTurn(local)
			if local.DecisionMaker(game.Player2, game.DecisionRules) != game.Player2 {
				t.Error("parallel clone retained expired policy")
			}
		})
	}
	wg.Wait()
	if g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player1 {
		t.Fatal("parallel games aliased live decision state")
	}
}
