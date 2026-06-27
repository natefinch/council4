package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// sourceSpellReductionCard models a spell that costs PerObjectReduction generic
// less to cast for each battlefield permanent matching selection, encoded as the
// AffectedSource spell cost modifier the cardgen backend emits for the
// "This spell costs {N} less to cast for each <object>" ability.
func sourceSpellReductionCard(name string, manaCost cost.Mana, selection game.Selection, perObject int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(manaCost),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCostModifier,
				AffectedSource: true,
				CostModifier: game.CostModifier{
					Kind:               game.CostModifierSpell,
					PerObjectReduction: perObject,
					CountSelection:     &selection,
				},
			}},
		}},
	}}
}

// sourceSpellGenericReduction sums the generic reductions the rules engine
// resolves for casting card from the player's hand, which for a clean game is
// exactly the source-scoped per-object reduction.
func sourceSpellGenericReduction(g *game.Game, playerID game.PlayerID, card *game.CardDef) int {
	state := &rulesPaymentState{g: g}
	total := 0
	for _, modifier := range state.CostModifiersForSpell(playerID, card, 0, zone.Hand, nil) {
		total += modifier.GenericReduction
	}
	return total
}

func anyCreatureSelection() game.Selection {
	return game.Selection{RequiredTypes: []types.Card{types.Creature}}
}

// sourceSpellZoneReductionCard models a spell that costs perObject generic less
// to cast for each card in the caster's own zone matching selection, encoded as
// the AffectedSource spell cost modifier the cardgen backend emits for the
// "This spell costs {N} less to cast for each <card> in your graveyard/hand"
// ability.
func sourceSpellZoneReductionCard(name string, manaCost cost.Mana, selection game.Selection, perObject int, cardZone zone.Type) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(manaCost),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCostModifier,
				AffectedSource: true,
				CostModifier: game.CostModifier{
					Kind:               game.CostModifierSpell,
					PerObjectReduction: perObject,
					CountSelection:     &selection,
					CountZone:          opt.Val(cardZone),
				},
			}},
		}},
	}}
}

func graveyardCreatureCard(g *game.Game, playerID game.PlayerID) {
	cardID := addCardToHand(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Creature",
		Types: []types.Card{types.Creature},
	}})
	g.Players[playerID].Hand.Remove(cardID)
	g.Players[playerID].Graveyard.Add(cardID)
}

func TestSourceSpellCostReductionCountsGraveyardCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardCreatureCard(g, game.Player1)
	graveyardCreatureCard(g, game.Player1)
	graveyardCreatureCard(g, game.Player2)
	// A battlefield creature must not be counted: only the caster's graveyard.
	addCreaturePermanent(g, game.Player1)
	card := sourceSpellZoneReductionCard("Hollow Marauder", cost.Mana{cost.O(6), cost.B}, anyCreatureSelection(), 1, zone.Graveyard)

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 2 {
		t.Fatalf("reduction for each creature card in your graveyard = %d, want 2", got)
	}
}

func TestSourceSpellCostReductionZeroGraveyardCardsNoReduction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := sourceSpellZoneReductionCard("Hollow Marauder", cost.Mana{cost.O(6), cost.B}, anyCreatureSelection(), 1, zone.Graveyard)

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 0 {
		t.Fatalf("reduction with an empty graveyard = %d, want 0", got)
	}
}

func TestSourceSpellCostReductionZeroCreaturesNoReduction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := sourceSpellReductionCard("Blasphemous Act", cost.Mana{cost.O(8), cost.R}, anyCreatureSelection(), 1)

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 0 {
		t.Fatalf("reduction with no creatures = %d, want 0", got)
	}
}

func TestSourceSpellCostReductionPerCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	card := sourceSpellReductionCard("Blasphemous Act", cost.Mana{cost.O(8), cost.R}, anyCreatureSelection(), 1)

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 3 {
		t.Fatalf("reduction with three battlefield creatures = %d, want 3", got)
	}
}

func TestSourceSpellCostReductionCountsControllerScopedSelection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	selection := anyCreatureSelection()
	selection.Controller = game.ControllerOpponent
	card := sourceSpellReductionCard("Primeval Protector", cost.Mana{cost.O(6), cost.G}, selection, 1)

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 1 {
		t.Fatalf("reduction for each creature opponents control = %d, want 1", got)
	}
}

