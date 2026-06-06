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
	for i := range abilities {
		if abilities[i].Body == nil {
			t.Fatalf("ability %d has nil Body: %+v", i, abilities[i])
		}
	}
	if !face.HasKeyword(Flying) {
		t.Fatal("categorized static keyword was not visible through HasKeyword")
	}
}

func TestCardFaceAbilityDefsPopulatesBodyForLegacyAbilities(t *testing.T) {
	face := CardFace{
		Abilities: []AbilityDef{
			{Kind: SpellAbility, Effects: []Effect{{Type: EffectDraw}}},
			{Kind: ActivatedAbility, ManaCost: opt.Val(cost.Mana{cost.G}), Effects: []Effect{{Type: EffectGainLife}}},
			{Kind: ActivatedAbility, IsManaAbility: true, Effects: []Effect{{Type: EffectAddMana}}},
			{Kind: ActivatedAbility, IsLoyaltyAbility: true, LoyaltyCost: 1, Effects: []Effect{{Type: EffectDraw}}},
			{Kind: TriggeredAbility, Trigger: opt.Val(TriggerCondition{Pattern: TriggerPattern{Event: EventPermanentEnteredBattlefield}}), Effects: []Effect{{Type: EffectDraw}}},
			{Kind: StaticAbility, KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Flying}}},
		},
	}

	abilities := face.AbilityDefs()

	if len(abilities) != len(face.Abilities) {
		t.Fatalf("abilities = %d, want %d", len(abilities), len(face.Abilities))
	}
	for i := range abilities {
		if abilities[i].Body == nil {
			t.Fatalf("ability %d has nil Body: %+v", i, abilities[i])
		}
	}
	if !abilities[0].IsSpell() || !abilities[1].IsActivated() || !abilities[2].IsMana() || !abilities[3].IsLoyalty() || !abilities[4].IsTriggered() || !abilities[5].IsStatic() {
		t.Fatalf("legacy abilities did not normalize to expected bodies: %+v", abilities)
	}
}

