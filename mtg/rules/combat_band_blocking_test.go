package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

// Source review: official CR effective 2026-09-25 (verified copy
// MagicCompRules 20260925.txt, SHA256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), read
// 702.22a-m, 509.1a-i, 509.3a-g, 510.1a-e and 702.19b-c. 702.22h: if an
// attacking creature becomes blocked by a creature, each other creature in the
// same band becomes blocked by that same blocking creature; blocking
// restrictions and evasion (509.1b) are checked only for the attacker the
// blocker is declared against, as in 702.22h's example (a flying banding
// creature and a swampwalker in one band, blocked by a flyer when the defender
// controls a Swamp). 509.3b/d: "blocks a creature" and "becomes blocked by a
// creature" trigger once per blocked/blocking pair, so each indirectly blocked
// member's pair has its own block events; 509.3c "becomes blocked" fires for
// each member that was unblocked. 702.22k: a blocker blocking a creature with
// banding has its damage divided by the active player among the creatures it
// blocks; 702.22e: the band persists after banding is lost, and with no member
// that has banding the division reverts to the blocking creature's controller
// (510.1d). 702.19b-c: a trampler's lethal damage counts damage the other
// attackers assign to a shared blocker. Card faces (pinned corpus): Mesa
// Pegasus ({1}{W} 1/1 Flying, banding), Anaconda ({3}{G} 3/3 Swampwalk), War
// Elephant ({3}{W} 2/2 Trample, banding), Kjeldoran Escort ({2}{W}{W} 2/3
// Banding), Benalish Hero ({W} 1/1 Banding), Merfolk Trickster ({U}{U} 2/2
// Flash; "When this creature enters, tap target creature an opponent controls.
// It loses all abilities until end of turn."), Infiltration Lens ({1}
// Equipment; "Whenever equipped creature becomes blocked by a creature, you may
// draw two cards." Equip {1}) standing in for Battering Ram's "Whenever this
// creature becomes blocked by a Wall" event shape. Effect-caused blocking
// (702.22i), "banded with" (#3483) and "bands with other" (#3484) are separate.

// bandBlockingAgent declares exactly the given blocks, optionally casting a card
// (targeting target) when it has priority during castStep, and answers damage
// divisions through bandingDivisionAgent.
type bandBlockingAgent struct {
	bandingDivisionAgent
	blocks   []game.BlockDeclaration
	cast     id.ID
	castStep game.Step
	target   id.ID
	casted   bool
}

func (a *bandBlockingAgent) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	for _, candidate := range legal {
		if _, ok := candidate.DeclareBlockersPayload(); ok {
			return action.DeclareBlockers(a.blocks)
		}
	}
	if a.cast != 0 && !a.casted && obs.Turn.Step == a.castStep {
		for _, candidate := range legal {
			if cast, ok := candidate.CastSpellPayload(); ok && cast.CardID == a.cast {
				a.casted = true
				return candidate
			}
		}
	}
	return action.Pass()
}

func (a *bandBlockingAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceTarget && a.target != 0 {
		for _, option := range request.Options {
			if slices.ContainsFunc(option.Targets, func(target game.Target) bool { return target.PermanentID == a.target }) {
				return []int{option.Index}
			}
		}
	}
	return a.bandingDivisionAgent.ChooseChoice(obs, request)
}

// bandCombatGame returns a game at Player1's declare attackers step.
func bandCombatGame() *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	return g
}

// bandAttack declares the given creatures attacking Player2 as one band.
func bandAttack(members ...*game.Permanent) action.DeclareAttackersAction {
	declaration := action.DeclareAttackersAction{Bands: []game.AttackingBand{bandOf(members...)}}
	for _, member := range members {
		member.SummoningSick = false
		declaration.Attackers = append(declaration.Attackers, attackAt(member, game.AttackTarget{Player: game.Player2}))
	}
	return declaration
}

