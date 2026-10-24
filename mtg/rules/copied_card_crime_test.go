package rules

import (
	"slices"
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

// atKnifepointFixture is an unregistered proving fixture for At Knifepoint
// (oracle bddb2064-afdb-4184-ae4a-633745e64a3c, Scryfall revision
// 897d594d-b5b0-43dc-b877-b483942416ce). It is not admitted to the card
// registry; it only proves the shared crime boundary (CR 700.13, 707.12).
//
//	{1}{B}{R} Enchantment
//	During your turn, outlaws you control have first strike. (Assassins,
//	Mercenaries, Pirates, Rogues, and Warlocks are outlaws.)
//	Whenever you commit a crime, create a 1/1 red Mercenary creature token with
//	"{T}: Target creature you control gets +1/+0 until end of turn. Activate
//	only as a sorcery." This ability triggers only once each turn.
func atKnifepointFixture() *game.CardDef {
	outlaws := game.Selection{SubtypesAny: []types.Sub{
		types.Assassin, types.Mercenary, types.Pirate, types.Rogue, types.Warlock,
	}}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:     "At Knifepoint",
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B, cost.R}),
			Colors:   []color.Color{color.Black, color.Red},
			Types:    []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{{
				Text: "During your turn, outlaws you control have first strike.",
				Condition: opt.Val(game.Condition{
					Text:                 "During your turn",
					SourceControllerTurn: true,
				}),
				ContinuousEffects: []game.ContinuousEffect{{
					Duration:    game.DurationPermanent,
					Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), outlaws),
					Layer:       game.LayerAbility,
					AddKeywords: []game.Keyword{game.FirstStrike},
				}},
			}},
			TriggeredAbilities: []game.TriggeredAbility{{
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:  game.EventCrimeCommitted,
						Player: game.TriggerPlayerYou,
					},
				},
				MaxTriggersPerTurn: 1,
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.CreateToken{
						Amount: game.Fixed(1),
						Source: game.TokenDef(atKnifepointMercenaryToken()),
					},
				}}}.Ability(),
			}},
			OracleText: `
			During your turn, outlaws you control have first strike. (Assassins, Mercenaries, Pirates, Rogues, and Warlocks are outlaws.)
			Whenever you commit a crime, create a 1/1 red Mercenary creature token with "{T}: Target creature you control gets +1/+0 until end of turn. Activate only as a sorcery." This ability triggers only once each turn.
		`,
		},
	}
}

func atKnifepointMercenaryToken() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Mercenary",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Mercenary},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:            "{T}: Target creature you control gets +1/+0 until end of turn. Activate only as a sorcery.",
			AdditionalCosts: cost.Tap,
			ZoneOfFunction:  zone.Battlefield,
			Timing:          game.SorceryOnly,
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature you control",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.ModifyPT{
						Object:         game.TargetPermanentReference(0),
						PowerDelta:     game.Fixed(1),
						ToughnessDelta: game.Fixed(0),
						Duration:       game.DurationUntilEndOfTurn,
					},
				}},
			}.Ability(),
		}},
	}}
}

// crimeChoiceAgent answers choices through a test callback, falling back to
// the request default.
type crimeChoiceAgent func(game.ChoiceRequest) ([]int, bool)

func (crimeChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a crimeChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if selected, ok := a(request); ok {
		return selected
	}
	return request.DefaultSelection
}

// retargetTo accepts optional copy retargeting and chooses the target option
// selecting exactly want.
func retargetTo(want ...game.Target) crimeChoiceAgent {
	return func(request game.ChoiceRequest) ([]int, bool) {
		if request.Kind == game.ChoiceMay {
			return []int{1}, true
		}
		for _, option := range request.Options {
			if len(option.Targets) > 0 && slices.Equal(option.Targets, want) {
				return []int{option.Index}, true
			}
		}
		return nil, false
	}
}

func crimeTargetSpell(name string, specs ...game.TargetSpec) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		SpellAbility: opt.Val(game.Mode{
			Targets: specs,
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Player: game.ControllerReference(), Amount: game.Fixed(1),
			}}},
		}.Ability()),
	}}
}

func playerTargets(minimum, maximum int) game.TargetSpec {
	return game.TargetSpec{MinTargets: minimum, MaxTargets: maximum, Constraint: "player", Allow: game.TargetAllowPlayer}
}

func permanentTargetSpec() game.TargetSpec {
	return game.TargetSpec{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent}
}

func stackTargetSpec() game.TargetSpec {
	return game.TargetSpec{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{
			game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility,
		}}}
}

// sameTargetObjects compares chosen targets by object identity, ignoring the
// zone version the engine records for card targets.
func sameTargetObjects(got, want []game.Target) bool {
	return slices.EqualFunc(got, want, func(a, b game.Target) bool {
		a.CardZoneVersion, a.CardZoneVersionSet = 0, false
		b.CardZoneVersion, b.CardZoneVersionSet = 0, false
		return a == b
	})
}

func graveyardCardTargetSpec() game.TargetSpec {
	return game.TargetSpec{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowCard, TargetZone: zone.Graveyard}
}

