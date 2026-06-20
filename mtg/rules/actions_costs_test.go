package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

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

func TestCastSpellTwoTypeSacrificeCostAcceptsAltType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Costly Plunder",
		ManaCost: opt.Val(cost.Mana{cost.B}),
		Types:    []types.Card{types.Instant},
		AdditionalCosts: []cost.Additional{
			{
				Kind:               cost.AdditionalSacrifice,
				Text:               "sacrifice an artifact or creature",
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Artifact,
				PermanentTypeAlt:   types.Creature,
			},
		},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})
	// The controller has only an artifact; the union's alternative permanent
	// type must allow paying the cost by sacrificing it.
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Clue Token",
		Types: []types.Card{types.Artifact}},
	})
	swamp := addBasicLandPermanent(g, game.Player1, types.Swamp)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.CastSpell(spellID, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("spell with payable artifact-or-creature sacrifice cost was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(cast with two-type sacrifice cost) = false, want true")
	}
	if _, ok := permanentByObjectID(g, artifact.ObjectID); ok {
		t.Fatal("sacrificed artifact remained on battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(artifact.CardInstanceID) {
		t.Fatal("sacrificed artifact was not put into graveyard")
	}
	if !swamp.Tapped {
		t.Fatal("swamp was not tapped to pay mana cost")
	}
	obj, ok := g.Stack.Peek()
	if !ok || len(obj.AdditionalCostsPaid) != 1 ||
		obj.AdditionalCostsPaid[0] != "sacrifice an artifact or creature" {
		t.Fatalf("stack additional costs paid = %+v, want two-type sacrifice cost", obj)
	}
}

func TestCastSpellExileXCardsAdditionalCostBindsX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Harvest Pyre",
		ManaCost: opt.Val(cost.Mana{cost.X, cost.R}),
		Types:    []types.Card{types.Instant},
		AdditionalCosts: []cost.Additional{{
			Kind:        cost.AdditionalExile,
			Text:        "exile X cards from your graveyard",
			AmountFromX: true,
			Source:      zone.Graveyard,
		}},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})
	firstID := addCardToHand(g, game.Player1, greenCreature())
	g.Players[game.Player1].Hand.Remove(firstID)
	g.Players[game.Player1].Graveyard.Add(firstID)
	secondID := addCardToHand(g, game.Player1, greenCreature())
	g.Players[game.Player1].Hand.Remove(secondID)
	g.Players[game.Player1].Graveyard.Add(secondID)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	setSorcerySpeedTurn(g, game.Player1)

	// X=2 requires exiling two graveyard cards; the additional cost amount is
	// bound from the announced X value.
	act := action.CastSpell(spellID, nil, 2, nil)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(cast exile-X with X=2) = false, want true")
	}
	player := g.Players[game.Player1]
	if player.Graveyard.Contains(firstID) || player.Graveyard.Contains(secondID) {
		t.Fatal("exile-X additional cost left graveyard cards behind")
	}
	if !player.Exile.Contains(firstID) || !player.Exile.Contains(secondID) {
		t.Fatal("exile-X additional cost did not exile both graveyard cards")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.XValue != 2 || len(obj.AdditionalCostsPaid) != 1 {
		t.Fatalf("stack object = %+v, want X=2 with one additional cost paid", obj)
	}
}

func TestCastSpellExileXCardsAdditionalCostFailsWithoutEnoughCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Harvest Pyre",
		ManaCost: opt.Val(cost.Mana{cost.X, cost.R}),
		Types:    []types.Card{types.Instant},
		AdditionalCosts: []cost.Additional{{
			Kind:        cost.AdditionalExile,
			Text:        "exile X cards from your graveyard",
			AmountFromX: true,
			Source:      zone.Graveyard,
		}},
		SpellAbility: opt.Val(game.AbilityContent{})},
	})
	onlyID := addCardToHand(g, game.Player1, greenCreature())
	g.Players[game.Player1].Hand.Remove(onlyID)
	g.Players[game.Player1].Graveyard.Add(onlyID)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	setSorcerySpeedTurn(g, game.Player1)

	// Only one card is available to exile, so X=2 cannot pay the additional cost.
	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 2, nil)) {
		t.Fatal("exile-X cast was legal for X=2 with only one graveyard card")
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
	assertEvent(t, g.Events, game.EventAbilityActivated, func(event game.Event) bool {
		return event.Player == game.Player1 &&
			event.PermanentID == source.ObjectID &&
			event.AbilityIndex == 0 &&
			!event.ManaAbility
	})
}

func TestActivatedAbilityTapPermanentsUnionCostTapsEitherType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Sunshot Militia",
		Types: []types.Card{types.Enchantment},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{{
				Kind:               cost.AdditionalTapPermanents,
				Text:               "Tap two untapped artifacts and/or creatures you control",
				Amount:             2,
				MatchPermanentType: true,
				PermanentType:      types.Artifact,
				PermanentTypeAlt:   types.Creature,
			}},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}},
	}})
	artifactPermanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Idol",
		Types: []types.Card{types.Artifact},
	}})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Militiaman",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	enchantment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Banner",
		Types: []types.Card{types.Enchantment},
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("applyAction(tap artifact-and/or-creature union ability) = false, want true")
	}
	if !artifactPermanent.Tapped || !creature.Tapped {
		t.Fatalf("union tapped artifact/creature = %v/%v, want both true", artifactPermanent.Tapped, creature.Tapped)
	}
	if enchantment.Tapped {
		t.Fatal("enchantment outside the union was tapped")
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

func TestCommanderControlledAlternativeCostNormalAndFreeChoices(t *testing.T) {
	for _, test := range []struct {
		name       string
		choice     int
		wantTapped bool
	}{
		{name: "normal", choice: 0, wantTapped: true},
		{name: "free", choice: 1, wantTapped: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			spellID := addCardToHand(g, game.Player1, commanderAlternativeTestSpell(nil))
			commander := addCombatPermanent(g, game.Player1, greenCommanderWithCost())
			g.CommanderIDs[commander.CardInstanceID] = true
			island := addBasicLandPermanent(g, game.Player1, types.Island)
			g.Turn.Phase = game.PhasePrecombatMain
			g.Turn.Step = game.StepNone
			agents := [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: [][]int{{test.choice}}},
			}

			if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
				t.Fatal("cast failed")
			}
			if island.Tapped != test.wantTapped {
				t.Fatalf("island tapped = %v, want %v", island.Tapped, test.wantTapped)
			}
		})
	}
}