func blockEvents(events []game.Event, kind game.EventKind) map[[2]id.ID]game.Event {
	found := make(map[[2]id.ID]game.Event)
	for _, event := range events {
		if event.Kind != kind {
			continue
		}
		switch kind {
		case game.EventAttackerBecameBlocked:
			found[[2]id.ID{event.RelatedPermanentID, event.PermanentID}] = event
		default:
			found[[2]id.ID{event.PermanentID, event.BlockedAttackerID}] = event
		}
	}
	return found
}

func markedDamage(t *testing.T, g *game.Game, permanent *game.Permanent) int {
	t.Helper()
	after, ok := permanentByObjectID(g, permanent.ObjectID)
	if !ok {
		t.Fatalf("%v left the battlefield", permanent.CardInstanceID)
	}
	return after.MarkedDamage
}

// TestBlockingBandMemberBlocksWholeBand is CR 702.22h's example in a live
// combat: Mesa Pegasus and Anaconda attack as a band while the defender
// controls a Swamp. A flyer can't be declared against swampwalking Anaconda,
// but blocking Pegasus blocks Anaconda too. Both become blocked by the flyer in
// one declaration batch, both block pairs are recorded, the active player
// divides the flyer's damage between them (702.22k) and nothing reaches the
// defending player.
func TestBlockingBandMemberBlocksWholeBand(t *testing.T) {
	g := bandCombatGame()
	pegasus := registryCombatPermanent(t, g, game.Player1, "Mesa Pegasus")
	anaconda := registryCombatPermanent(t, g, game.Player1, "Anaconda")
	addCombatPermanent(g, game.Player2, registryCardDef(t, "Swamp"))
	flyer := bandingTestCreature(g, game.Player2, 2, 4, game.Flying)
	attack := bandAttack(pegasus, anaconda)
	engine := NewEngine(nil)

	declared, ok := engine.Simulator().Apply(g, game.Player1, action.DeclareAttackersWithBands(attack.Attackers, attack.Bands), simPassPolicies())
	if !ok {
		t.Fatal("band declaration rejected")
	}
	declared.Turn.Step = game.StepDeclareBlockers
	ce := combatEngine{engine}
	if ce.blockersLegal(declared, game.Player2, action.DeclareBlockersAction{Blockers: []game.BlockDeclaration{{Blocker: flyer.ObjectID, Blocking: anaconda.ObjectID}}}) {
		t.Fatal("flyer was allowed to be declared against swampwalking Anaconda")
	}
	if !ce.applyBlockers(declared, game.Player2, action.DeclareBlockersAction{Blockers: []game.BlockDeclaration{{Blocker: flyer.ObjectID, Blocking: pegasus.ObjectID}}}) {
		t.Fatal("flyer blocking Mesa Pegasus was rejected")
	}
	want := []game.BlockDeclaration{{Blocker: flyer.ObjectID, Blocking: pegasus.ObjectID}, {Blocker: flyer.ObjectID, Blocking: anaconda.ObjectID}}
	if !slices.Equal(declared.Combat.Blockers, want) {
		t.Fatalf("blockers = %+v, want %+v", declared.Combat.Blockers, want)
	}
	for _, attacker := range []*game.Permanent{pegasus, anaconda} {
		if !declared.Combat.BlockedAttackers[attacker.ObjectID] || !slices.Equal(declared.Combat.BlockerOrder[attacker.ObjectID], []id.ID{flyer.ObjectID}) {
			t.Fatalf("attacker %v blocked=%v order=%v, want blocked by the flyer", attacker.ObjectID, declared.Combat.BlockedAttackers[attacker.ObjectID], declared.Combat.BlockerOrder[attacker.ObjectID])
		}
	}

	attackerAgent := &bandingDivisionAgent{}
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{Blocker: flyer.ObjectID, Blocking: pegasus.ObjectID}}}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = attackerAgent, defender
	life := g.Players[game.Player2].Life
	after, ok := engine.Simulator().ResolveCombatWithAttackers(g, game.Player1, attack, policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	becameBlocked := blockEvents(after.Events, game.EventAttackerBecameBlocked)
	declaredEvents := blockEvents(after.Events, game.EventBlockerDeclared)
	var batch id.ID
	for _, attacker := range []*game.Permanent{pegasus, anaconda} {
		pair := [2]id.ID{flyer.ObjectID, attacker.ObjectID}
		blocked, blockedOK := becameBlocked[pair]
		blocking, blockingOK := declaredEvents[pair]
		if !blockedOK || !blockingOK {
			t.Fatalf("attacker %v: became-blocked=%v blocker-declared=%v, want both events", attacker.ObjectID, blockedOK, blockingOK)
		}
		if blocked.Player != game.Player2 || blocked.AttackTarget.Player != game.Player2 || blocked.SimultaneousID == 0 ||
			blocking.SimultaneousID != blocked.SimultaneousID || batch != 0 && blocked.SimultaneousID != batch {
			t.Fatalf("attacker %v events = %+v / %+v, want one declaration batch against Player2", attacker.ObjectID, blocked, blocking)
		}
		batch = blocked.SimultaneousID
	}
	if len(defender.requests) != 0 || len(attackerAgent.requests) != 1 || attackerAgent.requests[0].CombatBlocker != flyer.ObjectID ||
		len(attackerAgent.requests[0].Options) != 2 {
		t.Fatalf("division requests: active %+v defender %+v, want the active player dividing the flyer's damage between both members", attackerAgent.requests, defender.requests)
	}
	if markedDamage(t, after, pegasus) != 0 || markedDamage(t, after, anaconda) != 2 {
		t.Fatal("default division should put the flyer's 2 damage below lethal on Anaconda")
	}
	if permanentWithCardOnBattlefield(after, flyer.CardInstanceID) {
		t.Fatal("the flyer should die to the band's 4 damage")
	}
	if after.Players[game.Player2].Life != life {
		t.Fatalf("defender life = %d, want %d: the whole band was blocked", after.Players[game.Player2].Life, life)
	}
}

