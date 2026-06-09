package game

import "github.com/natefinch/council4/opt"

func objectReferenceForTest(kind ObjectReferenceKind, targetIndex int, linkID string) ObjectReference {
	return ObjectReference{kind: kind, targetIndex: targetIndex, linkID: linkID}
}

func playerReferenceForTest(kind PlayerReferenceKind, targetIndex int, object opt.V[ObjectReference]) PlayerReference {
	return PlayerReference{kind: kind, targetIndex: targetIndex, object: object}
}
