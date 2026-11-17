package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned corpus Qarsi Deceiver e12681cc-c66e-4dac-882c-bd8638a2e12c
// (DTK printing 64d64264-895f-4d4d-8e5c-ab1d0d6e340b): "{T}: Add {C}. Spend
// this mana only to cast a face-down creature spell, pay a mana cost to turn a
// manifested creature face up, or pay a morph cost. (A megamorph cost is a
// morph cost.)" Its rulings_uri returned four 2015-02-25 rulings: the mana
// can't cast a spell that instructs you to manifest, megamorph turn-up is a
// special action, a manifested megamorph card turned up for its mana cost gets
// no counter, and megamorph is a morph variant. CR effective November 6, 2026:
// 106.6 (spend restrictions), 116.2b (turning face up is a special action),
// 701.40a-c (manifest; mana cost turn-up), 701.58a-c (a cloaked permanent is
// not a manifested one), 702.37b-e (morph/megamorph costs), 702.168d (a
// disguise cost is not a morph cost) and 708.4 (face-down spells). The card is
// not registered; qarsiDeceiverFixture carries only its mana ability.
func qarsiDeceiverFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Qarsi Deceiver",
		ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.U}),
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Snake, types.Wizard},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 4}),
		ManaAbilities: []game.ManaAbility{{
			Text:            "{T}: Add {C}. Spend this mana only to cast a face-down creature spell, pay a mana cost to turn a manifested creature face up, or pay a morph cost. (A megamorph cost is a morph cost.)",
			AdditionalCosts: cost.Tap,
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{
				Amount:    game.Fixed(1),
				ManaColor: mana.C,
				SpendRider: opt.Val(game.ManaSpendRider{
					Condition:   game.ManaSpendFaceDownOrTurnFaceUp,
					Restriction: game.ManaSpendRestrictedToCondition,
				}),
			}}}}.Ability(),
		}},
		OracleText: `{T}: Add {C}. Spend this mana only to cast a face-down creature spell, pay a mana cost to turn a manifested creature face up, or pay a morph cost. (A megamorph cost is a morph cost.)`,
	}}
}

func qarsiMainPhaseGame(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	qarsi := addCombatPermanent(g, game.Player1, qarsiDeceiverFixture())
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	return g, engine, qarsi
}

func floatQarsiMana(g *game.Game) {
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.C, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:       mana.Unit{Color: mana.C},
		Controller: game.Player1,
		Rider: game.ManaSpendRider{
			Condition:   game.ManaSpendFaceDownOrTurnFaceUp,
			Restriction: game.ManaSpendRestrictedToCondition,
		},
	})
}

// TestQarsiDeceiverManaCastsFaceDownCreatureSpells proves Qarsi's mana pays
// for a face-down creature spell (CR 708.4) but not for the same card cast
// face up, including as a mana ability activated while paying.
func TestQarsiDeceiverManaCastsFaceDownCreatureSpells(t *testing.T) {
	t.Run("face-down cast", func(t *testing.T) {
		g, engine, qarsi := qarsiMainPhaseGame(t)
		cardID := addCardToHand(g, game.Player1, ainokSurvivalistFixture())
		g.Players[game.Player1].ManaPool.Add(mana.G, 2)
		act := actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownMorph)
		if !containsAction(engine.legalActions(g, game.Player1), act) {
			t.Fatal("face-down cast payable with Qarsi mana is not a legal action")
		}
		if !engine.applyAction(g, game.Player1, act) {
			t.Fatal("face-down cast for {3} with Qarsi Deceiver's mana failed")
		}
		if !qarsi.Tapped {
			t.Fatal("Qarsi Deceiver was not tapped to pay for the face-down spell")
		}
		player := g.Players[game.Player1]
		if player.ManaPool.Total() != 0 || len(player.ManaRiders) != 0 {
			t.Fatalf("pool=%d riders=%+v after face-down cast, want all spent", player.ManaPool.Total(), player.ManaRiders)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		ainok := battlefieldPermanentForCard(g, cardID)
		if ainok == nil || !ainok.FaceDown {
			t.Fatalf("face-down Ainok = %+v", ainok)
		}
	})
	t.Run("face-up cast", func(t *testing.T) {
		g, engine, qarsi := qarsiMainPhaseGame(t)
		cardID := addCardToHand(g, game.Player1, ainokSurvivalistFixture())
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Forest", Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Forest},
			ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.G)},
		}})
		act := actionBuild.castSpell(cardID, zone.Hand, game.FaceFront, nil, 0, nil)
		if containsAction(engine.legalActions(g, game.Player1), act) {
			t.Fatal("face-up Ainok Survivalist payable only with Qarsi mana is a legal action")
		}
		if engine.applyAction(g, game.Player1, act) {
			t.Fatal("Qarsi Deceiver's mana paid for a face-up creature spell")
		}
		if qarsi.Tapped {
			t.Fatal("failed face-up cast tapped Qarsi Deceiver")
		}
	})
}