// TestBandTramplerCountsMemberDamageToSharedBlocker: War Elephant and a 2/2 attack
// as a band and a 2/3 blocks the 2/2, so it blocks War Elephant too. The 2/2
// assigns 2 to the blocker, so War Elephant needs only 1 more for lethal and
// tramples 1 over (702.19b-c). The active player divides the blocker's 2
// damage (702.22k): by default 1 to each so neither dies; a scripted answer
// puts both on the 2/2.
func TestBandTramplerCountsMemberDamageToSharedBlocker(t *testing.T) {
	for _, test := range []struct {
		name                     string
		amounts                  func(elephant, bear *game.Permanent) map[id.ID]int
		elephantDamage, bearDies int
	}{
		{name: "default", elephantDamage: 1},
		{name: "scripted", amounts: func(_, bear *game.Permanent) map[id.ID]int {
			return map[id.ID]int{bear.CardInstanceID: 2}
		}, bearDies: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := bandCombatGame()
			elephant := registryCombatPermanent(t, g, game.Player1, "War Elephant")
			bear := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			blocker := bandingTestCreature(g, game.Player2, 2, 3)
			attackerAgent := &bandingDivisionAgent{}
			if test.amounts != nil {
				amounts := test.amounts(elephant, bear)
				attackerAgent.answer = func(request game.ChoiceRequest) []int {
					if request.Kind != game.ChoiceCombatDamageDivision {
						return request.DefaultSelection
					}
					return divisionSelection(request, amounts)
				}
			}
			defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: bear.ObjectID}}}
			policies := simPassPolicies()
			policies[game.Player1], policies[game.Player2] = attackerAgent, defender
			life := g.Players[game.Player2].Life

			after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, bandAttack(elephant, bear), policies)
			if !ok {
				t.Fatal("combat not accepted")
			}
			if after.Players[game.Player2].Life != life-1 {
				t.Fatalf("defender life = %d, want %d: War Elephant tramples 1", after.Players[game.Player2].Life, life-1)
			}
			if permanentWithCardOnBattlefield(after, blocker.CardInstanceID) {
				t.Fatal("blocker should die to 2 from the 2/2 and 1 from War Elephant")
			}
			var division []game.ChoiceRequest
			for _, request := range attackerAgent.requests {
				if request.Kind == game.ChoiceCombatDamageDivision {
					division = append(division, request)
				}
			}
			if len(division) != 1 || division[0].CombatBlocker != blocker.ObjectID || len(division[0].Options) != 2 {
				t.Fatalf("division requests = %+v, want one for the blocker's damage between both members", division)
			}
			if markedDamage(t, after, elephant) != test.elephantDamage {
				t.Fatalf("War Elephant damage = %d, want %d", markedDamage(t, after, elephant), test.elephantDamage)
			}
			bearDied := 0
			if !permanentWithCardOnBattlefield(after, bear.CardInstanceID) {
				bearDied = 1
			}
			if bearDied != test.bearDies {
				t.Fatalf("2/2 died = %d, want %d", bearDied, test.bearDies)
			}
		})
	}
}

