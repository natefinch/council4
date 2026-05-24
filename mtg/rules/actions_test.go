package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestLegalActionsIncludesPlayableLandBeforePass(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Forest",
		Types: []game.CardType{game.TypeLand},
	})
	addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Sol Ring",
		Types: []game.CardType{game.TypeArtifact},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if len(legal) != 2 {
		t.Fatalf("legal actions = %d, want 2", len(legal))
	}
	if !actionsEqual(legal[0], action.PlayLand(landID)) {
		t.Fatalf("first legal action = %+v, want PlayLand(%v)", legal[0], landID)
	}
	if legal[1].Kind != action.ActionPass {
		t.Fatalf("last legal action kind = %v, want %v", legal[1].Kind, action.ActionPass)
	}
}

func TestLegalActionsIncludesCastSpellAfterPlayLandBeforePass(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Forest",
		Types: []game.CardType{game.TypeLand},
	})
	spellID := addCardToHand(g, game.Player1, greenCreature())
	addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if len(legal) != 3 {
		t.Fatalf("legal actions = %d, want 3", len(legal))
	}
	if !actionsEqual(legal[0], action.PlayLand(landID)) {
		t.Fatalf("first legal action = %+v, want PlayLand(%v)", legal[0], landID)
	}
	if !actionsEqual(legal[1], action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatalf("second legal action = %+v, want CastSpell(%v)", legal[1], spellID)
	}
	if legal[2].Kind != action.ActionPass {
		t.Fatalf("last legal action kind = %v, want %v", legal[2].Kind, action.ActionPass)
	}
}

func TestLegalActionsDoesNotIncludePlayLandWhenUnavailable(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*game.Game)
	}{
		{
			name: "outside main phase",
			setup: func(g *game.Game) {
				g.Turn.Phase = game.PhaseBeginning
				g.Turn.Step = game.StepDraw
			},
		},
		{
			name: "land already played",
			setup: func(g *game.Game) {
				g.Turn.Phase = game.PhasePrecombatMain
				g.Turn.Step = game.StepNone
				g.Turn.LandsPlayedThisTurn = 1
			},
		},
		{
			name: "non-active player",
			setup: func(g *game.Game) {
				g.Turn.Phase = game.PhasePrecombatMain
				g.Turn.Step = game.StepNone
				g.Turn.PriorityPlayer = game.Player2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCardToHand(g, game.Player1, &game.CardDef{
				Name:  "Forest",
				Types: []game.CardType{game.TypeLand},
			})
			tt.setup(g)

			legal := engine.legalActions(g, game.Player1)

			if len(legal) != 1 {
				t.Fatalf("legal actions = %d, want 1", len(legal))
			}
			if legal[0].Kind != action.ActionPass {
				t.Fatalf("legal action kind = %v, want %v", legal[0].Kind, action.ActionPass)
			}
		})
	}
}

func TestCreatureAndSorceryLegalOnlyAtSorcerySpeed(t *testing.T) {
	tests := []struct {
		name  string
		card  *game.CardDef
		setup func(*game.Game)
	}{
		{
			name: "creature outside main phase",
			card: greenCreature(),
			setup: func(g *game.Game) {
				g.Turn.Phase = game.PhaseBeginning
				g.Turn.Step = game.StepDraw
			},
		},
		{
			name: "sorcery while stack non-empty",
			card: greenSorcery(),
			setup: func(g *game.Game) {
				g.Turn.Phase = game.PhasePrecombatMain
				g.Turn.Step = game.StepNone
				g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell})
			},
		},
		{
			name: "creature non-active player",
			card: greenCreature(),
			setup: func(g *game.Game) {
				g.Turn.Phase = game.PhasePrecombatMain
				g.Turn.Step = game.StepNone
				g.Turn.PriorityPlayer = game.Player2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			spellID := addCardToHand(g, game.Player1, tt.card)
			addBasicLandPermanent(g, game.Player1, "Forest")
			tt.setup(g)

			if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
				t.Fatal("cast spell action was legal outside sorcery timing")
			}
		})
	}
}

