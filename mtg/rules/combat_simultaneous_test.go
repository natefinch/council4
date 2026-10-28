package rules

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCombatSimultaneousShieldMultiBlock(t *testing.T) {
	for _, shields := range []int{0, 1, 2} {
		t.Run(strconv.Itoa(shields), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
			first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
			second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
			attacker.Counters.Add(counter.Shield, shields)
			g.Combat = &game.CombatState{
				Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
				Blockers: []game.BlockDeclaration{
					{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
					{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
				},
			}

			NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

			wantDamage := 0
			if shields == 0 {
				wantDamage = 4
			}
			if attacker.MarkedDamage != wantDamage || attacker.Counters.Get(counter.Shield) != max(0, shields-1) {
				t.Fatalf("marked damage/shields = %d/%d, want %d/%d",
					attacker.MarkedDamage, attacker.Counters.Get(counter.Shield), wantDamage, max(0, shields-1))
			}
		})
	}
}

// TestCombatSimultaneousAssignmentsPrecedeWither: the trampler's lethal damage
// counts the 3 the wither creature is assigning to the 5/5 in the same step
// (CR 702.19b) but not the -1/-1 counters that damage will become, so it
// assigns 2 to the blocker and tramples 1 over.
func TestCombatSimultaneousAssignmentsPrecedeWither(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Wither)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Trample)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind: game.RuleEffectCanBlockAdditional, SourceObjectID: blocker.ObjectID, AffectedSource: true, AdditionalBlockCount: 1,
	})
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: first.ObjectID},
			{Blocker: blocker.ObjectID, Blocking: second.ObjectID},
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})
	if g.Players[game.Player2].Life != 39 || blocker.MarkedDamage != 2 || blocker.Counters.Get(counter.MinusOneMinusOne) != 3 {
		t.Fatalf("life/marked damage/counters = %d/%d/%d, want 39/2/3",
			g.Players[game.Player2].Life, blocker.MarkedDamage, blocker.Counters.Get(counter.MinusOneMinusOne))
	}
}

func TestCombatSimultaneousLifelinkDoesNotChangeAnotherSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 37
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Lifelink)
	second := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Conditional lifelink", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 3}), Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{{
			Condition: opt.Val(game.Condition{Aggregates: []game.AggregateComparison{
				{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: 40},
			}}),
			ContinuousEffects: []game.ContinuousEffect{{Layer: game.LayerAbility, AffectedSource: true, AddKeywords: []game.Keyword{game.Lifelink}}},
		}},
	}})
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}
	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})
	if g.Players[game.Player1].Life != 40 || g.Players[game.Player2].Life != 34 {
		t.Fatalf("life = %d/%d, want 40/34", g.Players[game.Player1].Life, g.Players[game.Player2].Life)
	}
}

func TestCombatSimultaneousFinitePreventionAllocation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}
	g.PreventionShields = append(g.PreventionShields, game.PreventionShield{
		ID: g.IDGen.Next(), Player: game.Player2, Amount: 3,
	})
	allocations := 0
	policy := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind != game.ChoicePreventionAllocation {
			return request.DefaultSelection
		}
		if obs.Player != game.Player2 || g.Players[game.Player2].Life != 40 {
			t.Fatal("prevention chooser observed damage already committed")
		}
		want := second.CardInstanceID
		if allocations == 0 {
			want = first.CardInstanceID
		}
		allocations++
		for _, option := range request.Options {
			if option.Card.Exists && option.Card.Val.CardID == want {
				return []int{option.Index}
			}
		}
		t.Fatal("missing source in prevention allocation")
		return nil
	}}
	engine := NewEngine(nil)
	log := &TurnLog{}
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player2: policy}, log)
	defer g.ClearChoiceContext()
	engine.resolveCombatDamage(g, log)
	if allocations != 3 || g.Players[game.Player2].Life != 38 || len(log.CombatDamage) != 1 ||
		log.CombatDamage[0].SourceID != first.CardInstanceID || log.CombatDamage[0].Damage != 2 {
		t.Fatalf("allocation/source outcome = %d/%d/%+v, want 3/38/first source 2", allocations, g.Players[game.Player2].Life, log.CombatDamage)
	}
}