// TestDeathtouchBlockerOfBandKillsAtMostOneByDefault: a 2/2 deathtouch creature
// blocks a band of Benalish Hero and a 2/2. The active player divides its
// damage (702.22k); any damage is lethal, so the default puts all of it on one
// member, the lowest-power Hero, and the 2/2 survives.
func TestDeathtouchBlockerOfBandKillsAtMostOneByDefault(t *testing.T) {
	g := bandCombatGame()
	hero := registryCombatPermanent(t, g, game.Player1, "Benalish Hero")
	bear := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := bandingTestCreature(g, game.Player2, 2, 2, game.Deathtouch)
	attackerAgent := &bandingDivisionAgent{}
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: bear.ObjectID}}}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = attackerAgent, defender

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, bandAttack(hero, bear), policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	if len(attackerAgent.requests) != 1 || attackerAgent.requests[0].Player != game.Player1 || attackerAgent.requests[0].CombatBlocker != blocker.ObjectID {
		t.Fatalf("division requests = %+v, want the active player dividing the deathtouch blocker's damage", attackerAgent.requests)
	}
	if permanentWithCardOnBattlefield(after, hero.CardInstanceID) || !permanentWithCardOnBattlefield(after, bear.CardInstanceID) {
		t.Fatal("default division should kill only Benalish Hero")
	}
	if permanentWithCardOnBattlefield(after, blocker.CardInstanceID) {
		t.Fatal("deathtouch blocker should die to the band's 3 damage")
	}
}

