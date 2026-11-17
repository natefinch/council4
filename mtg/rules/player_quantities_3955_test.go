package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3955 (sub-issue of #2098). Official CR effective
// 2026-11-06 (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt),
// read 202.3, 406.3a, 608.2h, 611.2c, 613.4b, 701.54a-e, 702.124a, 708.2, 903.3 and
// 903.3a-e. 903.3/903.3e: a commander is a commander in every zone and has its
// current characteristics there; 608.2h: an effect's amount is determined once,
// as it resolves, so a P/T-setting "X/X" is locked (611.2c). 701.54: each time
// the Ring tempts you counts, even if you control no creature. Cards (pinned
// corpus; the fixtures below follow this Oracle text and are not registered):
// Cactus Preserve (oracle 8da29533-f389-4bc2-ab9b-b469f893a362, Land — Desert)
// "{3}: Until end of turn, this land becomes an X/X green Plant creature with
// reach, where X is the greatest mana value among your commanders. It's still
// a land." (rulings https://api.scryfall.com/cards/ad9d426f-5870-42bb-a589-9218f7e35d62/rulings,
// 2024-04-12: it doesn't matter where your commander is, or commanders are with
// partner; X uses the commander's current characteristics, which may differ if
// it became a copy of something else). Frodo, Sauron's Bane (oracle
// 8be62f5c-0ab2-4869-97a2-d8a69a3fde95) "{B}{B}{B}: If Frodo is a Scout, it
// becomes a Halfling Rogue with "Whenever this creature deals combat damage to
// a player, that player loses the game if the Ring has tempted you four or more
// times this game. Otherwise, the Ring tempts you."" (rulings
// https://api.scryfall.com/cards/7d86dc2e-6f0c-4714-9d30-5d099d3b895c/rulings,
// 2023-06-16: the Ring can tempt you even if you control no creature). The
// Rogue fixture carries only the granted trigger; Frodo's type-checking
// animation is separate #2098 work.

func cactusPreserveFixture() *game.CardDef {
	x := game.DynamicAmount{Kind: game.DynamicAmountGreatestCommanderManaValue, Multiplier: 1}
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Cactus Preserve",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Desert},
		ActivatedAbilities: []game.ActivatedAbility{animateSelfAbility(
			"{3}: Until end of turn, this land becomes an X/X green Plant creature with reach, where X is the greatest mana value among your commanders. It's still a land.",
			cost.Mana{cost.O(3)},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerType,
				AddTypes: []types.Card{types.Creature}, AddSubtypes: []types.Sub{types.Plant},
			},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerColor,
				SetColors: []color.Color{color.Green},
			},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerAbility,
				AddKeywords: []game.Keyword{game.Reach},
			},
			game.ContinuousEffect{
				Duration: game.DurationPermanent, Layer: game.LayerPowerToughnessSet,
				SetPowerDynamic: opt.Val(x), SetToughnessDynamic: opt.Val(x),
			},
		)},
	}}
}

func testCommander(name string, manaCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		ManaCost:   opt.Val(manaCost),
		Types:      []types.Card{types.Creature},
		Supertypes: []types.Super{types.Legendary},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 3}),
	}}
}

func commanderQuantityGame(configs [game.NumPlayers]game.PlayerConfig) (*game.Game, *Engine) {
	g := game.NewGame(configs)
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	return g, NewEngine(nil)
}

func activateCactusPreserve(t *testing.T, g *game.Game, engine *Engine, cactus *game.Permanent) {
	t.Helper()
	activateAndResolve(t, g, engine, game.Player1, cactus, mana.C, mana.C, mana.C)
	if !permanentHasType(g, cactus, types.Creature) || !permanentHasType(g, cactus, types.Land) ||
		!permanentHasSubtype(g, cactus, types.Plant) || !permanentHasSubtype(g, cactus, types.Desert) {
		t.Fatal("Cactus Preserve did not become a Plant land creature")
	}
	if colors := permanentEffectiveColors(g, cactus); len(colors) != 1 || colors[0] != color.Green {
		t.Fatalf("Cactus Preserve colors = %v, want green", colors)
	}
	if !hasKeyword(g, cactus, game.Reach) {
		t.Fatal("Cactus Preserve lacks reach")
	}
}