func TestInstantLegalWhileStackNonEmpty(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	instantID := addCardToHand(g, game.Player2, greenInstant())
	addBasicLandPermanent(g, game.Player2, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player2
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell})

	if !containsAction(engine.legalActions(g, game.Player2), action.CastSpell(instantID, nil, 0, nil)) {
		t.Fatal("instant was not legal while another object was on the stack")
	}
}

func TestFlashCreatureLegalAtInstantSpeed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	flashID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:      "Ambush Viper",
		ManaCost:  greenCost(),
		Types:     []game.CardType{game.TypeCreature},
		Abilities: []game.AbilityDef{{Keywords: []game.Keyword{game.Flash}}},
	})
	addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell})

	if !containsAction(engine.legalActions(g, game.Player1), action.CastSpell(flashID, nil, 0, nil)) {
		t.Fatal("flash creature was not legal at instant speed")
	}
}

func TestUnpayableSpellIsNotLegal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, greenCreature())
	addBasicLandPermanent(g, game.Player1, "Mountain")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("unpayable spell was legal")
	}
}

func TestLegalActionsIncludesPayableXValues(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cost := mana.Cost{mana.VariableMana(), mana.ColoredMana(mana.Green)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Gelatinous Genesis",
		ManaCost: &cost,
		Types:    []game.CardType{game.TypeSorcery},
	})
	addBasicLandPermanent(g, game.Player1, "Forest")
	addBasicLandPermanent(g, game.Player1, "Mountain")
	addBasicLandPermanent(g, game.Player1, "Island")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)
	for _, xValue := range []int{0, 1, 2} {
		if !containsAction(legal, action.CastSpell(spellID, nil, xValue, nil)) {
			t.Fatalf("legal actions do not include X=%d cast: %+v", xValue, legal)
		}
	}
	if containsAction(legal, action.CastSpell(spellID, nil, 3, nil)) {
		t.Fatalf("legal actions include unpayable X=3 cast: %+v", legal)
	}
}

func TestApplyActionCastXSpellPaysChosenX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cost := mana.Cost{mana.VariableMana(), mana.ColoredMana(mana.Green)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Gelatinous Genesis",
		ManaCost: &cost,
		Types:    []game.CardType{game.TypeSorcery},
	})
	addBasicLandPermanent(g, game.Player1, "Forest")
	addBasicLandPermanent(g, game.Player1, "Mountain")
	addBasicLandPermanent(g, game.Player1, "Island")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 2, nil)) {
		t.Fatal("applyAction(cast X=2) = false, want true")
	}
	obj := g.Stack.Peek()
	if obj == nil {
		t.Fatal("stack is empty after casting X spell")
	}
	if obj.XValue != 2 {
		t.Fatalf("stack X value = %d, want 2", obj.XValue)
	}
}

func TestCastSpellWithSacrificeAdditionalCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cost := mana.Cost{mana.ColoredMana(mana.Green)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Village Rites",
		ManaCost: &cost,
		Types:    []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				AdditionalCosts: []game.AdditionalCost{
					{Kind: game.AdditionalCostSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: game.TypeCreature},
				},
			},
		},
	})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:      "Goblin Token",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &game.PT{Value: 1},
		Toughness: &game.PT{Value: 1},
	})
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.CastSpell(spellID, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("spell with payable sacrifice cost was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(cast with sacrifice cost) = false, want true")
	}
	if permanentByObjectID(g, creature.ObjectID) != nil {
		t.Fatal("sacrificed creature remained on battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(creature.CardInstanceID) {
		t.Fatal("sacrificed creature was not put into graveyard")
	}
	if !forest.Tapped {
		t.Fatal("forest was not tapped to pay mana cost")
	}
	obj := g.Stack.Peek()
	if obj == nil || len(obj.AdditionalCostsPaid) != 1 || obj.AdditionalCostsPaid[0] != "Sacrifice a creature" {
		t.Fatalf("stack additional costs paid = %+v, want sacrifice cost", obj)
	}
}

