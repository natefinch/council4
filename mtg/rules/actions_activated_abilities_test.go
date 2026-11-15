package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestEquipAbilityUsesStackAndAttachesOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	equipment := addCombatPermanent(g, game.Player1, equipEquipment())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Runeclaw Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.ActivateAbility(equipment.ObjectID, 0, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("equip activation was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(equip) = false, want true")
	}
	if !forest.Tapped {
		t.Fatal("forest was not tapped to pay equip cost")
	}
	if equipment.AttachedTo.Exists {
		t.Fatal("equipment attached before equip ability resolved")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, nil)
	if !equipment.AttachedTo.Exists || equipment.AttachedTo.Val != creature.ObjectID {
		t.Fatalf("equipment attached to = %v, want %v", equipment.AttachedTo, creature.ObjectID)
	}
	if !permanentIDsContain(creature.Attachments, equipment.ObjectID) {
		t.Fatal("equipped creature does not reference equipment")
	}
}

func TestEquipAbilityOnlyAsSorceryToCreatureYouControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	equipment := addCombatPermanent(g, game.Player1, equipEquipment())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Runeclaw Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	opponentCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Silvercoat Lion",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep

	if containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(equipment.ObjectID, 0, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0)) {
		t.Fatal("equip activation was legal outside sorcery speed")
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(equipment.ObjectID, 0, []game.Target{game.PermanentTarget(opponentCreature.ObjectID)}, 0)) {
		t.Fatal("equip activation was legal targeting opponent's creature")
	}
}

func TestGeneralActivatedAbilityUsesStackAndResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		ManaCost: greenCost(),
		Content: game.Mode{
			Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"}},
			Sequence: []game.Instruction{{Primitive: game.Damage{Recipient: game.AnyTargetDamageRecipient(0), Amount: game.Fixed(2)}}},
		}.Ability(),
	}))
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep
	act := action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PlayerTarget(game.Player2)}, 0)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("instant-speed activated ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(activated ability) = false, want true")
	}
	if !forest.Tapped {
		t.Fatal("forest was not tapped to pay activation cost")
	}
	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("player 2 life before resolution = %d, want 40", got)
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, nil)
	if got := g.Players[game.Player2].Life; got != 38 {
		t.Fatalf("player 2 life after resolution = %d, want 38", got)
	}
}

func TestModalActivatedAbilityEnumeratesPaysAndResolvesChosenMode(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		ManaCost: greenCost(),
		Content: game.AbilityContent{
			MinModes: 1,
			MaxModes: 1,
			Modes: []game.Mode{
				{
					Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"}},
					Sequence: []game.Instruction{{Primitive: game.Damage{Recipient: game.AnyTargetDamageRecipient(0), Amount: game.Fixed(2)}}},
				},
				{
					Sequence: []game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)}}},
				},
			},
		},
	}))
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1
	act := action.ActivateAbilityWithModes(source.ObjectID, 0, nil, 0, []int{1})
	targetedAct := action.ActivateAbilityWithModes(source.ObjectID, 0, []game.Target{game.PlayerTarget(game.Player2)}, 0, []int{0})

	legal := engine.legalActions(g, game.Player1)
	if !containsAction(legal, act) || !containsAction(legal, targetedAct) {
		t.Fatal("modal activated ability choices were not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(modal activated ability) = false, want true")
	}
	if !forest.Tapped {
		t.Fatal("forest was not tapped to pay modal activation cost")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !slices.Equal(obj.ChosenModes, []int{1}) {
		t.Fatalf("stack chosen modes = %#v, want [1]", obj)
	}
	engine.resolveTopOfStack(g, nil)
	if got := g.Players[game.Player1].Life; got != 43 {
		t.Fatalf("controller life = %d, want 43", got)
	}
	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("unchosen damage mode changed player 2 life to %d", got)
	}
}

func TestModalActivatedAbilityPreservesOptionalTargetOwnership(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		Content: game.AbilityContent{
			MinModes: 2,
			MaxModes: 2,
			Modes: []game.Mode{
				{
					Targets:  []game.TargetSpec{{MinTargets: 0, MaxTargets: 1, Constraint: "creature"}},
					Sequence: []game.Instruction{{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}}},
				},
				{
					Targets:  []game.TargetSpec{{MinTargets: 0, MaxTargets: 1, Constraint: "creature"}},
					Sequence: []game.Instruction{{Primitive: game.Untap{Object: game.TargetPermanentReference(0)}}},
				},
			},
		},
	}))
	target := addCreaturePermanent(g, game.Player2)
	g.Turn.PriorityPlayer = game.Player1
	targets := []game.Target{game.PermanentTarget(target.ObjectID)}
	modes := []int{0, 1}
	ambiguous := action.ActivateAbilityWithModes(source.ObjectID, 0, targets, 0, modes)
	if engine.applyAction(g, game.Player1, ambiguous) {
		t.Fatal("activation without ambiguous target ownership was accepted")
	}
	act := action.ActivateAbilityWithModesAndTargetCounts(source.ObjectID, 0, targets, []int{1, 0}, 0, modes)
	legal := engine.legalActions(g, game.Player1)
	if containsAction(legal, ambiguous) {
		t.Fatal("action equality ignored modal target ownership")
	}
	if !containsAction(legal, act) {
		t.Fatal("modal activation with explicit target ownership was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(modal activation with target ownership) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !slices.Equal(obj.TargetCounts, []int{1, 0}) {
		t.Fatalf("stack target counts = %#v, want [1 0]", obj)
	}
	engine.resolveTopOfStack(g, nil)
	if !target.Tapped {
		t.Fatal("target assigned to tap mode was not tapped")
	}
}

