package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCardDefDefaultFaceUsesFrontFace(t *testing.T) {
	card := &CardDef{
		Name:          "Front Name",
		Layout:        LayoutModalDFC,
		ManaCost:      opt.Val(mana.Cost{mana.U}),
		Colors:        []color.Color{color.Blue},
		ColorIdentity: mana.NewColorIdentity(color.Blue, color.Green),
		Types:         []types.Card{types.Instant},
		Back: opt.Val(CardFace{
			Name:     "Back Name",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest},
		}),
	}

	if !card.HasType(types.Instant) || card.HasType(types.Land) {
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
		Types:  []types.Card{types.Land},
		Back:   opt.Val(CardFace{Name: "Large Creature", Types: []types.Card{types.Creature}}),
	}

	if !card.CanChooseLandFace(FaceFront) {
		t.Fatal("front land face of transform card was not playable")
	}
	if card.CanChooseCastFace(FaceBack) || card.CanChooseLandFace(FaceBack) {
		t.Fatal("transform back face should not be a cast/play choice")
	}
}
