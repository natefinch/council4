package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestConditionControllerControlsPermanentFilter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "types.Snow Mountain",
		Supertypes: []types.Super{types.Basic, types.Snow},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Mountain}},
	})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Large Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 7}),
		Toughness: opt.Val(game.PT{Value: 7})},
	})

	condition := opt.Val(game.Condition{
		ControllerControls: game.PermanentFilter{
			Types:      []types.Card{types.Land},
			Supertypes: []types.Super{types.Basic},
			SubtypesAny: []types.Sub{
				types.Swamp,
				types.Mountain,
			},
			MinCount: 1,
		},
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition did not match controlled basic Mountain")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
		t.Fatal("condition matched another player's Mountain")
	}

	powerCondition := opt.Val(game.Condition{
		ControllerControls: game.PermanentFilter{
			Types: []types.Card{types.Creature},
			Power: opt.Val(compare.Int{
				Op:    compare.GreaterOrEqual,
				Value: 7,
			}),
		},
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, powerCondition) {
		t.Fatal("condition did not match controlled creature with power >= 7")
	}
	powerCondition.Val.Negate = true
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, powerCondition) {
		t.Fatal("negated condition matched")
	}
}

func TestConditionControllerControlsPermanentFilterCanExcludeSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Source",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5})},
	})
	condition := opt.Val(game.Condition{
		ControllerControls: game.PermanentFilter{
			Types:         []types.Card{types.Creature},
			ExcludeSource: true,
			Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
		},
	})
	if conditionSatisfied(g, conditionContext{controller: game.Player1, source: source}, condition) {
		t.Fatal("condition matched source as another creature")
	}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Other",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4})},
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1, source: source}, condition) {
		t.Fatal("condition did not match another large creature")
	}
}

func TestConditionControlsMatchingIgnoresPhasedOutPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Relic",
		Types: []types.Card{types.Artifact},
	}})
	condition := opt.Val(game.Condition{
		ControlsMatching: opt.Val(game.SelectionCount{
			Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
		}),
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition did not count in-phase artifact")
	}

	artifact.PhasedOut = true
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition counted phased-out artifact")
	}
}

func TestConditionObjectMatchesSourceLiveState(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Relic",
		Types: []types.Card{types.Artifact},
	}})
	condition := opt.Val(game.Condition{
		Object: opt.Val(game.SourcePermanentReference()),
		ObjectMatches: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Artifact},
			Tapped:        game.TriFalse,
		}),
	})
	ctx := conditionContext{controller: game.Player1, source: source}
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("source object condition did not match live untapped artifact")
	}
	source.Tapped = true
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("source object condition matched tapped artifact")
	}

	exists := opt.Val(game.Condition{Object: opt.Val(game.SourcePermanentReference())})
	if !conditionSatisfied(g, ctx, exists) {
		t.Fatal("source existence condition did not match battlefield source")
	}
	g.Battlefield = nil
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, exists) {
		t.Fatal("source existence condition matched absent source")
	}
}

func TestConditionObjectMatchesEventPermanentLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	human := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Departed Human",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Human},
	}})
	snapshot := snapshotPermanent(g, human, zone.Battlefield)
	rememberLastKnown(g, &snapshot)
	g.Battlefield = nil
	ctx := conditionContext{
		controller: game.Player1,
		event:      &game.Event{Kind: game.EventPermanentDied, PermanentID: human.ObjectID},
	}
	condition := opt.Val(game.Condition{
		Object: opt.Val(game.EventPermanentReference()),
		ObjectMatches: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			SubtypesAny:   []types.Sub{types.Human},
		}),
	})
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("event object condition did not match creature/Human LKI")
	}
	condition.Val.ObjectMatches.Val.SubtypesAny = []types.Sub{types.Elf}
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("event object condition matched wrong LKI subtype")
	}
	legacy := opt.Val(game.Condition{
		Object: opt.Val(game.EventPermanentReference()),
		Types:  []types.Card{types.Creature},
	})
	if !conditionSatisfied(g, ctx, legacy) {
		t.Fatal("legacy Object+Types condition no longer matched LKI")
	}
}

