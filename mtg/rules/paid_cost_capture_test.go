package rules

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestPaidCostRandomDiscardCapturesOnlyActualMember(t *testing.T) {
	seen := map[bool]bool{}
	for seed := uint64(1); seed <= 12; seed++ {
		g := game.NewGameWithRand([game.NumPlayers]game.PlayerConfig{}, rand.New(rand.NewPCG(1, seed)))
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, compiledPaidSubjectCard(t,
			"Discard a card at random: You gain 1 life. If the discarded card was multicolored, you gain 2 life.", "Artifact"))
		definitions := []*game.CardDef{
			{CardFace: game.CardFace{Name: "Multicolor", Types: []types.Card{types.Creature}, Colors: []color.Color{color.Red, color.Blue}}},
			{CardFace: game.CardFace{Name: "Monocolor", Types: []types.Card{types.Instant}, Colors: []color.Color{color.Red}}},
			{CardFace: game.CardFace{Name: "Colorless", Types: []types.Card{types.Land}}},
		}
		originalColors := map[id.ID][]color.Color{}
		for _, def := range definitions {
			cardID := addCardToHand(g, game.Player1, def)
			card, _ := g.GetCardInstance(cardID)
			card.ZoneVersion = 7
			originalColors[cardID] = slices.Clone(def.Colors)
		}
		setSorcerySpeedTurn(g, game.Player1)
		life := g.Players[game.Player1].Life
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
			t.Fatal("random-discard activation failed")
		}
		obj, ok := g.Stack.Peek()
		if !ok || len(obj.PaidCostSubjects) != 1 || g.Players[game.Player1].Hand.Size() != 2 {
			t.Fatal("random cost did not consume exactly one card")
		}
		fact := obj.PaidCostSubjects[0]
		if !g.Players[game.Player1].Graveyard.Contains(fact.Snapshot.CardID) ||
			!slices.Equal(fact.Snapshot.Colors, originalColors[fact.Snapshot.CardID]) ||
			!fact.CardZoneVersion.Exists || fact.CardZoneVersion.Val != 7 {
			t.Fatal("random predicate captured a candidate rather than the actual discarded incarnation")
		}
		qualifying := len(originalColors[fact.Snapshot.CardID]) == 2
		seen[qualifying] = true
		actual, _ := g.GetCardInstance(fact.Snapshot.CardID)
		actual.Def = definitions[2]
		actual.ZoneVersion += 2
		engine.resolveTopOfStack(g, &TurnLog{})
		expected := life + 1
		if qualifying {
			expected += 2
		}
		if g.Players[game.Player1].Life != expected {
			t.Fatal("random predicate did not read frozen actual colors")
		}
	}
	if !seen[true] || !seen[false] {
		t.Fatal("seeded payment table did not cover qualifying and near-miss subjects")
	}
}

