package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func TestCanPayCostWithUntappedForest(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	manaCost := cost.Mana{cost.G}

	if !canPayCost(g, game.Player1, &manaCost) {
		t.Fatal("canPayCost() = false, want true")
	}
}

func TestCanPayCostRejectsWrongBasicLandColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	manaCost := cost.Mana{cost.G}

	if canPayCost(g, game.Player1, &manaCost) {
		t.Fatal("canPayCost() = true with wrong basic land color, want false")
	}
}

func TestCanPayGenericCostWithAnyBasicLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	manaCost := cost.Mana{cost.O(1)}

	if !canPayCost(g, game.Player1, &manaCost) {
		t.Fatal("canPayCost() = false, want true")
	}
}

func TestGenericSymbolsDoNotConsumeManaNeededByColoredSymbols(t *testing.T) {
	t.Run("pool", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Players[game.Player1].ManaPool.Add(mana.W, 1)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		manaCost := cost.Mana{cost.O(1), cost.W}

		if !canPayCost(g, game.Player1, &manaCost) {
			t.Fatal("canPayCost() = false for pool {W,G} paying {1}{W}, want true")
		}
	})
	t.Run("lands", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addBasicLandPermanent(g, game.Player1, types.Plains)
		addBasicLandPermanent(g, game.Player1, types.Forest)
		manaCost := cost.Mana{cost.O(1), cost.W}

		if !canPayCost(g, game.Player1, &manaCost) {
			t.Fatal("canPayCost() = false for Plains+Forest paying {1}{W}, want true")
		}
	})
}

func TestCanPayColorlessCostOnlyWithColorlessMana(t *testing.T) {
	tests := []struct {
		name string
		add  func(*game.Game)
		want bool
	}{
		{
			name: "colorless source",
			add: func(g *game.Game) {
				addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Wastes",
					Types: []types.Card{types.Land}},
				}, mana.C, 1)
			},
			want: true,
		},
		{
			name: "colored source",
			add: func(g *game.Game) {
				addBasicLandPermanent(g, game.Player1, types.Forest)
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			tt.add(g)
			manaCost := cost.Mana{cost.C}

			if got := canPayCost(g, game.Player1, &manaCost); got != tt.want {
				t.Fatalf("canPayCost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPayCostTapsLandUsed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	manaCost := cost.Mana{cost.G}

	if !payTestGenericCost(g, game.Player1, &manaCost) {
		t.Fatal("payCost() = false, want true")
	}
	if !forest.Tapped {
		t.Fatal("forest was not tapped")
	}
	if !g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatalf("mana pool total = %d, want 0", g.Players[game.Player1].ManaPool.Total())
	}
}

func TestTappedLandCannotPayAgain(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	manaCost := cost.Mana{cost.G}

	if !payTestGenericCost(g, game.Player1, &manaCost) {
		t.Fatal("first payCost() = false, want true")
	}
	if canPayCost(g, game.Player1, &manaCost) {
		t.Fatal("canPayCost() = true after land tapped, want false")
	}
}

func TestPayCostFailureDoesNotMutate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	manaCost := cost.Mana{cost.G, cost.G}

	if payTestGenericCost(g, game.Player1, &manaCost) {
		t.Fatal("payCost() = true with insufficient mana, want false")
	}
	if forest.Tapped {
		t.Fatal("forest was tapped by failed payment")
	}
	if !g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatalf("mana pool total = %d, want 0", g.Players[game.Player1].ManaPool.Total())
	}
}

func TestPayCostUsesPoolBeforeTappingLands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	manaCost := cost.Mana{cost.G}

	if !payTestGenericCost(g, game.Player1, &manaCost) {
		t.Fatal("payCost() = false, want true")
	}
	if forest.Tapped {
		t.Fatal("forest was tapped even though pool could pay")
	}
}

func TestManaPoolsEmptyAfterMainPhase(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	engine := NewEngine(nil)

	engine.runMainPhase(g, [game.NumPlayers]PlayerAgent{}, game.PhasePrecombatMain, &TurnLog{})

	if !g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatalf("mana pool total = %d, want 0", g.Players[game.Player1].ManaPool.Total())
	}
}