// TestQarsiDeceiverManaPaysMorphCosts proves Qarsi's mana pays megamorph and
// morph costs (CR 702.37b, 116.2b) and its rider is consumed with the mana.
func TestQarsiDeceiverManaPaysMorphCosts(t *testing.T) {
	t.Run("megamorph", func(t *testing.T) {
		g, engine, qarsi := qarsiMainPhaseGame(t)
		ainok := addFaceDownPermanent(g, game.Player1, ainokSurvivalistFixture(), game.FaceDownMorph)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		agents := megamorphAgents(&megamorphMayPolicy{})
		if !engine.canTurnFaceUp(g, game.Player1, ainok.ObjectID) {
			t.Fatal("megamorph {1}{G} payable with Qarsi mana and {G} is not offered")
		}
		if !engine.applyActionWithChoices(g, game.Player1, actionBuild.turnFaceUp(ainok.ObjectID), agents, &TurnLog{}) {
			t.Fatal("turn face up for megamorph with Qarsi mana failed")
		}
		if ainok.FaceDown || !qarsi.Tapped {
			t.Fatalf("Ainok faceDown=%t Qarsi tapped=%t, want face up and tapped", ainok.FaceDown, qarsi.Tapped)
		}
		if got := ainok.Counters.Get(counter.PlusOnePlusOne); got != 1 {
			t.Fatalf("+1/+1 counters = %d, want 1 for megamorph", got)
		}
	})
	t.Run("morph with floating mana", func(t *testing.T) {
		g, engine, _ := qarsiMainPhaseGame(t)
		g.Turn.Phase = game.PhaseCombat
		g.Turn.Step = game.StepDeclareBlockers
		morph := addFaceDownPermanent(g, game.Player1, morphCreature(cost.Mana{cost.O(2)}), game.FaceDownMorph)
		floatQarsiMana(g)
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)
		if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(morph.ObjectID)) {
			t.Fatal("turn face up for morph {2} with floating Qarsi mana failed")
		}
		player := g.Players[game.Player1]
		if morph.FaceDown || player.ManaPool.Total() != 0 || len(player.ManaRiders) != 0 {
			t.Fatalf("faceDown=%t pool=%d riders=%+v, want face up with Qarsi mana and rider spent", morph.FaceDown, player.ManaPool.Total(), player.ManaRiders)
		}
	})
}

