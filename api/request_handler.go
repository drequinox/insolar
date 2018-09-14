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

package api

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/messagerouter/message"
	"github.com/jbenet/go-base58"
	"github.com/pkg/errors"
)

func makeRootDomainReference() core.RecordRef {
	const ref = "1111111-1111111-11111111-1111111"
	return core.String2Ref(base58.Encode([]byte(ref)))
}

var rootDomainReference = makeRootDomainReference()

func extractCreateMemberResponse(data []byte) (*string, error) {
	var typeHolder string
	refOrig, err := UnMarshalResponse(data, []interface{}{typeHolder})
	if err != nil {
		return nil, errors.Wrap(err, "[ extractCreateMemberResponse ]")
	}

	reference, ok := refOrig[0].(string)
	if !ok {
		msg := fmt.Sprintf("Can't cast response to string. orig: %T", refOrig)
		return nil, errors.New(msg)
	}

	return &reference, nil
}

// RequestHandler encapsulate processing of request
type RequestHandler struct {
	qid                 string
	params              *Params
	messageRouter       core.MessageRouter
	rootDomainReference core.RecordRef
}

// Creates new query handler
func NewRequestHandler(params *Params, router core.MessageRouter) *RequestHandler {
	return &RequestHandler{
		qid:                 params.QID,
		params:              params,
		messageRouter:       router,
		rootDomainReference: rootDomainReference,
	}
}

func (rh *RequestHandler) routeCall(ref core.RecordRef, method string, args core.Arguments) (*core.Response, error) {
	if rh.messageRouter == nil {
		return nil, errors.New("[ RouteCall ] message router was not set during initialization")
	}

	msg := &message.CallMethodMessage{
		ObjectRef: ref,
		Method:    method,
		Arguments: args,
	}

	res, err := rh.messageRouter.Route(msg)
	if err != nil {
		return nil, errors.Wrap(err, "[ RouteCall ] couldn't route message")
	}
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "[ RouteCall ] couldn't route message (error in response)")
	}

	return &res, nil
}

// Process CreateMember query type
func (rh *RequestHandler) ProcessCreateMember() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if len(rh.params.Name) == 0 {
		return nil, errors.New("field 'name' is required")
	}

	routResult, err := rh.sendRequest("CreateMember", []interface{}{rh.params.Name})
	if err != nil {
		return nil, errors.Wrap(err, "[ ProcessCreateMember ]")
	}

	memberRef, err := extractCreateMemberResponse(routResult.Result)
	if err != nil {
		return nil, errors.Wrap(err, "[ ProcessCreateMember ]")
	}

	result["reference"] = memberRef

	return result, nil
}

func extractGetBalanceResponse(data []byte) (uint, error) {
	var typeHolder uint
	dataUnmarsh, err := UnMarshalResponse(data, []interface{}{typeHolder})
	if err != nil {
		return 0, errors.Wrap(err, "[ extractGetBalanceResponse ]")
	}

	balance, ok := dataUnmarsh[0].(uint)
	if !ok {
		msg := fmt.Sprintf("Can't cast response to uint. orig: %s", reflect.TypeOf(dataUnmarsh[0]).String())
		return 0, errors.New(msg)
	}

	return balance, nil
}

func (rh *RequestHandler) sendRequest(method string, argsIn []interface{}) (*core.Response, error) {
	args, err := MarshalArgs(argsIn...)
	if err != nil {
		return nil, errors.Wrap(err, "[ SendRequest ]")
	}

	routResult, err := rh.routeCall(rh.rootDomainReference, method, args)
	if err != nil {
		return nil, errors.Wrap(err, "[ SendRequest ]")
	}

	return routResult, nil
}

// Process GetBalance query type
func (rh *RequestHandler) ProcessGetBalance() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["currency"] = "RUB"

	if len(rh.params.Reference) == 0 {
		return nil, errors.New("field 'reference' is required")
	}

	routResult, err := rh.sendRequest("GetBalance", []interface{}{rh.params.Reference})
	if err != nil {
		return nil, errors.Wrap(err, "[ ProcessGetBalance ]")
	}

	amount, err := extractGetBalanceResponse(routResult.Result)
	if err != nil {
		return nil, errors.Wrap(err, "[ ProcessGetBalance ]")
	}

	result["amount"] = amount

	return result, nil
}

func extractSendMoneyResponse(data []byte) (bool, error) {
	var typeHolder bool
	dataUnmarsh, err := UnMarshalResponse(data, []interface{}{typeHolder})
	if err != nil {
		return false, errors.Wrap(err, "[ extractSendMoneyResponse ]")
	}

	isSent, ok := dataUnmarsh[0].(bool)
	if !ok {
		msg := fmt.Sprintf("Can't cast response to bool. orig: %T", dataUnmarsh)
		return false, errors.New(msg)
	}

	return isSent, nil
}

// Process SendMoney query type
func (rh *RequestHandler) ProcessSendMoney() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if len(rh.params.From) == 0 {
		return nil, errors.New("field 'from' is required")
	}

	if len(rh.params.To) == 0 {
		return nil, errors.New("field 'from' is required")
	}
	if rh.params.Amount == 0 {
		return nil, errors.New("field 'amount' is required")
	}

	routResult, err := rh.sendRequest("SendMoney", []interface{}{rh.params.From, rh.params.To, rh.params.Amount})
	if err != nil {
		return nil, errors.Wrap(err, "[ ProcessSendMoney ]")
	}

	isSent, err := extractSendMoneyResponse(routResult.Result)
	if err != nil {
		return nil, errors.Wrap(err, "[ ProcessSendMoney ]")
	}

	result["success"] = isSent

	return result, nil
}

func extractDumpAllUsersResponse(data []byte) ([]byte, error) {
	var typeHolder []byte
	dataUnmarsh, err := UnMarshalResponse(data, []interface{}{typeHolder})
	if err != nil {
		return nil, errors.Wrap(err, "[ extractDumpAllUsersResponse ]")
	}

	dumpJSON, ok := dataUnmarsh[0].([]byte)
	if !ok {
		msg := fmt.Sprintf("Can't cast response to []byte. orig: %s", reflect.TypeOf(dataUnmarsh[0]))
		return nil, errors.New(msg)
	}

	return dumpJSON, nil
}

// Process Dump users query type
func (rh *RequestHandler) ProcessDumpUsers(all bool) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var err error
	var routResult *core.Response
	if all {
		routResult, err = rh.sendRequest("DumpAllUsers", []interface{}{})
	} else {
		if len(rh.params.Name) == 0 {
			return nil, errors.New("field 'name' is required")
		}
		routResult, err = rh.sendRequest("DumpUserInfo", []interface{}{rh.params.Name})
	}

	if err != nil {
		return nil, errors.Wrap(err, "[ ProcessDumpUsers ]")
	}

	serJSONDump, err := extractDumpAllUsersResponse(routResult.Result)
	if err != nil {
		return nil, errors.Wrap(err, "[ ProcessDumpUsers ]")
	}

	var dumpInfo interface{}
	err = json.Unmarshal(serJSONDump, &dumpInfo)
	if err != nil {
		return nil, errors.Wrap(err, "[ ProcessDumpUsers ]")
	}
	result["dump_info"] = dumpInfo

	return result, nil
}