// TestBandBlockedAfterBandingLostUsesDefenderDivision: Benalish Hero and a 2/2
// attack as a band and a 3/3 blocks the 2/2. The defender flashes in Merfolk
// Trickster so Hero, the only member with banding, loses all abilities, either
// before blockers are declared or after the block has propagated. Either way
// the 3/3 blocks Hero too (the band persists, 702.22e), but no blocked attacker
// has banding when damage is assigned, so the active player isn't asked: the
// blocking creature's controller divides its damage (510.1d). Its answer is
// honored, all 3 on Hero sparing the 2/2, and its default lethal-in-order
// division kills both.
func TestBandBlockedAfterBandingLostUsesDefenderDivision(t *testing.T) {
	for _, step := range []game.Step{game.StepDeclareAttackers, game.StepDeclareBlockers} {
		for _, allOnHero := range []bool{false, true} {
			name := step.String() + "/default"
			if allOnHero {
				name = step.String() + "/all on Hero"
			}
			t.Run(name, func(t *testing.T) {
				g := bandCombatGame()
				hero := registryCombatPermanent(t, g, game.Player1, "Benalish Hero")
				bear := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
				blocker := bandingTestCreature(g, game.Player2, 3, 3)
				island := registryCardDef(t, "Island")
				addCombatPermanent(g, game.Player2, island)
				addCombatPermanent(g, game.Player2, island)
				trickster := addCardToHand(g, game.Player2, registryCardDef(t, "Merfolk Trickster"))
				attackerAgent := &bandingDivisionAgent{}
				defender := &bandBlockingAgent{
					blocks: []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: bear.ObjectID}},
					cast:   trickster, castStep: step, target: hero.ObjectID,
				}
				if allOnHero {
					defender.answer = func(request game.ChoiceRequest) []int {
						return divisionSelection(request, map[id.ID]int{hero.CardInstanceID: 3})
					}
				}
				policies := simPassPolicies()
				policies[game.Player1], policies[game.Player2] = attackerAgent, defender

				after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, bandAttack(hero, bear), policies)
				if !ok {
					t.Fatal("combat not accepted")
				}
				if !defender.casted {
					t.Fatal("defender never cast Merfolk Trickster")
				}
				if _, ok := blockEvents(after.Events, game.EventAttackerBecameBlocked)[[2]id.ID{blocker.ObjectID, hero.ObjectID}]; !ok {
					t.Fatal("Benalish Hero did not become blocked with its band")
				}
				if len(attackerAgent.requests) != 0 {
					t.Fatalf("active player was asked %+v; with banding lost the blocker's controller divides", attackerAgent.requests)
				}
				if len(defender.requests) != 1 {
					t.Fatalf("defender division requests = %d, want 1", len(defender.requests))
				}
				request := defender.requests[0]
				if request.Kind != game.ChoiceCombatDamageDivision || request.Player != game.Player2 ||
					request.CombatBlocker != blocker.ObjectID || request.MinChoices != 3 || request.MaxChoices != 3 || len(request.Options) != 2 {
					t.Fatalf("division request = %+v, want Player2 dividing the 3/3's 3 damage between Hero and the 2/2", request)
				}
				if permanentWithCardOnBattlefield(after, hero.CardInstanceID) {
					t.Fatal("Benalish Hero should die to the 3/3's damage")
				}
				if bearSurvives := permanentWithCardOnBattlefield(after, bear.CardInstanceID); bearSurvives != allOnHero {
					t.Fatalf("2/2 on battlefield = %v, want %v", bearSurvives, allOnHero)
				}
				if permanentWithCardOnBattlefield(after, blocker.CardInstanceID) {
					t.Fatal("the 3/3 should die to the band's 3 damage")
				}
			})
		}
	}
}

// TestBlockerControllerDividesDamageAmongUnbandedAttackers is CR 510.1d
// without banding: a 3/3 that can block an additional creature blocks two
// unbanded 2/2s. Its controller, not the active player, is asked to divide its
// damage, and an answer the lethal-in-order default wouldn't give (all 3 on
// the second attacker) is honored.
func TestBlockerControllerDividesDamageAmongUnbandedAttackers(t *testing.T) {
	g := bandCombatGame()
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	first.SummoningSick, second.SummoningSick = false, false
	blocker := bandingTestCreature(g, game.Player2, 3, 5)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{Kind: game.RuleEffectCanBlockAdditional, SourceObjectID: blocker.ObjectID, AffectedSource: true, AdditionalBlockCount: 1})
	attackerAgent := &bandingDivisionAgent{}
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{
		{Blocker: blocker.ObjectID, Blocking: first.ObjectID},
		{Blocker: blocker.ObjectID, Blocking: second.ObjectID},
	}}
	defender.answer = func(request game.ChoiceRequest) []int {
		return divisionSelection(request, map[id.ID]int{second.CardInstanceID: 3})
	}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = attackerAgent, defender
	declaration := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		attackAt(first, game.AttackTarget{Player: game.Player2}),
		attackAt(second, game.AttackTarget{Player: game.Player2}),
	}}

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declaration, policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	if len(attackerAgent.requests) != 0 {
		t.Fatalf("active player was asked %+v; the blocker's controller divides", attackerAgent.requests)
	}
	if len(defender.requests) != 1 {
		t.Fatalf("defender division requests = %d, want 1", len(defender.requests))
	}
	request := defender.requests[0]
	if request.Kind != game.ChoiceCombatDamageDivision || request.Player != game.Player2 || request.CombatBlocker != blocker.ObjectID {
		t.Fatalf("division request = %+v, want Player2 dividing the blocker's damage", request)
	}
	if !permanentWithCardOnBattlefield(after, first.CardInstanceID) {
		t.Fatal("the first attacker was assigned no damage and should survive")
	}
	if permanentWithCardOnBattlefield(after, second.CardInstanceID) {
		t.Fatal("the second attacker should die to the 3 damage assigned to it")
	}
	if got := markedDamage(t, after, blocker); got != 4 {
		t.Fatalf("blocker damage = %d, want 4", got)
	}
}

