package rules

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestCompleteAttackDeclarationOutsidePolicySamples(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	var creatures []*game.Permanent
	for range 3 {
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
		creature.SummoningSick = false
		creatures = append(creatures, creature)
	}
	want := []game.AttackDeclaration{
		{Attacker: creatures[0].ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: creatures[2].ObjectID, Target: game.AttackTarget{Player: game.Player3}},
	}
	engine := NewEngine(nil)
	ce := combatEngine{engine}
	proposal := action.DeclareAttackers(want)
	if containsAction(ce.legalAttackers(g, game.Player1), proposal) {
		t.Fatal("fixture no longer distinguishes a full declaration from representative candidates")
	}
	truth := g.Clone()
	if !ce.applyAttackers(truth, game.Player1, action.DeclareAttackersAction{Attackers: want}) {
		t.Fatal("mixed-defender attack fixture is not a legal complete declaration")
	}
	agents := simPassPolicies()
	agents[game.Player1] = &exactNumericPolicy{announcement: proposal}
	ce.declareAttackers(g, agents, &TurnLog{})
	if !reflect.DeepEqual(g.Combat.Attackers, want) || creatures[1].Tapped || !creatures[0].Tapped || !creatures[2].Tapped {
		t.Fatal("live combat rejected an exact legal subset/defender distribution outside policy samples")
	}
}

func TestCompleteBlockDeclarationOutsidePolicyMatchings(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
	g.Combat = &game.CombatState{}
	var attackers, blockers []*game.Permanent
	for range 4 {
		attacker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Must be blocked", Types: []types.Card{types.Creature},
			StaticAbilities: []game.StaticAbility{game.MustBeBlockedStaticBody},
		}})
		attackers = append(attackers, attacker)
		blockers = append(blockers, addCombatCreaturePermanentWithPower(g, game.Player2, 3))
		g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{
			Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
		})
	}
	engine := NewEngine(nil)
	ce := combatEngine{engine}
	candidates := ce.legalBlockers(g, game.Player2)
	var want []game.BlockDeclaration
	var walk func([]int, []bool)
	walk = func(order []int, used []bool) {
		if want != nil {
			return
		}
		if len(order) == len(blockers) {
			var proposed []game.BlockDeclaration
			for index, attacker := range order {
				proposed = append(proposed, game.BlockDeclaration{Blocker: blockers[index].ObjectID, Blocking: attackers[attacker].ObjectID})
			}
			sampled := false
			for _, candidate := range candidates {
				payload, present := candidate.DeclareBlockersPayload()
				if !present || len(payload.Blockers) != len(proposed) {
					continue
				}
				same := true
				for _, pair := range proposed {
					same = same && slices.Contains(payload.Blockers, pair)
				}
				sampled = sampled || same
			}
			if !sampled {
				truth := g.Clone()
				if ce.applyBlockers(truth, game.Player2, action.DeclareBlockersAction{Blockers: proposed}) {
					want = proposed
				}
			}
			return
		}
		for index := range attackers {
			if !used[index] {
				used[index] = true
				walk(append(order, index), used)
				used[index] = false
			}
		}
	}
	walk(nil, make([]bool, len(attackers)))
	if want == nil {
		t.Fatal("fixture did not find a distinct legal complete matching outside representative samples")
	}
	agents := simPassPolicies()
	agents[game.Player2] = &exactNumericPolicy{announcement: action.DeclareBlockers(want)}
	ce.declareBlockers(g, agents, &TurnLog{})
	if !reflect.DeepEqual(g.Combat.Blockers, want) {
		t.Fatal("live combat rejected a legal complete matching outside policy samples")
	}
}

type stagedCombatPolicy struct {
	t          *testing.T
	proposal   action.Action
	attacks    []opt.V[game.AttackDeclaration]
	blocks     []game.BlockDeclaration
	attackStep int
	oracle     [][]game.BlockDeclaration
	fixed      map[game.BlockDeclaration]bool
}

func (p *stagedCombatPolicy) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return p.proposal
}

