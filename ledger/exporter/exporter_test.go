package exporter

import (
	"testing"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/ledger/record"
	"github.com/insolar/insolar/ledger/storage/storagetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExporter_Export(t *testing.T) {
	ctx := inslogger.TestContext(t)
	db, clean := storagetest.TmpDB(ctx, t, "")
	defer clean()

	exporter := NewExporter(db)

	err := db.AddPulse(ctx, core.Pulse{PulseNumber: 0})
	require.NoError(t, err)

	_, err = db.SetRecord(ctx, 0, &record.GenesisRecord{})
	require.NoError(t, err)
	_, err = db.SetRecord(ctx, 0, &record.ObjectActivateRecord{
		ObjectStateRecord: record.ObjectStateRecord{},
		IsDelegate:        true,
	})
	require.NoError(t, err)

	data, err := exporter.Export(ctx, 0, 1)
	require.NoError(t, err)
	str := string(data)

	assert.True(t, len(data) > 0)

	_ = str
}