func TestCommanderControlledAlternativeCostConditionChanges(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, commanderAlternativeTestSpell(nil))
	commandZoneCommanderID := g.IDGen.Next()
	g.CardInstances[commandZoneCommanderID] = &game.CardInstance{
		ID:    commandZoneCommanderID,
		Def:   greenCommanderWithCost(),
		Owner: game.Player1,
	}
	g.CommanderIDs[commandZoneCommanderID] = true
	g.Players[game.Player1].CommanderInstanceID = commandZoneCommanderID
	g.Players[game.Player1].CommandZone.Add(commandZoneCommanderID)
	commander := addCombatPermanent(g, game.Player2, greenCommanderWithCost())
	g.CommanderIDs[commander.CardInstanceID] = true
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.CastSpell(spellID, nil, 0, nil)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("command-zone or opponent-controlled commander enabled free cast")
	}
	commander.Controller = game.Player1
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("controlled opponent commander did not enable free cast")
	}
	commander.PhasedOut = true
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("phased-out commander enabled free cast")
	}
	commander.PhasedOut = false
	commander.Controller = game.Player2
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("cast succeeded after commander control changed before payment")
	}
}

func TestCommanderControlledAlternativeCostRequiresAdditionalCosts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	additional := []cost.Additional{{Kind: cost.AdditionalDiscard, Text: "Discard a card"}}
	spellID := addCardToHand(g, game.Player1, commanderAlternativeTestSpell(additional))
	fodderID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fodder"}})
	commander := addCombatPermanent(g, game.Player1, greenCommanderWithCost())
	g.CommanderIDs[commander.CardInstanceID] = true
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("free cast with payable additional cost failed")
	}
	if g.Players[game.Player1].Hand.Contains(fodderID) ||
		!g.Players[game.Player1].Graveyard.Contains(fodderID) {
		t.Fatal("required discard additional cost was not paid")
	}
}

func TestCommanderControlledAlternativeCostWithGrantedGraveyardCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, commanderAlternativeTestSpell(nil))
	commander := addCombatPermanent(g, game.Player1, greenCommanderWithCost())
	g.CommanderIDs[commander.CardInstanceID] = true
	g.Players[game.Player1].Hand.Remove(spellID)
	g.Players[game.Player1].Graveyard.Add(spellID)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCastFromZone,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		CastFromZone:   zone.Graveyard,
		AffectedCardID: spellID,
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.CastSpellFromZone(spellID, zone.Graveyard, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("granted graveyard cast could not use commander alternative cost")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("free cast from graveyard failed")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.SourceZone != zone.Graveyard || obj.Flashback {
		t.Fatalf("stack object = %+v, want non-flashback graveyard cast", obj)
	}
}

func TestCommanderControlledAlternativeCostDoesNotReplaceForcedFlashback(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spell := commanderAlternativeTestSpell(nil)
	spell.AlternativeCosts = append(spell.AlternativeCosts, cost.Alternative{
		Label:    flashbackAlternativeLabel,
		ManaCost: opt.Val(cost.Mana{cost.G}),
	})
	spell.StaticAbilities = []game.StaticAbility{{
		KeywordAbilities: game.SimpleKeywords(game.Flashback),
	}}
	spellID := addCardToHand(g, game.Player1, spell)
	commander := addCombatPermanent(g, game.Player1, greenCommanderWithCost())
	g.CommanderIDs[commander.CardInstanceID] = true
	g.Players[game.Player1].Hand.Remove(spellID)
	g.Players[game.Player1].Graveyard.Add(spellID)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.CastSpellFromZone(spellID, zone.Graveyard, nil, 0, nil)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("commander alternative bypassed forced flashback cost")
	}
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("payable flashback cast failed")
	}
	if !forest.Tapped {
		t.Fatal("flashback mana cost was not paid")
	}
	if obj, ok := g.Stack.Peek(); !ok || !obj.Flashback {
		t.Fatalf("stack object = %+v, want flashback cast", obj)
	}
}

func TestCommanderControlledAlternativeCostFromExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, commanderAlternativeTestSpell(nil))
	commander := addCombatPermanent(g, game.Player1, greenCommanderWithCost())
	g.CommanderIDs[commander.CardInstanceID] = true
	g.Players[game.Player1].Hand.Remove(spellID)
	g.Players[game.Player1].Exile.Add(spellID)
	g.AdventureCards[spellID] = true
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.CastSpellFromZone(spellID, zone.Exile, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("permitted exile cast could not use commander alternative cost")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("free cast from exile failed")
	}
}

func commanderAlternativeTestSpell(additional []cost.Additional) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:            "Commander Alternate",
		ManaCost:        opt.Val(cost.Mana{cost.U}),
		Types:           []types.Card{types.Instant},
		SpellAbility:    opt.Val(game.AbilityContent{}),
		AdditionalCosts: additional,
		AlternativeCosts: []cost.Alternative{{
			Label:     "Cast without paying mana cost",
			Condition: cost.AlternativeConditionControlsCommander,
		}},
	}}
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