// castExiledCardCopy casts a copy of the exiled card through the shared
// resolution-time copy-card casting operation (CR 707.12).
func castExiledCardCopy(engine *Engine, g *game.Game, controller game.PlayerID, cardID id.ID, targets []game.Target) bool {
	card, _ := g.GetCardInstance(cardID)
	return engine.castFreeOptionWithCopy(g, nil, controller, card, action.CastSpellAction{
		CardID: cardID, SourceZone: zone.Exile, Face: game.FaceFront, Targets: targets,
	}, false, true, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

func preparedCrimeCard(specs ...game.TargetSpec) *game.CardDef {
	def := prepareCard()
	alternate := def.Alternate.Val
	alternate.Name = "Prepared Crime"
	alternate.Types = []types.Card{types.Instant}
	alternate.SpellAbility = opt.Val(game.Mode{
		Targets: specs,
		Sequence: []game.Instruction{{Primitive: game.GainLife{
			Player: game.ControllerReference(), Amount: game.Fixed(1),
		}}},
	}.Ability())
	def.Alternate = opt.Val(alternate)
	return def
}

// addPreparedPermanent puts a prepared permanent whose alternate face is the
// given instant onto the battlefield under controller with {W} available.
func addPreparedPermanent(g *game.Game, controller game.PlayerID, def *game.CardDef) *game.Permanent {
	cardID := g.IDGen.Next()
	card := &game.CardInstance{ID: cardID, Owner: controller, Def: def}
	g.CardInstances[cardID] = card
	permanent, ok := createCardPermanent(g, card, controller, zone.Hand)
	if !ok || !permanent.Prepared {
		panic("prepared fixture permanent creation failed")
	}
	addBasicLandPermanent(g, controller, types.Plains)
	return permanent
}

func castPreparedCopy(engine *Engine, g *game.Game, controller game.PlayerID, permanent *game.Permanent, targets []game.Target) bool {
	g.Turn.PriorityPlayer = controller
	return engine.applyAction(g, controller, action.CastSpellFaceFromZone(
		permanent.CardInstanceID, zone.Battlefield, game.FaceAlternate, targets, 0, nil))
}

func mercenaryTokens(g *game.Game, controller game.PlayerID) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.TokenDef != nil && permanent.TokenDef.Name == "Mercenary" && permanent.Controller == controller {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}

func resolveKnifepointTriggers(t *testing.T, engine *Engine, g *game.Game) int {
	t.Helper()
	if !engine.putTriggeredAbilitiesOnStack(g) {
		return 0
	}
	resolved := 0
	for {
		top, ok := g.Stack.Peek()
		if !ok || top.Kind != game.StackTriggeredAbility {
			return resolved
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		resolved++
	}
}

// TestCopiedCardCrimeIsochronCopyCommitsCrime is the regression for a genuine
// cast of a copy of an exiled targeted instant (CR 707.12): it is cast, so it
// commits a crime (CR 700.13) and triggers At Knifepoint once.
func TestCopiedCardCrimeIsochronCopyCommitsCrime(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	addCombatPermanent(g, game.Player1, atKnifepointFixture())
	scepter := addCombatPermanent(g, game.Player1, isochronScepterDef())
	obj := activatedObjFor(scepter)
	cardID := addExiledInstant(g, game.Player1, damageInstantDef("Shock"))
	imprintLink(g, obj, cardID)

	engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	copySpell, ok := g.Stack.Peek()
	if !ok || !copySpell.Copy || copySpell.SourceTokenDef == nil || copySpell.SourceTokenDef.Name != "Shock" ||
		!slices.Equal(copySpell.Targets, []game.Target{game.PlayerTarget(game.Player2)}) {
		t.Fatalf("stack top = %+v, want cast Shock copy targeting Player2", copySpell)
	}
	if got := countEventsOfKind(g.Events, game.EventSpellCast); got != 1 {
		t.Fatalf("spell cast events = %d, want 1", got)
	}
	if got := countEventsOfKind(g.Events, game.EventCrimeCommitted); got != 1 {
		t.Fatalf("crime events = %d, want 1 for the genuinely cast copy", got)
	}
	crime := g.Events[slices.IndexFunc(g.Events, func(event game.Event) bool { return event.Kind == game.EventCrimeCommitted })]
	if crime.Player != game.Player1 || crime.StackObjectID != copySpell.ID {
		t.Fatalf("crime event = %+v, want Player1's crime by the copy", crime)
	}
	if got := resolveKnifepointTriggers(t, engine, g); got != 1 {
		t.Fatalf("resolved At Knifepoint triggers = %d, want 1", got)
	}
	tokens := mercenaryTokens(g, game.Player1)
	if len(tokens) != 1 {
		t.Fatalf("Mercenary tokens = %d, want 1", len(tokens))
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("copied physical card left exile")
	}
	token := tokens[0]
	token.SummoningSick = false
	if engine.applyAction(g, game.Player1, action.ActivateAbility(token.ObjectID, 0,
		[]game.Target{game.PermanentTarget(token.ObjectID)}, 0)) {
		t.Fatal("sorcery-speed Mercenary ability activated with a nonempty stack")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Life != 38 || g.Stack.Size() != 0 {
		t.Fatalf("copy resolution: opponent life %d stack %d, want 38 and empty", g.Players[game.Player2].Life, g.Stack.Size())
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) || g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("copy resolution moved the physical exiled card")
	}
	assertKnifepointMercenary(t, engine, g, token)
	if got := countEventsOfKind(g.Events, game.EventCrimeCommitted); got != 1 {
		t.Fatalf("crime events after resolution = %d, want 1", got)
	}
}

func assertKnifepointMercenary(t *testing.T, engine *Engine, g *game.Game, token *game.Permanent) {
	t.Helper()
	def := token.TokenDef
	if def.Name != "Mercenary" || !slices.Equal(def.Colors, []color.Color{color.Red}) ||
		!def.HasType(types.Creature) || !permanentHasSubtype(g, token, types.Mercenary) ||
		effectivePower(g, token) != 1 {
		t.Fatalf("token = %+v, want 1/1 red Mercenary creature", def)
	}
	if toughness, _ := effectiveToughness(g, token); toughness != 1 {
		t.Fatalf("token toughness = %d, want 1", toughness)
	}
	setMainPhasePriority(g, game.Player1)
	if !hasKeyword(g, token, game.FirstStrike) {
		t.Fatal("outlaw Mercenary lacks first strike during its controller's turn")
	}
	nonOutlaw := addCombatPermanent(g, game.Player1, vanillaCreature("Bystander", 2, 2))
	opponentOutlaw := addCombatPermanent(g, game.Player2, atKnifepointMercenaryToken())
	if hasKeyword(g, nonOutlaw, game.FirstStrike) || hasKeyword(g, opponentOutlaw, game.FirstStrike) {
		t.Fatal("At Knifepoint granted first strike beyond outlaws its controller controls")
	}
	g.Turn.ActivePlayer = game.Player2
	if hasKeyword(g, token, game.FirstStrike) {
		t.Fatal("outlaw kept first strike outside its controller's turn")
	}
	setMainPhasePriority(g, game.Player1)
	token.SummoningSick = false
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(token.ObjectID, 0,
		[]game.Target{game.PermanentTarget(token.ObjectID)}, 0)) {
		t.Fatal("Mercenary token ability did not activate as a sorcery")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if effectivePower(g, token) != 2 || !token.Tapped {
		t.Fatalf("Mercenary ability: power %d tapped %v, want 2 and tapped", effectivePower(g, token), token.Tapped)
	}
}

func newKnifepointGame() (*game.Game, *Engine, *game.Permanent) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	return g, engine, addCombatPermanent(g, game.Player1, atKnifepointFixture())
}

func crimeCount(g *game.Game) int {
	return countEventsOfKind(g.Events, game.EventCrimeCommitted)
}

func pushCrimeStackSpell(g *game.Game, controller game.PlayerID) *game.StackObject {
	addEffectSpellToStack(g, controller, game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}, nil)
	top, _ := g.Stack.Peek()
	return top
}

func pushCrimeStackAbility(g *game.Game, controller game.PlayerID) *game.StackObject {
	obj := activatedObjFor(addCombatPermanent(g, controller, crimeRodDef()))
	obj.ID = g.IDGen.Next()
	obj.AbilityIndex = 0
	g.Stack.Push(obj)
	return obj
}

// TestCopiedCardCrimeEffectivePermanentController proves a permanent target
// qualifies by its current controller, not its owner or base controller
// (CR 700.13, 108.4), on both copy-card casting routes.
func TestCopiedCardCrimeEffectivePermanentController(t *testing.T) {
	tests := []struct {
		name           string
		base, actual   game.PlayerID
		wantCrimeCount int
	}{
		{"own permanent stolen by opponent", game.Player1, game.Player2, 1},
		{"opponent permanent stolen by caster", game.Player2, game.Player1, 0},
	}
	routes := []struct {
		name string
		cast func(*Engine, *game.Game, []game.Target) bool
	}{
		{"resolution copy", func(engine *Engine, g *game.Game, targets []game.Target) bool {
			cardID := addExiledInstant(g, game.Player1, crimeTargetSpell("Copied Crime", permanentTargetSpec()))
			return castExiledCardCopy(engine, g, game.Player1, cardID, targets)
		}},
		{"prepared copy", func(engine *Engine, g *game.Game, targets []game.Target) bool {
			prepared := addPreparedPermanent(g, game.Player1, preparedCrimeCard(permanentTargetSpec()))
			return castPreparedCopy(engine, g, game.Player1, prepared, targets)
		}},
	}
	for _, route := range routes {
		for _, tc := range tests {
			t.Run(route.name+"/"+tc.name, func(t *testing.T) {
				g, engine, _ := newKnifepointGame()
				victim := addCombatPermanent(g, tc.base, vanillaCreature("Victim", 2, 2))
				g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
					ID: g.IDGen.Next(), AffectedObjectID: victim.ObjectID,
					Layer: game.LayerControl, NewController: opt.Val(tc.actual),
				})
				if effectiveController(g, victim) != tc.actual {
					t.Fatal("control-changing effect did not apply")
				}
				if !route.cast(engine, g, []game.Target{game.PermanentTarget(victim.ObjectID)}) {
					t.Fatal("copy cast failed")
				}
				if got := crimeCount(g); got != tc.wantCrimeCount {
					t.Fatalf("crime events = %d, want %d (base controller %v, current controller %v)",
						got, tc.wantCrimeCount, victim.Controller, effectiveController(g, victim))
				}
				if got := resolveKnifepointTriggers(t, engine, g); got != tc.wantCrimeCount {
					t.Fatalf("At Knifepoint triggers = %d, want %d", got, tc.wantCrimeCount)
				}
			})
		}
	}
}

