package rules

// Source review: Aftermath (Engine gap #3509).
// CR 702.127a (Comprehensive Rules effective 2026-11-06): "Aftermath is an
// ability found on some split cards ... 'You may cast this half of this split
// card from your graveyard,' 'This half of this split card can't be cast from
// any zone other than a graveyard,' and 'If this spell was cast from a
// graveyard, exile it instead of putting it anywhere else any time it would
// leave the stack.'" CR 709.3a: only the chosen half is evaluated to see if it
// can be cast. CR 709.3c/707.12: a cast copy of a split card keeps both halves.
// Commit // Memory rulings (2017-04-18): if another effect allows you to cast
// a split card with aftermath from any zone other than a graveyard, you can't
// cast the half with aftermath; if another effect allows you to cast it from a
// graveyard, you may cast either half, and the aftermath half is exiled if it
// would leave the stack; a spell with aftermath cast from a graveyard is always
// exiled afterward, whether it resolves, is countered, or leaves the stack in
// some other way. (The 2017 ruling that a split card's combined mana value is
// checked while casting is superseded by current CR 709.3a.) Pinned corpus
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4 prints
// aftermath cards with layout "split" and Aftermath on the second half.

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// aftermathCard is a split card shaped like Commit // Memory: an instant first
// half cast from hand ("Draw a card.") and a sorcery Aftermath half ("Tap
// target creature. You gain 2 life.").
func aftermathCard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Aftermath First",
			Types:    []types.Card{types.Instant},
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
			Colors:   []color.Color{color.White},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
				}},
			}.Ability()),
		},
		Layout: game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{
			Name:            "Aftermath Second",
			Types:           []types.Card{types.Sorcery},
			ManaCost:        opt.Val(cost.Mana{cost.O(1), cost.B}),
			Colors:          []color.Color{color.Black},
			StaticAbilities: []game.StaticAbility{game.AftermathStaticBody},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}},
				Sequence: []game.Instruction{
					{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}},
					{Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()}},
				},
			}.Ability()),
		}),
	}
}

func newAftermathGame() (*Engine, *game.Game) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, types.Plains)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	return NewEngine(nil), g
}

func castAftermathHalfFromGraveyard(t *testing.T, engine *Engine, g *game.Game, cardID id.ID, target *game.Permanent) *game.StackObject {
	t.Helper()
	cast := action.CastSpellFaceFromZone(cardID, zone.Graveyard, game.FaceAlternate, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("casting the Aftermath half from its owner's graveyard failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Face != game.FaceAlternate || obj.SourceZone != zone.Graveyard {
		t.Fatalf("stack object = %+v, want the Aftermath half cast from the graveyard", obj)
	}
	return obj
}

// TestAftermathCastsFirstHalfFromHandThenAftermathHalfFromGraveyard plays an
// aftermath card through its whole life: only the first half can be cast from
// hand, and it resolves to the graveyard as usual; from the graveyard only the
// Aftermath half can be cast, paying its normal mana cost, and it is exiled
// when it resolves (CR 702.127a).
func TestAftermathCastsFirstHalfFromHandThenAftermathHalfFromGraveyard(t *testing.T) {
	engine, g := newAftermathGame()
	cardID := addCardToHand(g, game.Player1, aftermathCard())
	drawnID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	addBasicLandPermanent(g, game.Player1, types.Plains)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	target := addCreaturePermanent(g, game.Player2)
	targets := []game.Target{game.PermanentTarget(target.ObjectID)}

	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, action.CastSpellFace(cardID, game.FaceFront, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, want the first half castable from hand", legal)
	}
	afterFromHand := action.CastSpellFace(cardID, game.FaceAlternate, targets, 0, nil)
	if actionsContain(legal, afterFromHand) || engine.applyAction(g, game.Player1, afterFromHand) {
		t.Fatal("the Aftermath half was castable from hand")
	}

	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceFront, nil, 0, nil)) {
		t.Fatal("casting the first half from hand failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	player := g.Players[game.Player1]
	if !player.Graveyard.Contains(cardID) || !player.Hand.Contains(drawnID) {
		t.Fatal("the first half did not draw a card and go to its owner's graveyard")
	}

	legal = engine.legalActions(g, game.Player1)
	if actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Graveyard, game.FaceFront, nil, 0, nil)) ||
		engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Graveyard, game.FaceFront, nil, 0, nil)) {
		t.Fatal("Aftermath let the first half be cast from the graveyard")
	}
	if !actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Graveyard, game.FaceAlternate, targets, 0, nil)) {
		t.Fatalf("legal actions = %+v, want the Aftermath half castable from the graveyard", legal)
	}
	castAftermathHalfFromGraveyard(t, engine, g, cardID, target)
	if slices.ContainsFunc(g.Battlefield, func(p *game.Permanent) bool {
		return p.Controller == game.Player1 && !p.Tapped
	}) {
		t.Fatal("the Aftermath half's normal {1}{B} cost was not paid")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !target.Tapped || player.Life != 42 {
		t.Fatalf("target tapped=%t life=%d, want the Aftermath half to tap the creature and gain 2 life", target.Tapped, player.Life)
	}
	if !player.Exile.Contains(cardID) || player.Graveyard.Contains(cardID) {
		t.Fatal("the resolved Aftermath half was not exiled")
	}
}