func TestPaymentChoiceSelectsSacrificeAdditionalCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cost := mana.Cost{mana.ColoredMana(mana.Green)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Chosen Offering",
		ManaCost: &cost,
		Types:    []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				AdditionalCosts: []game.AdditionalCost{
					{Kind: game.AdditionalCostSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: game.TypeCreature},
				},
			},
		},
	})
	first := addCombatPermanent(g, game.Player1, &game.CardDef{Name: "First", Types: []game.CardType{game.TypeCreature}, Power: &game.PT{Value: 1}, Toughness: &game.PT{Value: 1}})
	second := addCombatPermanent(g, game.Player1, &game.CardDef{Name: "Second", Types: []game.CardType{game.TypeCreature}, Power: &game.PT{Value: 1}, Toughness: &game.PT{Value: 1}})
	addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	log := TurnLog{}

	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &log) {
		t.Fatal("applyActionWithChoices(cast with sacrifice choice) = false, want true")
	}
	if permanentByObjectID(g, first.ObjectID) == nil {
		t.Fatal("first creature was sacrificed, want second")
	}
	if permanentByObjectID(g, second.ObjectID) != nil {
		t.Fatal("chosen second creature remained on battlefield")
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoicePayment || log.Choices[0].Selected[0] != 1 {
		t.Fatalf("payment choice log = %+v, want selected payment option 1", log.Choices)
	}
}

func TestPaymentChoiceCanPayPhyrexianManaWithLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cost := mana.Cost{mana.PhyrexianMana(mana.Green)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Phyrexian Choice",
		ManaCost: &cost,
		Types:    []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{Kind: game.SpellAbility},
		},
	})
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	log := TurnLog{}

	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &log) {
		t.Fatal("applyActionWithChoices(cast with phyrexian life choice) = false, want true")
	}
	if got := g.Players[game.Player1].Life; got != 38 {
		t.Fatalf("life = %d, want 38", got)
	}
	if forest.Tapped {
		t.Fatal("forest was tapped despite choosing phyrexian life payment")
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoicePayment || log.Choices[0].Selected[0] != 1 {
		t.Fatalf("payment choice log = %+v, want selected payment option 1", log.Choices)
	}
}

func TestPaymentChoiceRejectsUnavailablePhyrexianLifeOption(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].Life = 1
	cost := mana.Cost{mana.PhyrexianMana(mana.Green)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Phyrexian Choice",
		ManaCost: &cost,
		Types:    []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{Kind: game.SpellAbility},
		},
	})
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("applyActionWithChoices(cast with invalid phyrexian life choice) = false, want fallback to mana")
	}
	if got := g.Players[game.Player1].Life; got != 1 {
		t.Fatalf("life = %d, want 1", got)
	}
	if !forest.Tapped {
		t.Fatal("forest was not tapped after invalid life choice fell back to mana")
	}
}

func TestPaymentChoiceDoesNotOvercommitPhyrexianLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].Life = 3
	cost := mana.Cost{mana.PhyrexianMana(mana.Black), mana.PhyrexianMana(mana.Black)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Double Phyrexian Choice",
		ManaCost: &cost,
		Types:    []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{Kind: game.SpellAbility},
		},
	})
	firstSwamp := addBasicLandPermanent(g, game.Player1, "Swamp")
	secondSwamp := addBasicLandPermanent(g, game.Player1, "Swamp")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {1}}}}

	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("applyActionWithChoices(cast with overcommitted phyrexian choices) = false, want fallback to mana")
	}
	if got := g.Players[game.Player1].Life; got != 1 {
		t.Fatalf("life = %d, want 1", got)
	}
	if !firstSwamp.Tapped && !secondSwamp.Tapped {
		t.Fatal("neither swamp was tapped for the second phyrexian symbol")
	}
}

