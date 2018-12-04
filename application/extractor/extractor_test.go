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

package extractor

import (
	"testing"

	"github.com/insolar/insolar/core"
	"github.com/stretchr/testify/require"
)

// base.go
func TestStringResponse(t *testing.T) {
	testValue := "test_string"

	data, err := core.Serialize([]interface{}{testValue, nil})
	require.NoError(t, err)

	result, err := stringResponse(data)

	require.NoError(t, err)
	require.Equal(t, testValue, result)
}
