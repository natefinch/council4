package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review (Engine gap #3408, checked 2026 against Scryfall Oracle,
// rulings, and the Comprehensive Rules effective 2026-09-25):
//
// Court of Locthwain (oracle 6106ba65-f07b-41ec-8b59-e052bbde5529): "At the
// beginning of your upkeep, exile the top card of target opponent's library.
// You may play that card for as long as it remains exiled, and mana of any type
// can be spent to cast it. If you're the monarch, until end of turn, you may
// cast a spell from among cards exiled with this enchantment without paying its
// mana cost." Ruling 2023-09-01: "You may pay any alternative costs the card
// has rather than the card's mana cost."
//
// Faerie Guidemother // Gift of the Fae (oracle
// 04fbb7aa-0dd4-4ac6-837d-822227fb6356): {W} 1/1 Faerie with flying //
// {1}{W} Sorcery — Adventure "Target creature gets +2/+1 and gains flying until
// end of turn." Ruling 2019-10-04: "Casting a card as an Adventure isn't
// casting it for an alternative cost. Effects that allow you to cast a spell
// for an alternative cost or without paying its mana cost may allow you to
// apply those to the Adventure." Ruling 2019-10-04: an Adventure that leaves
// the stack other than by resolving goes to its owner's graveyard.
//
// CR 709.3/709.3a: a player chooses which half of a split card to cast; only
// that half is evaluated. CR 715.3/715.3a: an adventurer card may be cast as an
// Adventure, evaluating only its alternative characteristics. CR 715.3d: other
// effects that allow a player to cast the card may allow casting it as an
// Adventure. CR 712.11b: a modal double-faced card's caster chooses a face.
// CR 118.9a: only one alternative cost can be applied to a spell.
//
// So "cast a spell from among cards exiled with this" and a plain "you may
// cast that card" authorize every legal spell face, a free cast is still the
// one alternative cost, and the one-shot linked permission is used once.

// courtLinkedFreeOnly resolves the real Court of Locthwain upkeep for a
// monarch, exiling def from Player2's library, then removes Court's separate
// any-type play permission so only the monarch's linked free cast authorizes
// the card. Player1 has no mana and sorcery-speed priority.
func courtLinkedFreeOnly(t *testing.T, monarch bool, def *game.CardDef) (*game.Game, *Engine, id.ID) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].IsMonarch = monarch
	cardID := driveLocthwainUpkeepWithCard(t, engine, g, def)
	removed := false
	effects := g.RuleEffects[:0]
	for _, effect := range g.RuleEffects {
		if effect.Kind == game.RuleEffectPlayFromZone && effect.AffectedCardID == cardID {
			removed = true
			continue
		}
		effects = append(effects, effect)
	}
	g.RuleEffects = effects
	if !removed {
		t.Fatal("precondition: Court's any-type play permission is missing")
	}
	want := 0
	if monarch {
		want = 1
	}
	if got := linkedFreeCastPermissions(g); got != want {
		t.Fatalf("precondition: linked free-cast permissions = %d, want %d", got, want)
	}
	giveMainPhasePriority(g, game.Player1)
	return g, engine, cardID
}

func giftOfTheFaeFromExile(cardID id.ID, target *game.Permanent) action.Action {
	return action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceAlternate,
		[]game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
}

