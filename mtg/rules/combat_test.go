package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCombatPhaseVisitsPriorityStepsInOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// With an attacker declared, all five combat steps occur in order.
	addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	engine := NewEngine(nil)
	recorder := &firstLegalStepRecorder{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: recorder,
		game.Player2: recorder,
		game.Player3: recorder,
		game.Player4: recorder,
	}

	engine.runCombatPhase(g, agents, &TurnLog{})

	want := []game.Step{
		game.StepBeginningOfCombat,
		game.StepDeclareAttackers,
		game.StepDeclareBlockers,
		game.StepCombatDamage,
		game.StepEndOfCombat,
	}
	if !slices.Equal(recorder.firstVisits, want) {
		t.Fatalf("visited combat steps = %v, want %v", recorder.firstVisits, want)
	}
}

// TestCombatPhaseSkipsBlockerAndDamageStepsWithoutAttackers covers CR 508.8: when
// no creature is declared as an attacker (and none is put onto the battlefield
// attacking), the declare blockers and combat damage steps are skipped and the
// phase proceeds directly from declare attackers to end of combat.
func TestCombatPhaseSkipsBlockerAndDamageStepsWithoutAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	recorder := &combatStepRecorder{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: recorder,
		game.Player2: recorder,
		game.Player3: recorder,
		game.Player4: recorder,
	}

	engine.runCombatPhase(g, agents, &TurnLog{})

	want := []game.Step{
		game.StepBeginningOfCombat,
		game.StepDeclareAttackers,
		game.StepEndOfCombat,
	}
	if !slices.Equal(recorder.firstVisits, want) {
		t.Fatalf("visited combat steps = %v, want %v (declare blockers and combat damage skipped)", recorder.firstVisits, want)
	}
}

func TestCombatPhaseInitializesAndClearsCombatState(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	recorder := &combatStateRecorder{game: g}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: recorder,
		game.Player2: recorder,
		game.Player3: recorder,
		game.Player4: recorder,
	}

	engine.runCombatPhase(g, agents, &TurnLog{})

	if !recorder.sawCombatState {
		t.Fatal("agent never observed initialized combat state during combat")
	}
	if g.Combat != nil {
		t.Fatalf("combat state after combat = %+v, want nil", g.Combat)
	}
}

func TestCombatPhasePriorityWindowsPassThroughWithoutActions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if g.Turn.Phase != game.PhaseCombat {
		t.Fatalf("phase = %v, want %v", g.Turn.Phase, game.PhaseCombat)
	}
	if g.Turn.Step != game.StepEndOfCombat {
		t.Fatalf("step = %v, want %v", g.Turn.Step, game.StepEndOfCombat)
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
	passCount := 0
	declareAttackersCount := 0
	for _, logged := range log.Actions {
		switch logged.Action.Kind {
		case action.ActionPass:
			passCount++
		case action.ActionDeclareAttackers:
			declareAttackersCount++
		default:
			t.Fatalf("logged action kind = %v, want pass or declare attackers", logged.Action.Kind)
		}
	}
	// CR 508.8: with no attackers, the declare blockers and combat damage steps
	// are skipped, so only the beginning of combat, declare attackers, and end of
	// combat steps open a priority window.
	if passCount != game.NumPlayers*3 {
		t.Fatalf("logged pass actions = %d, want %d", passCount, game.NumPlayers*3)
	}
	if declareAttackersCount != 1 {
		t.Fatalf("logged declare attackers actions = %d, want 1", declareAttackersCount)
	}
}

func TestCombatPhaseSkipsFirstStrikeStepWithoutFirstOrDoubleStrike(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	recorder := &combatStepRecorder{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: recorder,
		game.Player2: recorder,
		game.Player3: recorder,
		game.Player4: recorder,
	}

	engine.runCombatPhase(g, agents, &TurnLog{})

	if slices.Contains(recorder.firstVisits, game.StepFirstStrikeDamage) {
		t.Fatalf("visited steps = %v, want no first-strike damage step", recorder.firstVisits)
	}
}

func TestCleanupStepClearsMarkedDamageOnSurvivors(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	survivor := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	survivor.MarkedDamage = 2
	survivor.MarkedDeathtouchDamage = true
	engine := NewEngine(nil)

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if survivor.MarkedDamage != 0 {
		t.Fatalf("marked damage after cleanup = %d, want 0", survivor.MarkedDamage)
	}
	if survivor.MarkedDeathtouchDamage {
		t.Fatal("marked deathtouch damage was not cleared during cleanup")
	}
}

func TestCleanupStepPreservesMarkedDamageOnPhasedOutPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	permanent.PhasedOut = true
	permanent.MarkedDamage = 2
	permanent.MarkedDeathtouchDamage = true
	permanent.TemporaryPowerModifier = 2
	permanent.TemporaryToughnessModifier = 2
	permanent.RegenerationShields = 1

	NewEngine(nil).runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if permanent.MarkedDamage != 2 || !permanent.MarkedDeathtouchDamage {
		t.Fatalf("phased-out damage after cleanup = %d/deathtouch:%v, want retained", permanent.MarkedDamage, permanent.MarkedDeathtouchDamage)
	}
	if permanent.TemporaryPowerModifier != 0 ||
		permanent.TemporaryToughnessModifier != 0 ||
		permanent.RegenerationShields != 0 {
		t.Fatal("turn-based modifiers did not expire on phased-out permanent")
	}
}

