package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// These scenarios prove "you may spend mana as though it were mana of any
// color/type to cast that spell" is a spending permission on the total cost
// the player actually chose (CR 609.4b, CR 118.14), not a rewrite of the mana
// cost into an alternative cost of its own: every cost choice the spell
// otherwise has (overload, kicker, evoke, additional costs, Mutate, and a
// "without paying its mana cost" permission) stays available and can be paid
// with off-color mana, and "any color" never lets colored mana pay {C}
// (CR 107.4c) while "any type" does.

// riderExileCard puts def into Player2's exile with a per-card permission for
// Player1 to cast it (Grenzo, Havoc Raiser and Court of Locthwain grant this
// shape) carrying spendAs, and gives Player1 sorcery-speed priority.
func riderExileCard(g *game.Game, def *game.CardDef, spendAs game.ManaSpendPermission) id.ID {
	cardID := addCardToExile(g, game.Player2, def)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectPlayFromZone,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		Duration:       game.DurationPermanent,
		CastFromZone:   zone.Exile,
		AffectedCardID: cardID,
		ExpiresFor:     game.Player1,
		SpendAnyMana:   spendAs,
	})
	giveMainPhasePriority(g, game.Player1)
	return cardID
}

func assertExileCastListed(t *testing.T, g *game.Game, engine *Engine, act action.Action, want bool) {
	t.Helper()
	if got := actionsContain(engine.legalActions(g, game.Player1), act); got != want {
		t.Fatalf("legal actions contain %v = %v, want %v", act, got, want)
	}
}

func TestSpendAnyColorPaysOverloadFromExile(t *testing.T) {
	t.Parallel()
	setup := func(spendAs game.ManaSpendPermission) (*game.Game, *Engine, id.ID, *game.Permanent) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		artifact := addOverloadArtifact(g, game.Player2, false)
		cardID := riderExileCard(g, vandalblastDefinition(), spendAs)
		g.Players[game.Player1].ManaPool.Add(mana.U, 5)
		return g, NewEngine(nil), cardID, artifact
	}
	overloadFrom := func(cardID id.ID) action.Action {
		return action.CastOverloadedSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil)
	}

	g, engine, cardID, artifact := setup(game.ManaSpendAsAnyColor)
	overload := overloadFrom(cardID)
	assertExileCastListed(t, g, engine, overload, true)
	if !engine.applyAction(g, game.Player1, overload) {
		t.Fatal("five blue mana could not pay overload {4}{R} under the any-color permission")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || !obj.Overloaded || obj.SourceZone != zone.Exile {
		t.Fatalf("stack object = %+v, want the overloaded spell cast from exile", obj)
	}
	// CR 609.4b: the permission changes only which mana may pay the cost; the
	// mana actually spent is still blue.
	if obj.ColorsOfManaSpentToCast != 1 || obj.ManaSpentByColorToCast[color.Blue] != 5 ||
		obj.ManaSpentByColorToCast[color.Red] != 0 {
		t.Fatalf("spent colors = %d %v, want five blue", obj.ColorsOfManaSpentToCast, obj.ManaSpentByColorToCast)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, artifact.ObjectID); ok {
		t.Fatal("overloaded spell did not destroy the opponent's artifact")
	}

	g, engine, cardID, _ = setup(game.ManaSpendNormal)
	overload = overloadFrom(cardID)
	assertExileCastListed(t, g, engine, overload, false)
	if engine.applyAction(g, game.Player1, overload) {
		t.Fatal("blue mana paid {4}{R} without the spending permission")
	}
	if !g.Players[game.Player2].Exile.Contains(cardID) {
		t.Fatal("rejected overload cast did not return the card to exile")
	}
}

// kickerExileFixture is a {G} sorcery with kicker {1}{G} and an additional
// "pay 2 life" cost, so one cast combines the mana cost, an optional
// additional mana cost, and a mandatory non-mana additional cost.
func kickerExileFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Kicked Exile Probe",
		ManaCost: opt.Val(cost.Mana{cost.G}),
		Types:    []types.Card{types.Sorcery},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.O(1), cost.G}}},
		}},
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2}},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability()),
	}}
}

func TestSpendAnyColorPaysKickerAndAdditionalCostsFromExile(t *testing.T) {
	t.Parallel()
	setup := func(blue int) (*game.Game, *Engine, id.ID) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		cardID := riderExileCard(g, kickerExileFixture(), game.ManaSpendAsAnyColor)
		g.Players[game.Player1].ManaPool.Add(mana.U, blue)
		return g, NewEngine(nil), cardID
	}
	kicked := func(cardID id.ID) action.Action {
		return action.CastKickedSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	}
	unkicked := func(cardID id.ID) action.Action {
		return action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	}

	t.Run("kicked", func(t *testing.T) {
		g, engine, cardID := setup(3)
		life := g.Players[game.Player1].Life
		assertExileCastListed(t, g, engine, kicked(cardID), true)
		if !engine.applyAction(g, game.Player1, kicked(cardID)) {
			t.Fatal("three blue mana could not pay {G} plus kicker {1}{G}")
		}
		obj, ok := g.Stack.Peek()
		if !ok || obj.SourceID != cardID || !obj.KickerPaid || obj.ManaSpentToCast != 3 {
			t.Fatalf("stack object = %+v, want the kicked spell paid with three mana", obj)
		}
		if got := g.Players[game.Player1].Life; got != life-2 {
			t.Fatalf("life = %d, want the additional 2 life paid (%d)", got, life-2)
		}
	})
	t.Run("declined kicker", func(t *testing.T) {
		g, engine, cardID := setup(1)
		assertExileCastListed(t, g, engine, kicked(cardID), false)
		assertExileCastListed(t, g, engine, unkicked(cardID), true)
		if !engine.applyAction(g, game.Player1, unkicked(cardID)) {
			t.Fatal("one blue mana could not pay the unkicked {G}")
		}
		if obj, ok := g.Stack.Peek(); !ok || obj.KickerPaid {
			t.Fatalf("stack object = %+v, want the unkicked spell", obj)
		}
	})
	t.Run("unaffordable kicker", func(t *testing.T) {
		g, engine, cardID := setup(2)
		if engine.applyAction(g, game.Player1, kicked(cardID)) {
			t.Fatal("two mana paid a three-mana kicked total cost")
		}
		if !g.Players[game.Player2].Exile.Contains(cardID) || g.Players[game.Player1].ManaPool.Total() != 2 {
			t.Fatal("rejected kicked cast was not undone")
		}
	})
}

