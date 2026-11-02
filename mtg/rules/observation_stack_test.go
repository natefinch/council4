package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func pushSpell(g *game.Game, controller game.PlayerID, def *game.CardDef, faceDown bool) {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: controller}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Controller: controller,
		FaceDown:   faceDown,
	})
}

func TestStackObjectViewExposesSpellCharacteristics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name:     "Hill Giant",
		Types:    []types.Card{types.Creature},
		Colors:   []color.Color{color.Red},
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
	}}
	pushSpell(g, game.Player2, def, false)

	stack := observe(g, game.Player1).Stack()
	if len(stack) != 1 {
		t.Fatalf("Stack() = %d objects, want 1", len(stack))
	}
	view := stack[0]
	if view.ManaValue != 3 {
		t.Errorf("ManaValue = %d, want 3", view.ManaValue)
	}
	if !slices.Contains(view.Types, types.Creature) {
		t.Errorf("Types = %v, want to contain Creature", view.Types)
	}
	if !slices.Contains(view.Colors, color.Red) {
		t.Errorf("Colors = %v, want to contain Red", view.Colors)
	}
}

func TestStackObjectViewRedactsFaceDownSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name:     "Hidden Morph",
		Types:    []types.Card{types.Creature},
		Colors:   []color.Color{color.Green},
		ManaCost: opt.Val(cost.Mana{cost.O(5)}),
	}}
	pushSpell(g, game.Player2, def, true)

	view := observe(g, game.Player1).Stack()[0]
	if view.Name != "" {
		t.Errorf("Name = %q, want empty for a face-down spell", view.Name)
	}
	if view.ManaValue != 0 || !slices.Equal(view.Types, []types.Card{types.Creature}) || view.Colors != nil {
		t.Errorf("face-down spell view: MV=%d Types=%v Colors=%v, want colorless creature with MV 0", view.ManaValue, view.Types, view.Colors)
	}
	if view.FaceDownCard.Exists {
		t.Error("opponent sees the face-down spell's hidden card")
	}
}

// Source review: pinned corpus Gravel Slinger
// 87551307-6b5f-4f12-aa1f-4beebefad3b3 ({3}{W} Creature — Human Soldier;
// "{T}: This creature deals 1 damage to target attacking or blocking
// creature." and Morph {1}{W}; no rulings at
// https://api.scryfall.com/cards/87551307-6b5f-4f12-aa1f-4beebefad3b3/rulings)
// and Biotransference 75a787a3-5838-4121-bbaf-745d22c4b0bc ({2}{B}{B}
// Enchantment; "Creatures you control are artifacts in addition to their other
// types. The same is true for creature spells you control and creature cards
// you own that aren't on the battlefield." plus its artifact-spell cast
// trigger; one ruling, 2022-10-07, about enter-the-battlefield replacement
// order). CR 2026-09-25 read: 708.2/708.2a (a face-down spell has only the
// listed characteristics: a 2/2 creature with no text, name, subtypes or mana
// cost), 708.4 (effects that care about a spell's characteristics see only the
// face-down characteristics), 611.3a (a static ability's effect is not locked
// in) and 613.1d (type-changing effects apply in layer 4).

// TestStackObjectViewFaceDownSpellPublicTypes is the #3526 scenario: a
// face-down Gravel Slinger spell shows both seats its public face-down
// characteristics, including Biotransference's artifact type while
// Biotransference is on the battlefield, and never its hidden name, mana value,
// colors or private card to the opponent. Its controller keeps the authorized
// FaceDownCard view.
func TestStackObjectViewFaceDownSpellPublicTypes(t *testing.T) {
	for _, withModifier := range []bool{false, true} {
		t.Run(map[bool]string{false: "without Biotransference", true: "with Biotransference"}[withModifier], func(t *testing.T) {
			g, engine := spellTypeScenario(t)
			var modifier *game.Permanent
			if withModifier {
				modifier = addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
			}
			cardID := addCardToHand(g, game.Player1, productionChoiceCard(t, "Gravel Slinger"))
			g.Players[game.Player1].ManaPool.Add(mana.C, 3)
			if !engine.applyAction(g, game.Player1, actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownMorph)) {
				t.Fatal("face-down Gravel Slinger cast failed")
			}
			obj, ok := g.Stack.Peek()
			if !ok || !obj.FaceDown || obj.SourceID != cardID {
				t.Fatalf("stack top = %+v, want face-down Gravel Slinger", obj)
			}

			want := []types.Card{types.Creature}
			if withModifier {
				want = []types.Card{types.Creature, types.Artifact}
			}
			assertFaceDownSpellView(t, g, obj, cardID, want)

			if !withModifier {
				return
			}
			removePermanentForSpellTypeTest(t, g, engine, modifier)
			if top, ok := g.Stack.Peek(); !ok || top.ID != obj.ID {
				t.Fatal("face-down spell is no longer on top of the stack")
			}
			assertFaceDownSpellView(t, g, obj, cardID, []types.Card{types.Creature})
		})
	}
}

func assertFaceDownSpellView(t *testing.T, g *game.Game, obj *game.StackObject, cardID id.ID, want []types.Card) {
	t.Helper()
	for _, viewer := range []game.PlayerID{game.Player1, game.Player2} {
		var view StackObjectView
		found := false
		for _, candidate := range observe(g, viewer).Stack() {
			if candidate.ID == obj.ID {
				view, found = candidate, true
			}
		}
		if !found {
			t.Fatalf("viewer %d: face-down spell missing from stack view", viewer)
		}
		if !view.FaceDown || view.Name != "" || view.ManaValue != 0 || view.Colors != nil {
			t.Errorf("viewer %d: face-down spell view FaceDown=%v Name=%q MV=%d Colors=%v, want nameless colorless MV 0", viewer, view.FaceDown, view.Name, view.ManaValue, view.Colors)
		}
		if !sameTypes(view.Types, want) {
			t.Errorf("viewer %d: public types = %v, want %v", viewer, view.Types, want)
		}
		switch viewer {
		case game.Player1:
			if !view.FaceDownCard.Exists || view.FaceDownCard.Val.CardInstanceID != cardID || view.FaceDownCard.Val.Name != "Gravel Slinger" {
				t.Errorf("controller FaceDownCard = %+v, want Gravel Slinger", view.FaceDownCard)
			}
		default:
			if view.FaceDownCard.Exists {
				t.Errorf("viewer %d sees the hidden card %+v", viewer, view.FaceDownCard.Val)
			}
		}
	}
}

func sameTypes(got, want []types.Card) bool {
	if len(got) != len(want) {
		return false
	}
	for _, typ := range want {
		if !slices.Contains(got, typ) {
			return false
		}
	}
	return true
}
