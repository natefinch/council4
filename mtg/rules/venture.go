package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// This file implements the "venture into the dungeon" and "venture into
// Undercity" keyword actions (CR 309.6) and dungeon completion (CR 309.7). A
// venture enters a dungeon or advances the current one; entering a room queues
// the room's ability to be put on the stack using the normal triggered-ability
// machinery, so room-ability trigger multipliers apply. A dungeon's final room
// ability ends with a CompleteDungeon instruction, so completion is recorded as
// the final room's ability resolves.

// ventureIntoDungeon performs the "venture into the dungeon" keyword action for
// playerID. If the player is not in a dungeon they choose one of the three
// ordinary dungeons and enter its first room; if they are already in a dungeon
// they advance to the next room. It reports whether a room was entered.
func (e *Engine) ventureIntoDungeon(g *game.Game, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated {
		return false
	}
	if player.Dungeon.Exists {
		return e.advanceDungeon(g, playerID, agents, log)
	}
	dungeonID, ok := e.chooseDungeon(g, playerID, game.OrdinaryDungeons(), agents, log)
	if !ok {
		return false
	}
	return e.enterDungeon(g, playerID, dungeonID, agents, log)
}

// ventureIntoUndercity performs the "venture into Undercity" keyword action for
// playerID. If the player is already in a dungeon they advance that dungeon;
// otherwise they enter Undercity, the only way to enter it. It reports whether a
// room was entered.
func (e *Engine) ventureIntoUndercity(g *game.Game, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated {
		return false
	}
	if player.Dungeon.Exists {
		return e.advanceDungeon(g, playerID, agents, log)
	}
	return e.enterDungeon(g, playerID, game.DungeonUndercity, agents, log)
}

// enterDungeon puts playerID into the given dungeon and enters its first room,
// assigning the dungeon a fresh synthetic object id. An ordinary dungeon enters
// its entrance (room 0); a free-traversal dungeon lets the player choose any
// room as their first.
func (e *Engine) enterDungeon(g *game.Game, playerID game.PlayerID, dungeonID game.DungeonID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok {
		return false
	}
	def, ok := game.DungeonByID(dungeonID)
	if !ok {
		return false
	}
	state := game.DungeonState{
		ObjectID: g.IDGen.Next(),
		Dungeon:  dungeonID,
		Room:     0,
	}
	if def.FreeTraversal {
		room, ok := e.chooseUnvisitedRoom(g, playerID, def, state, agents, log)
		if !ok {
			return false
		}
		state.Room = room
	}
	player.Dungeon = opt.Val(state)
	enterDungeonRoom(g, playerID)
	return true
}

// advanceDungeon advances playerID from their current room to the next one. An
// ordinary dungeon advances along the room's Next edges (choosing among branches);
// a free-traversal dungeon advances to any room the player has not yet visited.
func (e *Engine) advanceDungeon(g *game.Game, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Dungeon.Exists {
		return false
	}
	state := player.Dungeon.Val
	def, ok := game.DungeonByID(state.Dungeon)
	if !ok {
		return false
	}
	var next int
	if def.FreeTraversal {
		next, ok = e.chooseUnvisitedRoom(g, playerID, def, state, agents, log)
		if !ok {
			return false
		}
	} else {
		room, ok := def.Room(state.Room)
		if !ok || len(room.Next) == 0 {
			return false
		}
		next = room.Next[0]
		if len(room.Next) > 1 {
			next = e.chooseDungeonBranch(g, playerID, def, room.Next, agents, log)
		}
	}
	state.Room = next
	player.Dungeon = opt.Val(state)
	enterDungeonRoom(g, playerID)
	return true
}

