package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// parityBoard is a diverse battlefield used to characterize the shared Selection
// matcher against the legacy per-field logic it replaced.
type parityBoard struct {
	g                *game.Game
	all              []*game.Permanent
	redFlyerTapped   *game.Permanent
	whiteCreature    *game.Permanent
	artifact         *game.Permanent
	enchantment      *game.Permanent
	forest           *game.Permanent
	artifactCreature *game.Permanent
	greenCreatureP2  *game.Permanent
	tokenCreature    *game.Permanent
	equipment        *game.Permanent
}

func newParityBoard(t *testing.T) parityBoard {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	board := parityBoard{g: g}

	board.redFlyerTapped = addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Red Flyer",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Red},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.R}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Flying),
		}},
	}})
	board.redFlyerTapped.Tapped = true

	board.whiteCreature = addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "White Squire",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.White},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		ManaCost:  opt.Val(cost.Mana{cost.W}),
	}})

	board.artifact = addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Stone Relic",
		Types:    []types.Card{types.Artifact},
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
	}})

	board.enchantment = addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Aura Veil",
		Types: []types.Card{types.Enchantment},
	}})

	board.forest = addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Snow Forest",
		Types:      []types.Card{types.Land},
		Supertypes: []types.Super{types.Basic, types.Snow},
		Subtypes:   []types.Sub{types.Forest},
	}})

	board.artifactCreature = addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Brass Golem",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 4}),
		ManaCost:  opt.Val(cost.Mana{cost.O(4)}),
	}})

	board.greenCreatureP2 = addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Forest Brute",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Green},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
		ManaCost:  opt.Val(cost.Mana{cost.O(4), cost.G}),
	}})

	board.tokenCreature = addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Spirit Token",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.White},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	board.tokenCreature.Token = true

	board.equipment = addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Iron Blade",
		Types: []types.Card{types.Artifact},
	}})
	board.equipment.AttachedTo = opt.Val(board.whiteCreature.ObjectID)

	board.all = []*game.Permanent{
		board.redFlyerTapped, board.whiteCreature, board.artifact, board.enchantment,
		board.forest, board.artifactCreature, board.greenCreatureP2, board.tokenCreature,
		board.equipment,
	}
	return board
}

// referenceConditionFilterMatches reproduces the legacy
// permanentMatchesConditionFilter per-permanent semantics against a Selection.
func referenceConditionFilterMatches(g *game.Game, permanent *game.Permanent, sel game.Selection, useBase bool) bool {
	var values permanentEffectiveValues
	if useBase {
		values = basePermanentValues(g, permanent)
	} else {
		values = effectivePermanentValues(g, permanent)
	}
	for _, cardType := range sel.RequiredTypes {
		if !slices.Contains(values.types, cardType) {
			return false
		}
	}
	for _, supertype := range sel.Supertypes {
		if !slices.Contains(values.supertypes, supertype) {
			return false
		}
	}
	if len(sel.SubtypesAny) > 0 && !slices.ContainsFunc(sel.SubtypesAny, func(subtype types.Sub) bool {
		return slices.Contains(values.subtypes, subtype)
	}) {
		return false
	}
	if sel.Power.Exists {
		if useBase {
			return false
		}
		if !values.powerOK || !sel.Power.Val.Matches(values.power) {
			return false
		}
	}
	if sel.Toughness.Exists {
		if useBase {
			return false
		}
		if !values.toughnessOK || !sel.Toughness.Val.Matches(values.toughness) {
			return false
		}
	}
	return true
}

func matchConditionFilter(g *game.Game, permanent *game.Permanent, sel game.Selection, useBase bool) bool {
	var values permanentEffectiveValues
	if useBase {
		values = basePermanentValues(g, permanent)
	} else {
		values = effectivePermanentValues(g, permanent)
	}
	subject := selectionSubject{
		kind:      subjectPermanent,
		g:         g,
		permanent: permanent,
		values:    &values,
		viewer:    game.Player1,
		useBase:   useBase,
	}
	return matchSelection(&subject, &sel)
}

