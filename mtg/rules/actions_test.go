package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLegalActionsIncludesPlayableLandBeforePass(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
		Types: []types.Card{types.Land}},
	})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Unsupported Card"}})
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
	landID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
		Types: []types.Card{types.Land}},
	})
	spellID := addCardToHand(g, game.Player1, greenCreature())
	addBasicLandPermanent(g, game.Player1, types.Forest)
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
			addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
				Types: []types.Card{types.Land}},
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
			addBasicLandPermanent(g, game.Player1, types.Forest)
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
	addBasicLandPermanent(g, game.Player2, types.Forest)
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
	flashID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Ambush Viper",
		ManaCost:        greenCost(),
		Types:           []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{game.FlashStaticBody}},
	})
	addBasicLandPermanent(g, game.Player1, types.Forest)
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
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("unpayable spell was legal")
	}
}

func TestLegalActionsIncludesPayableXValues(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	manaCost := cost.Mana{cost.X, cost.G}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gelatinous Genesis",
		ManaCost: opt.Val(manaCost),
		Types:    []types.Card{types.Sorcery}},
	})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addBasicLandPermanent(g, game.Player1, types.Island)
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
	manaCost := cost.Mana{cost.X, cost.G}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gelatinous Genesis",
		ManaCost: opt.Val(manaCost),
		Types:    []types.Card{types.Sorcery}},
	})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addBasicLandPermanent(g, game.Player1, types.Island)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 2, nil)) {
		t.Fatal("applyAction(cast X=2) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack is empty after casting X spell")
	}
	if obj.XValue != 2 {
		t.Fatalf("stack X value = %d, want 2", obj.XValue)
	}
}

func TestCastSpellWithSacrificeAdditionalCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	manaCost := cost.Mana{cost.G}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Costly Creature",
		ManaCost: opt.Val(manaCost),
		Types:    []types.Card{types.Creature},
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: types.Creature},
		},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Goblin Token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.CastSpell(spellID, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("spell with payable sacrifice cost was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(cast with sacrifice cost) = false, want true")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("sacrificed creature remained on battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(creature.CardInstanceID) {
		t.Fatal("sacrificed creature was not put into graveyard")
	}
	if !forest.Tapped {
		t.Fatal("forest was not tapped to pay mana cost")
	}
	obj, ok := g.Stack.Peek()
	if !ok || len(obj.AdditionalCostsPaid) != 1 || obj.AdditionalCostsPaid[0] != "Sacrifice a creature" {
		t.Fatalf("stack additional costs paid = %+v, want sacrifice cost", obj)
	}
}

func TestCastSpellTapPermanentsCostRetriesAroundManaSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Tap Offering",
		ManaCost: opt.Val(cost.Mana{cost.G}),
		Types:    []types.Card{types.Sorcery},
		AdditionalCosts: []cost.Additional{{
			Kind:               cost.AdditionalTapPermanents,
			Text:               "Tap an untapped creature you control",
			Amount:             1,
			MatchPermanentType: true,
			PermanentType:      types.Creature,
		}},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})
	dork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Elvish Mystic",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}, mana.G, 1)
	dork.SummoningSick = false
	bear := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	setSorcerySpeedTurn(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast with tap-permanents cost) = false, want retry to use Bear for tap cost")
	}
	if !dork.Tapped {
		t.Fatal("mana creature was not tapped for mana")
	}
	if !bear.Tapped {
		t.Fatal("alternate creature was not tapped for tap-permanents cost")
	}
}

func TestCastSpellCannotReusePermanentForTapAndSacrificeCosts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spell := &game.CardDef{CardFace: game.CardFace{Name: "Double Offering",
		Types: []types.Card{types.Sorcery},
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalTapPermanents, Text: "Tap an untapped creature you control", Amount: 1, MatchPermanentType: true, PermanentType: types.Creature},
			{Kind: cost.AdditionalSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: types.Creature},
		},
		SpellAbility: opt.Val(game.AbilityContent{})},
	}
	spellID := addCardToHand(g, game.Player1, spell)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Only Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	setSorcerySpeedTurn(g, game.Player1)
	act := action.CastSpell(spellID, nil, 0, nil)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("spell was legal by reusing one creature for tap and sacrifice costs")
	}
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(cast with one tap-and-sacrifice creature) = true, want false")
	}

	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine = NewEngine(nil)
	spellID = addCardToHand(g, game.Player1, spell)
	onlyCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Only Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	second := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	setSorcerySpeedTurn(g, game.Player1)
	act = action.CastSpell(spellID, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("spell with separate tap and sacrifice creatures was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(cast with separate tap and sacrifice creatures) = false, want true")
	}
	_, firstStillPresent := permanentByObjectID(g, onlyCreature.ObjectID)
	_, secondStillPresent := permanentByObjectID(g, second.ObjectID)
	if firstStillPresent == secondStillPresent {
		t.Fatalf("battlefield presence first/second = %v/%v, want one sacrificed and one tapped", firstStillPresent, secondStillPresent)
	}
	if firstStillPresent && !onlyCreature.Tapped {
		t.Fatal("remaining first creature was not tapped")
	}
	if secondStillPresent && !second.Tapped {
		t.Fatal("remaining second creature was not tapped")
	}
}

func TestCastSpellRevealCostAttributesSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forestID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Forest",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
	}})
	spellID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Revealing Spell",
		Types: []types.Card{types.Sorcery},
		AdditionalCosts: []cost.Additional{{
			Kind:        cost.AdditionalReveal,
			Source:      zone.Hand,
			SubtypesAny: cost.SubtypeSet{types.Forest},
		}},
		SpellAbility: opt.Val(game.AbilityContent{}),
	}})
	setSorcerySpeedTurn(g, game.Player2)

	if !engine.applyAction(g, game.Player2, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast with reveal cost) = false, want true")
	}
	if !g.Players[game.Player2].Hand.Contains(forestID) {
		t.Fatal("revealed Forest left its owner's hand")
	}
	if !eventRevealedCardFromZone(g, game.Player2, spellID, forestID, zone.Hand) {
		t.Fatal("spell reveal cost did not attribute the source spell")
	}
}

func TestPaymentChoiceSelectsSacrificeAdditionalCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	manaCost := cost.Mana{cost.G}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Chosen Offering",
		ManaCost: opt.Val(manaCost),
		Types:    []types.Card{types.Sorcery},
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: types.Creature},
		},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})
	first := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1})}})
	second := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1})}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	log := TurnLog{}

	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &log) {
		t.Fatal("applyActionWithChoices(cast with sacrifice choice) = false, want true")
	}
	if _, ok := permanentByObjectID(g, first.ObjectID); !ok {
		t.Fatal("first creature was sacrificed, want second")
	}
	if _, ok := permanentByObjectID(g, second.ObjectID); ok {
		t.Fatal("chosen second creature remained on battlefield")
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoicePayment || log.Choices[0].Selected[0] != 1 {
		t.Fatalf("payment choice log = %+v, want selected payment option 1", log.Choices)
	}
}

func TestActivatedAbilityTapPermanentsCostRequiresUntappedMatches(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Aether Grid",
		Types: []types.Card{types.Enchantment},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{{
				Kind:               cost.AdditionalTapPermanents,
				Text:               "Tap two untapped artifacts you control",
				Amount:             2,
				MatchPermanentType: true,
				PermanentType:      types.Artifact,
			}},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Untapped Artifact", Types: []types.Card{types.Artifact}}})
	tapped := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Tapped Artifact", Types: []types.Card{types.Artifact}}})
	tapped.Tapped = true
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("tap-permanents activated ability was legal with too few untapped matching permanents")
	}
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(tap-permanents ability) = true, want false with too few untapped artifacts")
	}
}

func TestActivatedAbilityTapPermanentsCostTapsRequiredMatches(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "School Summoner",
		Types: []types.Card{types.Enchantment},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{{
				Kind:        cost.AdditionalTapPermanents,
				Text:        "Tap two untapped Merfolk you control",
				Amount:      2,
				SubtypesAny: cost.SubtypeSet{types.Merfolk},
			}},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}},
	}})
	first := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First Merfolk",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Merfolk},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	second := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second Merfolk",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Merfolk},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	soldier := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Soldier",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Soldier},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("applyAction(tap Merfolk ability) = false, want true")
	}
	if !first.Tapped || !second.Tapped {
		t.Fatalf("Merfolk tapped = %v/%v, want both true", first.Tapped, second.Tapped)
	}
	if soldier.Tapped {
		t.Fatal("non-Merfolk creature was tapped")
	}
}

