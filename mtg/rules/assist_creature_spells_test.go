package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

const (
	chargingBinoxAssistOracleID = "0f10f9d1-f9f9-4356-9208-0f21f44bfb97"
	magmaHellionAssistOracleID  = "88f6575d-a6cb-4b2c-90dd-c6239f3c6e4d"
)

// Source review: pinned-corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Charging Binox Oracle ID 0f10f9d1-f9f9-4356-9208-0f21f44bfb97, printing
// revision 68222ab7-7b9c-43e5-b80e-db643d80a6d9 (Battlebond #66), {7}{G},
// Creature — Beast 7/5; exact Oracle text:
// "Assist (Another player can pay up to {7} of this spell's cost.)
// Trample"
// Its Scryfall rulings URI
// https://api.scryfall.com/cards/68222ab7-7b9c-43e5-b80e-db643d80a6d9/rulings
// was retrieved 2026-11-22 and returned four WotC rulings dated 2018-06-08:
// targets precede helper selection, Assist pays generic mana only, colored
// costs stay with the caster, and the limit follows the spell's actual generic
// cost. Full response: .cardwork/assist-creatures/charging-binox-rulings.json.
// Read the run-pinned official Wizards Comprehensive Rules effective
// 2026-09-25 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt;
// verified local copy .cardwork/assist-creatures/MagicCompRules-20260925.txt
// (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca).
// Read 601.2f-h and 702.132a for locked cost, helper mana windows and
// generic-only contribution; 702.19a-b and 510.1c require attacking Trample
// to assign lethal damage to blockers before its excess may hit the player.
func TestChargingBinoxAssistAndTrample(t *testing.T) {
	g, engine, binox := castAssistCreature(t, chargingBinoxAssistOracleID, "Charging Binox", 7, mana.G)
	beginSicknessTurn(g, engine, game.Player1)
	if binox.SummoningSick {
		t.Fatal("Charging Binox remained summoning sick after its controller's next turn began")
	}
	resolveAssistCreatureTrampleCombat(t, g, engine, binox, 5)
}

// Source review: pinned-corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Magma Hellion Oracle ID 88f6575d-a6cb-4b2c-90dd-c6239f3c6e4d, printing
// revision ef9d02dc-4cb2-4ae1-84c3-e70aeb3c29ff (Battlebond #61), {6}{R},
// Creature — Hellion 5/4; exact Oracle text:
// "Assist (Another player can pay up to {6} of this spell's cost.)
// Trample, haste"
// Its Scryfall rulings URI
// https://api.scryfall.com/cards/ef9d02dc-4cb2-4ae1-84c3-e70aeb3c29ff/rulings
// was retrieved 2026-11-22 and returned four WotC rulings dated 2018-06-08:
// targets precede helper selection, Assist pays generic mana only, colored
// costs stay with the caster, and the limit follows the spell's actual generic
// cost. Full response: .cardwork/assist-creatures/magma-hellion-rulings.json.
// Read the run-pinned official Wizards Comprehensive Rules effective
// 2026-09-25 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt;
// verified local copy .cardwork/assist-creatures/MagicCompRules-20260925.txt
// (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca).
// Read 601.2f-h and 702.132a for locked cost and generic-only Assist payment,
// 302.6 and 702.10b for haste's exception to summoning sickness, and
// 702.19a-b and 510.1c for excess combat damage trampling over blockers.
func TestMagmaHellionAssistTrampleAndHaste(t *testing.T) {
	g, engine, hellion := castAssistCreature(t, magmaHellionAssistOracleID, "Magma Hellion", 6, mana.R)
	if !hellion.SummoningSick || !hasKeyword(g, hellion, game.Haste) {
		t.Fatal("Magma Hellion should be summoning sick and have haste immediately after entering")
	}
	resolveAssistCreatureTrampleCombat(t, g, engine, hellion, 3)
}

type assistCreatureChoiceAgent struct {
	simPassPolicy
	helperChosen game.PlayerID
	amountPaid   int
}

func (a *assistCreatureChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Prompt {
	case assistPlayerPrompt:
		for _, option := range request.Options {
			for _, target := range option.Targets {
				if target.Kind == game.TargetPlayer && target.PlayerID == game.Player2 {
					a.helperChosen = target.PlayerID
					return []int{option.Index}
				}
			}
		}
	case assistContributionPrompt:
		a.amountPaid = request.NumberRange.Val.Maximum.Val
		return []int{a.amountPaid}
	}
	return request.DefaultSelection
}

func castAssistCreature(t *testing.T, oracleID, name string, assistAmount int, casterMana mana.Color) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact Assist creature registry lookup failed: %v", err)
	}
	if def.Name != name {
		t.Fatalf("Oracle ID %s resolved to %q, want %q", oracleID, def.Name, name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(casterMana, 1)
	g.Players[game.Player2].ManaPool.Add(mana.C, assistAmount)
	cardID := addCardToHand(g, game.Player1, def)

	agent := &assistCreatureChoiceAgent{}
	agents := simPassPolicies()
	agents[game.Player1], agents[game.Player2] = agent, agent
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatalf("%s failed to cast with Assist", def.Name)
	}
	if agent.helperChosen != game.Player2 || agent.amountPaid != assistAmount {
		t.Fatalf("Assist choices = helper %v, contribution %d; want Player 2 and %d", agent.helperChosen, agent.amountPaid, assistAmount)
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 || g.Players[game.Player2].ManaPool.Total() != 0 {
		t.Fatalf("Assist did not charge the caster's colored cost and helper's generic cost: caster=%v helper=%v",
			g.Players[game.Player1].ManaPool, g.Players[game.Player2].ManaPool)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return g, engine, permanent
		}
	}
	t.Fatalf("%s did not enter the battlefield", def.Name)
	return nil, nil, nil
}

func resolveAssistCreatureTrampleCombat(t *testing.T, g *game.Game, engine *Engine, attacker *game.Permanent, expectedPlayerDamage int) {
	t.Helper()
	setAssistCreatureCombatStep(g, game.StepDeclareAttackers)
	g.Combat = &game.CombatState{}
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	attack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: attacker.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}}))
	if !engine.applyDeclareAttackers(g, game.Player1, attack) {
		t.Fatal("the Assist creature could not attack")
	}

	setAssistCreatureCombatStep(g, game.StepDeclareBlockers)
	blocks := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker:  blocker.ObjectID,
		Blocking: attacker.ObjectID,
	}}))
	if !engine.applyDeclareBlockers(g, game.Player2, blocks) {
		t.Fatal("the 2/2 blocker could not block the Assist creature")
	}

	life := g.Players[game.Player2].Life
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := life - g.Players[game.Player2].Life; got != expectedPlayerDamage {
		t.Fatalf("defending player took %d combat damage, want %d excess damage trampled over the 2/2 blocker",
			got, expectedPlayerDamage)
	}
	if blocker.MarkedDamage != 2 {
		t.Fatalf("blocker marked damage = %d, want the 2 lethal damage assigned before trample", blocker.MarkedDamage)
	}
}

func setAssistCreatureCombatStep(g *game.Game, step game.Step) {
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, step
}
