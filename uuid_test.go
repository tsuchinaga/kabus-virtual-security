package virtual_security

import (
	"reflect"
	"testing"
)

type testUUIDGenerator struct {
	i          int
	generator1 []string
}

func (t *testUUIDGenerator) generate() string {
	defer func() { t.i++ }()
	return t.generator1[t.i%len(t.generator1)]
}

func Test_newUUIDGenerator(t *testing.T) {
	want := &uuidGenerator{}
	got := newUUIDGenerator()
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
		got := uuidGenerator.generate()
		for j := 0; j < i; j++ {
			if got == uuids[j] {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), uuids, got)
			}
		}
		uuids[i] = got
	}
}