func TestSpendAnyColorKeepsEvokeAvailable(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	pt := game.PT{Value: 2}
	cardID := riderExileCard(g, &game.CardDef{CardFace: game.CardFace{
		Name:      "Evoke Exile Probe",
		Types:     []types.Card{types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.W}),
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		AlternativeCosts: []cost.Alternative{{
			Label:    "Evoke",
			Mechanic: cost.AlternativeMechanicEvoke,
			ManaCost: opt.Val(cost.Mana{cost.W}),
		}},
		TriggeredAbilities: []game.TriggeredAbility{game.EvokeSacrificeTriggeredAbility()},
	}}, game.ManaSpendAsAnyColor)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	act := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	assertExileCastListed(t, g, engine, act, true)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: evokeCostAgent{costLabel: "Evoke"}}
	if !engine.applyActionWithChoices(g, game.Player1, act, agents, &TurnLog{}) {
		t.Fatal("one blue mana could not pay evoke {W} under the any-color permission")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != cardID || !obj.Evoked {
		t.Fatalf("stack object = %+v, want the evoked spell", obj)
	}
}

func TestSpendAnyColorDoesNotPayColorlessRequirement(t *testing.T) {
	t.Parallel()
	reshaper := func() *game.CardDef {
		pt := game.PT{Value: 3}
		return &game.CardDef{CardFace: game.CardFace{
			Name:      "Colorless Requirement Probe",
			Types:     []types.Card{types.Creature},
			ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.C}),
			Power:     opt.Val(pt),
			Toughness: opt.Val(pt),
		}}
	}
	perCard := func(g *game.Game, cardID id.ID, spendAs game.ManaSpendPermission) {
		g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
			ID:             g.IDGen.Next(),
			Kind:           game.RuleEffectPlayFromZone,
			Controller:     game.Player1,
			AffectedPlayer: game.PlayerYou,
			Duration:       game.DurationPermanent,
			CastFromZone:   zone.Exile,
			AffectedCardID: cardID,
			ExpiresFor:     game.Player1,
			SpendAnyMana:   spendAs,
		})
	}
	cases := []struct {
		name    string
		spendAs game.ManaSpendPermission
		pool    map[mana.Color]int
		want    bool
	}{
		// CR 107.4c: {C} can be paid only with colorless mana, and treating mana
		// as any color does not make it colorless.
		{"any color with only blue", game.ManaSpendAsAnyColor, map[mana.Color]int{mana.U: 3}, false},
		{"any color with colorless", game.ManaSpendAsAnyColor, map[mana.Color]int{mana.U: 2, mana.C: 1}, true},
		// CR 118.14: mana of any type includes colorless.
		{"any type with only blue", game.ManaSpendAsAnyType, map[mana.Color]int{mana.U: 3}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToExile(g, game.Player1, reshaper())
			perCard(g, cardID, tc.spendAs)
			giveMainPhasePriority(g, game.Player1)
			for c, n := range tc.pool {
				g.Players[game.Player1].ManaPool.Add(c, n)
			}
			act := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
			assertExileCastListed(t, g, engine, act, tc.want)
			if got := engine.applyAction(g, game.Player1, act); got != tc.want {
				t.Fatalf("cast = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSpendAnyColorLetsColorlessManaPayColoredRequirement(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	artifact := addOverloadArtifact(g, game.Player2, false)
	cardID := riderExileCard(g, vandalblastDefinition(), game.ManaSpendAsAnyColor)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	act := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront,
		[]game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil)
	assertExileCastListed(t, g, engine, act, true)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("colorless mana could not pay {R} under the any-color permission")
	}
}

func TestSpendAnyColorPaysKickerOnFreeExileCast(t *testing.T) {
	t.Parallel()
	setup := func(spendAs game.ManaSpendPermission) (*game.Game, *Engine, id.ID) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		cardID := addCardToExile(g, game.Player1, kickerExileFixture())
		g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
			ID:                    g.IDGen.Next(),
			Kind:                  game.RuleEffectPlayFromZone,
			Controller:            game.Player1,
			AffectedPlayer:        game.PlayerYou,
			Duration:              game.DurationPermanent,
			CastFromZone:          zone.Exile,
			AffectedCardID:        cardID,
			ExpiresFor:            game.Player1,
			WithoutPayingManaCost: true,
			SpendAnyMana:          spendAs,
		})
		giveMainPhasePriority(g, game.Player1)
		g.Players[game.Player1].ManaPool.Add(mana.U, 2)
		return g, NewEngine(nil), cardID
	}
	kicked := func(cardID id.ID) action.Action {
		return action.CastKickedSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	}

	g, engine, cardID := setup(game.ManaSpendAsAnyColor)
	assertExileCastListed(t, g, engine, kicked(cardID), true)
	if !engine.applyAction(g, game.Player1, kicked(cardID)) {
		t.Fatal("blue mana could not pay kicker {1}{G} on a free cast with the spending permission")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.KickerPaid || obj.ManaSpentToCast != 2 {
		t.Fatalf("stack object = %+v, want a free cast that paid only the kicker", obj)
	}

	g, engine, cardID = setup(game.ManaSpendNormal)
	assertExileCastListed(t, g, engine, kicked(cardID), false)
	if engine.applyAction(g, game.Player1, kicked(cardID)) {
		t.Fatal("blue mana paid kicker {1}{G} without the spending permission")
	}
}

