package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestSpellResolvesForRemainingLegalTargetWithoutShiftingSlots(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	shrouded := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	legal := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	sourceID := addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0)}},
		{Primitive: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(1)}},
	}, []game.Target{
		game.PermanentTarget(shrouded.ObjectID),
		game.PermanentTarget(legal.ObjectID),
	})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("spell was not put on stack")
	}
	obj.TargetCounts = []int{1, 1}
	addShroudGranter(g, game.Player2)
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "graveyard" {
		t.Fatalf("resolve log = %+v, want spell resolved to graveyard", log.Resolves)
	}
	if shrouded.MarkedDamage != 0 {
		t.Fatalf("shrouded target marked damage = %d, want 0", shrouded.MarkedDamage)
	}
	if legal.MarkedDamage != 3 {
		t.Fatalf("legal target marked damage = %d, want 3", legal.MarkedDamage)
	}
}

func TestSpellIsCounteredWhenAllTargetsGainShroud(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	sourceID := addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0)}},
		{Primitive: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(1)}},
	}, []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("spell was not put on stack")
	}
	obj.TargetCounts = []int{1, 1}
	addShroudGranter(g, game.Player2)
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "countered by rules" {
		t.Fatalf("resolve log = %+v, want countered by rules", log.Resolves)
	}
	if first.MarkedDamage != 0 || second.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, %d; want no damage", first.MarkedDamage, second.MarkedDamage)
	}
}

func TestPermanentTargetedDamageMarksDamageOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	sourceID := addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:    game.Fixed(3),
		Recipient: game.AnyTargetDamageRecipient(0),
	}, []game.Target{game.PermanentTarget(target.ObjectID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	// Set target spec on the spell's content to require a creature target
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}}

	engine.resolveTopOfStack(g, &TurnLog{})

	if target.MarkedDamage != 3 {
		t.Fatalf("marked damage = %d, want 3", target.MarkedDamage)
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("resolved spell did not move to graveyard")
	}
}

func TestTargetChoiceKindsAtActionEnumerationLevel(t *testing.T) {
	tests := []struct {
		name            string
		setupSpell      func() *game.CardDef
		setupBoard      func(g *game.Game)
		wantCastActions int // number of cast-spell actions for the spell
	}{
		{
			name: "no targets required produces one cast action",
			setupSpell: func() *game.CardDef {
				return &game.CardDef{CardFace: game.CardFace{Name: "Shock No Target",
					Types:        []types.Card{types.Sorcery},
					SpellAbility: opt.Val(game.AbilityContent{})},
				}
			},
			setupBoard:      func(g *game.Game) {},
			wantCastActions: 1,
		},
		{
			name: "required target with one legal candidate produces one cast action",
			setupSpell: func() *game.CardDef {
				return permanentTargetSpell("creature")
			},
			setupBoard: func(g *game.Game) {
				addCreaturePermanent(g, game.Player2)
				addBasicLandPermanent(g, game.Player1, types.Forest)
			},
			wantCastActions: 1,
		},
		{
			name: "required target with no legal candidates produces no cast actions",
			setupSpell: func() *game.CardDef {
				return permanentTargetSpell("planeswalker")
			},
			setupBoard:      func(g *game.Game) {},
			wantCastActions: 0,
		},
		{
			name: "invalid target spec (min > max) produces no cast actions",
			setupSpell: func() *game.CardDef {
				return permanentTargetSpellWithRange("creature", 3, 1)
			},
			setupBoard: func(g *game.Game) {
				addCreaturePermanent(g, game.Player2)
			},
			wantCastActions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			spellID := addCardToHand(g, game.Player1, tt.setupSpell())
			tt.setupBoard(g)
			g.Turn.Phase = game.PhasePrecombatMain
			g.Turn.Step = game.StepNone

			legal := engine.legalActions(g, game.Player1)

			var castCount int
			for _, act := range legal {
				if cast, ok := act.CastSpellPayload(); ok && cast.CardID == spellID {
					castCount++
				}
			}
			if castCount != tt.wantCastActions {
				t.Errorf("cast actions = %d, want %d", castCount, tt.wantCastActions)
			}
		})
	}
}