func TestGeneralActivatedAbilityTapCostRespectsSummoningSickness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: cost.Tap,
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability(),
	}))
	source.SummoningSick = true
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("tap activated ability was legal while source creature was summoning sick")
	}
	source.SummoningSick = false
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("tap activated ability was not legal after summoning sickness ended")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(tap activated ability) = false, want true")
	}
	if !source.Tapped {
		t.Fatal("source was not tapped to pay activation cost")
	}
	engine.resolveTopOfStack(g, nil)
	if got := g.Players[game.Player1].Life; got != 41 {
		t.Fatalf("player 1 life = %d, want 41", got)
	}
}

func TestActivatedAbilityExilesSourceAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalExileSource,
			Text:   "Exile this artifact",
			Amount: 1,
			Source: zone.Battlefield,
		}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("source-exile activated ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(source-exile ability) = false, want true")
	}
	if _, ok := permanentByObjectID(g, source.ObjectID); ok {
		t.Fatal("source remained on battlefield after paying exile cost")
	}
	if !g.Players[game.Player1].Exile.Contains(source.CardInstanceID) {
		t.Fatal("source was not put into exile")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, nil)
	if got := g.Players[game.Player1].Life; got != 41 {
		t.Fatalf("player 1 life = %d, want 41", got)
	}
}

func TestActivatedAbilityExilesMatchingGraveyardCardAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:          cost.AdditionalExile,
			Text:          "Exile a creature card from your graveyard",
			Amount:        1,
			MatchCardType: true,
			CardType:      types.Creature,
			Source:        zone.Graveyard,
		}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard-exile ability was legal without a matching card")
	}
	instantID := addCardToHand(g, game.Player1, greenInstant())
	g.Players[game.Player1].Hand.Remove(instantID)
	g.Players[game.Player1].Graveyard.Add(instantID)
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard-exile ability was legal with only a nonmatching card")
	}
	creatureID := addCardToHand(g, game.Player1, greenCreature())
	g.Players[game.Player1].Hand.Remove(creatureID)
	g.Players[game.Player1].Graveyard.Add(creatureID)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard-exile ability was not legal with a matching card")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(graveyard-exile ability) = false, want true")
	}
	if g.Players[game.Player1].Graveyard.Contains(creatureID) ||
		!g.Players[game.Player1].Exile.Contains(creatureID) {
		t.Fatal("matching creature card was not moved from graveyard to exile")
	}
	if !g.Players[game.Player1].Graveyard.Contains(instantID) {
		t.Fatal("nonmatching instant card left the graveyard")
	}
}

func TestActivatedAbilityUntapsSourceAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalUntap, Text: "{Q}"}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("untap-cost ability was legal while source was untapped")
	}
	source.Tapped = true
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("untap-cost ability was not legal while source was tapped")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(untap-cost ability) = false, want true")
	}
	if source.Tapped {
		t.Fatal("source remained tapped after paying untap cost")
	}
}

func TestActivatedAbilityUntapCostRespectsSummoningSickness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalUntap, Text: "{Q}"}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	source.Tapped = true
	source.SummoningSick = true
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("untap activated ability was legal while source creature was summoning sick")
	}
	source.SummoningSick = false
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("untap activated ability was not legal after summoning sickness ended")
	}
}

func TestActivatedAbilityRemovesSourceCounterAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:        cost.AdditionalRemoveCounter,
			Text:        "Remove a charge counter from this creature",
			Amount:      1,
			CounterKind: counter.Charge,
		}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("counter-removal ability was legal without a counter")
	}
	source.Counters.Add(counter.Charge, 2)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("counter-removal ability was not legal with a counter")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(counter-removal ability) = false, want true")
	}
	if got := source.Counters.Get(counter.Charge); got != 1 {
		t.Fatalf("charge counters = %d, want 1 after payment", got)
	}
}

func TestActivatedAbilityRemovesMultipleSourceCountersAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:        cost.AdditionalRemoveCounter,
			Text:        "Remove two charge counters from this creature",
			Amount:      2,
			CounterKind: counter.Charge,
		}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	source.Counters.Add(counter.Charge, 1)
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("counter-removal ability was legal with too few counters")
	}
	source.Counters.Add(counter.Charge, 2)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(counter-removal ability) = false, want true")
	}
	if got := source.Counters.Get(counter.Charge); got != 1 {
		t.Fatalf("charge counters = %d, want 1 after payment", got)
	}
}