// TestSpendAnyColorPaysMutateFromExile casts a Mutate creature from its
// owner's exile through a group cast-from-exile permission (the Adventure
// permission belongs to the player who stole the Adventure spell) and pays the
// Mutate {G} alternative cost with blue mana only because of the rider.
func TestSpendAnyColorPaysMutateFromExile(t *testing.T) {
	t.Parallel()
	setup := func(t *testing.T, spendAs game.ManaSpendPermission) (*game.Game, *Engine, id.ID, action.Action) {
		t.Helper()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		target := addCombatPermanent(g, game.Player1, mutateTargetCard())
		cardID := addCardToHand(g, game.Player1, mutatingAdventureFixture())
		addBasicLandPermanent(g, game.Player1, types.Forest)
		giveMainPhasePriority(g, game.Player1)
		if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
			t.Fatal("Small Errand cast failed")
		}
		obj, _ := g.Stack.Peek()
		stealStackSpell(t, g, engine, obj.ID, game.Player2)
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatal("Small Errand did not resolve into its owner's exile")
		}
		g.Players[game.Player2].ManaPool.Empty()
		permission := castSpellsFromExileEffect(g, game.Player1)
		permission.SpendAnyMana = spendAs
		g.RuleEffects = append(g.RuleEffects, permission)
		addBasicLandPermanent(g, game.Player1, types.Island)
		giveMainPhasePriority(g, game.Player1)
		return g, engine, cardID, action.CastMutateSpellFromZone(cardID, zone.Exile, target.ObjectID)
	}

	g, engine, cardID, act := setup(t, game.ManaSpendAsAnyColor)
	assertExileCastListed(t, g, engine, act, true)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("an Island could not pay Mutate {G} under the any-color permission")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != cardID || !obj.Mutate || obj.SourceZone != zone.Exile {
		t.Fatalf("stack object = %+v, want the Mutate cast from exile", obj)
	}

	g, engine, cardID, act = setup(t, game.ManaSpendNormal)
	assertExileCastListed(t, g, engine, act, false)
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("an Island paid Mutate {G} without the spending permission")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("rejected Mutate cast did not return the card to exile")
	}
}

// TestEvelynGroupPermissionSpendsAnyColorAndSurvivesClone uses Evelyn's group
// cast-from-exile permission (collection counters) and proves the rider covers
// the kicked total cost on both direct validation and application, that the
// permission without the rider does not, and that the rider lives in the
// game-owned rule effect so a cloned game casts with it independently.
func TestEvelynGroupPermissionSpendsAnyColorAndSurvivesClone(t *testing.T) {
	t.Parallel()
	setup := func(rider bool) (*game.Game, *Engine, id.ID) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		permanent := evelynCovetousPermanent(g, game.Player1)
		if !rider {
			effects := g.CardInstances[permanent.CardInstanceID].Def.StaticAbilities[0].RuleEffects
			for i := range effects {
				effects[i].SpendAnyMana = game.ManaSpendNormal
			}
		}
		cardID := addCardToExile(g, game.Player2, kickerExileFixture())
		g.AddExileCounterFromController(cardID, counter.Collection, 1, game.Player1)
		giveMainPhasePriority(g, game.Player1)
		g.Players[game.Player1].ManaPool.Add(mana.U, 3)
		return g, NewEngine(nil), cardID
	}
	kicked := func(cardID id.ID) action.Action {
		return action.CastKickedSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	}
	canCastKicked := func(g *game.Game, engine *Engine, cardID id.ID) bool {
		return engine.canCastSpellFaceFromZoneWithKicker(g, game.Player1, cardID, zone.Exile, game.FaceFront, nil, 0, nil, true)
	}

	g, engine, cardID := setup(false)
	if canCastKicked(g, engine, cardID) || engine.applyAction(g, game.Player1, kicked(cardID)) {
		t.Fatal("blue mana paid {G} plus kicker {1}{G} without Evelyn's rider")
	}

	g, engine, cardID = setup(true)
	if !canCastKicked(g, engine, cardID) {
		t.Fatal("validation rejected the kicked cast under Evelyn's rider")
	}
	clone := g.Clone()
	if !engine.applyAction(clone, game.Player1, kicked(cardID)) {
		t.Fatal("the cloned game could not use Evelyn's rider")
	}
	if obj, ok := clone.Stack.Peek(); !ok || !obj.KickerPaid || obj.SourceZone != zone.Exile {
		t.Fatalf("clone stack object = %+v, want the kicked spell cast from exile", obj)
	}
	if !g.Players[game.Player2].Exile.Contains(cardID) || g.Stack.Size() != 0 ||
		g.Players[game.Player1].ManaPool.Total() != 3 {
		t.Fatal("casting in the clone changed the original game")
	}
	if !engine.applyAction(g, game.Player1, kicked(cardID)) {
		t.Fatal("the original game could not use Evelyn's rider after cloning")
	}
}

