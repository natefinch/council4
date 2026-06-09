package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// referenceGroupIDs manually enumerates group membership in battlefield order so
// it characterizes both the membership and ordering of groupMembers.
func referenceGroupIDs(g *game.Game, obj *game.StackObject, source *game.Permanent, controller game.PlayerID, group game.GroupReference) []id.ID {
	switch group.Domain() {
	case game.GroupDomainObjectControlled:
		if obj == nil || !obj.HasTriggerEvent || obj.TriggerEvent.PermanentID == 0 {
			return []id.ID{}
		}
		resolved, ok := resolvePermanentOrLastKnown(g, obj.TriggerEvent.PermanentID)
		if !ok {
			return []id.ID{}
		}
		defendingPlayer, ok := resolved.controller(g)
		if !ok {
			return []id.ID{}
		}
		ids := make([]id.ID, 0, len(g.Battlefield))
		for _, permanent := range g.Battlefield {
			if permanent.ObjectID == obj.TriggerEvent.PermanentID {
				continue
			}
			if effectiveController(g, permanent) != defendingPlayer || !permanentHasType(g, permanent, types.Creature) {
				continue
			}
			ids = append(ids, permanent.ObjectID)
		}
		return ids
	case game.GroupDomainBattlefield:
		selection := group.Selection()
		excluded := id.ID(0)
		if exclude, ok := group.Exclusion(); ok {
			resolverObj := obj
			if resolverObj == nil {
				resolverObj = &game.StackObject{Controller: controller}
			}
			excluded, _ = newReferenceResolverWithSource(g, resolverObj, source).objectIdentityID(exclude)
		}
		ids := make([]id.ID, 0, len(g.Battlefield))
		for _, permanent := range g.Battlefield {
			if excluded != 0 && permanent.ObjectID == excluded {
				continue
			}
			values := effectivePermanentValues(g, permanent)
			subject := selectionSubject{
				kind:      subjectPermanent,
				g:         g,
				permanent: permanent,
				values:    &values,
				viewer:    controller,
			}
			if selection.Controller != game.ControllerAny {
				subject.controller = values.controller
			}
			if source != nil {
				subject.sourceObjectID = source.ObjectID
			}
			if matchSelection(&subject, &selection) {
				ids = append(ids, permanent.ObjectID)
			}
		}
		return ids
	case game.GroupDomainAttachedObject:
		if source == nil || !source.AttachedTo.Exists {
			return []id.ID{}
		}
		return []id.ID{source.AttachedTo.Val}
	default:
		return []id.ID{}
	}
}

func TestGroupMembersManualParity(t *testing.T) {
	board := newParityBoard(t)
	g := board.g

	battlefieldObj := &game.StackObject{
		Controller: game.Player1,
		SourceID:   board.equipment.ObjectID,
		Targets:    []game.Target{game.PermanentTarget(board.whiteCreature.ObjectID)},
	}
	defendingObj := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventDamageDealt,
			PermanentID: board.greenCreatureP2.ObjectID,
		},
	}

	type groupCase struct {
		name  string
		group game.GroupReference
		obj   *game.StackObject
	}
	cases := []groupCase{
		{name: "all creatures", group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}), obj: battlefieldObj},
		{name: "all artifacts", group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}}), obj: battlefieldObj},
		{name: "all enchantments", group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}}), obj: battlefieldObj},
		{name: "all nonland permanents", group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}}), obj: battlefieldObj},
		{name: "all permanents", group: game.BattlefieldGroup(game.Selection{}), obj: battlefieldObj},
		{name: "creatures you control", group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}), obj: battlefieldObj},
		{name: "other creatures you control", group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true}), obj: battlefieldObj},
		{name: "equipped creature", group: game.AttachedObjectGroup(game.SourcePermanentReference()), obj: battlefieldObj},
		{name: "all creatures except target", group: game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.TargetPermanentReference(0)), obj: battlefieldObj},
		{name: "other creatures defending player controls", group: game.ObjectControlledGroupExcluding(game.EventPermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.EventPermanentReference()), obj: defendingObj},
	}

	for _, c := range cases {
		source, _ := permanentByObjectID(g, c.obj.SourceID)
		want := referenceGroupIDs(g, c.obj, source, c.obj.Controller, c.group)
		got := newReferenceResolver(g, c.obj).groupMembers(c.group)
		if !slices.Equal(got, want) {
			t.Errorf("%s groupMembers = %v, want %v", c.name, got, want)
		}
	}
}

