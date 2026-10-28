package rules

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func copiedMaxSpeedDef(pattern game.TriggerPattern) *game.CardDef {
	triggerType := game.TriggerWhenever
	if pattern.Event == game.EventBeginningOfStep {
		triggerType = game.TriggerAt
	}
	trigger := game.TriggeredAbility{
		Text:    "You gain 1 life.",
		Trigger: game.TriggerCondition{Type: triggerType, Pattern: pattern},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Max speed blueprint", Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{
			Condition: opt.Val(game.Condition{ControllerHasMaxSpeed: true}),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility, AffectedSource: true,
				AddAbilities: []game.Ability{&trigger},
			}},
		}},
	}}
}

func becomeStaticCopy(t testing.TB, g *game.Game, source, blueprint *game.Permanent, untilEndOfTurn bool) {
	t.Helper()
	r := &effectResolver{engine: NewEngine(nil), game: g, log: &TurnLog{}, obj: &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility,
		SourceID: source.ObjectID, Controller: source.Controller,
		Targets: []game.Target{game.PermanentTarget(blueprint.ObjectID)},
	}}
	if !handleBecomeCopy(r, game.BecomeCopy{
		Object: game.TargetPermanentReference(0), UntilEndOfTurn: untilEndOfTurn,
	}).succeeded {
		t.Fatal("BecomeCopy failed")
	}
}

func copiedStaticTriggerCount(g *game.Game, source *game.Permanent) int {
	count := 0
	for _, ability := range permanentEffectiveAbilitiesView(g, source) {
		if _, ok := ability.(*game.TriggeredAbility); ok {
			count++
		}
	}
	return count
}

func TestCopiedStaticMaxSpeedTriggers(t *testing.T) {
	for _, event := range []game.Event{
		{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player3},
		{Kind: game.EventCardDrawn, Player: game.Player3},
	} {
		for _, speed := range []int{0, 1, 3, 4} {
			t.Run(fmt.Sprintf("%v/speed%d", event.Kind, speed), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				blueprint := addCombatPermanent(g, game.Player1, copiedMaxSpeedDef(game.TriggerPattern{
					Event: event.Kind, Step: event.Step, Player: game.TriggerPlayerYou,
				}))
				source := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
					Name: "Physical copy", Types: []types.Card{types.Artifact},
				}})
				source.Controller = game.Player3
				g.Players[game.Player1].Speed = 4
				g.Players[game.Player3].Speed = speed
				becomeStaticCopy(t, g, source, blueprint, false)
				if copiedStaticTriggerCount(g, blueprint) != 1 {
					t.Fatal("uncopied max-speed trigger control failed")
				}
				if permanentEffectiveName(g, source) != permanentEffectiveName(g, blueprint) {
					t.Fatal("copy did not acquire the blueprint's name")
				}
				want := 0
				if speed == 4 {
					want = 1
				}
				if got := copiedStaticTriggerCount(g, source); got != want {
					t.Fatalf("copy has %d triggers, want %d", got, want)
				}
				emitEvent(g, event)
				if placed := engine.putTriggeredAbilitiesOnStack(g); placed != (want == 1) {
					t.Fatalf("placed=%v, want %v", placed, want == 1)
				}
				if want == 0 {
					return
				}
				obj, ok := g.Stack.Peek()
				if !ok || obj.SourceID != source.ObjectID || obj.SourceCardID != source.CardInstanceID ||
					obj.Controller != game.Player3 || obj.InlineTrigger == nil ||
					obj.InlineTrigger.Trigger.InterveningCondition.Exists {
					t.Fatal("trigger lost physical source/current controller or substituted intervening-if")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				if g.Players[game.Player3].Life != 41 || g.Players[game.Player1].Life != 40 ||
					g.Players[game.Player2].Life != 40 {
					t.Fatal("copied trigger did not resolve for its current controller alone")
				}
			})
		}
	}
}