func TestPaymentChoiceFallbackSelectsMultipleAdditionalCostObjects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cost := mana.Cost{mana.ColoredMana(mana.Green)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Double Offering",
		ManaCost: &cost,
		Types:    []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				AdditionalCosts: []game.AdditionalCost{
					{Kind: game.AdditionalCostSacrifice, Text: "Sacrifice two creatures", Amount: 2, MatchPermanentType: true, PermanentType: game.TypeCreature},
				},
			},
		},
	})
	first := addCombatPermanent(g, game.Player1, &game.CardDef{Name: "First", Types: []game.CardType{game.TypeCreature}, Power: &game.PT{Value: 1}, Toughness: &game.PT{Value: 1}})
	second := addCombatPermanent(g, game.Player1, &game.CardDef{Name: "Second", Types: []game.CardType{game.TypeCreature}, Power: &game.PT{Value: 1}, Toughness: &game.PT{Value: 1}})
	third := addCombatPermanent(g, game.Player1, &game.CardDef{Name: "Third", Types: []game.CardType{game.TypeCreature}, Power: &game.PT{Value: 1}, Toughness: &game.PT{Value: 1}})
	addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast with two sacrifices) = false, want true")
	}
	if permanentByObjectID(g, first.ObjectID) != nil || permanentByObjectID(g, second.ObjectID) != nil {
		t.Fatal("fallback did not sacrifice the first two creatures")
	}
	if permanentByObjectID(g, third.ObjectID) == nil {
		t.Fatal("fallback sacrificed the third creature, want it to remain")
	}
}

func TestAlternativeCostCanMakeSpellPayable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	normalCost := mana.Cost{mana.GenericMana(5)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Free Alternate",
		ManaCost: &normalCost,
		Types:    []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				AlternativeCosts: []game.AlternativeCost{
					{Label: "Cast for free"},
				},
			},
		},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.CastSpell(spellID, nil, 0, nil)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("spell with payable alternative cost was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(cast with alternative cost) = false, want true")
	}
	if obj := g.Stack.Peek(); obj == nil || obj.SourceID != spellID {
		t.Fatalf("stack top = %+v, want alternative-cost spell", obj)
	}
}

func TestPaymentChoiceSelectsAlternativeCostWithAdditionalCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	normalCost := mana.Cost{mana.ColoredMana(mana.Green)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Alternate Offering",
		ManaCost: &normalCost,
		Types:    []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				AlternativeCosts: []game.AlternativeCost{
					{
						Label: "Sacrifice instead",
						AdditionalCosts: []game.AdditionalCost{
							{Kind: game.AdditionalCostSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: game.TypeCreature},
						},
					},
				},
			},
		},
	})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{Name: "Offering", Types: []game.CardType{game.TypeCreature}, Power: &game.PT{Value: 1}, Toughness: &game.PT{Value: 1}})
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("applyActionWithChoices(cast with chosen alternative cost) = false, want true")
	}
	if forest.Tapped {
		t.Fatal("normal mana cost was paid despite choosing alternative cost")
	}
	if permanentByObjectID(g, creature.ObjectID) != nil {
		t.Fatal("alternative additional sacrifice cost was not paid")
	}
}

func TestSacrificedPermanentIsExcludedFromManaPaymentPlan(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cost := mana.Cost{mana.ColoredMana(mana.Green)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Costly Harvest",
		ManaCost: &cost,
		Types:    []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				AdditionalCosts: []game.AdditionalCost{
					{Kind: game.AdditionalCostSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: game.TypeCreature},
				},
			},
		},
	})
	dork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{
		Name:      "Llanowar Elves",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &game.PT{Value: 1},
		Toughness: &game.PT{Value: 1},
	}, mana.Green, 1)
	dork.SummoningSick = false
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("spell was legal by using the creature to both produce mana and pay sacrifice cost")
	}
	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast) = true, want false")
	}
	if dork.Tapped {
		t.Fatal("sacrificed candidate was tapped by failed payment")
	}
	if permanentByObjectID(g, dork.ObjectID) == nil {
		t.Fatal("sacrificed candidate left battlefield after failed payment")
	}
	if !g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("spell left hand after failed payment")
	}
}