// TestSpendRiderStaysWithItsOwnPermission gives Player1 two independent
// permissions for one exiled card: a "without paying its mana cost" permission
// with no rider, and a paying permission whose rider lets mana of any type be
// spent (Court of Locthwain's shape). A rider applies only to a spell cast
// through its own permission (CR 118.14), so the free cast cannot use the other
// permission's rider for kicker, while casting through the rider's permission
// pays the whole kicked cost with any mana.
func TestSpendRiderStaysWithItsOwnPermission(t *testing.T) {
	t.Parallel()
	setup := func(pool ...mana.Color) (*game.Game, *Engine, id.ID) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		cardID := riderExileCard(g, kickerExileFixture(), game.ManaSpendAsAnyType)
		g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
			ID:                    g.IDGen.Next(),
			Kind:                  game.RuleEffectPlayFromZone,
			Controller:            game.Player1,
			AffectedPlayer:        game.PlayerYou,
			Duration:              game.DurationPermanent,
			CastFromZone:          zone.Exile,
			AffectedCardID:        cardID,
			ExpiresFor:            game.Player1,
			WithoutPayingManaCost: true,
		})
		for _, c := range pool {
			g.Players[game.Player1].ManaPool.Add(c, 1)
		}
		return g, NewEngine(nil), cardID
	}
	kicked := func(cardID id.ID) action.Action {
		return action.CastKickedSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	}
	unkicked := func(cardID id.ID) action.Action {
		return action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	}
	assertCast := func(t *testing.T, g *game.Game, engine *Engine, cardID id.ID, act action.Action, wantKicker bool, wantSpent int) {
		t.Helper()
		assertExileCastListed(t, g, engine, act, true)
		if !engine.applyAction(g, game.Player1, act) {
			t.Fatalf("%v was rejected", act)
		}
		obj, ok := g.Stack.Peek()
		if !ok || obj.SourceID != cardID || obj.KickerPaid != wantKicker || obj.ManaSpentToCast != wantSpent {
			t.Fatalf("stack object = %+v, want kicker %v with %d mana spent", obj, wantKicker, wantSpent)
		}
	}

	t.Run("free cast cannot borrow the other permission's rider", func(t *testing.T) {
		t.Parallel()
		g, engine, cardID := setup(mana.U, mana.U)
		assertExileCastListed(t, g, engine, kicked(cardID), false)
		if engine.applyAction(g, game.Player1, kicked(cardID)) {
			t.Fatal("blue mana paid kicker {1}{G} on a free cast authorized by a permission without a rider")
		}
		if !g.Players[game.Player2].Exile.Contains(cardID) || g.Players[game.Player1].ManaPool.Total() != 2 {
			t.Fatal("the rejected cast was not undone")
		}
	})
	t.Run("free cast pays kicker normally", func(t *testing.T) {
		t.Parallel()
		g, engine, cardID := setup(mana.U, mana.G)
		assertCast(t, g, engine, cardID, kicked(cardID), true, 2)
	})
	t.Run("free cast needs no mana", func(t *testing.T) {
		t.Parallel()
		g, engine, cardID := setup()
		assertCast(t, g, engine, cardID, unkicked(cardID), false, 0)
	})
	t.Run("rider permission pays the whole kicked cost", func(t *testing.T) {
		t.Parallel()
		g, engine, cardID := setup(mana.U, mana.U, mana.U)
		assertCast(t, g, engine, cardID, kicked(cardID), true, 3)
	})
}

// TestEvelynRiderPaysMutateAfterCardLeavesExile casts an Adventure-exiled
// Mutate creature through Evelyn's collection-counter group permission. The
// permission and its rider are read before the card leaves exile, since the
// collection counter and so the permission's match are gone once the card is
// moved to the stack; the off-color Mutate cost is then paid with blue mana and
// the once-each-turn use is recorded, so a second card cannot be cast this turn.
func TestEvelynRiderPaysMutateAfterCardLeavesExile(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	evelynCovetousPermanent(g, game.Player1)
	target := addCombatPermanent(g, game.Player1, mutateTargetCard())
	exiled := func() id.ID {
		cardID := addCardToExile(g, game.Player1, mutatingAdventureFixture())
		g.AdventureCards[cardID] = game.AdventureExile{Player: game.Player2, ZoneVersion: g.CardInstances[cardID].ZoneVersion}
		g.AddExileCounterFromController(cardID, counter.Collection, 1, game.Player1)
		return cardID
	}
	first, second := exiled(), exiled()
	giveMainPhasePriority(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.U, 2)

	act := action.CastMutateSpellFromZone(first, zone.Exile, target.ObjectID)
	assertExileCastListed(t, g, engine, act, true)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("Evelyn's rider did not pay Mutate {G} with blue mana")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != first || !obj.Mutate || obj.ManaSpentToCast != 1 {
		t.Fatalf("stack object = %+v, want the Mutate cast paid with one mana", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	giveMainPhasePriority(g, game.Player1)
	if canCastMutateSpell(g, game.Player1, second, zone.Exile, target.ObjectID) {
		t.Fatal("Evelyn's once-each-turn permission was not used by the Mutate cast")
	}
}

// routeLabelAgent answers the spell-cost choice with the option labeled label
// (or the default when absent) and records every label it was offered.
type routeLabelAgent struct {
	label   string
	offered *[]string
}

func (routeLabelAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a routeLabelAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoicePayment {
		return request.DefaultSelection
	}
	selected := request.DefaultSelection
	for _, option := range request.Options {
		*a.offered = append(*a.offered, option.Label)
		if option.Label == a.label {
			selected = []int{option.Index}
		}
	}
	return selected
}

// freePlayPermission lets Player1 play cardID from exile without paying its mana
// cost (Dauthi Voidwalker's shape), with no spending rider.
func freePlayPermission(g *game.Game, cardID id.ID) {
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:                    g.IDGen.Next(),
		Kind:                  game.RuleEffectPlayFromZone,
		Controller:            game.Player1,
		AffectedPlayer:        game.PlayerYou,
		Duration:              game.DurationPermanent,
		CastFromZone:          zone.Exile,
		AffectedCardID:        cardID,
		ExpiresFor:            game.Player1,
		WithoutPayingManaCost: true,
	})
}

