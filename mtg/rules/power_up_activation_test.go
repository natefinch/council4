package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Official Wizards Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 702.193a-b binds Power-up to the entering permanent and permits one use;
// 118.7a-c subtract generic and matching colored symbols separately, with
// excess colored reduction lowering generic; 601.2f-h and 602.2b lock the
// total before mana abilities or payment; 400.7 and 602.5b track use on the
// object, not the physical card or controller; 122.1a defines the counters;
// 707.2 includes mana cost in copiable values.

type powerUpDiscountProbe struct {
	name           string
	oracleText     string
	powerUpText    string
	sourceCost     cost.Mana
	sourceGeneric  int
	sourceColor    mana.Color
	activationCost cost.Mana
	reducedGeneric int
	counters       int
	types          []types.Card
}

func runPowerUpDiscountProbe(t *testing.T, probe powerUpDiscountProbe) {
	t.Helper()
	def := &game.CardDef{CardFace: game.CardFace{
		Name: probe.name, ManaCost: opt.Val(probe.sourceCost), Types: probe.types,
		Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
		OracleText: probe.oracleText,
		ActivatedAbilities: []game.ActivatedAbility{
			powerUpCounterAbility(probe.powerUpText, probe.activationCost, probe.counters),
		},
	}}
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("fixture declaration rejected: %+v", issues)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].ManaPool.Add(mana.C, probe.sourceGeneric)
	g.Players[game.Player1].ManaPool.Add(probe.sourceColor, 1)
	source := castPowerUpSource(t, engine, g, def)
	abilityAction := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	pool := &g.Players[game.Player1].ManaPool
	pool.Add(mana.C, probe.reducedGeneric-1)
	if containsAction(engine.legalActions(g, game.Player1), abilityAction) || engine.applyAction(g, game.Player1, abilityAction) {
		t.Fatal("Power-up was legal with one generic mana short")
	}
	pool.Add(mana.C, 1)
	playPowerUp(t, engine, g, source.ObjectID, probe.reducedGeneric, probe.counters, [game.NumPlayers]PlayerAgent{})
}

func powerUpCounterAbility(text string, manaCost cost.Mana, counters int) game.ActivatedAbility {
	return game.ActivatedAbility{
		Text: text, ManaCost: opt.Val(manaCost), ActivateOnlyOnce: true,
		ReduceBySourceManaCostIfEnteredThisTurn: true,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddCounter{
			Object: game.SourcePermanentReference(), CounterKind: counter.PlusOnePlusOne, Amount: game.Fixed(counters),
		}}}}.Ability(),
	}
}

func castPowerUpSource(t *testing.T, engine *Engine, g *game.Game, def *game.CardDef) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("could not cast source")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	source, ok := findPermanentByCardID(g, cardID)
	if !ok || !permanentEnteredThisTurn(g, source.ObjectID) {
		t.Fatal("source did not enter the battlefield this turn")
	}
	setMainPhasePriority(g, game.Player1)
	return source
}

