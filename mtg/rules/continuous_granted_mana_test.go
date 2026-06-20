package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func addAnyColorLandGrantSource(g *game.Game, controller game.PlayerID) *game.Permanent {
	ability := game.TapAnyColorManaAbility()
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Chromatic Lantern",
		Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Land}},
				),
				AddAbilities: []game.Ability{&ability},
			}},
		}},
	}})
}

func countAnyColorManaAbilities(g *game.Game, permanent *game.Permanent) int {
	count := 0
	for _, ability := range permanentEffectiveAbilities(g, permanent) {
		if body, ok := ability.(*game.ManaAbility); ok && game.IsTapAnyColorManaAbility(body) {
			count++
		}
	}
	return count
}

func anyColorManaAbilityIndex(g *game.Game, permanent *game.Permanent) (int, bool) {
	for index, ability := range permanentEffectiveAbilities(g, permanent) {
		if body, ok := ability.(*game.ManaAbility); ok && game.IsTapAnyColorManaAbility(body) {
			return index, true
		}
	}
	return 0, false
}

func TestGrantedManaAbilityTracksSourceAndLandControllers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addAnyColorLandGrantSource(g, game.Player1)
	owned := addBasicLandWithManaPermanent(g, game.Player1, "Mountain", types.Mountain, mana.R)
	opposing := addBasicLandWithManaPermanent(g, game.Player2, "Island", types.Island, mana.U)
	nonland := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Bear", Types: []types.Card{types.Creature},
	}})

	if got := countAnyColorManaAbilities(g, owned); got != 1 {
		t.Fatalf("controlled land grants = %d, want 1", got)
	}
	if got := countAnyColorManaAbilities(g, opposing); got != 0 {
		t.Fatalf("opposing land grants = %d, want 0", got)
	}
	if got := countAnyColorManaAbilities(g, nonland); got != 0 {
		t.Fatalf("controlled nonland grants = %d, want 0", got)
	}

	owned.Controller = game.Player2
	opposing.Controller = game.Player1
	if got := countAnyColorManaAbilities(g, owned); got != 0 {
		t.Fatalf("land after control loss grants = %d, want 0", got)
	}
	if got := countAnyColorManaAbilities(g, opposing); got != 1 {
		t.Fatalf("land after control gain grants = %d, want 1", got)
	}

	source.Controller = game.Player2
	if got := countAnyColorManaAbilities(g, owned); got != 1 {
		t.Fatalf("Player2 land after source control change grants = %d, want 1", got)
	}
	if got := countAnyColorManaAbilities(g, opposing); got != 0 {
		t.Fatalf("Player1 land after source control change grants = %d, want 0", got)
	}

	g.Battlefield = slices.DeleteFunc(g.Battlefield, func(permanent *game.Permanent) bool {
		return permanent == source
	})
	if got := countAnyColorManaAbilities(g, owned); got != 0 {
		t.Fatalf("land after source left grants = %d, want 0", got)
	}
}

