package orm_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/orm"
	"github.com/cosmos/cosmos-sdk/orm/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Testing ORM with arbitrary address length
const addrLen = 10

func TestExportTableData(t *testing.T) {
	storeKey := sdk.NewKVStoreKey("test")
	const prefix = iota
	table := orm.NewTableBuilder(prefix, storeKey, &testdata.GroupMetadata{}, orm.FixLengthIndexKeys(1)).Build()

	ctx := orm.NewMockContext()
	testRecordsNum := 2
	testRecords := make([]testdata.GroupMetadata, testRecordsNum)
	for i := 1; i <= testRecordsNum; i++ {
		myAddr := sdk.AccAddress(bytes.Repeat([]byte{byte(i)}, addrLen))
		g := testdata.GroupMetadata{
			Description: fmt.Sprintf("my test %d", i),
			Admin:       myAddr,
		}
		err := table.Create(ctx, []byte{byte(i)}, &g)
		require.NoError(t, err)
		testRecords[i-1] = g
	}

	jsonModels, _, err := orm.ExportTableData(ctx, table)
	require.NoError(t, err)
	exp := `[
	{
	"key" : "AQ==",
	"value": {"admin":"cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpjnp7du", "description":"my test 1"}
	},
	{
	"key":"Ag==", 
	"value": {"admin":"cosmos1qgpqyqszqgpqyqszqgpqyqszqgpqyqszrh8mx2", "description":"my test 2"}
	}
]`
	assert.JSONEq(t, exp, string(jsonModels))
}

func TestImportTableData(t *testing.T) {
	storeKey := sdk.NewKVStoreKey("test")
	const prefix = iota
	table := orm.NewTableBuilder(prefix, storeKey, &testdata.GroupMetadata{}, orm.FixLengthIndexKeys(1)).Build()

	ctx := orm.NewMockContext()

	jsonModels := `[
	{
	"key" : "AQ==",
	"value": {"admin":"cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpjnp7du", "description":"my test 1"}
	},
	{
	"key":"Ag==", 
	"value": {"admin":"cosmos1qgpqyqszqgpqyqszqgpqyqszqgpqyqszrh8mx2", "description":"my test 2"}
	}
]`
	// when
	err := orm.ImportTableData(ctx, table, []byte(jsonModels), 0)
	require.NoError(t, err)

	// then
	for i := 1; i < 3; i++ {
		var loaded testdata.GroupMetadata
		err := table.GetOne(ctx, []byte{byte(i)}, &loaded)
		require.NoError(t, err)

		exp := testdata.GroupMetadata{
			Description: fmt.Sprintf("my test %d", i),
			Admin:       sdk.AccAddress(bytes.Repeat([]byte{byte(i)}, addrLen)),
		}
		require.Equal(t, exp, loaded)
	}

}