func TestSharedMatcherPermanentFilterParity(t *testing.T) {
	board := newParityBoard(t)
	g := board.g

	filters := map[string]game.Selection{
		"empty":          {},
		"creature":       {RequiredTypes: []types.Card{types.Creature}},
		"artifact-crea":  {RequiredTypes: []types.Card{types.Artifact, types.Creature}},
		"snow-super":     {Supertypes: []types.Super{types.Basic}},
		"forest-subtype": {SubtypesAny: []types.Sub{types.Forest}},
		"power-ge-3":     {Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})},
		"toughness-ge-4": {Toughness: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})},
	}

	for name, filter := range filters {
		for _, useBase := range []bool{false, true} {
			for _, permanent := range board.all {
				want := referenceConditionFilterMatches(g, permanent, filter, useBase)
				got := matchConditionFilter(g, permanent, filter, useBase)
				if got != want {
					t.Errorf("filter %q useBase=%v on %s: matcher=%v reference=%v", name, useBase, permanentDebugName(g, permanent), got, want)
				}
			}
		}
	}
}

// referenceContinuousGroupMatches is an independent reference implementation of
// continuous-effect group matching. It must not call matchSelection,
// continuousSelectionApplies, or any production group-matching helper so that
// TestSharedMatcherContinuousGroupParity can detect predicate bugs in those
// production paths.
func referenceContinuousGroupMatches(g *game.Game, source *game.Permanent, controller game.PlayerID, permanent *game.Permanent, group game.GroupReference) bool {
	switch group.Domain() {
	case game.GroupDomainAttachedObject:
		// The group is the single permanent the source is currently attached to.
		return source != nil && source.AttachedTo.Exists && permanent.ObjectID == source.AttachedTo.Val

	case game.GroupDomainBattlefield:
		sel := group.Selection()
		// ExcludeSource: the permanent must not be the source permanent itself.
		if sel.ExcludeSource && (source == nil || permanent.ObjectID == source.ObjectID) {
			return false
		}
		// All required types must be present on the permanent.
		for _, required := range sel.RequiredTypes {
			if !permanentHasType(g, permanent, required) {
				return false
			}
		}
		// No excluded type may be present.
		for _, excluded := range sel.ExcludedTypes {
			if permanentHasType(g, permanent, excluded) {
				return false
			}
		}
		// Controller filter applied directly using effective controller.
		permController := effectiveController(g, permanent)
		switch sel.Controller {
		case game.ControllerYou:
			if permController != controller {
				return false
			}
		case game.ControllerOpponent, game.ControllerNotYou:
			if permController == controller {
				return false
			}
		default:
		}
		return true

	case game.GroupDomainObjectControlled:
		// All permanents whose effective controller equals the anchor's effective
		// controller. Only SourcePermanentReference anchors are handled here because
		// that is the only form used in the parity test cases.
		anchor, ok := group.Anchor()
		if !ok || anchor.Kind() != game.ObjectReferenceSourcePermanent || source == nil {
			return false
		}
		anchorController := effectiveController(g, source)
		sel := group.Selection()
		if sel.ExcludeSource && permanent.ObjectID == source.ObjectID {
			return false
		}
		for _, required := range sel.RequiredTypes {
			if !permanentHasType(g, permanent, required) {
				return false
			}
		}
		for _, excluded := range sel.ExcludedTypes {
			if permanentHasType(g, permanent, excluded) {
				return false
			}
		}
		return effectiveController(g, permanent) == anchorController

	default:
		return false
	}
}

func TestSharedMatcherContinuousGroupParity(t *testing.T) {
	board := newParityBoard(t)
	g := board.g
	source := board.equipment

	cases := []struct {
		name  string
		group game.GroupReference
	}{
		{name: "all creatures", group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})},
		{name: "all artifacts", group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}})},
		{name: "all enchantments", group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}})},
		{name: "all nonland permanents", group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}})},
		{name: "all permanents", group: game.BattlefieldGroup(game.Selection{})},
		{name: "creatures you control", group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou})},
		{name: "creatures opponent controls", group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerNotYou})},
		{name: "other creatures you control", group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true})},
		{name: "equipped creature", group: game.AttachedObjectGroup(game.SourcePermanentReference())},
		{name: "creatures source controls", group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}})},
		{name: "permanents source controls", group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{})},
	}

	for _, tc := range cases {
		for _, permanent := range board.all {
			want := referenceContinuousGroupMatches(g, source, game.Player1, permanent, tc.group)

			values := effectivePermanentValues(g, permanent)
			gotContinuous := continuousEffectApplies(g, permanent, &values, &game.ContinuousEffect{
				SourceObjectID: source.ObjectID,
				Controller:     game.Player1,
				Group:          tc.group,
			})
			if gotContinuous != want {
				t.Errorf("%s continuous path on %s: matcher=%v reference=%v", tc.name, permanentDebugName(g, permanent), gotContinuous, want)
			}
		}
	}
}

