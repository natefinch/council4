package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestCompiledLegacyReturnedPermanentConditions(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, text, subtype string
		blink               bool
		counters, tokens    int
	}{
		{"Defy Death", "Return target creature card from your graveyard to the battlefield. If it's an Angel, put two +1/+1 counters on it.", "Angel", false, 2, 0},
		{"Fearsome Awakening", "Return target creature card from your graveyard to the battlefield. If it's a Dragon, put two +1/+1 counters on it.", "Dragon", false, 2, 0},
		{"Return Upon the Tide", "Return target creature card from your graveyard to the battlefield. If it's an Elf, create two 1/1 green Elf Warrior creature tokens.", "Elf", false, 0, 2},
		{"Essence Flux", "Exile target creature you control, then return that card to the battlefield under its owner's control. If it's a Spirit, put a +1/+1 counter on it.", "Spirit", true, 1, 0},
	} {
		for _, matches := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/matches=%t", test.name, matches), func(t *testing.T) {
				t.Parallel()
				def := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: test.name, Layout: "normal", TypeLine: "Instant", OracleText: test.text,
				})
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				targetDef := &game.CardDef{CardFace: game.CardFace{Name: "Same Name", Types: []types.Card{types.Creature}}}
				if matches {
					targetDef.Subtypes = []types.Sub{types.Sub(test.subtype)}
				}
				decoy := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
					Name: "Same Name", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Sub(test.subtype)},
				}})
				targetID := addCardToGraveyard(g, game.Player1, targetDef)
				target := currentCardTarget(t, g, targetID)
				if test.blink {
					old := addCombatPermanent(g, game.Player1, targetDef)
					targetID = old.CardInstanceID
					target = game.PermanentTarget(old.ObjectID)
					// The departed object deliberately has the opposite subtype.
					// The new incarnation must use its own current values.
					change := game.ContinuousEffect{ID: g.IDGen.Next(), AffectedObjectID: old.ObjectID, Layer: game.LayerType}
					if matches {
						change.RemoveSubtypes = []types.Sub{types.Sub(test.subtype)}
					} else {
						change.AddSubtypes = []types.Sub{types.Sub(test.subtype)}
					}
					g.ContinuousEffects = append(g.ContinuousEffects, change)
				}
				obj := &game.StackObject{
					ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1,
					SourceCardID: addCardInstance(g, game.Player1, def), Targets: []game.Target{target},
				}
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				var returned *game.Permanent
				tokens := 0
				for _, permanent := range g.Battlefield {
					if permanent.CardInstanceID == targetID {
						returned = permanent
					}
					if permanent.Token && permanent.Controller == game.Player1 {
						tokens++
					}
				}
				if returned == nil {
					t.Fatal("target did not return")
				}
				wantCounters, wantTokens := 0, 0
				if matches {
					wantCounters, wantTokens = test.counters, test.tokens
				}
				if got := returned.Counters.Get(counter.PlusOnePlusOne); got != wantCounters || tokens != wantTokens {
					t.Fatalf("returned counters=%d tokens=%d, want %d/%d", got, tokens, wantCounters, wantTokens)
				}
				if decoy.Counters.Get(counter.PlusOnePlusOne) != 0 {
					t.Fatal("same-named opponent decoy was modified")
				}
			})
		}
	}
}
func TestCompiledReturnedCurrentCharacteristicsAndTargetSlot(t *testing.T) {
	t.Parallel()
	for _, addElf := range []bool{false, true} {
		t.Run(fmt.Sprintf("addElf=%t", addElf), func(t *testing.T) {
			t.Parallel()
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Returned Incarnation", Layout: "normal", TypeLine: "Instant",
				OracleText: "Tap target creature. Return target creature card from a graveyard to the battlefield under your control. You gain 1 life. If it's an Elf, put a +1/+1 counter on it.",
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			prior := addCombatCreaturePermanent(g, game.Player2, 2, 2)
			cardID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Returned Card", Types: []types.Card{types.Creature},
			}})
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1,
				SourceCardID: addCardInstance(g, game.Player1, def),
				Targets:      []game.Target{game.PermanentTarget(prior.ObjectID), currentCardTarget(t, g, cardID)},
			}
			sequence := def.SpellAbility.Val.Modes[0].Sequence
			put, ok := sequence[1].Primitive.(game.PutOnBattlefield)
			if !ok {
				t.Fatal("return did not lower to a battlefield entry")
			}
			ref := game.LinkedObjectReference(string(put.PublishLinked))
			// One owning resolution: the returned object's characteristics change
			// between its publication and the conditional consumer, and its owner
			// and controller are read through the same published link.
			resolved := append([]game.Instruction(nil), sequence[:2]...)
			if addElf {
				resolved = append(resolved, game.Instruction{Primitive: game.ApplyContinuous{
					Object:            opt.Val(ref),
					ContinuousEffects: []game.ContinuousEffect{{Layer: game.LayerType, AddSubtypes: []types.Sub{types.Elf}}},
				}})
			}
			resolved = append(resolved,
				game.Instruction{Primitive: game.GainLife{Player: game.ObjectOwnerReference(ref), Amount: game.Fixed(100)}},
				game.Instruction{Primitive: game.GainLife{Player: game.ObjectControllerReference(ref), Amount: game.Fixed(1000)}},
			)
			resolved = append(resolved, sequence[2:]...)
			NewEngine(nil).resolveInstructionSequence(g, obj, resolved, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			returned := permanentForCard(g, cardID)
			if returned == nil {
				t.Fatal("nonzero card target did not publish its entered object")
			}
			want := 0
			if addElf {
				want = 1
			}
			if got := returned.Counters.Get(counter.PlusOnePlusOne); got != want || !prior.Tapped {
				t.Fatalf("current counter=%d want %d, prior tapped=%t", got, want, prior.Tapped)
			}
			// Owner Player2 and controller Player1 were read from the entered object;
			// the unconditional 1 life still reaches only the controller.
			if g.Players[game.Player1].Life != 41+1000 || g.Players[game.Player2].Life != 40+100 {
				t.Fatalf("owner/controller or player isolation changed: p1=%d p2=%d",
					g.Players[game.Player1].Life, g.Players[game.Player2].Life)
			}
			if _, ok := resolveObjectReference(g, obj, ref); ok {
				t.Fatal("local entered-object link leaked past its owning resolution")
			}
		})
	}
}