// TestCopiedCardCrimeResolutionCopyTargetMatrix covers each qualifying target
// class (CR 700.13) for a copy cast through the resolution-time copy-card route.
func TestCopiedCardCrimeResolutionCopyTargetMatrix(t *testing.T) {
	tests := []struct {
		name    string
		spec    game.TargetSpec
		targets func(g *game.Game) []game.Target
		crime   bool
	}{
		{"opponent player", playerTargets(1, 1), func(*game.Game) []game.Target {
			return []game.Target{game.PlayerTarget(game.Player2)}
		}, true},
		{"self player", playerTargets(1, 1), func(*game.Game) []game.Target {
			return []game.Target{game.PlayerTarget(game.Player1)}
		}, false},
		{"multiple qualifying players", playerTargets(2, 2), func(*game.Game) []game.Target {
			return []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)}
		}, true},
		{"self and opponent", playerTargets(2, 2), func(*game.Game) []game.Target {
			return []game.Target{game.PlayerTarget(game.Player1), game.PlayerTarget(game.Player4)}
		}, true},
		{"opponent permanent", permanentTargetSpec(), func(g *game.Game) []game.Target {
			return []game.Target{game.PermanentTarget(addCombatPermanent(g, game.Player2, vanillaCreature("Victim", 2, 2)).ObjectID)}
		}, true},
		{"own permanent", permanentTargetSpec(), func(g *game.Game) []game.Target {
			return []game.Target{game.PermanentTarget(addCombatPermanent(g, game.Player1, vanillaCreature("Ally", 2, 2)).ObjectID)}
		}, false},
		{"opponent spell", stackTargetSpec(), func(g *game.Game) []game.Target {
			addEffectSpellToStack(g, game.Player2, game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}, nil)
			top, _ := g.Stack.Peek()
			return []game.Target{game.StackObjectTarget(top.ID)}
		}, true},
		{"own spell", stackTargetSpec(), func(g *game.Game) []game.Target {
			addEffectSpellToStack(g, game.Player1, game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}, nil)
			top, _ := g.Stack.Peek()
			return []game.Target{game.StackObjectTarget(top.ID)}
		}, false},
		{"opponent ability", stackTargetSpec(), func(g *game.Game) []game.Target {
			obj := activatedObjFor(addCombatPermanent(g, game.Player2, crimeRodDef()))
			obj.ID = g.IDGen.Next()
			obj.AbilityIndex = 0
			g.Stack.Push(obj)
			return []game.Target{game.StackObjectTarget(obj.ID)}
		}, true},
		{"opponent graveyard card", graveyardCardTargetSpec(), func(g *game.Game) []game.Target {
			return []game.Target{game.CardTarget(addCardToGraveyard(g, game.Player3, vanillaCreature("Corpse", 1, 1)))}
		}, true},
		{"own graveyard card", graveyardCardTargetSpec(), func(g *game.Game) []game.Target {
			return []game.Target{game.CardTarget(addCardToGraveyard(g, game.Player1, vanillaCreature("Corpse", 1, 1)))}
		}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, engine, _ := newKnifepointGame()
			cardID := addExiledInstant(g, game.Player1, crimeTargetSpell("Copied Crime", tc.spec))
			targets := tc.targets(g)
			stackBefore := g.Stack.Size()
			if !castExiledCardCopy(engine, g, game.Player1, cardID, targets) {
				t.Fatal("copy-card cast failed")
			}
			copySpell, _ := g.Stack.Peek()
			if g.Stack.Size() != stackBefore+1 || !copySpell.Copy || !sameTargetObjects(copySpell.Targets, targets) {
				t.Fatalf("stack top = %+v, want copy with targets %+v", copySpell, targets)
			}
			if got := countEventsOfKind(g.Events, game.EventSpellCast); got != 1 {
				t.Fatalf("spell cast events = %d, want 1", got)
			}
			want := 0
			if tc.crime {
				want = 1
			}
			if got := crimeCount(g); got != want {
				t.Fatalf("crime events = %d, want %d", got, want)
			}
			if got := resolveKnifepointTriggers(t, engine, g); got != want {
				t.Fatalf("At Knifepoint triggers = %d, want %d", got, want)
			}
			if got := len(mercenaryTokens(g, game.Player1)); got != want {
				t.Fatalf("Mercenary tokens = %d, want %d", got, want)
			}
			if !g.Players[game.Player1].Exile.Contains(cardID) {
				t.Fatal("copied physical card left exile")
			}
		})
	}
}

