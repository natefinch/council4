package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
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
