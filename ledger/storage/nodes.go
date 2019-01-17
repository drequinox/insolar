/*
 *    Copyright 2018 Insolar
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

package storage

import (
	"bytes"
	"crypto"
	"encoding/binary"
	"io"
	"strconv"

	"github.com/insolar/insolar/core"
	"github.com/pkg/errors"
)

type Node struct {
	FID   core.RecordRef
	FRole core.StaticRole
}

func (Node) GetGlobuleID() core.GlobuleID {
	panic("implement me")
}

func (n Node) ID() core.RecordRef {
	return n.FID
}

func (Node) PhysicalAddress() string {
	panic("implement me")
}

func (Node) PublicKey() crypto.PublicKey {
	panic("implement me")
}

func (n Node) Role() core.StaticRole {
	return n.FRole
}

func (Node) ShortID() core.ShortNodeID {
	panic("implement me")
}

func (Node) Version() string {
	panic("implement me")
}

func (n Node) Serialize() ([]byte, error) {
	buf := make([]byte, 0, core.RecordRefSize+strconv.IntSize)
	result := bytes.NewBuffer(buf)
	if _, err := result.Write(n.FID[:]); err != nil {
		return nil, errors.Wrap(err, "Failed to serialize node ID")
	}
	if err := binary.Write(result, binary.BigEndian, int32(n.FRole)); err != nil {
		return nil, errors.Wrap(err, "Failed to serialize node role")
	}
	return result.Bytes(), nil
}

func DeserializeNode(reader io.Reader) (Node, error) {
	result := Node{}
	if err := binary.Read(reader, binary.BigEndian, result.FID[:]); err != nil {
		return result, errors.Wrap(err, "Failed to deserialize node ID")
	}
	var role int32
	if err := binary.Read(reader, binary.BigEndian, &role); err != nil {
		return result, errors.Wrap(err, "Failed to deserialize node role")
	}
	result.FRole = core.StaticRole(role)
	return result, nil
}