func TestInvalidTargetSpecAbilityProducesNoActivateActions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Broken Ability Source",
		Types: []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{
				Targets:  []game.TargetSpec{{MinTargets: 3, MaxTargets: 1, Constraint: "creature"}},
				Sequence: []game.Instruction{{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}}},
			}.Ability(),
		}}},
	})
	addCreaturePermanent(g, game.Player2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	for _, act := range legal {
		if activate, ok := act.ActivateAbilityPayload(); ok && activate.SourceID == source.ObjectID {
			t.Fatalf("invalid ability target spec produced activate action: %+v", act)
		}
	}
}

func TestOpponentChosenTargetSlotUsesDeferredLegalAction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, opponentChosenTargetAbilitySource())
	own := addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	addCreaturePermanent(g, game.Player2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	var matching []action.ActivateAbilityAction
	for _, act := range legal {
		activate, ok := act.ActivateAbilityPayload()
		if ok && activate.SourceID == source.ObjectID {
			matching = append(matching, activate)
		}
	}
	if len(matching) != 1 {
		t.Fatalf("activate actions = %d, want 1 canonical action", len(matching))
	}
	if got := matching[0].Targets; len(got) != 2 || got[0] != game.PermanentTarget(own.ObjectID) || got[1].Kind != game.TargetDeferred {
		t.Fatalf("targets = %+v, want own creature plus deferred opponent-chosen slot", got)
	}
}

func TestOpponentChosenTargetSlotIsChosenDuringAnnouncement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, opponentChosenTargetAbilitySource())
	own := addCreaturePermanent(g, game.Player1)
	placeholder := addCreaturePermanent(g, game.Player2)
	_ = addCreaturePermanent(g, game.Player3)
	chosen := addCreaturePermanent(g, game.Player3)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}

	ok := engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, []game.Target{
		game.PermanentTarget(own.ObjectID),
		game.PermanentTarget(placeholder.ObjectID),
	}, 0), agents, &log)

	if !ok {
		t.Fatal("applyActionWithChoices() = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("activated ability was not put on the stack")
	}
	if got := obj.Targets; len(got) != 2 || got[0] != game.PermanentTarget(own.ObjectID) || got[1] != game.PermanentTarget(chosen.ObjectID) {
		t.Fatalf("stack targets = %+v, want opponent's chosen target %d", got, chosen.ObjectID)
	}
	if len(log.Choices) != 2 || log.Choices[0].Request.Player != game.Player1 || log.Choices[1].Request.Player != game.Player3 {
		t.Fatalf("choice log = %+v, want controller opponent choice then opponent target choice", log.Choices)
	}
}

func TestOpponentChosenTargetSlotFallsBackDeterministically(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, opponentChosenTargetAbilitySource())
	own := addCreaturePermanent(g, game.Player1)
	fallback := addCreaturePermanent(g, game.Player2)
	addCreaturePermanent(g, game.Player3)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	ok := engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, []game.Target{
		game.PermanentTarget(own.ObjectID),
		game.DeferredTarget(),
	}, 0), [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !ok {
		t.Fatal("applyActionWithChoices() = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("activated ability was not put on the stack")
	}
	if got := obj.Targets[1]; got != game.PermanentTarget(fallback.ObjectID) {
		t.Fatalf("opponent-chosen target = %+v, want first Player2 creature %d", got, fallback.ObjectID)
	}
}

func TestOpponentChosenTargetSlotKeepsSourceControllerProtection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := opponentChosenTargetAbilitySource()
	content := source.ActivatedAbilities[0].Content
	if content.IsModal() || len(content.Modes) != 1 || len(content.Modes[0].Targets) < 2 {
		t.Fatal("source card missing expected ability targets")
	}
	spec := content.Modes[0].Targets[1]
	hexproof := addHexproofPermanent(g, game.Player2)
	normal := addCreaturePermanent(g, game.Player2)

	candidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player2, source, 0, &spec)

	if slices.Contains(candidates, game.PermanentTarget(hexproof.ObjectID)) {
		t.Fatal("opponent chooser could choose hexproof creature against source controller")
	}
	if !slices.Contains(candidates, game.PermanentTarget(normal.ObjectID)) {
		t.Fatal("opponent chooser could not choose non-hexproof creature they control")
	}
}