// TestCactusPreserveUsesGreatestOfYourPartnerCommanders is the #3955 partner
// scenario: with partner commanders of mana value 3 and 5 in the command zone,
// Cactus Preserve becomes a 5/5; an opponent's mana value 8 commander is not
// one of yours.
func TestCactusPreserveUsesGreatestOfYourPartnerCommanders(t *testing.T) {
	var configs [game.NumPlayers]game.PlayerConfig
	configs[game.Player1].Commanders = []*game.CardDef{
		testCommander("Partner Three", cost.Mana{cost.O(2), cost.G}),
		testCommander("Partner Five", cost.Mana{cost.O(3), cost.G, cost.G}),
	}
	configs[game.Player2].Commander = testCommander("Opposing Eight", cost.Mana{cost.O(7), cost.B})
	g, engine := commanderQuantityGame(configs)
	cactus := addCombatPermanent(g, game.Player1, cactusPreserveFixture())

	activateCactusPreserve(t, g, engine, cactus)
	assertPowerToughness(t, g, cactus, 5, 5)

	clone := g.Clone()
	x := game.DynamicAmount{Kind: game.DynamicAmountGreatestCommanderManaValue, Multiplier: 1}
	for player, want := range map[game.PlayerID]int{game.Player1: 5, game.Player2: 8, game.Player3: 0} {
		if got, cloned := dynamicAmountValue(g, nil, player, x), dynamicAmountValue(clone, nil, player, x); got != want || cloned != want {
			t.Fatalf("player %d greatest commander mana value = %d (clone %d), want %d", player, got, cloned, want)
		}
	}
	clonedCactus, ok := findPermanentByCardID(clone, cactus.CardInstanceID)
	if !ok {
		t.Fatal("cloned game lacks Cactus Preserve")
	}
	assertPowerToughness(t, clone, clonedCactus, 5, 5)
}