// TestAftermathHalfIsExiledWhenCountered counters the Aftermath half cast from
// a graveyard; it is exiled instead of going to the graveyard.
func TestAftermathHalfIsExiledWhenCountered(t *testing.T) {
	engine, g := newAftermathGame()
	cardID := addCardToGraveyard(g, game.Player1, aftermathCard())
	target := addCreaturePermanent(g, game.Player2)

	obj := castAftermathHalfFromGraveyard(t, engine, g, cardID, target)
	if !counterStackObject(g, obj.ID) {
		t.Fatal("counterStackObject() = false, want true")
	}
	player := g.Players[game.Player1]
	if !player.Exile.Contains(cardID) || player.Graveyard.Contains(cardID) {
		t.Fatal("the countered Aftermath half was not exiled")
	}
}

// TestAftermathHalfIsExiledWhenItsTargetIsIllegal resolves the Aftermath half
// after its only target left the battlefield; the spell doesn't resolve (CR
// 608.2b) and is exiled rather than put into the graveyard.
func TestAftermathHalfIsExiledWhenItsTargetIsIllegal(t *testing.T) {
	engine, g := newAftermathGame()
	cardID := addCardToGraveyard(g, game.Player1, aftermathCard())
	target := addCreaturePermanent(g, game.Player2)

	castAftermathHalfFromGraveyard(t, engine, g, cardID, target)
	movePermanentToZone(g, target, zone.Graveyard)
	engine.resolveTopOfStack(g, &TurnLog{})

	player := g.Players[game.Player1]
	if player.Life != 40 {
		t.Fatalf("life = %d, want the fizzled Aftermath half not to resolve", player.Life)
	}
	if !player.Exile.Contains(cardID) || player.Graveyard.Contains(cardID) {
		t.Fatal("the fizzled Aftermath half was not exiled")
	}
}

// TestAftermathHalfCannotBeCastFromExileByAnotherPermission gives the card a
// "you may cast that card" permission in exile: the first half may be cast
// with it, but the Aftermath half can't be cast from any zone other than a
// graveyard, normally or without paying its mana cost.
func TestAftermathHalfCannotBeCastFromExileByAnotherPermission(t *testing.T) {
	engine, g := newAftermathGame()
	cardID := addCardToExile(g, game.Player1, aftermathCard())
	target := addCreaturePermanent(g, game.Player2)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCastFromZone,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		CastFromZone:   zone.Exile,
		AffectedCardID: cardID,
	})

	afterFromExile := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceAlternate, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
	legal := engine.legalActions(g, game.Player1)
	if actionsContain(legal, afterFromExile) {
		t.Fatalf("legal actions = %+v, did not want the Aftermath half from exile", legal)
	}
	card, _ := g.GetCardInstance(cardID)
	for _, option := range engine.freeCastOptionsForCard(g, game.Player1, card, zone.Exile, game.Selection{}, opt.V[int]{}) {
		if option.cast.Face == game.FaceAlternate {
			t.Fatal("a free cast from exile offered the Aftermath half")
		}
	}
	if engine.applyAction(g, game.Player1, afterFromExile) {
		t.Fatal("the Aftermath half was cast from exile")
	}
	front := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	if !actionsContain(legal, front) || !engine.applyAction(g, game.Player1, front) {
		t.Fatal("the exile permission did not let the first half be cast")
	}
}