func TestActivatedAbilityTapPermanentsCostCannotReuseTapSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bird Keeper",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bird, types.Wizard},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 2}),
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalTap},
				{
					Kind:        cost.AdditionalTapPermanents,
					Text:        "Tap two untapped Birds you control",
					Amount:      2,
					SubtypesAny: cost.SubtypeSet{types.Bird},
				},
			},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}},
	}})
	first := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First Bird",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bird},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("tap-source plus tap-two-Birds ability was legal with only one other untapped Bird")
	}
	second := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second Bird",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bird},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(tap source plus tap Birds ability) = false, want true with two other Birds")
	}
	if !source.Tapped || !first.Tapped || !second.Tapped {
		t.Fatalf("tapped source/first/second = %v/%v/%v, want all true", source.Tapped, first.Tapped, second.Tapped)
	}
}

func TestActivatedAbilityTapPermanentsCostExcludesManaTappedPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mana Grid",
		Types: []types.Card{types.Enchantment},
		ActivatedAbilities: []game.ActivatedAbility{{
			ManaCost: opt.Val(cost.Mana{cost.G}),
			AdditionalCosts: []cost.Additional{{
				Kind:               cost.AdditionalTapPermanents,
				Text:               "Tap an untapped creature you control",
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Creature,
			}},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}},
	}})
	dork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Elvish Mystic",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}, mana.G, 1)
	dork.SummoningSick = false
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("tap-permanents ability was legal by reusing the only creature as a mana source")
	}
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(tap-permanents mana ability) = true, want false when only creature is needed for mana")
	}
	if dork.Tapped {
		t.Fatal("mana creature was tapped while rejected activation was applied")
	}
}

func TestActivatedAbilityTapPermanentsCostRetriesAroundManaSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mana Grid",
		Types: []types.Card{types.Enchantment},
		ActivatedAbilities: []game.ActivatedAbility{{
			ManaCost: opt.Val(cost.Mana{cost.G}),
			AdditionalCosts: []cost.Additional{{
				Kind:               cost.AdditionalTapPermanents,
				Text:               "Tap an untapped creature you control",
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Creature,
			}},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}},
	}})
	dork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Elvish Mystic",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}, mana.G, 1)
	dork.SummoningSick = false
	bear := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(tap-permanents mana ability) = false, want retry to use Bear for tap cost")
	}
	if !dork.Tapped {
		t.Fatal("mana creature was not tapped for mana")
	}
	if !bear.Tapped {
		t.Fatal("alternate creature was not tapped for tap-permanents cost")
	}
}

func TestPaymentChoiceSelectsTapPermanentsAdditionalCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Apothecary",
		Types: []types.Card{types.Enchantment},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{{
				Kind:               cost.AdditionalTapPermanents,
				Text:               "Tap an untapped creature you control",
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Creature,
			}},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}},
	}})
	first := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	second := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	log := TurnLog{}

	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0), agents, &log) {
		t.Fatal("applyActionWithChoices(activate with tap choice) = false, want true")
	}
	if first.Tapped {
		t.Fatal("first creature was tapped, want chosen second creature")
	}
	if !second.Tapped {
		t.Fatal("chosen second creature was not tapped")
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoicePayment || log.Choices[0].Selected[0] != 1 {
		t.Fatalf("payment choice log = %+v, want selected payment option 1", log.Choices)
	}
}

func TestPaymentChoiceCanPayPhyrexianManaWithLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	manaCost := cost.Mana{cost.PhyrexianMana(mana.G)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Phyrexian Choice",
		ManaCost:     opt.Val(manaCost),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
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
	manaCost := cost.Mana{cost.PhyrexianMana(mana.G)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Phyrexian Choice",
		ManaCost:     opt.Val(manaCost),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
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
	manaCost := cost.Mana{cost.PhyrexianMana(mana.B), cost.PhyrexianMana(mana.B)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Double Phyrexian Choice",
		ManaCost:     opt.Val(manaCost),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})
	firstSwamp := addBasicLandPermanent(g, game.Player1, types.Swamp)
	secondSwamp := addBasicLandPermanent(g, game.Player1, types.Swamp)
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
	manaCost := cost.Mana{cost.G}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Double Offering",
		ManaCost: opt.Val(manaCost),
		Types:    []types.Card{types.Sorcery},
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalSacrifice, Text: "Sacrifice two creatures", Amount: 2, MatchPermanentType: true, PermanentType: types.Creature},
		},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})
	first := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1})}})
	second := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1})}})
	third := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Third", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1})}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast with two sacrifices) = false, want true")
	}
	if _, ok := permanentByObjectID(g, first.ObjectID); ok {
		t.Fatal("fallback did not sacrifice the first creature")
	}
	if _, ok := permanentByObjectID(g, second.ObjectID); ok {
		t.Fatal("fallback did not sacrifice the second creature")
	}
	if _, ok := permanentByObjectID(g, third.ObjectID); !ok {
		t.Fatal("fallback sacrificed the third creature, want it to remain")
	}
}