func TestConditionProvenControllerSelections(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for i := range 2 {
		creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:      fmt.Sprintf("Tapped Creature %d", i),
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
		}})
		creature.Tapped = true
	}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Gate",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Gate},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Equipment",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
	}})
	ctx := conditionContext{controller: game.Player1}
	for _, condition := range []game.Condition{
		{ControlsMatching: opt.Val(game.SelectionCount{
			MinCount:  2,
			Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriTrue},
		})},
		{ControlsMatching: opt.Val(game.SelectionCount{
			Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5})},
		})},
		{ControlsMatching: opt.Val(game.SelectionCount{
			Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, SubtypesAny: []types.Sub{types.Gate}},
		})},
		{ControlsMatching: opt.Val(game.SelectionCount{
			Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}, SubtypesAny: []types.Sub{types.Equipment}},
		})},
	} {
		if !conditionSatisfied(g, ctx, opt.Val(condition)) {
			t.Fatalf("controller selection condition did not match: %+v", condition)
		}
	}
	noCreatures := opt.Val(game.Condition{
		Negate: true,
		ControlsMatching: opt.Val(game.SelectionCount{
			MinCount:  1,
			Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		}),
	})
	if conditionSatisfied(g, ctx, noCreatures) {
		t.Fatal("no-creatures condition matched controlled creatures")
	}
}

func TestInterveningHadCountersUsesEventPermanentLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Charged Relic",
		Types: []types.Card{types.Artifact},
	}})
	permanent.Counters.Add(counter.Charge, 1)
	snapshot := snapshotPermanent(g, permanent, zone.Battlefield)
	rememberLastKnown(g, &snapshot)
	g.Battlefield = nil
	trigger := game.TriggerCondition{InterveningIfEventPermanentHadCounters: true}
	event := game.Event{Kind: game.EventZoneChanged, PermanentID: permanent.ObjectID}
	if !triggerInterveningIf(g, nil, game.Player1, &trigger, &event) {
		t.Fatal("had-counters intervening condition did not match LKI counters")
	}
	snapshot.Counters = counter.Set{}
	rememberLastKnown(g, &snapshot)
	if triggerInterveningIf(g, nil, game.Player1, &trigger, &event) {
		t.Fatal("had-counters intervening condition matched empty LKI counters")
	}
}

func TestConditionControllerLiveStatePredicates(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for i, cardTypes := range [][]types.Card{
		{types.Artifact, types.Creature},
		{types.Enchantment},
		{types.Instant},
		{types.Land},
		{types.Creature},
		{types.Creature},
		{types.Creature},
	} {
		addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:  fmt.Sprintf("Graveyard Card %d", i),
			Types: cardTypes,
		}})
	}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Dual Land",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Plains, types.Island},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Forest",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
	}})
	for _, power := range []int{-1, 0, 4} {
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:      fmt.Sprintf("Creature %d", power),
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: power}),
			Toughness: opt.Val(game.PT{Value: 1}),
		}})
	}

	condition := opt.Val(game.Condition{
		ControllerHandEmpty:                     true,
		ControllerGraveyardCardCountAtLeast:     7,
		ControllerGraveyardCardTypeCountAtLeast: 4,
		ControllerBasicLandTypeCountAtLeast:     3,
		ControllerCreaturePowerDiversityAtLeast: 3,
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition did not match controller live state")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
		t.Fatal("condition matched another player's live state")
	}
}

func TestConditionCardCountsIgnoreTransientTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Hand.Add(g.IDGen.Next())
	g.Players[game.Player1].Graveyard.Add(g.IDGen.Next())

	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, opt.Val(game.Condition{
		ControllerHandEmpty: true,
	})) {
		t.Fatal("transient token in hand prevented empty-hand condition")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, opt.Val(game.Condition{
		ControllerHandSizeAtLeast: 1,
	})) {
		t.Fatal("transient token in hand counted toward hand size")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, opt.Val(game.Condition{
		ControllerGraveyardCardCountAtLeast: 1,
	})) {
		t.Fatal("transient token in graveyard counted as a card")
	}
	if got := dynamicAmountValue(g, nil, game.Player1, game.DynamicAmount{
		Kind:       game.DynamicAmountControllerHandSize,
		Multiplier: 1,
	}); got != 0 {
		t.Fatalf("dynamic hand size = %d, want 0", got)
	}
	if got := dynamicAmountValue(g, nil, game.Player1, game.DynamicAmount{
		Kind:       game.DynamicAmountControllerGraveyardSize,
		Multiplier: 1,
	}); got != 0 {
		t.Fatalf("dynamic graveyard size = %d, want 0", got)
	}
}

func TestConditionDeliriumCombinesSplitCardTypesOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToGraveyard(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Split", Types: []types.Card{types.Instant}},
		Layout:   game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{
			Name:  "Other Half",
			Types: []types.Card{types.Sorcery},
		}),
	})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic", Types: []types.Card{types.Artifact}}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Aura", Types: []types.Card{types.Enchantment}}})

	delirium := opt.Val(game.Condition{ControllerGraveyardCardTypeCountAtLeast: 4})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, delirium) {
		t.Fatal("split card did not contribute both card types to Delirium")
	}

	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToGraveyard(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Adventurer", Types: []types.Card{types.Creature}},
		Layout:   game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name:  "Adventure",
			Types: []types.Card{types.Instant},
		}),
	})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic", Types: []types.Card{types.Artifact}}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Aura", Types: []types.Card{types.Enchantment}}})
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, delirium) {
		t.Fatal("Adventure face contributed its card type to Delirium")
	}
}

func TestConditionTargetEnteredThisTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "New Creature",
		Types: []types.Card{types.Creature}},
	})
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(creature.ObjectID)},
	}
	condition := opt.Val(game.Condition{TargetEnteredThisTurn: opt.Val(0)})
	if conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition) {
		t.Fatal("condition matched before enter event")
	}
	emitEvent(g, game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: creature.ObjectID})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition) {
		t.Fatal("condition did not match target that entered this turn")
	}
	g.Turn.TurnNumber++
	g.EventTurnStarts = append(g.EventTurnStarts, len(g.Events))
	if conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition) {
		t.Fatal("condition matched target that entered on a previous turn")
	}
}

func TestConditionCastFromZoneRequiresNonCopyStackObject(t *testing.T) {
	condition := opt.Val(game.Condition{CastFromZone: opt.Val(zone.Graveyard)})
	obj := &game.StackObject{
		Kind:       game.StackSpell,
		Controller: game.Player1,
		SourceZone: zone.Graveyard,
	}
	if !conditionSatisfied(game.NewGame([game.NumPlayers]game.PlayerConfig{}), conditionContext{controller: game.Player1, obj: obj}, condition) {
		t.Fatal("condition did not match spell cast from graveyard")
	}
	obj.Copy = true
	if conditionSatisfied(game.NewGame([game.NumPlayers]game.PlayerConfig{}), conditionContext{controller: game.Player1, obj: obj}, condition) {
		t.Fatal("condition matched copied stack object")
	}
	obj.Copy = false
	obj.SourceZone = zone.Hand
	if conditionSatisfied(game.NewGame([game.NumPlayers]game.PlayerConfig{}), conditionContext{controller: game.Player1, obj: obj}, condition) {
		t.Fatal("condition matched spell cast from hand")
	}
}

func TestConditionControllerControlsTotalPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Small", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 3})}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Medium", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 5})}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Land", Types: []types.Card{types.Land}}})

	condition := opt.Val(game.Condition{
		ControllerControls: game.PermanentFilter{
			Types:      []types.Card{types.Creature},
			TotalPower: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 8}),
		},
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition did not match total creature power >= 8")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
		t.Fatal("condition matched another player's creatures")
	}
}