func TestCopiedStaticCapturedTriggerSurvivesLoss(t *testing.T) {
	for _, loss := range []string{"speed before placement", "speed after placement", "blueprint departure", "copy departure", "queued clone"} {
		t.Run(loss, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			blueprint := addCombatPermanent(g, game.Player1, copiedMaxSpeedDef(game.TriggerPattern{
				Event: game.EventCardDrawn, Player: game.TriggerPlayerYou,
			}))
			source := addCombatPermanent(g, game.Player3, &game.CardDef{})
			g.Players[game.Player3].Speed = 4
			becomeStaticCopy(t, g, source, blueprint, false)
			emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player3})
			if loss == "speed before placement" {
				g.Players[game.Player3].Speed = 0
			}
			if loss == "blueprint departure" {
				movePermanentToZone(g, blueprint, zone.Exile)
				if copiedStaticTriggerCount(g, source) != 1 {
					t.Fatal("blueprint departure erased the copy's static body")
				}
			}
			if loss == "copy departure" {
				movePermanentToZone(g, source, zone.Exile)
			}
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("captured trigger disappeared before placement")
			}
			if loss == "queued clone" {
				cloned := g.Clone()
				cloneSource, _ := permanentByObjectID(cloned, source.ObjectID)
				movePermanentToZone(cloned, cloneSource, zone.Exile)
				cloned.Players[game.Player3].Speed = 0
				engine.resolveTopOfStack(cloned, &TurnLog{})
				if cloned.Players[game.Player3].Life != 41 || g.Players[game.Player3].Life != 40 ||
					g.Stack.Size() != 1 {
					t.Fatal("queued copied trigger did not survive clone independently")
				}
			}
			g.Players[game.Player3].Speed = 0
			engine.resolveTopOfStack(g, &TurnLog{})
			if g.Players[game.Player3].Life != 41 {
				t.Fatal("queued max-speed trigger was incorrectly gated at resolution")
			}
			emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player3})
			if engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("unavailable static grant produced another trigger")
			}
		})
	}
}

func TestCopiedStaticSourcesReplaceIntrinsicAndMergedBodies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	blueprint := addCombatPermanent(g, game.Player1, copiedMaxSpeedDef(game.TriggerPattern{
		Event: game.EventCardDrawn, Player: game.TriggerPlayerYou,
	}))
	intrinsic := &game.CardDef{CardFace: game.CardFace{
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility, AffectedSource: true, AddKeywords: []game.Keyword{game.Flying},
			}},
		}},
	}}
	source := addCombatPermanent(g, game.Player2, intrinsic)
	source.Controller = game.Player3
	mergedID := addCardToHand(g, game.Player2, intrinsic)
	source.MergedCards = []game.MergedCard{{CardInstanceID: mergedID}}
	if !hasKeyword(g, source, game.Flying) {
		t.Fatal("uncopied intrinsic control failed")
	}
	g.Players[game.Player3].Speed = 4
	becomeStaticCopy(t, g, source, blueprint, false)
	components := 0
	visitPermanentStaticAbilityComponents(g, source, func(component permanentAbilityComponent) {
		components++
		if component.cardID != source.CardInstanceID ||
			component.originID != source.CardInstanceID || component.faceIndex != source.Face {
			t.Fatal("copied component adopted blueprint or merged identity")
		}
	})
	if components != 1 || hasKeyword(g, source, game.Flying) || copiedStaticTriggerCount(g, source) != 1 {
		t.Fatal("copy retained overwritten intrinsic/merged static effects or omitted copied body")
	}
	g.ContinuousEffects = nil
	if !hasKeyword(g, source, game.Flying) {
		t.Fatal("removing copy did not restore physical/merged static sources")
	}
}

func TestCopiedStaticUsesCurrentFaces(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	back := copiedMaxSpeedDef(game.TriggerPattern{Event: game.EventCardDrawn, Player: game.TriggerPlayerYou})
	blueprint := addCombatPermanent(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Blueprint front"},
		Back:     opt.Val(back.CardFace), Layout: game.LayoutTransform,
	})
	blueprint.Face = game.FaceBack
	source := addCombatPermanent(g, game.Player3, &game.CardDef{
		CardFace: game.CardFace{Name: "Physical front", StaticAbilities: []game.StaticAbility{game.FlyingStaticBody}},
		Back: opt.Val(game.CardFace{
			Name: "Physical back", StaticAbilities: []game.StaticAbility{game.ReachStaticBody},
		}),
		Layout: game.LayoutTransform,
	})
	source.Face = game.FaceBack
	g.Players[game.Player3].Speed = 4
	becomeStaticCopy(t, g, source, blueprint, true)
	visitPermanentStaticAbilityComponents(g, source, func(component permanentAbilityComponent) {
		if component.faceIndex != game.FaceBack || component.cardID != source.CardInstanceID {
			t.Fatal("copy replaced physical face identity")
		}
	})
	if copiedStaticTriggerCount(g, source) != 1 || hasKeyword(g, source, game.Reach) ||
		permanentEffectiveName(g, source) != back.Name {
		t.Fatal("copy read wrong blueprint face or retained intrinsic back-face statics")
	}
	expireCleanupDurations(g)
	if !hasKeyword(g, source, game.Reach) || hasKeyword(g, source, game.Flying) {
		t.Fatal("copy expiry did not restore the physical permanent's current face")
	}
}