func TestGrantedManaAbilityManualAndAutomaticPayment(t *testing.T) {
	t.Run("manual color choice produces mana", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addAnyColorLandGrantSource(g, game.Player1)
		land := addBasicLandWithManaPermanent(g, game.Player1, "Mountain", types.Mountain, mana.R)
		index, ok := anyColorManaAbilityIndex(g, land)
		if !ok {
			t.Fatal("granted any-color mana ability not found")
		}
		activate := action.ActivateAbility(land.ObjectID, index, nil, 0)
		if !containsAction(engine.legalActions(g, game.Player1), activate) {
			t.Fatal("granted choice mana ability was not exposed for manual activation")
		}
		agents := [game.NumPlayers]PlayerAgent{
			game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
		}
		if !engine.applyActionWithChoices(g, game.Player1, activate, agents, &TurnLog{}) {
			t.Fatal("manual granted mana activation failed")
		}
		if got := g.Players[game.Player1].ManaPool.Amount(mana.U); got != 1 {
			t.Fatalf("chosen blue mana = %d, want 1", got)
		}
	})

	t.Run("automatic colored payment chooses needed color", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addAnyColorLandGrantSource(g, game.Player1)
		land := addBasicLandWithManaPermanent(g, game.Player1, "Mountain", types.Mountain, mana.R)
		blueCost := cost.Mana{cost.U}
		if !payTestGenericCost(g, game.Player1, &blueCost) {
			t.Fatal("automatic payment with granted blue mana failed")
		}
		if !land.Tapped {
			t.Fatal("land was not tapped for automatic colored payment")
		}
		if !g.Players[game.Player1].ManaPool.IsEmpty() {
			t.Fatal("automatic colored payment left mana in pool")
		}
	})

	t.Run("one flexible permanent cannot pay twice", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addAnyColorLandGrantSource(g, game.Player1)
		addBasicLandWithManaPermanent(g, game.Player1, "Mountain", types.Mountain, mana.R)
		twoColorCost := cost.Mana{cost.U, cost.G}
		if canPayCost(g, game.Player1, &twoColorCost) {
			t.Fatal("one granted flexible source paid two colored symbols")
		}
	})

	t.Run("colored payment preserves flexible source", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addAnyColorLandGrantSource(g, game.Player1)
		flexible := addBasicLandWithManaPermanent(g, game.Player1, "Mountain", types.Mountain, mana.R)
		rigid := addComplexManaAbilityPermanent(
			g,
			game.Player1,
			&game.CardDef{CardFace: game.CardFace{Name: "Blue Rock", Types: []types.Card{types.Artifact}}},
			new(game.TapManaAbility(mana.U)),
		)
		cost := cost.Mana{cost.U, cost.R}
		if !payTestGenericCost(g, game.Player1, &cost) {
			t.Fatal("payment greedily consumed flexible source before rigid blue source")
		}
		if !flexible.Tapped || !rigid.Tapped {
			t.Fatalf("payment taps flexible/rigid = %v/%v, want both", flexible.Tapped, rigid.Tapped)
		}
	})
}

func TestGrantedManaAbilityLegalityAndDuplicateActions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addAnyColorLandGrantSource(g, game.Player1)
	addAnyColorLandGrantSource(g, game.Player1)
	land := addBasicLandWithManaPermanent(g, game.Player1, "Forest", types.Forest, mana.G)

	if got := countAnyColorManaAbilities(g, land); got != 2 {
		t.Fatalf("effective duplicate grants = %d, want 2 distinct continuous grants", got)
	}
	manualActions := 0
	for _, candidate := range engine.legalActions(g, game.Player1) {
		payload, ok := candidate.ActivateAbilityPayload()
		if ok && payload.SourceID == land.ObjectID {
			manualActions++
		}
	}
	if manualActions != 1 {
		t.Fatalf("equivalent manual mana actions = %d, want 1", manualActions)
	}

	index, ok := anyColorManaAbilityIndex(g, land)
	if !ok {
		t.Fatal("granted ability not found")
	}
	activate := action.ActivateAbility(land.ObjectID, index, nil, 0)
	land.Tapped = true
	if containsAction(engine.legalActions(g, game.Player1), activate) || engine.applyAction(g, game.Player1, activate) {
		t.Fatal("tapped land activated granted tap ability")
	}

	ability := game.TapAnyColorManaAbility()
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Mana Bear", Types: []types.Card{types.Creature},
		ManaAbilities: []game.ManaAbility{ability},
	}})
	creature.SummoningSick = true
	creatureActivate := action.ActivateAbility(creature.ObjectID, 0, nil, 0)
	if containsAction(engine.legalActions(g, game.Player1), creatureActivate) ||
		engine.applyAction(g, game.Player1, creatureActivate) {
		t.Fatal("summoning-sick creature activated a tap mana ability")
	}
}
