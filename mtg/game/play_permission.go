package game

import "github.com/natefinch/council4/mtg/game/types"

// PlayPermission identifies one source-incarnation/type permission announced
// before playing a card. The zero value selects an independent permission.
type PlayPermission struct {
	SourceID ObjectID
	Key      LinkedKey
	Type     types.Card
}

// PlayPermissionUse separates players' allowances even if control of the same
// source changes during a turn. The map containing these keys resets each turn.
type PlayPermissionUse struct {
	Permission PlayPermission
	Player     PlayerID
}
