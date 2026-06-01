package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

func TestCardDefDefaultFaceUsesFrontFace(t *testing.T) {
	card := &CardDef{
		Name:          "Front Name",
		Layout:        LayoutModalDFC,
		ColorIdentity: mana.NewColorIdentity(mana.Blue, mana.Green),
		Faces: []CardFace{
			{
				Name:      "Front Name",
				ManaCost:  opt.Val(mana.Cost{mana.ColoredMana(mana.Blue)}),
				ManaValue: 1,
				Colors:    []mana.Color{mana.Blue},
				Types:     []CardType{TypeInstant},
			},
			{
				Name:      "Back Name",
				ManaValue: 0,
				Types:     []CardType{TypeLand},
				Subtypes:  []string{LandSubtypeForest},
			},
		},
	}

	if !card.HasType(TypeInstant) || card.HasType(TypeLand) {
		t.Fatalf("default face types = %v, want front instant only", card.DefaultFace().Types)
	}
	if !card.CanChooseCastFace(FaceFront) || card.CanChooseCastFace(FaceBack) {
		t.Fatalf("cast face legality front/back = %v/%v, want true/false", card.CanChooseCastFace(FaceFront), card.CanChooseCastFace(FaceBack))
	}
	if !card.CanChooseLandFace(FaceBack) {
		t.Fatal("modal DFC land back face was not playable as a land")
	}
}

func TestTransformFrontLandCanBePlayedAsLand(t *testing.T) {
	card := &CardDef{
		Name:   "Transforming Land",
		Layout: LayoutTransform,
		Faces: []CardFace{
			{Name: "Transforming Land", Types: []CardType{TypeLand}},
			{Name: "Large Creature", Types: []CardType{TypeCreature}},
		},
	}

	if !card.CanChooseLandFace(FaceFront) {
		t.Fatal("front land face of transform card was not playable")
	}
	if card.CanChooseCastFace(FaceBack) || card.CanChooseLandFace(FaceBack) {
		t.Fatal("transform back face should not be a cast/play choice")
	}
}