func TestReturnedPublisherSkippedOrStaleTargetDoesNotReuseResult(t *testing.T) {
	t.Parallel()
	for _, skipped := range []bool{false, true} {
		t.Run(fmt.Sprintf("skipped=%t", skipped), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Elf", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf},
			}})
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1, SourceCardID: g.IDGen.Next(),
				Targets: []game.Target{currentCardTarget(t, g, cardID)}}
			decoy := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Elf", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf},
			}})
			key := game.LinkedKey("returned")
			rememberLinkedObject(g, linkedObjectSourceKey(g, obj, string(key)), permanentObjectBindingRef(decoy))
			publisher := game.Instruction{Primitive: game.PutOnBattlefield{
				Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}), PublishLinked: key,
			}}
			if skipped {
				addCardToHand(g, game.Player1, evidenceCard("Nonempty", 1))
				publisher.Condition = opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{ControllerHandEmpty: true})})
			} else {
				moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Hand)
				moveCardBetweenZones(g, game.Player1, cardID, zone.Hand, zone.Graveyard)
			}
			resolver := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}
			resolver.resolveInstruction(&publisher)
			ref := game.LinkedObjectReference(string(key))
			if _, ok := resolveObjectReference(g, obj, ref); ok {
				t.Fatal("skipped/failed producer reused a previous publication or current card")
			}
			resolver.resolveInstruction(&game.Instruction{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
				Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
					Object: opt.Val(ref), ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Elf}}),
				})}),
			})
			if g.Players[game.Player1].Life != 40 {
				t.Fatal("failed producer's condition fired")
			}
		})
	}
}

