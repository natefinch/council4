package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// courtOfLocthwainLink is the source-keyed link Court of Locthwain publishes for
// the cards it exiles from opponents' libraries, mirroring the constant the
// cardgen lowering emits.
const courtOfLocthwainLink = "court-of-locthwain-exile"

// TestImpulseExileFromTargetOpponentGrantsAnyManaPlay proves Court of Locthwain's
// upkeep clause: exiling the top card of a target opponent's library grants the
// controller a play-from-exile permission that lets mana of any type pay for it,
// and records the exiled card under the source-keyed linked set.
func TestImpulseExileFromTargetOpponentGrantsAnyManaPlay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	topID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Opponent Spell",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B, cost.B}),
	}})

	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
		SourceID:     g.IDGen.Next(),
		Targets:      []game.Target{game.PlayerTarget(game.Player2)},
	}
	resolveInstruction(engine, g, obj, game.ImpulseExile{
		Player:        game.TargetPlayerReference(0),
		Amount:        game.Fixed(1),
		Duration:      game.DurationPermanent,
		SpendAnyMana:  true,
		PublishLinked: courtOfLocthwainLink,
	}, &TurnLog{})

	if !g.Players[game.Player2].Exile.Contains(topID) {
		t.Fatal("top card of the target opponent's library was not exiled")
	}
	var effect *game.RuleEffect
	for i := range g.RuleEffects {
		if g.RuleEffects[i].Kind == game.RuleEffectPlayFromZone &&
			g.RuleEffects[i].AffectedCardID == topID {
			effect = &g.RuleEffects[i]
		}
	}
	if effect == nil {
		t.Fatal("no play-from-zone rule effect was created for the exiled card")
	}
	if !effect.SpendAnyMana {
		t.Fatal("play-from-zone permission does not carry SpendAnyMana")
	}
	if effect.ExpiresFor != game.Player1 || effect.CastFromZone != zone.Exile {
		t.Fatalf("rule effect = %+v", *effect)
	}
	refs := linkedObjects(g, linkedObjectSourceKey(g, obj, courtOfLocthwainLink))
	if len(refs) != 1 || refs[0].CardID != topID {
		t.Fatalf("linked objects = %v, want one ref to %v", refs, topID)
	}
}

// TestPlayFromExilePermissionSurvivesSourceLeaving proves Court of Locthwain's
// ruling that its exiled cards stay playable "even if Court of Locthwain leaves
// the battlefield": the DurationPermanent play-from-exile permission it grants
// persists — both as an active rule effect and across a turn's cleanup — for as
// long as the affected card remains exiled, without the source permanent, and
// ends only once the card leaves exile.
func TestPlayFromExilePermissionSurvivesSourceLeaving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	topID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Opponent Spell",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B, cost.B}),
	}})
	// The source's object ID names no battlefield permanent, standing in for a
	// Court of Locthwain that has already left the battlefield.
	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
		SourceID:     g.IDGen.Next(),
		Targets:      []game.Target{game.PlayerTarget(game.Player2)},
	}
	resolveInstruction(engine, g, obj, game.ImpulseExile{
		Player:        game.TargetPlayerReference(0),
		Amount:        game.Fixed(1),
		Duration:      game.DurationPermanent,
		SpendAnyMana:  true,
		PublishLinked: courtOfLocthwainLink,
	}, &TurnLog{})

	if !castFromZoneAllowsAnyMana(g, game.Player1, topID, zone.Exile, game.FaceFront) {
		t.Fatal("permission must be active while its source is off the battlefield")
	}
	expireRuleEffects(g)
	if !castFromZoneAllowsAnyMana(g, game.Player1, topID, zone.Exile, game.FaceFront) {
		t.Fatal("cleanup must not delete the permission while the card remains exiled")
	}
	g.Players[game.Player2].Exile.Remove(topID)
	if castFromZoneAllowsAnyMana(g, game.Player1, topID, zone.Exile, game.FaceFront) {
		t.Fatal("permission must end once the card leaves exile")
	}
}

// permission carrying SpendAnyMana lets the controller pay a {2}{B}{B} card's
// cost entirely with green mana, and that the permission without SpendAnyMana
// does not.
func TestPlayFromExileWithAnyManaCastsWithOffColorMana(t *testing.T) {
	setup := func(spendAny bool) (*Engine, *game.Game, id.ID, action.Action) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:     "Exiled Ogre",
			Types:    []types.Card{types.Creature},
			ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B, cost.B}),
		}})
		g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
			ID:             g.IDGen.Next(),
			Kind:           game.RuleEffectPlayFromZone,
			Controller:     game.Player1,
			AffectedPlayer: game.PlayerYou,
			Duration:       game.DurationPermanent,
			CreatedTurn:    1,
			CastFromZone:   zone.Exile,
			AffectedCardID: cardID,
			ExpiresFor:     game.Player1,
			SpendAnyMana:   spendAny,
		})
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone
		g.Turn.ActivePlayer = game.Player1
		g.Turn.PriorityPlayer = game.Player1
		g.Players[game.Player1].ManaPool.Add(mana.G, 4)
		return engine, g, cardID, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	}

	engine, g, cardID, act := setup(true)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("casting the exiled card with any-color mana failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID {
		t.Fatalf("stack top = %#v, want the exiled card", obj)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana pool = %d, want 0 (four green paid the {2}{B}{B} cost)", got)
	}

	engineNo, gNo, _, actNo := setup(false)
	if engineNo.applyAction(gNo, game.Player1, actNo) {
		t.Fatal("casting a {2}{B}{B} card with only green mana must fail without SpendAnyMana")
	}
}

