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
	"github.com/natefinch/council4/mtg/rules/payment"
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

func TestActivatedAbilityRevealXColoredCardsAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	firstBlue := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Blue One",
		Colors: []color.Color{color.Blue},
	}})
	red := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Red",
		Colors: []color.Color{color.Red},
	}})
	secondBlue := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Blue Two",
		Colors: []color.Color{color.Blue},
	}})
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:           cost.AdditionalReveal,
			Text:           "Reveal X blue cards from your hand",
			AmountFromX:    true,
			Source:         zone.Hand,
			MatchCardColor: true,
			CardColor:      color.Blue,
		}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)
	actX2 := action.ActivateAbility(source.ObjectID, 0, nil, 2)
	if !containsAction(legal, actX2) {
		t.Fatal("reveal-X ability was not legal for X=2 with two blue cards in hand")
	}
	if containsAction(legal, action.ActivateAbility(source.ObjectID, 0, nil, 3)) {
		t.Fatal("reveal-X ability was legal for X=3 with only two blue cards in hand")
	}
	if !engine.applyAction(g, game.Player1, actX2) {
		t.Fatal("applyAction(reveal-X ability) = false, want true")
	}
	if !g.Players[game.Player1].Hand.Contains(firstBlue) ||
		!g.Players[game.Player1].Hand.Contains(red) ||
		!g.Players[game.Player1].Hand.Contains(secondBlue) {
		t.Fatal("revealed cards should remain in hand")
	}
	if !eventRevealedCardFromZone(g, game.Player1, source.CardInstanceID, firstBlue, zone.Hand) ||
		!eventRevealedCardFromZone(g, game.Player1, source.CardInstanceID, secondBlue, zone.Hand) {
		t.Fatal("reveal-X ability did not emit reveal events for both blue cards")
	}
	if eventRevealedCardFromZone(g, game.Player1, source.CardInstanceID, red, zone.Hand) {
		t.Fatal("reveal-X ability revealed a nonmatching red card")
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

func TestActivatedAbilityCollectEvidenceExilesSelectedGraveyardCardsAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalCollectEvidence,
			Text:   "Collect evidence 4",
			Amount: 4,
			Source: zone.Graveyard,
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
	first := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Two", 2))
	second := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Three", 3))

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("collect-evidence ability was not legal with enough graveyard mana value")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(collect-evidence ability) = false, want true")
	}
	if g.Players[game.Player1].Graveyard.Contains(first) ||
		g.Players[game.Player1].Graveyard.Contains(second) ||
		!g.Players[game.Player1].Exile.Contains(first) ||
		!g.Players[game.Player1].Exile.Contains(second) {
		t.Fatal("evidence cards did not move from graveyard to exile")
	}
}

func TestActivatedAbilityCollectEvidenceRequiresEnoughManaValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalCollectEvidence,
			Text:   "Collect evidence 4",
			Amount: 4,
			Source: zone.Graveyard,
		}},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}}}}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	cardID := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Two", 2))

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("collect-evidence ability was legal with insufficient graveyard mana value")
	}
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(insufficient collect evidence) = true, want false")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) || g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("failed collect-evidence payment mutated zones")
	}
}

func TestCollectEvidenceRejectsStalePreferenceWithoutMutation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Evidence Source"}})
	graveyardCard := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Four", 4))
	handCard := addCardToHand(g, game.Player1, evidenceCard("Stale Evidence", 4))

	ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID: game.Player1,
		Source:   source,
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalCollectEvidence,
			Text:   "Collect evidence 4",
			Amount: 4,
			Source: zone.Graveyard,
		}},
		Prefs: &payment.Preferences{EvidenceChoices: []id.ID{handCard}},
	})
	if ok {
		t.Fatal("stale collect-evidence preference paid successfully")
	}
	if !g.Players[game.Player1].Graveyard.Contains(graveyardCard) ||
		g.Players[game.Player1].Exile.Contains(graveyardCard) ||
		!g.Players[game.Player1].Hand.Contains(handCard) {
		t.Fatal("stale collect-evidence preference mutated zones")
	}
}

