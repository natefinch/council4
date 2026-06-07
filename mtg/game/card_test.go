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

func TestCardFaceAbilityCountAndBodyAtUsesCanonicalOrder(t *testing.T) {
	face := CardFace{
		SpellAbility: opt.Val(Mode{Sequence: []Instruction{{Primitive: Draw{}}}}.Ability()),
		ManaAbilities: []ManaAbilityBody{{
			Text:    "Add one mana.",
			Content: Mode{Sequence: []Instruction{{Primitive: AddMana{}}}}.Ability(),
		}},
		TriggeredAbilities: []TriggeredAbilityBody{{
			Text: "When this enters, draw a card.",
			Trigger: TriggerCondition{
				Pattern: TriggerPattern{Event: EventPermanentEnteredBattlefield},
			},
			Content: Mode{Sequence: []Instruction{{Primitive: Draw{}}}}.Ability(),
		}},
		ReplacementAbilities: []ReplacementAbilityBody{{
			Text: "If this would die, exile it instead.",
		}},
		StaticAbilities: []StaticAbilityBody{{
			Text:             "Flying",
			KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Flying}},
		}},
	}

	if face.AbilityCount() != 5 {
		t.Fatalf("ability count = %d, want five categorized abilities", face.AbilityCount())
	}
	if _, ok := face.BodyAt(0).(ModalAbilityContent); !ok {
		t.Fatalf("BodyAt(0) = %T, want ModalAbilityContent", face.BodyAt(0))
	}
	if _, ok := face.BodyAt(1).(ManaAbilityBody); !ok {
		t.Fatalf("BodyAt(1) = %T, want ManaAbilityBody", face.BodyAt(1))
	}
	if _, ok := face.BodyAt(2).(TriggeredAbilityBody); !ok {
		t.Fatalf("BodyAt(2) = %T, want TriggeredAbilityBody", face.BodyAt(2))
	}
	if _, ok := face.BodyAt(3).(ReplacementAbilityBody); !ok {
		t.Fatalf("BodyAt(3) = %T, want ReplacementAbilityBody", face.BodyAt(3))
	}
	if _, ok := face.BodyAt(4).(StaticAbilityBody); !ok {
		t.Fatalf("BodyAt(4) = %T, want StaticAbilityBody", face.BodyAt(4))
	}
	if !face.HasKeyword(Flying) {
		t.Fatal("categorized static keyword was not visible through HasKeyword")
	}
}

func TestCardFaceCloneClonesSpellCosts(t *testing.T) {
	face := CardFace{
		AdditionalCosts: []cost.Additional{{
			Kind: cost.AdditionalPayLife,
			Text: "Pay 2 life",
		}},
		AlternativeCosts: []cost.Alternative{{
			Label:    "Flashback",
			ManaCost: opt.Val(cost.Mana{cost.O(3), cost.U}),
			AdditionalCosts: []cost.Additional{{
				Kind: cost.AdditionalDiscard,
				Text: "Discard a card",
			}},
		}},
		SpellAbility: opt.Val(ModalAbilityContent{}),
	}

	cloned := face.clone()

	face.AdditionalCosts[0].Text = "changed"
	face.AlternativeCosts[0].Label = "changed"
	face.AlternativeCosts[0].ManaCost.Val[0] = cost.O(9)
	face.AlternativeCosts[0].AdditionalCosts[0].Text = "changed"
	if cloned.AdditionalCosts[0].Text != "Pay 2 life" ||
		cloned.AlternativeCosts[0].Label != "Flashback" ||
		cloned.AlternativeCosts[0].ManaCost.Val[0] != cost.O(3) ||
		cloned.AlternativeCosts[0].AdditionalCosts[0].Text != "Discard a card" {
		t.Fatalf("cloned face costs alias source costs: %+v %+v", cloned.AdditionalCosts, cloned.AlternativeCosts)
	}
}

func TestClearAbilitiesRemovesCategorizedAbilities(t *testing.T) {
	face := CardFace{
		SpellAbility:         opt.Val(ModalAbilityContent{}),
		ActivatedAbilities:   []ActivatedAbilityBody{{Text: "Act"}},
		ManaAbilities:        []ManaAbilityBody{{Text: "Mana"}},
		LoyaltyAbilities:     []LoyaltyAbilityBody{{Text: "Loyal"}},
		TriggeredAbilities:   []TriggeredAbilityBody{{Text: "Trig"}},
		ReplacementAbilities: []ReplacementAbilityBody{{Text: "Replace"}},
		StaticAbilities:      []StaticAbilityBody{{Text: "Static"}},
	}

	face.ClearAbilities()

	if face.AbilityCount() != 0 {
		t.Fatalf("AbilityCount = %d, want 0", face.AbilityCount())
	}
	if face.SpellAbility.Exists || len(face.ActivatedAbilities) != 0 || len(face.ManaAbilities) != 0 ||
		len(face.LoyaltyAbilities) != 0 || len(face.TriggeredAbilities) != 0 ||
		len(face.ReplacementAbilities) != 0 || len(face.StaticAbilities) != 0 {
		t.Fatalf("ClearAbilities did not clear categorized fields: %+v", face)
	}
}
