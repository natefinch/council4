package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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

func TestCardDefAlternateFaceAdventure(t *testing.T) {
	card := &CardDef{
		CardFace: CardFace{Name: "Questing Squire"},
		Layout:   LayoutAdventure,
		Alternate: opt.Val(CardFace{
			Name:     "Seek the Way",
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
			Types:    []types.Card{types.Sorcery},
		}),
	}

	face, ok := card.AlternateFace()
	if !ok {
		t.Fatal("AlternateFace() reported absent alternate face")
	}
	if face.Name != "Seek the Way" ||
		!face.ManaCost.Exists ||
		len(face.Types) != 1 ||
		face.Types[0] != types.Sorcery {
		t.Fatalf("alternate face = %+v", face)
	}

	face, ok = card.Face(FaceAlternate)
	if !ok || face.Name != "Seek the Way" {
		t.Fatalf("Face(FaceAlternate) = %+v, %v", face, ok)
	}
	def, ok := card.FaceDef(FaceAlternate)
	if !ok || def.Name != "Seek the Way" {
		t.Fatalf("FaceDef(FaceAlternate) = %+v, %v", def, ok)
	}
	if got := card.FaceIndexes(); len(got) != 2 || got[0] != FaceFront || got[1] != FaceAlternate {
		t.Fatalf("FaceIndexes() = %v, want [FaceFront FaceAlternate]", got)
	}
	if !card.CanChooseCastFace(FaceAlternate) {
		t.Fatal("adventure alternate face was not castable")
	}
	if got := card.LegalCastFaces(); len(got) != 2 || got[0] != FaceFront || got[1] != FaceAlternate {
		t.Fatalf("LegalCastFaces() = %v, want [FaceFront FaceAlternate]", got)
	}
}

func TestCardDefFaceDefView(t *testing.T) {
	card := &CardDef{
		CardFace: CardFace{Name: "Front", Types: []types.Card{types.Creature}},
		Back: opt.Val(CardFace{
			Name:  "Back",
			Types: []types.Card{types.Land},
		}),
	}

	front, ok := card.FaceDefView(FaceFront)
	if !ok || front != card {
		t.Fatalf("FaceDefView(FaceFront) = %p, %v; want original card %p", front, ok, card)
	}
	back, ok := card.FaceDefView(FaceBack)
	if !ok {
		t.Fatal("FaceDefView(FaceBack) = false, want true")
	}
	if back.Name != "Back" || !back.HasType(types.Land) {
		t.Fatalf("FaceDefView(FaceBack) = %+v, want Back land", back)
	}
	backFace, ok := card.FaceView(FaceBack)
	if !ok || backFace != &card.Back.Val {
		t.Fatalf("FaceView(FaceBack) = %p, %v; want stored back face %p", backFace, ok, &card.Back.Val)
	}
}

func TestCardDefAlternateFaceAbsent(t *testing.T) {
	card := &CardDef{CardFace: CardFace{Name: "Ordinary Bear"}}

	if _, ok := card.AlternateFace(); ok {
		t.Fatal("AlternateFace() reported a face for single-faced card")
	}
}

func TestCardFaceToCardDefDeepClonesOverload(t *testing.T) {
	selection := Selection{RequiredTypes: []types.Card{types.Artifact}}
	face := CardFace{
		Name: "Overloaded Face",
		Overload: opt.Val(OverloadAbility{
			Cost: cost.Mana{cost.O(2), cost.R},
			SpellAbility: Mode{
				Sequence: []Instruction{{
					Description: "destroy artifacts",
					Primitive:   Destroy{Group: BattlefieldGroup(selection)},
				}},
			}.Ability(),
		}),
	}
	cloned := face.ToCardDef(&CardDef{})
	if !cloned.Overload.Exists {
		t.Fatal("ToCardDef omitted overload")
	}

	cloned.Overload.Val.Cost[0] = cost.O(9)
	cloned.Overload.Val.SpellAbility.Modes[0].Sequence[0].Description = "changed"
	destroy, ok := cloned.Overload.Val.SpellAbility.Modes[0].Sequence[0].Primitive.(Destroy)
	if !ok {
		t.Fatalf("cloned overload primitive = %T, want Destroy", cloned.Overload.Val.SpellAbility.Modes[0].Sequence[0].Primitive)
	}
	destroy.Group.selection.RequiredTypes[0] = types.Creature
	cloned.Overload.Val.SpellAbility.Modes[0].Sequence[0].Primitive = destroy

	originalDestroy, ok := face.Overload.Val.SpellAbility.Modes[0].Sequence[0].Primitive.(Destroy)
	if !ok {
		t.Fatalf("original overload primitive = %T, want Destroy", face.Overload.Val.SpellAbility.Modes[0].Sequence[0].Primitive)
	}
	if face.Overload.Val.Cost[0] != cost.O(2) ||
		face.Overload.Val.SpellAbility.Modes[0].Sequence[0].Description != "destroy artifacts" ||
		originalDestroy.Group.selection.RequiredTypes[0] != types.Artifact {
		t.Fatalf("mutating cloned overload changed original: %#v", face.Overload.Val)
	}
}