func TestSourceSpellCostReductionGenericFloorsAtZeroKeepsColored(t *testing.T) {
	makeGame := func(creatures int) (*game.Game, *game.CardDef) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		for range creatures {
			addCreaturePermanent(g, game.Player1)
		}
		addBasicLandPermanent(g, game.Player1, types.Mountain)
		card := sourceSpellReductionCard("Blasphemous Act", cost.Mana{cost.O(2), cost.R}, anyCreatureSelection(), 1)
		return g, card
	}

	t.Run("over-reduction floors generic at zero", func(t *testing.T) {
		g, card := makeGame(5)
		if !canPayTestSpellCosts(g, testSpellPaymentRequest{playerID: game.Player1, card: card, sourceZone: zone.Hand}) {
			t.Fatal("canPaySpellCosts() = false; five creatures should floor {2} to zero leaving {R} payable by one Mountain")
		}
	})

	t.Run("colored requirement preserved", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		for range 5 {
			addCreaturePermanent(g, game.Player1)
		}
		addBasicLandPermanent(g, game.Player1, types.Forest)
		card := sourceSpellReductionCard("Blasphemous Act", cost.Mana{cost.O(2), cost.R}, anyCreatureSelection(), 1)
		if canPayTestSpellCosts(g, testSpellPaymentRequest{playerID: game.Player1, card: card, sourceZone: zone.Hand}) {
			t.Fatal("canPaySpellCosts() = true; the {R} requirement must survive the generic reduction and a Forest cannot pay it")
		}
	})

	t.Run("no reduction below the printed cost without creatures", func(t *testing.T) {
		g, card := makeGame(0)
		if canPayTestSpellCosts(g, testSpellPaymentRequest{playerID: game.Player1, card: card, sourceZone: zone.Hand}) {
			t.Fatal("canPaySpellCosts() = true; {2}{R} needs three mana when no creatures reduce it")
		}
	})
}

func TestSourceSpellCostReductionAppliesOnlyToSourceSpell(t *testing.T) {

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player1)
	// A permanent on the battlefield that carries the self-scoped reduction static
	// ability must not reduce the cost of other spells the controller casts.
	addCombatPermanent(g, game.Player1, sourceSpellReductionCard(
		"Primeval Protector", cost.Mana{cost.O(6), cost.G}, anyCreatureSelection(), 1))

	other := &game.CardDef{CardFace: game.CardFace{
		Name:     "Unrelated Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.R}),
	}}

	if got := sourceSpellGenericReduction(g, game.Player1, other); got != 0 {
		t.Fatalf("a battlefield self-reduction leaked %d generic onto an unrelated spell, want 0", got)
	}
}

// sourceSpellDynamicReductionCard models a spell whose cast cost is reduced by a
// dynamic amount, encoded as the AffectedSource spell cost modifier the cardgen
// backend emits for "This spell costs {X} less to cast, where X is <dynamic
// amount>" (The Great Henge: the greatest power among creatures you control).
func sourceSpellDynamicReductionCard(name string, manaCost cost.Mana, dynamic *game.DynamicAmount) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(manaCost),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCostModifier,
				AffectedSource: true,
				CostModifier: game.CostModifier{
					Kind:             game.CostModifierSpell,
					DynamicReduction: dynamic,
				},
			}},
		}},
	}}
}

// addCreatureWithPower puts a vanilla creature with the given printed power onto
// the battlefield under controller, used to drive greatest-power dynamic amounts.
func addCreatureWithPower(g *game.Game, controller game.PlayerID, power int) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:      "Test Creature",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: power}),
			Toughness: opt.Val(game.PT{Value: power}),
		}},
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

func greatestPowerYouControlAmount() *game.DynamicAmount {
	return &game.DynamicAmount{
		Kind:  game.DynamicAmountGreatestPowerInGroup,
		Group: game.BattlefieldGroup(anyCreatureSelection()),
	}
}

func TestSourceSpellDynamicCostReductionGreatestPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCreatureWithPower(g, game.Player1, 2)
	addCreatureWithPower(g, game.Player1, 5)
	addCreatureWithPower(g, game.Player1, 3)
	card := sourceSpellDynamicReductionCard("The Great Henge", cost.Mana{cost.O(7), cost.G, cost.G}, greatestPowerYouControlAmount())

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 5 {
		t.Fatalf("dynamic reduction = %d, want 5 (greatest power among controlled creatures)", got)
	}
}