// assertExileCastValid checks direct validation of casting cardID's front face
// from exile. Group cast-from-exile permissions are not yet offered by legal
// action enumeration (#3357), so group-only casts are checked here directly.
func assertExileCastValid(t *testing.T, g *game.Game, engine *Engine, cardID id.ID, want bool) {
	t.Helper()
	if got := engine.canCastSpellFaceFromZoneWithKicker(g, game.Player1, cardID, zone.Exile, game.FaceFront, nil, 0, nil, false); got != want {
		t.Fatalf("casting %v from exile valid = %v, want %v", cardID, got, want)
	}
}

// collectionExiledCard puts def into Player2's exile with a collection counter
// from Player1, so Evelyn's group permission for Player1 matches it.
func collectionExiledCard(g *game.Game, def *game.CardDef) id.ID {
	cardID := addCardToExile(g, game.Player2, def)
	g.AddExileCounterFromController(cardID, counter.Collection, 1, game.Player1)
	return cardID
}

// TestCastRouteRecordsOnlyThePermissionUsed casts a collection-countered card
// that a separate "without paying its mana cost" permission also covers. The
// caster chooses which permission to cast through (CR 601.2b); only that
// permission's usage is spent, so casting through the free permission leaves
// Evelyn's once-each-turn permission available for another card, while
// choosing Evelyn's route spends it.
func TestCastRouteRecordsOnlyThePermissionUsed(t *testing.T) {
	t.Parallel()
	setup := func() (*game.Game, *Engine, *game.Permanent, id.ID, id.ID) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		evelyn := evelynCovetousPermanent(g, game.Player1)
		first := collectionExiledCard(g, kickerExileFixture())
		second := collectionExiledCard(g, kickerExileFixture())
		freePlayPermission(g, first)
		giveMainPhasePriority(g, game.Player1)
		return g, NewEngine(nil), evelyn, first, second
	}
	cast := func(cardID id.ID) action.Action {
		return action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	}

	for _, pool := range [][]mana.Color{nil, {mana.U}} {
		g, engine, evelyn, first, second := setup()
		for _, c := range pool {
			g.Players[game.Player1].ManaPool.Add(c, 1)
		}
		if !engine.applyAction(g, game.Player1, cast(first)) {
			t.Fatalf("pool %v: the free cast was rejected", pool)
		}
		if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != first || obj.ManaSpentToCast != 0 {
			t.Fatalf("pool %v: stack object = %+v, want the free cast", pool, obj)
		}
		if g.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID] {
			t.Fatalf("pool %v: the free permission's cast spent Evelyn's once-each-turn use", pool)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		giveMainPhasePriority(g, game.Player1)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		assertExileCastValid(t, g, engine, second, true)
		if !engine.applyAction(g, game.Player1, cast(second)) {
			t.Fatalf("pool %v: Evelyn could not authorize a second card after the free cast", pool)
		}
		if !g.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID] {
			t.Fatalf("pool %v: casting through Evelyn did not spend her once-each-turn use", pool)
		}
	}

	t.Run("choosing Evelyn's route spends her use", func(t *testing.T) {
		t.Parallel()
		g, engine, evelyn, first, second := setup()
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		var offered []string
		agents := [game.NumPlayers]PlayerAgent{game.Player1: routeLabelAgent{label: "Normal cost", offered: &offered}}
		if !engine.applyActionWithChoices(g, game.Player1, cast(first), agents, &TurnLog{}) {
			t.Fatal("casting through Evelyn's route was rejected")
		}
		if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != first || obj.ManaSpentToCast != 1 {
			t.Fatalf("stack object = %+v, want the cast paid with one blue mana; offered %q", obj, offered)
		}
		if !g.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID] {
			t.Fatal("casting through Evelyn's route did not spend her once-each-turn use")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		giveMainPhasePriority(g, game.Player1)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		assertExileCastValid(t, g, engine, second, false)
	})
}

// TestCastRouteOffersEveryRoutesCostOptions gives one exiled evoke creature a
// "without paying its mana cost" permission and a separate paying permission
// whose any-color rider can pay its evoke cost. Choosing a permission and an
// alternative cost are one choice among every payable option (CR 601.2b,
// 118.9), so the free route does not hide evoke through the rider's
// permission, and each option keeps only its own route's rider.
func TestCastRouteOffersEveryRoutesCostOptions(t *testing.T) {
	t.Parallel()
	setup := func() (*game.Game, *Engine, id.ID) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		pt := game.PT{Value: 2}
		cardID := riderExileCard(g, &game.CardDef{CardFace: game.CardFace{
			Name:      "Evoke Route Probe",
			Types:     []types.Card{types.Creature},
			ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.W}),
			Power:     opt.Val(pt),
			Toughness: opt.Val(pt),
			AlternativeCosts: []cost.Alternative{{
				Label:    "Evoke",
				Mechanic: cost.AlternativeMechanicEvoke,
				ManaCost: opt.Val(cost.Mana{cost.W}),
			}},
			TriggeredAbilities: []game.TriggeredAbility{game.EvokeSacrificeTriggeredAbility()},
		}}, game.ManaSpendAsAnyColor)
		freePlayPermission(g, cardID)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		return g, NewEngine(nil), cardID
	}
	act := func(cardID id.ID) action.Action {
		return action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	}

	t.Run("evoke through the rider permission", func(t *testing.T) {
		t.Parallel()
		g, engine, cardID := setup()
		var offered []string
		agents := [game.NumPlayers]PlayerAgent{game.Player1: routeLabelAgent{label: "Evoke", offered: &offered}}
		if !engine.applyActionWithChoices(g, game.Player1, act(cardID), agents, &TurnLog{}) {
			t.Fatal("cast rejected")
		}
		obj, ok := g.Stack.Peek()
		if !ok || obj.SourceID != cardID || !obj.Evoked || obj.ManaSpentToCast != 1 {
			t.Fatalf("stack object = %+v, want evoke paid with one blue mana; offered %q", obj, offered)
		}
		if !slices.Contains(offered, "Without paying its mana cost") || slices.Contains(offered, "Normal cost") {
			t.Fatalf("offered %q, want the free cast and evoke but not the unpayable normal cost", offered)
		}
	})
	t.Run("the free route stays the default", func(t *testing.T) {
		t.Parallel()
		g, engine, cardID := setup()
		if !engine.applyAction(g, game.Player1, act(cardID)) {
			t.Fatal("cast rejected")
		}
		if obj, ok := g.Stack.Peek(); !ok || obj.Evoked || obj.ManaSpentToCast != 0 {
			t.Fatalf("stack object = %+v, want the free cast", obj)
		}
	})
}