func (p *stagedCombatPolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceCombatDeclaration {
		return request.DefaultSelection
	}
	if p.proposal.Kind == action.ActionDeclareAttackers {
		want := p.attacks[p.attackStep]
		p.attackStep++
		for _, option := range request.Options {
			if option.Attack == want {
				return []int{option.Index}
			}
		}
	} else {
		if p.fixed == nil {
			p.fixed = make(map[game.BlockDeclaration]bool)
		}
		edge := game.BlockDeclaration{Blocker: request.CombatBlocker, Blocking: request.CombatAttacker}
		for _, option := range request.Options {
			p.fixed[edge] = option.Block.Exists
			possible := slices.ContainsFunc(p.oracle, func(candidate []game.BlockDeclaration) bool {
				for fixed, included := range p.fixed {
					if slices.Contains(candidate, fixed) != included {
						return false
					}
				}
				return true
			})
			if len(p.oracle) > 0 && !possible {
				p.t.Fatalf("offered a block prefix without a globally optimum completion: %+v", request)
			}
		}
		p.fixed[edge] = slices.Contains(p.blocks, edge)
		for _, option := range request.Options {
			if option.Block.Exists && slices.Contains(p.blocks, option.Block.Val) {
				return []int{option.Index}
			}
		}
		for _, option := range request.Options {
			if !option.Block.Exists {
				return []int{option.Index}
			}
		}
	}
	p.t.Fatalf("desired legal declaration prefix was not offered: %+v", request)
	return nil
}

func TestStagedAttackDeclarationsExhaustiveMixedDefenders(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	for range 3 {
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
		creature.SummoningSick = false
	}
	for _, kind := range []types.Card{types.Planeswalker, types.Battle} {
		addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name: "Permanent defender", Types: []types.Card{kind},
		}})
	}
	engine := NewEngine(nil)
	problem := newAttackDeclarationProblem(g, game.Player1)
	examined := 0
	var visit func(int, []opt.V[game.AttackDeclaration], []game.AttackDeclaration)
	visit = func(index int, choices []opt.V[game.AttackDeclaration], declarations []game.AttackDeclaration) {
		if index != len(problem.domains) {
			for _, choice := range problem.domains[index] {
				next := slices.Clone(declarations)
				var selected opt.V[game.AttackDeclaration]
				if choice.attacking {
					next, selected = append(next, choice.declaration), opt.Val(choice.declaration)
				}
				visit(index+1, append(slices.Clone(choices), selected), next)
			}
			return
		}
		examined++
		proposal := action.DeclareAttackersAction{Attackers: declarations}
		if !(combatEngine{engine}).attackersLegal(g, game.Player1, proposal) {
			t.Fatalf("ordinary subset/defender combination rejected: %+v", declarations)
		}
		policies := simPassPolicies()
		policies[game.Player1] = &stagedCombatPolicy{t: t, proposal: action.ChooseAttackers(), attacks: choices}
		result, ok := engine.Simulator().Apply(g, game.Player1, action.ChooseAttackers(), policies)
		if !ok || !slices.Equal(result.Combat.Attackers, declarations) || !result.Combat.AttackDeclarationComplete {
			t.Fatalf("staged declaration differs from exact proposal: %+v", declarations)
		}
		if g.Combat.AttackDeclarationComplete || len(g.Combat.Attackers) != 0 {
			t.Fatal("Simulator changed its source")
		}
		if _, repeated := engine.Simulator().Apply(result, game.Player1, action.ChooseAttackers(), policies); repeated {
			t.Fatal("attack window reopened during priority")
		}
	}
	visit(0, nil, nil)
	if examined != 216 {
		t.Fatalf("examined %d declarations, want all 6^3 stay/player/planeswalker/battle combinations", examined)
	}
}

