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
		Abilities: []game.AbilityDef{{
			Kind: game.StaticAbility,
			Effects: []game.Effect{{
				Type: game.EffectApplyRule,
				RuleEffects: []game.RuleEffect{{
					Kind: game.RuleEffectCostModifier,
					CostModifier: game.CostModifier{
						Kind:            game.CostModifierSpell,
						GenericIncrease: 1,
					},
				}},
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
	def.Abilities = append(def.Abilities, game.AbilityDef{
		Kind: game.ActivatedAbility,
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalTap},
		},
		IsManaAbility: true,
		Effects: []game.Effect{
			{
				Type:      game.EffectAddMana,
				ManaColor: m,
				Amount:    amount,
			},
		},
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