func TestPublishedBlinkSourceAndFreshObjectLKI(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Blinking Source", Layout: "normal", TypeLine: "Creature — Elf",
		OracleText: "{2}: Exile this creature, then return it to the battlefield under its owner's control. If it's an Elf, put a +1/+1 counter on it.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	old := addCombatPermanent(g, game.Player1, def)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), AffectedObjectID: old.ObjectID, Layer: game.LayerType, RemoveSubtypes: []types.Sub{types.Elf},
	})
	obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackActivatedAbility,
		SourceID: old.ObjectID, SourceCardID: old.CardInstanceID, Controller: game.Player1}
	engine := NewEngine(nil)
	content := def.ActivatedAbilities[0].Content
	// Hold the same local-product frame resolveInstructionSequence uses open so
	// the published link's identity and LKI are checked during its valid lifetime.
	restore := resolveSequenceInOwningFrame(g, obj, content.Modes[0].Sequence)
	put, ok := content.Modes[0].Sequence[1].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatal("source blink did not lower to a battlefield entry")
	}
	ref := game.LinkedObjectReference(string(put.PublishLinked))
	result, ok := resolveObjectReference(g, obj, ref)
	if !ok || result.permanent == nil || result.permanent.ObjectID == old.ObjectID ||
		result.permanent.Counters.Get(counter.PlusOnePlusOne) != 1 {
		t.Fatal("source blink did not condition and modify the new incarnation")
	}
	returned := result.permanent
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), AffectedObjectID: returned.ObjectID, Layer: game.LayerType,
		AddTypes: []types.Card{types.Artifact},
	})
	destroyPermanent(g, returned.ObjectID)
	card, _ := g.GetCardInstance(returned.CardInstanceID)
	r := effectResolver{engine: engine, game: g, obj: obj, log: &TurnLog{}}
	reentered, entered := r.putResolvedCardOnBattlefieldValue(card, zone.Graveyard, game.Player2, nil, permanentCreationOptions{})
	if !entered || reentered.ObjectID == returned.ObjectID {
		t.Fatal("test failed to create a third incarnation")
	}
	result, ok = resolveObjectReference(g, obj, ref)
	if !ok || result.permanent != nil || result.snapshot.ObjectID != returned.ObjectID ||
		!resolvedObjectHasType(g, &result, types.Artifact) {
		t.Fatal("published returned object lost its own departure LKI or rebound to the later object")
	}
	restore()
	if _, ok := resolveObjectReference(g, obj, ref); ok {
		t.Fatal("local returned-object link leaked past its owning resolution")
	}
}

// resolveSequenceInOwningFrame resolves sequence inside the runtime's own
// local-product frame and leaves it open; the caller closes it with the
// returned function after observing products during their valid lifetime.
func resolveSequenceInOwningFrame(g *game.Game, obj *game.StackObject, sequence []game.Instruction) func() {
	restore := enterLocalProductFrame(g, obj, sequence)
	resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	for i := range sequence {
		resolver.resolveInstruction(&sequence[i])
	}
	return restore
}

func TestBlinkInputCardIncarnationCannotLeaveAndReenterExile(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatCreaturePermanent(g, game.Player1, 2, 2)
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1, SourceCardID: g.IDGen.Next(),
		Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
	r := effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}
	r.resolveInstruction(&game.Instruction{Primitive: game.MovePermanent{
		Object: game.TargetPermanentReference(0), Destination: zone.Exile, PublishLinked: "departed",
	}})
	moveCardBetweenZones(g, game.Player1, target.CardInstanceID, zone.Exile, zone.Hand)
	moveCardBetweenZones(g, game.Player1, target.CardInstanceID, zone.Hand, zone.Exile)
	r.resolveInstruction(&game.Instruction{Primitive: game.PutOnBattlefield{
		Source: game.LinkedBattlefieldSource("departed"), PublishLinked: "returned",
	}})
	if len(g.Battlefield) != 0 || len(linkedObjects(g, linkedObjectSourceKey(g, obj, "returned"))) != 0 {
		t.Fatal("blink resurrected a later exile incarnation or fabricated a returned result")
	}
}