func TestLegalActionsIncludesModalSpellModeChoices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, modalCharm())
	target := addCombatPermanent(g, game.Player2, &game.CardDef{
		Name:      "Silvercoat Lion",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &game.PT{Value: 2},
		Toughness: &game.PT{Value: 2},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)
	if !containsAction(legal, action.CastSpell(spellID, nil, 0, []int{0})) {
		t.Fatalf("legal actions do not include untargeted modal choice: %+v", legal)
	}
	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, []int{1})) {
		t.Fatalf("legal actions do not include targeted modal choice: %+v", legal)
	}
	if containsAction(legal, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatalf("legal actions include modal spell without chosen mode: %+v", legal)
	}
}

func TestModalSpellResolvesChosenModeOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, modalCharm())
	target := addCombatPermanent(g, game.Player2, &game.CardDef{
		Name:      "Silvercoat Lion",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &game.PT{Value: 2},
		Toughness: &game.PT{Value: 2},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.CastSpell(spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, []int{1})
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(modal spell) = false, want true")
	}
	engine.resolveTopOfStack(g, nil)
	if target.MarkedDamage != 2 {
		t.Fatalf("target damage = %d, want 2", target.MarkedDamage)
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("controller life = %d, want 40 because unchosen mode did not resolve", got)
	}
}

func TestEquipAbilityUsesStackAndAttachesOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	equipment := addCombatPermanent(g, game.Player1, equipEquipment())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:      "Runeclaw Bear",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &game.PT{Value: 2},
		Toughness: &game.PT{Value: 2},
	})
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
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
	if equipment.AttachedTo != nil {
		t.Fatal("equipment attached before equip ability resolved")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, nil)
	if equipment.AttachedTo == nil || *equipment.AttachedTo != creature.ObjectID {
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
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:      "Runeclaw Bear",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &game.PT{Value: 2},
		Toughness: &game.PT{Value: 2},
	})
	opponentCreature := addCombatPermanent(g, game.Player2, &game.CardDef{
		Name:      "Silvercoat Lion",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &game.PT{Value: 2},
		Toughness: &game.PT{Value: 2},
	})
	addBasicLandPermanent(g, game.Player1, "Forest")
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
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.AbilityDef{
		Kind:     game.ActivatedAbility,
		ManaCost: greenCost(),
		Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"}},
		Effects:  []game.Effect{{Type: game.EffectDamage, TargetIndex: 0, Amount: 2}},
	}))
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
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

func TestGeneralActivatedAbilityTapCostRespectsSummoningSickness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.AbilityDef{
		Kind:           game.ActivatedAbility,
		AdditionalCost: "{T}",
		Effects:        []game.Effect{{Type: game.EffectGainLife, TargetIndex: -1, Amount: 1}},
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

func TestOncePerTurnActivatedAbilityIsTrackedAndResets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.AbilityDef{
		Kind:    game.ActivatedAbility,
		Timing:  game.OncePerTurn,
		Effects: []game.Effect{{Type: game.EffectGainLife, TargetIndex: -1, Amount: 1}},
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

func TestActivatedAbilityWithSacrificeCostResolvesAfterSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.AbilityDef{
		Kind: game.ActivatedAbility,
		AdditionalCosts: []game.AdditionalCost{
			{Kind: game.AdditionalCostSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: game.TypeCreature},
		},
		Effects: []game.Effect{{Type: game.EffectDraw, TargetIndex: -1, Amount: 1}},
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(sacrifice ability) = false, want true")
	}
	if permanentByObjectID(g, source.ObjectID) != nil {
		t.Fatal("source was not sacrificed as an activation cost")
	}
	engine.resolveTopOfStack(g, nil)
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want sacrificed source ability to draw", got)
	}
}

func TestTargetedSpellIsNotLegalBeforeTargetingSupport(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Giant Growth",
		ManaCost: greenCost(),
		Types:    []game.CardType{game.TypeInstant},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Targets: []game.TargetSpec{
					{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
				},
			},
		},
	})
	addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("targeted spell was legal before targeting support")
	}
}

func TestApplyActionPlayLandMovesCardToBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Forest",
		Types: []game.CardType{game.TypeLand},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.PlayLand(landID)) {
		t.Fatal("applyAction() = false, want true")
	}
	if g.Players[game.Player1].Hand.Contains(landID) {
		t.Fatal("land remained in hand")
	}
	if g.Turn.LandsPlayedThisTurn != 1 {
		t.Fatalf("lands played = %d, want 1", g.Turn.LandsPlayedThisTurn)
	}
	if len(g.Battlefield) != 1 {
		t.Fatalf("battlefield permanents = %d, want 1", len(g.Battlefield))
	}
	permanent := g.Battlefield[0]
	if permanent.CardInstanceID != landID {
		t.Fatalf("permanent card ID = %v, want %v", permanent.CardInstanceID, landID)
	}
	if permanent.Controller != game.Player1 {
		t.Fatalf("permanent controller = %v, want %v", permanent.Controller, game.Player1)
	}
	if !permanent.SummoningSick {
		t.Fatal("permanent summoning sick = false, want true")
	}
}

func TestApplyActionCastSpellPaysAndPushesStackObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, greenCreature())
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction() = false, want true")
	}
	if g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("spell remained in hand")
	}
	if !forest.Tapped {
		t.Fatal("forest was not tapped to pay cost")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", g.Stack.Size())
	}
	obj := g.Stack.Peek()
	if obj.SourceID != spellID {
		t.Fatalf("stack source ID = %v, want %v", obj.SourceID, spellID)
	}
	if obj.Controller != game.Player1 {
		t.Fatalf("stack controller = %v, want %v", obj.Controller, game.Player1)
	}
}

func TestApplyActionInvalidPlayLandDoesNotMutate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Forest",
		Types: []game.CardType{game.TypeLand},
	})
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepDraw

	if engine.applyAction(g, game.Player1, action.PlayLand(landID)) {
		t.Fatal("applyAction() = true, want false")
	}
	if !g.Players[game.Player1].Hand.Contains(landID) {
		t.Fatal("land was removed from hand")
	}
	if len(g.Battlefield) != 0 {
		t.Fatalf("battlefield permanents = %d, want 0", len(g.Battlefield))
	}
	if g.Turn.LandsPlayedThisTurn != 0 {
		t.Fatalf("lands played = %d, want 0", g.Turn.LandsPlayedThisTurn)
	}
}

func TestApplyActionInvalidCastDoesNotMutate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, greenCreature())
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepDraw

	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction() = true, want false")
	}
	if !g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("spell was removed from hand")
	}
	if forest.Tapped {
		t.Fatal("forest was tapped by invalid cast")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
}

func TestSplitSecondAllowsOnlyManaAbilitiesAndPass(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	manaRock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{
		Name:  "Mana Rock",
		Types: []game.CardType{game.TypeArtifact},
	}, mana.Colorless, 1)
	instantID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Response",
		Types: []game.CardType{game.TypeInstant},
	})
	splitSecondID := g.IDGen.Next()
	g.CardInstances[splitSecondID] = &game.CardInstance{
		ID: splitSecondID,
		Def: &game.CardDef{
			Name:  "Split Second Spell",
			Types: []game.CardType{game.TypeInstant},
			Abilities: []game.AbilityDef{{
				Kind:     game.StaticAbility,
				Keywords: []game.Keyword{game.SplitSecond},
			}},
		},
		Owner: game.Player2,
	}
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: splitSecondID, Controller: game.Player2})
	g.Turn.PriorityPlayer = game.Player1

	actions := engine.legalActions(g, game.Player1)

	if containsAction(actions, action.CastSpell(instantID, nil, 0, nil)) {
		t.Fatal("split second allowed casting a non-mana response")
	}
	if !containsAction(actions, action.ActivateAbility(manaRock.ObjectID, 0, nil, 0)) {
		t.Fatal("split second suppressed a mana ability")
	}
	if !containsAction(actions, action.Pass()) {
		t.Fatal("split second legal actions omitted pass")
	}
}