func TestStackSpellTargetCandidatesRespectTypeFilters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addStackSpell(g, game.Player2, "Creature Spell", []types.Card{types.Creature})
	sorcery := addStackSpell(g, game.Player2, "Sorcery Spell", []types.Card{types.Sorcery})
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Constraint: "noncreature spell",
		Predicate: game.TargetPredicate{
			ExcludedSpellCardTypes: []types.Card{types.Creature},
			StackObjectKinds:       []game.StackObjectKind{game.StackSpell},
		},
	}

	source := counterTargetSpell(&spec)

	candidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player1, source, 0, &spec)

	if !slices.Contains(candidates, game.StackObjectTarget(sorcery.ID)) {
		t.Fatalf("candidates = %+v, want noncreature stack target %d", candidates, sorcery.ID)
	}
	if slices.Contains(candidates, game.StackObjectTarget(creature.ID)) {
		t.Fatal("candidates included excluded creature spell target")
	}
}

func TestStackSpellTargetCandidatesRespectTypeUnion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	enchantment := addStackSpell(g, game.Player2, "Enchantment Spell", []types.Card{types.Enchantment})
	instant := addStackSpell(g, game.Player2, "Instant Spell", []types.Card{types.Instant})
	sorcery := addStackSpell(g, game.Player2, "Sorcery Spell", []types.Card{types.Sorcery})
	creature := addStackSpell(g, game.Player2, "Creature Spell", []types.Card{types.Creature})
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{
			SpellCardTypesAny: []types.Card{types.Enchantment, types.Instant, types.Sorcery},
			StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
		},
	}
	source := counterTargetSpell(&spec)

	candidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player1, source, 0, &spec)

	for _, want := range []*game.StackObject{enchantment, instant, sorcery} {
		if !slices.Contains(candidates, game.StackObjectTarget(want.ID)) {
			t.Errorf("candidates = %+v, want stack target %d", candidates, want.ID)
		}
	}
	if slices.Contains(candidates, game.StackObjectTarget(creature.ID)) {
		t.Fatal("candidates included creature spell outside union")
	}
}

func TestStackSpellTargetCandidatesIncludeCantBeCounteredSpells(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	protected := addStackSpellWithFace(g, game.Player2, &game.CardFace{
		Name:            "Protected Spell",
		Types:           []types.Card{types.Sorcery},
		StaticAbilities: []game.StaticAbility{game.CantBeCounteredStaticBody},
	})
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Constraint: "spell",
		Predicate:  game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
	}
	source := counterTargetSpell(&spec)

	candidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player1, source, 0, &spec)

	if !slices.Contains(candidates, game.StackObjectTarget(protected.ID)) {
		t.Fatal("candidates omitted can't-be-countered stack target")
	}
}

func TestStackSpellTargetCandidatesUseFaceDownCreatureType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	faceDown := addFaceDownStackSpell(g, game.Player2, "Hidden Sorcery", []types.Card{types.Sorcery})
	creatureSpec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Constraint: "creature spell",
		Predicate: game.TargetPredicate{
			SpellCardTypes:   []types.Card{types.Creature},
			StackObjectKinds: []game.StackObjectKind{game.StackSpell},
		},
	}
	noncreatureSpec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Constraint: "noncreature spell",
		Predicate: game.TargetPredicate{
			ExcludedSpellCardTypes: []types.Card{types.Creature},
			StackObjectKinds:       []game.StackObjectKind{game.StackSpell},
		},
	}
	source := counterTargetSpell(&creatureSpec)

	creatureCandidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player1, source, 0, &creatureSpec)
	noncreatureCandidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player1, source, 0, &noncreatureSpec)

	if !slices.Contains(creatureCandidates, game.StackObjectTarget(faceDown.ID)) {
		t.Fatal("creature-spell candidates omitted face-down spell")
	}
	if slices.Contains(noncreatureCandidates, game.StackObjectTarget(faceDown.ID)) {
		t.Fatal("noncreature-spell candidates included face-down spell")
	}
}

func TestStackAbilityTargetCandidatesRespectKindsAndExcludeSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spell := addStackSpell(g, game.Player2, "Spell", []types.Card{types.Instant})
	activated := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, Controller: game.Player2}
	triggered := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player2}
	futureManaAbility := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackObjectKind(99), Controller: game.Player2}
	g.Stack.Push(activated)
	g.Stack.Push(triggered)
	g.Stack.Push(futureManaAbility)
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{
			StackObjectKinds: []game.StackObjectKind{game.StackActivatedAbility, game.StackTriggeredAbility},
		},
	}

	candidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player1, nil, activated.ID, &spec)

	if slices.Contains(candidates, game.StackObjectTarget(spell.ID)) {
		t.Fatal("ability target candidates included spell")
	}
	if slices.Contains(candidates, game.StackObjectTarget(activated.ID)) {
		t.Fatal("ability target candidates included source stack object")
	}
	if !slices.Contains(candidates, game.StackObjectTarget(triggered.ID)) {
		t.Fatal("ability target candidates omitted triggered ability")
	}
	if slices.Contains(candidates, game.StackObjectTarget(futureManaAbility.ID)) {
		t.Fatal("ability target candidates included unknown future stack-object kind")
	}
}

