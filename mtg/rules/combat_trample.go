package rules

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// trampleStep makes the trampling attackers' combat damage assignments of one
// combat damage step as one joint assignment (CR 510.1, 702.19b). A trampler
// may assign damage to the attack target only once every creature blocking it
// is assigned lethal damage, counting damage already marked and the damage all
// other creatures assign in the same step. Tramplers that share a blocker can
// therefore each rely on the others' damage to it. Their assignments are made
// one at a time, and each is offered exactly the choices that the attackers
// still to assign can complete into a legal joint assignment, so the legal
// assignments don't depend on the order attackers were declared. An attacker
// that relied on a later one's damage commits that later attacker to assign it.
type trampleStep struct {
	// assigned is the combat damage assigned to each permanent so far.
	assigned map[id.ID]assignedCombatDamage
	// required holds the blockers an attacker that assigned excess damage
	// requires to be assigned lethal damage. The value is false when only
	// attackers with deathtouch require it.
	required map[id.ID]bool
	// blockers resolves the IDs in required and pending.
	blockers map[id.ID]*game.Permanent
	// pending holds the trampling attackers still to assign their damage.
	pending []pendingTrampler
}

type assignedCombatDamage struct {
	amount     int
	deathtouch bool
}

type pendingTrampler struct {
	attacker   id.ID
	power      int
	deathtouch bool
	blockers   []id.ID
}

func newTrampleStep() *trampleStep {
	return &trampleStep{
		assigned: map[id.ID]assignedCombatDamage{},
		required: map[id.ID]bool{},
		blockers: map[id.ID]*game.Permanent{},
	}
}

// record adds the combat damage source assigns to permanents in events.
func (s *trampleStep) record(g *game.Game, source *game.Permanent, events []damageEvent) {
	deathtouch := hasKeyword(g, source, game.Deathtouch)
	for _, event := range events {
		if event.permanent != nil && event.amount > 0 {
			s.add(s.assigned, event.permanent.ObjectID, event.amount, deathtouch)
		}
	}
}

func (*trampleStep) add(assigned map[id.ID]assignedCombatDamage, permanentID id.ID, amount int, deathtouch bool) {
	damage := assigned[permanentID]
	damage.amount += amount
	damage.deathtouch = damage.deathtouch || deathtouch
	assigned[permanentID] = damage
}

// addPending registers a trampling attacker that will assign its combat damage
// to blockers in this step, after the attackers already registered.
func (s *trampleStep) addPending(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent) {
	power := combatDamageAssignedBy(g, attacker)
	if power <= 0 {
		return
	}
	pending := pendingTrampler{attacker: attacker.ObjectID, power: power, deathtouch: hasKeyword(g, attacker, game.Deathtouch)}
	for _, blocker := range blockers {
		if blocker != nil {
			pending.blockers = append(pending.blockers, blocker.ObjectID)
			s.blockers[blocker.ObjectID] = blocker
		}
	}
	if len(pending.blockers) > 0 {
		s.pending = append(s.pending, pending)
	}
}

