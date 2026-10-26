package game

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func fuseTestHalf(name string, manaCost cost.Mana, c color.Color, cardType types.Card, statics ...StaticAbility) CardFace {
	return CardFace{
		Name:            name,
		ManaCost:        opt.Val(manaCost),
		Colors:          []color.Color{c},
		Types:           []types.Card{cardType},
		StaticAbilities: statics,
		SpellAbility:    opt.Val(Mode{Sequence: []Instruction{{Primitive: Destroy{Object: TargetPermanentReference(0)}}}}.Ability()),
		OracleText:      name + " text.",
	}
}

func fuseTestCard(left, right CardFace) *CardDef {
	return &CardDef{CardFace: left, Layout: LayoutSplit, Alternate: opt.Val(right)}
}

func TestFusedFaceCombinesBothHalves(t *testing.T) {
	card := fuseTestCard(
		fuseTestHalf("Left", cost.Mana{cost.O(1), cost.R}, color.Red, types.Instant, FuseStaticBody),
		fuseTestHalf("Right", cost.Mana{cost.O(2), cost.W}, color.White, types.Sorcery, FuseStaticBody),
	)
	if !card.CanFuse() || !card.CanChooseCastFace(FaceFused) {
		t.Fatal("fuse split card cannot choose FaceFused")
	}
	if got := card.FaceIndexes(); slices.Contains(got, FaceFused) {
		t.Fatalf("FaceIndexes() = %v, want printed faces only", got)
	}
	if got := card.LegalCastFaces(); slices.Contains(got, FaceFused) {
		t.Fatalf("LegalCastFaces() = %v, want fused offered only by the hand cast route", got)
	}

	fused, ok := card.FaceDefView(FaceFused)
	if !ok {
		t.Fatal("FaceDefView(FaceFused) = false, want true")
	}
	if fused.Name != "Left // Right" || card.FaceLabel(FaceFused) != "Left // Right (fused)" || card.FaceLabel(FaceAlternate) != "Right" {
		t.Fatalf("fused name = %q, labels %q / %q", fused.Name, card.FaceLabel(FaceFused), card.FaceLabel(FaceAlternate))
	}
	if fused.ManaValue() != 5 || len(fused.ManaCost.Val) != 4 {
		t.Fatalf("fused mana cost = %v (value %d), want {1}{R}{2}{W} with value 5", fused.ManaCost.Val, fused.ManaValue())
	}
	if !slices.Equal(fused.Colors, []color.Color{color.White, color.Red}) {
		t.Fatalf("fused colors = %v, want white and red", fused.Colors)
	}
	if !slices.Equal(fused.Types, []types.Card{types.Instant, types.Sorcery}) {
		t.Fatalf("fused types = %v, want instant and sorcery", fused.Types)
	}
	content := fused.SpellAbility.Val
	if len(content.Modes) != 2 || content.MinModes != 2 || content.MaxModes != 2 || !slices.Equal(FusedModes(), []int{0, 1}) {
		t.Fatalf("fused spell body = %+v, want both halves locked as modes 0 and 1", content)
	}
	if len(card.CardFace.ManaCost.Val) != 2 || len(card.Alternate.Val.ManaCost.Val) != 2 {
		t.Fatal("building the fused view mutated a printed half")
	}
}

func TestCanFuseRejectsUnsupportedSplitCards(t *testing.T) {
	plain := func() *CardDef {
		return fuseTestCard(
			fuseTestHalf("Left", cost.Mana{cost.R}, color.Red, types.Instant, FuseStaticBody),
			fuseTestHalf("Right", cost.Mana{cost.W}, color.White, types.Instant, FuseStaticBody),
		)
	}
	for name, mutate := range map[string]func(*CardDef){
		"no fuse keyword": func(c *CardDef) {
			c.StaticAbilities = nil
			c.Alternate.Val.StaticAbilities = nil
		},
		"not split layout": func(c *CardDef) { c.Layout = LayoutAdventure },
		"X cost": func(c *CardDef) {
			c.ManaCost = opt.Val(cost.Mana{cost.X, cost.R})
		},
		"modal half": func(c *CardDef) {
			c.Alternate.Val.SpellAbility.Val.MaxModes = 2
		},
		"permanent half": func(c *CardDef) { c.Types = []types.Card{types.Creature} },
		"legacy implementation": func(c *CardDef) {
			c.Alternate.Val.ImplementationID = "legacy"
		},
	} {
		t.Run(name, func(t *testing.T) {
			card := plain()
			if !card.CanFuse() {
				t.Fatal("baseline fuse card cannot fuse")
			}
			mutate(card)
			if card.CanFuse() || card.CanChooseCastFace(FaceFused) {
				t.Fatal("CanFuse = true, want the fused choice withheld")
			}
			if _, ok := card.FaceView(FaceFused); ok {
				t.Fatal("FaceView(FaceFused) = true, want false")
			}
		})
	}
}
