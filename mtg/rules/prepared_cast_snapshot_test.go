package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
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

// Source review (#3220, authoritative cast-face characteristic snapshots):
// Carnivorous Cultivator // Enroot, oracle db655c5a-06f6-42bc-8180-d9ca1c1d939a,
// FRA 100 printing
// 79dd5c54-5ea5-47b5-8f9b-50ed57a5ea45 (layout prepare; Enroot is a green {G}
// Sorcery with no subtypes); https://api.scryfall.com/cards/79dd5c54-5ea5-47b5-8f9b-50ed57a5ea45/rulings
// returned 15 WotC rulings dated 2026-08-21, including that only the prepared
// creature's current controller can cast the copy. Lys Alana Huntmaster, oracle
// 3f3439e1-75ce-482d-881f-836492dca6e9, KHC 67
// printing 0849f67b-da5d-49f7-840f-f87f48518156; its rulings URL returned one
// ruling dated 2016-06-08. Official CR TXT effective 2026-09-25
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt):
// 722.3c (the prepare copy has only the prepare spell's characteristics, which
// become its normal characteristics), 722.4, 601.2i (cast triggers see the spell
// as cast), 707.2 (copiable values, as modified by face-down status), 707.9b
// (a copy effect's characteristic exception becomes part of the copy's copiable
// values), 707.10
// and 707.10c (a copy of a spell copies its characteristics and choices; "copy
// target spell" effects), 707.12 (casting a copy of an object follows 601.2a-h
// and the copy becomes cast), 702.40a (storm),
// 702.37a/c and 708.2/708.2a (a face-down morph spell has no name, no subtypes,
// and only the listed characteristics), and 110.2 (controller). Required
// behavior: cast and copy events describe the spell on the stack, so empty
// subtype, supertype and color fields stay empty instead of reading the
// originating permanent or front face, and the caster stays the event's
// controller after the originating permanent changes control.
// The default registry entries for these cards carry no oracle identity, so
// they are looked up by exact name.
func lookupSnapshotTestCard(t *testing.T, name string) *game.CardDef {
	t.Helper()
	def := cards.NewDefaultRegistry().Lookup(name)
	if def == nil || def.Name != name {
		t.Fatalf("default-registry lookup of %q returned %+v", name, def)
	}
	return def
}

func snapshotTestMainPhase(g *game.Game) {
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
}

func triggeredAbilitiesFrom(g *game.Game, sourceID id.ID) int {
	count := 0
	for _, obj := range g.Stack.Objects() {
		if obj.Kind == game.StackTriggeredAbility && obj.SourceID == sourceID {
			count++
		}
	}
	return count
}

func lastSnapshotEvent(t *testing.T, g *game.Game, kind game.EventKind) game.Event {
	t.Helper()
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == kind {
			return g.Events[i]
		}
	}
	t.Fatalf("no %v event recorded", kind)
	return game.Event{}
}