func TestSourceSpellDynamicCostReductionNoCreaturesNoReduction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := sourceSpellDynamicReductionCard("The Great Henge", cost.Mana{cost.O(7), cost.G, cost.G}, greatestPowerYouControlAmount())

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 0 {
		t.Fatalf("dynamic reduction with no creatures = %d, want 0", got)
	}
}

// addArtifactWithManaValue puts a vanilla artifact with the given printed mana
// value onto the battlefield under controller, used to drive total-mana-value
// dynamic amounts.
func addArtifactWithManaValue(g *game.Game, controller game.PlayerID, manaValue int) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:     "Test Artifact",
			Types:    []types.Card{types.Artifact},
			ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
		}},
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

func totalManaValueArtifactsYouControlAmount() *game.DynamicAmount {
	return &game.DynamicAmount{
		Kind:  game.DynamicAmountTotalManaValueInGroup,
		Group: game.BattlefieldGroup(artifactsYouControlSelection()),
	}
}

func TestSourceSpellDynamicCostReductionTotalManaValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addArtifactWithManaValue(g, game.Player1, 2)
	addArtifactWithManaValue(g, game.Player1, 5)
	addArtifactWithManaValue(g, game.Player2, 4)
	card := sourceSpellDynamicReductionCard("Metalwork Colossus", cost.Mana{cost.O(11)}, totalManaValueArtifactsYouControlAmount())

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 7 {
		t.Fatalf("dynamic reduction = %d, want 7 (total mana value of controlled artifacts)", got)
	}
}

func artifactsYouControlSelection() game.Selection {
	return game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou}
}

// TestSourceSpellCostReductionAffinityForArtifacts exercises the cost modifier
// that "Affinity for artifacts" lowers to: the spell costs {1} less to cast for
// each artifact its caster controls, counting only the caster's artifacts.
func TestSourceSpellCostReductionAffinityForArtifacts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addArtifactPermanent(g, game.Player1)
	addArtifactPermanent(g, game.Player1)
	addArtifactPermanent(g, game.Player1)
	addArtifactPermanent(g, game.Player2)
	addCreaturePermanent(g, game.Player1)
	card := sourceSpellReductionCard("Thought Monitor", cost.Mana{cost.O(5), cost.U, cost.U}, artifactsYouControlSelection(), 1)

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 3 {
		t.Fatalf("Affinity reduction with three controlled artifacts = %d, want 3", got)
	}
}

// sourceSpellConditionalReductionCard models a spell that costs GenericReduction
// generic less to cast when the caster satisfies condition, encoded as the
// AffectedSource spell cost modifier the cardgen backend emits for the
// "This spell costs {N} less to cast if <condition>" ability (Wizard's Lightning,
// Squash, Draconic Lore).
func sourceSpellConditionalReductionCard(name string, manaCost cost.Mana, condition game.Condition, generic int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(manaCost),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCostModifier,
				AffectedSource: true,
				CostModifier: game.CostModifier{
					Kind:               game.CostModifierSpell,
					GenericReduction:   generic,
					ReductionCondition: opt.Val(condition),
				},
			}},
		}},
	}}
}

func controlsWizardCondition() game.Condition {
	return game.Condition{
		ControlsMatching: opt.Val(game.SelectionCount{
			Selection: game.Selection{SubtypesAny: []types.Sub{types.Wizard}},
		}),
	}
}

func addWizardPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:     "Test Wizard",
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Wizard},
		}},
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

func TestSourceSpellCostReductionConditionalAppliesWhenSatisfied(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addWizardPermanent(g, game.Player1)
	card := sourceSpellConditionalReductionCard("Wizard's Lightning", cost.Mana{cost.O(2), cost.R}, controlsWizardCondition(), 2)

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 2 {
		t.Fatalf("reduction while controlling a Wizard = %d, want 2", got)
	}
}

func TestSourceSpellCostReductionConditionalNoReductionWhenUnsatisfied(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// A Wizard controlled by an opponent must not satisfy "you control a Wizard".
	addWizardPermanent(g, game.Player2)
	card := sourceSpellConditionalReductionCard("Wizard's Lightning", cost.Mana{cost.O(2), cost.R}, controlsWizardCondition(), 2)

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 0 {
		t.Fatalf("reduction without controlling a Wizard = %d, want 0", got)
	}
}