func TestAlternativeCostCanMakeSpellPayable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	normalCost := cost.Mana{cost.O(5)}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Free Alternate",
		ManaCost: opt.Val(normalCost),
		Types:    []types.Card{types.Creature},
		AlternativeCosts: []cost.Alternative{
			{Label: "Cast for free"},
		},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
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
	if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != spellID {
		t.Fatalf("stack top = %+v, want alternative-cost spell", obj)
	}
}

func TestPaymentChoiceSelectsAlternativeCostWithAdditionalCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	normalCost := cost.Mana{cost.G}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Alternate Offering",
		ManaCost: opt.Val(normalCost),
		Types:    []types.Card{types.Sorcery},
		AlternativeCosts: []cost.Alternative{
			{
				Label: "Sacrifice instead",
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: types.Creature},
				},
			},
		},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Offering", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1})}})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("applyActionWithChoices(cast with chosen alternative cost) = false, want true")
	}
	if forest.Tapped {
		t.Fatal("normal mana cost was paid despite choosing alternative cost")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("alternative additional sacrifice cost was not paid")
	}
}

func TestSacrificedPermanentIsExcludedFromManaPaymentPlan(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	manaCost := cost.Mana{cost.G}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Costly Harvest",
		ManaCost: opt.Val(manaCost),
		Types:    []types.Card{types.Sorcery},
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: types.Creature},
		},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})
	dork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Llanowar Elves",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	}, mana.G, 1)
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
	if _, ok := permanentByObjectID(g, dork.ObjectID); !ok {
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
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Silvercoat Lion",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
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

func TestModalSpellSupportsChooseTwo(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, modalSpellWithModeRange(2, 2))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(spellID, nil, 0, []int{0, 1})) {
		t.Fatal("legal actions did not include first choose-two combination")
	}
	if !containsAction(legal, action.CastSpell(spellID, nil, 0, []int{1, 2})) {
		t.Fatal("legal actions did not include second choose-two combination")
	}
	if containsAction(legal, action.CastSpell(spellID, nil, 0, []int{0})) {
		t.Fatal("legal actions included too few chosen modes")
	}
}

func TestModalSpellSupportsOneOrBothAndUpToOne(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	oneOrBothID := addCardToHand(g, game.Player1, modalSpellWithModeRange(1, 2))
	upToOneID := addCardToHand(g, game.Player1, modalSpellWithModeRange(0, 1))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(oneOrBothID, nil, 0, []int{0, 1})) {
		t.Fatal("one-or-both modal spell did not include both modes")
	}
	if containsAction(legal, action.CastSpell(oneOrBothID, nil, 0, nil)) {
		t.Fatal("one-or-both modal spell included no modes")
	}
	if !containsAction(legal, action.CastSpell(upToOneID, nil, 0, nil)) {
		t.Fatal("choose-up-to-one modal spell did not include no modes")
	}
	if !containsAction(legal, action.CastSpell(upToOneID, nil, 0, []int{1})) {
		t.Fatal("choose-up-to-one modal spell did not include one mode")
	}
}

func TestModalSpellSupportsChooseThreeAndAllModes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	chooseThreeID := addCardToHand(g, game.Player1, modalSpellWithModeRange(3, 3))
	allModesID := addCardToHand(g, game.Player1, modalSpellWithModeRange(3, 3))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(chooseThreeID, nil, 0, []int{0, 1, 2})) {
		t.Fatal("choose-three modal spell did not include all three modes")
	}
	if containsAction(legal, action.CastSpell(chooseThreeID, nil, 0, []int{0, 1})) {
		t.Fatal("choose-three modal spell included too few modes")
	}
	if !containsAction(legal, action.CastSpell(allModesID, nil, 0, []int{0, 1, 2})) {
		t.Fatal("all-modes modal spell did not include all modes")
	}
}

