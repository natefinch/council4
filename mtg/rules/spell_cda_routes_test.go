package rules

// Source review: Engine gap #3444 makes spell cast events and spell cost
// determination read a spell's characteristic-defining abilities. Pinned corpus
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4 and Scryfall
// rulings retrieved for this unit (default-registry cards, exact Oracle text):
//   - Woodland Changeling 338f0ab6-f744-4e5a-a358-05e71b93758d {1}{G} Creature —
//     Shapeshifter: "Changeling (This card is every creature type.)"
//   - Nameless Inversion d6911456-42cb-4b23-a560-dc98784530f5 {1}{B} Kindred
//     Instant — Shapeshifter with Changeling. Rulings 2025-11-17: a card with
//     changeling is every creature type; changeling is a characteristic-defining
//     ability that functions in all zones; kindred lets a noncreature card have
//     creature types. The Kindred Changeling stand-ins below mirror this shape.
//   - Rohgahh, Kher Keep Overlord a14c692e-dce6-4121-84ec-7fa28ffe4d6b
//     {3}{B}{R}: "Whenever you cast a Kobold spell, you may pay {2}. If you do,
//     create a 4/4 red Dragon creature token with flying." / "Whenever you cast a
//     Dragon spell, create a 0/1 red Kobold creature token named Kobolds of Kher
//     Keep." Ruling 2022-09-09: each trigger resolves before the spell.
//   - Goblin Warchief 39882df0-c20f-469d-94ba-1617224e71a1 {1}{R}{R}: "Goblin
//     spells you cast cost {1} less to cast." Ruling 2018-04-27: it reduces only
//     generic mana.
//   - Mistform Ultimus 7bf0e628-06b5-4e12-a526-a2bce08844f2 {3}{U}: "Mistform
//     Ultimus is every creature type (even if this card isn't on the
//     battlefield)."
//   - Adaptive Automaton 53c730c6-2f8c-4af8-b400-b9d573a71e60 {3}: "As this
//     creature enters, choose a creature type." / "This creature is the chosen
//     type in addition to its other types." As a spell it is only a Construct.
//   - Thalia, Guardian of Thraben c9f8b8fb-1cd8-450e-a1fe-892e7a323479 {1}{W}:
//     "First strike" / "Noncreature spells cost {1} more to cast."
// Comprehensive Rules (official copy effective 2026-09-25, MagicCompRules
// 20260925.txt): 601.2f (cost reductions apply while the total cost is
// determined), 601.2i (cast triggers fire once the spell becomes cast),
// 604.3/604.3a (characteristic-defining abilities define colors or subtypes, are
// unconditional and function in all zones), 702.73a (Changeling: this object is
// every creature type), 702.114a (Devoid: this object is colorless), 308.2
// (kindred subtypes are creature types), 702.16 (protection from a quality checks
// the source's characteristics), 702.8a (flash timing) and 707.10 (a non-cast
// copy is not cast), 702.35a-b (a madness cast pays the madness cost as an
// alternative cost following 601.2b and 601.2f-h, so cost increases and
// reductions apply to it) and 601.2a (the card moves to the stack first).