func blockOracle(p *blockDeclarationProblem) [][]game.BlockDeclaration {
	var legal [][]game.BlockDeclaration
	best := -1
	for mask := 0; mask < 1<<(len(p.blockers)*len(p.attackers)); mask++ {
		var declarations []game.BlockDeclaration
		valid := true
		counts := make([]int, len(p.attackers))
		score := 0
		for blocker, source := range p.blockers {
			capacity := 0
			for attacker, target := range p.attackers {
				if mask&(1<<(blocker*len(p.attackers)+attacker)) == 0 {
					continue
				}
				capacity++
				valid = valid && p.allowed[blocker][attacker]
				counts[attacker]++
				score += p.lures[blocker][attacker]
				declarations = append(declarations, game.BlockDeclaration{Blocker: source.ObjectID, Blocking: target.ObjectID})
			}
			valid = valid && capacity <= p.capacities[blocker]
		}
		for attacker, count := range counts {
			valid = valid && (count == 0 || count >= p.minimum[attacker])
			valid = valid && (count <= 1 || !ruleEffectLimitsBlockersToOne(p.game, p.attackers[attacker]))
			if count >= p.minimum[attacker] {
				score += p.weights[attacker]
			}
		}
		valid = valid && blockDeclarationsSatisfyAloneRestriction(p.game, declarations)
		if !valid || score < best {
			continue
		}
		if score > best {
			best, legal = score, nil
		}
		legal = append(legal, declarations)
	}
	return legal
}

func TestStagedBlockDeclarationsExhaustiveGlobalRequirements(t *testing.T) {
	for _, scenario := range []string{"unit", "lure-menace", "lure-max-one", "multiple-capacity-alone"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.ActivePlayer = game.Player1
			g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
			g.Combat = &game.CombatState{}
			for index := range 3 {
				var keywords []game.Keyword
				if scenario == "lure-menace" && index == 0 {
					keywords = []game.Keyword{game.Menace}
				}
				attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, keywords...)
				g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{
					Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
				})
				addRuleEffectSource(g, game.Player1, game.RuleEffectMustBeBlocked, attacker.ObjectID)
				if index == 0 && scenario != "unit" {
					addRuleEffectSource(g, game.Player1, game.RuleEffectMustBeBlockedByAllAble, attacker.ObjectID)
				}
				if index == 0 && scenario == "lure-max-one" {
					addRuleEffectSource(g, game.Player1, game.RuleEffectCantBeBlockedByMoreThanOne, attacker.ObjectID)
				}
				var static []game.StaticAbility
				if index == 0 && scenario == "multiple-capacity-alone" {
					static = append(static, game.StaticAbility{
						RuleEffects: []game.RuleEffect{
							{Kind: game.RuleEffectCanBlockAdditional, AffectedSource: true, AdditionalBlockCount: 1},
							{Kind: game.RuleEffectCantBlockAlone, AffectedSource: true},
						},
					})
				}
				addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
					Name: "Blocker", Types: []types.Card{types.Creature},
					Power: opt.Val(game.PT{Value: 3}), Toughness: opt.Val(game.PT{Value: 3}), StaticAbilities: static,
				}})
			}
			engine := NewEngine(nil)
			problem := newBlockDeclarationProblem(g, game.Player2)
			oracle := blockOracle(problem)
			maximum, preferred := problem.maximum()
			if len(oracle) == 0 || maximum != problem.score(oracle[0]) || problem.score(preferred) != maximum {
				t.Fatalf("global optimum differs from exhaustive oracle: %d, %+v", maximum, preferred)
			}
			for _, declarations := range oracle {
				policies := simPassPolicies()
				policies[game.Player2] = &stagedCombatPolicy{t: t, proposal: action.ChooseBlockers(), blocks: declarations, oracle: oracle}
				result, ok := engine.Simulator().Apply(g, game.Player2, action.ChooseBlockers(), policies)
				if !ok || !slices.Equal(result.Combat.Blockers, declarations) {
					t.Fatalf("could not express legal oracle declaration: %+v", declarations)
				}
				if _, repeated := engine.Simulator().Apply(result, game.Player2, action.ChooseBlockers(), policies); repeated {
					t.Fatal("blocker window reopened during priority")
				}
			}

			for mask := range 1 << (len(problem.blockers) * len(problem.attackers)) {
				var declaration []game.BlockDeclaration
				for blocker, source := range problem.blockers {
					for attacker, target := range problem.attackers {
						if mask&(1<<(blocker*len(problem.attackers)+attacker)) != 0 {
							declaration = append(declaration, game.BlockDeclaration{Blocker: source.ObjectID, Blocking: target.ObjectID})
						}
					}
				}
				want := slices.ContainsFunc(oracle, func(candidate []game.BlockDeclaration) bool { return slices.Equal(candidate, declaration) })
				if got := (combatEngine{engine}).blockersLegal(g, game.Player2, action.DeclareBlockersAction{Blockers: declaration}); got != want {
					t.Fatalf("validator disagrees with oracle for %+v: got %v want %v", declaration, got, want)
				}
			}
		})
	}
}