func TestPaidCostCastSacrificeCapturesModifiedToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := &game.Permanent{
		ObjectID: g.IDGen.Next(), Owner: game.Player2, Controller: game.Player1, Token: true,
		TokenDef: &game.CardDef{CardFace: game.CardFace{
			Name: "Copied Artifact", Types: []types.Card{types.Artifact}, Colors: []color.Color{color.Blue},
		}},
	}
	g.Battlefield = append(g.Battlefield, token)
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{Layer: game.LayerType, AffectedObjectID: token.ObjectID, AddSupertypes: []types.Super{types.Legendary}},
		game.ContinuousEffect{Layer: game.LayerColor, AffectedObjectID: token.ObjectID, SetColors: []color.Color{color.Red}},
	)
	spell := compiledPaidSubjectCard(t,
		"As an additional cost to cast this spell, sacrifice an artifact.\nYou gain 1 life. If the sacrificed artifact was legendary, you gain 2 life.", "Sorcery")
	spellID := addCardToHand(g, game.Player1, spell)
	setSorcerySpeedTurn(g, game.Player1)
	life := g.Players[game.Player1].Life
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("actual additional-cast sacrifice failed")
	}
	obj, _ := g.Stack.Peek()
	if len(obj.PaidCostSubjects) != 1 {
		t.Fatal("cast did not capture actual token payment")
	}
	fact := obj.PaidCostSubjects[0]
	if fact.Snapshot.ObjectID != token.ObjectID || fact.Snapshot.CardID != 0 ||
		fact.CardZoneVersion.Exists || !fact.CharacteristicsKnown ||
		!slices.Contains(fact.Snapshot.Supertypes, types.Legendary) ||
		!slices.Equal(fact.Snapshot.Colors, []color.Color{color.Red}) {
		t.Fatalf("token LKI was not captured: %#v", fact)
	}
	delete(g.LastKnownInformation, token.ObjectID)
	g.ContinuousEffects = nil
	copied := game.NewStackObjectCopy(obj, g.IDGen.Next())
	g.Stack.Push(copied)
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != life+6 {
		t.Fatal("original and copied spell did not independently consume the same actual cost facts")
	}
}

func TestPaidCostSnapshotUsesActualFaceAndCopyValues(t *testing.T) {
	for _, copied := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, compiledPaidSubjectCard(t,
			"Sacrifice a creature: You gain 1 life. If a Saproling was sacrificed this way, you gain 2 life.", "Artifact"))
		def := &game.CardDef{
			CardFace: game.CardFace{Name: "Front", Types: []types.Card{types.Artifact}, Colors: []color.Color{color.Blue}},
			Back: opt.Val(game.CardFace{Name: "Back", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Saproling},
				Colors: []color.Color{color.Green}, Power: opt.Val(game.PT{Value: 0}), Toughness: opt.Val(game.PT{Value: 4})}),
		}
		paid := addCombatPermanent(g, game.Player1, def)
		paid.Face = game.FaceBack
		if copied {
			g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
				Layer: game.LayerCopy, AffectedObjectID: paid.ObjectID,
				CopyValues: opt.Val(game.CopyableValues{
					Name: "Copied Bear", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Bear},
					Colors: []color.Color{color.Red}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
				}),
			})
		}
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			Layer: game.LayerPowerToughnessModify, AffectedObjectID: paid.ObjectID, PowerDelta: 3, ToughnessDelta: 1,
		})
		setSorcerySpeedTurn(g, game.Player1)
		life := g.Players[game.Player1].Life
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
			t.Fatal("effective cost selection failed")
		}
		obj, _ := g.Stack.Peek()
		fact := obj.PaidCostSubjects[0]
		if fact.Snapshot.Face != game.FaceBack || !fact.Snapshot.Power.Exists || fact.Snapshot.Power.Val != map[bool]int{false: 3, true: 5}[copied] ||
			!fact.Snapshot.Toughness.Exists || fact.Snapshot.Toughness.Val != map[bool]int{false: 5, true: 3}[copied] {
			t.Fatalf("wrong frozen effective face/PT: %#v", fact.Snapshot)
		}
		delete(g.LastKnownInformation, paid.ObjectID)
		engine.resolveTopOfStack(g, &TurnLog{})
		expected := life + 1
		if !copied {
			expected += 2
		}
		if g.Players[game.Player1].Life != expected {
			t.Fatal("this-way cost predicate read printed front face instead of exact copiable subject")
		}
	}
}

func TestPaidCostUnknownDiscardRefusesPaymentWithoutPublication(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, compiledPaidSubjectCard(t,
		"Discard a card: You gain 1 life. If the discarded card wasn't a land card, you gain 2 life.", "Artifact"))
	paidID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Unknown"}})
	card, _ := g.GetCardInstance(paidID)
	card.Def = nil
	setSorcerySpeedTurn(g, game.Player1)
	if engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("unknown-definition card bypassed payment eligibility")
	}
	if _, exists := g.Stack.Peek(); exists || !g.Players[game.Player1].Hand.Contains(paidID) {
		t.Fatal("failed unknown-definition payment moved or published a subject")
	}
}