func TestCourtLinkedFreeCastAllowsAdventureFromOpponentExile(t *testing.T) {
	t.Parallel()
	g, engine, cardID := courtLinkedFreeOnly(t, true, registryCardDef(t, "Faerie Guidemother"))
	target := addCombatPermanent(g, game.Player1, vanillaCreature("Target", 1, 1))
	gift := giftOfTheFaeFromExile(cardID, target)

	assertExileCastListed(t, g, engine, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil), true)
	assertExileCastListed(t, g, engine, gift, true)
	if !engine.applyAction(g, game.Player1, gift) {
		t.Fatal("free Gift of the Fae cast rejected")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.Face != game.FaceAlternate || obj.ManaSpentToCast != 0 {
		t.Fatalf("stack object = %+v, want Gift of the Fae cast for free", obj)
	}
	if linkedFreeCastPermissions(g) != 0 {
		t.Fatal("the free Adventure cast did not consume the one-shot linked permission")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectivePower(g, target); got != 3 {
		t.Fatalf("target power = %d, want 3 after +2/+1", got)
	}
	if !g.Players[game.Player2].Exile.Contains(cardID) {
		t.Fatal("resolved Adventure was not exiled into its owner's exile")
	}
	if adventure, ok := g.AdventureCards[cardID]; !ok || adventure.Player != game.Player1 {
		t.Fatalf("AdventureCards[card] = %+v, %v; want Player1's CR 715.3d permission", adventure, ok)
	}
	giveMainPhasePriority(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	front := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	assertExileCastListed(t, g, engine, front, true)
	if !engine.applyAction(g, game.Player1, front) {
		t.Fatal("the Adventure's controller could not cast the creature from exile")
	}
}

func TestCourtLinkedFreeAdventureCounteredGoesToOwnersGraveyard(t *testing.T) {
	t.Parallel()
	g, engine, cardID := courtLinkedFreeOnly(t, true, registryCardDef(t, "Faerie Guidemother"))
	target := addCreaturePermanent(g, game.Player1)
	if !engine.applyAction(g, game.Player1, giftOfTheFaeFromExile(cardID, target)) {
		t.Fatal("free Gift of the Fae cast rejected")
	}
	obj, _ := g.Stack.Peek()
	if !counterStackObject(g, obj.ID) {
		t.Fatal("counterStackObject = false")
	}
	if !g.Players[game.Player2].Graveyard.Contains(cardID) {
		t.Fatal("countered Adventure did not go to its owner's graveyard")
	}
	if _, ok := g.AdventureCards[cardID]; ok {
		t.Fatal("countered Adventure granted a CR 715.3d permission")
	}
}

func TestCourtLinkedFreeCastRequiresMonarchForAdventure(t *testing.T) {
	t.Parallel()
	g, engine, cardID := courtLinkedFreeOnly(t, false, registryCardDef(t, "Faerie Guidemother"))
	target := addCreaturePermanent(g, game.Player1)
	gift := giftOfTheFaeFromExile(cardID, target)
	assertExileCastListed(t, g, engine, gift, false)
	if engine.applyAction(g, game.Player1, gift) {
		t.Fatal("Gift of the Fae was cast from exile without any permission")
	}
}

func TestCourtLinkedFreeCastAllowsAdventureFromOwnExile(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].IsMonarch = true
	obj := locthwainSource(g)
	cardID := addCardToExile(g, game.Player1, registryCardDef(t, "Faerie Guidemother"))
	rememberLinkedObject(g, linkedObjectSourceKey(g, obj, courtOfLocthwainLink), game.LinkedObjectRef{CardID: cardID})
	installLocthwainFreeCast(engine, g, obj)
	giveMainPhasePriority(g, game.Player1)
	target := addCreaturePermanent(g, game.Player2)
	gift := giftOfTheFaeFromExile(cardID, target)

	assertExileCastListed(t, g, engine, gift, true)
	if !engine.applyAction(g, game.Player1, gift) {
		t.Fatal("free Gift of the Fae cast from own exile rejected")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.Face != game.FaceAlternate || obj.ManaSpentToCast != 0 {
		t.Fatalf("stack object = %+v, want free Adventure", obj)
	}
	if linkedFreeCastPermissions(g) != 0 {
		t.Fatal("linked permission not consumed")
	}
}

// linkedSplitPool puts two split cards into Player1's exile under Court's
// linked pool with the monarch's free cast installed and no other permission.
func linkedSplitPool(t *testing.T, def func() *game.CardDef) (*game.Game, *Engine, id.ID, id.ID) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].IsMonarch = true
	obj := locthwainSource(g)
	key := linkedObjectSourceKey(g, obj, courtOfLocthwainLink)
	first := addCardToExile(g, game.Player1, def())
	second := addCardToExile(g, game.Player1, def())
	rememberLinkedObject(g, key, game.LinkedObjectRef{CardID: first})
	rememberLinkedObject(g, key, game.LinkedObjectRef{CardID: second})
	installLocthwainFreeCast(engine, g, obj)
	if linkedFreeCastPermissions(g) != 1 {
		t.Fatal("precondition: linked free-cast permission missing")
	}
	giveMainPhasePriority(g, game.Player1)
	return g, engine, first, second
}

