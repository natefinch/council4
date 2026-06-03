package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCardDefDefaultFaceUsesFrontFace(t *testing.T) {
	card := &CardDef{CardFace: CardFace{Name: "Front Name",

		ManaCost: opt.Val(cost.Mana{cost.U}),
		Colors:   []color.Color{color.Blue},

		Types: []types.Card{types.Instant}}, Layout: LayoutModalDFC,

		ColorIdentity: color.NewIdentity(color.Blue, color.Green),

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
	card := &CardDef{CardFace: CardFace{Name: "Transforming Land",

		Types: []types.Card{types.Land}}, Layout: LayoutTransform,

		Back: opt.Val(CardFace{Name: "Large Creature", Types: []types.Card{types.Creature}}),
	}

	if !card.CanChooseLandFace(FaceFront) {
		t.Fatal("front land face of transform card was not playable")
	}
	if card.CanChooseCastFace(FaceBack) || card.CanChooseLandFace(FaceBack) {
		t.Fatal("transform back face should not be a cast/play choice")
	}
}

func TestCardFaceAbilityDefsIncludesCategorizedAbilities(t *testing.T) {
	face := CardFace{
		SpellAbility: opt.Val(SpellAbilityBody{
			Text:    "Draw a card.",
			Content: PlainAbilityContent{Sequence: []Effect{{Type: EffectDraw}}},
		}),
		ManaAbilities: []ManaAbilityBody{{
			Text:     "Add one mana.",
			Sequence: []Effect{{Type: EffectAddMana}},
		}},
		TriggeredAbilities: []TriggeredAbilityBody{{
			Text: "When this enters, draw a card.",
			Trigger: TriggerCondition{
				Pattern: TriggerPattern{Event: EventPermanentEnteredBattlefield},
			},
			Content: PlainAbilityContent{Sequence: []Effect{{Type: EffectDraw}}},
		}},
		ReplacementAbilities: []ReplacementAbilityDef{{
			Text:    "If this would die, exile it instead.",
			Effects: []Effect{{Type: EffectReplace}},
		}},
		StaticAbilities: []StaticAbilityBody{{
			Text:             "Flying",
			KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Flying}},
		}},
	}

	abilities := face.AbilityDefs()

	if len(abilities) != 5 {
		t.Fatalf("abilities = %d, want five categorized abilities", len(abilities))
	}
	if !abilities[0].IsSpell() || !abilities[1].IsMana() || !abilities[2].IsTriggered() || !abilities[3].IsStatic() || !abilities[4].IsStatic() {
		t.Fatalf("ability kinds = %+v, want spell/mana/triggered/replacement-static/static", abilities)
	}
	if !face.HasKeyword(Flying) {
		t.Fatal("categorized static keyword was not visible through HasKeyword")
	}
}
