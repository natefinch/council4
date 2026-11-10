package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Oracle ID 1cd9a3ba-a115-4e59-983d-605e4435f3a8, Scryfall printing
// 34406508-da49-4b63-b123-55f0924a0675, checked 2026-11-10:
// "Buyback {4} (You may pay an additional {4} as you cast this spell. If you do, put this card into your hand as it resolves.)"
// "Searing Touch deals 1 damage to any target." Rulings endpoint
// https://api.scryfall.com/cards/34406508-da49-4b63-b123-55f0924a0675/rulings returned 0 records.
// Official CR, effective 2026-11-06, verified at https://magic.wizards.com/en/rules
// and read from https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 115.1a/115.4, 120.1/120.3a/120.4b, 601.2b/c/f-h, 608.2b/c/n, 702.27a.
func TestSearingTouchScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g, engine, cardID, spell := castBuybackRepeaterScenario(
		t,
		g,
		authored.SearingTouch,
		game.PlayerTarget(game.Player2),
		[]mana.Color{mana.R, mana.C, mana.C, mana.C, mana.C},
	)
	if len(spell.Targets) != 1 || spell.Targets[0] != game.PlayerTarget(game.Player2) {
		t.Fatal("Searing Touch did not cast targeting Player2")
	}
	startingLife := g.Players[game.Player2].Life

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != startingLife-1 {
		t.Fatalf("target player's life = %d, want %d after 1 damage", got, startingLife-1)
	}
	assertBuybackRepeaterReturned(t, g, cardID)
}

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Oracle ID 2effcd1c-5833-4e0e-bb67-2d2ffb22a014, Scryfall printing
// 2294ef08-c27a-4bb4-bd59-434516717961, checked 2026-11-10:
// "Buyback {3} (You may pay an additional {3} as you cast this spell. If you do, put this card into your hand as it resolves.)"
// "Target creature gets +3/+0 until end of turn." Rulings endpoint
// https://api.scryfall.com/cards/2294ef08-c27a-4bb4-bd59-434516717961/rulings returned 0 records.
// Official CR, effective 2026-11-06, verified at https://magic.wizards.com/en/rules
// and read from https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 115.1a/115.2, 118.8a-b, 307.1, 601.2b/c/f-h, 608.2b/c/n, 611.2a/c,
// 613.4c, 514.2, 702.27a.
func TestSeethingAngerScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Seething Anger target", 2, 3))
	g, engine, cardID, spell := castBuybackRepeaterScenario(
		t,
		g,
		authored.SeethingAnger,
		game.PermanentTarget(target.ObjectID),
		[]mana.Color{mana.R, mana.C, mana.C, mana.C},
	)
	if len(spell.Targets) != 1 || spell.Targets[0] != game.PermanentTarget(target.ObjectID) {
		t.Fatal("Seething Anger did not cast targeting its creature")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, target); got != 5 {
		t.Fatalf("target creature power = %d, want 5 after +3/+0", got)
	}
	if got, ok := effectiveToughness(g, target); !ok || got != 3 {
		t.Fatalf("target creature toughness = %d, valid=%v, want 3", got, ok)
	}
	assertBuybackRepeaterReturned(t, g, cardID)
}

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Oracle ID 4f424d94-1544-4086-a57d-2add17332094, Scryfall printing
// 92bd531c-fdd1-4a22-816e-258b2975c1a3, checked 2026-11-10:
// "Buyback {2}{B}{B} (You may pay an additional {2}{B}{B} as you cast this spell. If you do, put this card into your hand as it resolves.)"
// "Target player discards a card." Rulings endpoint
// https://api.scryfall.com/cards/92bd531c-fdd1-4a22-816e-258b2975c1a3/rulings returned 0 records.
// Official CR, effective 2026-11-06, verified at https://magic.wizards.com/en/rules
// and read from https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 115.1a/115.2, 118.8a-b, 307.1, 601.2b/c/f-h, 608.2b/c/n, 701.9a-b,
// 702.27a.
func TestMindPeelScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g, engine, cardID, spell := castBuybackRepeaterScenario(
		t,
		g,
		authored.MindPeel,
		game.PlayerTarget(game.Player2),
		[]mana.Color{mana.B, mana.B, mana.B, mana.C, mana.C},
	)
	if len(spell.Targets) != 1 || spell.Targets[0] != game.PlayerTarget(game.Player2) {
		t.Fatal("Mind Peel did not cast targeting Player2")
	}
	chosen := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Mind Peel choice"}})
	kept := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Mind Peel kept"}})
	discarder := &orderedHandChoiceAgent{order: []string{"Mind Peel choice"}}
	agents := [game.NumPlayers]PlayerAgent{}
	agents[game.Player2] = discarder

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player2].Graveyard.Contains(chosen) || g.Players[game.Player2].Hand.Contains(chosen) {
		t.Fatal("target player did not discard the card they chose")
	}
	if !g.Players[game.Player2].Hand.Contains(kept) {
		t.Fatal("target player's unchosen card left their hand")
	}
	if len(discarder.requests) != 1 || discarder.requests[0].Player != game.Player2 ||
		discarder.requests[0].Kind != game.ChoiceResolution {
		t.Fatalf("discard choice = %+v, want target player to choose", discarder.requests)
	}
	assertBuybackRepeaterReturned(t, g, cardID)
}

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Oracle ID 557d2af0-04c5-461d-826b-7208c4535dc2, Scryfall printing
// e5a24943-0121-4c64-9624-2e4cff37e6ca, checked 2026-11-10:
// "Buyback {4} (You may pay an additional {4} as you cast this spell. If you do, put this card into your hand as it resolves.)"
// "Target creature gets +2/+2 until end of turn." Rulings endpoint
// https://api.scryfall.com/cards/e5a24943-0121-4c64-9624-2e4cff37e6ca/rulings returned one ruling dated 2019-06-14:
// "If the target creature is an illegal target by the time Elvish Fury tries to resolve, the spell doesn’t resolve. If its buyback cost was paid, you won’t return it to your hand."
// Official CR, effective 2026-11-06, verified at https://magic.wizards.com/en/rules
// and read from https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 115.1a/115.2, 118.8a-b, 601.2b/c/f-h, 608.2b/c/n, 611.2a/c, 613.4c,
// 514.2, 702.27a.
func TestElvishFuryScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Elvish Fury target", 2, 2))
	g, engine, cardID, spell := castBuybackRepeaterScenario(
		t,
		g,
		authored.ElvishFury,
		game.PermanentTarget(target.ObjectID),
		[]mana.Color{mana.G, mana.C, mana.C, mana.C, mana.C},
	)
	if len(spell.Targets) != 1 || spell.Targets[0] != game.PermanentTarget(target.ObjectID) {
		t.Fatal("Elvish Fury did not cast targeting its creature")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, target); got != 4 {
		t.Fatalf("target creature power = %d, want 4 after +2/+2", got)
	}
	if got, ok := effectiveToughness(g, target); !ok || got != 4 {
		t.Fatalf("target creature toughness = %d, valid=%v, want 4", got, ok)
	}
	assertBuybackRepeaterReturned(t, g, cardID)
}