func TestPlaneswalkerLoyaltyAbilityPaysLoyaltyAndOncePerTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	planeswalker := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:    "Test Walker",
		Types:   []game.CardType{game.TypePlaneswalker},
		Loyalty: intPtr(3),
		Abilities: []game.AbilityDef{{
			Kind:             game.ActivatedAbility,
			IsLoyaltyAbility: true,
			LoyaltyCost:      -2,
			Effects:          []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}},
		}},
	})
	planeswalker.Counters.Add(counter.Loyalty, 3)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	act := action.ActivateAbility(planeswalker.ObjectID, 0, nil, 0)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction loyalty ability = false, want true")
	}
	if got := planeswalker.Counters.Get(counter.Loyalty); got != 1 {
		t.Fatalf("loyalty counters = %d, want 1", got)
	}
	card := g.GetCardInstance(planeswalker.CardInstanceID)
	if canActivateLoyaltyAbility(g, game.Player1, planeswalker, &card.Def.Abilities[0], 0, nil, 0) {
		t.Fatal("loyalty ability could be activated twice in one turn")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("loyalty ability did not resolve its effect")
	}
}

func TestKickerSpellPaysKickerAndAppliesKickerEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	kickerCost := mana.Cost{mana.ColoredMana(mana.Green)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Kicker Spell",
		Types: []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{{
			Kind:          game.SpellAbility,
			Effects:       []game.Effect{{Type: game.EffectGainLife, Amount: 1, TargetIndex: -1}},
			KickerCost:    &kickerCost,
			KickerEffects: []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}},
		}},
	})
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastKickedSpell(spellID, nil, 0, nil)) {
		t.Fatal("kicked spell cast failed")
	}

	if !forest.Tapped {
		t.Fatal("kicker cost did not tap mana source")
	}
	obj := g.Stack.Peek()
	if obj == nil || !obj.KickerPaid {
		t.Fatalf("stack object = %+v, want KickerPaid", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != 41 || g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatalf("life/hand = %d/%d, want base and kicker effects", g.Players[game.Player1].Life, g.Players[game.Player1].Hand.Size())
	}
}

func TestKickedSpellPlansBaseAndKickerTogether(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	baseCost := mana.Cost{mana.GenericMana(1)}
	kickerCost := mana.Cost{mana.ColoredMana(mana.Green)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Greedy Kicker Spell",
		Types:    []game.CardType{game.TypeSorcery},
		ManaCost: &baseCost,
		Abilities: []game.AbilityDef{{
			Kind:          game.SpellAbility,
			KickerCost:    &kickerCost,
			KickerEffects: []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}},
		}},
	})
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if engine.canCastSpellWithKicker(g, game.Player1, spellID, nil, 0, nil, true) {
		t.Fatal("canCastSpellWithKicker() = true with one Forest for {1}+{G}, want false")
	}
	if engine.applyAction(g, game.Player1, action.CastKickedSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction kicked spell = true, want false")
	}
	if forest.Tapped || !g.Players[game.Player1].Hand.Contains(spellID) || g.Stack.Size() != 0 {
		t.Fatal("failed kicked cast mutated mana, hand, or stack")
	}
}

func TestFightEffectDealsMutualCreatureDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets: []game.Target{
			game.PermanentTarget(first.ObjectID),
			game.PermanentTarget(second.ObjectID),
		},
	}

	engine.resolveEffect(g, obj, game.Effect{Type: game.EffectFight}, nil)

	if first.MarkedDamage != 2 || second.MarkedDamage != 3 {
		t.Fatalf("fight damage = %d/%d, want 2/3", first.MarkedDamage, second.MarkedDamage)
	}
}

func TestTransformPhaseOutAndEmblemEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: permanent.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(permanent.ObjectID)},
	}

	engine.resolveEffect(g, obj, game.Effect{Type: game.EffectTransform, TargetIndex: 0}, nil)
	engine.resolveEffect(g, obj, game.Effect{Type: game.EffectPhaseOut, TargetIndex: 0}, nil)
	emblemAbility := game.AbilityDef{Kind: game.StaticAbility, Text: "Test emblem ability"}
	engine.resolveEffect(g, obj, game.Effect{Type: game.EffectCreateEmblem, EmblemAbilities: []game.AbilityDef{emblemAbility}}, nil)

	if !permanent.Transformed || !permanent.PhasedOut {
		t.Fatalf("permanent transformed/phased = %v/%v, want true/true", permanent.Transformed, permanent.PhasedOut)
	}
	if len(g.Combat.Attackers) != 0 {
		t.Fatalf("attackers after phase out = %+v, want removed from combat", g.Combat.Attackers)
	}
	if len(g.Emblems) != 1 || g.Emblems[0].Owner != game.Player1 || len(g.Emblems[0].Abilities) != 1 || g.Emblems[0].Abilities[0].Text != emblemAbility.Text {
		t.Fatalf("emblems = %+v, want one Player1 emblem", g.Emblems)
	}
}

func TestPhasedOutPermanentsPhaseInAndCannotActivate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	manaRock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{
		Name:  "Mana Rock",
		Types: []game.CardType{game.TypeArtifact},
	}, mana.Colorless, 1)
	manaRock.PhasedOut = true
	g.Turn.PriorityPlayer = game.Player1

	if len(engine.legalManaAbilityActions(g, game.Player1)) != 0 {
		t.Fatal("phased-out permanent produced a legal mana ability")
	}

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if manaRock.PhasedOut {
		t.Fatal("phased-out permanent did not phase in during controller's untap step")
	}
}

func addCardToHand(g *game.Game, playerID game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   def,
		Owner: playerID,
	}
	g.Players[playerID].Hand.Add(cardID)
	return cardID
}

func greenCreature() *game.CardDef {
	return &game.CardDef{
		Name:      "Runeclaw Bear",
		ManaCost:  greenCost(),
		ManaValue: 1,
		Types:     []game.CardType{game.TypeCreature},
	}
}

func greenInstant() *game.CardDef {
	return &game.CardDef{
		Name:     "Giant Growth",
		ManaCost: greenCost(),
		Types:    []game.CardType{game.TypeInstant},
	}
}

func greenSorcery() *game.CardDef {
	return &game.CardDef{
		Name:     "Explore",
		ManaCost: greenCost(),
		Types:    []game.CardType{game.TypeSorcery},
	}
}

func modalCharm() *game.CardDef {
	return &game.CardDef{
		Name:  "Test Charm",
		Types: []game.CardType{game.TypeInstant},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Modes: []game.Mode{
					{
						Text:    "You gain 3 life.",
						Effects: []game.Effect{{Type: game.EffectGainLife, TargetIndex: -1, Amount: 3}},
					},
					{
						Text:    "Deal 2 damage to target creature.",
						Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
						Effects: []game.Effect{{Type: game.EffectDamage, TargetIndex: 0, Amount: 2}},
					},
				},
			},
		},
	}
}

func equipEquipment() *game.CardDef {
	cost := mana.Cost{mana.ColoredMana(mana.Green)}
	return &game.CardDef{
		Name:     "Test Sword",
		Types:    []game.CardType{game.TypeArtifact},
		Subtypes: []string{"Equipment"},
		Abilities: []game.AbilityDef{
			{
				Kind:     game.ActivatedAbility,
				Keywords: []game.Keyword{game.Equip},
				ManaCost: &cost,
				Timing:   game.SorceryOnly,
				Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature you control"}},
			},
		},
	}
}

func activatedAbilityPermanent(ability *game.AbilityDef) *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{
		Name:      "Activated Creature",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &pt,
		Toughness: &pt,
		Abilities: []game.AbilityDef{*ability},
	}
}

func greenCost() *mana.Cost {
	cost := mana.Cost{mana.ColoredMana(mana.Green)}
	return &cost
}