func TestVariableCostUsesChosenX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	manaCost := cost.Mana{cost.X}

	if !canPayCostWithX(g, game.Player1, &manaCost, 2) {
		t.Fatal("canPayCostWithX(X=2) = false, want true")
	}
	if canPayCostWithX(g, game.Player1, &manaCost, 3) {
		t.Fatal("canPayCostWithX(X=3) = true with two lands, want false")
	}
	if canPayCostWithX(g, game.Player1, &manaCost, -1) {
		t.Fatal("canPayCostWithX(X=-1) = true, want false")
	}
}

func TestVariableCostCanIncludeFixedColoredSymbols(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addBasicLandPermanent(g, game.Player1, types.Island)
	manaCost := cost.Mana{cost.X, cost.G}

	if !canPayCostWithX(g, game.Player1, &manaCost, 2) {
		t.Fatal("canPayCostWithX(X=2) = false for {X}{G} with three lands, want true")
	}
}

func TestSpellCostReductionIncreaseAndMinimumGeneric(t *testing.T) {
	t.Run("reduction", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addBasicLandPermanent(g, game.Player1, types.Forest)
		addBasicLandPermanent(g, game.Player1, types.Mountain)
		manaCost := cost.Mana{cost.O(3), cost.G}
		card := &game.CardDef{CardFace: game.CardFace{Name: "Reduced Creature", Types: []types.Card{types.Creature}, ManaCost: opt.Val(manaCost)}}
		g.CostModifiers = append(g.CostModifiers, game.CostModifier{
			Kind:             game.CostModifierSpell,
			GenericReduction: 2,
		})

		if !canPayTestSpellCosts(g, testSpellPaymentRequest{playerID: game.Player1, card: card, sourceZone: zone.Hand}) {
			t.Fatal("canPaySpellCosts() = false, want reduction to make {3}{G} payable with two lands")
		}
	})
	t.Run("increase", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addBasicLandPermanent(g, game.Player1, types.Forest)
		manaCost := cost.Mana{cost.G}
		card := &game.CardDef{CardFace: game.CardFace{Name: "Taxed Creature", Types: []types.Card{types.Creature}, ManaCost: opt.Val(manaCost)}}
		g.CostModifiers = append(g.CostModifiers, game.CostModifier{
			Kind:            game.CostModifierSpell,
			GenericIncrease: 1,
		})

		if canPayTestSpellCosts(g, testSpellPaymentRequest{playerID: game.Player1, card: card, sourceZone: zone.Hand}) {
			t.Fatal("canPaySpellCosts() = true, want increase to require a second mana")
		}
	})
	t.Run("minimum", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		manaCost := cost.Mana{cost.O(3)}
		card := &game.CardDef{CardFace: game.CardFace{Name: "Minimum Creature", Types: []types.Card{types.Creature}, ManaCost: opt.Val(manaCost)}}
		g.CostModifiers = append(g.CostModifiers, game.CostModifier{
			Kind:             game.CostModifierSpell,
			GenericReduction: 5,
			MinimumGeneric:   1,
		})
		if canPayTestSpellCosts(g, testSpellPaymentRequest{playerID: game.Player1, card: card, sourceZone: zone.Hand}) {
			t.Fatal("canPaySpellCosts() = true without mana, want minimum generic cost")
		}
		addBasicLandPermanent(g, game.Player1, types.Island)
		if !canPayTestSpellCosts(g, testSpellPaymentRequest{playerID: game.Player1, card: card, sourceZone: zone.Hand}) {
			t.Fatal("canPaySpellCosts() = false with one mana, want minimum generic cost payable")
		}
	})
}

func TestStaticRuleEffectModifiesSpellCosts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	manaCost := cost.Mana{cost.G}
	card := &game.CardDef{CardFace: game.CardFace{Name: "Taxed Creature", Types: []types.Card{types.Creature}, ManaCost: opt.Val(manaCost)}}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Spell Tax",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind: game.RuleEffectCostModifier,
				CostModifier: game.CostModifier{
					Kind:            game.CostModifierSpell,
					GenericIncrease: 1,
				},
			}},
		}}},
	})

	if canPayTestSpellCosts(g, testSpellPaymentRequest{playerID: game.Player1, card: card, sourceZone: zone.Hand}) {
		t.Fatal("static spell tax allowed {G} spell with only one mana")
	}
}

func TestHybridCostCanBePaidByEitherColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, types.Plains)
	manaCost := cost.Mana{cost.HybridMana(mana.G, mana.W)}

	if !canPayCost(g, game.Player1, &manaCost) {
		t.Fatal("canPayCost() = false for {G/W} with Plains, want true")
	}
}

func TestMonoHybridCostCanBePaidByColorOrGeneric(t *testing.T) {
	tests := []struct {
		name string
		add  func(*game.Game)
	}{
		{
			name: "colored",
			add:  func(g *game.Game) { addBasicLandPermanent(g, game.Player1, types.Forest) },
		},
		{
			name: "generic",
			add: func(g *game.Game) {
				addBasicLandPermanent(g, game.Player1, types.Mountain)
				addBasicLandPermanent(g, game.Player1, types.Island)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			tt.add(g)
			manaCost := cost.Mana{cost.Twobrid(mana.G)}

			if !canPayCost(g, game.Player1, &manaCost) {
				t.Fatal("canPayCost() = false for mono-hybrid cost, want true")
			}
		})
	}
}

func TestPhyrexianCostCanBePaidWithManaOrLife(t *testing.T) {
	t.Run("mana", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addBasicLandPermanent(g, game.Player1, types.Forest)
		manaCost := cost.Mana{cost.PhyrexianMana(mana.G)}

		if !payTestGenericCost(g, game.Player1, &manaCost) {
			t.Fatal("payCost() = false for phyrexian mana with Forest, want true")
		}
		if got := g.Players[game.Player1].Life; got != 40 {
			t.Fatalf("life = %d, want 40", got)
		}
	})
	t.Run("life", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addBasicLandPermanent(g, game.Player1, types.Forest)
		manaCost := cost.Mana{cost.PhyrexianMana(mana.G)}
		prefs := &payment.Preferences{PhyrexianLifeChoices: []bool{true}}

		if !payTestGenericCostWithPreferences(g, game.Player1, &manaCost, prefs) {
			t.Fatal("payCost() = false for phyrexian mana with life, want true")
		}
		if got := g.Players[game.Player1].Life; got != 38 {
			t.Fatalf("life = %d, want 38", got)
		}
	})
}

func TestSnowCostRequiresSnowMana(t *testing.T) {
	t.Run("non-snow source rejected", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addBasicLandPermanent(g, game.Player1, types.Forest)
		manaCost := cost.Mana{cost.S}

		if canPayCost(g, game.Player1, &manaCost) {
			t.Fatal("canPayCost() = true for {S} with non-snow Forest, want false")
		}
	})
	t.Run("snow source accepted", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addSnowBasicLandPermanent(g, game.Player1, types.Forest)
		manaCost := cost.Mana{cost.S}

		if !payTestGenericCost(g, game.Player1, &manaCost) {
			t.Fatal("payCost() = false for {S} with snow Forest, want true")
		}
	})
}

func TestColoredSymbolDoesNotUseSnowSourceNeededLater(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addSnowBasicLandPermanent(g, game.Player1, types.Plains)
	addBasicLandPermanent(g, game.Player1, types.Plains)
	manaCost := cost.Mana{cost.W, cost.S}

	if !canPayCost(g, game.Player1, &manaCost) {
		t.Fatal("canPayCost() = false for {W}{S} with snow and non-snow Plains, want true")
	}
}

func TestColoredSymbolDoesNotSpendFloatingSnowNeededLater(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].ManaPool.AddSnow(mana.W, 1)
	addBasicLandPermanent(g, game.Player1, types.Plains)
	manaCost := cost.Mana{cost.W, cost.S}

	if !canPayCost(g, game.Player1, &manaCost) {
		t.Fatal("canPayCost() = false for {W}{S} with floating snow and non-snow Plains, want true")
	}
}