func TestActivatedAbilityRemovesAnyKindSourceCounterAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:           cost.AdditionalRemoveCounter,
			Text:           "Remove a counter from this creature",
			Amount:         1,
			AnyCounterKind: true,
		}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("any-kind counter-removal ability was legal without a counter")
	}
	source.Counters.Add(counter.MinusOneMinusOne, 2)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("any-kind counter-removal ability was not legal with a counter")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(any-kind counter-removal ability) = false, want true")
	}
	if got := source.Counters.Get(counter.MinusOneMinusOne); got != 1 {
		t.Fatalf("-1/-1 counters = %d, want 1 after payment", got)
	}
}

func TestActivatedAbilityPaysEnergyCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalEnergy,
			Text:   "Pay {E}{E}",
			Amount: 2,
		}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	g.Players[game.Player1].EnergyCounters = 1
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("energy-cost ability was legal with too little energy")
	}
	g.Players[game.Player1].EnergyCounters = 3
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("energy-cost ability was not legal with enough energy")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(energy-cost ability) = false, want true")
	}
	if got := g.Players[game.Player1].EnergyCounters; got != 1 {
		t.Fatalf("energy counters = %d, want 1 after payment", got)
	}
}

func TestActivatedAbilityExertsSourceAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalExert,
			Text:   "Exert this creature",
			Amount: 1,
		}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("exert ability was not legal for untapped source")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(exert ability) = false, want true")
	}
	if source.Tapped || !source.Exerted {
		t.Fatalf("source tapped/exerted = %v/%v, want false/true", source.Tapped, source.Exerted)
	}
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("exert ability was not legal while source was already exerted")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(second exert ability) = false, want true")
	}

	source.SummoningSick = true
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if source.Tapped || source.Exerted || source.SummoningSick {
		t.Fatalf("after exerted untap step tapped/exerted/sick = %v/%v/%v, want false/false/false", source.Tapped, source.Exerted, source.SummoningSick)
	}
}

func TestActivatedAbilityTapAndExertSourceAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{
			cost.T,
			{
				Kind:   cost.AdditionalExert,
				Text:   "Exert this creature",
				Amount: 1,
			},
		},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("tap-and-exert ability was not legal for untapped source")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(tap-and-exert ability) = false, want true")
	}
	if !source.Tapped || !source.Exerted {
		t.Fatalf("source tapped/exerted = %v/%v, want true/true", source.Tapped, source.Exerted)
	}
	source.Tapped = false
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("tap-and-exert ability was not legal after source was untapped while already exerted")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(second tap-and-exert ability) = false, want true")
	}
	if !source.Tapped || !source.Exerted {
		t.Fatalf("after second activation tapped/exerted = %v/%v, want true/true", source.Tapped, source.Exerted)
	}

	source.SummoningSick = true
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !source.Tapped || source.Exerted || source.SummoningSick {
		t.Fatalf("after exerted untap step tapped/exerted/sick = %v/%v/%v, want true/false/false", source.Tapped, source.Exerted, source.SummoningSick)
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Next Draw Step Card"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if source.Tapped {
		t.Fatal("source did not untap on the next untap step after exert cleared")
	}
}

func TestExertedPhasedOutPermanentClearsExertionDuringUntap(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Exerted Phased Creature",
		Types: []types.Card{types.Creature},
	}})
	permanent.Tapped = true
	permanent.Exerted = true
	permanent.PhasedOut = true
	permanent.SummoningSick = true
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if permanent.PhasedOut {
		t.Fatal("permanent did not phase in")
	}
	if !permanent.Tapped || permanent.Exerted || permanent.SummoningSick {
		t.Fatalf("after phased exerted untap tapped/exerted/sick = %v/%v/%v, want true/false/false", permanent.Tapped, permanent.Exerted, permanent.SummoningSick)
	}
}

func TestActivatedAbilityMillsCardsAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalMill,
			Text:   "Mill four cards",
			Amount: 4,
		}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	first := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	second := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("mill-cost ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(mill-cost ability) = false, want true")
	}
	if g.Players[game.Player1].Library.Size() != 0 {
		t.Fatalf("library size = %d, want 0", g.Players[game.Player1].Library.Size())
	}
	if !g.Players[game.Player1].Graveyard.Contains(first) ||
		!g.Players[game.Player1].Graveyard.Contains(second) {
		t.Fatal("milled cards did not move to graveyard")
	}
}

func TestActivatedAbilityPutsCounterOnSourceAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:        cost.AdditionalPutCounter,
			Text:        "Put a verse counter on this creature",
			Amount:      1,
			CounterKind: counter.Verse,
		}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("put-counter-cost ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(put-counter-cost ability) = false, want true")
	}
	if got := source.Counters.Get(counter.Verse); got != 1 {
		t.Fatalf("verse counters = %d, want 1", got)
	}
}