func playPowerUp(t *testing.T, engine *Engine, g *game.Game, sourceID id.ID, manaSpent, countersAdded int, agents [game.NumPlayers]PlayerAgent) *TurnLog {
	t.Helper()
	act := action.ActivateAbility(sourceID, 0, nil, 0)
	pool := &g.Players[game.Player1].ManaPool
	spent := pool.Spent()
	source, ok := permanentByObjectID(g, sourceID)
	if !ok {
		t.Fatal("missing source")
	}
	before := source.Counters.Get(counter.PlusOnePlusOne)
	log := &TurnLog{}
	if !containsAction(engine.legalActions(g, game.Player1), act) || !engine.applyActionWithChoices(g, game.Player1, act, agents, log) {
		t.Fatalf("Power-up was not offered and payable with %d floating mana", pool.Total())
	}
	if got := pool.Spent() - spent; got != manaSpent {
		t.Fatalf("spent %d mana, want %d", got, manaSpent)
	}
	engine.resolveTopOfStack(g, log)
	if got := source.Counters.Get(counter.PlusOnePlusOne) - before; got != countersAdded {
		t.Fatalf("placed %d +1/+1 counters, want %d", got, countersAdded)
	}
	return log
}

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; MSH 186, printing revision 72727c2a-08a9-47bc-a5eb-cda52e21684b; Oracle: “Deathtouch / Power-up — {3}{G}: Put two +1/+1 counters on this creature. (Activate each power-up ability only once. Reduce the cost by its mana cost if it entered this turn.)” Rulings: https://api.scryfall.com/cards/72727c2a-08a9-47bc-a5eb-cda52e21684b/rulings was retrieved on 2026-11-22 and returned an empty list. Official Wizards Comprehensive Rules effective 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt; read 601.2f, 602.2b, 602.5b, 118.7a-c, 122.1a, and 400.7: activated abilities determine/lock their activation costs like spells; generic and colored reductions affect distinct mana components; +1/+1 counters change creature power/toughness; a zone change creates a new object.
func TestSerpentSpecialistPowerUpDiscountGapProbe(t *testing.T) {
	runPowerUpDiscountProbe(t, powerUpDiscountProbe{
		name: "Serpent Specialist", oracleText: "Deathtouch\nPower-up — {3}{G}: Put two +1/+1 counters on this creature.",
		powerUpText: "Power-up — {3}{G}: Put two +1/+1 counters on this creature.",
		sourceCost:  cost.Mana{cost.G}, sourceColor: mana.G,
		activationCost: cost.Mana{cost.O(3), cost.G},
		reducedGeneric: 3, counters: 2, types: []types.Card{types.Creature},
	})
}

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; MSH 8, printing revision 2242b5c3-42ef-4be0-a61f-65c93e56fcab; Oracle: “Lifelink / Power-up — {4}{W}: Put two +1/+1 counters on this creature. (Activate each power-up ability only once. Reduce the cost by its mana cost if it entered this turn.)” Rulings: https://api.scryfall.com/cards/2242b5c3-42ef-4be0-a61f-65c93e56fcab/rulings was retrieved on 2026-11-22 and returned an empty list. Official Wizards Comprehensive Rules effective 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt; read 601.2f, 602.2b, 602.5b, 118.7a-c, 122.1a, and 400.7: activated abilities determine/lock their activation costs like spells; generic and colored reductions affect distinct mana components; +1/+1 counters change creature power/toughness; a zone change creates a new object.
func TestBraveBrawlerPowerUpDiscountGapProbe(t *testing.T) {
	runPowerUpDiscountProbe(t, powerUpDiscountProbe{
		name: "Brave Brawler", oracleText: "Lifelink\nPower-up — {4}{W}: Put two +1/+1 counters on this creature.",
		powerUpText: "Power-up — {4}{W}: Put two +1/+1 counters on this creature.",
		sourceCost:  cost.Mana{cost.O(1), cost.W}, sourceGeneric: 1, sourceColor: mana.W,
		activationCost: cost.Mana{cost.O(4), cost.W},
		reducedGeneric: 3, counters: 2, types: []types.Card{types.Creature},
	})
}

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; MSH 159, printing revision 4f48ba40-9934-40fd-a251-e15a68e772d2; Oracle: “Haste / Power-up — {5}{R}: Put two +1/+1 counters on this creature. (Activate each power-up ability only once. Reduce the cost by its mana cost if it entered this turn.)” Rulings: https://api.scryfall.com/cards/4f48ba40-9934-40fd-a251-e15a68e772d2/rulings was retrieved on 2026-11-22 and returned an empty list. Official Wizards Comprehensive Rules effective 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt; read 601.2f, 602.2b, 602.5b, 118.7a-c, 122.1a, and 400.7: activated abilities determine/lock their activation costs like spells; generic and colored reductions affect distinct mana components; +1/+1 counters change creature power/toughness; a zone change creates a new object.
func TestVolcanicVillainPowerUpDiscountGapProbe(t *testing.T) {
	runPowerUpDiscountProbe(t, powerUpDiscountProbe{
		name: "Volcanic Villain", oracleText: "Haste\nPower-up — {5}{R}: Put two +1/+1 counters on this creature.",
		powerUpText: "Power-up — {5}{R}: Put two +1/+1 counters on this creature.",
		sourceCost:  cost.Mana{cost.O(2), cost.R}, sourceGeneric: 2, sourceColor: mana.R,
		activationCost: cost.Mana{cost.O(5), cost.R},
		reducedGeneric: 3, counters: 2, types: []types.Card{types.Creature},
	})
}

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; MSH 43, printing revision e727ec1c-dc3b-4f1a-8a62-18549f118b89; Oracle: “Flying / Power-up — {5}{U}: Put three +1/+1 counters on this creature. (Activate each power-up ability only once. Reduce the cost by its mana cost if it entered this turn.)” Rulings: https://api.scryfall.com/cards/e727ec1c-dc3b-4f1a-8a62-18549f118b89/rulings was retrieved on 2026-11-22 and returned an empty list. Official Wizards Comprehensive Rules effective 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt; read 601.2f, 602.2b, 602.5b, 118.7a-c, 122.1a, and 400.7: activated abilities determine/lock their activation costs like spells; generic and colored reductions affect distinct mana components; +1/+1 counters change creature power/toughness; a zone change creates a new object.
func TestAerialDoombotPowerUpDiscountGapProbe(t *testing.T) {
	runPowerUpDiscountProbe(t, powerUpDiscountProbe{
		name: "Aerial Doombot", oracleText: "Flying\nPower-up — {5}{U}: Put three +1/+1 counters on this creature.",
		powerUpText: "Power-up — {5}{U}: Put three +1/+1 counters on this creature.",
		sourceCost:  cost.Mana{cost.U}, sourceColor: mana.U,
		activationCost: cost.Mana{cost.O(5), cost.U},
		reducedGeneric: 5, counters: 3, types: []types.Card{types.Artifact, types.Creature},
	})
}