// TestCactusPreserveXLocksAndReadsCommanderInEveryZone is the #3955 single
// commander scenario. X reads the commander in the command zone (mana value
// 2), on the stack as a spell cast with X=5 (mana value 7, CR 202.3e) and on
// the battlefield (mana value 2 again). Each activation's X is fixed when it
// resolves (CR 608.2h), so moving the commander doesn't change an earlier
// result; the newest activation's 7b effect wins (CR 613.7).
func TestCactusPreserveXLocksAndReadsCommanderInEveryZone(t *testing.T) {
	var configs [game.NumPlayers]game.PlayerConfig
	configs[game.Player1].Commander = testCommander("Hydra Commander", cost.Mana{cost.X, cost.G, cost.G})
	g, engine := commanderQuantityGame(configs)
	commanderID := g.Players[game.Player1].CommanderInstanceID
	cactus := addCombatPermanent(g, game.Player1, cactusPreserveFixture())

	activateCactusPreserve(t, g, engine, cactus)
	assertPowerToughness(t, g, cactus, 2, 2)

	g.Players[game.Player1].ManaPool.Add(mana.C, 5)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	if !engine.applyAction(g, game.Player1, action.CastCommanderSpell(commanderID, nil, 5, nil)) {
		t.Fatal("casting the commander from the command zone failed")
	}
	if g.Players[game.Player1].CommandZone.Contains(commanderID) {
		t.Fatal("the commander is still in the command zone")
	}
	assertPowerToughness(t, g, cactus, 2, 2)

	activateCactusPreserve(t, g, engine, cactus)
	assertPowerToughness(t, g, cactus, 7, 7)

	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := findPermanentByCardID(g, commanderID); !ok {
		t.Fatal("the commander did not resolve onto the battlefield")
	}
	assertPowerToughness(t, g, cactus, 7, 7)

	activateCactusPreserve(t, g, engine, cactus)
	assertPowerToughness(t, g, cactus, 2, 2)

	engine.runCleanupStep(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if permanentHasType(g, cactus, types.Creature) {
		t.Fatal("Cactus Preserve is still a creature after cleanup")
	}
}

// TestCactusPreserveIgnoresFaceDownExiledCommanderManaValue is the #3955
// hidden-information scenario. A commander exiled face down has no
// characteristics (CR 406.3a), so its mana value is 0 and its printed mana
// value doesn't leak: with partners of mana value 3 and 5 and the 5 exiled face
// down, X is 3; with only the 5 commander, exiled face down, X is 0. Once the
// card is face up in exile its printed mana value counts again.
func TestCactusPreserveIgnoresFaceDownExiledCommanderManaValue(t *testing.T) {
	exileFaceDown := func(t *testing.T, g *game.Game, commanderID id.ID) {
		t.Helper()
		player := g.Players[game.Player1]
		if !player.CommandZone.Remove(commanderID) {
			t.Fatal("the commander is not in the command zone")
		}
		player.Exile.Add(commanderID)
		player.Exile.SetFaceDown(commanderID, true)
	}
	five := testCommander("Partner Five", cost.Mana{cost.O(3), cost.G, cost.G})

	var partners [game.NumPlayers]game.PlayerConfig
	partners[game.Player1].Commanders = []*game.CardDef{testCommander("Partner Three", cost.Mana{cost.O(2), cost.G}), five}
	g, engine := commanderQuantityGame(partners)
	var fiveID id.ID
	for _, cardID := range g.Players[game.Player1].CommandZone.All() {
		if card, ok := g.GetCardInstance(cardID); ok && card.Def.Name == five.Name {
			fiveID = cardID
		}
	}
	exileFaceDown(t, g, fiveID)
	cactus := addCombatPermanent(g, game.Player1, cactusPreserveFixture())
	activateCactusPreserve(t, g, engine, cactus)
	assertPowerToughness(t, g, cactus, 3, 3)

	var single [game.NumPlayers]game.PlayerConfig
	single[game.Player1].Commander = five
	g, engine = commanderQuantityGame(single)
	commanderID := g.Players[game.Player1].CommanderInstanceID
	exileFaceDown(t, g, commanderID)
	cactus = addCombatPermanent(g, game.Player1, cactusPreserveFixture())
	activateCactusPreserve(t, g, engine, cactus)
	assertPowerToughness(t, g, cactus, 0, 0)

	g.Players[game.Player1].Exile.SetFaceDown(commanderID, false)
	activateCactusPreserve(t, g, engine, cactus)
	assertPowerToughness(t, g, cactus, 5, 5)
}

// TestCactusPreserveWithoutCommanderIsZeroZeroAndDies is the #3955 no
// commander scenario: X is 0, so the 0/0 land creature is put into its owner's
// graveyard as a state-based action (CR 704.5f).
func TestCactusPreserveWithoutCommanderIsZeroZeroAndDies(t *testing.T) {
	g, engine := commanderQuantityGame([game.NumPlayers]game.PlayerConfig{})
	cactus := addCombatPermanent(g, game.Player1, cactusPreserveFixture())

	activateCactusPreserve(t, g, engine, cactus)
	assertPowerToughness(t, g, cactus, 0, 0)
	engine.applyStateBasedActions(g)
	if _, ok := findPermanentByCardID(g, cactus.CardInstanceID); ok {
		t.Fatal("the 0/0 Cactus Preserve stayed on the battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cactus.CardInstanceID) {
		t.Fatal("Cactus Preserve is not in its owner's graveyard")
	}
}

func frodoRogueFixture() *game.CardDef {
	temptedFourTimes := game.Condition{Aggregates: []game.AggregateComparison{{
		Aggregate: game.AggregateControllerRingTemptedCount, Op: compare.GreaterOrEqual, Value: 4,
	}}}
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Frodo, Sauron's Bane",
		Types:      []types.Card{types.Creature},
		Supertypes: []types.Super{types.Legendary},
		Subtypes:   []types.Sub{types.Halfling, types.Rogue},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 3}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "Whenever this creature deals combat damage to a player, that player loses the game if the Ring has tempted you four or more times this game. Otherwise, the Ring tempts you.",
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:               game.EventDamageDealt,
					Source:              game.TriggerSourceSelf,
					Subject:             game.TriggerSubjectDamageSource,
					DamageRecipient:     game.DamageRecipientPlayer,
					RequireCombatDamage: true,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{
				{
					Primitive:        game.PlayerLosesGame{Player: game.EventPlayerReference()},
					Condition:        opt.Val(game.EffectCondition{Condition: opt.Val(temptedFourTimes)}),
					PublishCondition: "tempted-four",
				},
				{
					Primitive:           game.RingTempts{Player: game.ControllerReference()},
					ConditionGate:       "tempted-four",
					ConditionGateNegate: true,
				},
			}}.Ability(),
		}},
	}}
}