func TestPreparedEnrootCastDoesNotTriggerLysAlanaHuntmaster(t *testing.T) {
	for _, origin := range []string{"card", "token"} {
		t.Run(origin, func(t *testing.T) {
			huntmasterDef := lookupSnapshotTestCard(t, "Lys Alana Huntmaster")
			cultivatorDef := lookupSnapshotTestCard(t, "Carnivorous Cultivator")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			huntmaster := addCombatPermanent(g, game.Player1, huntmasterDef)
			snapshotTestMainPhase(g)

			var castFrom id.ID
			if origin == "card" {
				cardID := addCardToHand(g, game.Player1, cultivatorDef)
				g.Players[game.Player1].ManaPool.Add(mana.G, 2)
				if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
					t.Fatal("casting Carnivorous Cultivator failed")
				}
				if !engine.putTriggeredAbilitiesOnStack(g) || triggeredAbilitiesFrom(g, huntmaster.ObjectID) != 1 {
					t.Fatal("casting the Elf creature face did not trigger Lys Alana Huntmaster")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				engine.resolveTopOfStack(g, &TurnLog{})
				permanent := permanentForCard(g, cardID)
				if permanent == nil || !permanent.Prepared {
					t.Fatalf("Carnivorous Cultivator permanent = %+v, want prepared", permanent)
				}
				castFrom = cardID
			} else {
				token, ok := createTokenPermanent(g, game.Player1, copyCardDef(cultivatorDef))
				if !ok || !token.Prepared {
					t.Fatalf("token = %+v, want prepared Carnivorous Cultivator copy", token)
				}
				castFrom = token.ObjectID
			}
			if !g.Stack.IsEmpty() {
				t.Fatalf("stack size = %d before Enroot, want empty", g.Stack.Size())
			}

			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(castFrom, zone.Battlefield, game.FaceAlternate, nil, 0, nil)) {
				t.Fatal("casting prepared Enroot failed")
			}
			cast := lastSnapshotEvent(t, g, game.EventSpellCast)
			if !cast.CardCharacteristicsKnown || len(cast.CardSubtypes) != 0 || len(cast.CardSupertypes) != 0 ||
				!slices.Equal(cast.CardTypes, []types.Card{types.Sorcery}) || !slices.Equal(cast.Colors, []color.Color{color.Green}) {
				t.Fatalf("Enroot cast event = %+v, want authoritative green Sorcery with no subtypes", cast)
			}
			engine.putTriggeredAbilitiesOnStack(g)
			if got := triggeredAbilitiesFrom(g, huntmaster.ObjectID); got != 0 {
				t.Fatalf("Lys Alana Huntmaster triggers = %d for Enroot, want 0 (Enroot is not an Elf spell)", got)
			}
		})
	}
}

// legendaryElfColorlessPrepareCard is a legendary green Elf whose prepare spell
// is a colorless Sorcery with no supertypes or subtypes, so every characteristic
// that differs between the faces would leak if the event were backfilled.
func legendaryElfColorlessPrepareCard(storm bool) *game.CardDef {
	alternate := game.CardFace{
		Name:     "Colorless Preparation",
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}},
		}.Ability()),
	}
	if storm {
		alternate.StaticAbilities = []game.StaticAbility{game.StormStaticBody}
	}
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:           "Legendary Elf Preparer",
			ManaCost:       opt.Val(cost.Mana{cost.G}),
			Colors:         []color.Color{color.Green},
			Supertypes:     []types.Super{types.Legendary},
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Elf, types.Druid},
			Power:          opt.Val(game.PT{Value: 2}),
			Toughness:      opt.Val(game.PT{Value: 2}),
			EntersPrepared: true,
		},
		Layout:    game.LayoutPrepare,
		Alternate: opt.Val(alternate),
	}
}

func addPreparedOriginPermanent(t *testing.T, g *game.Game, def *game.CardDef) (id.ID, *game.Permanent) {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, def)
	card, _ := g.GetCardInstance(cardID)
	g.Players[game.Player1].Hand.Remove(cardID)
	permanent, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok || !permanent.Prepared {
		t.Fatalf("prepared origin = %+v, want prepared permanent", permanent)
	}
	return cardID, permanent
}

type snapshotObserver struct {
	name    string
	pattern game.TriggerPattern
	want    bool
}

func preparedSnapshotObservers(matchCopy bool) []snapshotObserver {
	pattern := func(sel game.Selection) game.TriggerPattern {
		return game.TriggerPattern{Event: game.EventSpellCast, Controller: game.TriggerControllerYou, MatchSpellCopy: matchCopy, CardSelection: sel}
	}
	historic := pattern(game.Selection{})
	historic.RequireHistoric = true
	return []snapshotObserver{
		{name: "Elf spell", pattern: pattern(game.Selection{SubtypesAny: []types.Sub{types.Elf}})},
		{name: "legendary spell", pattern: pattern(game.Selection{Supertypes: []types.Super{types.Legendary}})},
		{name: "green spell", pattern: pattern(game.Selection{ColorsAny: []color.Color{color.Green}})},
		{name: "creature spell", pattern: pattern(game.Selection{RequiredTypes: []types.Card{types.Creature}})},
		{name: "historic spell", pattern: historic},
		{name: "colorless spell", pattern: pattern(game.Selection{Colorless: true}), want: true},
		{name: "sorcery spell", pattern: pattern(game.Selection{RequiredTypes: []types.Card{types.Sorcery}}), want: true},
	}
}