func TestCleanupStepDiscardsActivePlayerToMaximumHandSize(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	var cards []id.ID
	for i := range 10 {
		cards = append(cards, addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: string(rune('A' + i))}}))
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if got := g.Players[game.Player1].Hand.Size(); got != maximumHandSize {
		t.Fatalf("hand size = %d, want %d", got, maximumHandSize)
	}
	for _, cardID := range cards[:3] {
		if !g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatalf("oldest overflow card %v was not discarded", cardID)
		}
	}
	for _, cardID := range cards[3:] {
		if !g.Players[game.Player1].Hand.Contains(cardID) {
			t.Fatalf("card %v should have remained in hand", cardID)
		}
	}
}

func TestCleanupStepDoesNotDiscardAtOrBelowMaximumHandSize(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for i := range maximumHandSize {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: string(rune('A' + i))}})
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if got := g.Players[game.Player1].Hand.Size(); got != maximumHandSize {
		t.Fatalf("hand size = %d, want %d", got, maximumHandSize)
	}
	if got := g.Players[game.Player1].Graveyard.Size(); got != 0 {
		t.Fatalf("graveyard size = %d, want 0", got)
	}
}

func TestCleanupStepSkipsDiscardWhenControllerHasNoMaximumHandSize(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "No Max Hand",
		Types:           []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{game.NoMaximumHandSizeStaticBody},
	}})
	var cards []id.ID
	for i := range 10 {
		cards = append(cards, addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: string(rune('A' + i))}}))
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if got := g.Players[game.Player1].Hand.Size(); got != len(cards) {
		t.Fatalf("hand size = %d, want %d", got, len(cards))
	}
	if got := g.Players[game.Player1].Graveyard.Size(); got != 0 {
		t.Fatalf("graveyard size = %d, want 0", got)
	}
}

func TestCleanupStepDiscardsWhenOnlyOpponentHasNoMaximumHandSize(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// The static belongs to Player2, so it must not exempt the active Player1.
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:            "No Max Hand",
		Types:           []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{game.NoMaximumHandSizeStaticBody},
	}})
	for i := range 10 {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: string(rune('A' + i))}})
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if got := g.Players[game.Player1].Hand.Size(); got != maximumHandSize {
		t.Fatalf("hand size = %d, want %d", got, maximumHandSize)
	}
}

type combatStepRecorder struct {
	firstVisits []game.Step
	seen        map[game.Step]bool
}

func (r *combatStepRecorder) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if r.seen == nil {
		r.seen = make(map[game.Step]bool)
	}
	if obs.Turn.Phase == game.PhaseCombat && !r.seen[obs.Turn.Step] {
		r.seen[obs.Turn.Step] = true
		r.firstVisits = append(r.firstVisits, obs.Turn.Step)
	}
	return action.Pass()
}

// firstLegalStepRecorder records the first time it is asked to act in each combat
// step (like combatStepRecorder) but takes the first legal action rather than
// always passing, so it will declare an attacker when one is available.
type firstLegalStepRecorder struct {
	firstVisits []game.Step
	seen        map[game.Step]bool
}

func (r *firstLegalStepRecorder) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if r.seen == nil {
		r.seen = make(map[game.Step]bool)
	}
	if obs.Turn.Phase == game.PhaseCombat && !r.seen[obs.Turn.Step] {
		r.seen[obs.Turn.Step] = true
		r.firstVisits = append(r.firstVisits, obs.Turn.Step)
	}
	if len(legal) == 0 {
		return action.Pass()
	}
	return legal[0]
}

type combatStateRecorder struct {
	game           *game.Game
	sawCombatState bool
}

func (r *combatStateRecorder) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if obs.Turn.Phase == game.PhaseCombat && r.game.Combat != nil {
		r.sawCombatState = true
	}
	return action.Pass()
}

func addCombatCreaturePermanent(g *game.Game, controller game.PlayerID, keywords ...game.Keyword) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Combat Creature",
		Types: []types.Card{
			types.Creature,
		},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(keywords...),
		}}},
	})
}

func addCombatCreaturePermanentWithPower(g *game.Game, controller game.PlayerID, power int, keywords ...game.Keyword) *game.Permanent {
	pt := game.PT{Value: power}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Powered Combat Creature",
		Types: []types.Card{
			types.Creature,
		},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(keywords...),
		}}},
	})
}

func blockedCombat(attacker, blocker *game.Permanent) *game.CombatState {
	return &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: blocker.Controller}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}
}

func addCombatPermanent(g *game.Game, controller game.PlayerID, def *game.CardDef) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   def,
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func declareAttackersActionsContainTarget(actions []action.Action, attacker id.ID, target game.AttackTarget) bool {
	for _, act := range actions {
		payload, ok := act.DeclareAttackersPayload()
		if !ok {
			continue
		}
		for _, declaration := range payload.Attackers {
			if declaration.Attacker == attacker && declaration.Target == target {
				return true
			}
		}
	}
	return false
}

func mustDeclareAttackersPayload(t *testing.T, act action.Action) action.DeclareAttackersAction {
	t.Helper()
	payload, ok := act.DeclareAttackersPayload()
	if !ok {
		t.Fatalf("DeclareAttackersPayload() ok = false for %+v", act)
	}
	return payload
}

func mustDeclareBlockersPayload(t *testing.T, act action.Action) action.DeclareBlockersAction {
	t.Helper()
	payload, ok := act.DeclareBlockersPayload()
	if !ok {
		t.Fatalf("DeclareBlockersPayload() ok = false for %+v", act)
	}
	return payload
}

func intPtr(value int) *int {
	return new(value)
}

func permanentIDs(permanents []*game.Permanent) []id.ID {
	ids := make([]id.ID, 0, len(permanents))
	for _, permanent := range permanents {
		ids = append(ids, permanent.ObjectID)
	}
	return ids
}