func TestPaidCostHandSelfDiscardPublishesItsOwnCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := compiledPaidSubjectCard(t,
		"Discard this card: You gain 1 life. If the discarded card wasn't a land card, you gain 2 life.", "Creature")
	sourceID := addCardToHand(g, game.Player1, def)
	setSorcerySpeedTurn(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(sourceID, 0, nil, 0)) {
		t.Fatal("self-discard activation failed")
	}
	obj, _ := g.Stack.Peek()
	if len(obj.PaidCostSubjects) != 1 || obj.PaidCostSubjects[0].Snapshot.CardID != sourceID {
		t.Fatal("special hand-source payment did not publish its exact subject")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
}

func TestPaidCostRepaidCardDoesNotReusePreviousSnapshot(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	const text = "Discard this card: You gain 1 life. If the discarded card was blue, you gain 2 life."
	firstDef := compiledPaidSubjectCard(t, text, "Creature")
	firstDef.Colors = []color.Color{color.Blue}
	sourceID := addCardToHand(g, game.Player1, firstDef)
	setSorcerySpeedTurn(g, game.Player1)
	life := g.Players[game.Player1].Life
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(sourceID, 0, nil, 0)) {
		t.Fatal("first payment failed")
	}
	first, _ := g.Stack.Peek()
	firstFact := first.PaidCostSubjects[0]
	if !moveCardBetweenZones(g, game.Player1, sourceID, zone.Graveyard, zone.Hand) {
		t.Fatal("returning physical card for independent payment failed")
	}
	secondDef := compiledPaidSubjectCard(t, text, "Creature")
	secondDef.Colors = []color.Color{color.Red}
	card, _ := g.GetCardInstance(sourceID)
	card.Def = secondDef
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(sourceID, 0, nil, 0)) {
		t.Fatal("second payment failed")
	}
	second, _ := g.Stack.Peek()
	secondFact := second.PaidCostSubjects[0]
	if firstFact.Snapshot.CardID != secondFact.Snapshot.CardID || firstFact.CardZoneVersion == secondFact.CardZoneVersion ||
		!slices.Equal(firstFact.Snapshot.Colors, []color.Color{color.Blue}) ||
		!slices.Equal(secondFact.Snapshot.Colors, []color.Color{color.Red}) {
		t.Fatal("two payments of the same physical card shared an incarnation or mutable facts")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != life+4 {
		t.Fatal("independent stack entries read the same payment snapshot")
	}
}

func TestPaidCostInsteadChoosesOneActualDrawBranch(t *testing.T) {
	for _, legendary := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		paid := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Paid", Types: []types.Card{types.Creature}}})
		if legendary {
			g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
				Layer: game.LayerType, AffectedObjectID: paid.ObjectID, AddSupertypes: []types.Super{types.Legendary},
			})
		}
		for range 4 {
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reward", Types: []types.Card{types.Land}}})
		}
		spellID := addCardToHand(g, game.Player1, compiledPaidSubjectCard(t,
			"As an additional cost to cast this spell, sacrifice a creature.\nDraw two cards. If the sacrificed creature was legendary, draw three cards instead.", "Sorcery"))
		setSorcerySpeedTurn(g, game.Player1)
		if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
			t.Fatal("replacement-branch cast failed")
		}
		g.ContinuousEffects = nil
		delete(g.LastKnownInformation, paid.ObjectID)
		engine.resolveTopOfStack(g, &TurnLog{})
		expected := 2
		if legendary {
			expected = 3
		}
		if g.Players[game.Player1].Hand.Size() != expected {
			t.Fatalf("draw branch = %d, want %d rather than both branches", g.Players[game.Player1].Hand.Size(), expected)
		}
	}
}
