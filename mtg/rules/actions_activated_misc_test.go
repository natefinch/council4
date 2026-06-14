package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestActivatedAbilityReturnsPermanentsToOwnersHandAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:        cost.AdditionalReturnToHand,
			Text:        "Return two Islands you control to their owner's hand",
			Amount:      2,
			SubtypesAny: cost.SubtypeSet{types.Island},
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

	first := addBasicLandPermanent(g, game.Player1, types.Island)
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("return-cost ability was legal with too few Islands")
	}
	second := addBasicLandPermanent(g, game.Player1, types.Island)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("return-cost ability was not legal with enough Islands")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(return-cost ability) = false, want true")
	}
	if !g.Players[game.Player1].Hand.Contains(first.CardInstanceID) ||
		!g.Players[game.Player1].Hand.Contains(second.CardInstanceID) {
		t.Fatal("returned Islands did not move to owner's hand")
	}
}

func TestActivatedAbilityReturnToHandCostRequiresTappedPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:               cost.AdditionalReturnToHand,
			Text:               "Return a tapped creature you control to its owner's hand",
			Amount:             1,
			MatchPermanentType: true,
			PermanentType:      types.Creature,
			RequireTapped:      true,
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

	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Test Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("return-cost ability was legal with only an untapped creature")
	}
	creature.Tapped = true
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(tapped return-cost ability) = false, want true")
	}
	if !g.Players[game.Player1].Hand.Contains(creature.CardInstanceID) {
		t.Fatal("returned tapped creature did not move to owner's hand")
	}
}

func TestActivatedAbilityReturnToHandCostMovesToOwnerHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Return Engine",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{{
				Kind:               cost.AdditionalReturnToHand,
				Text:               "Return a creature you control to its owner's hand",
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Creature,
			}},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.GainLife{
					Amount: game.Fixed(1),
					Player: game.ControllerReference(),
				}}},
			}.Ability(),
		}},
	}})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Borrowed Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	creature.Owner = game.Player2
	g.CardInstances[creature.CardInstanceID].Owner = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("applyAction(owner-hand return-cost ability) = false, want true")
	}
	if !g.Players[game.Player2].Hand.Contains(creature.CardInstanceID) ||
		g.Players[game.Player1].Hand.Contains(creature.CardInstanceID) {
		t.Fatal("returned creature did not move to owner's hand")
	}
}

func TestManaAbilityUntapsSourceAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := game.TapManaAbility(mana.G)
	body.Text = "{Q}: Add {G}."
	body.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalUntap, Text: "{Q}"}}
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:          "Untap Mana Engine",
		Types:         []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{body},
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("untap-cost mana ability was legal while source was untapped")
	}
	source.Tapped = true
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("untap-cost mana ability was not legal while source was tapped")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(untap-cost mana ability) = false, want true")
	}
	if source.Tapped {
		t.Fatal("mana source remained tapped after paying untap cost")
	}
}

func TestOncePerTurnActivatedAbilityIsTrackedAndResets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		Timing: game.OncePerTurn,
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(once-per-turn ability) = false, want true")
	}
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("once-per-turn ability remained legal after activation")
	}
	engine.resolveTopOfStack(g, nil)
	engine.advanceToNextTurn(g)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("once-per-turn ability was not legal after turn reset")
	}
}

func TestDuringUpkeepActivatedAbilityRequiresControllersUpkeep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		Timing: game.DuringUpkeep,
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player1
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("during-your-upkeep ability was legal during an opponent's upkeep")
	}
	g.Turn.ActivePlayer = game.Player1
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("during-your-upkeep ability was not legal during controller's upkeep")
	}
}

func TestActivatedAbilityWithSacrificeCostResolvesAfterSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: types.Creature},
		},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(sacrifice ability) = false, want true")
	}
	if _, ok := permanentByObjectID(g, source.ObjectID); ok {
		t.Fatal("source was not sacrificed as an activation cost")
	}
	engine.resolveTopOfStack(g, nil)
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want sacrificed source ability to draw", got)
	}
}