func TestPreparedCastAndStormCopyKeepEmptyCharacteristicSnapshot(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID, origin := addPreparedOriginPermanent(t, g, legendaryElfColorlessPrepareCard(true))
	observer := addCombatCreaturePermanent(g, game.Player1)
	elfWatcher := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventSpellCast, Controller: game.TriggerControllerYou, MatchSpellCopy: true,
		CardSelection: game.Selection{SubtypesAny: []types.Sub{types.Elf}},
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	sorceryWatcher := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventSpellCast, Controller: game.TriggerControllerYou, MatchSpellCopy: true,
		CardSelection: game.Selection{RequiredTypes: []types.Card{types.Sorcery}},
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	firstID := addCardToHand(g, game.Player1, simpleGainLifeInstant("First Spell"))
	snapshotTestMainPhase(g)
	if !engine.applyAction(g, game.Player1, action.CastSpell(firstID, nil, 0, nil)) {
		t.Fatal("first spell cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	g.TriggerEventCursor = len(g.Events)

	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("casting prepared storm spell failed")
	}
	resolveStormTrigger(t, engine, g, nil)
	cast := lastSnapshotEvent(t, g, game.EventSpellCast)
	copied := lastSnapshotEvent(t, g, game.EventSpellCopied)
	for _, event := range []game.Event{cast, copied} {
		if !event.CardCharacteristicsKnown || len(event.CardSubtypes) != 0 || len(event.CardSupertypes) != 0 ||
			len(event.Colors) != 0 || !slices.Equal(event.CardTypes, []types.Card{types.Sorcery}) {
			t.Fatalf("%v event = %+v, want authoritative colorless Sorcery with no supertypes or subtypes", event.Kind, event)
		}
	}
	if cast.PermanentID != origin.ObjectID {
		t.Fatalf("cast event PermanentID = %v, want originating permanent %v", cast.PermanentID, origin.ObjectID)
	}
	for _, tc := range preparedSnapshotObservers(false) {
		if got := triggerMatchesEvent(g, observer, &tc.pattern, cast); got != tc.want {
			t.Errorf("cast: %s observer matched = %v, want %v", tc.name, got, tc.want)
		}
	}
	for _, tc := range preparedSnapshotObservers(true) {
		if got := triggerMatchesEvent(g, observer, &tc.pattern, copied); got != tc.want {
			t.Errorf("storm copy: %s observer matched = %v, want %v", tc.name, got, tc.want)
		}
	}

	engine.putTriggeredAbilitiesOnStack(g)
	if got := triggeredAbilitiesFrom(g, elfWatcher.ObjectID); got != 0 {
		t.Fatalf("Elf watcher triggers = %d, want 0 for prepared cast and storm copy", got)
	}
	if got := triggeredAbilitiesFrom(g, sorceryWatcher.ObjectID); got != 2 {
		t.Fatalf("Sorcery watcher triggers = %d, want 2 (prepared cast and storm copy)", got)
	}
}

func TestPreparedCastKeepsCasterAfterOriginControlChanges(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID, origin := addPreparedOriginPermanent(t, g, legendaryElfColorlessPrepareCard(false))
	source1 := addCombatCreaturePermanent(g, game.Player1)
	source2 := addCombatCreaturePermanent(g, game.Player2)
	snapshotTestMainPhase(g)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("casting prepared spell failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	origin.Controller = game.Player2

	castSorcery := opt.Val(game.Condition{EventHistory: opt.Val(game.EventHistoryCondition{
		Pattern: game.TriggerPattern{
			Event:         game.EventSpellCast,
			Controller:    game.TriggerControllerYou,
			CardSelection: game.Selection{RequiredTypes: []types.Card{types.Sorcery}},
		},
		Window: game.EventHistoryCurrentTurn,
	})})
	check := func(t *testing.T, g *game.Game, source1, source2 *game.Permanent) {
		t.Helper()
		if !conditionSatisfied(g, conditionContext{controller: game.Player1, source: source1}, castSorcery) {
			t.Fatal("caster's \"you cast a sorcery this turn\" condition is false after the origin changed control")
		}
		if conditionSatisfied(g, conditionContext{controller: game.Player2, source: source2}, castSorcery) {
			t.Fatal("origin's new controller counts as having cast the prepared spell")
		}
		cast := lastSnapshotEvent(t, g, game.EventSpellCast)
		if !cast.CardCharacteristicsKnown || cast.Controller != game.Player1 {
			t.Fatalf("cast event = %+v, want authoritative Player1 snapshot", cast)
		}
	}
	check(t, g, source1, source2)

	clone := g.Clone()
	cloneSource1, ok1 := permanentByObjectID(clone, source1.ObjectID)
	cloneSource2, ok2 := permanentByObjectID(clone, source2.ObjectID)
	if !ok1 || !ok2 {
		t.Fatal("cloned observer permanents missing")
	}
	check(t, clone, cloneSource1, cloneSource2)
}

func TestFaceDownMorphCastDoesNotLeakFrontFaceCharacteristics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := morphCreature(cost.Mana{cost.G})
	def.Colors = []color.Color{color.Green}
	def.Supertypes = []types.Super{types.Legendary}
	def.Subtypes = []types.Sub{types.Elf}
	cardID := addCardToHand(g, game.Player1, def)
	observer := addCombatCreaturePermanent(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 3)
	snapshotTestMainPhase(g)

	if !engine.applyAction(g, game.Player1, actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownMorph)) {
		t.Fatal("face-down cast failed")
	}
	cast := lastSnapshotEvent(t, g, game.EventSpellCast)
	if !cast.CardCharacteristicsKnown || cast.CardID != cardID || len(cast.CardSubtypes) != 0 ||
		len(cast.CardSupertypes) != 0 || len(cast.Colors) != 0 {
		t.Fatalf("face-down cast event = %+v, want authoritative creature with no subtypes, supertypes or colors", cast)
	}
	for _, tc := range preparedSnapshotObservers(false) {
		want := tc.want
		switch tc.name {
		case "creature spell":
			want = true
		case "sorcery spell":
			want = false
		default:
		}
		if got := triggerMatchesEvent(g, observer, &tc.pattern, cast); got != want {
			t.Errorf("face-down cast: %s observer matched = %v, want %v", tc.name, got, want)
		}
	}
}

