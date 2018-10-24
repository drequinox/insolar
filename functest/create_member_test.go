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

package functest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMember(t *testing.T) {
	result, err := signedRequest(&root, "CreateMember", "Member", "000")
	assert.NoError(t, err)
	ref, ok := result.(string)
	assert.True(t, ok)
	assert.NotEqual(t, "", ref)
}

func TestCreateMemberWrongNameType(t *testing.T) {
	_, err := signedRequest(&root, "CreateMember", 111, "000")
	assert.EqualError(t, err, "unexpected EOF")
}

func TestCreateMemberWrongKeyType(t *testing.T) {
	_, err := signedRequest(&root, "CreateMember", "Member", 111)
	assert.EqualError(t, err, "EOF")
}

// no error
func _TestCreateMemberOneParameter(t *testing.T) {
	_, err := signedRequest(&root, "CreateMember", "text")
	assert.Error(t, err)
}

func TestCreateMemberOneParameterOtherType(t *testing.T) {
	_, err := signedRequest(&root, "CreateMember", 111)
	assert.EqualError(t, err, "EOF")
}

func TestCreateMembersWithSameName(t *testing.T) {
	firstMemberRef, err := signedRequest(&root, "CreateMember", "Member", "000")
	assert.NoError(t, err)
	secondMemberRef, err := signedRequest(&root, "CreateMember", "Member", "000")
	assert.NoError(t, err)

	assert.NotEqual(t, firstMemberRef, secondMemberRef)
}