func powerUpFixture(manaCost, activationCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Power-up fixture", ManaCost: opt.Val(manaCost), Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
		ActivatedAbilities: []game.ActivatedAbility{powerUpCounterAbility("Power-up", activationCost, 2)},
	}}
}

func TestPowerUpDiscountFollowsExactEntryObject(t *testing.T) {
	g, engine := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
	def := powerUpFixture(cost.Mana{cost.O(1), cost.W}, cost.Mana{cost.O(4), cost.W})
	pool := &g.Players[game.Player1].ManaPool
	pool.Add(mana.C, 1)
	pool.Add(mana.W, 1)
	older := castPowerUpSource(t, engine, g, def)
	engine.advanceToNextTurn(g)
	pool.Add(mana.C, 1)
	pool.Add(mana.W, 1)
	fresh := castPowerUpSource(t, engine, g, def)

	pool.Add(mana.C, 4)
	olderAction := action.ActivateAbility(older.ObjectID, 0, nil, 0)
	if containsAction(engine.legalActions(g, game.Player1), olderAction) || engine.applyAction(g, game.Player1, olderAction) {
		t.Fatal("older source got a discount from the identical source entering this turn")
	}
	pool.Add(mana.W, 1)
	playPowerUp(t, engine, g, older.ObjectID, 5, 2, [game.NumPlayers]PlayerAgent{})
	pool.Add(mana.C, 3)
	playPowerUp(t, engine, g, fresh.ObjectID, 3, 2, [game.NumPlayers]PlayerAgent{})

	pool.Add(mana.C, 4)
	pool.Add(mana.W, 1)
	if containsAction(engine.legalActions(g, game.Player1), olderAction) || engine.applyAction(g, game.Player1, olderAction) {
		t.Fatal("ability could be used twice on the same object")
	}
	pool.Empty()
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1,
		Targets: []game.Target{game.PermanentTarget(older.ObjectID)},
	}
	for _, instr := range immediateBlinkInstructions(game.PutOnBattlefield{}) {
		engine.resolveInstructionWithChoices(g, obj, &instr, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	}
	returned, ok := findPermanentByCardID(g, older.CardInstanceID)
	if !ok || returned.ObjectID == older.ObjectID {
		t.Fatal("blink did not return a new source object")
	}
	pool.Add(mana.C, 3)
	playPowerUp(t, engine, g, returned.ObjectID, 3, 2, [game.NumPlayers]PlayerAgent{})
}

