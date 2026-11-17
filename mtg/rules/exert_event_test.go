package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3935, the exert event: "Whenever you exert a creature"
// triggers and "If this creature hasn't been exerted this turn" (sub-issue of
// #1751, after #3934's exert-as-it-attacks cost). Pinned-corpus Oracle text
// (fixtures below are unregistered; a later card unit admits them): Vizier of
// the True HOU 28 printing c1800882-e886-48d5-ac85-0a8e5841dc9a (oracle
// 0e31d70e-2b5e-4c32-81cf-aed2a13e8aa5), Resolute Survivors HOU 142 printing
// 713f469e-58c3-4db4-aa7a-77126bda13f2 (oracle
// 3d699db7-cc52-4ee3-947b-0ba291bf037a), Trueheart Twins J25 605 printing
// 34e6cfec-b6a0-4079-8d06-2a99a70d8b37 (oracle
// 5d3be0d2-0940-4904-9f12-de592c990bdb) and Combat Celebrant TRC 451 printing
// ee9d77bb-ed05-49d9-8ced-d956101d3c65 (oracle
// 5e15ff93-99a0-4000-918e-4bd2c257188d). Scryfall rulings
// (https://api.scryfall.com/cards/<printing>/rulings) returned 6, 6, 5 and 9
// rulings: "whenever you exert a creature" triggers when you exert that
// creature or any other creature you control; it can be exerted with no legal
// target for its trigger; tap-and-freeze effects (Decision Paralysis) aren't
// exert; exert triggers resolve before blockers; Combat Celebrant untaps all
// your other creatures, not just attackers, untapping an attacker doesn't
// remove it from combat, there's no main phase between the combats, and the
// additional combat happens even if Combat Celebrant doesn't survive; exert
// Celebrants one at a time in successive combats to keep attacking (2017-04-18;
// 2017-07-14 Two-Headed Giant team damage is not modeled). Registered Pride
// Sovereign (oracle a131c32d-2b4f-4ee4-aab2-ab05ab010978) drives the exert-cost
// route and Glacial Grasp the tap-and-freeze route. CR 2025-09-19 from the
// verified official copy: 701.43a-d, 508.1g, 603.2, 603.3b, 603.12, 607.2h,
// 500.8, 506.1, 400.7.

const (
	exertOnlyText         = "You may exert this creature as it attacks. (It won't untap during your next untap step.)"
	vizierOfTheTrueText   = "Whenever you exert a creature, tap target creature an opponent controls."
	resoluteSurvivorsText = "Whenever you exert a creature, this creature deals 1 damage to each opponent and you gain 1 life."
	trueheartTwinsText    = "Whenever you exert a creature, creatures you control get +1/+0 until end of turn."
	combatCelebrantText   = "If this creature hasn't been exerted this turn, you may exert it as it attacks. When you do, untap all other creatures you control and after this phase, there is an additional combat phase. (An exerted creature won't untap during your next untap step.)"
)

// whenYouExertCreatureFixture is a creature with a bare "You may exert this
// creature as it attacks." and a "Whenever you exert a creature" trigger.
func whenYouExertCreatureFixture(name string, manaCost cost.Mana, colors []color.Color, subtypes []types.Sub, power, toughness int, triggerText string, content game.AbilityContent) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		ManaCost:  opt.Val(manaCost),
		Colors:    colors,
		Types:     []types.Card{types.Creature},
		Subtypes:  subtypes,
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
		StaticAbilities: []game.StaticAbility{{
			Text:             exertOnlyText,
			ExertAsItAttacks: opt.Val(game.ExertAttackCost{}),
		}},
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: triggerText,
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:            game.EventPermanentExerted,
					Player:           game.TriggerPlayerYou,
					SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
				},
			},
			Content: content,
		}},
		OracleText: exertOnlyText + "\n" + triggerText,
	}}
}

func vizierOfTheTrueFixture() *game.CardDef {
	return whenYouExertCreatureFixture("Vizier of the True", cost.Mana{cost.O(3), cost.W}, []color.Color{color.White},
		[]types.Sub{"Human", "Cleric"}, 3, 2, vizierOfTheTrueText,
		game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature an opponent controls",
				Allow:      game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					RequiredTypesAny: []types.Card{types.Creature},
					Controller:       game.ControllerOpponent,
				}),
			}},
			Sequence: []game.Instruction{{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}}},
		}.Ability())
}

