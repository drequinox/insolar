/*
 *    Copyright 2018 INS Ecosystem
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package artifactmanager

import (
	"github.com/pkg/errors"

	"github.com/insolar/insolar/ledger/index"
	"github.com/insolar/insolar/ledger/record"
	"github.com/insolar/insolar/ledger/storage"
)

// LedgerArtifactManager provides concrete API to storage for virtual processing module
type LedgerArtifactManager struct {
	storer   storage.LedgerStorer
	archPref []record.ArchType
}

func (m *LedgerArtifactManager) checkRequestRecord(requestID record.ID) error {
	// TODO: implement request check
	return nil
}

func (m *LedgerArtifactManager) getCodeRecord(codeID record.ID) (*record.CodeRecord, error) {
	rec, err := m.storer.GetRecord(codeID)
	if err != nil {
		return nil, errors.Wrap(err, "code record is not found")
	}
	codeRec, ok := rec.(*record.CodeRecord)
	if !ok {
		return nil, errors.New("provided reference is not a code reference")
	}
	return codeRec, nil
}

func (m *LedgerArtifactManager) getCodeRecordCode(codeID record.ID) ([]byte, error) {
	codeRec, err := m.getCodeRecord(codeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve code record")
	}
	code, err := codeRec.GetCode(m.archPref)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve code")
	}

	return code, nil
}

func (m *LedgerArtifactManager) storeRecord(rec record.Record) (record.Reference, error) {
	id, err := m.storer.SetRecord(rec)
	if err != nil {
		return record.Reference{}, errors.Wrap(err, "record store failed")
	}
	return record.Reference{Domain: rec.Domain(), Record: id}, nil
}

func (m *LedgerArtifactManager) getActiveClass(classID record.ID) (
	*record.ClassActivateRecord, *record.ClassAmendRecord, *index.ClassLifeline, error,
) {
	classRecord, err := m.storer.GetRecord(classID)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "class record is not found")
	}
	activateRec, ok := classRecord.(*record.ClassActivateRecord)
	if !ok {
		return nil, nil, nil, errors.New("provided reference is not a class record")
	}
	classIndex, isFound := m.storer.GetClassIndex(classID)
	if !isFound {
		return nil, nil, nil, errors.New("inconsistent class index")
	}
	latestClassRecord, err := m.storer.GetRecord(classIndex.LatestStateID)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "latest class record is not found")
	}
	if _, ok := latestClassRecord.(*record.DeactivationRecord); ok {
		return nil, nil, nil, errors.New("class is deactivated")
	}
	amendRecord, ok := latestClassRecord.(*record.ClassAmendRecord)
	if classID != classIndex.LatestStateID && !ok {
		return nil, nil, nil, errors.New("wrong index record")
	}

	return activateRec, amendRecord, classIndex, nil
}

func (m *LedgerArtifactManager) getActiveObject(objID record.ID) (
	*record.ObjectActivateRecord, *record.ObjectAmendRecord, *index.ObjectLifeline, error,
) {
	objRecord, err := m.storer.GetRecord(objID)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "object record is not found")
	}
	activateRec, ok := objRecord.(*record.ObjectActivateRecord)
	if !ok {
		return nil, nil, nil, errors.New("provided reference is not an object record")
	}

	objIndex, isFound := m.storer.GetObjectIndex(objID)
	if !isFound {
		return nil, nil, nil, errors.New("inconsistent object index")
	}
	latestObjRecord, err := m.storer.GetRecord(objIndex.LatestStateID)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "latest object record is not found")
	}
	if _, ok := latestObjRecord.(*record.DeactivationRecord); ok {
		return nil, nil, nil, errors.New("object is deactivated")
	}
	amendRecord, ok := latestObjRecord.(*record.ObjectAmendRecord)
	if objID != objIndex.LatestStateID && !ok {
		return nil, nil, nil, errors.New("wrong index record")
	}

	return activateRec, amendRecord, objIndex, nil
}

func (m *LedgerArtifactManager) SetArchPref(pref []record.ArchType) {
	m.archPref = pref
}

// DeployCode deploys new code to storage (CodeRecord).
func (m *LedgerArtifactManager) DeployCode(
	requestRef record.Reference, codeMap map[record.ArchType][]byte,
) (record.Reference, error) {
	err := m.checkRequestRecord(requestRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	rec := record.CodeRecord{
		StorageRecord: record.StorageRecord{
			StatefulResult: record.StatefulResult{
				ResultRecord: record.ResultRecord{
					RequestRecord: requestRef,
				},
			},
		},
		TargetedCode: codeMap,
	}
	return m.storeRecord(&rec)
}

// ActivateClass activates class from given code (ClassActivateRecord).
func (m *LedgerArtifactManager) ActivateClass(
	requestRef, codeRef record.Reference, memory record.Memory,
) (record.Reference, error) {
	err := m.checkRequestRecord(requestRef.Record)
	if err != nil {
		return record.Reference{}, err
	}
	_, err = m.getCodeRecord(codeRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	rec := record.ClassActivateRecord{
		ActivationRecord: record.ActivationRecord{
			StatefulResult: record.StatefulResult{
				ResultRecord: record.ResultRecord{
					RequestRecord: requestRef,
				},
			},
		},
		CodeRecord:    codeRef,
		DefaultMemory: memory,
	}
	classRef, err := m.storeRecord(&rec)
	if err != nil {
		return record.Reference{}, err
	}
	err = m.storer.SetClassIndex(classRef.Record, &index.ClassLifeline{
		LatestStateID: classRef.Record,
	})
	if err != nil {
		return record.Reference{}, errors.Wrap(err, "failed to store lifeline index")
	}

	return classRef, nil
}

// DeactivateClass deactivates class (DeactivationRecord)
func (m *LedgerArtifactManager) DeactivateClass(
	requestRef, classRef record.Reference,
) (record.Reference, error) {
	err := m.checkRequestRecord(requestRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	_, _, classIndex, err := m.getActiveClass(classRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	rec := record.DeactivationRecord{
		AmendRecord: record.AmendRecord{
			StatefulResult: record.StatefulResult{
				ResultRecord: record.ResultRecord{
					RequestRecord: requestRef,
				},
			},
			HeadRecord: classRef,
			AmendedRecord: record.Reference{
				Domain: classRef.Domain,
				Record: classIndex.LatestStateID,
			},
		},
	}
	deactivationRef, err := m.storeRecord(&rec)
	if err != nil {
		return record.Reference{}, errors.New("failed to store deactivation record")
	}
	classIndex.LatestStateID = deactivationRef.Record
	err = m.storer.SetClassIndex(classRef.Record, classIndex)
	if err != nil {
		// TODO: add transaction
		return record.Reference{}, errors.New("failed to store lifeline index")
	}

	return deactivationRef, nil
}

// UpdateClass allows to change class code etc. (ClassAmendRecord).
func (m *LedgerArtifactManager) UpdateClass(
	requestRef, classRef, codeRef record.Reference, migrationRefs []record.Reference,
) (record.Reference, error) {
	err := m.checkRequestRecord(requestRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	_, _, classIndex, err := m.getActiveClass(classRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	_, err = m.getCodeRecord(codeRef.Record)
	if err != nil {
		return record.Reference{}, err
	}
	for _, migrationRef := range migrationRefs {
		_, err = m.getCodeRecord(migrationRef.Record)
		if err != nil {
			return record.Reference{}, errors.Wrap(err, "invalid migrations")
		}
	}

	rec := record.ClassAmendRecord{
		AmendRecord: record.AmendRecord{
			StatefulResult: record.StatefulResult{
				ResultRecord: record.ResultRecord{
					RequestRecord: requestRef,
				},
			},
			HeadRecord: classRef,
			AmendedRecord: record.Reference{
				Domain: classRef.Domain,
				Record: classIndex.LatestStateID,
			},
		},
		NewCode:    codeRef,
		Migrations: migrationRefs,
	}

	amendRef, err := m.storeRecord(&rec)
	if err != nil {
		return record.Reference{}, errors.New("failed to store amend record")
	}
	classIndex.LatestStateID = amendRef.Record
	classIndex.AmendIDs = append(classIndex.AmendIDs, amendRef.Record)
	err = m.storer.SetClassIndex(classRef.Record, classIndex)
	if err != nil {
		// TODO: add transaction
		return record.Reference{}, errors.New("failed to store lifeline index")
	}

	return amendRef, nil
}

// ActivateObj creates and activates new object from given class (ObjectActivateRecord).
func (m *LedgerArtifactManager) ActivateObj(
	requestRef, classRef record.Reference, memory record.Memory,
) (record.Reference, error) {
	err := m.checkRequestRecord(requestRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	_, _, _, err = m.getActiveClass(classRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	rec := record.ObjectActivateRecord{
		ActivationRecord: record.ActivationRecord{
			StatefulResult: record.StatefulResult{
				ResultRecord: record.ResultRecord{
					RequestRecord: requestRef,
				},
			},
		},
		ClassActivateRecord: classRef,
		Memory:              memory,
	}

	objRef, err := m.storeRecord(&rec)
	if err != nil {
		return record.Reference{}, err
	}
	err = m.storer.SetObjectIndex(objRef.Record, &index.ObjectLifeline{
		ClassID:       classRef.Record,
		LatestStateID: objRef.Record,
	})
	if err != nil {
		return record.Reference{}, errors.Wrap(err, "failed to store lifeline index")
	}

	return objRef, nil
}

// DeactivateObj deactivates object (DeactivationRecord).
func (m *LedgerArtifactManager) DeactivateObj(requestRef, objRef record.Reference) (record.Reference, error) {
	err := m.checkRequestRecord(requestRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	_, _, objIndex, err := m.getActiveObject(objRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	rec := record.DeactivationRecord{
		AmendRecord: record.AmendRecord{
			StatefulResult: record.StatefulResult{
				ResultRecord: record.ResultRecord{
					RequestRecord: requestRef,
				},
			},
			HeadRecord: objRef,
			AmendedRecord: record.Reference{
				Domain: objRef.Domain,
				Record: objIndex.LatestStateID,
			},
		},
	}
	deactivationRef, err := m.storeRecord(&rec)
	if err != nil {
		return record.Reference{}, errors.New("failed to store deactivation record")
	}
	objIndex.LatestStateID = deactivationRef.Record
	err = m.storer.SetObjectIndex(objRef.Record, objIndex)
	if err != nil {
		// TODO: add transaction
		return record.Reference{}, errors.New("failed to store lifeline index")
	}
	return deactivationRef, nil
}

// UpdateObj allows to change object state (ObjectAmendRecord).
func (m *LedgerArtifactManager) UpdateObj(
	requestRef, objRef record.Reference, memory record.Memory,
) (record.Reference, error) {
	err := m.checkRequestRecord(requestRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	_, _, objIndex, err := m.getActiveObject(objRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	rec := record.ObjectAmendRecord{
		AmendRecord: record.AmendRecord{
			StatefulResult: record.StatefulResult{
				ResultRecord: record.ResultRecord{
					RequestRecord: requestRef,
				},
			},
			HeadRecord: objRef,
			AmendedRecord: record.Reference{
				Domain: objRef.Domain,
				Record: objIndex.LatestStateID,
			},
		},
		NewMemory: memory,
	}

	amendRef, err := m.storeRecord(&rec)
	if err != nil {
		return record.Reference{}, errors.New("failed to store amend record")
	}
	objIndex.LatestStateID = amendRef.Record
	objIndex.AppendIDs = []record.ID{}
	err = m.storer.SetObjectIndex(objRef.Record, objIndex)
	if err != nil {
		// TODO: add transaction
		return record.Reference{}, errors.New("failed to store lifeline index")
	}
	return amendRef, nil
}

func (m *LedgerArtifactManager) AppendObjDelegate(
	requestRef, objRef record.Reference, memory record.Memory,
) (record.Reference, error) {
	err := m.checkRequestRecord(requestRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	_, _, objIndex, err := m.getActiveObject(objRef.Record)
	if err != nil {
		return record.Reference{}, err
	}

	rec := record.ObjectAppendRecord{
		AmendRecord: record.AmendRecord{
			StatefulResult: record.StatefulResult{
				ResultRecord: record.ResultRecord{
					RequestRecord: requestRef,
				},
			},
			HeadRecord: objRef,
			AmendedRecord: record.Reference{
				Domain: objRef.Domain,
				Record: objIndex.LatestStateID,
			},
		},
		AppendMemory: memory,
	}

	appendRef, err := m.storeRecord(&rec)
	if err != nil {
		return record.Reference{}, errors.New("failed to store append record")
	}
	objIndex.AppendIDs = append(objIndex.AppendIDs, appendRef.Record)
	err = m.storer.SetObjectIndex(objRef.Record, objIndex)
	if err != nil {
		// TODO: add transaction
		return record.Reference{}, errors.New("failed to store lifeline index")
	}
	return appendRef, nil
}

func (m *LedgerArtifactManager) GetExactObj(
	classState, objectState record.Reference,
) ([]byte, record.Memory, error) {
	classRec, err := m.storer.GetRecord(classState.Record)
	if err != nil {
		return nil, nil, errors.Wrap(err, "class record not found")
	}
	var codeID record.ID
	var classHeadRef record.Reference
	switch rec := classRec.(type) {
	case *record.ClassActivateRecord:
		codeID = rec.CodeRecord.Record
		classHeadRef = classState
	case *record.ClassAmendRecord:
		codeID = rec.NewCode.Record
		classHeadRef = rec.HeadRecord
	default:
		return nil, nil, errors.New("wrong class reference")
	}
	code, err := m.getCodeRecordCode(codeID)
	if err != nil {
		return nil, nil, err
	}

	objectRec, err := m.storer.GetRecord(objectState.Record)
	if err != nil {
		return nil, nil, errors.Wrap(err, "object record not found")
	}
	var memory record.Memory
	var objectHeadRef record.Reference
	switch rec := objectRec.(type) {
	case *record.ObjectActivateRecord:
		memory = rec.Memory
		objectHeadRef = objectState
	case *record.ObjectAmendRecord:
		memory = rec.NewMemory
		objectHeadRef = rec.HeadRecord
	default:
		return nil, nil, errors.New("wrong object reference")
	}
	objectIndex, ok := m.storer.GetObjectIndex(objectHeadRef.Record)
	if !ok {
		return nil, nil, errors.New("object index not found")
	}

	if objectIndex.ClassID != classHeadRef.Record {
		return nil, nil, errors.New("the object does not belong to the class")
	}

	return code, memory, nil
}

func (m *LedgerArtifactManager) GetLatestObj(
	objectRef, storedClassState, storedObjState record.Reference,
) (*ClassDescriptor, *ObjectDescriptor, error) {
	var (
		class  *ClassDescriptor
		object *ObjectDescriptor
	)

	objActivateRec, objStateRec, objIndex, err := m.getActiveObject(objectRef.Record)
	if err != nil {
		return nil, nil, err
	}
	classActivateRec, classStateRec, classIndex, err := m.getActiveClass(objIndex.ClassID)
	if err != nil {
		return nil, nil, err
	}

	if storedClassState.Record != classIndex.LatestStateID {
		class = &ClassDescriptor{
			StateRef: record.Reference{
				Domain: storedClassState.Domain,
				Record: classIndex.LatestStateID,
			},

			manager:           m,
			fromState:         storedClassState,
			activateRecord:    classActivateRec,
			latestAmendRecord: classStateRec,
			lifelineIndex:     classIndex,
		}
	}

	if storedObjState.Record != objIndex.LatestStateID {
		object = &ObjectDescriptor{
			StateRef: record.Reference{
				Domain: storedObjState.Domain,
				Record: objIndex.LatestStateID,
			},

			manager:           m,
			activateRecord:    objActivateRec,
			latestAmendRecord: objStateRec,
			lifelineIndex:     objIndex,
		}
	}

	return class, object, nil
}