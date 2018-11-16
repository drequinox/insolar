package exporter

import (
	"context"
	"encoding/json"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/ledger/record"
	"github.com/insolar/insolar/ledger/storage"
	"github.com/jbenet/go-base58"
)

type Exporter struct {
	db *storage.DB
}

func NewExporter(db *storage.DB) *Exporter {
	return &Exporter{db: db}
}

type recordData struct {
	Type string
	Data record.Record
}

type recordsData map[string]recordData

type pulseData struct {
	Records recordsData
	Pulse   core.Pulse
}

func (e *Exporter) Export(ctx context.Context, fromPulse core.PulseNumber, size int) ([]byte, error) {
	results := make(map[core.PulseNumber]pulseData)

	current := &fromPulse
	for current != nil {
		pulse, err := e.db.GetPulse(ctx, *current)
		if err != nil {
			return nil, err
		}
		data, err := e.exportPulse(ctx, &pulse.Pulse)
		if err != nil {
			return nil, err
		}
		results[fromPulse] = *data

		current = pulse.Next
	}

	return json.Marshal(results)
}

func (e *Exporter) exportPulse(ctx context.Context, pulse *core.Pulse) (*pulseData, error) {
	records := recordsData{}
	err := e.db.IterateRecords(ctx, pulse.PulseNumber, func(id core.RecordID, rec record.Record) error {
		records[string(base58.Encode(id[:]))] = recordData{
			Type: rec.Type().String(),
			Data: rec,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	data := pulseData{
		Records: records,
		Pulse:   *pulse,
	}

	return &data, nil
}