// TestCastFromZoneAllowsAnyMana gates the any-mana permission to the affected
// player, card, and SpendAnyMana flag.
func TestCastFromZoneAllowsAnyMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Exiled Spell", Types: []types.Card{types.Instant}}})
	otherID := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Other Spell", Types: []types.Card{types.Instant}}})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectPlayFromZone,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		Duration:       game.DurationPermanent,
		CreatedTurn:    1,
		CastFromZone:   zone.Exile,
		AffectedCardID: cardID,
		ExpiresFor:     game.Player1,
		SpendAnyMana:   true,
	})

	if !castFromZoneAllowsAnyMana(g, game.Player1, cardID, zone.Exile, game.FaceFront) {
		t.Fatal("the affected card should allow any-mana casting")
	}
	if castFromZoneAllowsAnyMana(g, game.Player1, otherID, zone.Exile, game.FaceFront) {
		t.Fatal("the permission must not extend to another exiled card")
	}
	if castFromZoneAllowsAnyMana(g, game.Player2, cardID, zone.Exile, game.FaceFront) {
		t.Fatal("only the affected player may spend any mana")
	}
}

// TestAnyManaSymbols proves the mana-cost rewrite keeps the cost's size but drops
// every color restriction.
func TestAnyManaSymbols(t *testing.T) {
	got := anyManaSymbols(opt.Val(cost.Mana{
		cost.O(2), cost.B, cost.C, cost.HybridMana(mana.W, mana.U),
		cost.Twobrid(mana.R), cost.PhyrexianMana(mana.G), cost.X, cost.S,
	}))
	want := cost.Mana{
		cost.O(2), cost.O(1), cost.O(1), cost.O(1),
		cost.O(1), cost.PhyrexianGeneric(1), cost.X, cost.S,
	}
	if len(got) != len(want) {
		t.Fatalf("rewritten cost = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("symbol[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

// locthwainSource builds a triggered-ability stack object standing in for Court
// of Locthwain on the battlefield, with a stable source card identity so the
// linked-exile pool and the installed permission share a key.
func locthwainSource(g *game.Game) *game.StackObject {
	return &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
		SourceID:     g.IDGen.Next(),
	}
}

// installLocthwainFreeCast resolves the monarch-gated ApplyRule that installs the
// until-end-of-turn free-cast permission over the source's linked-exile pool.
func installLocthwainFreeCast(engine *Engine, g *game.Game, obj *game.StackObject) {
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCastLinkedExileForFree,
				AffectedPlayer: game.PlayerYou,
				ExiledLinkKey:  courtOfLocthwainLink,
			}},
			Duration: game.DurationUntilEndOfTurn,
		},
		Condition: opt.Val(game.EffectCondition{
			Condition: opt.Val(game.Condition{ControllerIsMonarch: true}),
		}),
	}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

// addLocthwainPoolCard exiles a card into Player1's exile and remembers it under
// the source's linked-exile pool, mirroring a card exiled by Court of Locthwain
// on an earlier turn.
func addLocthwainPoolCard(g *game.Game, obj *game.StackObject, name string) id.ID {
	cardID := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B, cost.B}),
	}})
	rememberLinkedObject(g, linkedObjectSourceKey(g, obj, courtOfLocthwainLink), game.LinkedObjectRef{CardID: cardID})
	return cardID
}

// TestCastLinkedExileForFreeWhenMonarch proves that, while the controller is the
// monarch, the installed permission lets them cast one card from the enchantment's
// linked-exile pool without paying its mana cost.
func TestCastLinkedExileForFreeWhenMonarch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := locthwainSource(g)
	cardID := addLocthwainPoolCard(g, obj, "Pooled Bolt")
	g.Players[game.Player1].IsMonarch = true
	installLocthwainFreeCast(engine, g, obj)

	if !castLinkedExileForFree(g, game.Player1, cardID) {
		t.Fatal("monarch should be permitted to cast the pooled card for free")
	}
	setSorcerySpeedTurn(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("casting the pooled card for free failed")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana pool = %d, want 0 (no mana was spent)", got)
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != cardID {
		t.Fatalf("stack top = %#v, want the pooled card", top)
	}
}

// TestCastLinkedExileForFreeGatedOffWhenNotMonarch proves the free cast is gated
// off when the controller is not the monarch: the monarch condition fails, no
// permission is installed, and the pooled card cannot be cast.
func TestCastLinkedExileForFreeGatedOffWhenNotMonarch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := locthwainSource(g)
	cardID := addLocthwainPoolCard(g, obj, "Pooled Bolt")
	// Controller is not the monarch.
	installLocthwainFreeCast(engine, g, obj)

	if castLinkedExileForFree(g, game.Player1, cardID) {
		t.Fatal("a non-monarch must not receive the free-cast permission")
	}
	setSorcerySpeedTurn(g, game.Player1)
	if engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("a non-monarch must not be able to cast the pooled card")
	}
}

// TestCastLinkedExileForFreeConsumedAfterOneCast proves the permission is a
// one-shot: after casting one pooled card for free, a second pooled card can no
// longer be cast for free.
func TestCastLinkedExileForFreeConsumedAfterOneCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := locthwainSource(g)
	firstID := addLocthwainPoolCard(g, obj, "First Bolt")
	secondID := addLocthwainPoolCard(g, obj, "Second Bolt")
	g.Players[game.Player1].IsMonarch = true
	installLocthwainFreeCast(engine, g, obj)

	setSorcerySpeedTurn(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(firstID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("casting the first pooled card for free failed")
	}
	if castLinkedExileForFree(g, game.Player1, secondID) {
		t.Fatal("the free-cast permission must be consumed after one cast")
	}
	if engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(secondID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("a second pooled card must not be castable for free after the permission is consumed")
	}
}