func TestStackSpellTargetColorQualifiersMatchSpellColors(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	blue := addStackSpellWithFace(g, game.Player2, &game.CardFace{
		Name: "Blue Spell", Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue},
	})
	red := addStackSpellWithFace(g, game.Player2, &game.CardFace{
		Name: "Red Spell", Types: []types.Card{types.Instant}, Colors: []color.Color{color.Red},
	})
	multi := addStackSpellWithFace(g, game.Player2, &game.CardFace{
		Name: "Gold Spell", Types: []types.Card{types.Instant}, Colors: []color.Color{color.White, color.Blue},
	})
	colorless := addStackSpellWithFace(g, game.Player2, &game.CardFace{
		Name: "Colorless Spell", Types: []types.Card{types.Artifact},
	})

	stackSpell := func(pred game.TargetPredicate) game.TargetSpec {
		pred.StackObjectKinds = []game.StackObjectKind{game.StackSpell}
		return game.TargetSpec{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject, Predicate: pred}
	}

	tests := []struct {
		name string
		spec game.TargetSpec
		want map[*game.StackObject]bool
	}{
		{
			name: "blue spell",
			spec: stackSpell(game.TargetPredicate{SpellColors: []color.Color{color.Blue}}),
			want: map[*game.StackObject]bool{blue: true, red: false, multi: true, colorless: false},
		},
		{
			name: "nonblue spell",
			spec: stackSpell(game.TargetPredicate{SpellExcludedColors: []color.Color{color.Blue}}),
			want: map[*game.StackObject]bool{blue: false, red: true, multi: false, colorless: true},
		},
		{
			name: "colorless spell",
			spec: stackSpell(game.TargetPredicate{SpellColorless: true}),
			want: map[*game.StackObject]bool{blue: false, red: false, multi: false, colorless: true},
		},
		{
			name: "multicolored spell",
			spec: stackSpell(game.TargetPredicate{SpellMulticolored: true}),
			want: map[*game.StackObject]bool{blue: false, red: false, multi: true, colorless: false},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := test.spec
			for obj, want := range test.want {
				got := targetMatchesSpec(g, game.Player1, 0, &spec, game.StackObjectTarget(obj.ID))
				if got != want {
					t.Fatalf("%s match for spell %d = %v, want %v", test.name, obj.ID, got, want)
				}
			}
		})
	}
}

func TestStackObjectTargetKindsMatchExactly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spell := addStackSpell(g, game.Player2, "Spell", []types.Card{types.Instant})
	activated := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, Controller: game.Player2}
	triggered := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player2}
	g.Stack.Push(activated)
	g.Stack.Push(triggered)

	tests := []struct {
		name  string
		kinds []game.StackObjectKind
		want  *game.StackObject
	}{
		{"spell", []game.StackObjectKind{game.StackSpell}, spell},
		{"activated ability", []game.StackObjectKind{game.StackActivatedAbility}, activated},
		{"triggered ability", []game.StackObjectKind{game.StackTriggeredAbility}, triggered},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := game.TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowStackObject,
				Predicate:  game.TargetPredicate{StackObjectKinds: test.kinds},
			}
			for _, obj := range []*game.StackObject{spell, activated, triggered} {
				got := targetMatchesSpec(g, game.Player1, 0, &spec, game.StackObjectTarget(obj.ID))
				if got != (obj == test.want) {
					t.Fatalf("target kind %v match = %v, want %v", obj.Kind, got, obj == test.want)
				}
			}
		})
	}
}

func playerDamageSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "player"}},
			Sequence: []game.Instruction{{Primitive: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)}}},
		}.Ability())},
	}
}

func permanentTargetSpell(constraint string) *game.CardDef {
	return permanentTargetSpellWithRange(constraint, 1, 1)
}

func optionalPermanentTargetSpell(constraint string) *game.CardDef {
	return permanentTargetSpellWithRange(constraint, 0, 1)
}

func permanentTargetSpellWithRange(constraint string, minTargets, maxTargets int) *game.CardDef {
	return permanentTargetSpellWithSpecs([]game.TargetSpec{
		{MinTargets: minTargets, MaxTargets: maxTargets, Constraint: constraint},
	})
}