import (
	"slices"
	"testing"

	carda "github.com/natefinch/council4/mtg/cards/a"
	cardg "github.com/natefinch/council4/mtg/cards/g"
	cardm "github.com/natefinch/council4/mtg/cards/m"
	cardr "github.com/natefinch/council4/mtg/cards/r"
	cardt "github.com/natefinch/council4/mtg/cards/t"
	cardw "github.com/natefinch/council4/mtg/cards/w"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func kindredChangelingInstantTestDef() *game.CardDef {
	def := plainInstantTestDef("Kindred Changeling Instant", 2)
	def.Types = []types.Card{types.Kindred, types.Instant}
	def.Subtypes = []types.Sub{types.Shapeshifter}
	def.StaticAbilities = append(def.StaticAbilities, game.ChangelingStaticBody)
	return def
}

func plainShapeshifterTestDef() *game.CardDef {
	def := vanillaCreature("Plain Shapeshifter", 2, 2)
	def.ManaCost = opt.Val(cost.Mana{cost.O(1), cost.G})
	def.Subtypes = []types.Sub{types.Shapeshifter}
	return def
}

// withKindredChangeling makes def a Kindred Shapeshifter with Changeling.
func withKindredChangeling(def *game.CardDef) *game.CardDef {
	def.Types = append([]types.Card{types.Kindred}, def.Types...)
	def.Subtypes = []types.Sub{types.Shapeshifter}
	def.StaticAbilities = append(def.StaticAbilities, game.ChangelingStaticBody)
	return def
}

func requireKoboldDragonCastEvent(t *testing.T, g *game.Game) {
	t.Helper()
	event := lastSpellCastEvent(t, g)
	if !slices.Contains(event.CardSubtypes, types.Kobold) || !slices.Contains(event.CardSubtypes, types.Dragon) ||
		!slices.Contains(event.CardSubtypes, types.Shapeshifter) {
		t.Fatalf("cast event subtypes = %v, want every creature type including Shapeshifter", event.CardSubtypes)
	}
}

// TestRohgahhCastTriggersSeeCharacteristicDefiningSubtypes proves "Whenever
// you cast a Kobold spell" and "Whenever you cast a Dragon spell" read a
// spell's subtypes as its characteristic-defining abilities define them: a
// Changeling creature spell, a Kindred Changeling instant and Mistform Ultimus
// trigger both abilities, while a Shapeshifter without Changeling and Adaptive
// Automaton (whose type comes from an entry choice, not a CDA) trigger neither.
func TestRohgahhCastTriggersSeeCharacteristicDefiningSubtypes(t *testing.T) {
	printedDragon := vanillaCreature("Printed Dragon", 2, 2)
	printedDragon.ManaCost = opt.Val(cost.Mana{cost.O(2)})
	printedDragon.Subtypes = []types.Sub{types.Dragon}
	for _, tc := range []struct {
		def  *game.CardDef
		pay  func(*mana.Pool)
		want int
	}{
		{cardw.WoodlandChangeling(), func(pool *mana.Pool) { pool.Add(mana.G, 2) }, 2},
		{kindredChangelingInstantTestDef(), func(pool *mana.Pool) { pool.Add(mana.C, 2) }, 2},
		{cardm.MistformUltimus(), func(pool *mana.Pool) { pool.Add(mana.U, 1); pool.Add(mana.C, 3) }, 2},
		{printedDragon, func(pool *mana.Pool) { pool.Add(mana.C, 2) }, 1},
		{plainShapeshifterTestDef(), func(pool *mana.Pool) { pool.Add(mana.G, 2) }, 0},
		{carda.AdaptiveAutomaton(), func(pool *mana.Pool) { pool.Add(mana.C, 3) }, 0},
	} {
		t.Run(tc.def.Name, func(t *testing.T) {
			engine, g := newEffectiveCascadeGame()
			rohgahh := addCombatPermanent(g, game.Player1, cardr.RohgahhKherKeepOverlord())
			cardID := addCardToHand(g, game.Player1, tc.def)
			tc.pay(&g.Players[game.Player1].ManaPool)
			if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
				t.Fatalf("casting %s failed", tc.def.Name)
			}
			engine.putTriggeredAbilitiesOnStack(g)
			if got := triggersFromSource(g, rohgahh.ObjectID); got != tc.want {
				t.Fatalf("%s Rohgahh triggers = %d, want %d", tc.def.Name, got, tc.want)
			}
		})
	}
}