// TestUnflaggedSpellCastEventKeepsFrontFaceFallback preserves the legacy
// fallback for spell-cast events that do not carry an authoritative snapshot.
func TestUnflaggedSpellCastEventKeepsFrontFaceFallback(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	observer := addCombatCreaturePermanent(g, game.Player1)
	elfID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Elf", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf},
	}})
	pattern := &game.TriggerPattern{
		Event: game.EventSpellCast, Controller: game.TriggerControllerYou,
		CardSelection: game.Selection{SubtypesAny: []types.Sub{types.Elf}},
	}
	event := game.Event{Kind: game.EventSpellCast, Controller: game.Player1, CardID: elfID}
	if !triggerMatchesEvent(g, observer, pattern, event) {
		t.Fatal("unflagged spell-cast event did not fall back to the front face")
	}
	event.CardCharacteristicsKnown = true
	if triggerMatchesEvent(g, observer, pattern, event) {
		t.Fatal("authoritative empty snapshot fell back to the front face")
	}
}

type snapshotWatcher struct {
	name      string
	pattern   game.TriggerPattern
	permanent *game.Permanent
}

// addSnapshotWatchers puts one "whenever you cast [filtered] spell" permanent
// per preparedSnapshotObservers pattern, plus a red-spell watcher, under player.
func addSnapshotWatchers(g *game.Game, player game.PlayerID, matchCopy bool) []snapshotWatcher {
	observers := preparedSnapshotObservers(matchCopy)
	observers = append(observers, snapshotObserver{name: "red spell", pattern: game.TriggerPattern{
		Event: game.EventSpellCast, Controller: game.TriggerControllerYou, MatchSpellCopy: matchCopy,
		CardSelection: game.Selection{ColorsAny: []color.Color{color.Red}},
	}})
	watchers := make([]snapshotWatcher, 0, len(observers))
	for _, observer := range observers {
		pattern := observer.pattern
		permanent := addTriggeredPermanent(g, player, &pattern,
			[]game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
		watchers = append(watchers, snapshotWatcher{name: observer.name, pattern: pattern, permanent: permanent})
	}
	return watchers
}

func checkSnapshotWatchers(t *testing.T, g *game.Game, label string, watchers []snapshotWatcher, want map[string]bool) {
	t.Helper()
	for _, watcher := range watchers {
		wantCount := 0
		if want[watcher.name] {
			wantCount = 1
		}
		if got := triggeredAbilitiesFrom(g, watcher.permanent.ObjectID); got != wantCount {
			t.Errorf("%s: %s watcher triggers = %d, want %d", label, watcher.name, got, wantCount)
		}
	}
}

// TestCopyStackObjectOfPreparedSpellKeepsEmptySnapshot covers the non-storm
// "copy target spell" route (CR 707.10): the copy has the prepared spell's
// copiable values, as modified by the copy effect's color exception, and is
// controlled by the copier, so neither the originating permanent's Elf,
// legendary and green characteristics nor its controller are attributed to it.
func TestCopyStackObjectOfPreparedSpellKeepsEmptySnapshot(t *testing.T) {
	for _, tc := range []struct {
		name      string
		copier    game.PlayerID
		setColors []color.Color
	}{
		{name: "caster copies", copier: game.Player1},
		{name: "opponent copies as red", copier: game.Player2, setColors: []color.Color{color.Red}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID, _ := addPreparedOriginPermanent(t, g, legendaryElfColorlessPrepareCard(false))
			snapshotTestMainPhase(g)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)) {
				t.Fatal("casting prepared spell failed")
			}
			prepared, ok := g.Stack.Peek()
			if !ok {
				t.Fatal("prepared spell missing from stack")
			}
			g.TriggerEventCursor = len(g.Events)
			watchers := map[game.PlayerID][]snapshotWatcher{
				game.Player1: addSnapshotWatchers(g, game.Player1, true),
				game.Player2: addSnapshotWatchers(g, game.Player2, true),
			}

			addEffectSpellToStack(g, tc.copier,
				game.CopyStackObject{Object: game.TargetStackObjectReference(0), SetColors: tc.setColors},
				[]game.Target{game.StackObjectTarget(prepared.ID)})
			engine.resolveTopOfStack(g, &TurnLog{})
			copyObj, ok := g.Stack.Peek()
			if !ok || !copyObj.Copy || copyObj.ID == prepared.ID || copyObj.Controller != tc.copier {
				t.Fatalf("top of stack = %+v, want copy controlled by %v", copyObj, tc.copier)
			}
			copied := lastSnapshotEvent(t, g, game.EventSpellCopied)
			if !copied.CardCharacteristicsKnown || copied.Controller != tc.copier || copied.CardID != cardID ||
				len(copied.CardSubtypes) != 0 || len(copied.CardSupertypes) != 0 ||
				!slices.Equal(copied.Colors, tc.setColors) || !slices.Equal(copied.CardTypes, []types.Card{types.Sorcery}) {
				t.Fatalf("copy event = %+v, want authoritative %v Sorcery controlled by %v", copied, tc.setColors, tc.copier)
			}

			engine.putTriggeredAbilitiesOnStack(g)
			want := map[string]bool{"sorcery spell": true}
			if tc.setColors == nil {
				want["colorless spell"] = true
			} else {
				want["red spell"] = true
			}
			checkSnapshotWatchers(t, g, "copier", watchers[tc.copier], want)
			other := game.Player2
			if tc.copier == game.Player2 {
				other = game.Player1
			}
			checkSnapshotWatchers(t, g, "non-copier", watchers[other], nil)
		})
	}
}