// TestAftermathGraveyardPermissionCastsEitherHalf gives the card's owner a
// separate "you may cast that card from your graveyard" permission. Either half
// may then be cast from the graveyard: the first half, which lacks Aftermath,
// goes to the graveyard as usual, while the Aftermath half is still exiled.
func TestAftermathGraveyardPermissionCastsEitherHalf(t *testing.T) {
	engine, g := newAftermathGame()
	cardID := addCardToGraveyard(g, game.Player1, aftermathCard())
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	addBasicLandPermanent(g, game.Player1, types.Plains)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	target := addCreaturePermanent(g, game.Player2)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCastFromZone,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		CastFromZone:   zone.Graveyard,
		AffectedCardID: cardID,
	})

	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Graveyard, game.FaceFront, nil, 0, nil)) {
		t.Fatal("the graveyard permission did not let the first half be cast")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	player := g.Players[game.Player1]
	if !player.Graveyard.Contains(cardID) || player.Exile.Contains(cardID) {
		t.Fatal("the first half cast from the graveyard did not return to the graveyard")
	}

	castAftermathHalfFromGraveyard(t, engine, g, cardID, target)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !player.Exile.Contains(cardID) || player.Graveyard.Contains(cardID) {
		t.Fatal("the Aftermath half cast from the graveyard was not exiled")
	}
}

// TestAftermathHalfCastFreeFromGraveyardIsExiled casts the Aftermath half from
// its owner's graveyard without paying its mana cost through a resolving
// effect's sorcery-only free cast; it is still exiled when it resolves.
func TestAftermathHalfCastFreeFromGraveyardIsExiled(t *testing.T) {
	engine, g := newAftermathGame()
	cardID := addCardToGraveyard(g, game.Player1, aftermathCard())
	target := addCreaturePermanent(g, game.Player2)
	card, _ := g.GetCardInstance(cardID)
	sorcery := game.Selection{RequiredTypesAny: []types.Card{types.Sorcery}}

	options := engine.freeCastOptionsForCard(g, game.Player1, card, zone.Graveyard, sorcery, opt.V[int]{})
	if !slices.ContainsFunc(options, func(option freeCastOption) bool { return option.cast.Face == game.FaceAlternate }) {
		t.Fatalf("free cast options = %+v, want the Aftermath half from the graveyard", options)
	}
	player := g.Players[game.Player1]
	if !engine.chooseAndCastFreeSpellFromSource(g, player, game.Player1, cardID, zone.Graveyard,
		sorcery, opt.V[int]{}, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("the free cast of the Aftermath half failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Face != game.FaceAlternate || obj.SourceZone != zone.Graveyard {
		t.Fatalf("stack object = %+v, want the Aftermath half cast from the graveyard", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !target.Tapped || !player.Exile.Contains(cardID) || player.Graveyard.Contains(cardID) {
		t.Fatal("the free Aftermath half did not resolve and get exiled")
	}
}

// TestAftermathExileSurvivesGameClone clones the game with the Aftermath half
// on the stack; the clone's spell is exiled when it resolves, and the original
// game is unaffected.
func TestAftermathExileSurvivesGameClone(t *testing.T) {
	engine, g := newAftermathGame()
	cardID := addCardToGraveyard(g, game.Player1, aftermathCard())
	target := addCreaturePermanent(g, game.Player2)
	castAftermathHalfFromGraveyard(t, engine, g, cardID, target)

	clone := g.Clone()
	engine.resolveTopOfStack(clone, &TurnLog{})

	if !clone.Players[game.Player1].Exile.Contains(cardID) || clone.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("the cloned Aftermath half was not exiled")
	}
	if g.Stack.Size() != 1 || g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("resolving the clone changed the original game")
	}
}