// TestEveryMatchingGroupPermissionIsARoute casts a card that both Grolnok's
// croak-counter permission (no rider) and Evelyn's collection-counter
// permission (once each turn, any-color rider) cover, in both orders of the
// permissions. Each matching group permission is a separate route with its own
// rider (CR 118.14): off-color mana casts through Evelyn and spends her use,
// while on-color mana casts through Grolnok and leaves Evelyn's use unspent.
func TestEveryMatchingGroupPermissionIsARoute(t *testing.T) {
	t.Parallel()
	for _, evelynFirst := range []bool{false, true} {
		for _, pool := range []mana.Color{mana.U, mana.G} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			var evelyn *game.Permanent
			if evelynFirst {
				evelyn = evelynCovetousPermanent(g, game.Player1)
				grolnokOmnivorePermanent(g, game.Player1)
			} else {
				grolnokOmnivorePermanent(g, game.Player1)
				evelyn = evelynCovetousPermanent(g, game.Player1)
			}
			cardID := collectionExiledCard(g, kickerExileFixture())
			g.AddExileCounter(cardID, counter.Croak, 1)
			g.Players[game.Player1].ManaPool.Add(pool, 1)
			giveMainPhasePriority(g, game.Player1)
			act := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
			assertExileCastValid(t, g, engine, cardID, true)
			if !engine.applyAction(g, game.Player1, act) {
				t.Fatalf("evelynFirst=%v pool=%v: cast rejected", evelynFirst, pool)
			}
			if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != cardID || obj.ManaSpentToCast != 1 {
				t.Fatalf("evelynFirst=%v pool=%v: stack object = %+v", evelynFirst, pool, obj)
			}
			if used, want := g.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID], pool == mana.U; used != want {
				t.Fatalf("evelynFirst=%v pool=%v: Evelyn used = %v, want %v", evelynFirst, pool, used, want)
			}
		}
	}
}

// TestMutateCastChoosesAmongGroupPermissionRoutes casts an Adventure-exiled
// Mutate creature that Grolnok's and Evelyn's group permissions both cover.
// The Mutate cost {G} is paid through whichever permission's route can pay it:
// green mana through Grolnok's permission leaves Evelyn's once-each-turn use
// unspent, while blue mana needs Evelyn's any-color rider and spends her use.
func TestMutateCastChoosesAmongGroupPermissionRoutes(t *testing.T) {
	t.Parallel()
	for _, pool := range []mana.Color{mana.G, mana.U} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		evelyn := evelynCovetousPermanent(g, game.Player1)
		grolnokOmnivorePermanent(g, game.Player1)
		target := addCombatPermanent(g, game.Player1, mutateTargetCard())
		cardID := addCardToExile(g, game.Player1, mutatingAdventureFixture())
		g.AdventureCards[cardID] = game.AdventureExile{Player: game.Player2, ZoneVersion: g.CardInstances[cardID].ZoneVersion}
		g.AddExileCounterFromController(cardID, counter.Collection, 1, game.Player1)
		g.AddExileCounter(cardID, counter.Croak, 1)
		g.Players[game.Player1].ManaPool.Add(pool, 1)
		giveMainPhasePriority(g, game.Player1)
		if !canCastMutateSpell(g, game.Player1, cardID, zone.Exile, target.ObjectID) {
			t.Fatalf("pool %v: Mutate cast not valid", pool)
		}
		if !engine.applyAction(g, game.Player1, action.CastMutateSpellFromZone(cardID, zone.Exile, target.ObjectID)) {
			t.Fatalf("pool %v: Mutate cast rejected", pool)
		}
		if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != cardID || !obj.Mutate || obj.ManaSpentToCast != 1 {
			t.Fatalf("pool %v: stack object = %+v", pool, obj)
		}
		if used, want := g.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID], pool == mana.U; used != want {
			t.Fatalf("pool %v: Evelyn used = %v, want %v", pool, used, want)
		}
	}
}

// registryCardDef looks up a registered card by its exact name.
func registryCardDef(t *testing.T, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupChecked(name)
	if err != nil {
		t.Fatalf("default-registry lookup of %s failed: %v", name, err)
	}
	return def
}

// linkedFreeCastPermissions counts the active one-shot linked free-cast
// permissions (Court of Locthwain's monarch clause).
func linkedFreeCastPermissions(g *game.Game) int {
	count := 0
	for _, effect := range g.RuleEffects {
		if effect.Kind == game.RuleEffectCastLinkedExileForFree {
			count++
		}
	}
	return count
}

