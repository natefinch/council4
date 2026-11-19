package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: Oracle ID f77a9160-1d37-476d-8908-29ddcc982328, pinned
// printing 50162cdd-ba30-48df-93ff-197c7f4a2913 (Streets of New Capenna 114),
// and pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// The current exact Scryfall card response (retrieved 2026-11-19 from
// https://api.scryfall.com/cards/50162cdd-ba30-48df-93ff-197c7f4a2913 and
// saved at .cardwork/sources/mayhem-patrol-card.json) confirms {1}{R},
// Creature — Devil Warrior 1/2, and the full text: menace; whenever it attacks,
// target creature gets +1/+0 until end of turn; Blitz {1}{R} with haste,
// death-draw, and a next-end-step sacrifice if the Blitz cost was paid. Its
// exact-printing rulings response from
// https://api.scryfall.com/cards/50162cdd-ba30-48df-93ff-197c7f4a2913/rulings,
// saved at .cardwork/sources/mayhem-patrol-rulings.json and retrieved
// 2026-11-19, contains five WotC rulings dated 2022-04-29: Blitz is still a
// cast with normal timing/countering; the delayed sacrifice applies only while
// the same permanent remains on the battlefield; attacking is optional; copies
// do not inherit Blitz; and the death-draw triggers on any death. Official CR
// effective 2026-09-25:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
// (verified local copy .cardwork/sources/MagicCompRules-20260925.txt,
// SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca).
// Read 118.9, 601.2b/c, 302.6, 508.1a/f, 603.2-3d, 115.1d, 611.2a, 613.4c,
// 702.10b, 702.111b, 702.152a, 603.7a-c, 513.1, 700.4, and 121.1: Blitz is an
// alternative cost, haste permits an immediate attack, the attack trigger's
// target is chosen as it goes on the stack, the +1/+0 applies in layer 7c for
// its stated duration, menace restricts blocks, and Blitz's conditional
// consequences are distinct from the attack trigger.
const mayhemPatrolScenarioOracleID = "f77a9160-1d37-476d-8908-29ddcc982328"

type mayhemPatrolChoiceAgent struct {
	evokeCostAgent
	target  game.Target
	offered []game.Target
}

func (a *mayhemPatrolChoiceAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceTarget {
		selected := -1
		for _, option := range request.Options {
			for _, target := range option.Targets {
				a.offered = append(a.offered, target)
				if target == a.target {
					selected = option.Index
				}
			}
		}
		if selected >= 0 {
			return []int{selected}
		}
	}
	return a.evokeCostAgent.ChooseChoice(obs, request)
}

func TestMayhemPatrolBlitzAttackTargetsAnyCreatureAndGivesPlusOnePower(t *testing.T) {
	if authored.MayhemPatrol.OracleID != mayhemPatrolScenarioOracleID {
		t.Fatalf("authored Mayhem Patrol Oracle ID = %q", authored.MayhemPatrol.OracleID)
	}
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(authored.MayhemPatrol.OracleID)
	if err != nil || def == nil || def.Name != "Mayhem Patrol" {
		t.Fatalf("checked Mayhem Patrol identity lookup = %v, %v", def, err)
	}

	engine := NewEngineWithPrograms(nil, registry.Programs())
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	selfTarget := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	defendingTarget := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	otherTarget := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	targetPowers := map[game.ObjectID]int{
		selfTarget.ObjectID:      effectivePower(g, selfTarget),
		defendingTarget.ObjectID: effectivePower(g, defendingTarget),
		otherTarget.ObjectID:     effectivePower(g, otherTarget),
	}
	card := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	agent := &mayhemPatrolChoiceAgent{
		evokeCostAgent: evokeCostAgent{costLabel: "Blitz"},
		target:         game.PermanentTarget(selfTarget.ObjectID),
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(card, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("could not cast Mayhem Patrol for its Blitz cost")
	}
	spell, ok := g.Stack.Peek()
	if !ok || !spell.Blitzed || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Mayhem Patrol did not pay {1}{R} for Blitz: spell=%+v mana=%d",
			spell, g.Players[game.Player1].ManaPool.Total())
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	patrol := permanentForCard(g, card)
	requireBlitzEntry(t, g, patrol, true, true)
	if !hasKeyword(g, patrol, game.Menace) {
		t.Fatal("Mayhem Patrol did not have menace after entering")
	}
	if !slices.Contains(eligibleAttackers(g, game.Player1), patrol) {
		t.Fatal("Blitz haste did not let Mayhem Patrol attack on the turn it entered")
	}

	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attack := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		attackAt(patrol, game.AttackTarget{Player: game.Player2}),
	}}
	if !engine.applyDeclareAttackers(g, game.Player1, attack) {
		t.Fatal("Blitzed Mayhem Patrol could not attack Player 2")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Mayhem Patrol's attack trigger was not put on the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || len(trigger.Targets) != 1 || trigger.Targets[0] != agent.target {
		t.Fatalf("attack-trigger target = %+v, want the selected creature %v", trigger, agent.target)
	}

	wantTargets := map[game.Target]bool{
		game.PermanentTarget(selfTarget.ObjectID):      true,
		game.PermanentTarget(defendingTarget.ObjectID): true,
		game.PermanentTarget(otherTarget.ObjectID):     true,
		game.PermanentTarget(patrol.ObjectID):          true,
	}
	if len(agent.offered) != len(wantTargets) {
		t.Fatalf("attack trigger offered targets %v, want all four creatures", agent.offered)
	}
	seenTargets := make(map[game.Target]bool, len(agent.offered))
	for _, offered := range agent.offered {
		if !wantTargets[offered] {
			t.Fatalf("attack trigger offered non-creature target %v", offered)
		}
		seenTargets[offered] = true
	}
	for expected := range wantTargets {
		if !seenTargets[expected] {
			t.Fatalf("attack trigger omitted legal creature target %v from %v", expected, agent.offered)
		}
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	for _, creature := range []*game.Permanent{selfTarget, defendingTarget, otherTarget} {
		want := targetPowers[creature.ObjectID]
		if creature == selfTarget {
			want++
		}
		if got := effectivePower(g, creature); got != want {
			t.Fatalf("creature %d power = %d, want %d", creature.ObjectID, got, want)
		}
	}
}