// TestSpellCopyEventKeepsCopyColorException covers a copy effect's color
// exception on the published copy snapshot (CR 707.9b): the exception becomes
// part of the copy's copiable values, so the copy event reports it, and a copy
// of that copy inherits it. It runs for a prepared spell and for a face-down
// morph spell through the CopyStackObject effect and the authored program copy
// operation. Copies of face-down spells still expose the hidden face's other
// characteristics (#3351), so only the color exception is asserted for them.
func TestSpellCopyEventKeepsCopyColorException(t *testing.T) {
	red := []color.Color{color.Red}
	for _, origin := range []string{"prepared", "face-down"} {
		for _, route := range []string{"copy effect", "authored program"} {
			t.Run(origin+"/"+route, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := permanentCopyEngine(t)
				snapshotTestMainPhase(g)
				if origin == "prepared" {
					cardID, _ := addPreparedOriginPermanent(t, g, legendaryElfColorlessPrepareCard(false))
					g.Players[game.Player1].ManaPool.Add(mana.C, 1)
					if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)) {
						t.Fatal("casting prepared spell failed")
					}
				} else {
					cardID := addCardToHand(g, game.Player1, lookupSnapshotTestCard(t, "Birchlore Rangers"))
					g.Players[game.Player1].ManaPool.Add(mana.G, 3)
					if !engine.applyAction(g, game.Player1, actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownMorph)) {
						t.Fatal("face-down cast failed")
					}
				}
				original, ok := g.Stack.Peek()
				if !ok || original.FaceDown != (origin == "face-down") {
					t.Fatalf("original spell = %+v, want %s spell on the stack", original, origin)
				}
				g.TriggerEventCursor = len(g.Events)
				watchers := addSnapshotWatchers(g, game.Player1, true)
				if origin == "face-down" {
					watchers = slices.DeleteFunc(watchers, func(watcher snapshotWatcher) bool {
						return watcher.name != "red spell" && watcher.name != "green spell" && watcher.name != "colorless spell"
					})
				}

				redCopy := resolveForkCopy(t, g, engine, original, red)
				first := lastSnapshotEvent(t, g, game.EventSpellCopied)
				if first.StackObjectID != redCopy.ID || !first.CardCharacteristicsKnown || !slices.Equal(first.Colors, red) {
					t.Fatalf("red copy event = %+v, want authoritative red snapshot of %v", first, redCopy.ID)
				}
				engine.putTriggeredAbilitiesOnStack(g)
				want := map[string]bool{"red spell": true, "sorcery spell": true}
				checkSnapshotWatchers(t, g, "red copy", watchers, want)

				var copyOfCopy *game.StackObject
				if route == "copy effect" {
					copyOfCopy = resolveForkCopy(t, g, engine, redCopy, nil)
				} else {
					copyOfCopy = copyPermanentSpell(t, engine, g, redCopy, game.Player1)
				}
				second := lastSnapshotEvent(t, g, game.EventSpellCopied)
				if second.StackObjectID != copyOfCopy.ID || !second.CardCharacteristicsKnown || !slices.Equal(second.Colors, red) {
					t.Fatalf("copy-of-copy event = %+v, want inherited red snapshot of %v", second, copyOfCopy.ID)
				}
				observer := addCombatCreaturePermanent(g, game.Player1)
				for _, watcher := range watchers {
					if got := triggerMatchesEvent(g, observer, &watcher.pattern, second); got != want[watcher.name] {
						t.Errorf("copy of copy: %s watcher matched = %v, want %v", watcher.name, got, want[watcher.name])
					}
				}
			})
		}
	}
}