// TestChangelingCastEventRoutes proves every spell-cast route records a
// Changeling spell's effective subtypes: a commander cast from the command
// zone, a free cast by cascade, a cast copy of a card, a paid cast during
// resolution, and suspend, rebound and madness casts from exile.
func TestChangelingCastEventRoutes(t *testing.T) {
	t.Run("commander from command zone", func(t *testing.T) {
		engine, g := newEffectiveCascadeGame()
		rohgahh := addCombatPermanent(g, game.Player1, cardr.RohgahhKherKeepOverlord())
		commanderID := addCommanderToCommandZone(g, game.Player1, cardw.WoodlandChangeling())
		g.Players[game.Player1].ManaPool.Add(mana.G, 2)
		if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(commanderID, zone.Command, nil, 0, nil)) {
			t.Fatal("casting the Changeling commander failed")
		}
		requireKoboldDragonCastEvent(t, g)
		engine.putTriggeredAbilitiesOnStack(g)
		if got := triggersFromSource(g, rohgahh.ObjectID); got != 2 {
			t.Fatalf("Rohgahh triggers = %d, want 2", got)
		}
	})
	t.Run("free cast by cascade", func(t *testing.T) {
		engine, g := newEffectiveCascadeGame()
		rohgahh := addCombatPermanent(g, game.Player1, cardr.RohgahhKherKeepOverlord())
		changelingID := addCardToLibrary(g, game.Player1, cardw.WoodlandChangeling())
		bigID := addCardToHand(g, game.Player1, cascadeSpell(4))
		g.Players[game.Player1].ManaPool.Add(mana.C, 4)
		if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(bigID, nil, 0, nil)); got != 1 {
			t.Fatalf("cascade triggers = %d, want 1", got)
		}
		if got := triggersFromSource(g, rohgahh.ObjectID); got != 0 {
			t.Fatalf("Rohgahh triggers from the cascade spell = %d, want 0", got)
		}
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if top, _ := g.Stack.Peek(); top == nil || top.SourceID != changelingID {
			t.Fatalf("stack top = %+v, want the cascaded Woodland Changeling", top)
		}
		requireKoboldDragonCastEvent(t, g)
		engine.putTriggeredAbilitiesOnStack(g)
		if got := triggersFromSource(g, rohgahh.ObjectID); got != 2 {
			t.Fatalf("Rohgahh triggers = %d, want 2", got)
		}
	})
	t.Run("cast copy of a card", func(t *testing.T) {
		engine, g := newEffectiveCascadeGame()
		rohgahh := addCombatPermanent(g, game.Player1, cardr.RohgahhKherKeepOverlord())
		def := vanillaCreature("Prepared Changeling Source", 2, 2)
		def.Layout, def.Alternate = game.LayoutPrepare, opt.Val(kindredChangelingInstantTestDef().CardFace)
		permanent := addCombatPermanent(g, game.Player1, def)
		permanent.Prepared = true
		g.Players[game.Player1].ManaPool.Add(mana.C, 2)
		cast := action.CastSpellFaceFromZone(permanent.CardInstanceID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
		if !engine.applyAction(g, game.Player1, cast) {
			t.Fatal("casting the prepared Changeling face failed")
		}
		if spell := lastSpellCastStackObject(t, g); !spell.Copy {
			t.Fatal("prepared face was not cast as a copy")
		}
		requireKoboldDragonCastEvent(t, g)
		engine.putTriggeredAbilitiesOnStack(g)
		if got := triggersFromSource(g, rohgahh.ObjectID); got != 2 {
			t.Fatalf("Rohgahh triggers = %d, want 2", got)
		}
	})
	t.Run("paid cast during resolution", func(t *testing.T) {
		engine, g := newEffectiveCascadeGame()
		addCombatPermanent(g, game.Player1, cardg.GoblinWarchief())
		cardID := addCardToGraveyard(g, game.Player1, cardw.WoodlandChangeling())
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		if !engine.castPaidTargetedSpell(g, game.Player1, cardID, zone.Graveyard, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
			t.Fatal("paid cast of the Changeling with Goblin Warchief's reduction failed")
		}
		requireKoboldDragonCastEvent(t, g)
	})
	t.Run("suspend", func(t *testing.T) {
		engine, g := newEffectiveCascadeGame()
		addSuspendedCard(g, game.Player1, withKindredChangeling(suspendSorcery(1, cost.Mana{cost.O(1)})), 1)
		engine.processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		requireKoboldDragonCastEvent(t, g)
	})
	t.Run("rebound", func(t *testing.T) {
		engine, g := newEffectiveCascadeGame()
		cardID := addCardToHand(g, game.Player1, withKindredChangeling(reboundSorcery()))
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
			t.Fatal("casting the rebound Changeling failed")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		engine.processReboundUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if event := lastSpellCastEvent(t, g); event.FromZone != zone.Exile {
			t.Fatalf("last cast from %v, want the rebound cast from exile", event.FromZone)
		}
		requireKoboldDragonCastEvent(t, g)
	})
	t.Run("madness", func(t *testing.T) {
		engine, g := newEffectiveCascadeGame()
		cardID := addCardToHand(g, game.Player1, withKindredChangeling(madnessSorcery(cost.Mana{cost.O(1)})))
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)
		discardCardFromHand(g, game.Player1, cardID)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("madness trigger was not put on the stack")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if event := lastSpellCastEvent(t, g); event.CardID != cardID || event.FromZone != zone.Exile {
			t.Fatalf("last cast event = %+v, want the madness cast", event)
		}
		requireKoboldDragonCastEvent(t, g)
	})
}