// TestWithAbilityBodiesNormalizesGrantedAbilities verifies that
// WithAbilityBodies lowers body-only AbilityDef entries nested inside
// ContinuousEffect.AddAbilities and Effect.EmblemAbilities, so that rules
// consumers see flat compatibility fields without hot-path allocation.
func TestWithAbilityBodiesNormalizesGrantedAbilities(t *testing.T) {
	grantedActivated := AbilityDef{Body: ActivatedAbilityBody{
		Text:     "{2}: This creature gets +1/+0 until end of turn.",
		ManaCost: opt.Val(cost.Mana{cost.O(2)}),
		Content: PlainAbilityContent{Sequence: []Effect{
			{Type: EffectModifyPT, PowerDelta: 1, UntilEndOfTurn: true, TargetIndex: TargetIndexSourcePermanent},
		}},
	}}
	emblemAbility := AbilityDef{Body: StaticAbilityBody{
		Text:             "You have emblematic power.",
		KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Flying}},
	}}

	card := &CardDef{CardFace: CardFace{
		StaticAbilities: []StaticAbilityBody{
			{
				Text: "Equipped creature has an ability.",
				Effects: []Effect{{
					Type: EffectApplyContinuous,
					ContinuousEffects: []ContinuousEffect{{
						Layer:        LayerAbility,
						Selector:     EffectSelectorEquippedCreature,
						AddAbilities: []AbilityDef{grantedActivated},
					}},
					EmblemAbilities: []AbilityDef{emblemAbility},
				}},
			},
		},
		SpellAbility: opt.Val(SpellAbilityBody{
			Text: "Choose one.",
			Content: ModalAbilityContent{Modes: []Mode{{
				Text: "Create an emblem.",
				Effects: []Effect{{
					Type:            EffectCreateEmblem,
					EmblemAbilities: []AbilityDef{emblemAbility},
				}},
			}}},
		}),
	}}

	sourceAbilities := card.AbilityDefs()
	sourceGranted := sourceAbilities[1].Effects[0].ContinuousEffects[0].AddAbilities[0]
	if sourceGranted.Kind != ActivatedAbility || len(sourceGranted.Effects) != 1 {
		t.Fatalf("slow-path granted ability was not normalized: %+v", sourceGranted)
	}
	sourceEmblem := sourceAbilities[0].Modes[0].Effects[0].EmblemAbilities[0]
	if sourceEmblem.Kind != StaticAbility || !sourceEmblem.HasKeyword(Flying) {
		t.Fatalf("slow-path emblem ability was not normalized: %+v", sourceEmblem)
	}

	normalized := card.WithAbilityBodies()

	if !normalized.hasCategorizedAbilities() {
		t.Fatal("normalized runtime card lost categorized abilities")
	}

	abilities := normalized.AbilityDefs()
	if len(abilities) != 2 {
		t.Fatalf("len(abilities) = %d, want 2", len(abilities))
	}
	static := &abilities[1]
	if len(static.Effects) != 1 {
		t.Fatalf("static.Effects len = %d, want 1", len(static.Effects))
	}
	flatEffect := &static.Effects[0]

	// Verify ContinuousEffect.AddAbilities were normalized.
	if len(flatEffect.ContinuousEffects) != 1 {
		t.Fatalf("ContinuousEffects len = %d, want 1", len(flatEffect.ContinuousEffects))
	}
	granted := &flatEffect.ContinuousEffects[0].AddAbilities
	if len(*granted) != 1 {
		t.Fatalf("AddAbilities len = %d, want 1", len(*granted))
	}
	ab := &(*granted)[0]
	if ab.Body == nil {
		t.Fatal("granted AddAbility has nil Body after normalization")
	}
	if !ab.IsActivated() {
		t.Fatal("granted AddAbility IsActivated() = false, want true")
	}
	if ab.Kind != ActivatedAbility {
		t.Fatalf("granted AddAbility flat Kind = %v, want ActivatedAbility", ab.Kind)
	}
	if len(ab.Effects) != 1 || ab.Effects[0].Type != EffectModifyPT {
		t.Fatalf("granted AddAbility flat Effects = %+v, want ModifyPT", ab.Effects)
	}

	// Verify EmblemAbilities were normalized.
	if len(flatEffect.EmblemAbilities) != 1 {
		t.Fatalf("EmblemAbilities len = %d, want 1", len(flatEffect.EmblemAbilities))
	}
	emblem := &flatEffect.EmblemAbilities[0]
	if emblem.Body == nil {
		t.Fatal("emblem ability has nil Body after normalization")
	}
	if !emblem.IsStatic() {
		t.Fatal("emblem ability IsStatic() = false, want true")
	}
	if emblem.Kind != StaticAbility {
		t.Fatalf("emblem ability flat Kind = %v, want StaticAbility", emblem.Kind)
	}
	if !emblem.HasKeyword(Flying) {
		t.Fatal("emblem ability HasKeyword(Flying) = false after normalization")
	}

	modalEmblem := &abilities[0].Modes[0].Effects[0].EmblemAbilities[0]
	if modalEmblem.Kind != StaticAbility || !modalEmblem.HasKeyword(Flying) {
		t.Fatalf("modal emblem ability was not normalized: %+v", modalEmblem)
	}

	// Verify source card is not mutated (body-only form preserved).
	srcBody := card.StaticAbilities[0]
	if srcBody.Effects[0].ContinuousEffects[0].AddAbilities[0].Kind != 0 {
		t.Fatal("source card AddAbility was mutated; original body should be unchanged")
	}
	srcModalBody := card.SpellAbility.Val
	srcModal, ok := srcModalBody.Content.(ModalAbilityContent)
	if !ok {
		t.Fatal("source card modal ability content is not ModalAbilityContent")
	}
	if srcModal.Modes[0].Effects[0].EmblemAbilities[0].Kind != 0 {
		t.Fatal("source card modal EmblemAbility was mutated; original body should be unchanged")
	}
}

func TestWithAbilityBodiesPreservesCategorizedAbilities(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		ReplacementAbilities: []ReplacementAbilityDef{
			EntersTappedReplacement("This permanent enters tapped."),
		},
	}}

	normalized := card.WithAbilityBodies()

	if len(normalized.ReplacementAbilities) != 1 {
		t.Fatalf("ReplacementAbilities len = %d, want 1", len(normalized.ReplacementAbilities))
	}
	if !normalized.abilitiesNormalized {
		t.Fatal("runtime card abilities were not marked normalized")
	}
	abilities := normalized.AbilityDefs()
	if len(abilities) != 1 || abilities[0].Body == nil {
		t.Fatalf("compatibility abilities = %+v, want one body-backed replacement", abilities)
	}
}
