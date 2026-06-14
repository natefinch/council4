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
	assertEvent(t, g.Events, game.EventPermanentSacrificed, func(event game.Event) bool {
		return event.Player == game.Player1 && event.PermanentID == sourceID
	})
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