func TestCopiedStaticFramesControlAndClone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	blueprint := addCombatPermanent(g, game.Player1, copiedMaxSpeedDef(game.TriggerPattern{
		Event: game.EventCardDrawn, Player: game.TriggerPlayerYou,
	}))
	source := addCombatPermanent(g, game.Player2, &game.CardDef{})
	g.Players[game.Player3].Speed = 4
	becomeStaticCopy(t, g, source, blueprint, false)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), AffectedObjectID: source.ObjectID, Layer: game.LayerControl,
		NewController: opt.Val(game.Player3),
	})
	check := func(g *game.Game, source *game.Permanent, want int) {
		t.Helper()
		unframed := effectivePermanentValues(g, source)
		g.BeginStaticSourceFrame()
		framed := effectivePermanentValues(g, source)
		again := effectivePermanentValues(g, source)
		g.EndStaticSourceFrame()
		if !reflect.DeepEqual(unframed.abilities, framed.abilities) ||
			!reflect.DeepEqual(framed.abilities, again.abilities) ||
			framed.controller != unframed.controller || copiedStaticTriggerCount(g, source) != want {
			t.Fatal("copied source frame differed from uncached evaluation")
		}
	}
	check(g, source, 1)
	cloned := g.Clone()
	cloneSource, _ := permanentByObjectID(cloned, source.ObjectID)
	cloned.Players[game.Player3].Speed = 0
	check(cloned, cloneSource, 0)
	check(g, source, 1)
	cloned.ContinuousEffects = cloned.ContinuousEffects[:1]
	cloned.Players[game.Player2].Speed = 4
	check(cloned, cloneSource, 1)
	if effectiveController(cloned, cloneSource) != game.Player2 ||
		effectiveController(g, source) != game.Player3 {
		t.Fatal("control change or clone isolation failed")
	}
	movePermanentToZone(cloned, cloneSource, zone.Exile)
	for _, discovered := range staticAbilitySources(cloned) {
		if discovered.permanent != nil && discovered.permanent.ObjectID == source.ObjectID {
			t.Fatal("departed copy remained a static source")
		}
	}
	check(g, source, 1)
}

func TestCopiedStaticFaceDownTokenAndCopyChain(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	blueprint := addCombatPermanent(g, game.Player1, copiedMaxSpeedDef(game.TriggerPattern{
		Event: game.EventCardDrawn, Player: game.TriggerPlayerYou,
	}))
	g.Players[game.Player3].Speed = 4
	source := addCombatPermanent(g, game.Player3, &game.CardDef{})
	becomeStaticCopy(t, g, source, blueprint, false)
	chain := addCombatPermanent(g, game.Player3, &game.CardDef{})
	becomeStaticCopy(t, g, chain, source, false)
	def, ok := permanentCopyDef(g, source)
	if !ok {
		t.Fatal("copy values missing")
	}
	token, ok := createTokenPermanent(g, game.Player3, def)
	if !ok {
		t.Fatal("copy token creation failed")
	}
	becomeStaticCopy(t, g, token, blueprint, false)
	for _, permanent := range []*game.Permanent{source, chain, token} {
		if copiedStaticTriggerCount(g, permanent) != 1 {
			t.Fatal("copy chain or token omitted static body")
		}
	}
	visitPermanentStaticAbilityComponents(g, token, func(component permanentAbilityComponent) {
		if component.cardID != 0 || component.originID != token.ObjectID {
			t.Fatal("copied token borrowed blueprint card identity")
		}
	})
	source.FaceDown = true
	if copiedStaticTriggerCount(g, source) != 0 ||
		copiedStaticTriggerCount(g, chain) != 1 || copiedStaticTriggerCount(g, token) != 1 {
		t.Fatal("face-down status leaked into independent copies or retained hidden static body")
	}
	source.FaceDown = false
	source.PhasedOut = true
	for _, discovered := range staticAbilitySources(g) {
		if discovered.permanent == source {
			t.Fatal("phased-out copy remained a static source")
		}
	}
	source.PhasedOut = false
	if copiedStaticTriggerCount(g, source) != 1 {
		t.Fatal("turning face up/phasing in did not restore copied statics")
	}
}

