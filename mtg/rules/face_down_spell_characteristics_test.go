package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: issue #3374. CR 2026-09-25 (verified copy, effective
// September 25, 2026): 702.37c and 702.168b make a face-down morph/disguise
// spell a 2/2 creature with no text, name, subtypes or mana cost (disguise adds
// ward {2}); 708.2 says those are its only characteristics, 708.4 that effects
// caring about a spell's characteristics see only the face-down ones, and 708.5
// that only its controller may look at it; 702.16a-c and 702.21a read
// the source's characteristics for protection and ward. Default-registry cards
// checked against the pinned corpus: Birchlore Rangers
// 8ce3a3a1-3569-4909-a604-f78d4888781e (green Elf Druid Ranger, Morph {G}),
// Aveline de Grandpré db3b4f73-e241-4d60-a877-25327092c291 (Legendary green
// Human Assassin, Disguise {B}{G}; 2024-07-05 rulings: the face-down spell is a
// colorless 2/2 with ward {2} and no name, mana cost or creature types),
// Akroma, Angel of Fury 2b5ac7bd-3788-4c8b-b945-c1105986807a ("This spell
// can't be countered", Morph; 2021-03-19 ruling: cast face down it can be
// countered), Willbender 6e6a3551-250c-4563-8c30-6d8f49e6429f (blue Human
// Wizard, Morph), Zoetic Cavern 37f10035-bf05-460d-9390-433caa2570f4 (Land,
// Morph {2}), Deathgrip fc86164d-24f1-4fad-a358-f61dbfd86bd9 ("Counter target
// green spell"),
// Ceremonious Rejection 08c5ed8e-4804-4042-8a1d-ad24c6846816 ("Counter target
// colorless spell"), Tale's End 1421115b-9a98-4ab2-bcb2-7d8899ce12db ("...or
// legendary spell") and Pyroblast b029eb9a-dd7a-40c2-96c4-0063d9cc002c
// ("Counter target spell if it's blue"; 2016-06-08 ruling: blueness is checked
// on resolution). Goldbug, Scrappy Scout's "Human spells you control can't be
// countered" supplies the spell-subtype qualifier.

// pushFaceUpCardSpell puts a face-up spell for def on the stack as a control
// for the face-down casts.
func pushFaceUpCardSpell(g *game.Game, controller game.PlayerID, def *game.CardDef) *game.StackObject {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: controller}
	obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: cardID, Controller: controller}
	g.Stack.Push(obj)
	return obj
}

func stackTargetCandidateIDs(g *game.Game, controller game.PlayerID, source *game.CardDef, spec game.TargetSpec) map[id.ID]bool {
	got := make(map[id.ID]bool)
	for _, candidate := range targetCandidatesForSpec(g, controller, source, 0, game.Event{}, &spec) {
		if candidate.Kind == game.TargetStackObject {
			got[candidate.StackObjectID] = true
		}
	}
	return got
}

func assertSpellTargetable(t *testing.T, g *game.Game, cardName string, source *game.CardDef, spec game.TargetSpec, faceDown, faceUp *game.StackObject, wantFaceDown, wantFaceUp bool) {
	t.Helper()
	got := stackTargetCandidateIDs(g, game.Player2, source, spec)
	if got[faceDown.ID] != wantFaceDown {
		t.Errorf("%s can target face-down spell = %v, want %v", cardName, got[faceDown.ID], wantFaceDown)
	}
	if got[faceUp.ID] != wantFaceUp {
		t.Errorf("%s can target face-up spell = %v, want %v", cardName, got[faceUp.ID], wantFaceUp)
	}
}

func castFaceDownForOpponentResponse(t *testing.T, cardName string, kind game.FaceDownKind) (*game.Game, *Engine, *game.StackObject) {
	t.Helper()
	g, engine, cardID := faceDownCostGame(t, cardName, mana.G, 3)
	obj := castFaceDownPaying(t, engine, g, cardID, kind, 3)
	return g, engine, obj
}