func TestCourtLinkedFreeCastAllowsEitherSplitHalfOnce(t *testing.T) {
	t.Parallel()
	g, engine, first, second := linkedSplitPool(t, splitCard)
	for _, cardID := range []id.ID{first, second} {
		for _, face := range []game.FaceIndex{game.FaceFront, game.FaceAlternate} {
			assertExileCastListed(t, g, engine, action.CastSpellFaceFromZone(cardID, zone.Exile, face, nil, 0, nil), true)
		}
	}
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(first, zone.Exile, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("free split alternate-half cast rejected")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.Face != game.FaceAlternate || obj.ManaSpentToCast != 0 {
		t.Fatalf("stack object = %+v, want free alternate half", obj)
	}
	if linkedFreeCastPermissions(g) != 0 {
		t.Fatal("linked permission not consumed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Life; got != 42 {
		t.Fatalf("life = %d, want 42 from the alternate half", got)
	}
	giveMainPhasePriority(g, game.Player1)
	for _, face := range []game.FaceIndex{game.FaceFront, game.FaceAlternate} {
		act := action.CastSpellFaceFromZone(second, zone.Exile, face, nil, 0, nil)
		assertExileCastListed(t, g, engine, act, false)
		if engine.applyAction(g, game.Player1, act) {
			t.Fatalf("second pooled card face %v cast after the one-shot permission was used", face)
		}
	}
}

func TestCourtLinkedFreeSplitCastSurvivesClone(t *testing.T) {
	t.Parallel()
	g, engine, first, _ := linkedSplitPool(t, splitCard)
	clone := g.Clone()
	if !engine.applyAction(clone, game.Player1, action.CastSpellFaceFromZone(first, zone.Exile, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("cloned game rejected the free alternate-half cast")
	}
	if obj, ok := clone.Stack.Peek(); !ok || obj.Face != game.FaceAlternate || obj.ManaSpentToCast != 0 {
		t.Fatalf("clone stack object = %+v, want free alternate half", obj)
	}
	if linkedFreeCastPermissions(clone) != 0 || linkedFreeCastPermissions(g) != 1 {
		t.Fatal("clone and original share the linked permission's usage")
	}
	if !g.Players[game.Player1].Exile.Contains(first) || g.Stack.Size() != 0 {
		t.Fatal("casting on the clone changed the original game")
	}
}

// overloadSplitCard is a split card whose second half has overload, an
// alternative cost the free linked cast cannot combine with (CR 118.9a).
func overloadSplitCard() *game.CardDef {
	def := splitCard()
	selection := game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerNotYou}
	def.Alternate = opt.Val(game.CardFace{
		Name:     "Split Blast",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.R}),
		Colors:   []color.Color{color.Red},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1,
				Allow:     game.TargetAllowPermanent,
				Selection: opt.Val(selection),
			}},
			Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
		}.Ability()),
		Overload: opt.Val(game.OverloadAbility{
			Cost: cost.Mana{cost.O(4), cost.R},
			SpellAbility: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.Destroy{Group: game.BattlefieldGroup(selection)}}},
			}.Ability(),
		}),
	})
	return def
}

