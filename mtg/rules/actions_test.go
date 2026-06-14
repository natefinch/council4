package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

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