// assign makes a trampling attacker's assignment of damage among its blockers
// and the attack target, returning the excess assigned to the attack target.
// The defending player chooses it when a blocker has banding (CR 702.22j);
// otherwise a legal recorded assignment is honored, and the default assigns as
// much excess as the joint assignment allows. With no blockers left, all of the
// damage is assigned to the attack target (CR 702.19d).
func (s *trampleStep) assign(g *game.Game, chooser combatDamageChooser, attacker *game.Permanent, blockers []*game.Permanent, target game.AttackTarget, damage int) ([]creatureDamageAssignment, int) {
	s.pending = slices.DeleteFunc(s.pending, func(pending pendingTrampler) bool {
		return pending.attacker == attacker.ObjectID
	})
	blockers = slices.DeleteFunc(slices.Clone(blockers), func(blocker *game.Permanent) bool {
		return blocker == nil
	})
	if damage <= 0 {
		return nil, 0
	}
	if len(blockers) == 0 {
		return nil, damage
	}
	withoutExcess := s.minimalAssignments(g, attacker, blockers, false, damage)
	var withExcess [][]int
	maxExcess := 0
	if _, ok := attackTargetCombatDamageEvent(g, attacker, target, damage); ok {
		withExcess = s.minimalAssignments(g, attacker, blockers, true, damage-1)
		if len(withExcess) > 0 {
			maxExcess = damage - vectorTotal(withExcess[0])
		}
	}
	candidates := func(excess int) [][]int {
		options := withoutExcess
		if excess > 0 {
			options = withExcess
		}
		options = slices.DeleteFunc(slices.Clone(options), func(amounts []int) bool {
			return vectorTotal(amounts) > damage-excess
		})
		if len(options) == 0 {
			// Unreachable while every earlier assignment kept the joint
			// assignment completable; assign nothing required.
			options = [][]int{make([]int, len(blockers))}
		}
		return options
	}
	var amounts []int
	var excess int
	if defendingPlayerAssignsAttackerDamage(g, blockers) {
		amounts, excess = s.defenderAssignment(g, chooser, attacker, blockers, target, damage, maxExcess, candidates)
	} else if recorded, recordedExcess, ok := s.recordedAssignment(g, attacker, blockers, damage, maxExcess); ok {
		amounts, excess = recorded, recordedExcess
	} else {
		excess = maxExcess
		amounts = s.defaultAssignment(g, attacker, blockers, damage-excess, candidates(excess)[0])
	}
	s.commit(g, attacker, blockers, amounts, excess)
	var assignments []creatureDamageAssignment
	for i, blocker := range blockers {
		if amounts[i] > 0 {
			assignments = append(assignments, creatureDamageAssignment{permanent: blocker, damage: amounts[i]})
		}
	}
	return assignments, excess
}

// defenderAssignment has the defending player choose how much of the damage
// goes to the attack target, then how the rest is divided among the blockers:
// first which of the least legal assignments it builds on, when there is more
// than one, then where the damage beyond it goes.
func (s *trampleStep) defenderAssignment(g *game.Game, chooser combatDamageChooser, attacker *game.Permanent, blockers []*game.Permanent, target game.AttackTarget, damage, maxExcess int, candidates func(int) [][]int) ([]int, int) {
	player := attackerDefendingPlayer(g, attacker.ObjectID)
	excess := 0
	if maxExcess > 0 {
		excess = chooser.trampleExcessToAttackTarget(g, player, attacker, target, damage, maxExcess)
	}
	amounts := slices.Clone(chooser.chooseTrampleMinimum(g, player, attacker, blockers, candidates(excess)))
	rest := damage - excess - vectorTotal(amounts)
	if rest <= 0 {
		return amounts, excess
	}
	var division []creatureDamageAssignment
	if excess == 0 && vectorTotal(amounts) == 0 {
		division = chooser.divideCombatDamage(g, player, attacker, blockers, rest, true)
	} else {
		prompt := fmt.Sprintf("Divide the other %d of %s's combat damage among the creatures blocking it, beyond the damage it must assign to them", rest, permanentChoiceLabel(g, attacker))
		defaultDivision := protectiveCombatDamageDivision(g, attacker, blockers, rest)
		if excess > 0 {
			defaultDivision = []creatureDamageAssignment{{permanent: blockers[0], damage: rest}}
		}
		division = chooser.chooseCombatDamageDivision(g, player, attacker, blockers, rest, true, prompt, defaultDivision)
	}
	for _, assignment := range division {
		amounts[slices.Index(blockers, assignment.permanent)] += assignment.damage
	}
	return amounts, excess
}