func TestCopiedStaticCopyOrderingAndExpiry(t *testing.T) {
	for _, ordering := range []string{"timestamp", "dependency"} {
		t.Run(ordering, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player3].Speed = 4
			blueprint := addCombatPermanent(g, game.Player1, copiedMaxSpeedDef(game.TriggerPattern{
				Event: game.EventCardDrawn, Player: game.TriggerPlayerYou,
			}))
			source := addCombatPermanent(g, game.Player3, &game.CardDef{})
			becomeStaticCopy(t, g, source, blueprint, false)
			later := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Later blueprint", StaticAbilities: []game.StaticAbility{{
					ContinuousEffects: []game.ContinuousEffect{{
						Layer: game.LayerAbility, AffectedSource: true, AddKeywords: []game.Keyword{game.Flying},
					}},
				}},
			}})
			becomeStaticCopy(t, g, source, later, true)
			if !hasKeyword(g, source, game.Flying) || copiedStaticTriggerCount(g, source) != 0 {
				t.Fatal("latest copy did not replace the earlier static body")
			}
			if ordering == "dependency" {
				g.ContinuousEffects[0].DependsOn = []id.ID{g.ContinuousEffects[1].ID}
				if hasKeyword(g, source, game.Flying) || copiedStaticTriggerCount(g, source) != 1 {
					t.Fatal("static source discovery ignored copy-layer dependency order")
				}
			}
			expireCleanupDurations(g)
			if hasKeyword(g, source, game.Flying) || copiedStaticTriggerCount(g, source) != 1 {
				t.Fatal("copy expiry did not restore the previous static body")
			}
		})
	}
}

func TestCopiedStaticConditionReadsEarlierLayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := copiedMaxSpeedDef(game.TriggerPattern{Event: game.EventCardDrawn, Player: game.TriggerPlayerYou})
	def.StaticAbilities[0].Condition.Val.Object = opt.Val(game.SourcePermanentReference())
	def.StaticAbilities[0].Condition.Val.ObjectMatches = opt.Val(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
	})
	blueprint := addCombatPermanent(g, game.Player1, def)
	source := addCombatPermanent(g, game.Player3, vanillaCreature("Physical creature", 2, 2))
	g.Players[game.Player3].Speed = 4
	becomeStaticCopy(t, g, source, blueprint, false)
	if copiedStaticTriggerCount(g, source) != 0 {
		t.Fatal("condition read the physical creature instead of copied artifact")
	}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), AffectedObjectID: source.ObjectID,
		Layer: game.LayerType, AddTypes: []types.Card{types.Creature},
	})
	if copiedStaticTriggerCount(g, source) != 1 {
		t.Fatal("condition did not observe the earlier type layer")
	}
	g.BeginStaticSourceFrame()
	if copiedStaticTriggerCount(g, source) != 1 {
		t.Fatal("condition's before-layer cache changed the result")
	}
	g.EndStaticSourceFrame()
}

func TestCopiedStaticGroupUsesPhysicalSourceAndLiveController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	blueprint := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Grant blueprint",
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				}),
				AddKeywords: []game.Keyword{game.Haste},
			}},
		}},
	}})
	source := addCombatPermanent(g, game.Player2, &game.CardDef{})
	recipient := addCombatCreaturePermanent(g, game.Player3)
	ownerCreature := addCombatCreaturePermanent(g, game.Player2)
	becomeStaticCopy(t, g, source, blueprint, false)
	movePermanentToZone(g, blueprint, zone.Exile)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), AffectedObjectID: source.ObjectID,
		Layer: game.LayerControl, NewController: opt.Val(game.Player3),
	})
	if !hasKeyword(g, recipient, game.Haste) || hasKeyword(g, ownerCreature, game.Haste) {
		t.Fatal("copied group used blueprint identity or physical owner instead of live controller")
	}
	g.BeginStaticSourceFrame()
	if !hasKeyword(g, recipient, game.Haste) {
		t.Fatal("copied group disappeared in a static-source frame")
	}
	g.EndStaticSourceFrame()
	movePermanentToZone(g, source, zone.Exile)
	if hasKeyword(g, recipient, game.Haste) {
		t.Fatal("copied group effect survived its source's departure")
	}
}
