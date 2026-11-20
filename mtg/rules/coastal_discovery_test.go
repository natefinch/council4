package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: exact BFZ #73 printing c1df2411-0b0d-4b93-bf30-a22487b3c6ef,
// Oracle ID f7c84690-8c7c-41a6-b430-a9771030f903; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. Current
// Oracle verified 2026-11-19 at
// https://api.scryfall.com/cards/c1df2411-0b0d-4b93-bf30-a22487b3c6ef:
// "{3}{U} Sorcery. Draw two cards. Awaken 4—{5}{U} (If you cast this spell for
// {5}{U}, also put four +1/+1 counters on target land you control and it becomes
// a 0/0 Elemental creature with haste. It's still a land.)" Its rulings endpoint
// returned six WotC rulings dated 2015-08-25: the Awaken cost is optional; the
// land target exists only on that cost; all-illegal-targets means no resolution;
// the land retains its types, subtypes, mana abilities and color. Wizards
// Comprehensive Rules effective 2026-11-06, official copy
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt; read
// 601.2b-c/e/f, 608.2b, 611.2a/c, 122.1a, 205.1b, 613.1d/f/g, 613.4b-c,
// 302.6, 702.10b and 702.113a-b. Awaken replaces the mana cost and adds a target only
// on that cast branch; its effect adds counters, retains non-creature types and
// subtypes, adds Elemental creature/haste, and sets base power/toughness to 0/0.
func coastalDiscoveryEngine(t *testing.T) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog(authored.CoastalDiscovery.Programs)
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

func TestCoastalDiscoveryAwakenAlternativeDrawsAndAnimatesForest(t *testing.T) {
	engine := coastalDiscoveryEngine(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, authored.CoastalDiscovery.New())
	forestDef := &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Forest},
	}}
	forest := addManaAbilityPermanent(g, game.Player1, forestDef, mana.G, 1)
	forest.SummoningSick = true
	firstDraw := addCardToLibraryNamed(g, game.Player1, "Coastal Discovery draw one")
	secondDraw := addCardToLibraryNamed(g, game.Player1, "Coastal Discovery draw two")
	g.Players[game.Player1].ManaPool.Add(mana.C, 5)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)

	cast := action.CastSpell(spellID, []game.Target{game.PermanentTarget(forest.ObjectID)}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("Awaken cast with {5}{U} and a controlled Forest is not legal")
	}
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Awaken cast failed")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 1 || spell.ManaSpentToCast != 6 || len(spell.Targets) != 1 {
		t.Fatalf("cast spell = %+v, want Awaken alternative, six mana spent, one target", spell)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(firstDraw) || !g.Players[game.Player1].Hand.Contains(secondDraw) {
		t.Fatal("Coastal Discovery did not draw both cards")
	}
	if forest.Counters.Get(counter.PlusOnePlusOne) != 4 {
		t.Fatalf("Forest +1/+1 counters = %d, want 4", forest.Counters.Get(counter.PlusOnePlusOne))
	}
	if !permanentHasType(g, forest, types.Land) || !permanentHasType(g, forest, types.Creature) ||
		!permanentHasSubtype(g, forest, types.Forest) || !permanentHasSubtype(g, forest, types.Elemental) {
		t.Fatal("Awakened Forest did not retain its land type/subtype and become an Elemental creature")
	}
	if got := effectivePower(g, forest); got != 4 {
		t.Fatalf("Awakened Forest power = %d, want 4 from 0 base plus four counters", got)
	}
	if got := effectiveToughnessMust(t, g, forest); got != 4 {
		t.Fatalf("Awakened Forest toughness = %d, want 4 from 0 base plus four counters", got)
	}
	if !hasKeyword(g, forest, game.Haste) {
		t.Fatal("Awakened Forest does not have haste")
	}
	setMainPhasePriority(g, game.Player1)
	manaAbility := action.ActivateAbility(forest.ObjectID, forestDef.ManaAbilityIndex(0), nil, 0)
	if !engine.applyAction(g, game.Player1, manaAbility) {
		t.Fatal("newly animated Forest could not use its retained mana ability with haste")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("green mana from the animated Forest = %d, want 1", got)
	}
}

func TestCoastalDiscoveryBaseCostDrawsWithoutTarget(t *testing.T) {
	engine := coastalDiscoveryEngine(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, authored.CoastalDiscovery.New())
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	firstDraw := addCardToLibraryNamed(g, game.Player1, "Coastal Discovery draw one")
	secondDraw := addCardToLibraryNamed(g, game.Player1, "Coastal Discovery draw two")
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)

	cast := action.CastSpell(spellID, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("base-cost draw-only cast should be legal without targets")
	}
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("base-cost cast failed")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 0 || spell.ManaSpentToCast != 4 || len(spell.Targets) != 0 {
		t.Fatalf("cast spell = %+v, want base cost, four mana spent, no targets", spell)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(firstDraw) || !g.Players[game.Player1].Hand.Contains(secondDraw) {
		t.Fatal("Coastal Discovery did not draw both cards")
	}
	if permanentHasType(g, forest, types.Creature) || forest.Counters.Get(counter.PlusOnePlusOne) != 0 {
		t.Fatal("base-cost Coastal Discovery applied its Awaken effect")
	}
}