func TestStagedWarPrideDeclarationBoundedWork(t *testing.T) {
	const size = 16
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
	g.Combat = &game.CombatState{}
	for range size {
		attacker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "War-Pride requirement", Types: []types.Card{types.Creature},
			StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{
				{Kind: game.RuleEffectMustBeBlocked, AffectedSource: true},
				{Kind: game.RuleEffectCantBeBlockedByMoreThanOne, AffectedSource: true},
			}}},
		}})
		g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}})
		addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	}
	engine := NewEngine(nil)
	problem := newBlockDeclarationProblem(g, game.Player2)
	goal, _ := problem.maximum()
	if goal != size || !problem.unit {
		t.Fatalf("fixture did not use residual b-matching: goal %d, unit %v", goal, problem.unit)
	}
	fixed := make([]int8, size*size)
	for index := range fixed {
		fixed[index] = -1
	}
	var want []game.BlockDeclaration
	for blocker, source := range problem.blockers {
		for attacker, target := range problem.attackers {
			edge := blocker*size + attacker
			selected := int8(0)
			if attacker == size-1-blocker {
				selected = 1
				want = append(want, game.BlockDeclaration{Blocker: source.ObjectID, Blocking: target.ObjectID})
			}
			for _, decision := range []int8{0, 1} {
				fixed[edge] = decision
				_, _, found := problem.complete(fixed, goal)
				if decision == selected && !found {
					t.Fatal("residual matching rejected an extendable reversed permutation")
				}
			}
			fixed[edge] = selected
		}
	}
	if problem.states > 100000 {
		t.Fatalf("16x16 prefix matching expanded %d states, bound 100000", problem.states)
	}
	candidates := engine.Simulator().LegalActions(g, game.Player2)
	if len(candidates) > size*size+2 || !containsAction(candidates, action.ChooseBlockers()) {
		t.Fatalf("16x16 candidates = %d, want bounded samples plus staged route", len(candidates))
	}
	policies := simPassPolicies()
	policies[game.Player2] = &stagedCombatPolicy{t: t, proposal: action.ChooseBlockers(), blocks: want}
	result, ok := engine.Simulator().Apply(g, game.Player2, action.ChooseBlockers(), policies)
	if !ok || !slices.Equal(result.Combat.Blockers, want) {
		t.Fatal("large staged matching did not produce the chosen complete permutation")
	}
	t.Logf("16x16 matching: %d candidates, %d solver states", len(candidates), problem.states)
}

func TestInvalidCompleteDeclarationsAreAtomic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	creature.SummoningSick = false
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addAttackTaxPermanent(g, game.Player2, 1)
	engine := NewEngine(nil)
	before := g.Clone()
	invalid := []game.AttackDeclaration{
		{Attacker: creature.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: id.ID(999999), Target: game.AttackTarget{Player: game.Player2}},
	}
	if (combatEngine{engine}).applyAttackers(g, game.Player1, action.DeclareAttackersAction{Attackers: invalid}) {
		t.Fatal("accepted an ineligible second attacker")
	}
	if !reflect.DeepEqual(g.Combat, before.Combat) || !reflect.DeepEqual(g.Battlefield, before.Battlefield) ||
		!reflect.DeepEqual(g.Players, before.Players) || !reflect.DeepEqual(g.Events, before.Events) {
		t.Fatal("invalid declaration changed payment, tapping, events or combat")
	}
}

