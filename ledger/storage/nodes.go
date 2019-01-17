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

func SerializeNodes(nodes []Node) ([]byte, error) {
	var result []byte
	for _, node := range nodes {
		data, err := node.Serialize()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to serialize node %s", node.FID)
		}
		result = append(result, data...)
	}
	return result, nil
}

func DeserializeNodes(data []byte) ([]Node, error) {
	reader := bytes.NewReader(data)
	var result []Node
	for reader.Len() > 0 {
		node, err := DeserializeNode(reader)
		if err != nil {
			return nil, err
		}
		result = append(result, node)
	}
	return result, nil
}