// TestLinkedFreeCastIsOneRouteBesideCourtsPaidPermission casts registered cards
// Court of Locthwain exiled while its controller is the monarch. Court grants
// two separate permissions: playing the card with mana of any type, and,
// for the monarch, casting it without paying its mana cost. Each is a route
// the caster chooses with its own costs (CR 601.2b, 118.9, 118.14), so the
// free cast never hides evoke, kicker, or overload paid through the any-type
// permission, and the one-shot free permission is consumed only when the
// spell is cast through it.
func TestLinkedFreeCastIsOneRouteBesideCourtsPaidPermission(t *testing.T) {
	t.Parallel()
	setup := func(t *testing.T, name string, blue int) (*game.Game, *Engine, id.ID) {
		t.Helper()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		g.Players[game.Player1].IsMonarch = true
		cardID := driveLocthwainUpkeepWithCard(t, engine, g, registryCardDef(t, name))
		if linkedFreeCastPermissions(g) != 1 {
			t.Fatal("precondition: the monarch's linked free-cast permission is missing")
		}
		g.Players[game.Player1].ManaPool.Add(mana.U, blue)
		giveMainPhasePriority(g, game.Player1)
		return g, engine, cardID
	}

	t.Run("evoke through the any-type permission keeps the free cast", func(t *testing.T) {
		t.Parallel()
		g, engine, cardID := setup(t, "Wispmare", 1)
		var offered []string
		agents := [game.NumPlayers]PlayerAgent{game.Player1: routeLabelAgent{label: "Evoke", offered: &offered}}
		act := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
		if !engine.applyActionWithChoices(g, game.Player1, act, agents, &TurnLog{}) {
			t.Fatal("cast rejected")
		}
		obj, ok := g.Stack.Peek()
		if !ok || obj.SourceID != cardID || !obj.Evoked || obj.ManaSpentToCast != 1 {
			t.Fatalf("stack object = %+v, want evoke paid with one blue mana; offered %q", obj, offered)
		}
		if !slices.Contains(offered, "Without paying its mana cost") {
			t.Fatalf("offered %q, want the free cast beside evoke", offered)
		}
		if linkedFreeCastPermissions(g) != 1 {
			t.Fatal("casting through the any-type permission consumed the free-cast permission")
		}
	})
	t.Run("the free cast is the default and consumes its permission", func(t *testing.T) {
		t.Parallel()
		g, engine, cardID := setup(t, "Wispmare", 1)
		if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
			t.Fatal("cast rejected")
		}
		if obj, ok := g.Stack.Peek(); !ok || obj.Evoked || obj.ManaSpentToCast != 0 {
			t.Fatalf("stack object = %+v, want the free cast", obj)
		}
		if linkedFreeCastPermissions(g) != 0 {
			t.Fatal("the free cast did not consume its one-shot permission")
		}
	})
	t.Run("kicker paid with any-type mana", func(t *testing.T) {
		t.Parallel()
		g, engine, cardID := setup(t, "Goblin Bushwhacker", 2)
		act := action.CastKickedSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
		assertExileCastListed(t, g, engine, act, true)
		if !engine.applyAction(g, game.Player1, act) {
			t.Fatal("kicked cast rejected")
		}
		if obj, ok := g.Stack.Peek(); !ok || !obj.KickerPaid || obj.ManaSpentToCast != 2 {
			t.Fatalf("stack object = %+v, want {R} and kicker {R} paid with two blue mana", obj)
		}
		if linkedFreeCastPermissions(g) != 1 {
			t.Fatal("the paid cast consumed the free-cast permission")
		}
	})
	t.Run("overload paid with any-type mana", func(t *testing.T) {
		t.Parallel()
		g, engine, cardID := setup(t, "Vandalblast", 5)
		act := action.CastOverloadedSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil)
		assertExileCastListed(t, g, engine, act, true)
		if !engine.applyAction(g, game.Player1, act) {
			t.Fatal("overloaded cast rejected")
		}
		if obj, ok := g.Stack.Peek(); !ok || !obj.Overloaded || obj.ManaSpentToCast != 5 {
			t.Fatalf("stack object = %+v, want overload paid with five blue mana", obj)
		}
		if linkedFreeCastPermissions(g) != 1 {
			t.Fatal("the overloaded cast consumed the free-cast permission")
		}
	})
}

// ownPermissionExileCard puts def into Player1's own exile with a per-card
// permission for Player1 to cast it, letting mana be spent as though it were
// mana of any color.
func ownPermissionExileCard(g *game.Game, def *game.CardDef) id.ID {
	cardID := addCardToExile(g, game.Player1, def)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectPlayFromZone,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		Duration:       game.DurationPermanent,
		CastFromZone:   zone.Exile,
		AffectedCardID: cardID,
		ExpiresFor:     game.Player1,
		SpendAnyMana:   game.ManaSpendAsAnyColor,
	})
	return cardID
}