func TestCourtLinkedFreeSplitCastCannotCombineWithOverload(t *testing.T) {
	t.Parallel()
	g, engine, first, _ := linkedSplitPool(t, overloadSplitCard)
	artifact := addOverloadArtifact(g, game.Player2, false)
	g.Players[game.Player1].ManaPool.Add(mana.R, 5)

	overloaded := action.CastOverloadedSpellFaceFromZone(first, zone.Exile, game.FaceAlternate, nil)
	assertExileCastListed(t, g, engine, overloaded, false)
	if engine.applyAction(g, game.Player1, overloaded) {
		t.Fatal("overload combined with the linked free cast")
	}
	if linkedFreeCastPermissions(g) != 1 {
		t.Fatal("rejected overload cast consumed the linked permission")
	}

	free := action.CastSpellFaceFromZone(first, zone.Exile, game.FaceAlternate,
		[]game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil)
	assertExileCastListed(t, g, engine, free, true)
	if !engine.applyAction(g, game.Player1, free) {
		t.Fatal("free targeted alternate-half cast rejected")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.Overloaded || obj.ManaSpentToCast != 0 {
		t.Fatalf("stack object = %+v, want the free single-target half", obj)
	}

	// Control: a separate paid permission does authorize the overloaded half.
	g2, engine2, first2, _ := linkedSplitPool(t, overloadSplitCard)
	addOverloadArtifact(g2, game.Player2, false)
	g2.Players[game.Player1].ManaPool.Add(mana.R, 5)
	g2.RuleEffects = append(g2.RuleEffects, game.RuleEffect{
		ID:             g2.IDGen.Next(),
		Kind:           game.RuleEffectPlayFromZone,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		CastFromZone:   zone.Exile,
		AffectedCardID: first2,
		ExpiresFor:     game.Player1,
	})
	overloaded2 := action.CastOverloadedSpellFaceFromZone(first2, zone.Exile, game.FaceAlternate, nil)
	assertExileCastListed(t, g2, engine2, overloaded2, true)
	if !engine2.applyAction(g2, game.Player1, overloaded2) {
		t.Fatal("paid permission could not cast the overloaded half")
	}
	if obj, ok := g2.Stack.Peek(); !ok || !obj.Overloaded || obj.ManaSpentToCast != 5 {
		t.Fatalf("stack object = %+v, want paid overload", obj)
	}
	if linkedFreeCastPermissions(g2) != 1 {
		t.Fatal("paid overload cast consumed the linked permission")
	}
}

// impulseCastCard discards def through a real ExileForPlay cast-variant
// trigger, which exiles it with a "you may cast it this turn" permission naming
// no face, then gives Player1 sorcery-speed priority.
func impulseCastCard(t *testing.T, def *game.CardDef) (*game.Game, *Engine, id.ID) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	exileForPlayDiscardPermanent(g, game.Player1, true)
	cardID := addCardToHand(g, game.Player1, def)
	if !discardCardFromHand(g, game.Player1, cardID) {
		t.Fatal("discardCardFromHand = false")
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("precondition: the discarded card was not exiled for casting")
	}
	giveMainPhasePriority(g, game.Player1)
	return g, engine, cardID
}

func TestNoFaceCastPermissionAllowsAdventureFace(t *testing.T) {
	t.Parallel()
	g, engine, cardID := impulseCastCard(t, registryCardDef(t, "Faerie Guidemother"))
	target := addCreaturePermanent(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 2)
	gift := giftOfTheFaeFromExile(cardID, target)
	assertExileCastListed(t, g, engine, gift, true)
	if !engine.applyAction(g, game.Player1, gift) {
		t.Fatal("Gift of the Fae cast through a no-face cast permission rejected")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.Face != game.FaceAlternate || obj.ManaSpentToCast != 2 {
		t.Fatalf("stack object = %+v, want Gift of the Fae paid {1}{W}", obj)
	}
}

func TestNoFaceCastPermissionAllowsEitherSplitHalf(t *testing.T) {
	t.Parallel()
	g, engine, cardID := impulseCastCard(t, splitCard())
	g.Players[game.Player1].ManaPool.Add(mana.B, 2)
	front := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	alternate := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceAlternate, nil, 0, nil)
	assertExileCastListed(t, g, engine, alternate, true)
	if engine.applyAction(g, game.Player1, front) {
		t.Fatal("black mana paid the white front half")
	}
	if !engine.applyAction(g, game.Player1, alternate) {
		t.Fatal("alternate half cast through a no-face cast permission rejected")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Life; got != 42 {
		t.Fatalf("life = %d, want 42 from the alternate half", got)
	}
}

func TestNoFaceCastPermissionAllowsModalDFCBackSpell(t *testing.T) {
	t.Parallel()
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Modal Creature",
			Types:     []types.Card{types.Creature},
			ManaCost:  opt.Val(cost.Mana{cost.G}),
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name:     "Modal Sorcery",
			Types:    []types.Card{types.Sorcery},
			ManaCost: opt.Val(cost.Mana{cost.R}),
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()},
				}},
			}.Ability()),
		}),
	}
	g, engine, cardID := impulseCastCard(t, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	back := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceBack, nil, 0, nil)
	assertExileCastListed(t, g, engine, back, true)
	if !engine.applyAction(g, game.Player1, back) {
		t.Fatal("back face cast through a no-face cast permission rejected")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.Face != game.FaceBack {
		t.Fatalf("stack object = %+v, want the back face", obj)
	}
}