func TestConditionEventPermanentNameUniqueAmongControlledAndGraveyardCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	unique := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Unique", Types: []types.Card{types.Creature}}})
	condition := opt.Val(game.Condition{EventPermanentNameUniqueAmongControlledAndGraveyardCreatures: true})
	ctx := conditionContext{
		controller: game.Player1,
		event: &game.Event{
			PermanentID: unique.ObjectID,
		},
	}
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("condition counted the event permanent as another matching creature")
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Unique", Types: []types.Card{types.Creature}}})
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("condition matched with another controlled creature of the same name")
	}

	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	unique = addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Unique", Types: []types.Card{types.Creature}}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Unique", Types: []types.Card{types.Creature}}})
	ctx.event = &game.Event{PermanentID: unique.ObjectID}
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("condition matched with a same-name creature card in graveyard")
	}
}

func TestActivationConditionRestrictsExplicitAndAutoMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	verge := addCombatPermanent(g, game.Player1, conditionalRedManaLand())
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Red Spell",
		ManaCost:     opt.Val(cost.Mana{cost.R}),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})

	if containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(verge.ObjectID, 0, nil, 0)) {
		t.Fatal("conditional red mana ability was legal without Swamp or Mountain")
	}
	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("auto-payment treated conditional red mana as available without Swamp or Mountain")
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mountain",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Mountain}},
	})
	if !containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(verge.ObjectID, 0, nil, 0)) {
		t.Fatal("conditional red mana ability was not legal with Mountain")
	}
	if !containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("auto-payment did not use conditional red mana with Mountain")
	}
}

func TestInterveningConditionChecksControllerPermanentPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBugenhagenLikePermanent(g, game.Player1)
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	if _, ok := engine.drawCard(g, game.Player2); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger fired without controlled creature with power >= 7")
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Large Creature",
		Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 7})},
	})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Again"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reward"}})
	if _, ok := engine.drawCard(g, game.Player2); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger did not fire with controlled creature with power >= 7")
	}
	g.Battlefield = g.Battlefield[:1]
	log := TurnLog{}
	engine.resolveTopOfStack(g, &log)
	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want intervening-if recheck to skip draw", got)
	}
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "intervening if false" {
		t.Fatalf("resolve log = %+v, want intervening-if false", log.Resolves)
	}
}

func TestStaticConditionGraveyardAbilityGrantsHaste(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Creature",
		Types: []types.Card{types.Creature}},
	})
	other := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Other Creature",
		Types: []types.Card{types.Creature}},
	})
	angerID := addCardToGraveyard(g, game.Player1, angerLikeCard())
	angerCard := g.CardInstances[angerID]

	if hasKeyword(g, creature, game.Haste) {
		t.Fatal("Anger-like graveyard ability granted haste without Mountain")
	}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mountain",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Mountain}},
	})
	if !hasKeyword(g, creature, game.Haste) {
		t.Fatal("Anger-like graveyard ability did not grant haste with Mountain")
	}
	if hasKeyword(g, other, game.Haste) {
		t.Fatal("Anger-like graveyard ability granted haste to opponent's creature")
	}

	g.Players[game.Player1].Graveyard.Remove(angerID)
	battlefieldAnger := addCombatPermanent(g, game.Player1, angerCard.Def)
	if hasKeyword(g, creature, game.Haste) {
		t.Fatal("graveyard-only Anger-like static ability functioned from battlefield")
	}
	if !hasKeyword(g, battlefieldAnger, game.Haste) {
		t.Fatal("battlefield Anger-like source did not have its own haste keyword")
	}
}

func TestConditionalEntersTappedCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	cardID := addCardToHand(g, game.Player1, cinderLikeLand())
	engine := NewEngine(nil)
	if !engine.applyPlayLand(g, game.Player1, cardID) {
		t.Fatal("play land without basics failed")
	}
	if got := g.Battlefield[len(g.Battlefield)-1]; !got.Tapped {
		t.Fatalf("land = %+v, want tapped without two basic lands", got)
	}

	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Forest}},
	})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Island}},
	})
	cardID = addCardToHand(g, game.Player1, cinderLikeLand())
	if !engine.applyPlayLand(g, game.Player1, cardID) {
		t.Fatal("play land with basics failed")
	}
	if got := g.Battlefield[len(g.Battlefield)-1]; got.Tapped {
		t.Fatalf("land = %+v, want untapped with two basic lands", got)
	}
}

