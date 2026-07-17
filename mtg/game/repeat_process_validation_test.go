package game

import "testing"

func TestValidateConditionalRepeatProcess(t *testing.T) {
	t.Parallel()
	const key = ResultKey("continue")
	validBody := Mode{Sequence: []Instruction{{
		Primitive:     Draw{Player: ControllerReference(), Amount: Fixed(1)},
		PublishResult: key,
	}}}.Ability()
	for _, tc := range []struct {
		name    string
		repeat  RepeatProcess
		wantErr bool
	}{
		{
			name:   "published continuation",
			repeat: RepeatProcess{Body: validBody, ContinueResult: key},
		},
		{
			name:    "missing publication",
			repeat:  RepeatProcess{Body: validBody, ContinueResult: "other"},
			wantErr: true,
		},
		{
			name:    "conditional and bounded",
			repeat:  RepeatProcess{Times: Fixed(2), Body: validBody, ContinueResult: key},
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.repeat.validatePrimitive(nil, true)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validatePrimitive() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