// TestRedundantBandBlockRejected: a creature that can block an additional
// creature can't be declared against two members of one band, since blocking
// one blocks both (702.22h). It can still block two attackers that aren't
// banded, and the staged block choices never offer the redundant second block.
func TestRedundantBandBlockRejected(t *testing.T) {
	g := bandCombatGame()
	hero := registryCombatPermanent(t, g, game.Player1, "Benalish Hero")
	bear := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	loner := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := bandingTestCreature(g, game.Player2, 3, 3)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{Kind: game.RuleEffectCanBlockAdditional, SourceObjectID: blocker.ObjectID, AffectedSource: true, AdditionalBlockCount: 1})
	attack := bandAttack(hero, bear)
	loner.SummoningSick = false
	attack.Attackers = append(attack.Attackers, attackAt(loner, game.AttackTarget{Player: game.Player2}))
	engine := NewEngine(nil)
	declared, ok := engine.Simulator().Apply(g, game.Player1, action.DeclareAttackersWithBands(attack.Attackers, attack.Bands), simPassPolicies())
	if !ok {
		t.Fatal("band declaration rejected")
	}
	declared.Turn.Step = game.StepDeclareBlockers
	ce := combatEngine{engine}
	block := func(attackers ...*game.Permanent) action.DeclareBlockersAction {
		var declaration action.DeclareBlockersAction
		for _, attacker := range attackers {
			declaration.Blockers = append(declaration.Blockers, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID})
		}
		return declaration
	}
	if ce.blockersLegal(declared, game.Player2, block(hero, bear)) {
		t.Fatal("one blocker declared against two members of a band was accepted")
	}
	if !ce.blockersLegal(declared, game.Player2, block(hero, loner)) || !ce.blockersLegal(declared, game.Player2, block(bear)) {
		t.Fatal("non-redundant blocks were rejected")
	}
	for _, candidate := range ce.legalBlockers(declared, game.Player2) {
		payload, _ := candidate.DeclareBlockersPayload()
		if slices.Contains(payload.Blockers, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: hero.ObjectID}) &&
			slices.Contains(payload.Blockers, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: bear.ObjectID}) {
			t.Fatalf("legal blocker sample %+v is redundant", payload.Blockers)
		}
	}

	staged := &stagedBlockAll{}
	agents := simPassPolicies()
	agents[game.Player2] = staged
	chosen := ce.chooseStagedBlockers(declared, game.Player2, agents, &TurnLog{})
	payload, _ := chosen.DeclareBlockersPayload()
	if !ce.applyBlockers(declared, game.Player2, payload) {
		t.Fatalf("staged declaration %+v rejected", payload.Blockers)
	}
	if staged.redundantOffered(blocker.ObjectID, hero.ObjectID, bear.ObjectID) {
		t.Fatalf("staged choices offered a redundant band block: %+v", staged.requests)
	}
	for _, attacker := range []*game.Permanent{hero, bear, loner} {
		if !declared.Combat.BlockedAttackers[attacker.ObjectID] {
			t.Fatalf("attacker %v not blocked after staged declaration %+v", attacker.ObjectID, payload.Blockers)
		}
	}
	if len(payload.Blockers) != 2 {
		t.Fatalf("staged declaration = %+v, want one band member and the loner", payload.Blockers)
	}
}