// TestForetoldAndPlottedCardsKeepOtherPermissionRoutes proves a card's own
// foretell or plot permission is one route beside any other permission to cast
// it (CR 601.2b): a foretold card whose foretell cost is unaffordable can still
// be cast for its mana cost through another permission's any-color rider, and
// a plotted card cast at instant speed, when plot's sorcery-timing permission
// (CR 702.170d) is unavailable, can still be cast through another permission.
func TestForetoldAndPlottedCardsKeepOtherPermissionRoutes(t *testing.T) {
	t.Parallel()
	instant := func(keyword game.KeywordAbility) *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{
			Name:            "Own Permission Probe",
			Types:           []types.Card{types.Instant},
			ManaCost:        opt.Val(cost.Mana{cost.G}),
			SpellAbility:    opt.Val(game.AbilityContent{}),
			StaticAbilities: []game.StaticAbility{{KeywordAbilities: []game.KeywordAbility{keyword}}},
		}}
	}
	act := func(cardID id.ID) action.Action {
		return action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	}

	t.Run("foretold", func(t *testing.T) {
		t.Parallel()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID := ownPermissionExileCard(g, instant(game.ForetellKeyword{Cost: cost.Mana{cost.O(5)}}))
		g.ForetoldCards[cardID] = 1
		g.Turn.TurnNumber = 2
		giveMainPhasePriority(g, game.Player1)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		if !cardIsForetoldInExile(g, game.Player1, cardID) {
			t.Fatal("precondition: the card is not castable as foretold")
		}
		assertExileCastListed(t, g, engine, act(cardID), true)
		if !engine.applyAction(g, game.Player1, act(cardID)) {
			t.Fatal("cast through the any-color permission rejected")
		}
		if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != cardID || obj.ManaSpentToCast != 1 {
			t.Fatalf("stack object = %+v, want {G} paid with one blue mana", obj)
		}
	})
	t.Run("plotted at instant speed", func(t *testing.T) {
		t.Parallel()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID := ownPermissionExileCard(g, instant(game.PlotKeyword{Cost: cost.Mana{cost.O(1)}}))
		g.PlottedCards[cardID] = 1
		g.Turn.TurnNumber = 2
		g.Turn.ActivePlayer = game.Player2
		g.Turn.PriorityPlayer = game.Player1
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone
		if !cardIsPlottedInExile(g, game.Player1, cardID) {
			t.Fatal("precondition: the card is not plotted")
		}
		assertExileCastValid(t, g, engine, cardID, false)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		assertExileCastListed(t, g, engine, act(cardID), true)
		if !engine.applyAction(g, game.Player1, act(cardID)) {
			t.Fatal("instant-speed cast through the other permission rejected")
		}
		if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != cardID || obj.ManaSpentToCast != 1 {
			t.Fatalf("stack object = %+v, want {G} paid with one blue mana", obj)
		}
	})
}

// TestAdventurePermissionCoversOnlyThePermanentFace casts registered Faerie
// Guidemother, exiled on an Adventure and carrying Evelyn's collection counter.
// The Adventure permission covers only the creature face (CR 715.3d), so
// casting Gift of the Fae from exile must go through Evelyn's permission and
// spend her once-each-turn use, while the creature face defaults to the
// Adventure permission and leaves her use unspent.
func TestAdventurePermissionCoversOnlyThePermanentFace(t *testing.T) {
	t.Parallel()
	setup := func(t *testing.T) (*game.Game, *Engine, *game.Permanent, id.ID) {
		t.Helper()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		evelyn := groupExileRegistryPermanent(t, g, game.Player1, "Evelyn, the Covetous")
		cardID := addCardToExile(g, game.Player1, registryCardDef(t, "Faerie Guidemother"))
		g.AdventureCards[cardID] = game.AdventureExile{Player: game.Player1, ZoneVersion: g.CardInstances[cardID].ZoneVersion}
		g.AddExileCounterFromController(cardID, counter.Collection, 1, game.Player1)
		g.Players[game.Player1].ManaPool.Add(mana.W, 2)
		giveMainPhasePriority(g, game.Player1)
		return g, NewEngine(nil), evelyn, cardID
	}

	t.Run("Adventure face", func(t *testing.T) {
		t.Parallel()
		g, engine, evelyn, cardID := setup(t)
		act := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceAlternate, []game.Target{game.PermanentTarget(evelyn.ObjectID)}, 0, nil)
		if !engine.applyAction(g, game.Player1, act) {
			t.Fatal("Gift of the Fae cast through Evelyn rejected")
		}
		if !g.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID] {
			t.Fatal("the Adventure face was cast without spending Evelyn's once-each-turn use")
		}
	})
	t.Run("permanent face", func(t *testing.T) {
		t.Parallel()
		g, engine, evelyn, cardID := setup(t)
		if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
			t.Fatal("creature face cast rejected")
		}
		if g.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID] {
			t.Fatal("the creature face spent Evelyn's use although the Adventure permission covers it")
		}
	})
}

// TestFreeOnlyPermissionCannotAuthorizeAnotherAlternativeCost gives an exiled
// overload card a permission only to cast it without paying its mana cost.
// That permission cannot authorize overload instead (CR 118.9a), while a
// separate paying permission still can, with its own any-color rider.
func TestFreeOnlyPermissionCannotAuthorizeAnotherAlternativeCost(t *testing.T) {
	t.Parallel()
	setup := func(rider bool) (*game.Game, *Engine, id.ID) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addOverloadArtifact(g, game.Player2, false)
		var cardID id.ID
		if rider {
			cardID = riderExileCard(g, vandalblastDefinition(), game.ManaSpendAsAnyColor)
		} else {
			cardID = addCardToExile(g, game.Player2, vandalblastDefinition())
			giveMainPhasePriority(g, game.Player1)
		}
		freePlayPermission(g, cardID)
		return g, NewEngine(nil), cardID
	}

	g, engine, cardID := setup(false)
	g.Players[game.Player1].ManaPool.Add(mana.R, 5)
	overload := action.CastOverloadedSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil)
	assertExileCastListed(t, g, engine, overload, false)
	if engine.applyAction(g, game.Player1, overload) {
		t.Fatal("a free-only permission authorized an overloaded cast")
	}

	g, engine, cardID = setup(true)
	g.Players[game.Player1].ManaPool.Add(mana.U, 5)
	overload = action.CastOverloadedSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil)
	assertExileCastListed(t, g, engine, overload, true)
	if !engine.applyAction(g, game.Player1, overload) {
		t.Fatal("the paying permission could not authorize overload")
	}
	if obj, ok := g.Stack.Peek(); !ok || !obj.Overloaded || obj.ManaSpentToCast != 5 {
		t.Fatalf("stack object = %+v, want overload paid with five blue mana", obj)
	}
}