// TestCopiedCardCrimePreparedCopyCast covers the prepared spell-copy casting
// route, which also casts a copy of a card (CR 707.12).
func TestCopiedCardCrimePreparedCopyCast(t *testing.T) {
	tests := []struct {
		name    string
		spec    game.TargetSpec
		targets func(g *game.Game) []game.Target
		crime   bool
	}{
		{"opponent player", playerTargets(1, 1), func(*game.Game) []game.Target {
			return []game.Target{game.PlayerTarget(game.Player3)}
		}, true},
		{"self player", playerTargets(1, 1), func(*game.Game) []game.Target {
			return []game.Target{game.PlayerTarget(game.Player1)}
		}, false},
		{"multiple qualifying players", playerTargets(2, 2), func(*game.Game) []game.Target {
			return []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player4)}
		}, true},
		{"opponent permanent", permanentTargetSpec(), func(g *game.Game) []game.Target {
			return []game.Target{game.PermanentTarget(addCombatPermanent(g, game.Player4, vanillaCreature("Victim", 2, 2)).ObjectID)}
		}, true},
		{"opponent graveyard card", graveyardCardTargetSpec(), func(g *game.Game) []game.Target {
			return []game.Target{game.CardTarget(addCardToGraveyard(g, game.Player2, vanillaCreature("Corpse", 1, 1)))}
		}, true},
		{"opponent spell", stackTargetSpec(), func(g *game.Game) []game.Target {
			return []game.Target{game.StackObjectTarget(pushCrimeStackSpell(g, game.Player2).ID)}
		}, true},
		{"own spell", stackTargetSpec(), func(g *game.Game) []game.Target {
			return []game.Target{game.StackObjectTarget(pushCrimeStackSpell(g, game.Player1).ID)}
		}, false},
		{"opponent ability", stackTargetSpec(), func(g *game.Game) []game.Target {
			return []game.Target{game.StackObjectTarget(pushCrimeStackAbility(g, game.Player3).ID)}
		}, true},
		{"own ability", stackTargetSpec(), func(g *game.Game) []game.Target {
			return []game.Target{game.StackObjectTarget(pushCrimeStackAbility(g, game.Player1).ID)}
		}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, engine, _ := newKnifepointGame()
			prepared := addPreparedPermanent(g, game.Player1, preparedCrimeCard(tc.spec))
			targets := tc.targets(g)
			stackBefore := g.Stack.Size()
			if !castPreparedCopy(engine, g, game.Player1, prepared, targets) {
				t.Fatal("prepared copy cast failed")
			}
			copySpell, _ := g.Stack.Peek()
			if !copySpell.Copy || copySpell.Face != game.FaceAlternate || !sameTargetObjects(copySpell.Targets, targets) {
				t.Fatalf("stack top = %+v, want prepared copy with targets %+v", copySpell, targets)
			}
			if prepared.Prepared || permanentForCard(g, prepared.CardInstanceID) != prepared {
				t.Fatal("prepared source did not stay on the battlefield unprepared")
			}
			if got := countEventsOfKind(g.Events, game.EventSpellCast); got != 1 {
				t.Fatalf("spell cast events = %d, want 1", got)
			}
			want := 0
			if tc.crime {
				want = 1
			}
			if got := crimeCount(g); got != want {
				t.Fatalf("crime events = %d, want %d", got, want)
			}
			if got := resolveKnifepointTriggers(t, engine, g); got != want {
				t.Fatalf("At Knifepoint triggers = %d, want %d", got, want)
			}
			life := g.Players[game.Player1].Life
			engine.resolveTopOfStack(g, &TurnLog{})
			if g.Players[game.Player1].Life != life+1 || g.Stack.Size() != stackBefore {
				t.Fatal("prepared copy did not resolve and cease to exist")
			}
			if permanentForCard(g, prepared.CardInstanceID) != prepared {
				t.Fatal("prepared source left the battlefield when its copy resolved")
			}
		})
	}

	t.Run("unpayable", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		prepared := addPreparedPermanent(g, game.Player1, preparedCrimeCard(playerTargets(1, 1)))
		for _, permanent := range g.Battlefield {
			if permanent.Controller == game.Player1 && permanent.CardInstanceID != prepared.CardInstanceID {
				permanent.Tapped = true
			}
		}
		if castPreparedCopy(engine, g, game.Player1, prepared, []game.Target{game.PlayerTarget(game.Player2)}) {
			t.Fatal("unpayable prepared copy was cast")
		}
		if crimeCount(g) != 0 || countEventsOfKind(g.Events, game.EventSpellCast) != 0 || g.Stack.Size() != 0 ||
			!prepared.Prepared || len(g.TriggeredAbilitiesThisTurn) != 0 {
			t.Fatal("failed prepared payment published a cast or crime or spent the preparation")
		}
	})
}