func TestCombatSimultaneousPreventionChoicesAreAPNAP(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Combat = &game.CombatState{}
	for _, player := range []game.PlayerID{game.Player3, game.Player2} {
		for range 2 {
			source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{
				Attacker: source.ObjectID, Target: game.AttackTarget{Player: player},
			})
		}
		g.PreventionShields = append(g.PreventionShields, game.PreventionShield{ID: g.IDGen.Next(), Player: player, Amount: 1})
	}
	var choosers []game.PlayerID
	policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoicePreventionAllocation {
			choosers = append(choosers, request.Player)
		}
		return request.DefaultSelection
	}}
	engine := NewEngine(nil)
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player2: policy, game.Player3: policy}, &TurnLog{})
	defer g.ClearChoiceContext()
	engine.resolveCombatDamage(g, &TurnLog{})
	if !reflect.DeepEqual(choosers, []game.PlayerID{game.Player2, game.Player3}) {
		t.Fatalf("prevention chooser order = %v, want Player2 then Player3", choosers)
	}
}

func TestCombatSimultaneousRedirectionPreservesWholeEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	protector := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	protector.Counters.Add(counter.Shield, 1)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind: game.RuleEffectRedirectDamageToSource, Controller: game.Player2,
		AffectedPlayer: game.PlayerYou, SourceObjectID: protector.ObjectID,
	})
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}
	log := &TurnLog{}
	NewEngine(nil).resolveCombatDamage(g, log)
	if protector.MarkedDamage != 0 || protector.Counters.Get(counter.Shield) != 0 ||
		g.Players[game.Player2].Life != 40 || len(log.CombatDamage) != 0 {
		t.Fatalf("redirected simultaneous damage = %d, shields=%d, life=%d, logs=%+v",
			protector.MarkedDamage, protector.Counters.Get(counter.Shield), g.Players[game.Player2].Life, log.CombatDamage)
	}
}

func TestCombatSimultaneousOneShotChoosesOneSource(t *testing.T) {
	for _, reflectDamage := range []bool{false, true} {
		t.Run(strconv.FormatBool(reflectDamage), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			first := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
			second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			shieldSource := addColoredSourceCard(g, game.Player2, color.White)
			g.PreventionShields = append(g.PreventionShields, game.PreventionShield{
				All: true, OneShot: true, Player: game.Player2, Controller: game.Player2,
				RedirectToSourceController: reflectDamage, RedirectSourceID: shieldSource,
			})
			g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
				{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
				{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			}}
			policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				if request.Kind == game.ChoicePreventionAllocation {
					for _, option := range request.Options {
						if option.Card.Exists && option.Card.Val.CardID == second.CardInstanceID {
							return []int{option.Index}
						}
					}
					t.Fatal("second source was not offered")
				}
				return request.DefaultSelection
			}}
			engine := NewEngine(nil)
			engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player2: policy}, &TurnLog{})
			defer g.ClearChoiceContext()
			engine.resolveCombatDamage(g, &TurnLog{})
			wantAttackerLife := 40
			if reflectDamage {
				wantAttackerLife = 38
			}
			if g.Players[game.Player2].Life != 37 || g.Players[game.Player1].Life != wantAttackerLife || len(g.PreventionShields) != 0 {
				t.Fatalf("one-shot result = life %d/%d, shields %d; want %d/37, shields 0",
					g.Players[game.Player1].Life, g.Players[game.Player2].Life, len(g.PreventionShields), wantAttackerLife)
			}
		})
	}
}

func TestCombatSimultaneousGlobalPreventionPreservesEachChooser(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	attacker.Counters.Add(counter.Shield, 1)
	g.PreventionShields = append(g.PreventionShields, game.PreventionShield{All: true, Global: true})
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers: []game.BlockDeclaration{
			{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
			{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
		},
	}
	choseFog := false
	policy := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		for _, option := range request.Options {
			if strings.Contains(option.Label, "prevention shield") {
				choseFog = true
				return []int{option.Index}
			}
		}
		return request.DefaultSelection
	}}
	engine := NewEngine(nil)
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player1: policy}, &TurnLog{})
	defer g.ClearChoiceContext()
	engine.resolveCombatDamage(g, &TurnLog{})
	if !choseFog || attacker.Counters.Get(counter.Shield) != 1 || attacker.MarkedDamage != 0 {
		t.Fatalf("Fog choice = %v, shields=%d, damage=%d; want true/1/0",
			choseFog, attacker.Counters.Get(counter.Shield), attacker.MarkedDamage)
	}
}