// TestSharedMatcherContinuousGroupParityDetectsMismatch ensures the reference
// implementation is actually independent: an intentionally wrong predicate
// (always returns true) must produce at least one mismatch against production.
func TestSharedMatcherContinuousGroupParityDetectsMismatch(t *testing.T) {
	board := newParityBoard(t)
	g := board.g
	source := board.equipment
	group := game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})

	mismatchFound := false
	for _, permanent := range board.all {
		// Broken oracle: always matches.
		want := true
		values := effectivePermanentValues(g, permanent)
		got := continuousEffectApplies(g, permanent, &values, &game.ContinuousEffect{
			SourceObjectID: source.ObjectID,
			Controller:     game.Player1,
			Group:          group,
		})
		if got != want {
			mismatchFound = true
			break
		}
	}
	if !mismatchFound {
		t.Error("broken oracle did not trigger a mismatch: parity test cannot detect predicate bugs")
	}
}

// referenceEventPermanentFilters reproduces the legacy
// eventPermanentTypeFiltersMatch plus RequireNonToken behavior.
func referenceEventPermanentFilters(g *game.Game, event game.Event, required, excluded []types.Card, nonToken bool) bool {
	for _, cardType := range required {
		if !eventPermanentHasType(g, event, cardType) {
			return false
		}
	}
	for _, cardType := range excluded {
		if eventPermanentHasType(g, event, cardType) {
			return false
		}
	}
	if nonToken && eventPermanentIsToken(g, event) {
		return false
	}
	return true
}

func matchEventSubject(g *game.Game, event game.Event, sel game.Selection) bool {
	subject := selectionSubject{kind: subjectEventPermanent, g: g, event: event}
	return matchSelection(&subject, &sel)
}

func TestSharedMatcherTriggerSubjectParity(t *testing.T) {
	board := newParityBoard(t)
	g := board.g

	type spec struct {
		required []types.Card
		excluded []types.Card
		nonToken bool
	}
	specs := map[string]spec{
		"none":           {},
		"creature":       {required: []types.Card{types.Creature}},
		"artifact":       {required: []types.Card{types.Artifact}},
		"excluded-land":  {excluded: []types.Card{types.Land}},
		"nontoken":       {nonToken: true},
		"nontoken-creat": {required: []types.Card{types.Creature}, nonToken: true},
	}

	for name, s := range specs {
		for _, permanent := range board.all {
			event := game.Event{Kind: game.EventPermanentDied, PermanentID: permanent.ObjectID}
			sel := game.Selection{RequiredTypes: s.required, ExcludedTypes: s.excluded, NonToken: s.nonToken}
			want := referenceEventPermanentFilters(g, event, s.required, s.excluded, s.nonToken)
			got := matchEventSubject(g, event, sel)
			if got != want {
				t.Errorf("trigger subject %q on %s: matcher=%v reference=%v", name, permanentDebugName(g, permanent), got, want)
			}
		}
	}
}

// referenceEventCardFilters reproduces the legacy eventCardTypeFiltersMatch.
func referenceEventCardFilters(g *game.Game, event game.Event, required, excluded []types.Card) bool {
	cardTypes := event.CardTypes
	if len(cardTypes) == 0 && event.CardID != 0 {
		if card, ok := g.GetCardInstance(event.CardID); ok {
			cardTypes = cardFaceOrDefault(card, game.FaceFront).Types
		}
	}
	for _, cardType := range required {
		if !slices.Contains(cardTypes, cardType) {
			return false
		}
	}
	for _, cardType := range excluded {
		if slices.Contains(cardTypes, cardType) {
			return false
		}
	}
	return true
}

func matchCastSpellSubject(g *game.Game, event game.Event, sel game.Selection) bool {
	subject := selectionSubject{kind: subjectCastSpell, g: g, event: event, cardTypes: eventSpellCardTypes(g, event)}
	return matchSelection(&subject, &sel)
}