// crimeRodDef is an artifact with a tap ability targeting a player.
func crimeRodDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Crime Rod",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:            "{T}: Target player gains 1 life.",
			AdditionalCosts: cost.Tap,
			ZoneOfFunction:  zone.Battlefield,
			Content: game.Mode{
				Targets: []game.TargetSpec{playerTargets(1, 1)},
				Sequence: []game.Instruction{{Primitive: game.GainLife{
					Player: game.ControllerReference(), Amount: game.Fixed(1),
				}}},
			}.Ability(),
		}},
	}}
}

func castCrimeSpellFromHand(t *testing.T, engine *Engine, g *game.Game, def *game.CardDef, targets []game.Target, agents [game.NumPlayers]PlayerAgent) *game.StackObject {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, targets, 0, nil), agents, &TurnLog{}) {
		t.Fatal("hand cast failed")
	}
	for _, obj := range g.Stack.Objects() {
		if obj.SourceID == cardID && !obj.Copy {
			return obj
		}
	}
	t.Fatal("cast spell not on stack")
	return nil
}

func copyStackObject(engine *Engine, g *game.Game, original *game.StackObject, newTargets ...game.Target) *game.StackObject {
	addEffectSpellToStack(g, game.Player1,
		game.CopyStackObject{Object: game.TargetStackObjectReference(0), MayChooseNewTargets: len(newTargets) > 0},
		[]game.Target{game.StackObjectTarget(original.ID)})
	var agents [game.NumPlayers]PlayerAgent
	if len(newTargets) > 0 {
		agents[game.Player1] = retargetTo(newTargets...)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	top, _ := g.Stack.Peek()
	return top
}

// TestCopiedCardCrimeOriginalAnnouncementsAndCopies preserves crimes for
// original spells, activated abilities and triggered abilities while stack
// copies, storm copies and copied abilities are never announced (CR 707.10).
func TestCopiedCardCrimeOriginalAnnouncementsAndCopies(t *testing.T) {
	opponent := []game.Target{game.PlayerTarget(game.Player2)}
	self := []game.Target{game.PlayerTarget(game.Player1)}

	t.Run("hand cast", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		castCrimeSpellFromHand(t, engine, g, crimeTargetSpell("Hand Crime", playerTargets(1, 1)), opponent, [game.NumPlayers]PlayerAgent{})
		if crimeCount(g) != 1 || resolveKnifepointTriggers(t, engine, g) != 1 {
			t.Fatal("ordinary cast did not commit one crime")
		}
	})

	t.Run("activated ability", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		rod := addCombatPermanent(g, game.Player1, crimeRodDef())
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(rod.ObjectID, 0, opponent, 0)) {
			t.Fatal("activation failed")
		}
		if crimeCount(g) != 1 || resolveKnifepointTriggers(t, engine, g) != 1 {
			t.Fatal("activated ability did not commit one crime")
		}
	})

	t.Run("triggered ability", func(t *testing.T) {
		g, engine, knifepoint := newKnifepointGame()
		addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn, Player: game.TriggerPlayerYou},
			[]game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}}},
			[]game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "target opponent", Allow: game.TargetAllowPlayer}})
		emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("targeted trigger was not put on the stack")
		}
		top, _ := g.Stack.Peek()
		if len(top.Targets) != 1 || top.Targets[0].Kind != game.TargetPlayer || top.Targets[0].PlayerID == game.Player1 || crimeCount(g) != 1 {
			t.Fatalf("trigger %+v crimes %d, want opponent target and one crime", top.Targets, crimeCount(g))
		}
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("trigger crime did not trigger At Knifepoint")
		}
		knifeTrigger, _ := g.Stack.Peek()
		if knifeTrigger.SourceID != knifepoint.ObjectID {
			t.Fatalf("top trigger source = %v, want At Knifepoint", knifeTrigger.SourceID)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if len(mercenaryTokens(g, game.Player1)) != 1 || crimeCount(g) != 1 {
			t.Fatal("trigger crime did not create one Mercenary")
		}
	})

	t.Run("stack copies with changed targets", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		innocent := castCrimeSpellFromHand(t, engine, g, crimeTargetSpell("Innocent", playerTargets(1, 1)), self, [game.NumPlayers]PlayerAgent{})
		copied := copyStackObject(engine, g, innocent, opponent...)
		if !copied.Copy || !slices.Equal(copied.Targets, opponent) || crimeCount(g) != 0 {
			t.Fatalf("copy %+v crimes %d, want opponent-targeting copy without a crime", copied, crimeCount(g))
		}
		guilty := castCrimeSpellFromHand(t, engine, g, crimeTargetSpell("Guilty", playerTargets(1, 1)), opponent, [game.NumPlayers]PlayerAgent{})
		copyStackObject(engine, g, guilty)
		copyStackObject(engine, g, guilty, game.PlayerTarget(game.Player3))
		if crimeCount(g) != 1 || countEventsOfKind(g.Events, game.EventSpellCast) != 2 {
			t.Fatalf("crimes %d casts %d, want only the original announcements", crimeCount(g), countEventsOfKind(g.Events, game.EventSpellCast))
		}
	})

	t.Run("storm copies", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		firstID := addCardToHand(g, game.Player1, simpleGainLifeInstant("First Spell"))
		if !engine.applyAction(g, game.Player1, action.CastSpell(firstID, nil, 0, nil)) {
			t.Fatal("first spell cast failed")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		storm := crimeTargetSpell("Storm Crime", playerTargets(1, 1))
		storm.StaticAbilities = []game.StaticAbility{game.StormStaticBody}
		castCrimeSpellFromHand(t, engine, g, storm, opponent, [game.NumPlayers]PlayerAgent{})
		copies := 0
		for _, obj := range g.Stack.Objects() {
			if obj.Copy {
				copies++
				if !slices.Equal(obj.Targets, opponent) {
					t.Fatalf("storm copy targets = %+v, want the original's targets", obj.Targets)
				}
			}
		}
		if copies != 1 || crimeCount(g) != 1 {
			t.Fatalf("storm copies %d crimes %d, want one copy and one crime", copies, crimeCount(g))
		}
	})

	t.Run("copied ability", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		rod := addCombatPermanent(g, game.Player1, crimeRodDef())
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(rod.ObjectID, 0, self, 0)) {
			t.Fatal("activation failed")
		}
		ability, _ := g.Stack.Peek()
		copied := copyStackObject(engine, g, ability, opponent...)
		if !copied.Copy || copied.Kind != game.StackActivatedAbility || !slices.Equal(copied.Targets, opponent) || crimeCount(g) != 0 {
			t.Fatalf("copied ability %+v crimes %d, want retargeted copy without a crime", copied, crimeCount(g))
		}
	})

	t.Run("card copy cast triggers cast and crime abilities together", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{Event: game.EventSpellCast, Controller: game.TriggerControllerYou},
			[]game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(5)}}}, nil)
		cardID := addExiledInstant(g, game.Player1, crimeTargetSpell("Copied Crime", playerTargets(1, 1)))
		if !castExiledCardCopy(engine, g, game.Player1, cardID, opponent) {
			t.Fatal("copy-card cast failed")
		}
		if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 3 {
			t.Fatalf("stack size = %d, want copy plus simultaneous cast and crime triggers", g.Stack.Size())
		}
		life := g.Players[game.Player1].Life
		engine.resolveTopOfStack(g, &TurnLog{})
		engine.resolveTopOfStack(g, &TurnLog{})
		if g.Players[game.Player1].Life != life+5 || len(mercenaryTokens(g, game.Player1)) != 1 {
			t.Fatal("cast trigger or crime trigger did not resolve for the cast copy")
		}
	})
}