func TestLifeAndOpponentConditions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}

	g.Players[game.Player1].Life = 10
	if !conditionSatisfied(g, ctx, opt.Val(game.Condition{ControllerLifeAtLeast: 10})) {
		t.Fatal("controller life-at-least condition failed at threshold")
	}
	g.Players[game.Player1].Life = 9
	if conditionSatisfied(g, ctx, opt.Val(game.Condition{ControllerLifeAtLeast: 10})) {
		t.Fatal("controller life-at-least condition passed below threshold")
	}

	g.Players[game.Player2].Life = 13
	if !conditionSatisfied(g, ctx, opt.Val(game.Condition{AnyPlayerLifeAtMost: 13})) {
		t.Fatal("any-player life condition failed at threshold")
	}
	g.Players[game.Player1].Life = 14
	g.Players[game.Player2].Life = 14
	if conditionSatisfied(g, ctx, opt.Val(game.Condition{AnyPlayerLifeAtMost: 13})) {
		t.Fatal("any-player life condition passed with all players above threshold")
	}

	g.Players[game.Player3].Eliminated = true
	g.Players[game.Player4].Eliminated = true
	if conditionSatisfied(g, ctx, opt.Val(game.Condition{OpponentCountAtLeast: 2})) {
		t.Fatal("opponent-count condition included eliminated players")
	}
	g.Players[game.Player3].Eliminated = false
	if !conditionSatisfied(g, ctx, opt.Val(game.Condition{OpponentCountAtLeast: 2})) {
		t.Fatal("opponent-count condition failed with two alive opponents")
	}
}

func TestNegativeConditionThresholdsFailClosed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}
	conditions := []game.Condition{
		{Negate: true, ControllerLifeAtLeast: -1},
		{Negate: true, AnyPlayerLifeAtMost: -1},
		{Negate: true, OpponentCountAtLeast: -1},
	}
	for _, condition := range conditions {
		if conditionSatisfied(g, ctx, opt.Val(condition)) {
			t.Fatalf("conditionSatisfied(%+v) = true, want false", condition)
		}
		countConditions := []game.Condition{
			{Negate: true, ControllerControls: game.PermanentFilter{MinCount: -1}},
			{Negate: true, ControlsMatching: opt.Val(game.SelectionCount{MinCount: -1})},
			{Negate: true, AnyOpponentControls: opt.Val(game.SelectionCount{MinCount: -1})},
			{Negate: true, OpponentsControl: opt.Val(game.SelectionCount{MinCount: -1})},
		}
		for _, condition := range countConditions {
			if conditionSatisfied(g, ctx, opt.Val(condition)) {
				t.Fatalf("conditionSatisfied(%+v) = true, want false", condition)
			}
		}
	}
}

func TestOpponentPermanentCountConditions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}
	count := game.SelectionCount{
		Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
		MinCount:  2,
	}
	addBasicLandPermanent(g, game.Player2, types.Forest)
	addBasicLandPermanent(g, game.Player3, types.Island)

	if conditionSatisfied(g, ctx, opt.Val(game.Condition{AnyOpponentControls: opt.Val(count)})) {
		t.Fatal("any-opponent condition combined permanents across opponents")
	}
	if !conditionSatisfied(g, ctx, opt.Val(game.Condition{OpponentsControl: opt.Val(count)})) {
		t.Fatal("collective opponents condition did not combine permanents")
	}
	addBasicLandPermanent(g, game.Player2, types.Mountain)
	if !conditionSatisfied(g, ctx, opt.Val(game.Condition{AnyOpponentControls: opt.Val(count)})) {
		t.Fatal("any-opponent condition failed for one opponent with enough permanents")
	}
}