// TestNonCastCopyOfChangelingSpellIsNotCast proves a non-cast copy of a
// Changeling spell (CR 707.10) records its effective subtypes on the copy event
// but does not trigger "whenever you cast" abilities.
func TestNonCastCopyOfChangelingSpellIsNotCast(t *testing.T) {
	engine, g := newEffectiveCascadeGame()
	rohgahh := addCombatPermanent(g, game.Player1, cardr.RohgahhKherKeepOverlord())
	cardID := addCardToHand(g, game.Player1, cardw.WoodlandChangeling())
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("casting Woodland Changeling failed")
	}
	spell := lastSpellCastStackObject(t, g)
	engine.putTriggeredAbilitiesOnStack(g)
	before, beforeCasts := triggersFromSource(g, rohgahh.ObjectID), spellCastEventCount(g)
	addEffectSpellToStack(g, game.Player1, game.CopyStackObject{
		Object: game.TargetStackObjectReference(0),
	}, []game.Target{game.StackObjectTarget(spell.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if copied, _ := g.Stack.Peek(); copied == nil || !copied.Copy {
		t.Fatalf("stack top = %+v, want the non-cast copy", copied)
	}
	var copiedEvent game.Event
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == game.EventSpellCopied {
			copiedEvent = g.Events[i]
			break
		}
	}
	if !slices.Contains(copiedEvent.CardSubtypes, types.Kobold) {
		t.Fatalf("copy event subtypes = %v, want every creature type", copiedEvent.CardSubtypes)
	}
	engine.putTriggeredAbilitiesOnStack(g)
	if spellCastEventCount(g) != beforeCasts || triggersFromSource(g, rohgahh.ObjectID) != before {
		t.Fatal("a non-cast copy of a Changeling spell was cast or triggered Rohgahh")
	}
}

// TestGoblinWarchiefReducesCharacteristicDefiningGoblinSpells proves "Goblin
// spells you cast cost {1} less" reads a spell's characteristic-defining
// subtypes while its total cost is determined (CR 601.2f, 604.3): Woodland
// Changeling (from hand and as a commander from the command zone), a Kindred
// Changeling instant and Mistform Ultimus are castable with one less generic
// mana, while a Shapeshifter without Changeling and Adaptive Automaton are not
// reduced. The printed card is never mutated.
func TestGoblinWarchiefReducesCharacteristicDefiningGoblinSpells(t *testing.T) {
	for _, tc := range []struct {
		def      *game.CardDef
		from     zone.Type
		pay      func(*mana.Pool)
		castable bool
	}{
		{cardw.WoodlandChangeling(), zone.Hand, func(pool *mana.Pool) { pool.Add(mana.G, 1) }, true},
		{cardw.WoodlandChangeling(), zone.Command, func(pool *mana.Pool) { pool.Add(mana.G, 1) }, true},
		{kindredChangelingInstantTestDef(), zone.Hand, func(pool *mana.Pool) { pool.Add(mana.C, 1) }, true},
		{cardm.MistformUltimus(), zone.Hand, func(pool *mana.Pool) { pool.Add(mana.U, 1); pool.Add(mana.C, 2) }, true},
		{plainShapeshifterTestDef(), zone.Hand, func(pool *mana.Pool) { pool.Add(mana.G, 1) }, false},
		{carda.AdaptiveAutomaton(), zone.Hand, func(pool *mana.Pool) { pool.Add(mana.C, 2) }, false},
	} {
		t.Run(tc.def.Name+" from "+tc.from.String(), func(t *testing.T) {
			engine, g := newEffectiveCascadeGame()
			addCombatPermanent(g, game.Player1, cardg.GoblinWarchief())
			printed := slices.Clone(tc.def.Subtypes)
			var cast action.Action
			if tc.from == zone.Command {
				cardID := addCommanderToCommandZone(g, game.Player1, tc.def)
				cast = action.CastSpellFromZone(cardID, zone.Command, nil, 0, nil)
			} else {
				cast = action.CastSpell(addCardToHand(g, game.Player1, tc.def), nil, 0, nil)
			}
			tc.pay(&g.Players[game.Player1].ManaPool)
			if got := engine.applyAction(g, game.Player1, cast); got != tc.castable {
				t.Fatalf("cast %s with one less generic mana = %v, want %v", tc.def.Name, got, tc.castable)
			}
			if !slices.Equal(tc.def.Subtypes, printed) {
				t.Fatalf("%s printed subtypes mutated to %v", tc.def.Name, tc.def.Subtypes)
			}
		})
	}
}

// TestSpellColorCDAsReachCastTriggersAndCostModifiers proves color
// characteristic-defining abilities reach both routes (CR 604.3, 702.114a): a
// colorless-cost card that "is all colors" is a red spell for "Whenever you
// cast a red spell" and "Red spells you cast cost {1} less", while a red card
// with Devoid is neither.
func TestSpellColorCDAsReachCastTriggersAndCostModifiers(t *testing.T) {
	allColors := vanillaCreature("All-Colors Golem", 2, 2)
	allColors.ManaCost = opt.Val(cost.Mana{cost.O(2)})
	allColors.StaticAbilities = []game.StaticAbility{{
		Text: "This creature is all colors.",
		ContinuousEffects: []game.ContinuousEffect{{
			AffectedSource: true,
			Layer:          game.LayerColor,
			SetColors:      []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green},
		}},
	}}
	devoid := vanillaCreature("Devoid Red Drone", 2, 2, game.Devoid)
	devoid.ManaCost = opt.Val(cost.Mana{cost.O(2)})
	devoid.Colors = []color.Color{color.Red}
	redWatcher := vanillaCreature("Red Spell Watcher", 1, 1)
	redWatcher.StaticAbilities = []game.StaticAbility{{
		Text: "Red spells you cast cost {1} less to cast.",
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCostModifier,
			Duration:       game.DurationPermanent,
			AffectedPlayer: game.PlayerYou,
			CostModifier: game.CostModifier{
				Kind:             game.CostModifierSpell,
				GenericReduction: 1,
				CardSelection:    game.Selection{ColorsAny: []color.Color{color.Red}},
			},
		}},
	}}
	redWatcher.TriggeredAbilities = []game.TriggeredAbility{{
		Text: "Whenever you cast a red spell, draw a card.",
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event:           game.EventSpellCast,
				Controller:      game.TriggerControllerYou,
				CardSelection:   game.Selection{ColorsAny: []color.Color{color.Red}},
				StackObjectKind: game.StackSpell,
			},
		},
		Content: game.AbilityContent{Modes: []game.Mode{{Sequence: []game.Instruction{{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}}, MinModes: 1, MaxModes: 1},
	}}
	for _, tc := range []struct {
		def *game.CardDef
		red bool
	}{
		{allColors, true},
		{devoid, false},
	} {
		t.Run(tc.def.Name, func(t *testing.T) {
			engine, g := newEffectiveCascadeGame()
			watcher := addCombatPermanent(g, game.Player1, redWatcher)
			cardID := addCardToHand(g, game.Player1, tc.def)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			if got := engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)); got != tc.red {
				t.Fatalf("cast %s for {1} = %v, want %v", tc.def.Name, got, tc.red)
			}
			if !tc.red {
				g.Players[game.Player1].ManaPool.Add(mana.C, 1)
				if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
					t.Fatalf("cast %s for {2} failed", tc.def.Name)
				}
			}
			engine.putTriggeredAbilitiesOnStack(g)
			want := 0
			if tc.red {
				want = 1
			}
			if got := triggersFromSource(g, watcher.ObjectID); got != want {
				t.Fatalf("%s red-spell triggers = %d, want %d", tc.def.Name, got, want)
			}
		})
	}
}

