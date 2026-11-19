package rules

import (
	"slices"
	"testing"

	cardg "github.com/natefinch/council4/mtg/cards/g"
	cardp "github.com/natefinch/council4/mtg/cards/p"
	cardu "github.com/natefinch/council4/mtg/cards/u"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: #4015; Chaoslace (Oracle 08842aa3-f923-46e9-a106-f542331e9cc1,
// 4ED revision 476180df-8b88-4ead-b6c5-6ccb3e8a2cfd) changes a target spell or
// permanent to red without changing mana symbols. Scryfall returned no rulings:
// https://api.scryfall.com/cards/476180df-8b88-4ead-b6c5-6ccb3e8a2cfd/rulings
// Official CR effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// 105.3, 202.2, 611.2a/c, 613.1e/3/7b: apply layer 5 in timestamp order to
// the fixed object; 400.7/a, 608.3, 614.12: only the resulting permanent
// inherits it, including as it enters; 111.13, 707.2/10: it is not copied,
// but may be applied independently to a spell copy that becomes a token.
// 707.12: a cast copy is itself targetable as a spell; 708.2a, 708.4: face-down
// characteristics, not the underlying card, are the base for layer 5.
// 608.2b, 608.3b, 702.16b, 702.47c/d, 702.103e, 702.140b: protection
// rechecks use the resolving spell's effective color, including for spliced
// text. An illegal Aura target prevents entry, while bestow and mutate fall
// back to creature-spell resolution. Under 613.3 and 702.114a, the LayerColor
// effect applies after Devoid, not the other way around.

func resolveSpellColorEffect(g *game.Game, engine *Engine, target *game.StackObject, effect game.ContinuousEffect, duration game.EffectDuration) {
	addEffectSpellToStack(g, game.Player2, game.ApplyContinuous{
		Object:            opt.Val(game.TargetObjectReference(0)),
		ContinuousEffects: []game.ContinuousEffect{effect},
		Duration:          duration,
	}, []game.Target{game.StackObjectTarget(target.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
}

func paintSpellRed(g *game.Game, engine *Engine, target *game.StackObject) {
	resolveSpellColorEffect(g, engine, target, game.ContinuousEffect{
		Layer: game.LayerColor, SetColors: []color.Color{color.Red},
	}, game.DurationPermanent)
}

func requireSpellColors(t *testing.T, g *game.Game, spell *game.StackObject, want ...color.Color) {
	t.Helper()
	got, ok := stackObjectColors(g, spell)
	if !ok || !slices.Equal(got, want) {
		t.Fatalf("spell %v colors = %v (known=%v), want %v", spell.ID, got, ok, want)
	}
}

func castBlueColorTestCreature(t *testing.T, g *game.Game, engine *Engine) *game.StackObject {
	t.Helper()
	def := vanillaCreature("Blue creature", 2, 2)
	def.Colors, def.ManaCost = []color.Color{color.Blue}, opt.Val(cost.Mana{cost.U})
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("casting the blue creature failed")
	}
	return lastSpellCastStackObject(t, g)
}

func TestResolvingSpellColorEntersAndReplays(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spell := castBlueColorTestCreature(t, g, engine)
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventPermanentEnteredBattlefield,
		EntersTapped: true, EntersTappedOthers: true,
		EntersTappedSelection: &game.Selection{ColorsAny: []color.Color{color.Red}},
	})
	paintSpellRed(g, engine, spell)
	replay := g.Clone()
	resolveSpellColorEffect(g, engine, spell, game.ContinuousEffect{
		Layer: game.LayerColor, SetColors: []color.Color{color.Green},
	}, game.DurationUntilEndOfTurn)
	requireSpellColors(t, g, spell, color.Green)
	expireCleanupDurations(g)
	requireSpellColors(t, g, spell, color.Red)
	replaySpell, _ := replay.Stack.Peek()
	requireSpellColors(t, replay, replaySpell, color.Red)
	if manaValue, ok := stackObjectManaValue(g, spell); !ok || manaValue != 1 {
		t.Fatalf("recolored spell mana value = %d (%v), want 1", manaValue, ok)
	}
	for _, current := range []*game.Game{g, replay} {
		engine.resolveTopOfStack(current, &TurnLog{})
		permanent := permanentByCardID(current, spell.SourceID)
		if permanent == nil || permanent.ObjectID == spell.ID || !permanent.Tapped ||
			!slices.Equal(permanentEffectiveColors(current, permanent), []color.Color{color.Red}) {
			t.Fatalf("resulting permanent = %+v, want a new red permanent that entered tapped", permanent)
		}
		assertEvent(t, current.Events, game.EventSpellResolved, func(event game.Event) bool {
			return event.StackObjectID == spell.ID && slices.Equal(event.Colors, []color.Color{color.Red})
		})
		if !movePermanentToZone(current, permanent, zone.Graveyard) {
			t.Fatal("could not move the permanent to the graveyard")
		}
		def, _ := current.GetCardInstance(spell.SourceID)
		if !slices.Equal(def.Def.Colors, []color.Color{color.Blue}) || !slices.Equal(def.Def.ManaCost.Val, cost.Mana{cost.U}) {
			t.Fatalf("card definition mutated: colors %v, cost %v", def.Def.Colors, def.Def.ManaCost)
		}
	}
}

func TestResolvingSpellColorCopyRoutes(t *testing.T) {
	t.Run("copy of a permanent spell becomes a token", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spell := castBlueColorTestCreature(t, g, engine)
		paintSpellRed(g, engine, spell)
		copied := copyStackObject(engine, g, spell)
		if copied == nil || !copied.Copy {
			t.Fatal("could not copy the spell")
		}
		requireSpellColors(t, g, copied, color.Blue)
		paintSpellRed(g, engine, copied)
		engine.resolveTopOfStack(g, &TurnLog{})
		if len(g.Battlefield) != 1 || !g.Battlefield[0].Token ||
			!slices.Equal(permanentEffectiveColors(g, g.Battlefield[0]), []color.Color{color.Red}) ||
			!slices.Equal(g.Battlefield[0].TokenDef.Colors, []color.Color{color.Blue}) {
			t.Fatal("the targeted copy must become a red token with blue copiable values")
		}
		if casts := spellCastEventCount(g); casts != 1 {
			t.Fatalf("cast events = %d, want only the original cast", casts)
		}
	})
	t.Run("storm copy excludes recoloring", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		original := pushStackSpell(g, game.Player1, "Storm fixture", types.Instant)
		g.CardInstances[original.SourceID].Def.Colors = []color.Color{color.Blue}
		paintSpellRed(g, engine, original)
		trigger := &game.StackObject{Controller: game.Player1, HasTriggerEvent: true, TriggerEvent: game.Event{
			StackObjectID: original.ID, SpellsCastBeforeThisTurn: opt.Val(1),
		}}
		engine.resolveStormTriggeredAbility(g, trigger, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		copied, _ := g.Stack.Peek()
		if copied == nil || !copied.Copy {
			t.Fatal("storm did not create a copy")
		}
		requireSpellColors(t, g, copied, color.Blue)
		requireSpellColors(t, g, original, color.Red)
	})
	t.Run("cast copy of a card", func(t *testing.T) {
		engine, g := newEffectiveCascadeGame()
		def := vanillaCreature("Prepared spell source", 2, 2)
		def.Layout, def.Alternate = game.LayoutPrepare, opt.Val(game.CardFace{
			Name: "Prepared blue instant", Types: []types.Card{types.Instant},
			Colors: []color.Color{color.Blue}, ManaCost: opt.Val(cost.Mana{cost.U}),
		})
		permanent := addCombatPermanent(g, game.Player1, def)
		permanent.Prepared = true
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		cast := action.CastSpellFaceFromZone(permanent.CardInstanceID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
		if !engine.applyAction(g, game.Player1, cast) {
			t.Fatal("casting the prepared spell copy failed")
		}
		spell := lastSpellCastStackObject(t, g)
		if !spell.Copy || !spell.CastCopy {
			t.Fatal("the prepared spell is not a cast copy")
		}
		paintSpellRed(g, engine, spell)
		requireSpellColors(t, g, spell, color.Red)
		if cast := lastSpellCastEvent(t, g); !slices.Equal(cast.Colors, []color.Color{color.Blue}) {
			t.Fatalf("cast history color = %v, want blue", cast.Colors)
		}
	})
}

func TestResolvingSpellColorCounterAndFreeRecast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spell := castBlueColorTestCreature(t, g, engine)
	paintSpellRed(g, engine, spell)
	counter := pushStackSpell(g, game.Player2, "Red-only counter", types.Instant, game.StackObjectTarget(spell.ID))
	g.CardInstances[counter.SourceID].Def.SpellAbility = opt.Val(game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
			Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}, SpellColors: []color.Color{color.Red}},
		}},
		Sequence: []game.Instruction{{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}}},
	}.Ability())
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(spell.SourceID) || len(g.Battlefield) != 0 {
		t.Fatal("red-only counter did not see the recolored spell at resolution")
	}
	if !moveCardBetweenZones(g, game.Player1, spell.SourceID, zone.Graveyard, zone.Exile) {
		t.Fatal("could not move the countered card to exile")
	}
	if !engine.chooseAndCastFreeSpellFromSourceMay(g, g.Players[game.Player1], game.Player1, spell.SourceID,
		zone.Exile, game.Selection{}, opt.V[int]{}, false, true, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("could not cast the same card from exile without paying its mana cost")
	}
	recast := lastSpellCastStackObject(t, g)
	requireSpellColors(t, g, recast, color.Blue)
	paintSpellRed(g, engine, recast)
	cast := lastSpellCastEvent(t, g)
	if cast.FromZone != zone.Exile || !slices.Equal(cast.Colors, []color.Color{color.Blue}) {
		t.Fatalf("recoloring changed cast history: colors %v, source %v", cast.Colors, cast.FromZone)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if permanent := permanentByCardID(g, spell.SourceID); permanent == nil ||
		!slices.Equal(permanentEffectiveColors(g, permanent), []color.Color{color.Red}) {
		t.Fatal("the recolored free cast did not enter red")
	}
}