func TestManaAbilityActionResolvesImmediatelyWithoutStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	rock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Sol Ring",
		Types: []types.Card{types.Artifact}},
	}, mana.C, 2)

	legal := engine.legalActions(g, game.Player1)
	want := action.ActivateAbility(rock.ObjectID, 0, nil, 0)
	if !containsAction(legal, want) {
		t.Fatalf("legal actions do not contain mana ability activation %+v: %+v", want, legal)
	}
	if !engine.applyAction(g, game.Player1, want) {
		t.Fatal("applyAction(mana ability) = false, want true")
	}
	if !rock.Tapped {
		t.Fatal("mana rock was not tapped")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 2 {
		t.Fatalf("colorless mana = %d, want 2", got)
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0 for mana ability", got)
	}
}

func TestSnowManaAbilityAddsSnowMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	snowRock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "types.Snow Manalith",
		Supertypes: []types.Super{types.Snow},
		Types:      []types.Card{types.Artifact}},
	}, mana.G, 1)
	want := action.ActivateAbility(snowRock.ObjectID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, want) {
		t.Fatal("applyAction(snow mana ability) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("green mana = %d, want 1", got)
	}
	if got := g.Players[game.Player1].ManaPool.SnowAmount(); got != 1 {
		t.Fatalf("snow mana = %d, want 1", got)
	}
}

func TestCreatureTapManaAbilityRespectsSummoningSickness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	dork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Elvish Mystic",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	}, mana.G, 1)
	want := action.ActivateAbility(dork.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), want) {
		t.Fatal("summoning-sick creature mana ability was legal")
	}
	dork.SummoningSick = false
	if !containsAction(engine.legalActions(g, game.Player1), want) {
		t.Fatal("non-summoning-sick creature mana ability was not legal")
	}
}

func TestApplyManaAbilityRequiresPriority(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	rock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Sol Ring",
		Types: []types.Card{types.Artifact}},
	}, mana.C, 2)
	g.Turn.PriorityPlayer = game.Player2

	if engine.applyAction(g, game.Player1, action.ActivateAbility(rock.ObjectID, 0, nil, 0)) {
		t.Fatal("applyAction(mana ability without priority) = true, want false")
	}
	if rock.Tapped {
		t.Fatal("mana rock was tapped by activation without priority")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana pool total = %d, want 0", got)
	}
}

func TestPayCostAutoActivatesManaRock(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	rock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Sol Ring",
		Types: []types.Card{types.Artifact}},
	}, mana.C, 2)
	manaCost := cost.Mana{cost.O(2)}

	if !payTestGenericCost(g, game.Player1, &manaCost) {
		t.Fatal("payCost() = false, want true")
	}
	if !rock.Tapped {
		t.Fatal("mana rock was not tapped for payment")
	}
	if !g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatalf("mana pool total = %d, want 0 after exact payment", g.Players[game.Player1].ManaPool.Total())
	}
}

func TestPayCostAutoActivatesMultiOutputSourceForRequiredColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	dork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Llanowar Tribe",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3})},
	}, mana.G, 3)
	dork.SummoningSick = false
	manaCost := cost.Mana{cost.G, cost.G}

	if !payTestGenericCost(g, game.Player1, &manaCost) {
		t.Fatal("payCost() = false, want true")
	}
	if !dork.Tapped {
		t.Fatal("multi-output mana dork was not tapped for payment")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("floating green mana = %d, want 1", got)
	}
}

func TestPayCostAutoActivatesMultiOutputSourceForColorlessSymbols(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	rock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Sol Ring",
		Types: []types.Card{types.Artifact}},
	}, mana.C, 2)
	manaCost := cost.Mana{cost.C, cost.C}

	if !payTestGenericCost(g, game.Player1, &manaCost) {
		t.Fatal("payCost() = false, want true")
	}
	if !rock.Tapped {
		t.Fatal("multi-output mana rock was not tapped for payment")
	}
	if !g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatalf("mana pool total = %d, want 0 after exact colorless payment", g.Players[game.Player1].ManaPool.Total())
	}
}