func TestCompiledReturnedPublisherReplacementOrCounter(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, text string
		blink      bool
		countered  bool
	}{
		{"return replacement", "Return target creature card from your graveyard to the battlefield. If it's an Elf, create two 1/1 green Elf Warrior creature tokens.", false, false},
		{"blink replacement", "Exile target creature you control, then return that card to the battlefield under its owner's control. If it's a Spirit, put a +1/+1 counter on it.", true, false},
		{"countered return", "Return target creature card from your graveyard to the battlefield. If it's an Elf, create two 1/1 green Elf Warrior creature tokens.", false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: test.name, Layout: "normal", TypeLine: "Instant", OracleText: test.text,
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			targetDef := &game.CardDef{CardFace: game.CardFace{
				Name: "Subject", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf, types.Spirit},
			}}
			cardID := addCardToGraveyard(g, game.Player1, targetDef)
			target := currentCardTarget(t, g, cardID)
			fromZone := zone.Graveyard
			destination := zone.Exile
			if test.blink {
				old := addCombatPermanent(g, game.Player1, targetDef)
				cardID = old.CardInstanceID
				target = game.PermanentTarget(old.ObjectID)
				fromZone, destination = zone.Exile, zone.Hand
			}
			sourceID := addCardInstance(g, game.Player1, def)
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1,
				SourceID: sourceID, SourceCardID: sourceID, Targets: []game.Target{target},
			}
			g.Stack.Push(obj)
			if test.countered {
				if !counterStackObject(g, obj.ID) {
					t.Fatal("test spell was not countered")
				}
			} else {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: fromZone,
					MatchToZone: true, ToZone: zone.Battlefield, ReplaceToZone: destination,
				})
			}
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			if len(g.Battlefield) != 0 {
				t.Fatal("ineffective publisher returned a permanent or executed its gated token rider")
			}
			if !test.countered {
				actual, ok := cardZone(g, cardID)
				if !ok || actual != destination {
					t.Fatalf("replacement destination=%v/%t, want %v", actual, ok, destination)
				}
			}
		})
	}
}

func TestPublishedCardVersionAndPermanentFallbackStayDistinct(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Card", Types: []types.Card{types.Creature},
	}})
	card, _ := g.GetCardInstance(cardID)
	moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Hand)
	moveCardBetweenZones(g, game.Player1, cardID, zone.Hand, zone.Graveyard)
	ref := game.LinkedObjectRef{CardID: cardID, CardZoneVersion: card.ZoneVersion}
	if _, ok := resolveLinkedObjectRef(g, ref); !ok {
		t.Fatal("current reached card did not resolve")
	}
	moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Hand)
	if _, ok := resolveLinkedObjectRef(g, ref); ok {
		t.Fatal("a card link followed a different zone incarnation")
	}
	if _, ok := resolveLinkedObjectRef(g, game.LinkedObjectRef{ObjectID: g.IDGen.Next(), CardID: cardID}); ok {
		t.Fatal("unknown permanent identity silently fell back to printed card characteristics")
	}
}

func TestCompiledReturnedSubjectsRemainIndependent(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Independent Returns", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Return target creature card from your graveyard to the battlefield. If it's an Elf, create a 1/1 green Elf Warrior creature token. Return target creature card from your graveyard to the battlefield. If it's a Dragon, put a +1/+1 counter on it.",
	})
	for mask := range 4 {
		t.Run(fmt.Sprintf("matches=%d", mask), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			first := &game.CardDef{CardFace: game.CardFace{Name: "Same Name", Types: []types.Card{types.Creature}}}
			second := &game.CardDef{CardFace: game.CardFace{Name: "Same Name", Types: []types.Card{types.Creature}}}
			if mask&1 != 0 {
				first.Subtypes = []types.Sub{types.Elf}
			}
			if mask&2 != 0 {
				second.Subtypes = []types.Sub{types.Dragon}
			}
			firstID := addCardToGraveyard(g, game.Player1, first)
			secondID := addCardToGraveyard(g, game.Player1, second)
			sourceID := addCardInstance(g, game.Player1, def)
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1,
				SourceID: sourceID, SourceCardID: sourceID,
				Targets: []game.Target{currentCardTarget(t, g, firstID), currentCardTarget(t, g, secondID)},
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			tokens, counters := 0, 0
			for _, permanent := range g.Battlefield {
				if permanent.Token {
					tokens++
				}
				if permanent.CardInstanceID == firstID && permanent.Counters.Get(counter.PlusOnePlusOne) != 0 {
					t.Fatal("second return modified the first returned object")
				}
				if permanent.CardInstanceID == secondID {
					counters = permanent.Counters.Get(counter.PlusOnePlusOne)
				}
			}
			if tokens != mask&1 || counters != (mask>>1)&1 {
				t.Fatalf("tokens=%d counters=%d, want %d/%d", tokens, counters, mask&1, (mask>>1)&1)
			}
		})
	}
}