func TestCardFaceToCardDefDeepClonesCumulativeUpkeep(t *testing.T) {
	face := CardFace{
		Name: "Cumulative Face",
		TriggeredAbilities: []TriggeredAbility{
			CumulativeUpkeepTriggeredAbility(cost.Mana{cost.O(1), cost.U}),
		},
	}
	cloned := face.ToCardDef(&CardDef{})
	ability := &cloned.TriggeredAbilities[0]
	keyword, ok := ability.KeywordAbilities[0].(CumulativeUpkeepKeyword)
	if !ok {
		t.Fatalf("cloned keyword = %T; want CumulativeUpkeepKeyword", ability.KeywordAbilities[0])
	}
	keyword.Cost[0] = cost.O(9)
	ability.KeywordAbilities[0] = keyword
	pay, ok := ability.Content.Modes[0].Sequence[1].Primitive.(Pay)
	if !ok {
		t.Fatalf("cloned payment primitive = %T; want Pay", ability.Content.Modes[0].Sequence[1].Primitive)
	}
	pay.Payment.ManaCost.Val[0] = cost.O(9)
	pay.Payment.ManaCostMultiplier.Val.CounterKind = counter.Charge
	ability.Content.Modes[0].Sequence[1].Primitive = pay

	original := face.TriggeredAbilities[0]
	originalKeyword, ok := original.KeywordAbilities[0].(CumulativeUpkeepKeyword)
	if !ok {
		t.Fatalf("original keyword = %T; want CumulativeUpkeepKeyword", original.KeywordAbilities[0])
	}
	originalPay, ok := original.Content.Modes[0].Sequence[1].Primitive.(Pay)
	if !ok {
		t.Fatalf("original payment primitive = %T; want Pay", original.Content.Modes[0].Sequence[1].Primitive)
	}
	if originalKeyword.Cost[0] != cost.O(1) ||
		originalPay.Payment.ManaCost.Val[0] != cost.O(1) ||
		originalPay.Payment.ManaCostMultiplier.Val.CounterKind != counter.Age {
		t.Fatalf("mutating cloned cumulative upkeep changed original: %#v", original)
	}
}