// TestChangelingSpellMatchesSubtypeCantBeCountered proves the stack's
// subtype-filtered "can't be countered" check reads a Changeling spell's
// effective subtypes.
func TestChangelingSpellMatchesSubtypeCantBeCountered(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:                 g.IDGen.Next(),
		Kind:               game.RuleEffectCantBeCountered,
		Controller:         game.Player1,
		AffectedController: game.ControllerYou,
		SpellSubtypes:      []types.Sub{types.Goblin},
		Duration:           game.DurationPermanent,
	})
	changeling := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceTokenDef: cardw.WoodlandChangeling()}
	if stackSpellCanBeCountered(g, changeling) {
		t.Fatal("a Changeling spell must be a Goblin spell that can't be countered")
	}
	plain := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceTokenDef: plainShapeshifterTestDef()}
	if !stackSpellCanBeCountered(g, plain) {
		t.Fatal("a Shapeshifter without Changeling must remain counterable")
	}
}

// TestChangelingSpellMatchesSubtypeFlashPermission proves "You may cast Goblin
// spells as though they had flash" reads a Changeling spell's effective
// subtypes, and that a Shapeshifter without Changeling is not covered.
func TestChangelingSpellMatchesSubtypeFlashPermission(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	flashTimingStaticPermanent(g, game.Player1, nil, []types.Sub{types.Goblin})
	if !playerCanCastAsThoughFlash(g, game.Player1, cardw.WoodlandChangeling()) {
		t.Fatal("Goblin flash permission does not cover a Changeling spell")
	}
	if !playerCanCastAsThoughFlash(g, game.Player1, cardm.MistformUltimus()) {
		t.Fatal("Goblin flash permission does not cover Mistform Ultimus")
	}
	if playerCanCastAsThoughFlash(g, game.Player1, plainShapeshifterTestDef()) {
		t.Fatal("Goblin flash permission covers a Shapeshifter without Changeling")
	}
}