func TestResolvingSpellColorFaceDownPrivacy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := vanillaCreature("Hidden blue creature", 2, 2)
	def.Colors = []color.Color{color.Blue}
	def.StaticAbilities = []game.StaticAbility{{
		KeywordAbilities: []game.KeywordAbility{game.MorphKeyword{Cost: cost.Mana{cost.U}}},
	}}
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastFaceDown(cardID, game.FaceFront, game.FaceDownMorph)) {
		t.Fatal("casting the face-down creature failed")
	}
	spell := lastSpellCastStackObject(t, g)
	paintSpellRed(g, engine, spell)
	view := NewObservation(g, game.Player2).stackObjectView(spell)
	if view.Name != "" || view.FaceDownCard.Exists || !slices.Equal(view.Colors, []color.Color{color.Red}) {
		t.Fatalf("opponent view = %+v, want the public layer 5 change but not the hidden card", view)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if permanent := permanentByCardID(g, cardID); permanent == nil || !permanent.FaceDown ||
		!slices.Equal(permanentEffectiveColors(g, permanent), []color.Color{color.Red}) {
		t.Fatal("the face-down creature did not keep its color effect on entry")
	}
}

func TestResolvingSpellColorFizzle(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCreaturePermanent(g, game.Player2)
	spell := pushStackSpell(g, game.Player1, "Targeted instant", types.Instant, game.PermanentTarget(creature.ObjectID))
	def := g.CardInstances[spell.SourceID].Def
	def.Colors = []color.Color{color.Blue}
	def.SpellAbility = opt.Val(game.Mode{
		Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent}},
	}.Ability())
	paintSpellRed(g, engine, spell)
	requireSpellColors(t, g, spell, color.Red)
	if !movePermanentToZone(g, creature, zone.Exile) {
		t.Fatal("could not remove the instant's target")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(spell.SourceID) {
		t.Fatal("the instant did not fizzle to the graveyard")
	}
	assertNoEvent(t, g.Events, game.EventSpellResolved, func(event game.Event) bool {
		return event.StackObjectID == spell.ID
	})
}

func TestResolvingSpellColorProtectionOnCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	protected := addProtectionFromColorPermanent(g, game.Player2, color.Red)
	cardID := addCardToHand(g, game.Player1, cardu.Unsummon())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, []game.Target{game.PermanentTarget(protected.ObjectID)}, 0, nil)) {
		t.Fatal("casting the blue Unsummon failed")
	}
	original := lastSpellCastStackObject(t, g)
	copied := copyStackObject(engine, g, original)
	if copied == nil || !copied.Copy || copied.ID == original.ID {
		t.Fatal("the targeted spell did not produce a separate copy")
	}
	paintSpellRed(g, engine, copied)
	requireSpellColors(t, g, copied, color.Red)
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, protected.ObjectID); !ok {
		t.Fatal("the red copy illegally returned the creature with protection from red")
	}
	assertNoEvent(t, g.Events, game.EventSpellResolved, func(event game.Event) bool {
		return event.StackObjectID == copied.ID
	})
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Hand.Contains(protected.CardInstanceID) {
		t.Fatal("the original blue spell did not return the creature")
	}
}

func TestResolvingSpellColorProtectionWithSplice(t *testing.T) {
	for _, tc := range []struct {
		name        string
		blockHost   bool
		blockSplice bool
	}{
		{name: "only host illegal", blockHost: true},
		{name: "only splice illegal", blockSplice: true},
		{name: "both illegal", blockHost: true, blockSplice: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, engine := newSpliceGame()
			hostID := addCardToHand(g, game.Player1, cardg.GlacialRay())
			addCardToHand(g, game.Player1, cardg.GlacialRay())
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			g.Players[game.Player1].ManaPool.Add(mana.R, 2)
			protected := addProtectionFromColorPermanent(g, game.Player2, color.Blue)
			hostTarget, spliceTarget := game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player2)
			wantDamage := 4
			if tc.blockHost {
				hostTarget = game.PermanentTarget(protected.ObjectID)
				wantDamage -= 2
			}
			if tc.blockSplice {
				spliceTarget = game.PermanentTarget(protected.ObjectID)
				wantDamage -= 2
			}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{
				splices: []string{"Glacial Ray"}, targets: []game.Target{spliceTarget},
			}}
			if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, []game.Target{hostTarget}, 0, nil), agents, &TurnLog{}) {
				t.Fatal("casting red Glacial Ray with a spliced Glacial Ray failed")
			}
			spell := lastSpellCastStackObject(t, g)
			if len(spell.SplicedTargets) != 1 || !slices.Equal(spell.SplicedTargets[0], []game.Target{spliceTarget}) {
				t.Fatalf("spliced targets = %v, want %v", spell.SplicedTargets, spliceTarget)
			}
			resolveSpellColorEffect(g, engine, spell, game.ContinuousEffect{
				Layer: game.LayerColor, SetColors: []color.Color{color.Blue},
			}, game.DurationPermanent)
			requireSpellColors(t, g, spell, color.Blue)
			life := g.Players[game.Player2].Life
			engine.resolveTopOfStack(g, &TurnLog{})
			matchesSpell := func(event game.Event) bool { return event.StackObjectID == spell.ID }
			if wantDamage == 0 {
				assertNoEvent(t, g.Events, game.EventSpellResolved, matchesSpell)
			} else {
				assertEvent(t, g.Events, game.EventSpellResolved, matchesSpell)
			}
			// Damage prevention alone could hide a targeting failure, so also
			// check that each illegal target was screened out before effects ran.
			hostDeferred := spell.Targets[0].Kind == game.TargetDeferred
			spliceDeferred := spell.SplicedTargets[0][0].Kind == game.TargetDeferred
			if hostDeferred != tc.blockHost || spliceDeferred != tc.blockSplice {
				t.Fatalf("deferred host/splice targets = %v/%v, want %v/%v", hostDeferred, spliceDeferred, tc.blockHost, tc.blockSplice)
			}
			if got := life - g.Players[game.Player2].Life; got != wantDamage || protected.MarkedDamage != 0 {
				t.Fatalf("damage to player/protected creature = %d/%d, want %d/0", got, protected.MarkedDamage, wantDamage)
			}
		})
	}
}