func TestPayCostAutoActivatesNonSummoningSickManaDork(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	dork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Elvish Mystic",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	}, mana.G, 1)
	manaCost := cost.Mana{cost.G}

	if canPayCost(g, game.Player1, &manaCost) {
		t.Fatal("canPayCost() = true with summoning-sick mana dork, want false")
	}
	dork.SummoningSick = false
	if !payTestGenericCost(g, game.Player1, &manaCost) {
		t.Fatal("payCost() = false with ready mana dork, want true")
	}
	if !dork.Tapped {
		t.Fatal("mana dork was not tapped for payment")
	}
}

func TestPayCostAutoActivatesUntapManaAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Untap Mana Engine",
		Types: []types.Card{types.Artifact},
	}}, mana.G, 1)
	source.Tapped = true
	card, ok := permanentCardDef(g, source)
	if !ok {
		t.Fatal("mana source card definition not found")
	}
	card.ManaAbilities[0].AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalUntap, Text: "{Q}"}}
	manaCost := cost.Mana{cost.G}

	if !payTestGenericCost(g, game.Player1, &manaCost) {
		t.Fatal("payCost() = false with untap mana ability, want true")
	}
	if source.Tapped {
		t.Fatal("mana source remained tapped after automatic untap activation")
	}
}

func TestPayCostChoosesManaAbilityMatchingTapState(t *testing.T) {
	for _, test := range []struct {
		name       string
		tapped     bool
		firstCost  cost.Additional
		secondCost cost.Additional
	}{
		{
			name:       "untapped source skips untap ability",
			firstCost:  cost.Additional{Kind: cost.AdditionalUntap, Text: "{Q}"},
			secondCost: cost.T,
		},
		{
			name:       "tapped source skips tap ability",
			tapped:     true,
			firstCost:  cost.T,
			secondCost: cost.Additional{Kind: cost.AdditionalUntap, Text: "{Q}"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name:  "Modal Mana Engine",
				Types: []types.Card{types.Artifact},
			}}, mana.G, 1)
			card, ok := permanentCardDef(g, source)
			if !ok {
				t.Fatal("mana source card definition not found")
			}
			first := card.ManaAbilities[0]
			first.AdditionalCosts = []cost.Additional{test.firstCost}
			second := first
			second.AdditionalCosts = []cost.Additional{test.secondCost}
			card.ManaAbilities = []game.ManaAbility{first, second}
			source.Tapped = test.tapped
			manaCost := cost.Mana{cost.G}

			if !payTestGenericCost(g, game.Player1, &manaCost) {
				t.Fatal("payCost() = false with a mana ability matching source tap state")
			}
			if source.Tapped == test.tapped {
				t.Fatal("automatic mana ability did not change source tap state")
			}
		})
	}
}

func TestPayCostRespectsManaAbilityTimingAndUsage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Restricted Mana Engine",
		Types: []types.Card{types.Artifact},
	}}, mana.G, 1)
	card, ok := permanentCardDef(g, source)
	if !ok {
		t.Fatal("mana source card definition not found")
	}
	card.ManaAbilities[0].Timing = game.SorceryOncePerTurn
	manaCost := cost.Mana{cost.G}
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep

	if canPayCost(g, game.Player1, &manaCost) {
		t.Fatal("canPayCost() = true outside restricted mana ability timing")
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if !payTestGenericCost(g, game.Player1, &manaCost) {
		t.Fatal("payCost() = false during restricted mana ability timing")
	}
	source.Tapped = false
	if canPayCost(g, game.Player1, &manaCost) {
		t.Fatal("canPayCost() = true after once-per-turn mana ability was used")
	}
}