// chooseUnvisitedRoom asks playerID which unvisited room of a free-traversal
// dungeon to enter next, from every room whose visited bit is not set. It reports
// false only when every room has been visited (which cannot happen before
// completion).
func (e *Engine) chooseUnvisitedRoom(g *game.Game, playerID game.PlayerID, def *game.DungeonDef, state game.DungeonState, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (int, bool) {
	unvisited := make([]int, 0, len(def.Rooms))
	for i := range def.Rooms {
		if state.Visited&roomBit(i) == 0 {
			unvisited = append(unvisited, i)
		}
	}
	if len(unvisited) == 0 {
		return 0, false
	}
	if len(unvisited) == 1 {
		return unvisited[0], true
	}
	options := make([]game.ChoiceOption, len(unvisited))
	for i, roomIndex := range unvisited {
		label := ""
		if room, ok := def.Room(roomIndex); ok {
			label = room.Name
		}
		options[i] = game.ChoiceOption{Index: i, Label: label}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose a room to enter",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, log)
	if len(selected) == 0 || selected[0] < 0 || selected[0] >= len(unvisited) {
		return unvisited[0], true
	}
	return unvisited[selected[0]], true
}

// enterDungeonRoom marks the current room visited, registers the dungeon as a
// last-known object, emits the venture event, and queues the current room's
// ability to be put on the stack, marked final when entering it completed the
// dungeon's rooms.
func enterDungeonRoom(g *game.Game, playerID game.PlayerID) {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Dungeon.Exists {
		return
	}
	state := player.Dungeon.Val
	def, ok := game.DungeonByID(state.Dungeon)
	if !ok {
		return
	}
	room, ok := def.Room(state.Room)
	if !ok {
		return
	}
	state.Visited |= roomBit(state.Room)
	player.Dungeon = opt.Val(state)
	final := dungeonRoomIsFinal(def, state)
	rememberDungeonObject(g, playerID, state, def)
	emitEvent(g, game.Event{
		Kind:       game.EventVenturedIntoDungeon,
		Controller: playerID,
		Player:     playerID,
	})
	g.PendingRoomAbilities = append(g.PendingRoomAbilities, game.RoomAbilityTrigger{
		Controller:      playerID,
		DungeonObjectID: state.ObjectID,
		Dungeon:         state.Dungeon,
		Room:            state.Room,
		Final:           final,
		Ability:         game.TriggeredAbility{Content: room.Ability},
	})
}

// dungeonRoomIsFinal reports whether entering the player's current room completed
// the dungeon's rooms: the terminal room of an ordinary dungeon, or the last
// unvisited room of a free-traversal dungeon (every room's visited bit set).
func dungeonRoomIsFinal(def *game.DungeonDef, state game.DungeonState) bool {
	if def.FreeTraversal {
		return state.Visited == allRoomsVisitedMask(def)
	}
	room, ok := def.Room(state.Room)
	return ok && room.Final()
}

// allRoomsVisitedMask returns the visited-bitmask value in which every room of
// the dungeon has been visited.
func allRoomsVisitedMask(def *game.DungeonDef) uint32 {
	mask := uint32(0)
	for i := range def.Rooms {
		mask |= roomBit(i)
	}
	return mask
}

// roomBit returns the visited-bitmask bit for a room index, masked to the 32-bit
// visited-mask width (dungeons have far fewer than 32 rooms).
func roomBit(index int) uint32 {
	return uint32(1) << (index & 31)
}

// rememberDungeonObject records a last-known snapshot of the dungeon so
// room-ability trigger multipliers and player observation can resolve the
// dungeon's owner and type. Dungeons are shared objects that never enter the
// battlefield, so the snapshot is their only object representation.
func rememberDungeonObject(g *game.Game, playerID game.PlayerID, state game.DungeonState, def *game.DungeonDef) {
	snapshot := game.ObjectSnapshot{
		ObjectID:   state.ObjectID,
		Owner:      playerID,
		Controller: playerID,
		Name:       def.Name,
		Types:      []types.Card{types.Dungeon},
	}
	if def.ID == game.DungeonUndercity {
		snapshot.Subtypes = []types.Sub{types.Undercity}
	}
	rememberLastKnown(g, &snapshot)
}

// chooseDungeon asks playerID which dungeon to venture into, from the given
// options. It always returns one of the options, falling back to the first.
func (e *Engine) chooseDungeon(g *game.Game, playerID game.PlayerID, options []game.DungeonID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (game.DungeonID, bool) {
	if len(options) == 0 {
		return game.DungeonNone, false
	}
	choiceOptions := make([]game.ChoiceOption, len(options))
	for i, dungeonID := range options {
		label := ""
		if def, ok := game.DungeonByID(dungeonID); ok {
			label = def.Name
		}
		choiceOptions[i] = game.ChoiceOption{Index: i, Label: label}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose a dungeon to venture into",
		Options:          choiceOptions,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, log)
	if len(selected) == 0 || selected[0] < 0 || selected[0] >= len(options) {
		return options[0], true
	}
	return options[selected[0]], true
}

// chooseDungeonBranch asks playerID which of the given next rooms to advance to.
// It always returns one of the next-room indices, falling back to the first.
func (e *Engine) chooseDungeonBranch(g *game.Game, playerID game.PlayerID, def *game.DungeonDef, next []int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) int {
	choiceOptions := make([]game.ChoiceOption, len(next))
	for i, roomIndex := range next {
		label := ""
		if room, ok := def.Room(roomIndex); ok {
			label = room.Name
		}
		choiceOptions[i] = game.ChoiceOption{Index: i, Label: label}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose the next room",
		Options:          choiceOptions,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, log)
	if len(selected) == 0 || selected[0] < 0 || selected[0] >= len(next) {
		return next[0]
	}
	return next[selected[0]]
}

// handleVentureIntoDungeon resolves the VentureIntoDungeon primitive.
func handleVentureIntoDungeon(r *effectResolver, prim game.VentureIntoDungeon) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	res.succeeded = r.engine.ventureIntoDungeon(r.game, playerID, r.agents, r.log)
	return res
}

// handleVentureIntoUndercity resolves the VentureIntoUndercity primitive.
func handleVentureIntoUndercity(r *effectResolver, prim game.VentureIntoUndercity) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	res.succeeded = r.engine.ventureIntoUndercity(r.game, playerID, r.agents, r.log)
	return res
}

// completeDungeonForStackObject completes the dungeon when a final room's ability
// leaves the stack (CR 309.7), whether it resolved, was countered, or was removed
// for having no legal targets. It delegates to completeDungeonForRoomMark.
func completeDungeonForStackObject(g *game.Game, obj *game.StackObject) {
	if obj == nil {
		return
	}
	completeDungeonForRoomMark(g, obj.DungeonRoom)
}

// completeDungeonForRoomMark completes the dungeon for a final room-ability
// marker as its ability leaves the stack, or is dropped before reaching the
// stack for having no legal targets (CR 309.7). It runs once per dungeon entry:
// the marker records the venturing player, dungeon entry object id, and room, so
// a copied or doubled final-room ability completes the dungeon exactly once — the
// first to leave the stack clears the dungeon state, and later copies find no
// matching state. It clears the dungeon state, increments DungeonsCompleted, and
// emits EventCompletedDungeon.
func completeDungeonForRoomMark(g *game.Game, marked opt.V[game.DungeonRoomMark]) {
	if !marked.Exists || !marked.Val.Final {
		return
	}
	mark := marked.Val
	player, ok := playerByID(g, mark.Owner)
	if !ok || !player.Dungeon.Exists {
		return
	}
	state := player.Dungeon.Val
	if state.ObjectID != mark.ObjectID || state.Room != mark.Room {
		return
	}
	player.Dungeon = opt.V[game.DungeonState]{}
	player.DungeonsCompleted++
	emitEvent(g, game.Event{
		Kind:       game.EventCompletedDungeon,
		Controller: mark.Owner,
		Player:     mark.Owner,
	})
}

// drainReadyRoomAbilities removes every queued dungeon room ability and returns
// it as a pending triggered ability ready to be put on the stack. Each is an
// ordinary event-backed trigger whose source is the dungeon, so room-ability
// trigger multipliers see it as a room ability of that dungeon.
func drainReadyRoomAbilities(g *game.Game) []pendingTriggeredAbility {
	if len(g.PendingRoomAbilities) == 0 {
		return nil
	}
	pending := make([]pendingTriggeredAbility, 0, len(g.PendingRoomAbilities))
	for i := range g.PendingRoomAbilities {
		trigger := &g.PendingRoomAbilities[i]
		ability := trigger.Ability
		pending = append(pending, pendingTriggeredAbility{
			controller: trigger.Controller,
			sourceID:   trigger.DungeonObjectID,
			inline:     &ability,
			event: game.Event{
				Kind:       game.EventVenturedIntoDungeon,
				Controller: trigger.Controller,
				Player:     trigger.Controller,
			},
			hasEvent:        true,
			ordinaryTrigger: true,
			dungeonRoom: opt.Val(game.DungeonRoomMark{
				Owner:    trigger.Controller,
				ObjectID: trigger.DungeonObjectID,
				Room:     trigger.Room,
				Final:    trigger.Final,
			}),
		})
	}
	g.PendingRoomAbilities = nil
	return pending
}

// drainPendingInitiativeVentures resolves every queued initiative venture into
// Undercity (CR 720). It runs where player choices are available (the trigger-
// gathering point), so a venture queued during combat damage can make branch
// choices. Each queued venture is resolved for a player who is still in the game.
func (e *Engine) drainPendingInitiativeVentures(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if len(g.PendingInitiativeVentures) == 0 {
		return
	}
	ventures := g.PendingInitiativeVentures
	g.PendingInitiativeVentures = nil
	for _, playerID := range ventures {
		if player, ok := playerByID(g, playerID); !ok || player.Eliminated {
			continue
		}
		e.ventureIntoUndercity(g, playerID, agents, log)
	}
}