func TestInvalidCompleteBlockDeclarationIsAtomic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	pair := game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}
	before := g.Clone()
	if (combatEngine{NewEngine(nil)}).applyBlockers(g, game.Player2, action.DeclareBlockersAction{
		Blockers: []game.BlockDeclaration{pair, pair},
	}) {
		t.Fatal("accepted duplicate blocker pair")
	}
	if !reflect.DeepEqual(g.Combat, before.Combat) || !reflect.DeepEqual(g.Battlefield, before.Battlefield) ||
		!reflect.DeepEqual(g.Events, before.Events) {
		t.Fatal("invalid second pair partially changed blockers, events or declaration completion")
	}
}

func TestMultiBlockPolicySubsetsAreBounded(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
	g.Combat = &game.CombatState{}
	for range 24 {
		attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
		g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}})
	}
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Many-blocker", Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind: game.RuleEffectCanBlockAdditional, AffectedSource: true, AdditionalBlockCount: 23,
		}}}},
	}})
	candidates := (combatEngine{}).legalBlockers(g, game.Player2)
	if len(candidates) > 89 {
		t.Fatalf("24-attacker policy produced %d candidates, bound 24 + 64 subsets + empty", len(candidates))
	}
}

func TestVoluntaryAttackTaxDoesNotExcuseOtherRequirements(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	free := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	taxed := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	for _, requirement := range []struct {
		creature *game.Permanent
		target   game.PlayerID
	}{{free, game.Player3}, {taxed, game.Player2}} {
		g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
			ID: g.IDGen.Next(), Kind: game.RuleEffectMustAttack,
			AffectedObjectID: requirement.creature.ObjectID, RequiredAttackTarget: opt.Val(requirement.target),
		})
	}
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addAttackTaxPermanent(g, game.Player2, 1)
	engine := NewEngine(nil)
	onlyTaxed := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{{
		Attacker: taxed.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	if (combatEngine{engine}).attackersLegal(g, game.Player1, onlyTaxed) {
		t.Fatal("voluntarily paying B's tax waived A's freely satisfiable requirement")
	}
	freeDeclaration := game.AttackDeclaration{Attacker: free.ObjectID, Target: game.AttackTarget{Player: game.Player3}}
	for _, declarations := range [][]game.AttackDeclaration{{freeDeclaration}, {freeDeclaration, onlyTaxed.Attackers[0]}} {
		if !(combatEngine{engine}).attackersLegal(g, game.Player1, action.DeclareAttackersAction{Attackers: declarations}) {
			t.Fatal("tax payment was either required or failed to preserve unrelated requirements")
		}
	}
	problem := newAttackDeclarationProblem(g, game.Player1)
	fixed := make([]int, len(problem.creatures))
	for creature, domain := range problem.domains {
		for index, choice := range domain {
			if !choice.attacking && problem.creatures[creature].ObjectID == free.ObjectID ||
				choice.attacking && choice.declaration == onlyTaxed.Attackers[0] {
				fixed[creature] = index
			}
		}
	}
	if _, found := problem.complete(fixed, problem.maximum(), false); found {
		t.Fatal("staged choices offered the same invalid paid substitution")
	}
}

func TestNondefendingSeatsGetPriorityNotBlockDeclarations(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	g.Combat = &game.CombatState{
		AttackDeclarationComplete: true,
		Attackers:                 []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}
	g.Combat.BlockDeclarationsComplete[game.Player2] = true
	simulator := NewEngine(nil).Simulator()
	for player := range game.PlayerID(game.NumPlayers) {
		if containsAction(simulator.LegalActions(g, player), action.ChooseBlockers()) {
			t.Fatalf("seat %d retained a fictitious declaration window", player)
		}
		if _, ok := simulator.Apply(g, player, action.ChooseBlockers(), simPassPolicies()); ok {
			t.Fatalf("seat %d could declare blockers during priority", player)
		}
	}
	if !containsAction(simulator.LegalActions(g, game.Player1), action.Pass()) {
		t.Fatal("active player did not receive ordinary priority actions")
	}
}

func TestCombatDeclarationWindowRejectsPriorityAnnouncements(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	card := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Instant during declaration", Types: []types.Card{types.Instant},
	}})
	proposal := action.AnnounceSpell(action.CastSpellAction{CardID: card, SourceZone: zone.Hand})
	engine := NewEngine(nil)
	if !engine.announcementLegal(g, game.Player1, proposal) {
		t.Fatal("fixture needs an otherwise legal priority announcement")
	}
	policies := simPassPolicies()
	policies[game.Player1] = &exactNumericPolicy{announcement: proposal}
	samples := (combatEngine{engine}).legalAttackers(g, game.Player1)
	chosen := engine.chooseAction(g, policies, game.Player1, samples, samples[len(samples)-1], action.ActionDeclareAttackers)
	if chosen.Kind != action.ActionDeclareAttackers || g.Execution.InvalidActions != 1 || g.Execution.ActionFallbacks != 1 {
		t.Fatal("priority announcement escaped the explicit turn-based declaration window")
	}
	if _, accepted := engine.Simulator().Apply(g, game.Player1, proposal, policies); accepted {
		t.Fatal("Simulator accepted priority before attackers were declared")
	}
	g.Turn.Step = game.StepDeclareBlockers
	attacker := eligibleAttackers(g, game.Player1)[0]
	g.Combat.Attackers = []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}
	g.Combat.AttackDeclarationComplete = true
	if _, accepted := engine.Simulator().Apply(g, game.Player1, proposal, policies); accepted {
		t.Fatal("Simulator accepted priority before another seat declared blockers")
	}
}

