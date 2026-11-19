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

// Source review: Oracle ID dedf0d3a-090a-4cc3-b5ee-65b6cb990b72, Plasma Jockey,
// pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// printing 741ed811-cbcd-4178-a874-dc79d2a5d6f1. Exact text: {3}{R}, Creature —
// Lizard Warrior 3/1; “Whenever this creature attacks, target creature an
// opponent controls can't block this turn. Blitz {2}{R} (If you cast this spell
// for its blitz cost, it gains haste and ‘When this creature dies, draw a card.’
// Sacrifice it at the beginning of the next end step.)” Confirmed against
// Scryfall's current card response on 2026-11-19:
// https://api.scryfall.com/cards/741ed811-cbcd-4178-a874-dc79d2a5d6f1.
// The same day's /rulings response at
// https://api.scryfall.com/cards/741ed811-cbcd-4178-a874-dc79d2a5d6f1/rulings
// contains five WotC rulings, all dated 2022-04-29: Blitz is still a cast and
// can be countered; the delayed sacrifice applies only if that permanent is
// still on the battlefield; attacking is optional; permanent copies don't
// inherit Blitz; and the Blitz death draw triggers for any death.
// Official CR effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829;
// retained at .cardwork/sources/MagicCompRules-20261106.txt). Read 118.9 and
// 601.2b/f-h/702.152a (Blitz alternative payment and its conditional abilities),
// 302.6/702.10b (haste permits the new creature to attack), 508.1d and 603.2-3
// (attack trigger timing and independent resolution), 509.1b (blocking
// restrictions), 611.2a/c and 514.2 (the prohibition expires at end of turn),
// 608.2b (target legality), 603.7a-c/513.1 (next-end-step delayed sacrifice),
// and 700.4/121.1 (dies and drawing a card).
type plasmaJockeyChoiceAgent struct {
	evokeCostAgent
	target  game.Target
	offered []game.Target
}

func (a *plasmaJockeyChoiceAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
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

func TestPlasmaJockeyBlitzAttackPreventsOpponentCreatureBlocking(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(authored.PlasmaJockey.OracleID)
	if err != nil || def.Name != "Plasma Jockey" {
		t.Fatalf("checked Plasma Jockey identity lookup = %v, %v", def, err)
	}

	engine := NewEngineWithPrograms(nil, registry.Programs())
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	otherDefender := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	otherOpponent := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	card := addCardToHand(g, game.Player1, def)
	draw := addCardToLibraryNamed(g, game.Player1, "Plasma Jockey Blitz draw")
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	agent := &plasmaJockeyChoiceAgent{
		evokeCostAgent: evokeCostAgent{costLabel: "Blitz"},
		target:         game.PermanentTarget(target.ObjectID),
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(card, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("could not cast Plasma Jockey for its Blitz cost")
	}
	spell, ok := g.Stack.Peek()
	if !ok || !spell.Blitzed || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Plasma Jockey did not pay {2}{R} for Blitz: spell=%+v mana=%d", spell, g.Players[game.Player1].ManaPool.Total())
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	jockey := permanentForCard(g, card)
	requireBlitzEntry(t, g, jockey, true, true)
	if !slices.Contains(eligibleAttackers(g, game.Player1), jockey) {
		t.Fatal("Blitz haste did not let Plasma Jockey attack on the turn it entered")
	}

	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attack := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		attackAt(jockey, game.AttackTarget{Player: game.Player2}),
	}}
	if !engine.applyDeclareAttackers(g, game.Player1, attack) {
		t.Fatal("Blitzed Plasma Jockey could not attack Player 2")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Plasma Jockey's attack trigger was not put on the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || len(trigger.Targets) != 1 || trigger.Targets[0].PermanentID != target.ObjectID {
		t.Fatalf("Plasma Jockey attack trigger target = %+v, want the chosen opponent creature %v", trigger, target.ObjectID)
	}
	wantTargets := map[game.Target]bool{
		game.PermanentTarget(target.ObjectID):        true,
		game.PermanentTarget(otherDefender.ObjectID): true,
		game.PermanentTarget(otherOpponent.ObjectID): true,
	}
	if len(agent.offered) != len(wantTargets) {
		t.Fatalf("attack-trigger target options = %v, want all three opponent creatures", agent.offered)
	}
	seenTargets := make(map[game.Target]bool, len(agent.offered))
	for _, offered := range agent.offered {
		if !wantTargets[offered] {
			t.Fatalf("attack trigger offered non-opponent target %v", offered)
		}
		seenTargets[offered] = true
	}
	for expected := range wantTargets {
		if !seenTargets[expected] {
			t.Fatalf("attack trigger omitted legal opponent target %v from %v", expected, agent.offered)
		}
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	g.Turn.Step = game.StepDeclareBlockers
	canOtherDefenderBlock := false
	for _, candidate := range legalDeclareBlockersActions(g, game.Player2) {
		blocks, ok := candidate.DeclareBlockersPayload()
		if !ok {
			continue
		}
		for _, block := range blocks.Blockers {
			if block.Blocker == target.ObjectID {
				t.Fatal("the targeted opponent creature could block this turn")
			}
			if block.Blocker == otherDefender.ObjectID && block.Blocking == jockey.ObjectID {
				canOtherDefenderBlock = true
			}
		}
	}
	if !canOtherDefenderBlock {
		t.Fatal("Plasma Jockey's attack trigger prevented an untargeted creature from blocking")
	}

	blitzEndStep(engine, g)
	resolveStackWithTriggers(engine, g, agents)
	if !g.Players[game.Player1].Graveyard.Contains(card) || !g.Players[game.Player1].Hand.Contains(draw) {
		t.Fatal("Blitz did not sacrifice Plasma Jockey at the next end step and draw a card when it died")
	}
}