func TestSharedMatcherTriggerCardParity(t *testing.T) {
	board := newParityBoard(t)
	g := board.g

	instantID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Quick Bolt",
		Types: []types.Card{types.Instant},
	}})

	events := map[string]game.Event{
		"cardtypes-instant":  {Kind: game.EventSpellCast, CardTypes: []types.Card{types.Instant}},
		"cardtypes-creature": {Kind: game.EventSpellCast, CardTypes: []types.Card{types.Creature}},
		"cardid-instant":     {Kind: game.EventSpellCast, CardID: instantID},
		"empty":              {Kind: game.EventSpellCast},
	}

	type spec struct {
		required []types.Card
		excluded []types.Card
	}
	specs := map[string]spec{
		"require-instant":  {required: []types.Card{types.Instant}},
		"require-creature": {required: []types.Card{types.Creature}},
		"exclude-instant":  {excluded: []types.Card{types.Instant}},
	}

	for eventName, event := range events {
		for specName, s := range specs {
			sel := game.Selection{RequiredTypes: s.required, ExcludedTypes: s.excluded}
			want := referenceEventCardFilters(g, event, s.required, s.excluded)
			got := matchCastSpellSubject(g, event, sel)
			if got != want {
				t.Errorf("trigger card %q/%q: matcher=%v reference=%v", eventName, specName, got, want)
			}
		}
	}
}

// TestSharedMatcherRichCombination exercises a Selection combining predicate
// fields no single historical mass-group shortcut could express, confirming the matcher applies
// every constraint conjunctively.
func TestSharedMatcherRichCombination(t *testing.T) {
	board := newParityBoard(t)
	g := board.g

	sel := game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		ColorsAny:     []color.Color{color.Red},
		Keyword:       game.Flying,
		Tapped:        game.TriTrue,
		Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3}),
	}

	matched := make([]id.ID, 0)
	for _, permanent := range board.all {
		values := effectivePermanentValues(g, permanent)
		subject := selectionSubject{kind: subjectPermanent, g: g, permanent: permanent, values: &values, viewer: game.Player1, clampPower: true}
		if matchSelection(&subject, &sel) {
			matched = append(matched, permanent.ObjectID)
		}
	}

	if len(matched) != 1 || matched[0] != board.redFlyerTapped.ObjectID {
		t.Fatalf("rich combination matched %v, want only red tapped flyer %d", matched, board.redFlyerTapped.ObjectID)
	}
}

func TestSelectionOnlyTargetSpecMatchesPermanent(t *testing.T) {
	board := newParityBoard(t)
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}),
	}

	if !permanentTargetMatchesSpec(
		board.g,
		game.Player1,
		0,
		&spec,
		board.whiteCreature.ObjectID,
	) {
		t.Fatal("Selection-only creature target did not match a creature")
	}
	if permanentTargetMatchesSpec(
		board.g,
		game.Player1,
		0,
		&spec,
		board.artifact.ObjectID,
	) {
		t.Fatal("Selection-only creature target matched a noncreature artifact")
	}
}

func TestSelectionAnyOfKeepsQualifiedAlternativesAndCommonController(t *testing.T) {
	t.Parallel()

	board := newParityBoard(t)
	nonbasicLand := addCombatPermanent(board.g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Utility Land",
		Types: []types.Card{types.Land},
	}})
	basicArtifactLand := addCombatPermanent(board.g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:       "Basic Artifact Land",
		Types:      []types.Card{types.Artifact, types.Land},
		Supertypes: []types.Super{types.Basic},
	}})
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			Controller: game.ControllerOpponent,
			AnyOf: []game.Selection{
				{RequiredTypes: []types.Card{types.Artifact}},
				{RequiredTypes: []types.Card{types.Enchantment}},
				{RequiredTypes: []types.Card{types.Land}, ExcludedSupertype: types.Basic},
			},
		}),
	}

	for name, permanent := range map[string]*game.Permanent{
		"enchantment":         board.enchantment,
		"artifact creature":   board.artifactCreature,
		"nonbasic land":       nonbasicLand,
		"basic artifact land": basicArtifactLand,
	} {
		if !permanentTargetMatchesSpec(board.g, game.Player1, 0, &spec, permanent.ObjectID) {
			t.Errorf("%s did not match the qualified union", name)
		}
	}
	for name, permanent := range map[string]*game.Permanent{
		"your artifact":     board.artifact,
		"opponent creature": board.greenCreatureP2,
		"your basic land":   board.forest,
	} {
		if permanentTargetMatchesSpec(board.g, game.Player1, 0, &spec, permanent.ObjectID) {
			t.Errorf("%s unexpectedly matched the qualified union", name)
		}
	}
}

