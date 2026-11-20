package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Oracle ID 45aa8bdb-54da-4069-bd6c-c63396154464, Oath of the
// Gatewatch #61, printing revision 33ff1000-1a4e-43f6-aa02-1dbe9fac6901;
// pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Current Oracle retrieved 2026-11-20 from
// https://api.scryfall.com/cards/33ff1000-1a4e-43f6-aa02-1dbe9fac6901:
// "{2}{U}{U} Instant. Surge {U}{U} (You may cast this spell for its surge
// cost if you or a teammate has cast another spell this turn.) This spell
// can't be countered. Counter target spell."
// Rulings retrieved the same day from
// https://api.scryfall.com/cards/33ff1000-1a4e-43f6-aa02-1dbe9fac6901/rulings:
// four WotC rulings dated 2016-01-22. Surge is only an alternative cost here;
// a qualifying spell may still be on the stack, may have resolved, or may
// have been countered; copies inherit paid Surge; and Surge doesn't change
// mana cost or mana value.
// Official Wizards Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Read 113.6g, 115.1a, 118.9a-d, 601.2b,c,e,f-i, 608.2b, 701.6a-b, and
// 702.117a: Surge is an alternative cost gated on a prior completed cast;
// the spell targets a spell on announcement, its target is rechecked on
// resolution, and its own can't-be-countered ability functions on the stack.
// Another spell may still target it, but that counter effect cannot remove it.
func TestOverwhelmingDenialSurgeUncounterableAndCountersTargetSpell(t *testing.T) {
	t.Parallel()

	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("45aa8bdb-54da-4069-bd6c-c63396154464")
	if err != nil {
		t.Fatalf("exact Overwhelming Denial lookup failed: %v", err)
	}
	if def.Name != "Overwhelming Denial" {
		t.Fatalf("exact lookup returned %q", def.Name)
	}
	counterspellDef := &game.CardDef{CardFace: game.CardFace{
		Name: "Counter probe", ManaCost: opt.Val(cost.Mana{cost.U, cost.U}),
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Text: "Counter target spell.",
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowStackObject,
				Predicate: game.TargetPredicate{
					StackObjectKinds: []game.StackObjectKind{game.StackSpell},
				},
			}},
			Sequence: []game.Instruction{{
				Primitive: game.CounterObject{
					Object: game.TargetStackObjectReference(0),
				},
			}},
		}.Ability()),
	}}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	playerOneLife := g.Players[game.Player1].Life
	playerTwoLife := g.Players[game.Player2].Life
	earlierSpellID := addCardToHand(g, game.Player1, simpleGainLifeInstant("Earlier spell"))
	targetSpellID := addCardToHand(g, game.Player2, simpleGainLifeInstant("Target spell"))
	denialID := addCardToHand(g, game.Player1, def)
	counterProbeID := addCardToHand(g, game.Player2, counterspellDef)
	g.Players[game.Player1].ManaPool.Add(mana.U, 2)
	g.Players[game.Player2].ManaPool.Add(mana.U, 2)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(earlierSpellID, nil, 0, nil)) {
		t.Fatal("earlier spell did not cast")
	}
	earlierSpell, ok := g.Stack.Peek()
	if !ok || earlierSpell.SourceID != earlierSpellID {
		t.Fatalf("earlier spell = %+v, want the earlier spell on the stack", earlierSpell)
	}

	g.Turn.PriorityPlayer = game.Player2
	if !engine.applyAction(g, game.Player2, action.CastSpell(targetSpellID, nil, 0, nil)) {
		t.Fatal("target spell did not cast")
	}
	targetSpell, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("target spell is missing from the stack")
	}

	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(
		denialID,
		[]game.Target{game.StackObjectTarget(targetSpell.ID)},
		0,
		nil,
	)) {
		t.Fatal("Overwhelming Denial did not cast targeting the opponent's spell")
	}
	denial, ok := g.Stack.Peek()
	if !ok || denial.SourceID != denialID {
		t.Fatalf("top of stack = %+v, want Overwhelming Denial", denial)
	}
	if denial.AlternativeCost != 1 || denial.ManaSpentToCast != 2 ||
		g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Overwhelming Denial did not pay Surge {U}{U}: stack=%+v, mana=%d",
			denial, g.Players[game.Player1].ManaPool.Total())
	}

	g.Turn.PriorityPlayer = game.Player2
	counterProbeOnDenial := action.CastSpell(
		counterProbeID,
		[]game.Target{game.StackObjectTarget(denial.ID)},
		0,
		nil,
	)
	if !containsAction(engine.legalActions(g, game.Player2), counterProbeOnDenial) {
		t.Fatal("counter probe could not legally target Overwhelming Denial")
	}
	if !engine.applyAction(g, game.Player2, counterProbeOnDenial) {
		t.Fatal("counter probe did not cast targeting Overwhelming Denial")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := stackObjectByID(g, denial.ID); !ok {
		t.Fatal("counter probe countered Overwhelming Denial")
	}
	if !g.Players[game.Player2].Graveyard.Contains(counterProbeID) ||
		g.Players[game.Player2].ManaPool.Total() != 0 {
		t.Fatal("counter probe did not resolve and pay its {U}{U} cost")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(targetSpellID) {
		t.Fatal("Overwhelming Denial did not counter its target spell")
	}
	if !g.Players[game.Player1].Graveyard.Contains(denialID) {
		t.Fatal("resolved Overwhelming Denial did not go to its owner's graveyard")
	}
	if g.Players[game.Player2].Life != playerTwoLife {
		t.Fatal("countered target spell resolved its life-gain effect")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != playerOneLife+1 {
		t.Fatal("the earlier spell did not remain on the stack and resolve")
	}
}