// TestCopiedCardCrimePaymentAndLaterChanges proves the crime belongs to a
// completed announcement (CR 601.2i): failed or declined casts commit none and
// later countering, illegal targets or retargeting do not change it.
func TestCopiedCardCrimePaymentAndLaterChanges(t *testing.T) {
	opponent := []game.Target{game.PlayerTarget(game.Player2)}

	t.Run("additional cost", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		def := crimeTargetSpell("Costly Crime", playerTargets(1, 1))
		def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalSacrifice, MatchPermanentType: true, PermanentType: types.Creature}}
		cardID := addExiledInstant(g, game.Player1, def)
		if castExiledCardCopy(engine, g, game.Player1, cardID, opponent) {
			t.Fatal("copy was cast without its sacrifice cost")
		}
		if crimeCount(g) != 0 || countEventsOfKind(g.Events, game.EventSpellCast) != 0 || g.Stack.Size() != 0 ||
			engine.putTriggeredAbilitiesOnStack(g) || len(g.TriggeredAbilitiesThisTurn) != 0 {
			t.Fatal("failed payment published a cast or crime or consumed a trigger use")
		}
		fodder := addCombatPermanent(g, game.Player1, vanillaCreature("Fodder", 1, 1))
		if !castExiledCardCopy(engine, g, game.Player1, cardID, opponent) {
			t.Fatal("payable copy cast failed")
		}
		copySpell, _ := g.Stack.Peek()
		if permanentForCard(g, fodder.CardInstanceID) != nil || len(copySpell.AdditionalCostsPaid) == 0 {
			t.Fatal("combined additional cost was not paid")
		}
		if crimeCount(g) != 1 || resolveKnifepointTriggers(t, engine, g) != 1 || !g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatal("paid copy cast did not commit exactly one crime")
		}
	})

	t.Run("combined costs", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := buybackEngine(t)
		subject := castBuybackSubject(t, g, engine)
		addCombatPermanent(g, game.Player1, atKnifepointFixture())
		def := reiterateFixture()
		def.AdditionalCosts = []cost.Additional{
			{Kind: cost.AdditionalPayLife, Amount: 2, ChoiceGroup: 1},
			{Kind: cost.AdditionalSacrifice, Amount: 1, ChoiceGroup: 1, MatchPermanentType: true, PermanentType: types.Artifact},
		}
		def.StaticAbilities[0].KeywordAbilities = []game.KeywordAbility{game.BuybackKeyword{
			Cost: cost.Mana{cost.O(3)},
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalPayLife, Amount: 3, ChoiceGroup: 1},
				{Kind: cost.AdditionalDiscard, Amount: 1, ChoiceGroup: 1},
			},
		}}
		cardID := addExiledInstant(g, game.Player1, def)
		g.Players[game.Player1].ManaPool.Add(mana.C, 4)
		g.CostModifiers = []game.CostModifier{{Kind: game.CostModifierSpell, GenericIncrease: 1}}
		life := g.Players[game.Player1].Life
		card, _ := g.GetCardInstance(cardID)
		if !engine.castFreeOptionWithCopy(g, nil, game.Player1, card, action.CastSpellAction{
			CardID: cardID, SourceZone: zone.Exile, Targets: []game.Target{game.StackObjectTarget(subject.ID)},
		}, false, true, independentBuybackAgent(t, independentChooseBuyback), &TurnLog{}) {
			t.Fatal("combined-cost copy cast failed")
		}
		copySpell, _ := g.Stack.Peek()
		if !copySpell.Copy || !copySpell.BuybackPaid || copySpell.ManaSpentToCast != 4 ||
			g.Players[game.Player1].Life != life-5 || len(copySpell.AdditionalCostsPaid) != 2 {
			t.Fatalf("payment: copy %v buyback %v spent %d life %d costs %v", copySpell.Copy, copySpell.BuybackPaid,
				copySpell.ManaSpentToCast, g.Players[game.Player1].Life, copySpell.AdditionalCostsPaid)
		}
		if crimeCount(g) != 1 || countEventsOfKind(g.Events, game.EventSpellCast) != 2 || resolveKnifepointTriggers(t, engine, g) != 1 {
			t.Fatal("combined-cost copy cast did not commit exactly one crime")
		}
		if !g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatal("copied physical card left exile")
		}
	})

	t.Run("declined", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		scepter := addCombatPermanent(g, game.Player1, isochronScepterDef())
		obj := activatedObjFor(scepter)
		cardID := addExiledInstant(g, game.Player1, damageInstantDef("Shock"))
		imprintLink(g, obj, cardID)
		engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(),
			[game.NumPlayers]PlayerAgent{game.Player1: declineChoiceAgent{}}, &TurnLog{})
		if g.Stack.Size() != 0 || crimeCount(g) != 0 || countEventsOfKind(g.Events, game.EventSpellCast) != 0 ||
			engine.putTriggeredAbilitiesOnStack(g) || len(g.TriggeredAbilitiesThisTurn) != 0 {
			t.Fatal("declined copy cast committed a crime or consumed a trigger use")
		}
	})

	t.Run("countered", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		cardID := addExiledInstant(g, game.Player1, crimeTargetSpell("Countered Crime", playerTargets(1, 1)))
		castExiledCardCopy(engine, g, game.Player1, cardID, opponent)
		copySpell, _ := g.Stack.Peek()
		if !counterStackObject(g, copySpell.ID) {
			t.Fatal("copy was not countered")
		}
		if crimeCount(g) != 1 || resolveKnifepointTriggers(t, engine, g) != 1 || len(mercenaryTokens(g, game.Player1)) != 1 {
			t.Fatal("countering the copy undid its crime")
		}
		if !g.Players[game.Player1].Exile.Contains(cardID) || g.Players[game.Player1].Graveyard.Contains(cardID) || g.Stack.Size() != 0 {
			t.Fatal("countered copy moved its physical source or remained on the stack")
		}
	})

	t.Run("targets become illegal", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		victim := addCombatPermanent(g, game.Player2, vanillaCreature("Victim", 2, 2))
		cardID := addExiledInstant(g, game.Player1, crimeTargetSpell("Fizzled Crime", permanentTargetSpec()))
		castExiledCardCopy(engine, g, game.Player1, cardID, []game.Target{game.PermanentTarget(victim.ObjectID)})
		movePermanentToZone(g, victim, zone.Graveyard)
		if resolveKnifepointTriggers(t, engine, g) != 1 {
			t.Fatal("crime trigger did not resolve after its target left")
		}
		life := g.Players[game.Player1].Life
		engine.resolveTopOfStack(g, &TurnLog{})
		if crimeCount(g) != 1 || g.Players[game.Player1].Life != life || g.Stack.Size() != 0 ||
			!g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatal("illegal-target copy changed its crime, resolved, or moved its physical source")
		}
	})

	t.Run("retargeting", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		innocentID := addExiledInstant(g, game.Player1, crimeTargetSpell("Innocent Copy", playerTargets(1, 1)))
		castExiledCardCopy(engine, g, game.Player1, innocentID, []game.Target{game.PlayerTarget(game.Player1)})
		innocent, _ := g.Stack.Peek()
		agents := [game.NumPlayers]PlayerAgent{game.Player1: retargetTo(opponent...)}
		if !engine.retargetStackObjectChoice(g, game.Player1, innocent, agents, &TurnLog{}) ||
			!slices.Equal(innocent.Targets, opponent) || crimeCount(g) != 0 {
			t.Fatalf("retargeted copy %+v crimes %d, want retarget without crime", innocent.Targets, crimeCount(g))
		}
		guiltyID := addExiledInstant(g, game.Player1, crimeTargetSpell("Guilty Copy", playerTargets(1, 1)))
		castExiledCardCopy(engine, g, game.Player1, guiltyID, opponent)
		guilty, _ := g.Stack.Peek()
		agents[game.Player1] = retargetTo(game.PlayerTarget(game.Player1))
		if !engine.retargetStackObjectChoice(g, game.Player1, guilty, agents, &TurnLog{}) ||
			!slices.Equal(guilty.Targets, []game.Target{game.PlayerTarget(game.Player1)}) || crimeCount(g) != 1 {
			t.Fatalf("retargeted guilty copy %+v crimes %d, want its one crime retained", guilty.Targets, crimeCount(g))
		}
		if resolveKnifepointTriggers(t, engine, g) != 1 {
			t.Fatal("retained crime did not trigger At Knifepoint")
		}
	})
}