// TestProtectionFromSubtypeSeesChangelingSpell proves protection from Goblins
// (CR 702.16) sees a Changeling spell as a Goblin source on the stack, through
// its last-known information after it leaves the stack, and as a card, while a
// Shapeshifter without Changeling is not a Goblin source.
func TestProtectionFromSubtypeSeesChangelingSpell(t *testing.T) {
	for _, tc := range []struct {
		def  *game.CardDef
		want bool
	}{
		{cardw.WoodlandChangeling(), true},
		{plainShapeshifterTestDef(), false},
	} {
		t.Run(tc.def.Name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			protected := addProtectionFromSubtypesPermanent(g, game.Player2, types.Goblin)
			cardID := addCardToGraveyard(g, game.Player1, tc.def)
			if got := permanentProtectedFromSource(g, protected, cardID, 0); got != tc.want {
				t.Fatalf("protected from %s card = %v, want %v", tc.def.Name, got, tc.want)
			}
			spell := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceTokenDef: tc.def}
			g.Stack.Push(spell)
			if got := permanentProtectedFromSource(g, protected, 0, spell.ID); got != tc.want {
				t.Fatalf("protected from %s spell = %v, want %v", tc.def.Name, got, tc.want)
			}
			if got := slices.Contains(snapshotStackSpell(g, spell).Subtypes, types.Goblin); got != tc.want {
				t.Fatalf("%s last-known Goblin subtype = %v, want %v", tc.def.Name, got, tc.want)
			}
		})
	}
}

