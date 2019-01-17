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
	"testing"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeserializeNode(t *testing.T) {
	node := Node{
		FID:   testutils.RandomRef(),
		FRole: core.StaticRoleVirtual,
	}

	data, err := node.Serialize()
	require.NoError(t, err)
	node2, err := DeserializeNode(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, node, node2)
}

func TestDeserializeNodes(t *testing.T) {
	node1 := Node{
		FID:   testutils.RandomRef(),
		FRole: core.StaticRoleVirtual,
	}
	node2 := Node{
		FID:   testutils.RandomRef(),
		FRole: core.StaticRoleHeavyMaterial,
	}
	node3 := Node{
		FID:   testutils.RandomRef(),
		FRole: core.StaticRoleLightMaterial,
	}
	nodes := []Node{node1, node2, node3}
	data, err := SerializeNodes(nodes)
	require.NoError(t, err)
	nodes2, err := DeserializeNodes(data)
	require.NoError(t, err)
	assert.Equal(t, nodes, nodes2)
}
