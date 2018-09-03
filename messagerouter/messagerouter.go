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

package messagerouter

import (
	"bytes"
	"encoding/gob"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/network/servicenetwork"
)

const deliverRPCMethodName = "MessageRouter.Deliver"

// MessageRouter is component that routes application logic requests,
// e.g. glue between network and logic runner
type MessageRouter struct {
	LogicRunner LogicRunner
	service     *servicenetwork.ServiceNetwork
}

// LogicRunner is an interface that should satisfy logic executor
type LogicRunner interface {
	Execute(ref core.RecordRef, method string, args []byte) (data []byte, result []byte, err error)
}

// Response to a `Message`
type Response struct {
	Data   []byte
	Result []byte
	Error  error
}

// New is a `MessageRouter` constructor, takes an executor object
// that satisfies `LogicRunner` interface
func New(lr LogicRunner, service *servicenetwork.ServiceNetwork) (*MessageRouter, error) {
	mr := &MessageRouter{lr, service}
	mr.service.RemoteProcedureRegister(deliverRPCMethodName, mr.deliver)
	return mr, nil
}

// Route a `Message` and get a `Response` or error from remote host
func (r *MessageRouter) Route(msg *core.Message) (response Response, err error) {
	res, err := r.service.SendMessage(deliverRPCMethodName, msg)
	if err != nil {
		return response, err
	}

	return DeserializeResponse(res)
}

// Deliver method calls LogicRunner.Execute on local host
// this method is registered as RPC stub
func (r *MessageRouter) deliver(args [][]byte) (result []byte, err error) {

	msg, err := DeserializeMessage(args[0]) // TODO: check empty args
	if err != nil {
		return nil, err
	}

	data, res, err := r.LogicRunner.Execute(msg.Reference, msg.Method, msg.Arguments)
	return Serialize(Response{data, res, err})
}

// Serialize converts Message or Response to byte slice.
func Serialize(value interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(value)
	if err != nil {
		return nil, err
	}
	res := buffer.Bytes()
	return res, err
}

// DeserializeMessage reads packet from byte slice.
func DeserializeMessage(data []byte) (msg core.Message, err error) {
	err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&msg)
	return msg, err
}

// DeserializeResponse reads response from byte slice.
func DeserializeResponse(data []byte) (res Response, err error) {
	err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&res)
	return res, err
}

func init() {
	gob.Register(&core.Message{})
	gob.Register(&Response{})
}