// TestMadnessCastDeterminesTotalCostWithSpellCostModifiers proves a madness
// cast determines its total cost from the madness cost like any alternative
// cost (CR 702.35b, 601.2f): Goblin Warchief reduces a Kindred Changeling
// madness spell (a Goblin through its CDA) but not a non-Goblin madness spell,
// and Thalia, Guardian of Thraben increases a noncreature madness spell. An
// unpayable total leaves the card to go to its owner's graveyard.
func TestMadnessCastDeterminesTotalCostWithSpellCostModifiers(t *testing.T) {
	for _, tc := range []struct {
		name      string
		modifier  func() *game.CardDef
		def       func() *game.CardDef
		available int
		cast      bool
		spent     int
	}{
		{"Warchief reduced Changeling", cardg.GoblinWarchief, func() *game.CardDef {
			return withKindredChangeling(madnessSorcery(cost.Mana{cost.O(2)}))
		}, 1, true, 1},
		{"Warchief surplus Changeling", cardg.GoblinWarchief, func() *game.CardDef {
			return withKindredChangeling(madnessSorcery(cost.Mana{cost.O(2)}))
		}, 2, true, 1},
		{"Warchief unaffordable Changeling", cardg.GoblinWarchief, func() *game.CardDef {
			return withKindredChangeling(madnessSorcery(cost.Mana{cost.O(2)}))
		}, 0, false, 0},
		{"Warchief ignores non-Goblin", cardg.GoblinWarchief, func() *game.CardDef {
			return madnessSorcery(cost.Mana{cost.O(2)})
		}, 1, false, 0},
		{"Thalia taxed unpaid", cardt.ThaliaGuardianOfThraben, func() *game.CardDef {
			return madnessSorcery(cost.Mana{cost.O(1)})
		}, 1, false, 0},
		{"Thalia taxed paid", cardt.ThaliaGuardianOfThraben, func() *game.CardDef {
			return madnessSorcery(cost.Mana{cost.O(1)})
		}, 2, true, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine, g := newEffectiveCascadeGame()
			addCombatPermanent(g, game.Player1, tc.modifier())
			player := g.Players[game.Player1]
			cardID := addCardToHand(g, game.Player1, tc.def())
			player.ManaPool.Add(mana.C, tc.available)
			if !discardCardFromHand(g, game.Player1, cardID) {
				t.Fatal("discarding the madness card failed")
			}
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("madness trigger was not put on the stack")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			obj, onStack := g.Stack.Peek()
			cast := onStack && obj.Kind == game.StackSpell && obj.SourceID == cardID
			if cast != tc.cast {
				t.Fatalf("madness cast = %v, want %v", cast, tc.cast)
			}
			if !cast {
				if !player.Graveyard.Contains(cardID) || player.Exile.Contains(cardID) {
					t.Fatal("uncast madness card did not move from exile to the graveyard")
				}
				if got := player.ManaPool.Total(); got != tc.available {
					t.Fatalf("mana pool = %d after an unpaid madness cast, want %d", got, tc.available)
				}
				return
			}
			event := lastSpellCastEvent(t, g)
			if event.CardID != cardID || event.FromZone != zone.Exile {
				t.Fatalf("last cast event = %+v, want the madness cast from exile", event)
			}
			if !event.ManaSpentToCast.Exists || event.ManaSpentToCast.Val != tc.spent || obj.ManaSpentToCast != tc.spent ||
				player.ManaPool.Total() != tc.available-tc.spent {
				t.Fatalf("mana spent = %v (stack %d) with %d left, want %d spent of %d",
					event.ManaSpentToCast, obj.ManaSpentToCast, player.ManaPool.Total(), tc.spent, tc.available)
			}
		})
	}
}
