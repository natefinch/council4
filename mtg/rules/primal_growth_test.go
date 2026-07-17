package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func primalGrowthLikeDef() *game.CardDef {
	basicLand := game.Selection{
		RequiredTypes: []types.Card{types.Land},
		Supertypes:    []types.Super{types.Basic},
	}
	search := func(amount int, kicked bool) game.Instruction {
		condition := game.Condition{SpellWasKicked: true, Negate: !kicked}
		return game.Instruction{
			Primitive: game.Search{
				Player: game.ControllerReference(),
				Spec: game.SearchSpec{
					SourceZone:  zone.Library,
					Destination: zone.Battlefield,
					Filter:      basicLand,
				},
				Amount: game.Fixed(amount),
			},
			Condition: opt.Val(game.EffectCondition{Condition: opt.Val(condition)}),
		}
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Primal Growth",
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.G}),
		Types:    []types.Card{types.Sorcery},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{
				AdditionalCosts: []cost.Additional{{
					Kind:               cost.AdditionalSacrifice,
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Creature,
				}},
			}},
		}},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{
			search(1, false),
			search(2, true),
		}}.Ability()),
	}}
}

func addPrimalGrowthBasic(t *testing.T, g *game.Game, name string, subtype types.Sub) id.ID {
	t.Helper()
	return addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{subtype},
	}})
}

func preparePrimalGrowthCast(g *game.Game) {
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
}

func TestPrimalGrowthNormalSearchesForOneWithoutSacrifice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := addTokenCreaturePermanent(g, game.Player1, "Saproling")
	first := addPrimalGrowthBasic(t, g, "Plains", types.Plains)
	second := addPrimalGrowthBasic(t, g, "Island", types.Island)
	spellID := addCardToHand(g, game.Player1, primalGrowthLikeDef())
	preparePrimalGrowthCast(g)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("normal Primal Growth cast failed")
	}
	if _, ok := permanentByObjectID(g, token.ObjectID); !ok {
		t.Fatal("normal cast paid the unchosen Kicker sacrifice")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	found := 0
	for _, cardID := range []id.ID{first, second} {
		if permanent := permanentForCard(g, cardID); permanent != nil {
			found++
			if permanent.Controller != game.Player1 || permanent.Tapped {
				t.Fatalf("searched land = %#v, want untapped under caster's control", permanent)
			}
		}
	}
	if found != 1 {
		t.Fatalf("found lands = %d, want 1", found)
	}
}

func TestPrimalGrowthKickedSacrificesControlledTokenAndSearchesForTwo(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := addTokenCreaturePermanent(g, game.Player1, "Saproling")
	opponentCreature := addTokenCreaturePermanent(g, game.Player2, "Opponent Bear")
	first := addPrimalGrowthBasic(t, g, "Plains", types.Plains)
	second := addPrimalGrowthBasic(t, g, "Island", types.Island)
	third := addPrimalGrowthBasic(t, g, "Swamp", types.Swamp)
	spellID := addCardToHand(g, game.Player1, primalGrowthLikeDef())
	preparePrimalGrowthCast(g)

	if !engine.applyAction(g, game.Player1, action.CastKickedSpell(spellID, nil, 0, nil)) {
		t.Fatal("kicked Primal Growth cast failed")
	}
	if _, ok := permanentByObjectID(g, token.ObjectID); ok {
		t.Fatal("controlled creature token was not sacrificed")
	}
	if _, ok := permanentByObjectID(g, opponentCreature.ObjectID); !ok {
		t.Fatal("opponent-controlled creature was used to pay the Kicker cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	found := 0
	for _, cardID := range []id.ID{first, second, third} {
		if permanent := permanentForCard(g, cardID); permanent != nil {
			found++
			if permanent.Controller != game.Player1 || permanent.Tapped {
				t.Fatalf("searched land %v = %#v, want untapped under caster's control", cardID, permanent)
			}
		}
	}
	if found != 2 {
		t.Fatalf("found lands = %d, want 2", found)
	}
}

func TestPrimalGrowthFailedKickerPaymentRollsBack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := addTokenCreaturePermanent(g, game.Player1, "Saproling")
	spellID := addCardToHand(g, game.Player1, primalGrowthLikeDef())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if engine.applyAction(g, game.Player1, action.CastKickedSpell(spellID, nil, 0, nil)) {
		t.Fatal("kicked cast succeeded without mana")
	}
	if _, ok := permanentByObjectID(g, token.ObjectID); !ok {
		t.Fatal("failed Kicker payment sacrificed the token")
	}
	if !g.Players[game.Player1].Hand.Contains(spellID) || g.Stack.Size() != 0 {
		t.Fatal("failed Kicker payment did not restore the spell proposal")
	}
}

func TestDirectCastRejectsExcessiveMultikickerCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:         "Bounded Multikicker",
		ManaCost:     opt.Val(cost.Mana{cost.O(0)}),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{
				Cost:  cost.Mana{cost.O(0)},
				Multi: true,
			}},
		}},
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	act := action.CastMultikickedSpellFaceFromZone(
		spellID,
		zone.Hand,
		game.FaceFront,
		nil,
		0,
		nil,
		maxLegalMultikickCount+1,
	)
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("direct cast accepted excessive Multikicker count")
	}
	if !g.Players[game.Player1].Hand.Contains(spellID) || g.Stack.Size() != 0 {
		t.Fatal("rejected excessive Multikicker cast mutated the proposal")
	}
	negative := action.CastMultikickedSpellFaceFromZone(
		spellID,
		zone.Hand,
		game.FaceFront,
		nil,
		0,
		nil,
		-1,
	)
	if engine.applyAction(g, game.Player1, negative) {
		t.Fatal("direct cast accepted negative Multikicker count")
	}
}