// chooseTrampleMinimum asks player which of the least legal assignments of
// attacker's damage to its blockers to build on. With one, nothing is asked.
func (c combatDamageChooser) chooseTrampleMinimum(g *game.Game, player game.PlayerID, attacker *game.Permanent, blockers []*game.Permanent, candidates [][]int) []int {
	if len(candidates) == 1 {
		return candidates[0]
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, amounts := range candidates {
		var parts []string
		for j, amount := range amounts {
			if amount > 0 {
				parts = append(parts, fmt.Sprintf("%d to %s", amount, permanentChoiceLabel(g, blockers[j])))
			}
		}
		label := "none"
		if len(parts) > 0 {
			label = strings.Join(parts, ", ")
		}
		options[i] = game.ChoiceOption{Index: i, Label: label}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceModal,
		Player:           player,
		Prompt:           fmt.Sprintf("Choose the least combat damage %s assigns to the creatures blocking it, so that every trampling attacker assigning excess damage past them has lethal damage assigned to them", permanentChoiceLabel(g, attacker)),
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
		Subject:          permanentChoiceInfo(g, attacker),
		CombatAttacker:   attacker.ObjectID,
	}
	selected := c.choose(g, request)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(candidates) {
		return candidates[0]
	}
	return candidates[selected[0]]
}

// recordedAssignment honors a recorded division of the attacker's damage among
// its blockers, the rest going to the attack target, when it is legal jointly
// with the other assignments of the step.
func (s *trampleStep) recordedAssignment(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent, damage, maxExcess int) ([]int, int, bool) {
	recorded, total, ok := recordedAttackerDamageAssignments(g, attacker, blockers, damage)
	if !ok {
		return nil, 0, false
	}
	excess := damage - total
	if excess > maxExcess {
		return nil, 0, false
	}
	amounts := make([]int, len(blockers))
	for _, assignment := range recorded {
		amounts[slices.Index(blockers, assignment.permanent)] = assignment.damage
	}
	required, pending := s.requirements(g, attacker, blockers, excess > 0)
	if !s.feasible(g, s.withAssignment(g, attacker, blockers, amounts), required, pending) {
		return nil, 0, false
	}
	return amounts, excess, true
}

// defaultAssignment builds on base, a least legal assignment, by assigning the
// rest of the damage to the blockers in turn up to lethal damage each, any
// remainder going to the last blocker.
func (s *trampleStep) defaultAssignment(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent, damage int, base []int) []int {
	amounts := slices.Clone(base)
	rest := damage - vectorTotal(amounts)
	assigned := s.withAssignment(g, attacker, blockers, amounts)
	deathtouch := hasKeyword(g, attacker, game.Deathtouch)
	for i, blocker := range blockers {
		if rest <= 0 {
			break
		}
		lethal := s.lethalShortfall(g, blocker, assigned[blocker.ObjectID], deathtouch)
		amount := min(rest, lethal)
		amounts[i] += amount
		rest -= amount
	}
	if rest > 0 {
		amounts[len(amounts)-1] += rest
	}
	return amounts
}

// commit records the attacker's assignment and, when it assigned excess
// damage, requires lethal damage on each of its blockers.
func (s *trampleStep) commit(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent, amounts []int, excess int) {
	deathtouch := hasKeyword(g, attacker, game.Deathtouch)
	for i, blocker := range blockers {
		if amounts[i] > 0 {
			s.add(s.assigned, blocker.ObjectID, amounts[i], deathtouch)
		}
		if excess > 0 {
			s.required[blocker.ObjectID] = s.required[blocker.ObjectID] || !deathtouch
			s.blockers[blocker.ObjectID] = blocker
		}
	}
}

// lethalShortfall is how much more of a source's damage must be assigned to
// blocker, beyond assigned, for it to be assigned lethal damage (CR 702.19b).
// Any nonzero damage from a deathtouch source is lethal (CR 702.2c).
func (*trampleStep) lethalShortfall(g *game.Game, blocker *game.Permanent, assigned assignedCombatDamage, deathtouch bool) int {
	if assigned.deathtouch {
		return 0
	}
	shortfall := max(0, lethalDamageRemaining(g, blocker)-assigned.amount)
	if deathtouch {
		if blocker.MarkedDeathtouchDamage {
			return 0
		}
		return min(shortfall, 1)
	}
	return shortfall
}

// assignedLethal reports whether blocker has been assigned lethal damage. strict
// is false when only attackers with deathtouch require it, for which deathtouch
// damage already marked on it is lethal, as lethalDamageRemainingFromSource.
func (*trampleStep) assignedLethal(g *game.Game, blocker *game.Permanent, assigned assignedCombatDamage, strict bool) bool {
	if assigned.deathtouch || !strict && blocker.MarkedDeathtouchDamage {
		return true
	}
	return assigned.amount >= lethalDamageRemaining(g, blocker)
}

// requirements returns the blockers that must be assigned lethal damage if the
// attacker assigns excess damage (or not) and the attackers still to assign.
func (s *trampleStep) requirements(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent, excess bool) (map[id.ID]bool, []pendingTrampler) {
	required := maps.Clone(s.required)
	if excess {
		deathtouch := hasKeyword(g, attacker, game.Deathtouch)
		for _, blocker := range blockers {
			required[blocker.ObjectID] = required[blocker.ObjectID] || !deathtouch
			s.blockers[blocker.ObjectID] = blocker
		}
	}
	return required, s.pending
}

// withAssignment returns the step's assigned damage plus the attacker's
// amounts to its blockers.
func (s *trampleStep) withAssignment(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent, amounts []int) map[id.ID]assignedCombatDamage {
	assigned := maps.Clone(s.assigned)
	deathtouch := hasKeyword(g, attacker, game.Deathtouch)
	for i, blocker := range blockers {
		if amounts[i] > 0 {
			s.add(assigned, blocker.ObjectID, amounts[i], deathtouch)
		}
	}
	return assigned
}

// minimalAssignments returns the least assignments of at most budget of the
// attacker's damage to its blockers (amounts in blocker order) with which the
// attackers still to assign can complete a legal joint assignment, given that
// the attacker assigns excess damage to the attack target (or not). Every legal
// assignment is one of these plus more damage. They are ordered by total, then
// with more damage on earlier blockers.
func (s *trampleStep) minimalAssignments(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent, excess bool, budget int) [][]int {
	required, pending := s.requirements(g, attacker, blockers, excess)
	deathtouch := hasKeyword(g, attacker, game.Deathtouch)
	// Damage beyond what makes a blocker lethal can't help any requirement, and
	// a required blocker that no attacker still to assign is blocked by needs
	// all of that from this attacker unless it is already lethal.
	caps := make([]int, len(blockers))
	forced := make([]bool, len(blockers))
	base := make([]int, len(blockers))
	maxTotal := 0
	for i, blocker := range blockers {
		strict, ok := required[blocker.ObjectID]
		if !ok {
			forced[i] = true
			continue
		}
		assigned := s.assigned[blocker.ObjectID]
		if !assigned.deathtouch {
			caps[i] = max(0, lethalDamageRemaining(g, blocker)-assigned.amount)
			if deathtouch {
				caps[i] = min(caps[i], 1)
			}
		}
		maxTotal += caps[i]
		forced[i] = !slices.ContainsFunc(pending, func(pending pendingTrampler) bool {
			return slices.Contains(pending.blockers, blocker.ObjectID)
		})
		if forced[i] && !s.assignedLethal(g, blocker, assigned, strict) {
			base[i] = caps[i]
		}
	}
	var minimal [][]int
	for total := vectorTotal(base); total <= min(budget, maxTotal); total++ {
		for _, amounts := range boundedVectors(base, caps, forced, total) {
			if slices.ContainsFunc(minimal, func(smaller []int) bool { return vectorCovers(amounts, smaller) }) {
				continue
			}
			if s.feasible(g, s.withAssignment(g, attacker, blockers, amounts), required, pending) {
				minimal = append(minimal, amounts)
			}
		}
	}
	return minimal
}

// boundedVectors lists the vectors with the given total that equal base where
// forced and lie between base and caps elsewhere, with more on earlier entries
// first.
func boundedVectors(base, caps []int, forced []bool, total int) [][]int {
	var vectors [][]int
	current := slices.Clone(base)
	var fill func(i, remaining int)
	fill = func(i, remaining int) {
		if i == len(current) {
			if remaining == 0 {
				vectors = append(vectors, slices.Clone(current))
			}
			return
		}
		if forced[i] {
			fill(i+1, remaining)
			return
		}
		for amount := min(caps[i], remaining); amount >= 0; amount-- {
			current[i] = amount
			fill(i+1, remaining-amount)
		}
		current[i] = base[i]
	}
	fill(0, total-vectorTotal(base))
	return vectors
}

// feasible reports whether the attackers still to assign (pending) can add
// damage to assigned so that every required blocker is assigned lethal damage.
// One damage from a pending attacker with deathtouch makes a blocker lethal
// (CR 702.2c); the others' damage is a transportation problem.
func (s *trampleStep) feasible(g *game.Game, assigned map[id.ID]assignedCombatDamage, required map[id.ID]bool, pending []pendingTrampler) bool {
	var uncovered []*game.Permanent
	for blockerID, strict := range required {
		if blocker := s.blockers[blockerID]; !s.assignedLethal(g, blocker, assigned[blockerID], strict) {
			uncovered = append(uncovered, blocker)
		}
	}
	if len(uncovered) == 0 {
		return true
	}
	slices.SortFunc(uncovered, func(a, b *game.Permanent) int { return cmp.Compare(a.ObjectID, b.ObjectID) })
	deathtouchBudget := make([]int, len(pending))
	for i, attacker := range pending {
		if attacker.deathtouch {
			deathtouchBudget[i] = attacker.power
		}
	}
	return s.coverable(g, uncovered, nil, assigned, pending, deathtouchBudget)
}

// coverable tries covering each uncovered blocker with one damage from a
// pending deathtouch attacker or leaving it to the others' damage.
func (s *trampleStep) coverable(g *game.Game, uncovered, rest []*game.Permanent, assigned map[id.ID]assignedCombatDamage, pending []pendingTrampler, deathtouchBudget []int) bool {
	if len(uncovered) == 0 {
		return s.damageCovers(g, rest, assigned, pending)
	}
	blocker := uncovered[0]
	for i, attacker := range pending {
		if deathtouchBudget[i] > 0 && slices.Contains(attacker.blockers, blocker.ObjectID) {
			deathtouchBudget[i]--
			ok := s.coverable(g, uncovered[1:], rest, assigned, pending, deathtouchBudget)
			deathtouchBudget[i]++
			if ok {
				return true
			}
		}
	}
	return s.coverable(g, uncovered[1:], append(slices.Clip(rest), blocker), assigned, pending, deathtouchBudget)
}

// damageCovers reports whether the pending attackers without deathtouch can
// divide their damage to assign each blocker its lethal shortfall, by a maximum
// flow from attackers (capacity: power) through the blockers they're blocked
// by to the blockers' shortfalls.
func (s *trampleStep) damageCovers(g *game.Game, blockers []*game.Permanent, assigned map[id.ID]assignedCombatDamage, pending []pendingTrampler) bool {
	if len(blockers) == 0 {
		return true
	}
	var suppliers []pendingTrampler
	for _, attacker := range pending {
		if !attacker.deathtouch {
			suppliers = append(suppliers, attacker)
		}
	}
	nodes := len(suppliers) + len(blockers) + 2
	source, sink := 0, nodes-1
	capacity := make([][]int, nodes)
	for i := range capacity {
		capacity[i] = make([]int, nodes)
	}
	demand := 0
	for j, blocker := range blockers {
		shortfall := s.lethalShortfall(g, blocker, assigned[blocker.ObjectID], false)
		capacity[1+len(suppliers)+j][sink] = shortfall
		demand += shortfall
	}
	for i, attacker := range suppliers {
		capacity[source][1+i] = attacker.power
		for j, blocker := range blockers {
			if slices.Contains(attacker.blockers, blocker.ObjectID) {
				capacity[1+i][1+len(suppliers)+j] = demand
			}
		}
	}
	flow := 0
	for flow < demand {
		parent := make([]int, nodes)
		for i := range parent {
			parent[i] = -1
		}
		parent[source] = source
		queue := []int{source}
		for len(queue) > 0 && parent[sink] < 0 {
			node := queue[0]
			queue = queue[1:]
			for next := range nodes {
				if parent[next] < 0 && capacity[node][next] > 0 {
					parent[next] = node
					queue = append(queue, next)
				}
			}
		}
		if parent[sink] < 0 {
			break
		}
		augment := demand
		for node := sink; node != source; node = parent[node] {
			augment = min(augment, capacity[parent[node]][node])
		}
		for node := sink; node != source; node = parent[node] {
			capacity[parent[node]][node] -= augment
			capacity[node][parent[node]] += augment
		}
		flow += augment
	}
	return flow >= demand
}

func vectorTotal(amounts []int) int {
	total := 0
	for _, amount := range amounts {
		total += amount
	}
	return total
}

// vectorCovers reports whether amounts is at least smaller in every entry.
func vectorCovers(amounts, smaller []int) bool {
	for i := range amounts {
		if amounts[i] < smaller[i] {
			return false
		}
	}
	return true
}