func TestCollectEvidenceAndExileCostCannotReuseGraveyardCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Evidence Source"}})
	graveyardCard := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Four", 4))

	ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID: game.Player1,
		Source:   source,
		AdditionalCosts: []cost.Additional{
			{
				Kind:   cost.AdditionalCollectEvidence,
				Text:   "Collect evidence 4",
				Amount: 4,
				Source: zone.Graveyard,
			},
			{
				Kind:   cost.AdditionalExile,
				Text:   "Exile a card from your graveyard",
				Amount: 1,
				Source: zone.Graveyard,
			},
		},
	})
	if ok {
		t.Fatal("collect-evidence and exile costs reused the same graveyard card")
	}
	if !g.Players[game.Player1].Graveyard.Contains(graveyardCard) ||
		g.Players[game.Player1].Exile.Contains(graveyardCard) {
		t.Fatal("failed combined collect-evidence/exile payment mutated zones")
	}
}

func TestActivatedAbilityCollectEvidenceAndExileCostChoosesDistinctGraveyardCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{
			{
				Kind:   cost.AdditionalCollectEvidence,
				Text:   "Collect evidence 4",
				Amount: 4,
				Source: zone.Graveyard,
			},
			{
				Kind:   cost.AdditionalExile,
				Text:   "Exile a card from your graveyard",
				Amount: 1,
				Source: zone.Graveyard,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}}}}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	evidence := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Four", 4))
	firstFodder := addCardToGraveyard(g, game.Player1, evidenceCard("First Fodder", 1))
	secondFodder := addCardToGraveyard(g, game.Player1, evidenceCard("Second Fodder", 1))

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("combined collect-evidence/exile ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(combined collect-evidence/exile ability) = false, want true")
	}
	exiledCount := 0
	for _, cardID := range []id.ID{evidence, firstFodder, secondFodder} {
		if g.Players[game.Player1].Exile.Contains(cardID) {
			exiledCount++
		}
	}
	if exiledCount < 2 || !g.Players[game.Player1].Exile.Contains(evidence) {
		t.Fatal("combined collect-evidence/exile payment did not exile distinct graveyard cards")
	}
}

func TestActivatedAbilityCollectEvidencePreservesCardsForLaterEvidenceCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{
			{
				Kind:   cost.AdditionalCollectEvidence,
				Text:   "Collect evidence 6",
				Amount: 6,
				Source: zone.Graveyard,
			},
			{
				Kind:   cost.AdditionalCollectEvidence,
				Text:   "Collect evidence 10",
				Amount: 10,
				Source: zone.Graveyard,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}}}}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	six := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Six", 6))
	ten := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Ten", 10))

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("combined collect-evidence ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(combined collect-evidence ability) = false, want true")
	}
	if !g.Players[game.Player1].Exile.Contains(six) ||
		!g.Players[game.Player1].Exile.Contains(ten) {
		t.Fatal("combined collect-evidence payment did not preserve cards for the later threshold")
	}
}

func TestActivatedAbilityCollectEvidencePreservesCreatureForLaterExileCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{
			{
				Kind:   cost.AdditionalCollectEvidence,
				Text:   "Collect evidence 4",
				Amount: 4,
				Source: zone.Graveyard,
			},
			{
				Kind:          cost.AdditionalExile,
				Text:          "Exile a creature card from your graveyard",
				Amount:        1,
				Source:        zone.Graveyard,
				MatchCardType: true,
				CardType:      types.Creature,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}}}}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	nonCreature := addCardToGraveyard(g, game.Player1, evidenceCard("Noncreature Evidence", 4))
	creature := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Creature Evidence",
		ManaCost: opt.Val(cost.Mana{cost.O(4)}),
		Types:    []types.Card{types.Creature},
	}})

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("collect-evidence plus typed-exile ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(collect-evidence plus typed-exile ability) = false, want true")
	}
	if !g.Players[game.Player1].Exile.Contains(nonCreature) ||
		!g.Players[game.Player1].Exile.Contains(creature) {
		t.Fatal("collect-evidence payment did not preserve the creature card for typed exile")
	}
}

func TestCollectEvidenceRejectsUnsupportedVariableManaValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalCollectEvidence,
			Text:   "Collect evidence 1",
			Amount: 1,
			Source: zone.Graveyard,
		}},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}}}}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Variable Evidence",
		ManaCost: opt.Val(cost.Mana{cost.X}),
		Types:    []types.Card{types.Sorcery},
	}})

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("collect-evidence ability was legal with only variable mana value evidence")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("variable evidence card moved to exile")
	}
}