func TestFaceDownMorphSpellTargetQualifiersUseFaceDownCharacteristics(t *testing.T) {
	t.Parallel()
	g, _, faceDown := castFaceDownForOpponentResponse(t, "Birchlore Rangers", game.FaceDownMorph)
	faceUp := pushFaceUpCardSpell(g, game.Player1, productionChoiceCard(t, "Birchlore Rangers"))

	deathgrip := productionChoiceCard(t, "Deathgrip")
	assertSpellTargetable(t, g, "Deathgrip", deathgrip,
		deathgrip.ActivatedAbilities[0].Content.Modes[0].Targets[0], faceDown, faceUp, false, true)
	rejection := productionChoiceCard(t, "Ceremonious Rejection")
	assertSpellTargetable(t, g, "Ceremonious Rejection", rejection,
		rejection.SpellAbility.Val.Modes[0].Targets[0], faceDown, faceUp, true, false)

	chars, ok := stackObjectSourceChars(g, faceDown)
	if !ok || len(chars.colors) != 0 || len(chars.subtypes) != 0 || !slices.Equal(chars.types, []types.Card{types.Creature}) {
		t.Fatalf("face-down source characteristics = %+v (ok %v), want colorless Creature with no subtypes", chars, ok)
	}
	if colors, _ := stackObjectColors(g, faceDown); len(colors) != 0 {
		t.Fatalf("face-down spell colors = %v, want colorless", colors)
	}
}

func TestFaceDownDisguiseSpellHidesLegendaryGreenAndHuman(t *testing.T) {
	t.Parallel()
	g, _, faceDown := castFaceDownForOpponentResponse(t, "Aveline de Grandpré", game.FaceDownDisguise)
	faceUp := pushFaceUpCardSpell(g, game.Player1, productionChoiceCard(t, "Aveline de Grandpré"))

	talesEnd := productionChoiceCard(t, "Tale's End")
	assertSpellTargetable(t, g, "Tale's End", talesEnd,
		talesEnd.SpellAbility.Val.Modes[0].Targets[0], faceDown, faceUp, false, true)
	deathgrip := productionChoiceCard(t, "Deathgrip")
	assertSpellTargetable(t, g, "Deathgrip", deathgrip,
		deathgrip.ActivatedAbilities[0].Content.Modes[0].Targets[0], faceDown, faceUp, false, true)
	rejection := productionChoiceCard(t, "Ceremonious Rejection")
	assertSpellTargetable(t, g, "Ceremonious Rejection", rejection,
		rejection.SpellAbility.Val.Modes[0].Targets[0], faceDown, faceUp, true, false)
	assertSpellTargetable(t, g.Clone(), "Tale's End on a cloned game", talesEnd,
		talesEnd.SpellAbility.Val.Modes[0].Targets[0], faceDown, faceUp, false, true)

	if supers := stackSpellSupertypes(g, faceDown); len(supers) != 0 {
		t.Fatalf("face-down supertypes = %v, want none", supers)
	}
	chars, ok := stackObjectSourceChars(g, faceDown)
	if !ok || len(chars.colors) != 0 || len(chars.subtypes) != 0 {
		t.Fatalf("face-down source characteristics = %+v, want colorless with no subtypes", chars)
	}
	upChars, _ := stackObjectSourceChars(g, faceUp)
	if !slices.Contains(upChars.subtypes, types.Human) || !slices.Contains(upChars.colors, color.Green) {
		t.Fatalf("face-up source characteristics = %+v, want green Human", upChars)
	}
}

func TestFaceDownHumanSpellIsNotAHumanSpellForCantBeCountered(t *testing.T) {
	t.Parallel()
	g, _, faceDown := castFaceDownForOpponentResponse(t, "Aveline de Grandpré", game.FaceDownDisguise)
	faceUp := pushFaceUpCardSpell(g, game.Player1, productionChoiceCard(t, "Aveline de Grandpré"))
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:                 g.IDGen.Next(),
		Kind:               game.RuleEffectCantBeCountered,
		Controller:         game.Player1,
		AffectedController: game.ControllerYou,
		SpellSubtypes:      []types.Sub{types.Human},
		Duration:           game.DurationPermanent,
		CreatedTurn:        g.Turn.TurnNumber,
	})
	if !stackSpellCanBeCountered(g, faceDown) {
		t.Fatal("a face-down Human card is not a Human spell and must be counterable")
	}
	if stackSpellCanBeCountered(g, faceUp) {
		t.Fatal("the face-up Human spell must stay uncounterable")
	}
}