// TestCopiedCardCrimeKnifepointTriggerLimits covers At Knifepoint's "you" and
// "only once each turn" trigger semantics for genuinely cast copies.
func TestCopiedCardCrimeKnifepointTriggerLimits(t *testing.T) {
	opponent := []game.Target{game.PlayerTarget(game.Player2)}
	castCopy := func(t *testing.T, engine *Engine, g *game.Game, caster game.PlayerID, targets []game.Target) {
		t.Helper()
		cardID := addExiledInstant(g, caster, crimeTargetSpell("Copied Crime", playerTargets(1, 1)))
		if !castExiledCardCopy(engine, g, caster, cardID, targets) {
			t.Fatal("copy-card cast failed")
		}
	}

	t.Run("only controller crimes", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		castCopy(t, engine, g, game.Player2, []game.Target{game.PlayerTarget(game.Player1)})
		if crimeCount(g) != 1 || engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("an opponent's crime triggered At Knifepoint")
		}
	})

	t.Run("once each turn and reset", func(t *testing.T) {
		g, engine, knifepoint := newKnifepointGame()
		castCopy(t, engine, g, game.Player1, opponent)
		castCopy(t, engine, g, game.Player1, []game.Target{game.PlayerTarget(game.Player3)})
		if crimeCount(g) != 2 || resolveKnifepointTriggers(t, engine, g) != 1 {
			t.Fatal("two same-turn crimes did not trigger exactly once")
		}
		castCopy(t, engine, g, game.Player1, opponent)
		if engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("At Knifepoint triggered twice in one turn")
		}
		if got := g.TriggeredAbilitiesThisTurn[game.TriggeredAbilityUse{SourceID: knifepoint.ObjectID}]; got != 1 {
			t.Fatalf("trigger uses = %d, want 1", got)
		}
		engine.advanceToNextTurn(g)
		castCopy(t, engine, g, game.Player1, opponent)
		if resolveKnifepointTriggers(t, engine, g) != 1 || len(mercenaryTokens(g, game.Player1)) != 2 {
			t.Fatal("At Knifepoint did not reset on the next turn")
		}
	})

	t.Run("separate instances", func(t *testing.T) {
		g, engine, _ := newKnifepointGame()
		addCombatPermanent(g, game.Player1, atKnifepointFixture())
		castCopy(t, engine, g, game.Player1, opponent)
		if resolveKnifepointTriggers(t, engine, g) != 2 || len(mercenaryTokens(g, game.Player1)) != 2 {
			t.Fatal("separate At Knifepoints did not each trigger once")
		}
	})

	t.Run("fresh incarnation", func(t *testing.T) {
		g, engine, knifepoint := newKnifepointGame()
		castCopy(t, engine, g, game.Player1, opponent)
		if resolveKnifepointTriggers(t, engine, g) != 1 {
			t.Fatal("first crime did not trigger")
		}
		movePermanentToZone(g, knifepoint, zone.Graveyard)
		card, _ := g.GetCardInstance(knifepoint.CardInstanceID)
		g.Players[game.Player1].Graveyard.Remove(card.ID)
		returned, ok := createCardPermanent(g, card, game.Player1, zone.Graveyard)
		if !ok || returned.ObjectID == knifepoint.ObjectID {
			t.Fatal("At Knifepoint did not return as a new object")
		}
		castCopy(t, engine, g, game.Player1, opponent)
		if resolveKnifepointTriggers(t, engine, g) != 1 || len(mercenaryTokens(g, game.Player1)) != 2 {
			t.Fatal("fresh At Knifepoint incarnation did not trigger")
		}
	})

	t.Run("copied trigger and source departure", func(t *testing.T) {
		g, engine, knifepoint := newKnifepointGame()
		castCopy(t, engine, g, game.Player1, opponent)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("crime did not trigger")
		}
		trigger, _ := g.Stack.Peek()
		copied := copyStackObject(engine, g, trigger)
		if !copied.Copy || copied.Kind != game.StackTriggeredAbility {
			t.Fatalf("copied trigger = %+v", copied)
		}
		movePermanentToZone(g, knifepoint, zone.Graveyard)
		engine.resolveTopOfStack(g, &TurnLog{})
		engine.resolveTopOfStack(g, &TurnLog{})
		if len(mercenaryTokens(g, game.Player1)) != 2 {
			t.Fatal("original and copied triggers did not each create a Mercenary after source departure")
		}
		if got := g.TriggeredAbilitiesThisTurn[game.TriggeredAbilityUse{SourceID: knifepoint.ObjectID}]; got != 1 {
			t.Fatalf("trigger uses = %d, want copied trigger not to consume a use", got)
		}
	})
}

