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

func TestPaymentOnlyManaAbilityResolvesImmediatelyWithoutStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	rock := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Sol Ring",
		Types: []types.Card{types.Artifact}},
	}, mana.C, 2)

	legal := engine.legalActions(g, game.Player1)
	want := action.ActivateAbility(rock.ObjectID, 0, nil, 0)
	if containsAction(legal, want) {
		t.Fatalf("legal actions exposed payment-only mana ability activation %+v: %+v", want, legal)
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

	if engine.applyAction(g, game.Player1, want) {
		t.Fatal("summoning-sick creature mana ability was activatable")
	}
	dork.SummoningSick = false
	if containsAction(engine.legalActions(g, game.Player1), want) {
		t.Fatal("payment-only creature mana ability was exposed as a standalone action")
	}
	if !engine.applyAction(g, game.Player1, want) {
		t.Fatal("non-summoning-sick creature mana ability was not activatable")
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