// TestQarsiDeceiverManaPaysManifestedManaCost proves Qarsi's mana pays the
// mana cost of a manifested creature card (CR 701.40a) without adding the
// megamorph counter, while a cloaked creature's mana cost is not payable
// (CR 701.58: cloaked, not manifested).
func TestQarsiDeceiverManaPaysManifestedManaCost(t *testing.T) {
	t.Run("Temur War Shaman manifests Marsh Hulk", func(t *testing.T) {
		g, engine, qarsi := qarsiMainPhaseGame(t)
		hulkID := addCardToLibrary(g, game.Player1, marshHulkFixture())
		shamanID := addCardToHand(g, game.Player1, megamorphTestCard(t, "fb683e18-50b8-4821-9a1a-86c6dea63ac9", "Temur War Shaman"))
		g.Players[game.Player1].ManaPool.Add(mana.G, 6)
		agents := megamorphAgents(&megamorphMayPolicy{payment: []int{0}})
		if !engine.applyActionWithChoices(g, game.Player1, actionBuild.castSpell(shamanID, zone.Hand, game.FaceFront, nil, 0, nil), agents, &TurnLog{}) {
			t.Fatal("Temur War Shaman cast failed")
		}
		if qarsi.Tapped {
			t.Fatal("Qarsi Deceiver paid for Temur War Shaman, a face-up spell that manifests")
		}
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
		resolveMegamorphTriggers(t, g, engine, agents)
		hulk := battlefieldPermanentForCard(g, hulkID)
		if hulk == nil || !hulk.FaceDown || hulk.FaceDownKind != game.FaceDownManifest {
			t.Fatalf("manifested Marsh Hulk = %+v", hulk)
		}

		g.Players[game.Player1].ManaPool.Add(mana.B, 5)
		if !engine.applyActionWithChoices(g, game.Player1, actionBuild.turnFaceUp(hulk.ObjectID), agents, &TurnLog{}) {
			t.Fatal("turn manifested Marsh Hulk face up for {4}{B}{B} with Qarsi mana failed")
		}
		if hulk.FaceDown || !qarsi.Tapped {
			t.Fatalf("Marsh Hulk faceDown=%t Qarsi tapped=%t", hulk.FaceDown, qarsi.Tapped)
		}
		if got := hulk.Counters.Get(counter.PlusOnePlusOne); got != 0 {
			t.Fatalf("+1/+1 counters = %d, want 0 for the mana cost", got)
		}
		if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
			t.Fatalf("mana left = %d, want 0", got)
		}
	})
	for _, tc := range []struct {
		name string
		kind game.FaceDownKind
		want bool
	}{
		{name: "manifested", kind: game.FaceDownManifest, want: true},
		{name: "cloaked", kind: game.FaceDownCloak, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, engine, qarsi := qarsiMainPhaseGame(t)
			permanent := addFaceDownPermanent(g, game.Player1, manifestCreature(cost.Mana{cost.O(1), cost.G}), tc.kind)
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			if got := engine.canTurnFaceUp(g, game.Player1, permanent.ObjectID); got != tc.want {
				t.Fatalf("canTurnFaceUp = %t, want %t", got, tc.want)
			}
			if got := containsAction(engine.legalTurnFaceUpActions(g, game.Player1), actionBuild.turnFaceUp(permanent.ObjectID)); got != tc.want {
				t.Fatalf("turn face-up action listed = %t, want %t", got, tc.want)
			}
			if got := engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)); got != tc.want {
				t.Fatalf("turn face up = %t, want %t", got, tc.want)
			}
			if permanent.FaceDown == tc.want || qarsi.Tapped != tc.want {
				t.Fatalf("faceDown=%t Qarsi tapped=%t", permanent.FaceDown, qarsi.Tapped)
			}
		})
	}
}

// TestQarsiDeceiverManaCannotPayOtherCosts proves Qarsi's mana can't pay a
// disguise cost (CR 702.168d) or an activated ability's cost, and that
// unspent floating Qarsi mana empties with its rider (CR 106.4).
func TestQarsiDeceiverManaCannotPayOtherCosts(t *testing.T) {
	t.Run("disguise cost", func(t *testing.T) {
		g, engine, qarsi := qarsiMainPhaseGame(t)
		veiled := addFaceDownPermanent(g, game.Player1, disguiseCreature(cost.Mana{cost.O(1)}), game.FaceDownDisguise)
		if engine.canTurnFaceUp(g, game.Player1, veiled.ObjectID) {
			t.Fatal("disguise {1} is payable with Qarsi mana")
		}
		if containsAction(engine.legalTurnFaceUpActions(g, game.Player1), actionBuild.turnFaceUp(veiled.ObjectID)) {
			t.Fatal("disguise turn face-up action listed with only Qarsi mana")
		}
		if engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(veiled.ObjectID)) || !veiled.FaceDown || qarsi.Tapped {
			t.Fatalf("disguise turn-up with Qarsi mana: faceDown=%t Qarsi tapped=%t", veiled.FaceDown, qarsi.Tapped)
		}
	})
	t.Run("activated ability", func(t *testing.T) {
		g, engine, _ := qarsiMainPhaseGame(t)
		source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Life Totem", Types: []types.Card{types.Creature},
			Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{{
				ManaCost: opt.Val(cost.Mana{cost.O(1)}),
				Content: game.Mode{
					Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
				}.Ability(),
			}},
		}})
		floatQarsiMana(g)
		act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
		if containsAction(engine.legalActions(g, game.Player1), act) {
			t.Fatal("activated ability payable only with Qarsi mana is a legal action")
		}
		if engine.applyAction(g, game.Player1, act) {
			t.Fatal("Qarsi mana paid an activated ability cost")
		}
		emptyManaPools(g)
		player := g.Players[game.Player1]
		if player.ManaPool.Total() != 0 || len(player.ManaRiders) != 0 {
			t.Fatalf("pool=%d riders=%+v after emptying, want none", player.ManaPool.Total(), player.ManaRiders)
		}
	})
}

func battlefieldPermanentForCard(g *game.Game, cardID game.ObjectID) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return permanent
		}
	}
	return nil
}
