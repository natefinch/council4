package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestCanPayCostWithUntappedForest(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, "Forest")
	cost := mana.Cost{mana.ColoredMana(mana.Green)}

	if !canPayCost(g, game.Player1, &cost) {
		t.Fatal("canPayCost() = false, want true")
	}
}

func TestCanPayCostRejectsWrongBasicLandColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, "Mountain")
	cost := mana.Cost{mana.ColoredMana(mana.Green)}

	if canPayCost(g, game.Player1, &cost) {
		t.Fatal("canPayCost() = true with wrong basic land color, want false")
	}
}

func TestCanPayGenericCostWithAnyBasicLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, "Mountain")
	cost := mana.Cost{mana.GenericMana(1)}

	if !canPayCost(g, game.Player1, &cost) {
		t.Fatal("canPayCost() = false, want true")
	}
}

func TestGenericSymbolsDoNotConsumeManaNeededByColoredSymbols(t *testing.T) {
	t.Run("pool", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Players[game.Player1].ManaPool.Add(mana.White, 1)
		g.Players[game.Player1].ManaPool.Add(mana.Green, 1)
		cost := mana.Cost{mana.GenericMana(1), mana.ColoredMana(mana.White)}

		if !canPayCost(g, game.Player1, &cost) {
			t.Fatal("canPayCost() = false for pool {W,G} paying {1}{W}, want true")
		}
	})
	t.Run("lands", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addBasicLandPermanent(g, game.Player1, "Plains")
		addBasicLandPermanent(g, game.Player1, "Forest")
		cost := mana.Cost{mana.GenericMana(1), mana.ColoredMana(mana.White)}

		if !canPayCost(g, game.Player1, &cost) {
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
				addManaAbilityPermanent(g, game.Player1, &game.CardDef{
					Name:  "Wastes",
					Types: []game.CardType{game.TypeLand},
				}, mana.Colorless, 1)
			},
			want: true,
		},
		{
			name: "colored source",
			add: func(g *game.Game) {
				addBasicLandPermanent(g, game.Player1, "Forest")
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			tt.add(g)
			cost := mana.Cost{mana.ColorlessMana()}

			if got := canPayCost(g, game.Player1, &cost); got != tt.want {
				t.Fatalf("canPayCost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPayCostTapsLandUsed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	cost := mana.Cost{mana.ColoredMana(mana.Green)}

	if !payCost(g, game.Player1, &cost) {
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
	addBasicLandPermanent(g, game.Player1, "Forest")
	cost := mana.Cost{mana.ColoredMana(mana.Green)}

	if !payCost(g, game.Player1, &cost) {
		t.Fatal("first payCost() = false, want true")
	}
	if canPayCost(g, game.Player1, &cost) {
		t.Fatal("canPayCost() = true after land tapped, want false")
	}
}

func TestPayCostFailureDoesNotMutate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	cost := mana.Cost{mana.ColoredMana(mana.Green), mana.ColoredMana(mana.Green)}

	if payCost(g, game.Player1, &cost) {
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
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	g.Players[game.Player1].ManaPool.Add(mana.Green, 1)
	cost := mana.Cost{mana.ColoredMana(mana.Green)}

	if !payCost(g, game.Player1, &cost) {
		t.Fatal("payCost() = false, want true")
	}
	if forest.Tapped {
		t.Fatal("forest was tapped even though pool could pay")
	}
}

func TestManaPoolsEmptyAfterMainPhase(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].ManaPool.Add(mana.Green, 1)
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
	addBasicLandPermanent(g, game.Player1, "Forest")
	addBasicLandPermanent(g, game.Player1, "Mountain")
	cost := mana.Cost{mana.VariableMana()}

	if !canPayCostWithX(g, game.Player1, &cost, 2) {
		t.Fatal("canPayCostWithX(X=2) = false, want true")
	}
	if canPayCostWithX(g, game.Player1, &cost, 3) {
		t.Fatal("canPayCostWithX(X=3) = true with two lands, want false")
	}
	if canPayCostWithX(g, game.Player1, &cost, -1) {
		t.Fatal("canPayCostWithX(X=-1) = true, want false")
	}
}

func TestVariableCostCanIncludeFixedColoredSymbols(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, "Forest")
	addBasicLandPermanent(g, game.Player1, "Mountain")
	addBasicLandPermanent(g, game.Player1, "Island")
	cost := mana.Cost{mana.VariableMana(), mana.ColoredMana(mana.Green)}

	if !canPayCostWithX(g, game.Player1, &cost, 2) {
		t.Fatal("canPayCostWithX(X=2) = false for {X}{G} with three lands, want true")
	}
}

func TestHybridCostCanBePaidByEitherColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, "Plains")
	cost := mana.Cost{mana.HybridMana(mana.Green, mana.White)}

	if !canPayCost(g, game.Player1, &cost) {
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
			add:  func(g *game.Game) { addBasicLandPermanent(g, game.Player1, "Forest") },
		},
		{
			name: "generic",
			add: func(g *game.Game) {
				addBasicLandPermanent(g, game.Player1, "Mountain")
				addBasicLandPermanent(g, game.Player1, "Island")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			tt.add(g)
			cost := mana.Cost{mana.MonoHybridMana(mana.Green)}

			if !canPayCost(g, game.Player1, &cost) {
				t.Fatal("canPayCost() = false for mono-hybrid cost, want true")
			}
		})
	}
}