// addComplexManaAbilityPermanent places a permanent with a single mana ability
// built from the given body onto the battlefield for controller.
func addComplexManaAbilityPermanent(g *game.Game, controller game.PlayerID, def *game.CardDef, body *game.ManaAbility) *game.Permanent {
	def.ManaAbilities = append(def.ManaAbilities, *body)
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   def,
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

// TestComplexManaAbilityTapPayLifeResolvesImmediately verifies that a
// "{T}, Pay 1 life: Add {U} or {R}." mana ability resolves without a stack
// object, taps the source, deducts life, and adds the chosen mana.
func TestComplexManaAbilityTapPayLifeResolvesImmediately(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := game.ManaAbility{
		Text: "{T}, Pay 1 life: Add {U} or {R}.",
		AdditionalCosts: []cost.Additional{
			cost.T,
			{Kind: cost.AdditionalPayLife, Text: "Pay 1 life", Amount: 1},
		},
		Content: game.TapManaChoiceAbility(mana.U, mana.R).Content,
	}
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Pain Land", Types: []types.Card{types.Land}}},
		&body,
	)
	startLife := g.Players[game.Player1].Life
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	// Illegal while tapped.
	source.Tapped = true
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("tap+pay-life mana ability was legal while source was tapped")
	}
	source.Tapped = false

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(tap+pay-life mana ability) = false, want true")
	}
	if !source.Tapped {
		t.Fatal("source was not tapped after tap+pay-life activation")
	}
	if got := g.Players[game.Player1].Life; got != startLife-1 {
		t.Fatalf("life = %d, want %d after paying 1 life", got, startLife-1)
	}
	totalMana := g.Players[game.Player1].ManaPool.Amount(mana.U) + g.Players[game.Player1].ManaPool.Amount(mana.R)
	if totalMana != 1 {
		t.Fatalf("mana pool (U+R) = %d, want 1", totalMana)
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0 for mana ability", got)
	}
}

// TestComplexManaAbilityTapPayLifeFailsWithInsufficientLife verifies that the
// ability cannot be activated when the player has no life to pay, and leaves
// permanent state, life total, and mana pool unchanged.
func TestComplexManaAbilityTapPayLifeFailsWithInsufficientLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := game.ManaAbility{
		Text: "{T}, Pay 1 life: Add {U} or {R}.",
		AdditionalCosts: []cost.Additional{
			cost.T,
			{Kind: cost.AdditionalPayLife, Text: "Pay 1 life", Amount: 1},
		},
		Content: game.TapManaChoiceAbility(mana.U, mana.R).Content,
	}
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Pain Land", Types: []types.Card{types.Land}}},
		&body,
	)
	g.Players[game.Player1].Life = 0
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("tap+pay-life mana ability was legal with 0 life")
	}
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(tap+pay-life with 0 life) = true, want false")
	}
	if source.Tapped {
		t.Fatal("source was tapped by failed payment")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("mana pool = %d, want 0 after failed payment", g.Players[game.Player1].ManaPool.Total())
	}
}

// TestComplexManaAbilityManaCostAndTapAddsMultiSymbolOutput verifies that a
// "{1}, {T}: Add {G}{W}." mana ability consumes the mana, taps the source, and
// adds both mana symbols without creating a stack object.
func TestComplexManaAbilityManaCostAndTapAddsMultiSymbolOutput(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Provide the {1} mana needed by the activation cost.
	addBasicLandPermanent(g, game.Player1, types.Forest)
	body := game.ManaAbility{
		Text:            "{1}, {T}: Add {G}{W}.",
		ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
		AdditionalCosts: cost.Tap,
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.G}},
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.W}},
		}}.Ability(),
	}
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Signet", Types: []types.Card{types.Artifact}}},
		&body,
	)
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(mana-cost+tap mana ability) = false, want true")
	}
	if !source.Tapped {
		t.Fatal("source was not tapped after {1},{T} activation")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("green mana = %d, want 1", got)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.W); got != 1 {
		t.Fatalf("white mana = %d, want 1", got)
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0", got)
	}
}

// TestComplexManaAbilityManaCostFailsWithoutMana verifies that a mana-cost
// mana ability does not tap the source or add mana when there is no mana
// available to pay the activation cost.
func TestComplexManaAbilityManaCostFailsWithoutMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := game.ManaAbility{
		Text:            "{1}, {T}: Add {G}{W}.",
		ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
		AdditionalCosts: cost.Tap,
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.G}},
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.W}},
		}}.Ability(),
	}
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Signet", Types: []types.Card{types.Artifact}}},
		&body,
	)
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("mana-cost mana ability was legal without available mana")
	}
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(mana-cost ability without mana) = true, want false")
	}
	if source.Tapped {
		t.Fatal("source was tapped by failed payment")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("mana pool = %d, want 0 after failed payment", g.Players[game.Player1].ManaPool.Total())
	}
}