// TestCopiedCardCrimeCloneAndReplay proves crime events, targets and trigger
// uses are game-owned across clones and recorded-choice replay.
func TestCopiedCardCrimeCloneAndReplay(t *testing.T) {
	g, engine, knifepoint := newKnifepointGame()
	scepter := addCombatPermanent(g, game.Player1, isochronScepterDef())
	obj := activatedObjFor(scepter)
	cardID := addExiledInstant(g, game.Player1, crimeTargetSpell("Replayed Crime",
		game.TargetSpec{MinTargets: 1, MaxTargets: 1, Constraint: "target opponent", Allow: game.TargetAllowPlayer}))
	imprintLink(g, obj, cardID)
	replay := g.Clone()

	var recorded TurnLog
	recorder := crimeChoiceAgent(func(request game.ChoiceRequest) ([]int, bool) {
		if request.Kind == game.ChoiceMay {
			return []int{1}, true
		}
		return nil, false
	})
	engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(), [game.NumPlayers]PlayerAgent{game.Player1: recorder}, &recorded)
	castEvents := slices.Clone(g.Events)
	copySpell, _ := g.Stack.Peek()
	if !copySpell.Copy || crimeCount(g) != 1 || len(recorded.Choices) == 0 {
		t.Fatalf("recorded cast copy %+v crimes %d choices %d", copySpell, crimeCount(g), len(recorded.Choices))
	}
	if crimeCount(replay) != 0 || replay.Stack.Size() != 0 {
		t.Fatal("casting in the original mutated the pre-cast clone")
	}

	mid := g.Clone()
	if resolveKnifepointTriggers(t, engine, g) != 1 {
		t.Fatal("original did not trigger")
	}
	if len(mid.TriggeredAbilitiesThisTurn) != 0 || mid.Stack.Size() != 1 || len(mercenaryTokens(mid, game.Player1)) != 0 {
		t.Fatal("resolving the original trigger mutated the mid-cast clone")
	}
	if resolveKnifepointTriggers(t, engine, mid) != 1 || len(mercenaryTokens(mid, game.Player1)) != 1 {
		t.Fatal("mid-cast clone did not own its pending crime trigger")
	}
	use := game.TriggeredAbilityUse{SourceID: knifepoint.ObjectID}
	if g.TriggeredAbilitiesThisTurn[use] != 1 || mid.TriggeredAbilitiesThisTurn[use] != 1 || len(mercenaryTokens(g, game.Player1)) != 1 {
		t.Fatal("trigger use or token ownership crossed clones")
	}

	next := 0
	replayer := crimeChoiceAgent(func(game.ChoiceRequest) ([]int, bool) {
		selected := recorded.Choices[next].Selected
		next++
		return selected, true
	})
	replayObj := activatedObjFor(scepter)
	engine.resolveAbilityContentWithChoices(replay, replayObj, imprintCopyCastAbility(), [game.NumPlayers]PlayerAgent{game.Player1: replayer}, &TurnLog{})
	replayCopy, _ := replay.Stack.Peek()
	if !slices.EqualFunc(replay.Events, castEvents, func(a, b game.Event) bool {
		return a.Kind == b.Kind && a.Player == b.Player && a.StackObjectID == b.StackObjectID && a.Target == b.Target
	}) {
		t.Fatal("replayed cast events differ from the recorded cast")
	}
	if next != len(recorded.Choices) || !slices.Equal(replayCopy.Targets, copySpell.Targets) ||
		crimeCount(replay) != 1 || countEventsOfKind(replay.Events, game.EventSpellCast) != 1 {
		t.Fatalf("replay targets %+v crimes %d, want recorded %+v and one crime", replayCopy.Targets, crimeCount(replay), copySpell.Targets)
	}
	if resolveKnifepointTriggers(t, engine, replay) != 1 || replay.TriggeredAbilitiesThisTurn[use] != 1 {
		t.Fatal("replay did not reproduce the trigger use")
	}
	if !replay.Players[game.Player1].Exile.Contains(cardID) || !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("copied source card left exile")
	}
}