func permanentDebugName(g *game.Game, permanent *game.Permanent) string {
	if def, ok := permanentCardDef(g, permanent); ok {
		return def.Name
	}
	return "permanent"
}

// matchSelectionForPermanent runs sel through the shared matcher against a
// battlefield permanent, mirroring how permanentTargetMatchesSpec builds the
// subject so the RequiredCounter predicate is exercised on real counters.
func matchSelectionForPermanent(g *game.Game, controller game.PlayerID, sel game.Selection, permanent *game.Permanent) bool {
	values := effectivePermanentValues(g, permanent)
	subject := selectionSubject{
		kind:       subjectPermanent,
		g:          g,
		permanent:  permanent,
		values:     &values,
		viewer:     controller,
		clampPower: true,
	}
	if sel.Controller != game.ControllerAny {
		subject.controller = effectiveController(g, permanent)
	}
	return matchSelection(&subject, &sel)
}

func TestMatchSelectionExcludedSubtypes(t *testing.T) {
	board := newParityBoard(t)
	g := board.g
	excludeForest := game.Selection{
		ExcludedSubtype: types.Forest,
	}
	if matchSelectionForPermanent(g, game.Player1, excludeForest, board.forest) {
		t.Error("a Forest permanent must not match an ExcludedSubtypes{Forest} selection")
	}
	if !matchSelectionForPermanent(g, game.Player1, excludeForest, board.whiteCreature) {
		t.Error("a permanent without the Forest subtype should match an ExcludedSubtypes{Forest} selection")
	}
}

func TestMatchSelectionRequiredCounter(t *testing.T) {
	board := newParityBoard(t)
	g := board.g
	withCounter := board.whiteCreature
	withCounter.Counters.Add(counter.PlusOnePlusOne, 1)
	plusOne := game.Selection{
		RequiredTypes:   []types.Card{types.Creature},
		MatchCounter:    true,
		RequiredCounter: counter.PlusOnePlusOne,
	}
	if !matchSelectionForPermanent(g, game.Player1, plusOne, withCounter) {
		t.Error("creature with a +1/+1 counter should match RequiredCounter selection")
	}
	if matchSelectionForPermanent(g, game.Player1, plusOne, board.redFlyerTapped) {
		t.Error("creature without a +1/+1 counter must not match RequiredCounter selection")
	}
	charge := game.Selection{
		RequiredTypes:   []types.Card{types.Creature},
		MatchCounter:    true,
		RequiredCounter: counter.Charge,
	}
	if matchSelectionForPermanent(g, game.Player1, charge, withCounter) {
		t.Error("a +1/+1 counter must not satisfy a charge-counter requirement")
	}
}

func TestMatchSelectionAnyCounter(t *testing.T) {
	board := newParityBoard(t)
	g := board.g
	withCounter := board.whiteCreature
	withCounter.Counters.Add(counter.Charge, 1)
	anyCounter := game.Selection{MatchAnyCounter: true}
	if !matchSelectionForPermanent(g, game.Player1, anyCounter, withCounter) {
		t.Error("a permanent with any counter should match a MatchAnyCounter selection")
	}
	if matchSelectionForPermanent(g, game.Player1, anyCounter, board.redFlyerTapped) {
		t.Error("a permanent without counters must not match a MatchAnyCounter selection")
	}
}

func TestMatchSelectionNoCounters(t *testing.T) {
	board := newParityBoard(t)
	g := board.g
	withCounter := board.whiteCreature
	withCounter.Counters.Add(counter.Charge, 1)
	noCounters := game.Selection{MatchNoCounters: true}
	if matchSelectionForPermanent(g, game.Player1, noCounters, withCounter) {
		t.Error("a permanent carrying a counter must not match a MatchNoCounters selection")
	}
	if !matchSelectionForPermanent(g, game.Player1, noCounters, board.redFlyerTapped) {
		t.Error("a permanent with no counters should match a MatchNoCounters selection")
	}
}