// TestComplexManaAbilitySacrificeSourceAddsManaThenLeavesField verifies that a
// "Sacrifice this creature: Add {C}." mana ability sacrifices the source,
// adds colorless mana, and creates no stack object.
func TestComplexManaAbilitySacrificeSourceAddsManaThenLeavesField(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := game.ManaAbility{
		Text:            "Sacrifice this creature: Add {C}.",
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Text: "Sacrifice this creature", Amount: 1}},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.C}},
		}}.Ability(),
	}
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{
			Name:      "Eldrazi Scion",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		}},
		&body,
	)
	sourceID := source.ObjectID
	sourceCardID := source.CardInstanceID
	act := action.ActivateAbility(sourceID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(sacrifice-source mana ability) = false, want true")
	}
	// Source must have left the battlefield.
	for _, p := range g.Battlefield {
		if p.ObjectID == sourceID {
			t.Fatal("sacrificed source remained on battlefield")
		}
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceCardID) {
		t.Fatal("sacrificed source card not found in graveyard")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 1 {
		t.Fatalf("colorless mana = %d, want 1", got)
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0", got)
	}
}

// TestComplexManaAbilityTypedSacrificeAddsManaThenLeavesField verifies that a
// "Sacrifice a creature: Add {C}{C}." mana ability (Ashnod's Altar shape)
// sacrifices a chosen creature, adds two colorless mana, and creates no stack
// object.
func TestComplexManaAbilityTypedSacrificeAddsManaThenLeavesField(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := game.ManaAbility{
		Text: "Sacrifice a creature: Add {C}{C}.",
		AdditionalCosts: []cost.Additional{{
			Kind:               cost.AdditionalSacrifice,
			Text:               "Sacrifice a creature",
			Amount:             1,
			MatchPermanentType: true,
			PermanentType:      types.Creature,
		}},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.C}},
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.C}},
		}}.Ability(),
	}
	altar := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Ashnod's Altar", Types: []types.Card{types.Artifact}}},
		&body,
	)
	// Add a creature to sacrifice as the cost.
	creature := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{
			Name:      "Fodder",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		}},
		&game.ManaAbility{},
	)
	// Remove the placeholder mana ability added by addComplexManaAbilityPermanent.
	creatureCard, ok := permanentCardDef(g, creature)
	if !ok {
		t.Fatal("creature card def not found")
	}
	creatureCard.ManaAbilities = nil
	creatureCardID := creature.CardInstanceID

	act := action.ActivateAbility(altar.ObjectID, 0, nil, 0)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(typed-sacrifice mana ability) = false, want true")
	}
	// The creature must have left the battlefield.
	for _, p := range g.Battlefield {
		if p.CardInstanceID == creatureCardID {
			t.Fatal("sacrificed creature remained on battlefield")
		}
	}
	if !g.Players[game.Player1].Graveyard.Contains(creatureCardID) {
		t.Fatal("sacrificed creature not found in graveyard")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 2 {
		t.Fatalf("colorless mana = %d, want 2", got)
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0", got)
	}
}

// TestComplexManaAbilityPureManaConversionProducesOutput verifies that a
// "{R}: Add {B}." mana ability (Agent of Stromgald shape) consumes the red
// mana, adds one black mana, and creates no stack object.
func TestComplexManaAbilityPureManaConversionProducesOutput(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Float {R} in the mana pool to pay the activation cost.
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	body := game.ManaAbility{
		Text:     "{R}: Add {B}.",
		ManaCost: opt.Val(cost.Mana{cost.R}),
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.B}},
		}}.Ability(),
	}
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{
			Name:      "Agent of Stromgald",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		}},
		&body,
	)
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(pure-mana conversion) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 0 {
		t.Fatalf("red mana = %d, want 0 (consumed)", got)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.B); got != 1 {
		t.Fatalf("black mana = %d, want 1", got)
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0", got)
	}
}