func TestFaceDownAkromaCanBeCountered(t *testing.T) {
	t.Parallel()
	g, engine, faceDown := castFaceDownForOpponentResponse(t, "Akroma, Angel of Fury", game.FaceDownMorph)
	if !stackSpellCanBeCountered(g, faceDown) {
		t.Fatal("face-down Akroma must not have \"This spell can't be countered\"")
	}
	rejectionID := addCardToHand(g, game.Player2, productionChoiceCard(t, "Ceremonious Rejection"))
	g.Players[game.Player2].ManaPool.Add(mana.U, 1)
	g.Turn.PriorityPlayer = game.Player2
	if !engine.applyAction(g, game.Player2, actionBuild.castSpell(rejectionID, zone.Hand, game.FaceFront,
		[]game.Target{game.StackObjectTarget(faceDown.ID)}, 0, nil)) {
		t.Fatal("Ceremonious Rejection could not target the colorless face-down spell")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if stackContainsObject(g, faceDown.ID) || !g.Players[game.Player1].Graveyard.Contains(faceDown.SourceID) {
		t.Fatal("face-down Akroma was not countered into its owner's graveyard")
	}

	faceUp := pushFaceUpCardSpell(g, game.Player1, productionChoiceCard(t, "Akroma, Angel of Fury"))
	if stackSpellCanBeCountered(g, faceUp) {
		t.Fatal("face-up Akroma must keep \"This spell can't be countered\"")
	}
}

func TestPyroblastDoesNotCounterFaceDownBlueSpell(t *testing.T) {
	t.Parallel()
	for _, faceDownCast := range []bool{true, false} {
		g, engine, faceDown := castFaceDownForOpponentResponse(t, "Willbender", game.FaceDownMorph)
		target := faceDown
		if !faceDownCast {
			target = pushFaceUpCardSpell(g, game.Player1, productionChoiceCard(t, "Willbender"))
		}
		pyroblastID := addCardToHand(g, game.Player2, productionChoiceCard(t, "Pyroblast"))
		g.Players[game.Player2].ManaPool.Add(mana.R, 1)
		g.Turn.PriorityPlayer = game.Player2
		if !engine.applyAction(g, game.Player2, actionBuild.castSpell(pyroblastID, zone.Hand, game.FaceFront,
			[]game.Target{game.StackObjectTarget(target.ID)}, 0, []int{0})) {
			t.Fatal("Pyroblast cast failed")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if countered := !stackContainsObject(g, target.ID); countered == faceDownCast {
			t.Fatalf("Pyroblast countered face-down=%v Willbender = %v, want %v", faceDownCast, countered, !faceDownCast)
		}
	}
}

func TestFaceDownSpellResolvedEventUsesFaceDownCharacteristics(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"Birchlore Rangers", "Zoetic Cavern"} {
		g, engine, obj := castFaceDownForOpponentResponse(t, name, game.FaceDownMorph)
		engine.resolveTopOfStack(g, &TurnLog{})
		var resolved game.Event
		for _, event := range g.Events {
			if event.Kind == game.EventSpellResolved && event.StackObjectID == obj.ID {
				resolved = event
			}
		}
		if resolved.Kind != game.EventSpellResolved || resolved.CardID != obj.SourceID {
			t.Fatalf("%s: no spell-resolved event for the face-down spell", name)
		}
		watcher := addCreaturePermanent(g, game.Player2)
		matches := func(sel game.Selection) bool {
			pattern := &game.TriggerPattern{Event: game.EventSpellResolved, CardSelection: sel}
			return triggerMatchesEventForController(g, watcher, game.Player2, pattern, resolved)
		}
		if !matches(game.Selection{RequiredTypes: []types.Card{types.Creature}}) {
			t.Fatalf("%s: face-down spell did not resolve as a creature spell", name)
		}
		for _, sel := range []game.Selection{
			{RequiredTypes: []types.Card{types.Land}},
			{SubtypesAny: []types.Sub{types.Elf}},
			{ColorsAny: []color.Color{color.Green}},
		} {
			if matches(sel) {
				t.Fatalf("%s: face-down spell-resolved event matched hidden characteristic %+v", name, sel)
			}
		}
	}
}

func TestFaceDownSpellObservationStillShowsControllerHiddenCard(t *testing.T) {
	t.Parallel()
	g, _, obj := castFaceDownForOpponentResponse(t, "Aveline de Grandpré", game.FaceDownDisguise)
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		stack := NewObservation(g, viewer).Stack()
		if len(stack) != 1 || stack[0].ID != obj.ID || !stack[0].FaceDown || stack[0].Name != "" ||
			len(stack[0].Colors) != 0 || len(stack[0].Types) != 0 {
			t.Fatalf("viewer %v public stack view = %+v", viewer, stack)
		}
		private := stack[0].FaceDownCard
		if viewer == game.Player1 && (!private.Exists || private.Val.Name != "Aveline de Grandpré") {
			t.Fatalf("controller private face-down card = %+v, want Aveline de Grandpré", private)
		}
		if viewer != game.Player1 && private.Exists {
			t.Fatalf("opponent sees face-down card %+v", private.Val)
		}
	}
}
