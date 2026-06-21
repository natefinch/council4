package rules

import (
	"slices"
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

func TestSplitSecondOffersOnlyComplexManaAbilitiesAndPass(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	manaRock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mana Rock",
		Types: []types.Card{types.Artifact}},
	}, mana.C, 1)
	complexBody := painlandColoredManaAbility(mana.C, 1)
	complexSource := addComplexManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Pain Rock",
		Types: []types.Card{types.Artifact}},
	}, &complexBody)
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
	if containsAction(actions, action.ActivateAbility(manaRock.ObjectID, 0, nil, 0)) {
		t.Fatal("split second offered a payment-only mana ability")
	}
	if !containsAction(actions, action.ActivateAbility(complexSource.ObjectID, 0, nil, 0)) {
		t.Fatal("split second suppressed a complex mana ability")
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
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.FlashbackKeyword{Cost: flashbackCost}},
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

func TestFlashbackAlternativeSacrificeCostCastsFromGraveyardAndExiles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Sacrifice Flashback Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(5)}),
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{Kind: game.Flashback}},
		}},
		AlternativeCosts: []cost.Alternative{{
			Label: "Flashback",
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: types.Creature},
			},
		}}},
	})
	g.Players[game.Player1].Hand.Remove(cardID)
	g.Players[game.Player1].Graveyard.Add(cardID)
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Goblin Token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	act := action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("flashback cast paying the sacrifice cost failed")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("creature was not sacrificed to pay the flashback cost")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Flashback || obj.SourceZone != zone.Graveyard {
		t.Fatalf("stack object = %+v, want flashback graveyard cast", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("flashback spell was not exiled after resolving")
	}
}

func TestFlashbackAlternativeCostCannotBeUsedFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	flashbackCost := cost.Mana{cost.G}
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Expensive Flashback Spell",
		Types:        []types.Card{types.Sorcery},
		ManaCost:     opt.Val(cost.Mana{cost.O(5)}),
		SpellAbility: opt.Val(game.AbilityContent{}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.FlashbackKeyword{Cost: flashbackCost}},
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

func TestFlashbackExilesWhenCountered(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Countered Flashback Spell",
		Types:        []types.Card{types.Sorcery},
		ManaCost:     opt.Val(cost.Mana{cost.O(5)}),
		SpellAbility: opt.Val(game.AbilityContent{}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.FlashbackKeyword{Cost: cost.Mana{cost.R}}},
		}}},
	})
	g.Players[game.Player1].Hand.Remove(cardID)
	g.Players[game.Player1].Graveyard.Add(cardID)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("flashback cast from graveyard failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Flashback {
		t.Fatalf("stack object = %+v, want flashback marker", obj)
	}
	if !counterStackObject(g, obj.ID) {
		t.Fatal("counterStackObject() = false, want true")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) ||
		g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("countered flashback spell was not exiled")
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
	resolveInstruction(engine, g, obj, game.CreateEmblem{EmblemAbilities: []game.Ability{&emblemAbility}}, nil)

	if permanent.Transformed || !permanent.PhasedOut {
		t.Fatalf("permanent transformed/phased = %v/%v, want false/true", permanent.Transformed, permanent.PhasedOut)
	}
	if len(g.Combat.Attackers) != 0 {
		t.Fatalf("attackers after phase out = %+v, want removed from combat", g.Combat.Attackers)
	}
	if len(g.Emblems) != 1 || g.Emblems[0].Owner != game.Player1 || len(g.Emblems[0].Abilities) != 1 {
		t.Fatalf("emblems = %+v, want one Player1 emblem", g.Emblems)
	}
	body, ok := g.Emblems[0].Abilities[0].(*game.StaticAbility)
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