func TestResolvingSpellColorProtectionOnAura(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	protected := addProtectionFromColorPermanent(g, game.Player2, color.Red)
	cardID := addCardToHand(g, game.Player1, cardp.Pacifism())
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, []game.Target{game.PermanentTarget(protected.ObjectID)}, 0, nil)) {
		t.Fatal("casting the white Aura failed")
	}
	spell := lastSpellCastStackObject(t, g)
	paintSpellRed(g, engine, spell)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("the Aura with an illegal target did not go to the graveyard")
	}
	assertNoEvent(t, g.Events, game.EventPermanentEnteredBattlefield, func(event game.Event) bool {
		return event.CardID == cardID
	})
	assertNoEvent(t, g.Events, game.EventSpellResolved, func(event game.Event) bool {
		return event.StackObjectID == spell.ID
	})
}

func TestResolvingSpellColorProtectionOnBestow(t *testing.T) {
	g, engine := setupBestowMain(t)
	protected := addProtectionFromColorPermanent(g, game.Player2, color.Red)
	cardID := addCardToHand(g, game.Player1, bestowTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if !engine.applyAction(g, game.Player1, action.CastBestowSpell(cardID, []game.Target{game.PermanentTarget(protected.ObjectID)}, 0, nil)) {
		t.Fatal("casting the colorless spell bestowed failed")
	}
	spell := lastSpellCastStackObject(t, g)
	paintSpellRed(g, engine, spell)
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentByCardID(g, cardID)
	if spell.Targets[0].Kind != game.TargetDeferred || permanent == nil || permanent.Bestowed || permanent.AttachedTo.Exists ||
		!permanentHasType(g, permanent, types.Creature) || !slices.Equal(permanentEffectiveColors(g, permanent), []color.Color{color.Red}) {
		t.Fatal("the illegal bestow target should be screened out and the spell should enter as a red creature")
	}
}

func TestResolvingSpellColorMutate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player1)
	spell := castDirgeBatForMutate(t, g, engine, target)
	paintSpellRed(g, engine, spell)
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(target.MergedCards) != 1 || !slices.Equal(permanentEffectiveColors(g, target), []color.Color{color.Red}) {
		t.Fatal("the merged permanent did not keep the spell's color effect")
	}
}