func resoluteSurvivorsFixture() *game.CardDef {
	return whenYouExertCreatureFixture("Resolute Survivors", cost.Mana{cost.O(1), cost.R, cost.W}, []color.Color{color.Red, color.White},
		[]types.Sub{"Human", "Warrior"}, 3, 3, resoluteSurvivorsText,
		game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Damage{
				Amount:       game.Fixed(1),
				Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
				DamageSource: opt.Val(game.SourcePermanentReference()),
			}},
			{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}},
		}}.Ability())
}

func trueheartTwinsFixture() *game.CardDef {
	return whenYouExertCreatureFixture("Trueheart Twins", cost.Mana{cost.O(4), cost.R}, []color.Color{color.Red},
		[]types.Sub{"Jackal", "Warrior"}, 4, 4, trueheartTwinsText,
		game.Mode{Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
			ContinuousEffects: []game.ContinuousEffect{{
				Duration: game.DurationPermanent,
				Group: game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Controller:    game.ControllerYou,
				}),
				Layer:      game.LayerPowerToughnessModify,
				PowerDelta: 1,
			}},
			Duration: game.DurationUntilEndOfTurn,
		}}}}.Ability())
}

func combatCelebrantFixture() *game.CardDef {
	def := exertCreatureFixture("Combat Celebrant", cost.Mana{cost.O(2), cost.R}, []color.Color{color.Red},
		[]types.Sub{"Human", "Warrior"}, 4, 1, combatCelebrantText,
		game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Untap{Group: game.BattlefieldGroupExcluding(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerYou,
			}, game.SourcePermanentReference())}},
			{Primitive: game.AddExtraPhases{Combat: true}},
		}}.Ability())
	def.StaticAbilities[0].Condition = opt.Val(game.Condition{
		Text:   "If this creature hasn't been exerted this turn",
		Negate: true,
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:   game.EventPermanentExerted,
				Source:  game.TriggerSourceSelf,
				Subject: game.TriggerSubjectPermanent,
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})
	return def
}

// exertThenPick exerts the given attackers when asked for attack exertions and
// otherwise answers with the permanent option for target.
func exertThenPick(target *game.Permanent, exert ...*game.Permanent) func(game.ChoiceRequest) []int {
	return func(request game.ChoiceRequest) []int {
		if request.Kind == game.ChoicePayment {
			return pickExertOptions(exert...)(request)
		}
		if target != nil {
			if picked := pickPermanentOption(target)(request); picked != nil {
				return picked
			}
		}
		return fallbackChoice(request)
	}
}

func exertEvents(g *game.Game) []game.Event {
	var events []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentExerted {
			events = append(events, event)
		}
	}
	return events
}

func TestFixturesForExertEventValidate(t *testing.T) {
	for _, def := range []*game.CardDef{vizierOfTheTrueFixture(), resoluteSurvivorsFixture(), trueheartTwinsFixture(), combatCelebrantFixture()} {
		if issues := game.ValidateCardDef(def); len(issues) != 0 {
			t.Fatalf("%s: %v", def.Name, issues)
		}
	}
}

func TestVizierOfTheTrueTriggersOnlyWhenItsControllerExertsACreature(t *testing.T) {
	cases := []struct {
		name        string
		exertVizier bool
		exertOther  bool
		want        bool
	}{
		{name: "exerts itself", exertVizier: true, want: true},
		{name: "another creature exerted", exertOther: true, want: true},
		{name: "attacks without exerting"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newEnlistCombatGame()
			vizier := addCombatPermanent(g, game.Player1, vizierOfTheTrueFixture())
			brawler := addCombatPermanent(g, game.Player1, hoodedBrawlerFixture())
			victim := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
			var exert []*game.Permanent
			if tc.exertVizier {
				exert = append(exert, vizier)
			}
			if tc.exertOther {
				exert = append(exert, brawler)
			}
			policy := &chosenTapPolicy{pick: exertThenPick(victim, exert...)}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = policy
			engine := NewEngine(nil)

			declareExertAttack(t, g, engine, game.Player1, agents, game.Player2, vizier, brawler)
			events := exertEvents(g)
			if len(events) != len(exert) {
				t.Fatalf("exert events = %d, want %d", len(events), len(exert))
			}
			for i, event := range events {
				if event.Player != game.Player1 || event.PermanentID != exert[i].ObjectID {
					t.Fatalf("exert event %d = player %v permanent %v, want Player1 exerting %v", i, event.Player, event.PermanentID, exert[i].ObjectID)
				}
			}
			engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
			for g.Stack.Size() > 0 {
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			}
			if g.Turn.Step != game.StepDeclareAttackers {
				t.Fatalf("step = %v, want the trigger resolved before blockers", g.Turn.Step)
			}
			if victim.Tapped != tc.want {
				t.Fatalf("opponent's creature tapped = %v, want %v", victim.Tapped, tc.want)
			}
		})
	}
}

