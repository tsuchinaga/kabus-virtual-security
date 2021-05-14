package virtual_security

import (
	"reflect"
	"testing"
)

func Test_NewUUIDGenerator(t *testing.T) {
	want := &uuidGenerator{}
	got := NewUUIDGenerator()
	t.Parallel()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_uuidGenerator_Generate(t *testing.T) {
	l := 1000
	uuidGenerator := &uuidGenerator{}
	uuids := make([]string, l)
	for i := 0; i < l; i++ {
		got := uuidGenerator.Generate()
		for j := 0; j < i; j++ {
			if got == uuids[j] {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), uuids, got)
			}
		}
		uuids[i] = got
	}
}