func TestCardFaceToCardDefDeepClonesSacrificeConditionedReanimation(t *testing.T) {
	const resultKey = ResultKey("sacrifice-succeeded")
	face := CardFace{
		Name: "Twin Revival",
		SpellAbility: opt.Val(Mode{
			Targets: []TargetSpec{{
				MinTargets: 2,
				MaxTargets: 2,
				Allow:      TargetAllowCard,
				TargetZone: zone.Graveyard,
				Selection: opt.Val(Selection{
					RequiredTypes: []types.Card{types.Creature},
				}),
			}},
			Sequence: []Instruction{
				{
					Primitive: SacrificePermanents{
						Player:    ControllerReference(),
						Amount:    Fixed(1),
						Selection: Selection{RequiredTypes: []types.Card{types.Creature}},
					},
					PublishResult: resultKey,
				},
				{
					Primitive: PutOnBattlefield{
						Sources: []BattlefieldSource{
							CardBattlefieldSource(CardReference{Kind: CardReferenceTarget}),
							CardBattlefieldSource(CardReference{Kind: CardReferenceTarget, TargetIndex: 1}),
						},
						EntryTapped: true,
					},
					ResultGate: opt.Val(InstructionResultGate{
						Key:       resultKey,
						Succeeded: TriTrue,
					}),
				},
			},
		}.Ability()),
	}

	cloned := face.ToCardDef(&CardDef{})
	cloned.SpellAbility.Val.Modes[0].Targets[0].Selection.Val.RequiredTypes[0] = types.Artifact
	cloned.SpellAbility.Val.Modes[0].Sequence[0].PublishResult = "changed"
	sacrifice, ok := cloned.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(SacrificePermanents)
	if !ok {
		t.Fatalf("cloned primitive = %T; want SacrificePermanents", cloned.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	sacrifice.Selection.RequiredTypes[0] = types.Artifact
	cloned.SpellAbility.Val.Modes[0].Sequence[0].Primitive = sacrifice
	reanimation, ok := cloned.SpellAbility.Val.Modes[0].Sequence[1].Primitive.(PutOnBattlefield)
	if !ok {
		t.Fatalf("cloned primitive = %T; want PutOnBattlefield", cloned.SpellAbility.Val.Modes[0].Sequence[1].Primitive)
	}
	reanimation.Sources[0] = CardBattlefieldSource(CardReference{Kind: CardReferenceSource})
	cloned.SpellAbility.Val.Modes[0].Sequence[1].Primitive = reanimation

	originalMode := face.SpellAbility.Val.Modes[0]
	originalSacrifice, ok := originalMode.Sequence[0].Primitive.(SacrificePermanents)
	if !ok {
		t.Fatalf("original primitive = %T; want SacrificePermanents", originalMode.Sequence[0].Primitive)
	}
	if originalMode.Targets[0].Selection.Val.RequiredTypes[0] != types.Creature ||
		originalMode.Sequence[0].PublishResult != resultKey ||
		originalSacrifice.Selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("mutating cloned sequence changed original: %#v", originalMode)
	}
	originalReanimation, ok := originalMode.Sequence[1].Primitive.(PutOnBattlefield)
	if !ok {
		t.Fatalf("original primitive = %T; want PutOnBattlefield", originalMode.Sequence[1].Primitive)
	}
	if originalReanimation.Sources[0] != CardBattlefieldSource(CardReference{Kind: CardReferenceTarget}) {
		t.Fatalf("mutating cloned reanimation sources changed original: %#v", originalReanimation)
	}
}

func TestCardFaceAbilityCountAndBodyAtUsesCanonicalOrder(t *testing.T) {
	face := CardFace{
		SpellAbility: opt.Val(Mode{Sequence: []Instruction{{Primitive: Draw{}}}}.Ability()),
		ManaAbilities: []ManaAbility{{
			Text:    "Add one mana.",
			Content: Mode{Sequence: []Instruction{{Primitive: AddMana{}}}}.Ability(),
		}},
		TriggeredAbilities: []TriggeredAbility{{
			Text: "When this enters, draw a card.",
			Trigger: TriggerCondition{
				Pattern: TriggerPattern{Event: EventPermanentEnteredBattlefield},
			},
			Content: Mode{Sequence: []Instruction{{Primitive: Draw{}}}}.Ability(),
		}},
		ReplacementAbilities: []ReplacementAbility{{
			Text: "If this would die, exile it instead.",
		}},
		StaticAbilities: []StaticAbility{{
			Text:             "Flying",
			KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Flying}},
		}},
	}

	if face.AbilityCount() != 5 {
		t.Fatalf("ability count = %d, want five categorized abilities", face.AbilityCount())
	}
	if _, ok := face.BodyAt(0).(*AbilityContent); !ok {
		t.Fatalf("BodyAt(0) = %T, want ModalAbilityContent", face.BodyAt(0))
	}
	if _, ok := face.BodyAt(1).(*ManaAbility); !ok {
		t.Fatalf("BodyAt(1) = %T, want ManaAbilityBody", face.BodyAt(1))
	}
	if _, ok := face.BodyAt(2).(*TriggeredAbility); !ok {
		t.Fatalf("BodyAt(2) = %T, want TriggeredAbilityBody", face.BodyAt(2))
	}
	if _, ok := face.BodyAt(3).(*ReplacementAbility); !ok {
		t.Fatalf("BodyAt(3) = %T, want ReplacementAbilityBody", face.BodyAt(3))
	}
	if _, ok := face.BodyAt(4).(*StaticAbility); !ok {
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
		SpellAbility: opt.Val(AbilityContent{}),
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

func TestMutateStaticAbilityCopiesCost(t *testing.T) {
	manaCost := cost.Mana{cost.O(1), cost.G}
	ability := MutateStaticAbility(manaCost)
	face := CardFace{StaticAbilities: []StaticAbility{ability}}

	manaCost[0] = cost.O(9)
	got, ok := face.MutateCost()
	if !ok || got[0] != cost.O(1) || got[1] != cost.G {
		t.Fatalf("MutateCost() = %#v, %v, want copied {1}{G}", got, ok)
	}
	got[0] = cost.O(7)
	again, _ := face.MutateCost()
	if again[0] != cost.O(1) {
		t.Fatalf("MutateCost() returned aliased cost: %#v", again)
	}
}

func TestClearAbilitiesRemovesCategorizedAbilities(t *testing.T) {
	face := CardFace{
		SpellAbility:         opt.Val(AbilityContent{}),
		ActivatedAbilities:   []ActivatedAbility{{Text: "Act"}},
		ManaAbilities:        []ManaAbility{{Text: "Mana"}},
		LoyaltyAbilities:     []LoyaltyAbility{{Text: "Loyal"}},
		TriggeredAbilities:   []TriggeredAbility{{Text: "Trig"}},
		ReplacementAbilities: []ReplacementAbility{{Text: "Replace"}},
		StaticAbilities:      []StaticAbility{{Text: "Static"}},
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