// stagedBlockAll answers every staged block choice by blocking when it may.
type stagedBlockAll struct {
	simPassPolicy
	requests []game.ChoiceRequest
}

func (a *stagedBlockAll) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceCombatDeclaration {
		return request.DefaultSelection
	}
	a.requests = append(a.requests, request)
	for _, option := range request.Options {
		if option.Block.Exists {
			return []int{option.Index}
		}
	}
	return request.DefaultSelection
}

// redundantOffered reports whether a block option against second was offered
// after blocker had already been chosen to block first, or vice versa.
func (a *stagedBlockAll) redundantOffered(blocker, first, second id.ID) bool {
	chose := make(map[id.ID]bool)
	for _, request := range a.requests {
		if request.CombatBlocker != blocker {
			continue
		}
		for _, option := range request.Options {
			if !option.Block.Exists {
				continue
			}
			attacker := option.Block.Val.Blocking
			if attacker == first && chose[second] || attacker == second && chose[first] {
				return true
			}
			chose[attacker] = true
		}
	}
	return false
}

// TestBandBecomesBlockedByTriggerAndDamage: Kjeldoran Escort and War Elephant,
// equipped with Infiltration Lens, attack as a band and a 4/4 blocks the Escort.
// War Elephant becomes blocked by that creature too, so the Lens's "becomes
// blocked by a creature" trigger fires (509.3d) and its controller draws two.
// The Escort's 2 and War Elephant's 2 kill the 4/4 with nothing to trample over,
// and the active player divides the 4/4's damage (702.22k), here 3 to the
// Escort and 1 to War Elephant.
func TestBandBecomesBlockedByTriggerAndDamage(t *testing.T) {
	g := bandCombatGame()
	escort := registryCombatPermanent(t, g, game.Player1, "Kjeldoran Escort")
	elephant := registryCombatPermanent(t, g, game.Player1, "War Elephant")
	lens := addCombatPermanent(g, game.Player1, registryCardDef(t, "Infiltration Lens"))
	lens.AttachedTo = opt.Val(elephant.ObjectID)
	addCardToLibraryNamed(g, game.Player1, "Library Card A")
	addCardToLibraryNamed(g, game.Player1, "Library Card B")
	blocker := bandingTestCreature(g, game.Player2, 4, 4)
	attackerAgent := &bandingDivisionAgent{answer: func(request game.ChoiceRequest) []int {
		if request.Kind != game.ChoiceCombatDamageDivision {
			return request.DefaultSelection
		}
		return divisionSelection(request, map[id.ID]int{escort.CardInstanceID: 3, elephant.CardInstanceID: 1})
	}}
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: escort.ObjectID}}}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = attackerAgent, defender
	life := g.Players[game.Player2].Life
	hand := g.Players[game.Player1].Hand.Size()

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, bandAttack(escort, elephant), policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	if got := after.Players[game.Player1].Hand.Size(); got != hand+2 {
		t.Fatalf("Player1 hand = %d, want %d: Infiltration Lens should trigger for War Elephant", got, hand+2)
	}
	if after.Players[game.Player2].Life != life {
		t.Fatalf("defender life = %d, want %d", after.Players[game.Player2].Life, life)
	}
	if permanentWithCardOnBattlefield(after, blocker.CardInstanceID) {
		t.Fatal("4/4 should die to the band's 4 damage")
	}
	if permanentWithCardOnBattlefield(after, escort.CardInstanceID) {
		t.Fatal("Kjeldoran Escort should die to the 3 damage assigned to it")
	}
	if markedDamage(t, after, elephant) != 1 {
		t.Fatalf("War Elephant damage = %d, want 1", markedDamage(t, after, elephant))
	}
}