func TestStagedControlledAttackUsesActorPaymentAndLogs(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.TurnNumber = 1
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	g.PlayerControlEffects = []game.PlayerControlEffect{{
		ID: g.IDGen.Next(), Player: game.Player1, Controller: game.Player3, Turn: 1,
	}}
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	land := addBasicLandPermanent(g, game.Player1, types.Forest)
	addAttackTaxPermanent(g, game.Player2, 1)
	choiceCalls := 0
	policy := &controlScript{
		action: func(obs PlayerObservation, legal []action.Action) action.Action {
			if obs.Player != game.Player1 || obs.DecisionMaker != game.Player3 || !containsAction(legal, action.ChooseAttackers()) {
				t.Fatal("controlled declaration dispatch lost actor, decision maker or staged route")
			}
			return action.ChooseAttackers()
		},
		choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
			choiceCalls++
			if obs.Player != game.Player1 || obs.DecisionMaker != game.Player3 || request.Player != game.Player1 {
				t.Fatal("controlled staged decision used the wrong seat")
			}
			for _, option := range request.Options {
				if option.Attack.Exists && option.Attack.Val.Target.Player == game.Player2 {
					return []int{option.Index}
				}
			}
			t.Fatal("affordable voluntarily taxed attack missing")
			return nil
		},
	}
	policies := [game.NumPlayers]PlayerAgent{game.Player3: policy}
	engine := NewEngine(nil)
	simulated, ok := engine.Simulator().Apply(g, game.Player1, action.ChooseAttackers(), policies)
	if !ok {
		t.Fatal("Simulator rejected controlled staged attack")
	}
	simulatedLand, exists := permanentByObjectID(simulated, land.ObjectID)
	if !exists || !simulatedLand.Tapped || land.Tapped || attacker.Tapped {
		t.Fatal("Simulator taxed attack changed its source or failed actor payment")
	}
	log := &TurnLog{}
	(combatEngine{engine}).declareAttackers(g, policies, log)
	if !land.Tapped || !attacker.Tapped || choiceCalls != 2 || len(log.Choices) != 1 ||
		log.Choices[0].DecisionMaker != game.Player3 || !slices.Equal(g.Combat.Attackers, simulated.Combat.Attackers) {
		t.Fatal("live and simulated controlled declarations or actor payment diverged")
	}
}
