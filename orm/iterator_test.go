package orm_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/orm"
	"github.com/cosmos/cosmos-sdk/orm/testdata"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadAll(t *testing.T) {
	specs := map[string]struct {
		srcIT     orm.Iterator
		destSlice func() orm.ModelSlicePtr
		expErr    *errors.Error
		expIDs    []orm.RowID
		expResult orm.ModelSlicePtr
	}{
		"all good with object slice": {
			srcIT: mockIter(orm.EncodeSequence(1), &testdata.GroupMetadata{Description: "test"}),
			destSlice: func() orm.ModelSlicePtr {
				x := make([]testdata.GroupMetadata, 1)
				return &x
			},
			expIDs:    []orm.RowID{orm.EncodeSequence(1)},
			expResult: &[]testdata.GroupMetadata{{Description: "test"}},
		},
		"all good with pointer slice": {
			srcIT: mockIter(orm.EncodeSequence(1), &testdata.GroupMetadata{Description: "test"}),
			destSlice: func() orm.ModelSlicePtr {
				x := make([]*testdata.GroupMetadata, 1)
				return &x
			},
			expIDs:    []orm.RowID{orm.EncodeSequence(1)},
			expResult: &[]*testdata.GroupMetadata{{Description: "test"}},
		},
		"dest slice empty": {
			srcIT: mockIter(orm.EncodeSequence(1), &testdata.GroupMetadata{}),
			destSlice: func() orm.ModelSlicePtr {
				x := make([]testdata.GroupMetadata, 0)
				return &x
			},
			expIDs:    []orm.RowID{orm.EncodeSequence(1)},
			expResult: &[]testdata.GroupMetadata{{}},
		},
		"dest pointer with nil value": {
			srcIT: mockIter(orm.EncodeSequence(1), &testdata.GroupMetadata{}),
			destSlice: func() orm.ModelSlicePtr {
				return (*[]testdata.GroupMetadata)(nil)
			},
			expErr: orm.ErrArgument,
		},
		"iterator is nil": {
			srcIT:     nil,
			destSlice: func() orm.ModelSlicePtr { return new([]testdata.GroupMetadata) },
			expErr:    orm.ErrArgument,
		},
		"dest slice is nil": {
			srcIT:     noopIter(),
			destSlice: func() orm.ModelSlicePtr { return nil },
			expErr:    orm.ErrArgument,
		},
		"dest slice is not a pointer": {
			srcIT:     orm.IteratorFunc(nil),
			destSlice: func() orm.ModelSlicePtr { return make([]testdata.GroupMetadata, 1) },
			expErr:    orm.ErrArgument,
		},
		"error on loadNext is returned": {
			srcIT: orm.NewInvalidIterator(),
			destSlice: func() orm.ModelSlicePtr {
				x := make([]testdata.GroupMetadata, 1)
				return &x
			},
			expErr: orm.ErrIteratorInvalid,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			loaded := spec.destSlice()
			ids, err := orm.ReadAll(spec.srcIT, loaded)
			require.True(t, spec.expErr.Is(err), "expected %s but got %s", spec.expErr, err)
			assert.Equal(t, spec.expIDs, ids)
			if err == nil {
				assert.Equal(t, spec.expResult, loaded)
			}
		})
	}
}

func TestLimitedIterator(t *testing.T) {
	sliceIter := func(s ...string) orm.Iterator {
		var pos int
		return orm.IteratorFunc(func(dest orm.Persistent) (orm.RowID, error) {
			if pos == len(s) {
				return nil, orm.ErrIteratorDone
			}
			v := s[pos]

			*dest.(*persistentString) = persistentString(v)
			pos++
			return []byte(v), nil
		})
	}
	specs := map[string]struct {
		src orm.Iterator
		exp []persistentString
	}{
		"all from range with max > length": {
			src: orm.LimitIterator(sliceIter("a", "b", "c"), 4),
			exp: []persistentString{"a", "b", "c"},
		},
		"up to max": {
			src: orm.LimitIterator(sliceIter("a", "b", "c"), 2),
			exp: []persistentString{"a", "b"},
		},
		"none when max = 0": {
			src: orm.LimitIterator(sliceIter("a", "b", "c"), 0),
			exp: []persistentString{},
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			var loaded []persistentString
			_, err := orm.ReadAll(spec.src, &loaded)
			require.NoError(t, err)
			assert.EqualValues(t, spec.exp, loaded)
		})
	}
}

// mockIter amino encodes + decodes value object.
func mockIter(rowID orm.RowID, val orm.Persistent) orm.Iterator {
	b, err := val.Marshal()
	if err != nil {
		panic(err)
	}
	return orm.NewSingleValueIterator(rowID, b)
}

func noopIter() orm.Iterator {
	return orm.IteratorFunc(func(dest orm.Persistent) (orm.RowID, error) {
		return nil, nil
	})
}

type persistentString string

func (p persistentString) Marshal() ([]byte, error) {
	return []byte(p), nil
}

func (p *persistentString) Unmarshal(b []byte) error {
	s := persistentString(string(b))
	p = &s
	return nil
}

func (p persistentString) ValidateBasic() error {
	return nil
}