func permanentTargetSpellWithSpecs(specs []game.TargetSpec) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Permanent Target Spell",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: specs,
		}.Ability())},
	}
}

func counterTargetSpell(spec *game.TargetSpec) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Counterspell",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets:  []game.TargetSpec{*spec},
			Sequence: []game.Instruction{{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}}},
		}.Ability())},
	}
}

func addStackSpell(g *game.Game, controller game.PlayerID, name string, cardTypes []types.Card) *game.StackObject {
	return addStackSpellWithFace(g, controller, &game.CardFace{Name: name, Types: cardTypes})
}

func addStackSpellWithFace(g *game.Game, controller game.PlayerID, face *game.CardFace) *game.StackObject {
	cardID := addCardToHand(g, controller, &game.CardDef{CardFace: *face})
	g.Players[controller].Hand.Remove(cardID)
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Controller: controller,
	}
	g.Stack.Push(obj)
	return obj
}

func addFaceDownStackSpell(g *game.Game, controller game.PlayerID, name string, cardTypes []types.Card) *game.StackObject {
	cardID := addCardToHand(g, controller, &game.CardDef{CardFace: game.CardFace{Name: name, Types: cardTypes}})
	g.Players[controller].Hand.Remove(cardID)
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     cardID,
		Controller:   controller,
		FaceDown:     true,
		FaceDownFace: game.FaceFront,
		FaceDownKind: game.FaceDownMorph,
	}
	g.Stack.Push(obj)
	return obj
}

func opponentChosenTargetAbilitySource() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Arena-like Land",
		Types: []types.Card{types.Land},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{types.Creature},
							Controller:     game.ControllerYou,
						},
					},
					{
						MinTargets: 1,
						MaxTargets: 1,
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{types.Creature},
							Controller:     game.ControllerYou,
						},
						Chooser: game.TargetChooserOpponent,
					},
				},
				Sequence: []game.Instruction{
					{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}},
					{Primitive: game.Tap{Object: game.TargetPermanentReference(1)}},
					{Primitive: game.Fight{}},
				},
			}.Ability(),
		}}},
	}
}

func addShroudGranter(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Shroud Granter",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Creature}},
				),
				AddKeywords: []game.Keyword{game.Shroud},
			}},
		}},
	}})
}

// TestDistinctFromPriorTargetsExcludesSharedObject checks that a second target
// spec marked DistinctFromPriorTargets ("... another target creature") never
// pairs an object with itself: with two creatures available, the enumerated
// cast actions cover the ordered distinct pairs and omit the self-pairs.
func TestDistinctFromPriorTargetsExcludesSharedObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spell := permanentTargetSpellWithSpecs([]game.TargetSpec{
		{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowPermanent,
			Predicate:  game.TargetPredicate{PermanentTypes: []types.Card{types.Creature}},
		},
		{
			MinTargets:               1,
			MaxTargets:               1,
			Allow:                    game.TargetAllowPermanent,
			Predicate:                game.TargetPredicate{PermanentTypes: []types.Card{types.Creature}},
			DistinctFromPriorTargets: true,
		},
	})
	spellID := addCardToHand(g, game.Player1, spell)
	mine := addCreaturePermanent(g, game.Player1)
	theirs := addCreaturePermanent(g, game.Player2)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	var pairs [][2]id.ID
	for _, act := range legal {
		cast, ok := act.CastSpellPayload()
		if !ok || cast.CardID != spellID {
			continue
		}
		if len(cast.Targets) != 2 {
			t.Fatalf("cast targets = %d, want 2", len(cast.Targets))
		}
		if cast.Targets[0].PermanentID == cast.Targets[1].PermanentID {
			t.Fatalf("distinct spec produced self-pair: %+v", cast.Targets)
		}
		pairs = append(pairs, [2]id.ID{cast.Targets[0].PermanentID, cast.Targets[1].PermanentID})
	}
	wantPairs := map[[2]id.ID]bool{
		{mine.ObjectID, theirs.ObjectID}: true,
		{theirs.ObjectID, mine.ObjectID}: true,
	}
	if len(pairs) != len(wantPairs) {
		t.Fatalf("distinct fight pairs = %d, want %d (%+v)", len(pairs), len(wantPairs), pairs)
	}
	for _, pair := range pairs {
		if !wantPairs[pair] {
			t.Fatalf("unexpected pair %+v", pair)
		}
	}
}