func castBuybackRepeaterScenario(
	t *testing.T,
	g *game.Game,
	entry cardset.Entry,
	target game.Target,
	manaToPay []mana.Color,
) (*game.Game, *Engine, id.ID, *game.StackObject) {
	t.Helper()
	registry, engine := buybackInteractionRuntime(t)
	def, err := registry.LookupOracleID(entry.OracleID)
	if err != nil || def == nil || def.Name != entry.Name {
		t.Fatalf("exact production lookup for %s failed: %v", entry.Name, err)
	}
	cardID := addCardToHand(g, game.Player1, def)
	for _, color := range manaToPay {
		g.Players[game.Player1].ManaPool.Add(color, 1)
	}
	setMainPhasePriority(g, game.Player1)
	targets := []game.Target{target}
	if !engine.applyActionWithChoices(g, game.Player1,
		action.CastSpell(cardID, targets, 0, nil),
		buybackInteractionChoiceAgent(t, game.Player1, true),
		&TurnLog{},
	) {
		t.Fatalf("%s did not cast with its printed cost and buyback", entry.Name)
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.Kind != game.StackSpell || spell.SourceID != cardID ||
		!spell.BuybackPaid || spell.ManaSpentToCast != len(manaToPay) ||
		g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("%s cast did not record its target and chosen buyback payment: %+v", entry.Name, spell)
	}
	return g, engine, cardID, spell
}

func assertBuybackRepeaterReturned(t *testing.T, g *game.Game, cardID id.ID) {
	t.Helper()
	if !g.Players[game.Player1].Hand.Contains(cardID) ||
		g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("resolved spell with paid buyback did not return to its owner's hand")
	}
}