func TestLifeConditionalEntersTappedReplacement(t *testing.T) {
	for _, test := range []struct {
		name       string
		life       int
		wantTapped bool
	}{
		{name: "below threshold", life: 9, wantTapped: true},
		{name: "at threshold", life: 10, wantTapped: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = test.life
			setSorcerySpeedTurn(g, game.Player1)
			cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name:  "Life Land",
				Types: []types.Card{types.Land},
				ReplacementAbilities: []game.ReplacementAbility{
					game.EntersTappedIfReplacement("This land enters tapped unless you have 10 or more life.", &game.Condition{
						Negate:                true,
						ControllerLifeAtLeast: 10,
					}),
				},
			}})
			if !NewEngine(nil).applyPlayLand(g, game.Player1, cardID) {
				t.Fatal("applyPlayLand() = false")
			}
			if got := g.Battlefield[len(g.Battlefield)-1].Tapped; got != test.wantTapped {
				t.Fatalf("Tapped = %v, want %v", got, test.wantTapped)
			}
		})
	}
}

func setSorcerySpeedTurn(g *game.Game, playerID game.PlayerID) {
	g.Turn.ActivePlayer = playerID
	g.Turn.PriorityPlayer = playerID
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
}

func conditionalRedManaLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Conditional Red Land",
		Types: []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{{
			Text: "{T}: Add {R}. Activate only if you control a Swamp or a Mountain.",
			ActivationCondition: opt.Val(game.Condition{
				ControllerControls: game.PermanentFilter{
					SubtypesAny: []types.Sub{types.Swamp, types.Mountain},
				},
			}),
			AdditionalCosts: cost.Tap,
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.AddMana{
					Amount:    game.Fixed(1),
					ManaColor: mana.R,
				}}},
			}.Ability(),
		}}},
	}
}

func addBugenhagenLikePermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Bugenhagen-like",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "At the beginning of your upkeep, if you control a creature with power 7 or greater, draw a card.",
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event: game.EventCardDrawn,
				},
				InterveningIf: "if you control a creature with power 7 or greater",
				InterveningCondition: opt.Val(game.Condition{
					ControllerControls: game.PermanentFilter{
						Types: []types.Card{types.Creature},
						Power: opt.Val(compare.Int{
							Op:    compare.GreaterOrEqual,
							Value: 7,
						}),
					},
				}),
			},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}}},
	})
}

func angerLikeCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Anger-like",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{
			game.HasteStaticBody,
			{
				ZoneOfFunction: zone.Graveyard,
				Condition: opt.Val(game.Condition{
					ControllerControls: game.PermanentFilter{
						SubtypesAny: []types.Sub{types.Mountain},
					},
				}),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer: game.LayerAbility,
					Group: game.BattlefieldGroup(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Controller:    game.ControllerYou,
					}),
					AddKeywords: []game.Keyword{game.Haste},
				}},
			},
		}},
	}
}

func cinderLikeLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Cinder-like",
		Types: []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedIfReplacement("This land enters tapped unless you control two or more basic lands.", &game.Condition{
				Negate: true,
				ControllerControls: game.PermanentFilter{
					Types:      []types.Card{types.Land},
					Supertypes: []types.Super{types.Basic},
					MinCount:   2,
				},
			}),
		}},
	}
}

func TestEventHistoryConditionCurrentTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source1 := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})
	source2 := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Bear",
		Types: []types.Card{types.Creature},
	}})

	attackedCond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:      game.EventAttackerDeclared,
				Controller: game.TriggerControllerYou,
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})

	ctx1 := conditionContext{controller: game.Player1, source: source1}
	ctx2 := conditionContext{controller: game.Player2, source: source2}
	if conditionSatisfied(g, ctx1, attackedCond) {
		t.Fatal("condition satisfied before any attacks")
	}

	// Player1 attacks — satisfied for Player1's source but not Player2's.
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})

	if !conditionSatisfied(g, ctx1, attackedCond) {
		t.Fatal("condition not satisfied after Player1 attacked")
	}
	if conditionSatisfied(g, ctx2, attackedCond) {
		t.Fatal("condition satisfied for Player2 source when only Player1 attacked")
	}
}

func TestEventHistoryConditionPreviousTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})

	lifeLostCond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:  game.EventLifeLost,
				Player: game.TriggerPlayerOpponent,
			},
			Window: game.EventHistoryPreviousTurn,
		}),
	})

	ctx := conditionContext{controller: game.Player1, source: source}
	if conditionSatisfied(g, ctx, lifeLostCond) {
		t.Fatal("condition satisfied before any events")
	}

	// Emit life-loss on this turn then advance to the next turn so it becomes
	// "previous turn" for the upkeep check.
	emitEvent(g, game.Event{Kind: game.EventLifeLost, Player: game.Player2, Amount: 3})
	g.EventTurnStarts = append(g.EventTurnStarts, len(g.Events))
	g.Turn.TurnNumber++

	if !conditionSatisfied(g, ctx, lifeLostCond) {
		t.Fatal("condition not satisfied after opponent lost life last turn")
	}
}

func TestEventHistoryConditionNegatedNoSpells(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})

	noSpellsCond := opt.Val(game.Condition{
		Negate: true,
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{Event: game.EventSpellCast},
			Window:  game.EventHistoryPreviousTurn,
		}),
	})

	ctx := conditionContext{controller: game.Player1, source: source}

	// No previous turn yet — EventsPreviousTurn returns nil, no spells found →
	// condition (negated) is satisfied.
	if !conditionSatisfied(g, ctx, noSpellsCond) {
		t.Fatal("negated condition not satisfied when no previous-turn events exist")
	}

	// Emit a spell cast on "last turn" then advance.
	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	g.EventTurnStarts = append(g.EventTurnStarts, len(g.Events))
	g.Turn.TurnNumber++

	if conditionSatisfied(g, ctx, noSpellsCond) {
		t.Fatal("negated condition satisfied when a spell was cast last turn")
	}
}

func TestEventHistoryConditionCreatureDiedCurrentTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})

	creatureDiedCond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:            game.EventPermanentDied,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})

	ctx := conditionContext{controller: game.Player1, source: source}
	if conditionSatisfied(g, ctx, creatureDiedCond) {
		t.Fatal("condition satisfied before any creature deaths")
	}

	// A non-creature permanent died — should not satisfy.
	addAndEmitArtifactDied(g)
	if conditionSatisfied(g, ctx, creatureDiedCond) {
		t.Fatal("condition satisfied after non-creature died")
	}

	// A creature dies — now satisfied.
	emitCreatureDiedEvent(g)
	if !conditionSatisfied(g, ctx, creatureDiedCond) {
		t.Fatal("condition not satisfied after creature died")
	}
}

func TestEventHistoryConditionFailsClosedWithNilSource(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})

	cond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:      game.EventAttackerDeclared,
				Controller: game.TriggerControllerYou,
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})
	if conditionSatisfied(g, conditionContext{controller: game.Player1, source: nil}, cond) {
		t.Fatal("condition satisfied with nil source; should fail closed")
	}
}

// addAndEmitArtifactDied registers and emits an EventPermanentDied for a
// non-creature artifact so tests can verify creature-type filtering.
func addAndEmitArtifactDied(g *game.Game) {
	perm := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Dead Relic",
		Types: []types.Card{types.Artifact},
	}})
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentDied,
		PermanentID: perm.ObjectID,
		CardTypes:   []types.Card{types.Artifact},
	})
}

// emitCreatureDiedEvent emits an EventPermanentDied that looks like a creature
// died by recording creature card types on the event directly.
func emitCreatureDiedEvent(g *game.Game) {
	perm := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Dead Bear",
		Types: []types.Card{types.Creature},
	}})
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentDied,
		PermanentID: perm.ObjectID,
		CardTypes:   []types.Card{types.Creature},
	})
}
