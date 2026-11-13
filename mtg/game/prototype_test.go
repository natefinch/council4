package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func validPrototypeCard() *CardDef {
	return &CardDef{CardFace: CardFace{
		Name:      "Prototype Validation Automaton",
		ManaCost:  opt.Val(cost.Mana{cost.O(7)}),
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(PT{Value: 6}),
		Toughness: opt.Val(PT{Value: 4}),
		Prototype: opt.Val(PrototypeProfile{
			ManaCost:  cost.Mana{cost.O(1), cost.B, cost.R},
			Power:     PT{Value: 3},
			Toughness: PT{Value: 2},
		}),
	}}
}

// TestPrototypeFaceReplacesOnlyCostColorAndPowerToughness checks the synthetic
// prototyped face against CR 719.3b and 719.5: only the mana cost, colors and
// power/toughness change, and the card's normal face is untouched.
func TestPrototypeFaceReplacesOnlyCostColorAndPowerToughness(t *testing.T) {
	card := validPrototypeCard()
	if !card.CanChooseCastFace(FacePrototype) {
		t.Fatal("a prototype creature card can't choose the prototyped cast")
	}
	face, ok := card.FaceView(FacePrototype)
	if !ok {
		t.Fatal("prototype face unavailable")
	}
	if face.Name != card.Name || len(face.Types) != 2 || face.ManaValue() != 3 {
		t.Fatalf("prototype face = %q types %v MV %d", face.Name, face.Types, face.ManaValue())
	}
	if len(face.Colors) != 2 || face.Colors[0] != color.Black || face.Colors[1] != color.Red {
		t.Fatalf("prototype colors = %v, want black and red", face.Colors)
	}
	if face.Power.Val.Value != 3 || face.Toughness.Val.Value != 2 {
		t.Fatalf("prototype P/T = %d/%d, want 3/2", face.Power.Val.Value, face.Toughness.Val.Value)
	}
	if card.ManaValue() != 7 || len(card.Colors) != 0 || card.Power.Val.Value != 6 {
		t.Fatal("building the prototype face changed the card's normal characteristics")
	}
	if !face.Prototype.Exists {
		t.Fatal("the prototyped face lost its copiable prototype profile (CR 719.2a)")
	}
	if card.CanChooseCastFace(FaceFused) || (&CardDef{CardFace: CardFace{Name: "Plain"}}).CanChooseCastFace(FacePrototype) {
		t.Fatal("prototype choice offered for an unrelated face or card")
	}
}

func TestValidateCardDefPrototype(t *testing.T) {
	if issues := ValidateCardDef(validPrototypeCard()); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
	tests := []struct {
		name   string
		mutate func(*CardDef)
	}{
		{"noncreature", func(c *CardDef) { c.Types = []types.Card{types.Artifact} }},
		{"empty prototype cost", func(c *CardDef) { c.Prototype.Val.ManaCost = nil }},
		{"negative generic", func(c *CardDef) { c.Prototype.Val.ManaCost = cost.Mana{cost.O(-1), cost.R} }},
		{"same-color hybrid", func(c *CardDef) { c.Prototype.Val.ManaCost = cost.Mana{cost.HybridMana(mana.R, mana.R)} }},
		{"unprintable phyrexian generic", func(c *CardDef) { c.Prototype.Val.ManaCost = cost.Mana{cost.PhyrexianGeneric(2)} }},
		{"negative toughness", func(c *CardDef) { c.Prototype.Val.Toughness = PT{Value: -1} }},
		{"star power", func(c *CardDef) { c.Prototype.Val.Power = PT{IsStar: true} }},
		{"back face", func(c *CardDef) {
			c.Layout = LayoutTransform
			c.Back = opt.Val(CardFace{Name: "Back", Types: []types.Card{types.Creature}})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := validPrototypeCard()
			tt.mutate(card)
			found := false
			for _, issue := range ValidateCardDef(card) {
				if issue.Code == CardDefIssueInvalidPrototype {
					found = true
				}
			}
			if !found {
				t.Fatal("invalid prototype accepted")
			}
			if card.CanChooseCastFace(FacePrototype) {
				t.Fatal("invalid prototype offered as a cast choice")
			}
		})
	}
}