func TestModalSpellDuplicateModesAreCanonicalized(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, modalSpellWithDuplicateModes(2, 2))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	for _, modes := range [][]int{{0, 0}, {0, 1}, {1, 1}} {
		if !containsAction(legal, action.CastSpell(spellID, nil, 0, modes)) {
			t.Fatalf("legal actions did not include duplicate-mode choice %+v", modes)
		}
	}
	if containsAction(legal, action.CastSpell(spellID, nil, 0, []int{1, 0})) {
		t.Fatal("legal actions included non-canonical duplicate-mode permutation")
	}
}

func TestModalSpellResolvesChosenModeOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, modalCharm())
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Silvercoat Lion",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
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

func TestTargetedSpellIsNotLegalBeforeTargetingSupport(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Giant Growth",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
			},
		}.Ability())},
	})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("targeted spell was legal before targeting support")
	}
}

func TestApplyActionPlayLandMovesCardToBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
		Types: []types.Card{types.Land}},
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
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
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
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack is empty after casting spell")
	}
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
	landID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
		Types: []types.Card{types.Land}},
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
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
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
	manaRock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mana Rock",
		Types: []types.Card{types.Artifact}},
	}, mana.C, 1)
	instantID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Response",
		Types: []types.Card{types.Instant}},
	})
	splitSecondID := g.IDGen.Next()
	g.CardInstances[splitSecondID] = &game.CardInstance{
		ID: splitSecondID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Split Second Spell",
			Types:           []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{game.SplitSecondStaticBody}},
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

func TestKickerSpellPaysKickerAndAppliesKickerEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	kickerCost := cost.Mana{cost.G}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Kicker Spell",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{
				Cost: kickerCost,
				BonusContent: game.Mode{
					Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
				}.Ability(),
			}},
		}}},
	})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastKickedSpell(spellID, nil, 0, nil)) {
		t.Fatal("kicked spell cast failed")
	}

	if !forest.Tapped {
		t.Fatal("kicker cost did not tap mana source")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.KickerPaid {
		t.Fatalf("stack object = %+v, want KickerPaid", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != 41 || g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatalf("life/hand = %d/%d, want base and kicker effects", g.Players[game.Player1].Life, g.Players[game.Player1].Hand.Size())
	}
}

func TestKickedPermanentPreservesKickerOnEnterEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Kicked Creature",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.G}}},
		}},
	}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastKickedSpell(spellID, nil, 0, nil)) {
		t.Fatal("kicked permanent cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == game.EventPermanentEnteredBattlefield {
			if !g.Events[i].KickerPaid {
				t.Fatal("permanent enter event lost kicker payment")
			}
			return
		}
	}
	t.Fatal("missing permanent enter event")
}

func TestKickedSpellPlansBaseAndKickerTogether(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	baseCost := cost.Mana{cost.O(1)}
	kickerCost := cost.Mana{cost.G}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Greedy Kicker Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(baseCost),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{
				Cost: kickerCost,
				BonusContent: game.Mode{
					Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
				}.Ability(),
			}},
		}}},
	})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
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

func TestFlashbackCastsFromGraveyardAndExilesOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	flashbackCost := cost.Mana{cost.G}
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Flashback Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(5)}),
		AlternativeCosts: []cost.Alternative{{
			Label:    flashbackAlternativeLabel,
			ManaCost: opt.Val(flashbackCost),
		}},
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Flashback),
		}}},
	})
	g.Players[game.Player1].Hand.Remove(cardID)
	g.Players[game.Player1].Graveyard.Add(cardID)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	act := action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("flashback cast from graveyard failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Flashback {
		t.Fatalf("stack object = %+v, want flashback marker", obj)
	}
	if obj.SourceZone != zone.Graveyard {
		t.Fatalf("stack object source zone = %v, want graveyard", obj.SourceZone)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("flashback spell returned to graveyard")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("flashback spell was not exiled")
	}
}

func TestFlashbackAlternativeCostCannotBeUsedFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	flashbackCost := cost.Mana{cost.G}
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Expensive Flashback Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(5)}),
		AlternativeCosts: []cost.Alternative{{
			Label:    flashbackAlternativeLabel,
			ManaCost: opt.Val(flashbackCost),
		}},
		SpellAbility: opt.Val(game.AbilityContent{}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Flashback),
		}}},
	})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("flashback alternative cost was payable from hand")
	}
}

func TestGraveyardAbilityExilesSourceCardAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fanatic Source",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Snake, types.Druid},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 4}),
		ActivatedAbilities: []game.ActivatedAbility{
			game.EternalizeActivatedBody(cost.Mana{cost.O(0)}, types.Snake, types.Druid),
		}},
	})
	g.Players[game.Player1].Hand.Remove(cardID)
	g.Players[game.Player1].Graveyard.Add(cardID)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	act := action.ActivateAbility(cardID, 0, nil, 0)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard source-exile ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("graveyard source-exile ability activation failed")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("source card remained in graveyard after paying exile-source cost")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("source card was not exiled to pay cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	var token *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			token = permanent
			break
		}
	}
	if token == nil || token.TokenDef == nil {
		t.Fatal("source-card ability did not create a token")
	}
	if got := token.TokenDef.Subtypes; !slices.Equal(got, []types.Sub{types.Zombie, types.Snake, types.Druid}) {
		t.Fatalf("token subtypes = %+v, want Zombie Snake Druid", got)
	}
	if got := token.TokenDef.Colors; !slices.Equal(got, []color.Color{color.Black}) {
		t.Fatalf("token colors = %+v, want black", got)
	}
}

func TestGraveyardOnlyAbilityIsNotActivatedFromBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Wrong Zone Source",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		ActivatedAbilities: []game.ActivatedAbility{{
			ZoneOfFunction: zone.Graveyard,
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}}},
	})
	g.Turn.PriorityPlayer = game.Player1
	act := action.ActivateAbility(permanent.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard-only ability was legal from battlefield")
	}
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("graveyard-only ability activated from battlefield")
	}
}

func TestRuleEffectAllowsCastingFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, greenInstant())
	g.Players[game.Player1].Hand.Remove(cardID)
	g.Players[game.Player1].Graveyard.Add(cardID)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Graveyard Permission",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCastFromZone,
				AffectedPlayer: game.PlayerYou,
				CastFromZone:   zone.Graveyard,
			}},
		}}},
	})
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("rule effect did not allow graveyard cast")
	}
}

func TestProwessTriggersOnNoncreatureSpellCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	prowess := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Prowess)
	spellID := addCardToHand(g, game.Player1, greenInstant())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast instant failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("prowess trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, prowess); got != 3 {
		t.Fatalf("prowess power = %d, want 3", got)
	}
}

func TestFightEffectDealsMutualCreatureDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets: []game.Target{
			game.PermanentTarget(first.ObjectID),
			game.PermanentTarget(second.ObjectID),
		},
	}

	resolveFightTargets(g, obj, 0, 1)

	if first.MarkedDamage != 2 || second.MarkedDamage != 3 {
		t.Fatalf("fight damage = %d/%d, want 2/3", first.MarkedDamage, second.MarkedDamage)
	}
}

func TestFightEffectUsesExplicitRelatedTargetIndex(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ignored := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets: []game.Target{
			game.PermanentTarget(ignored.ObjectID),
			game.PermanentTarget(first.ObjectID),
			game.PermanentTarget(second.ObjectID),
		},
	}

	resolveFightTargets(g, obj, 1, 2)

	if ignored.MarkedDamage != 0 {
		t.Fatalf("ignored target marked damage = %d, want 0", ignored.MarkedDamage)
	}
	if first.MarkedDamage != 2 || second.MarkedDamage != 3 {
		t.Fatalf("fight damage = %d/%d, want 2/3", first.MarkedDamage, second.MarkedDamage)
	}
}

func TestFightEffectWithMissingRelatedTargetDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets: []game.Target{
			game.PermanentTarget(first.ObjectID),
		},
	}

	resolveFightTargets(g, obj, 0, 1)

	if first.MarkedDamage != 0 {
		t.Fatalf("target marked damage = %d, want 0", first.MarkedDamage)
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

	resolveInstruction(engine, g, obj, game.Transform{Object: game.TargetPermanentReference(0)}, nil)
	resolveInstruction(engine, g, obj, game.PhaseOut{Object: game.TargetPermanentReference(0)}, nil)
	emblemAbility := game.StaticAbility{Text: "Test emblem ability"}
	resolveInstruction(engine, g, obj, game.CreateEmblem{EmblemAbilities: []game.Ability{emblemAbility}}, nil)

	if permanent.Transformed || !permanent.PhasedOut {
		t.Fatalf("permanent transformed/phased = %v/%v, want false/true", permanent.Transformed, permanent.PhasedOut)
	}
	if len(g.Combat.Attackers) != 0 {
		t.Fatalf("attackers after phase out = %+v, want removed from combat", g.Combat.Attackers)
	}
	if len(g.Emblems) != 1 || g.Emblems[0].Owner != game.Player1 || len(g.Emblems[0].Abilities) != 1 {
		t.Fatalf("emblems = %+v, want one Player1 emblem", g.Emblems)
	}
	body, ok := g.Emblems[0].Abilities[0].(game.StaticAbility)
	if !ok || body.Text != emblemAbility.Text {
		t.Fatalf("emblem body = %+v, want static body %q", g.Emblems[0].Abilities[0], emblemAbility.Text)
	}
}