func TestPhyrexianCostCanBePaidWithManaOrLife(t *testing.T) {
	t.Run("mana", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addBasicLandPermanent(g, game.Player1, "Forest")
		cost := mana.Cost{mana.PhyrexianMana(mana.Green)}

		if !payCost(g, game.Player1, &cost) {
			t.Fatal("payCost() = false for phyrexian mana with Forest, want true")
		}
		if got := g.Players[game.Player1].Life; got != 40 {
			t.Fatalf("life = %d, want 40", got)
		}
	})
	t.Run("life", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		cost := mana.Cost{mana.PhyrexianMana(mana.Green)}

		if !payCost(g, game.Player1, &cost) {
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
		addBasicLandPermanent(g, game.Player1, "Forest")
		cost := mana.Cost{mana.SnowMana()}

		if canPayCost(g, game.Player1, &cost) {
			t.Fatal("canPayCost() = true for {S} with non-snow Forest, want false")
		}
	})
	t.Run("snow source accepted", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addSnowBasicLandPermanent(g, game.Player1, "Forest")
		cost := mana.Cost{mana.SnowMana()}

		if !payCost(g, game.Player1, &cost) {
			t.Fatal("payCost() = false for {S} with snow Forest, want true")
		}
	})
}

func TestColoredSymbolDoesNotUseSnowSourceNeededLater(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addSnowBasicLandPermanent(g, game.Player1, "Plains")
	addBasicLandPermanent(g, game.Player1, "Plains")
	cost := mana.Cost{mana.ColoredMana(mana.White), mana.SnowMana()}

	if !canPayCost(g, game.Player1, &cost) {
		t.Fatal("canPayCost() = false for {W}{S} with snow and non-snow Plains, want true")
	}
}

func TestColoredSymbolDoesNotSpendFloatingSnowNeededLater(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].ManaPool.AddSnow(mana.White, 1)
	addBasicLandPermanent(g, game.Player1, "Plains")
	cost := mana.Cost{mana.ColoredMana(mana.White), mana.SnowMana()}

	if !canPayCost(g, game.Player1, &cost) {
		t.Fatal("canPayCost() = false for {W}{S} with floating snow and non-snow Plains, want true")
	}
}

func TestManaAbilityActionResolvesImmediatelyWithoutStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	rock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{
		Name:  "Sol Ring",
		Types: []game.CardType{game.TypeArtifact},
	}, mana.Colorless, 2)

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
	if got := g.Players[game.Player1].ManaPool.Amount(mana.Colorless); got != 2 {
		t.Fatalf("colorless mana = %d, want 2", got)
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0 for mana ability", got)
	}
}

func TestSnowManaAbilityAddsSnowMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	snowRock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{
		Name:       "Snow Manalith",
		Supertypes: []game.Supertype{game.Snow},
		Types:      []game.CardType{game.TypeArtifact},
	}, mana.Green, 1)
	want := action.ActivateAbility(snowRock.ObjectID, 0, nil, 0)

	if !engine.applyAction(g, game.Player1, want) {
		t.Fatal("applyAction(snow mana ability) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.Green); got != 1 {
		t.Fatalf("green mana = %d, want 1", got)
	}
	if got := g.Players[game.Player1].ManaPool.SnowAmount(); got != 1 {
		t.Fatalf("snow mana = %d, want 1", got)
	}
}

func TestCreatureTapManaAbilityRespectsSummoningSickness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	dork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{
		Name:      "Elvish Mystic",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &game.PT{Value: 1},
		Toughness: &game.PT{Value: 1},
	}, mana.Green, 1)
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
	rock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{
		Name:  "Sol Ring",
		Types: []game.CardType{game.TypeArtifact},
	}, mana.Colorless, 2)
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
	rock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{
		Name:  "Sol Ring",
		Types: []game.CardType{game.TypeArtifact},
	}, mana.Colorless, 2)
	cost := mana.Cost{mana.GenericMana(2)}

	if !payCost(g, game.Player1, &cost) {
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
	dork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{
		Name:      "Llanowar Tribe",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &game.PT{Value: 3},
		Toughness: &game.PT{Value: 3},
	}, mana.Green, 3)
	dork.SummoningSick = false
	cost := mana.Cost{mana.ColoredMana(mana.Green), mana.ColoredMana(mana.Green)}

	if !payCost(g, game.Player1, &cost) {
		t.Fatal("payCost() = false, want true")
	}
	if !dork.Tapped {
		t.Fatal("multi-output mana dork was not tapped for payment")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.Green); got != 1 {
		t.Fatalf("floating green mana = %d, want 1", got)
	}
}

func TestPayCostAutoActivatesMultiOutputSourceForColorlessSymbols(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	rock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{
		Name:  "Sol Ring",
		Types: []game.CardType{game.TypeArtifact},
	}, mana.Colorless, 2)
	cost := mana.Cost{mana.ColorlessMana(), mana.ColorlessMana()}

	if !payCost(g, game.Player1, &cost) {
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
	dork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{
		Name:      "Elvish Mystic",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &game.PT{Value: 1},
		Toughness: &game.PT{Value: 1},
	}, mana.Green, 1)
	cost := mana.Cost{mana.ColoredMana(mana.Green)}

	if canPayCost(g, game.Player1, &cost) {
		t.Fatal("canPayCost() = true with summoning-sick mana dork, want false")
	}
	dork.SummoningSick = false
	if !payCost(g, game.Player1, &cost) {
		t.Fatal("payCost() = false with ready mana dork, want true")
	}
	if !dork.Tapped {
		t.Fatal("mana dork was not tapped for payment")
	}
}

func addBasicLandPermanent(g *game.Game, controller game.PlayerID, subtype string) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{
			Name:     subtype,
			Types:    []game.CardType{game.TypeLand},
			Subtypes: []string{subtype},
		},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       id.ID(g.IDGen.Next()),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func addSnowBasicLandPermanent(g *game.Game, controller game.PlayerID, subtype string) *game.Permanent {
	permanent := addBasicLandPermanent(g, controller, subtype)
	card := g.GetCardInstance(permanent.CardInstanceID)
	card.Def.Supertypes = append(card.Def.Supertypes, game.Snow)
	return permanent
}

func addManaAbilityPermanent(g *game.Game, controller game.PlayerID, def *game.CardDef, color mana.Color, amount int) *game.Permanent {
	def.Abilities = append(def.Abilities, game.AbilityDef{
		Kind:           game.ActivatedAbility,
		AdditionalCost: "{T}",
		IsManaAbility:  true,
		Effects: []game.Effect{
			{
				Type:      game.EffectAddMana,
				ManaColor: color,
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
	return createCardPermanent(g, card, controller, game.ZoneStack)
}