func TestPowerUpDiscountUsesCurrentManaCost(t *testing.T) {
	g, engine := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
	blueprint := addCombatPermanent(g, game.Player1, powerUpFixture(cost.Mana{cost.O(1), cost.W}, cost.Mana{cost.O(4), cost.W}))
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	source := castPowerUpSource(t, engine, g, &game.CardDef{CardFace: game.CardFace{
		Name: "Copy recipient", ManaCost: opt.Val(cost.Mana{cost.U}), Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
	}})
	becomeStaticCopy(t, g, source, blueprint, false)

	// The source's new {1}{W} cost applies, not the {U} it was cast for.
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	playPowerUp(t, engine, g, source.ObjectID, 3, 2, [game.NumPlayers]PlayerAgent{})
}

func TestPowerUpDiscountCombinedCostsCloneAndReplay(t *testing.T) {
	g, engine := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
	// {W} has no match in either activation option, so it reduces generic.
	def := powerUpFixture(cost.Mana{cost.W}, cost.Mana{cost.O(2), cost.R})
	def.ActivatedAbilities[0].CostModifiers = []game.CostModifier{{
		Kind: game.CostModifierAbility, GenericIncrease: 1,
	}}
	def.ActivatedAbilities[0].AlternativeCosts = []cost.Alternative{{
		Label: "Pay life and mana", ManaCost: opt.Val(cost.Mana{cost.O(3)}),
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2}},
	}}
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	source := castPowerUpSource(t, engine, g, def)
	pool := &g.Players[game.Player1].ManaPool
	pool.Add(mana.C, 2)
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	if containsAction(engine.legalActions(g, game.Player1), act) || engine.applyAction(g, game.Player1, act) {
		t.Fatal("alternative was offered when neither combined payment was possible")
	}
	pool.Add(mana.R, 1)
	initial, life := g.Clone(), g.Players[game.Player1].Life
	makeChoice := func(selectIndex int) [game.NumPlayers]PlayerAgent {
		return conspireTestAgent(func(req game.ChoiceRequest) []int {
			if req.Prompt == "Choose ability cost" {
				if len(req.Options) != 2 {
					t.Fatalf("expected both cost options: %v", req.Options)
				}
				return []int{selectIndex}
			}
			return nil
		})
	}
	log := playPowerUp(t, engine, g, source.ObjectID, 3, 2, makeChoice(1))
	if g.Players[game.Player1].Life != life-2 || len(log.Choices) == 0 {
		t.Fatal("alternative lost the nonmana payment or was not recorded")
	}
	replayed := initial.Clone()
	playPowerUp(t, engine, replayed, source.ObjectID, 3, 2, replayAgentsFromTurnLog(log))
	if replayed.Players[game.Player1].Life != life-2 {
		t.Fatal("replay did not choose the same alternative")
	}
	playPowerUp(t, engine, initial, source.ObjectID, 3, 2, makeChoice(0))
	if initial.Players[game.Player1].Life != life {
		t.Fatal("declining the alternative paid its life cost")
	}
}

func TestPowerUpDiscountRejectsUnsupportedDeclarations(t *testing.T) {
	t.Run("off_battlefield", func(t *testing.T) {
		def := powerUpFixture(cost.Mana{cost.W}, cost.Mana{cost.W})
		def.ActivatedAbilities[0].ZoneOfFunction = zone.Graveyard
		if issues := game.ValidateCardDef(def); len(issues) != 1 || issues[0].Code != game.CardDefIssueInvalidAbilityBody {
			t.Fatalf("expected off-battlefield declaration to be rejected, got %+v", issues)
		}
	})
	t.Run("hybrid_source_cost", func(t *testing.T) {
		def := powerUpFixture(cost.Mana{cost.HybridMana(mana.R, mana.G)}, cost.Mana{cost.W})
		if issues := game.ValidateCardDef(def); len(issues) != 1 || issues[0].Code != game.CardDefIssueInvalidAbilityBody {
			t.Fatalf("expected unsupported source symbols to be rejected, got %+v", issues)
		}
	})
}
