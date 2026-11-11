package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Recruit the Worthy, MH1 #22, Oracle ID e25e2a3f-4add-4744-b1f8-2e880998a51f,
// printing revision fb025c0c-7fdf-4f71-afd5-ef11cd2804eb. Oracle text:
// "Buyback {3} (You may pay an additional {3} as you cast this spell. If you do, put this card into your hand as it resolves.)
// Create a 1/1 white Soldier creature token." Rulings URI
// https://api.scryfall.com/cards/fb025c0c-7fdf-4f71-afd5-ef11cd2804eb/rulings
// returned 0 records on 2026-11-11. Official Wizards CR effective 2026-11-06,
// verified at https://magic.wizards.com/en/rules and read from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// 601.2b/f-h and 702.27a require announcing and paying buyback as an additional
// casting cost, 701.7a creates the Soldier token, 111.2 gives it to the player
// who created it, and 111.4 supplies the unspecified token name. Rule 608.2c
// resolves its instructions in order before buyback returns the spell to its owner.
func TestRecruitTheWorthyScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	registry, engine := buybackInteractionRuntime(t)
	cardID, spell := castSupportedBuybackSpell(
		t, g, registry, engine, authored.RecruitTheWorthy, nil,
		[]mana.Color{mana.W, mana.C, mana.C, mana.C},
	)
	if !spell.BuybackPaid {
		t.Fatal("Recruit the Worthy did not record its buyback payment")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	assertBuybackSpellReturned(t, g, cardID)
	soldiers := 0
	for _, permanent := range g.Battlefield {
		if !permanent.Token || permanent.TokenDef == nil || permanent.TokenDef.Name != "Soldier Token" {
			continue
		}
		soldiers++
		token := permanent.TokenDef
		if permanent.Controller != game.Player1 || permanent.Owner != game.Player1 ||
			permanent.Tapped || !slices.Equal(token.Types, []types.Card{types.Creature}) ||
			!slices.Equal(token.Subtypes, []types.Sub{types.Soldier}) ||
			!slices.Equal(token.Colors, []color.Color{color.White}) ||
			!token.Power.Exists || token.Power.Val.Value != 1 ||
			!token.Toughness.Exists || token.Toughness.Val.Value != 1 {
			t.Fatalf("wrong Soldier token controller or characteristics: %+v / %+v", permanent, token)
		}
	}
	if soldiers != 1 {
		t.Fatalf("Recruit the Worthy created %d Soldier tokens, want one", soldiers)
	}
}

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Mind Games, STH #38, Oracle ID ea43372f-b0a7-4d1b-be3e-a4a738b06d59,
// printing revision 4da50979-1f5d-48d1-9406-dfc785273c04. Oracle text:
// "Buyback {2}{U} (You may pay an additional {2}{U} as you cast this spell. If you do, put this card into your hand as it resolves.)
// Tap target artifact, creature, or land." Rulings URI
// https://api.scryfall.com/cards/4da50979-1f5d-48d1-9406-dfc785273c04/rulings
// returned 0 records on 2026-11-11. Official Wizards CR effective 2026-11-06,
// verified at https://magic.wizards.com/en/rules and read from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// 115.1a and 601.2c require choosing the target as the spell is cast;
// 608.2b rechecks its legality, 701.26a taps the permanent, and 702.27a applies
// the paid-buyback replacement only as the spell resolves.
func TestMindGamesScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addBattlefieldPermanent(g, game.Player2, "Mind Games land", []types.Card{types.Land})
	registry, engine := buybackInteractionRuntime(t)
	targetRef := game.PermanentTarget(target.ObjectID)
	cardID, spell := castSupportedBuybackSpell(
		t, g, registry, engine, authored.MindGames, []game.Target{targetRef},
		[]mana.Color{mana.U, mana.U, mana.C, mana.C},
	)
	if len(spell.Targets) != 1 || spell.Targets[0] != targetRef {
		t.Fatal("Mind Games did not retain the chosen land target")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if !target.Tapped {
		t.Fatal("Mind Games did not tap its target land")
	}
	assertBuybackSpellReturned(t, g, cardID)
}

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Haze of Rage, TSR #169, Oracle ID f17d0fb8-c157-43b8-be26-f5ba4c6aed14,
// printing revision c344b885-68c6-43d2-b6c1-6c89b3c94983. Oracle text:
// "Buyback {2} (You may pay an additional {2} as you cast this spell. If you do, put this card into your hand as it resolves.)
// Creatures you control get +1/+0 until end of turn.
// Storm (When you cast this spell, copy it for each spell cast before it this turn.)"
// Rulings URI https://api.scryfall.com/cards/c344b885-68c6-43d2-b6c1-6c89b3c94983/rulings
// returned six WotC rulings on 2026-11-11, including the 2021-03-19 ruling
// that the pump applies only to creatures controlled on resolution and that
// earlier casts count for storm but prior copies do not, and 2022-12-08 rulings
// that storm copies are put on the stack, aren't cast, and aren't counted by
// later storm spells. Official Wizards CR effective 2026-11-06, verified at
// https://magic.wizards.com/en/rules and read from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// 601.2b/f-h/i and 702.27a govern buyback casting and resolution; 702.40a-b and
// 707.10 govern storm's cast-triggered copies; 611.2a and 613.4c require the
// resolving +1/+0 effect to last through end of turn in layer 7c.
func TestHazeOfRageScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	registry, engine := buybackInteractionRuntime(t)
	earlierSpell := addCardToHand(g, game.Player1, simpleGainLifeInstant("Earlier spell"))
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(earlierSpell, nil, 0, nil)) {
		t.Fatal("setup spell did not cast before Haze of Rage")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	ally := addCombatPermanent(g, game.Player1, vanillaCreature("Haze ally", 2, 2))
	opponent := addCombatPermanent(g, game.Player2, vanillaCreature("Haze opponent", 3, 3))
	cardID, spell := castSupportedBuybackSpell(
		t, g, registry, engine, authored.HazeOfRage, nil,
		[]mana.Color{mana.R, mana.C, mana.C, mana.C},
	)
	if !spell.BuybackPaid {
		t.Fatal("Haze of Rage did not record its buyback payment")
	}

	resolveStormTrigger(t, engine, g, nil)
	if g.Stack.Size() != 2 {
		t.Fatalf("Haze storm left %d stack objects, want its original plus one copy", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("resolving the storm copy returned the physical Haze of Rage card")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, ally); got != 4 {
		t.Fatalf("Haze ally power = %d, want 4 after the spell and its one storm copy", got)
	}
	if got := effectivePower(g, opponent); got != 3 {
		t.Fatalf("opponent creature power = %d, want unchanged 3", got)
	}
	laterAlly := addCombatPermanent(g, game.Player1, vanillaCreature("Later Haze ally", 4, 4))
	if got := effectivePower(g, laterAlly); got != 4 {
		t.Fatalf("later creature power = %d, want unaffected 4", got)
	}
	assertBuybackSpellReturned(t, g, cardID)
}

func castSupportedBuybackSpell(
	t *testing.T,
	g *game.Game,
	registry *cards.Registry,
	engine *Engine,
	entry cardset.Entry,
	targets []game.Target,
	manaToPay []mana.Color,
) (id.ID, *game.StackObject) {
	t.Helper()
	def, err := registry.LookupOracleID(entry.OracleID)
	if err != nil || def == nil || def.Name != entry.Name {
		t.Fatalf("exact production lookup for %s failed: %v", entry.Name, err)
	}
	cardID := addCardToHand(g, game.Player1, def)
	for _, color := range manaToPay {
		g.Players[game.Player1].ManaPool.Add(color, 1)
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyActionWithChoices(
		g, game.Player1, action.CastSpell(cardID, targets, 0, nil),
		buybackInteractionChoiceAgent(t, game.Player1, true), &TurnLog{},
	) {
		t.Fatalf("%s did not cast with its printed cost and buyback", entry.Name)
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.Kind != game.StackSpell || spell.SourceID != cardID ||
		!spell.BuybackPaid || spell.ManaSpentToCast != len(manaToPay) ||
		g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("%s did not record its target and chosen buyback payment: %+v", entry.Name, spell)
	}
	return cardID, spell
}

func assertBuybackSpellReturned(t *testing.T, g *game.Game, cardID id.ID) {
	t.Helper()
	if !g.Players[game.Player1].Hand.Contains(cardID) ||
		g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("resolved spell with paid buyback did not return to its owner's hand")
	}
}