// TestAutoPaymentDoesNotUseComplexManaAbilities verifies that automatic mana
// payment (used when paying spell costs) never auto-activates mana abilities
// that carry a mana cost, pay-life cost, sacrifice cost, or produce more than
// one symbol of output. These abilities require explicit player action.
func TestAutoPaymentDoesNotUseComplexManaAbilities(t *testing.T) {
	tests := []struct {
		name      string
		body      game.ManaAbility
		wantCost  cost.Mana
		setupPool func(*game.Game)
	}{
		{
			name: "mana-cost conversion",
			body: game.ManaAbility{
				ManaCost:        opt.Val(cost.Mana{cost.R}),
				AdditionalCosts: cost.Tap,
				Content: game.Mode{Sequence: []game.Instruction{
					{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.B}},
				}}.Ability(),
			},
			wantCost: cost.Mana{cost.B},
			setupPool: func(g *game.Game) {
				g.Players[game.Player1].ManaPool.Add(mana.R, 1)
			},
		},
		{
			name: "tap+pay-life",
			body: game.ManaAbility{
				AdditionalCosts: []cost.Additional{
					cost.T,
					{Kind: cost.AdditionalPayLife, Text: "Pay 1 life", Amount: 1},
				},
				Content: game.TapManaChoiceAbility(mana.U, mana.R).Content,
			},
			wantCost: cost.Mana{cost.U},
		},
		{
			name: "sacrifice-source",
			body: game.ManaAbility{
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalSacrificeSource, Text: "Sacrifice this", Amount: 1},
				},
				Content: game.Mode{Sequence: []game.Instruction{
					{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.C}},
				}}.Ability(),
			},
			wantCost: cost.Mana{cost.C},
		},
		{
			name: "multi-symbol tap",
			body: game.ManaAbility{
				AdditionalCosts: cost.Tap,
				Content: game.Mode{Sequence: []game.Instruction{
					{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.G}},
					{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.W}},
				}}.Ability(),
			},
			wantCost: cost.Mana{cost.G},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addComplexManaAbilityPermanent(g, game.Player1,
				&game.CardDef{CardFace: game.CardFace{Name: "Complex Source", Types: []types.Card{types.Artifact}}},
				&test.body,
			)
			if test.setupPool != nil {
				test.setupPool(g)
			}
			if canPayCost(g, game.Player1, &test.wantCost) {
				t.Fatal("canPayCost() = true using complex mana ability, want false (auto-payment must be conservative)")
			}
		})
	}
}

// TestAutoPaymentStillUsesSimpleTapManaAbility confirms that automatic mana
// payment continues to work correctly for ordinary tap sources after the
// complex-ability changes.
func TestAutoPaymentStillUsesSimpleTapManaAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Forest Proxy",
		Types: []types.Card{types.Land},
	}}, mana.G, 1)
	manaCost := cost.Mana{cost.G}
	if !canPayCost(g, game.Player1, &manaCost) {
		t.Fatal("canPayCost() = false for simple tap source, want true")
	}
	if !payTestGenericCost(g, game.Player1, &manaCost) {
		t.Fatal("payCost() = false for simple tap source, want true")
	}
}

func addBasicLandPermanent(g *game.Game, controller game.PlayerID, subtype types.Sub) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: string(subtype),
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{subtype}},
		},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func addSnowBasicLandPermanent(g *game.Game, controller game.PlayerID, subtype types.Sub) *game.Permanent {
	permanent := addBasicLandPermanent(g, controller, subtype)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		panic("snow basic land card instance not found")
	}
	card.Def.Supertypes = append(card.Def.Supertypes, types.Snow)
	return permanent
}

func addManaAbilityPermanent(g *game.Game, controller game.PlayerID, def *game.CardDef, m mana.Color, amount int) *game.Permanent {
	def.ManaAbilities = append(def.ManaAbilities, game.ManaAbility{
		AdditionalCosts: cost.Tap,
		Content: game.Mode{
			Sequence: []game.Instruction{
				{Primitive: game.AddMana{
					ManaColor: m,
					Amount:    game.Fixed(amount),
				}},
			},
		}.Ability(),
	})
	cardID := g.IDGen.Next()
	card := &game.CardInstance{
		ID:    cardID,
		Def:   def,
		Owner: controller,
	}
	g.CardInstances[cardID] = card
	permanent, ok := createCardPermanent(g, card, controller, zone.Stack)
	if !ok {
		panic("mana ability permanent was not created")
	}
	return permanent
}