// frodoCombatAfterTemptations has the Ring tempt Player1 temptations times
// while they control no creature (2023-06-16 ruling), then Frodo deals unblocked combat damage to Player2 and its trigger resolves.
func frodoCombatAfterTemptations(t *testing.T, temptations int) *game.Game {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	tempter := ringTemptsObj(g, game.Player1)
	for range temptations {
		resolveInstruction(engine, g, tempter, game.RingTempts{Player: game.ControllerReference()}, &TurnLog{})
	}
	if got := g.Players[game.Player1].RingTemptedCount; got != temptations {
		t.Fatalf("fixture: Ring tempted Player1 %d times, want %d", got, temptations)
	}
	// Frodo enters after the temptations, so it isn't the Ring-bearer and the
	// Ring's own combat-damage ability doesn't add life loss to the scenario.
	frodo := addCombatPermanent(g, game.Player1, frodoRogueFixture())

	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: frodo.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	engine.resolveCombatDamagePass(g, normalCombatDamage, &TurnLog{})
	if g.Players[game.Player2].Life != g.Players[game.Player2].StartingLife-2 {
		t.Fatalf("Player2 life = %d, want %d", g.Players[game.Player2].Life, g.Players[game.Player2].StartingLife-2)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Frodo's combat damage trigger did not go on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyStateBasedActions(g)
	return g
}

// TestFrodoRogueDefenderLosesAfterFourTemptations is the #3955 Frodo
// scenario with four prior temptations: the damaged player loses the game and
// the Ring doesn't tempt again.
func TestFrodoRogueDefenderLosesAfterFourTemptations(t *testing.T) {
	g := frodoCombatAfterTemptations(t, 4)
	if !g.Players[game.Player2].Eliminated {
		t.Fatal("Player2 did not lose the game")
	}
	if got := g.Players[game.Player1].RingTemptedCount; got != 4 {
		t.Fatalf("RingTemptedCount = %d, want 4 (the Ring tempted again)", got)
	}
	clone := g.Clone()
	ctx := conditionContext{controller: game.Player1}
	got, ok := aggregateValue(g, ctx, game.AggregateControllerRingTemptedCount)
	cloned, clonedOK := aggregateValue(clone, ctx, game.AggregateControllerRingTemptedCount)
	if !ok || !clonedOK || got != 4 || cloned != 4 {
		t.Fatalf("Ring-tempt aggregate = %d/%v (clone %d/%v), want 4", got, ok, cloned, clonedOK)
	}
}

// TestFrodoRogueTemptsInsteadWithThreeTemptations is the #3955 Frodo scenario
// with three prior temptations: the damaged player stays in the game and the
// Ring tempts Frodo's controller a fourth time.
func TestFrodoRogueTemptsInsteadWithThreeTemptations(t *testing.T) {
	g := frodoCombatAfterTemptations(t, 3)
	if g.Players[game.Player2].Eliminated || g.MarkedToLoseGame[game.Player2] {
		t.Fatal("Player2 lost the game with only three temptations")
	}
	if got := g.Players[game.Player1].RingTemptedCount; got != 4 {
		t.Fatalf("RingTemptedCount = %d, want 4", got)
	}
	if got := g.Players[game.Player1].RingLevel; got != 4 {
		t.Fatalf("RingLevel = %d, want 4", got)
	}
}