func TestResolvingSpellColorProtectionOnMutateWithDevoid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	protected := addProtectionFromColorPermanent(g, game.Player1, color.Red)
	def := mutateCard()
	def.StaticAbilities = append(def.StaticAbilities, game.DevoidStaticBody)
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastMutateSpell(cardID, protected.ObjectID)) {
		t.Fatal("casting the colorless mutating spell failed")
	}
	spell := lastSpellCastStackObject(t, g)
	requireSpellColors(t, g, spell)
	paintSpellRed(g, engine, spell)
	requireSpellColors(t, g, spell, color.Red)
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentByCardID(g, cardID)
	if len(protected.MergedCards) != 0 || permanent == nil || permanent.ObjectID == protected.ObjectID {
		t.Fatal("protection from red should prevent merging and the spell should enter separately")
	}
}

func TestResolvingSpellColorRedirectedEntry(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spell := castBlueColorTestCreature(t, g, engine)
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventPermanentEnteredBattlefield,
		EntersTapped: true, EntersTappedOthers: true, ReplaceToZone: zone.Exile,
		EntersTappedSelection: &game.Selection{ColorsAny: []color.Color{color.Red}},
	})
	paintSpellRed(g, engine, spell)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(spell.SourceID) || len(g.Battlefield) != 0 {
		t.Fatal("the red-only replacement did not redirect the would-be red permanent to exile")
	}
	if !engine.chooseAndCastFreeSpellFromSourceMay(g, g.Players[game.Player1], game.Player1, spell.SourceID,
		zone.Exile, game.Selection{}, opt.V[int]{}, false, true, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("could not recast the exiled card")
	}
	requireSpellColors(t, g, lastSpellCastStackObject(t, g), color.Blue)
	engine.resolveTopOfStack(g, &TurnLog{})
	if permanent := permanentByCardID(g, spell.SourceID); permanent == nil || permanent.Tapped ||
		!slices.Equal(permanentEffectiveColors(g, permanent), []color.Color{color.Blue}) {
		t.Fatal("the second spell did not become an untapped blue permanent")
	}
}