func TestPhasedOutPermanentsPhaseInAndCannotActivate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	manaRock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mana Rock",
		Types: []types.Card{types.Artifact}},
	}, mana.C, 1)
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
	return &game.CardDef{CardFace: game.CardFace{Name: "Runeclaw Bear",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Creature}},
	}
}

func greenInstant() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Giant Growth",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant}},
	}
}

func greenSorcery() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Explore",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Sorcery}},
	}
}

func modalCharm() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Test Charm",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.AbilityContent{
			MinModes: 1,
			MaxModes: 1,
			Modes: []game.Mode{
				{
					Text:     "You gain 3 life.",
					Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(3), Player: game.ControllerReference()}}},
				},
				{
					Text:     "Deal 2 damage to target creature.",
					Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
					Sequence: []game.Instruction{{Primitive: game.Damage{Recipient: game.AnyTargetDamageRecipient(0), Amount: game.Fixed(2)}}},
				},
			},
		})},
	}
}

func modalSpellWithModeRange(minModes, maxModes int) *game.CardDef {
	return modalSpellWithModeRangeAndDuplicates(minModes, maxModes, false)
}

func modalSpellWithDuplicateModes(minModes, maxModes int) *game.CardDef {
	return modalSpellWithModeRangeAndDuplicates(minModes, maxModes, true)
}

func modalSpellWithModeRangeAndDuplicates(minModes, maxModes int, allowDuplicates bool) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Flexible Charm",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{
			MinModes:            minModes,
			MaxModes:            maxModes,
			AllowDuplicateModes: allowDuplicates,
			Modes: []game.Mode{
				{Text: "You gain 1 life.", Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}},
				{Text: "You gain 2 life.", Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()}}}},
				{Text: "You gain 3 life.", Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(3), Player: game.ControllerReference()}}}},
			},
		})},
	}
}

func equipEquipment() *game.CardDef {
	manaCost := cost.Mana{cost.G}
	return &game.CardDef{CardFace: game.CardFace{Name: "Test Sword",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
		ActivatedAbilities: []game.ActivatedAbility{{
			KeywordAbilities: []game.KeywordAbility{game.EquipKeyword{Cost: manaCost}},
			ManaCost:         opt.Val(manaCost),
			Timing:           game.SorceryOnly,
			Content: game.Mode{
				Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature you control"}},
			}.Ability(),
		}}},
	}
}

func activatedAbilityPermanent(ability *game.ActivatedAbility) *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{CardFace: game.CardFace{Name: "Activated Creature",
		Types:              []types.Card{types.Creature},
		Power:              opt.Val(pt),
		Toughness:          opt.Val(pt),
		ActivatedAbilities: []game.ActivatedAbility{*ability}},
	}
}

func greenCost() opt.V[cost.Mana] {
	manaCost := cost.Mana{cost.G}
	return opt.Val(manaCost)
}

func TestPlaneswalkerLoyaltyAbilityPaysLoyaltyAndOncePerTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	planeswalker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Test Walker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3),
		LoyaltyAbilities: []game.LoyaltyAbility{{
			LoyaltyCost: -2,
			Content: game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
				}},
			}.Ability(),
		}}},
	})
	planeswalker.Counters.Add(counter.Loyalty, 3)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
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
	card, ok := g.GetCardInstance(planeswalker.CardInstanceID)
	if !ok {
		t.Fatal("planeswalker card instance not found")
	}
	if canActivateLoyaltyAbility(g, game.Player1, planeswalker, &card.Def.LoyaltyAbilities[0], 0, nil, 0) {
		t.Fatal("loyalty ability could be activated twice in one turn")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("loyalty ability did not resolve its effect")
	}
}