func TestGroupMembersOtherCreaturesYouControlNilSourceMatchesNothing(t *testing.T) {
	board := newParityBoard(t)
	g := board.g
	obj := &game.StackObject{Controller: game.Player1}

	group := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
		ExcludeSource: true,
	})
	got := newReferenceResolver(g, obj).groupMembers(group)
	if len(got) != 0 {
		t.Fatalf("groupMembers with nil source = %v, want empty", got)
	}
}

func TestGroupMembersExceptTargetUsesPrimitiveTargetIndex(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	third := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets: []game.Target{
			game.PermanentTarget(first.ObjectID),
			game.PermanentTarget(second.ObjectID),
		},
	}

	group := game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.TargetPermanentReference(1))
	got := newReferenceResolver(g, obj).groupMembers(group)
	want := []id.ID{first.ObjectID, third.ObjectID}
	if !slices.Equal(got, want) {
		t.Fatalf("groupMembers excluding target 1 = %v, want %v", got, want)
	}
}

func TestAttachedObjectGroupResolvesAndEmptyWhenUnattached(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Creature",
		Types: []types.Card{types.Creature},
	}})
	equipment := addCombatPermanent(g, game.Player1, equipmentWithStaticEffect(nil))
	obj := &game.StackObject{Controller: game.Player1, SourceID: equipment.ObjectID}
	resolver := newReferenceResolver(g, obj)
	group := game.AttachedObjectGroup(game.SourcePermanentReference())

	if got := resolver.groupMembers(group); len(got) != 0 {
		t.Fatalf("unattached equipment group = %v, want empty", got)
	}

	if !attachPermanent(g, equipment, creature) {
		t.Fatal("attachPermanent failed")
	}
	got := resolver.groupMembers(group)
	if !slices.Equal(got, []id.ID{creature.ObjectID}) {
		t.Fatalf("attached equipment group = %v, want [%v]", got, creature.ObjectID)
	}
}

func TestReferenceResolverObjectUsesLastKnownInformation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Departed",
		Types: []types.Card{types.Creature},
	}})
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(creature.ObjectID)},
	}
	snapshot := snapshotPermanent(g, creature, zone.Battlefield)
	removePermanentFromBattlefield(g, creature.ObjectID)
	rememberLastKnown(g, &snapshot)

	resolver := newReferenceResolver(g, obj)
	resolved, ok := resolver.object(game.TargetPermanentReference(0))
	if !ok {
		t.Fatal("object reference did not resolve to last-known information")
	}
	if resolved.permanent != nil {
		t.Fatal("expected snapshot resolution, got a live permanent")
	}
	if controller, ok := resolved.controller(g); !ok || controller != game.Player2 {
		t.Fatalf("snapshot controller = %v (%v), want Player2", controller, ok)
	}
	if owner, ok := resolved.owner(); !ok || owner != game.Player2 {
		t.Fatalf("snapshot owner = %v (%v), want Player2", owner, ok)
	}
}

func TestReferenceResolverControllerAndOwnerReferences(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name:  "Borrowed Beast",
		Types: []types.Card{types.Creature},
	}})
	creature.Controller = game.Player2
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(creature.ObjectID)},
	}
	resolver := newReferenceResolver(g, obj)

	controller, ok := resolver.player(game.ObjectControllerReference(game.TargetPermanentReference(0)))
	if !ok || controller != game.Player2 {
		t.Fatalf("object controller reference = %v (%v), want Player2", controller, ok)
	}
	owner, ok := resolver.player(game.ObjectOwnerReference(game.TargetPermanentReference(0)))
	if !ok || owner != game.Player3 {
		t.Fatalf("object owner reference = %v (%v), want Player3", owner, ok)
	}
}

func TestReferenceResolverPlayerReferenceRejectsDeadPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player2].Eliminated = true
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player2)},
	}
	resolver := newReferenceResolver(g, obj)
	if _, ok := resolver.player(game.TargetPlayerReference(0)); ok {
		t.Fatal("target player reference resolved an eliminated player")
	}
}
