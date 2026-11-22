package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
)

const gangUpOracleID = "dea311c3-9be3-4059-a206-5fbbdca4d6c9"

// Source review: BBD #47 printing 10d01449-3e4e-44ef-90aa-9489c86c57df;
// pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Oracle ID dea311c3-9be3-4059-a206-5fbbdca4d6c9, exact text:
// "Assist (Another player can pay up to {X} of this spell's cost. You choose
// the value of X.) Destroy target creature with power X or less."
// Scryfall rulings URI
// https://api.scryfall.com/cards/10d01449-3e4e-44ef-90aa-9489c86c57df/rulings
// retrieved 2026-11-22: four WotC rulings, all dated 2018-06-08. They specify
// that targets are chosen before the helper; Assist pays only generic mana,
// leaving colored costs to the caster; color-flexible mana does not make
// colored costs generic; and cost changes adjust how much generic the helper
// may pay. The full response is retained in .cardwork/sources/gang-up-rulings.json.
// Read the official Wizards Comprehensive Rules effective 2026-11-06 from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified from https://magic.wizards.com/en/rules; SHA-256
// 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829; full TXT
// retained in .cardwork/sources/MagicCompRules-20261106.txt). Sections 107.3a/i and
// 601.2b establish the announced X and its shared value in cost and text;
// 115.1a and 601.2c place creature-target choice during casting; 601.2f-h and
// 702.132a lock the total cost, allow the chosen helper to activate mana, and
// let them pay only generic mana before the caster pays; 208.1, 613.4a-d and
// 613.5 determine current power; 608.2b rechecks target legality on resolution;
// and 701.8a-c define destruction and its regeneration replacement.
func TestGangUpAssistXTargetCreatureAndDestroys(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(gangUpOracleID)
	if err != nil {
		t.Fatal(err)
	}
	if def.Name != "Gang Up" {
		t.Fatalf("Oracle ID lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	atX := addCombatPermanent(g, game.Player2, vanillaCreature("Power two", 2, 2))
	aboveX := addCombatPermanent(g, game.Player2, vanillaCreature("Power three", 3, 3))
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player2].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player1)

	valid := action.CastSpell(spell, []game.Target{game.PermanentTarget(atX.ObjectID)}, 2, nil)
	invalid := action.CastSpell(spell, []game.Target{game.PermanentTarget(aboveX.ObjectID)}, 2, nil)
	legal := engine.legalActions(g, game.Player1)
	if !containsAction(legal, valid) {
		t.Fatal("Gang Up at X=2 was not offered with a legal power-2 target and an assisting payer")
	}
	if containsAction(legal, invalid) || engine.applyAction(g.Clone(), game.Player1, invalid) {
		t.Fatal("Gang Up at X=2 accepted a creature with power greater than X")
	}

	assistantChosen, contributionPaid := false, false
	agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
		switch request.Prompt {
		case assistPlayerPrompt:
			for _, option := range request.Options {
				if len(option.Targets) == 1 && option.Targets[0].Kind == game.TargetPlayer &&
					option.Targets[0].PlayerID == game.Player2 {
					assistantChosen = true
					return []int{option.Index}
				}
			}
			t.Fatal("Gang Up did not offer Player 2 as an Assist payer")
		case assistContributionPrompt:
			if request.Player != game.Player2 || request.NumberRange.Val.Maximum.Val != 2 {
				t.Fatalf("Assist contribution request = %+v, want Player 2 paying up to {2}", request)
			}
			contributionPaid = true
			return []int{2}
		}
		return request.DefaultSelection
	}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1], agents[game.Player2] = agent, agent
	if !engine.applyActionWithChoices(g, game.Player1, valid, agents, &TurnLog{}) {
		t.Fatal("Gang Up failed to cast for X=2 with the helper paying its generic cost")
	}
	if !assistantChosen || !contributionPaid {
		t.Fatal("Gang Up did not choose Player 2 and accept the two-generic Assist contribution")
	}
	spellObject, ok := g.Stack.Peek()
	if !ok || spellObject.XValue != 2 || spellObject.ManaSpentToCast != 3 ||
		spellObject.ManaSpentByColorToCast[color.Black] != 1 {
		t.Fatalf("Gang Up cast bookkeeping = %+v; want X=2 and {2}{B} paid", spellObject)
	}
	if !g.Players[game.Player1].ManaPool.IsEmpty() || !g.Players[game.Player2].ManaPool.IsEmpty() {
		t.Fatal("Gang Up did not consume the helper's generic mana and the caster's black mana")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(atX.CardInstanceID) ||
		!g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("Gang Up did not destroy the power-2 target and finish in the caster's graveyard")
	}
	if _, stillOnBattlefield := permanentByObjectID(g, aboveX.ObjectID); !stillOnBattlefield {
		t.Fatal("Gang Up destroyed the power-3 creature that was not its target")
	}
}