// TestCastCopyOfExiledCardKeepsCardSnapshot covers casting a copy of a card
// in exile without paying its mana cost (CR 707.12, Isochron Scepter): the copy
// is cast, so it counts as a cast spell for "whenever you cast" triggers, and
// those triggers see exactly the exiled card's characteristics, including
// legitimately empty ones.
func TestCastCopyOfExiledCardKeepsCardSnapshot(t *testing.T) {
	for _, tc := range []struct {
		name  string
		def   *game.CardDef
		types []types.Card
		want  map[string]bool
	}{
		{
			name:  "colorless instant",
			def:   plainInstantDef("Colorless Insight"),
			types: []types.Card{types.Instant},
			want:  map[string]bool{"colorless spell": true},
		},
		{
			name: "green legendary Kindred Elf instant",
			def: &game.CardDef{CardFace: game.CardFace{
				Name:         "Legendary Elf Insight",
				ManaCost:     opt.Val(cost.Mana{cost.G}),
				Colors:       []color.Color{color.Green},
				Supertypes:   []types.Super{types.Legendary},
				Types:        []types.Card{types.Kindred, types.Instant},
				Subtypes:     []types.Sub{types.Elf},
				SpellAbility: opt.Val(game.AbilityContent{}),
			}},
			types: []types.Card{types.Kindred, types.Instant},
			want: map[string]bool{
				"Elf spell": true, "legendary spell": true, "green spell": true, "historic spell": true,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			scepter := addCombatPermanent(g, game.Player1, isochronScepterDef())
			obj := activatedObjFor(scepter)
			cardID := addExiledInstant(g, game.Player1, tc.def)
			imprintLink(g, obj, cardID)
			watchers := addSnapshotWatchers(g, game.Player1, false)
			opponentWatchers := addSnapshotWatchers(g, game.Player2, false)
			snapshotTestMainPhase(g)
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)

			engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			copyObj, ok := g.Stack.Peek()
			if !ok || !copyObj.Copy || copyObj.Controller != game.Player1 || !g.Players[game.Player1].Exile.Contains(cardID) {
				t.Fatalf("top of stack = %+v, want Player1's cast copy with the original still exiled", copyObj)
			}
			if copyObj.ManaSpentToCast != 0 || g.Players[game.Player1].ManaPool.Total() != 2 {
				t.Fatalf("copy mana spent = %d with pool %d, want a free cast leaving the pool at 2",
					copyObj.ManaSpentToCast, g.Players[game.Player1].ManaPool.Total())
			}
			cast := lastSnapshotEvent(t, g, game.EventSpellCast)
			if cast.StackObjectID != copyObj.ID || !cast.CardCharacteristicsKnown || cast.Controller != game.Player1 ||
				cast.PlayerEventOrdinalThisTurn != 1 || cast.ManaSpentToCast != opt.Val(0) ||
				!slices.Equal(cast.CardTypes, tc.types) || !slices.Equal(cast.CardSubtypes, tc.def.Subtypes) ||
				!slices.Equal(cast.CardSupertypes, tc.def.Supertypes) || !slices.Equal(cast.Colors, tc.def.Colors) {
				t.Fatalf("copy cast event = %+v, want authoritative free cast of %s", cast, tc.def.Name)
			}
			if got := spellsCastEarlierThisTurn(g); got != 1 {
				t.Fatalf("spells cast this turn after the copy = %d, want 1", got)
			}

			engine.putTriggeredAbilitiesOnStack(g)
			checkSnapshotWatchers(t, g, "caster", watchers, tc.want)
			checkSnapshotWatchers(t, g, "opponent", opponentWatchers, nil)
		})
	}
}