func TestCombatSimultaneousShieldsDoNotLeakAcrossDamageEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.DoubleStrike)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.DoubleStrike)
	target.Counters.Add(counter.Shield, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: target.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers: []game.BlockDeclaration{
			{Blocker: first.ObjectID, Blocking: target.ObjectID},
			{Blocker: second.ObjectID, Blocking: target.ObjectID},
		},
	}
	engine := NewEngine(nil)
	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
	if target.MarkedDamage != 0 || target.Counters.Get(counter.Shield) != 1 {
		t.Fatal("first-strike damage did not consume exactly one shield")
	}
	engine.resolveCombatDamagePass(g, normalCombatDamage, &TurnLog{})
	if target.MarkedDamage != 0 || target.Counters.Get(counter.Shield) != 0 {
		t.Fatal("normal damage did not consume its own shield")
	}
	target.Counters.Add(counter.Shield, 1)
	if dealPermanentDamage(g, first.CardInstanceID, first.ObjectID, game.Player2, target, 2, true) != 0 ||
		dealPermanentDamage(g, second.CardInstanceID, second.ObjectID, game.Player2, target, 2, true) != 2 {
		t.Fatal("independent damage calls were incorrectly combined")
	}
}

type simultaneousCombatPolicy struct {
	attacker id.ID
	defender game.PlayerID
}

func (policy simultaneousCombatPolicy) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	for _, act := range legal {
		if attackers, ok := act.DeclareAttackersPayload(); ok && len(attackers.Attackers) == 1 &&
			attackers.Attackers[0].Attacker == policy.attacker && attackers.Attackers[0].Target.Player == policy.defender {
			return act
		}
		if blockers, ok := act.DeclareBlockersPayload(); ok && len(blockers.Blockers) == 2 {
			return act
		}
	}
	return action.Pass()
}

func (simultaneousCombatPolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	return request.DefaultSelection
}

func TestCombatSimultaneousShieldThroughLiveAndSimulator(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5, game.Lifelink)
	addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	attacker.Counters.Add(counter.Shield, 1)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	policies := simPassPolicies()
	policies[game.Player1] = simultaneousCombatPolicy{attacker: attacker.ObjectID, defender: game.Player2}
	policies[game.Player2] = simultaneousCombatPolicy{}
	declare := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}
	first, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declare, policies)
	if !ok {
		t.Fatal("Simulator rejected legal multi-block fixture")
	}
	second, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declare, policies)
	if !ok || !reflect.DeepEqual(first.Events, second.Events) {
		t.Fatal("equal Simulator choices did not reproduce events")
	}
	live := g.Clone()
	engine := NewEngine(nil)
	log := &TurnLog{}
	engine.setReplacementChoiceContext(live, policies, log)
	engine.runCombatPhase(live, policies, log)
	live.ClearChoiceContext()
	for _, state := range []*game.Game{first, second, live} {
		survivor, found := permanentByObjectID(state, attacker.ObjectID)
		if !found || survivor.MarkedDamage != 0 || survivor.Counters.Get(counter.Shield) != 0 ||
			state.Players[game.Player1].Life != 45 || len(state.Battlefield) != 1 {
			t.Fatalf("multi-block result incorrect: survivor=%+v life=%d battlefield=%d", survivor, state.Players[game.Player1].Life, len(state.Battlefield))
		}
		gains := 0
		batch := id.ID(0)
		for _, event := range state.Events {
			if event.Kind == game.EventDamageDealt || event.Kind == game.EventDamagePrevented {
				if batch == 0 {
					batch = event.SimultaneousID
				}
				if event.SimultaneousID == 0 || event.SimultaneousID != batch {
					t.Fatal("damage/prevention facts do not share the real damage event")
				}
			}
			if event.Kind == game.EventLifeGained {
				gains++
				if event.Amount != 5 || event.SimultaneousID != batch {
					t.Fatal("lifelink did not preserve the damage batch")
				}
			}
		}
		if gains != 1 {
			t.Fatalf("lifelink gains = %d, want one per source", gains)
		}
	}
	permanent, _ := permanentByObjectID(first, attacker.ObjectID)
	permanent.Counters.Add(counter.Shield, 4)
	original, _ := permanentByObjectID(g, attacker.ObjectID)
	other, _ := permanentByObjectID(second, attacker.ObjectID)
	if original.Counters.Get(counter.Shield) != 1 || other.Counters.Get(counter.Shield) != 0 {
		t.Fatal("damage or later mutation leaked between clones")
	}
}