func TestVizierOfTheTrueIgnoresAnOpponentExertingACreature(t *testing.T) {
	g := newEnlistCombatGame()
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player2, game.Player2
	addCombatPermanent(g, game.Player1, vizierOfTheTrueFixture())
	ownCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	opposingBrawler := addCombatPermanent(g, game.Player2, hoodedBrawlerFixture())
	policy := &chosenTapPolicy{pick: exertThenPick(nil, opposingBrawler)}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player2] = policy
	engine := NewEngine(nil)

	declareExertAttack(t, g, engine, game.Player2, agents, game.Player1, opposingBrawler)
	if events := exertEvents(g); len(events) != 1 || events[0].Player != game.Player2 {
		t.Fatalf("exert events = %+v, want Player2's exert", events)
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	if g.Stack.Size() != 1 {
		t.Fatalf("stack = %d, want only Hooded Brawler's own linked trigger", g.Stack.Size())
	}
	top, _ := g.Stack.Peek()
	if top.Controller != game.Player2 {
		t.Fatal("Vizier of the True triggered on an opponent exerting a creature")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if !opposingBrawler.Tapped || ownCreature.Tapped {
		t.Fatalf("brawler tapped=%v own creature tapped=%v, want only the attacker tapped", opposingBrawler.Tapped, ownCreature.Tapped)
	}
}

func TestTrueheartTwinsTriggersForEachCreatureExertedInOneDeclaration(t *testing.T) {
	g := newEnlistCombatGame()
	twins := addCombatPermanent(g, game.Player1, trueheartTwinsFixture())
	brawler := addCombatPermanent(g, game.Player1, hoodedBrawlerFixture())
	bystander := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	opponent := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	policy := &chosenTapPolicy{pick: exertThenPick(nil, twins, brawler)}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = policy
	engine := NewEngine(nil)

	declareExertAttack(t, g, engine, game.Player1, agents, game.Player2, twins, brawler)
	events := exertEvents(g)
	if len(events) != 2 || events[0].SimultaneousID == 0 || events[0].SimultaneousID != events[1].SimultaneousID {
		t.Fatalf("exert events = %+v, want two exerts sharing one batch", events)
	}
	var attackBatch id.ID
	for _, event := range g.Events {
		if event.Kind == game.EventAttackerDeclared {
			attackBatch = event.SimultaneousID
		}
	}
	if attackBatch != events[0].SimultaneousID {
		t.Fatalf("exert batch %v, want the attack declaration's batch %v", events[0].SimultaneousID, attackBatch)
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	if g.Stack.Size() != 3 {
		t.Fatalf("stack = %d, want two Trueheart Twins triggers and Hooded Brawler's linked trigger", g.Stack.Size())
	}
	for g.Stack.Size() > 0 {
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	}
	if got := effectivePower(g, twins); got != 6 {
		t.Fatalf("Trueheart Twins power = %d, want 4+2", got)
	}
	if got := effectivePower(g, brawler); got != 3+2+2 {
		t.Fatalf("Hooded Brawler power = %d, want 3+2 (Brawler)+2 (Twins)", got)
	}
	if got := effectivePower(g, bystander); got != 3 {
		t.Fatalf("non-attacking creature power = %d, want 1+2", got)
	}
	if got := effectivePower(g, opponent); got != 1 {
		t.Fatalf("opponent's creature power = %d, want 1", got)
	}
}

func TestResoluteSurvivorsTriggersOnExertCostButNotTapAndFreeze(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, resoluteSurvivorsFixture())
	sovereign := addCombatPermanent(g, game.Player1, registryCardDef(t, "Pride Sovereign"))
	frozen := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addCardToLibraryNamed(g, game.Player1, "Grasp Draw")
	for range 2 {
		addCardToLibraryNamed(g, game.Player1, "Milled")
	}
	life := make([]int, game.NumPlayers)
	for i := range life {
		life[i] = g.Players[i].Life
	}

	grasp := addCardToHand(g, game.Player1, registryCardDef(t, "Glacial Grasp"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 3)
	if !engine.applyAction(g, game.Player1, action.CastSpell(grasp, []game.Target{game.PermanentTarget(frozen.ObjectID)}, 0, nil)) {
		t.Fatal("casting Glacial Grasp failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.putTriggeredAbilitiesOnStack(g)
	if !frozen.Tapped || len(exertEvents(g)) != 0 || g.Stack.Size() != 0 {
		t.Fatalf("tap-and-freeze: tapped=%v exert events=%d stack=%d, want a tap that isn't exert", frozen.Tapped, len(exertEvents(g)), g.Stack.Size())
	}

	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(sovereign.ObjectID, 0, nil, 0)) {
		t.Fatal("activating Pride Sovereign failed")
	}
	if events := exertEvents(g); len(events) != 1 || events[0].PermanentID != sovereign.ObjectID || events[0].SimultaneousID != 0 {
		t.Fatalf("exert events = %+v, want Pride Sovereign's exert cost", events)
	}
	engine.putTriggeredAbilitiesOnStack(g)
	if g.Stack.Size() != 2 {
		t.Fatalf("stack = %d, want Pride Sovereign's ability and Resolute Survivors' trigger", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	for i := range life {
		want := life[i] - 1
		if game.PlayerID(i) == game.Player1 {
			want = life[i] + 1
		}
		if g.Players[i].Life != want {
			t.Fatalf("player %d life = %d, want %d", i, g.Players[i].Life, want)
		}
	}
}

func TestCombatCelebrantExtraCombatEvenIfItDiesAndOnlyOncePerTurn(t *testing.T) {
	g := newEnlistCombatGame()
	celebrant := addCombatPermanent(g, game.Player1, combatCelebrantFixture())
	tappedOwn := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	tappedOwn.Tapped = true
	tappedOpponent := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	tappedOpponent.Tapped = true
	policy := &chosenTapPolicy{pick: exertThenPick(nil, celebrant)}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = policy
	engine := NewEngine(nil)

	declareExertAttack(t, g, engine, game.Player1, agents, game.Player2, celebrant)
	if requests := policy.tapRequests(); len(requests) != 1 || !slices.Equal(optionCardIDs(requests[0]), []id.ID{celebrant.CardInstanceID}) {
		t.Fatalf("exert requests = %+v, want Combat Celebrant offered", requests)
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	if g.Stack.Size() != 1 {
		t.Fatalf("stack = %d, want Combat Celebrant's linked trigger", g.Stack.Size())
	}
	movePermanentToZone(g, celebrant, zone.Graveyard)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if !slices.Contains(g.Turn.ExtraPhases, game.PhaseCombat) {
		t.Fatalf("extra phases = %v, want an additional combat although Combat Celebrant died", g.Turn.ExtraPhases)
	}
	if tappedOwn.Tapped || !tappedOpponent.Tapped {
		t.Fatalf("own tapped=%v opponent tapped=%v, want only your other creatures untapped", tappedOwn.Tapped, tappedOpponent.Tapped)
	}

	g2 := newEnlistCombatGame()
	again := addCombatPermanent(g2, game.Player1, combatCelebrantFixture())
	policy2 := &chosenTapPolicy{pick: exertThenPick(nil, again)}
	agents[game.Player1] = policy2
	declareExertAttack(t, g2, engine, game.Player1, agents, game.Player2, again)
	if !again.ExertedBy[game.Player1] {
		t.Fatal("Combat Celebrant was not exerted")
	}
	clone := g2.Clone()
	for _, state := range []*game.Game{g2, clone} {
		exerted, _ := permanentByObjectID(state, again.ObjectID)
		exerted.Tapped = false
		state.Combat = &game.CombatState{}
		recorder := &chosenTapPolicy{pick: exertThenPick(nil, exerted)}
		var cloneAgents [game.NumPlayers]PlayerAgent
		cloneAgents[game.Player1] = recorder
		declareExertAttack(t, state, engine, game.Player1, cloneAgents, game.Player2, exerted)
		if requests := recorder.tapRequests(); len(requests) != 0 {
			t.Fatalf("exert requests = %+v, want none for a Combat Celebrant exerted this turn", requests)
		}
	}

	g2.Turn.TurnNumber++
	g2.Combat = &game.CombatState{}
	again.Tapped = false
	nextTurn := &chosenTapPolicy{pick: exertThenPick(nil)}
	agents[game.Player1] = nextTurn
	declareExertAttack(t, g2, engine, game.Player1, agents, game.Player2, again)
	if requests := nextTurn.tapRequests(); len(requests) != 1 {
		t.Fatalf("exert requests on a later turn = %+v, want Combat Celebrant offered again", requests)
	}
}

// celebrantTurnAgent declares the planned attackers against Player2 in each
// successive combat and exerts the planned creatures, recording each combat's
// exert request.
type celebrantTurnAgent struct {
	attacks  [][]*game.Permanent
	exerts   [][]*game.Permanent
	combat   int
	requests map[int][]game.ChoiceRequest
}

func (a *celebrantTurnAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	if a.combat >= len(a.attacks) {
		return action.Pass()
	}
	for _, act := range legal {
		payload, ok := act.DeclareAttackersPayload()
		if !ok || len(payload.Attackers) != len(a.attacks[a.combat]) {
			continue
		}
		matches := true
		for i, declaration := range payload.Attackers {
			if declaration.Target.Player != game.Player2 || declaration.Target.PlaneswalkerID != 0 ||
				!slices.ContainsFunc(a.attacks[a.combat], func(p *game.Permanent) bool { return p.ObjectID == payload.Attackers[i].Attacker }) {
				matches = false
				break
			}
		}
		if matches {
			a.combat++
			return act
		}
	}
	return action.Pass()
}

func (a *celebrantTurnAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoicePayment && a.combat > 0 {
		a.requests[a.combat-1] = append(a.requests[a.combat-1], request)
		return pickExertOptions(a.exerts[a.combat-1]...)(request)
	}
	return fallbackChoice(request)
}

func TestCombatCelebrantsChainAdditionalCombatsThroughTheTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatPermanent(g, game.Player1, combatCelebrantFixture())
	second := addCombatPermanent(g, game.Player1, combatCelebrantFixture())
	brawler := addCombatPermanent(g, game.Player1, hoodedBrawlerFixture())
	stockLibrary(g, game.Player1, 6)
	startingLife := g.Players[game.Player2].Life
	agent := &celebrantTurnAgent{
		attacks:  [][]*game.Permanent{{first}, {second}, {first, brawler}},
		exerts:   [][]*game.Permanent{{first}, {second}, {first, brawler}},
		requests: map[int][]game.ChoiceRequest{},
	}
	agents := simPassPolicies()
	agents[game.Player1] = agent

	log := engine.runTurn(g, agents)

	want := []game.Phase{game.PhaseBeginning, game.PhasePrecombatMain, game.PhaseCombat, game.PhaseCombat, game.PhaseCombat, game.PhasePostcombatMain, game.PhaseEnding}
	if got := phaseOrder(log); !slices.Equal(got, want) {
		t.Fatalf("phases = %v, want %v", got, want)
	}
	if agent.combat != 3 {
		t.Fatalf("declared %d planned attacks, want 3", agent.combat)
	}
	wantOffered := [][]id.ID{
		{first.CardInstanceID},
		{second.CardInstanceID},
		{brawler.CardInstanceID},
	}
	for combat, offered := range wantOffered {
		requests := agent.requests[combat]
		if len(requests) != 1 || !slices.Equal(optionCardIDs(requests[0]), offered) {
			t.Fatalf("combat %d exert requests = %+v, want options %v", combat+1, requests, offered)
		}
	}
	if !first.ExertedBy[game.Player1] || !second.ExertedBy[game.Player1] || !brawler.ExertedBy[game.Player1] {
		t.Fatal("a planned exert did not happen")
	}
	if got, wantDamage := startingLife-g.Players[game.Player2].Life, 4+4+4+5; got != wantDamage {
		t.Fatalf("Player2 took %d combat damage, want %d", got, wantDamage)
	}
}