func TestGraveyardActivatedAbilityChecksActivationCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Conditional Escape",
		Types: []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			ZoneOfFunction: zone.Graveyard,
			ActivationCondition: opt.Val(game.Condition{
				ControllerLifeAtLeast: 10,
			}),
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}}}.Ability(),
		}},
	}})
	g.Turn.PriorityPlayer = game.Player1
	act := action.ActivateAbility(sourceID, 0, nil, 0)

	g.Players[game.Player1].Life = 9
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard ability was legal while its activation condition was false")
	}
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(graveyard ability with false activation condition) = true, want false")
	}

	g.Players[game.Player1].Life = 10
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard ability was not legal while its activation condition was true")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(graveyard ability with true activation condition) = false, want true")
	}
}

func TestGraveyardCollectEvidencePreservesSourceForExileSourceCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Evidence Escape",
		ManaCost: opt.Val(cost.Mana{cost.O(4)}),
		Types:    []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			ZoneOfFunction: zone.Graveyard,
			AdditionalCosts: []cost.Additional{
				{
					Kind:   cost.AdditionalCollectEvidence,
					Text:   "Collect evidence 4",
					Amount: 4,
					Source: zone.Graveyard,
				},
				{
					Kind:   cost.AdditionalExileSource,
					Text:   "Exile this card from your graveyard",
					Amount: 1,
					Source: zone.Graveyard,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}}}.Ability(),
		}},
	}})
	otherEvidence := addCardToGraveyard(g, game.Player1, evidenceCard("Other Evidence", 4))
	g.Turn.PriorityPlayer = game.Player1
	act := action.ActivateAbility(sourceID, 0, nil, 0)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard collect-evidence/exile-source ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(graveyard collect-evidence/exile-source ability) = false, want true")
	}
	if !g.Players[game.Player1].Exile.Contains(sourceID) ||
		!g.Players[game.Player1].Exile.Contains(otherEvidence) {
		t.Fatal("graveyard collect-evidence payment did not preserve source for exile-source cost")
	}
}

func TestGraveyardExileSourcePreservesOtherCardForLaterCollectEvidenceCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Evidence Escape",
		ManaCost: opt.Val(cost.Mana{cost.O(4)}),
		Types:    []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			ZoneOfFunction: zone.Graveyard,
			AdditionalCosts: []cost.Additional{
				{
					Kind:   cost.AdditionalExileSource,
					Text:   "Exile this card from your graveyard",
					Amount: 1,
					Source: zone.Graveyard,
				},
				{
					Kind:   cost.AdditionalCollectEvidence,
					Text:   "Collect evidence 4",
					Amount: 4,
					Source: zone.Graveyard,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}}}.Ability(),
		}},
	}})
	otherEvidence := addCardToGraveyard(g, game.Player1, evidenceCard("Other Evidence", 4))
	g.Turn.PriorityPlayer = game.Player1
	act := action.ActivateAbility(sourceID, 0, nil, 0)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard exile-source/collect-evidence ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(graveyard exile-source/collect-evidence ability) = false, want true")
	}
	if !g.Players[game.Player1].Exile.Contains(sourceID) ||
		!g.Players[game.Player1].Exile.Contains(otherEvidence) {
		t.Fatal("graveyard exile-source payment did not preserve other card for collect-evidence cost")
	}
}

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

func evidenceCard(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
		Types:    []types.Card{types.Sorcery},
	}}
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

func TestLoyaltyActionPreservesAmbiguousTargetPartition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	planeswalker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Targeting Walker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3),
		LoyaltyAbilities: []game.LoyaltyAbility{{
			LoyaltyCost: 1,
			Content: game.Mode{Targets: []game.TargetSpec{
				{MinTargets: 0, MaxTargets: 1, Constraint: "creature"},
				{MinTargets: 0, MaxTargets: 1, Constraint: "creature"},
			}}.Ability(),
		}},
	}})
	planeswalker.Counters.Add(counter.Loyalty, 3)
	target := addCreaturePermanent(g, game.Player2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	act := action.ActivateAbilityWithModesAndTargetCounts(
		planeswalker.ObjectID,
		0,
		[]game.Target{game.PermanentTarget(target.ObjectID)},
		[]int{1, 0},
		0,
		nil,
	)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("loyalty action dropped ambiguous target partition")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(loyalty with target partition) = false, want true")
	}
}
