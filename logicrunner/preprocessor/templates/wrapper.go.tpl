//
// Copyright 2019 Insolar Technologies GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Code generated by insgocc. DO NOT EDIT.
// source template in logicrunner/preprocessor/templates

package {{ .Package }}

import (
{{- range $import, $i := .Imports }}
	{{ $import }}
{{- end }}
)

const PanicIsLogicalError = false

func INS_META_INFO() ([] map[string]string) {
	result := make([]map[string] string, 0)
	{{ range $method := .Methods }}
		{{ if $method.SagaInfo.IsSaga }}
		{
		info := make(map[string] string, 3)
		info["Type"] = "SagaInfo"
		info["MethodName"] = "{{ $method.Name }}"
		info["RollbackMethodName"] = "{{ $method.SagaInfo.RollbackMethodName }}"
		result = append(result, info)
		}
		{{end}}
	{{end}}
	return result
}

func INSMETHOD_GetCode(object []byte, data []byte) ([]byte, []byte, error) {
	ph := common.CurrentProxyCtx
	self := new({{ $.ContractType }})

	if len(object) == 0 {
		return nil, nil, &foundation.Error{S: "[ Fake GetCode ] ( Generated Method ) Object is nil"}
	}

	err := ph.Deserialize(object, self)
	if err != nil {
		e := &foundation.Error{ S: "[ Fake GetCode ] ( Generated Method ) Can't deserialize args.Data: " + err.Error() }
		return nil, nil, e
	}

	state := []byte{}
	err = ph.Serialize(self, &state)
	if err != nil {
		return nil, nil, err
	}

	ret := []byte{}
	err = ph.Serialize([]interface{} { self.GetCode().Bytes() }, &ret)

	return state, ret, err
}

func INSMETHOD_GetPrototype(object []byte, data []byte) ([]byte, []byte, error) {
	ph := common.CurrentProxyCtx
	self := new({{ $.ContractType }})

	if len(object) == 0 {
		return nil, nil, &foundation.Error{ S: "[ Fake GetPrototype ] ( Generated Method ) Object is nil"}
	}

	err := ph.Deserialize(object, self)
	if err != nil {
		e := &foundation.Error{ S: "[ Fake GetPrototype ] ( Generated Method ) Can't deserialize args.Data: " + err.Error() }
		return nil, nil, e
	}

	state := []byte{}
	err = ph.Serialize(self, &state)
	if err != nil {
		return nil, nil, err
	}

	ret := []byte{}
	err = ph.Serialize([]interface{} { self.GetPrototype().Bytes() }, &ret)

	return state, ret, err
}

{{ range $method := .Methods }}
func INSMETHOD_{{ $method.Name }}(object []byte, data []byte) (newState []byte, result []byte, err error) {
	ph := common.CurrentProxyCtx
	ph.SetSystemError(nil)

	self := new({{ $.ContractType }})

	if len(object) == 0 {
		err = &foundation.Error{ S: "[ Fake{{ $method.Name }} ] ( INSMETHOD_* ) ( Generated Method ) Object is nil"}
		return
	}

	err = ph.Deserialize(object, self)
	if err != nil {
		err = &foundation.Error{ S: "[ Fake{{ $method.Name }} ] ( INSMETHOD_* ) ( Generated Method ) Can't deserialize args.Data: " + err.Error() }
		return
	}

	{{ $method.ArgumentsZeroList }}
	err = ph.Deserialize(data, &args)
	if err != nil {
		err = &foundation.Error{ S: "[ Fake{{ $method.Name }} ] ( INSMETHOD_* ) ( Generated Method ) Can't deserialize args.Arguments: " + err.Error() }
		return
	}

	{{ $method.ResultDefinitions }}

	serializeResults := func() error {
		return ph.Serialize(
			foundation.Result{Returns:[]interface{}{ {{ $method.Results }} }},
			&result,
		)
	}

	needRecover := true
	defer func() {
		if !needRecover {
			return
		}
		if r := recover(); r != nil {
			recoveredError := errors.Wrap(errors.Errorf("%v", r), "Failed to execute method (panic)")
			recoveredError = ph.MakeErrorSerializable(recoveredError)

			if PanicIsLogicalError {
				ret{{ $method.LastErrorInRes }} = recoveredError

				newState = object
				err = serializeResults()
			} else {
				err = recoveredError
			}

		}
	}()

	{{ $method.Results }} = self.{{ $method.Name }}( {{ $method.Arguments }} )

	needRecover = false

	if ph.GetSystemError() != nil {
		return nil, nil, ph.GetSystemError()
	}

	err = ph.Serialize(self, &newState)
	if err != nil {
		return nil, nil, err
	}

{{ range $i := $method.ErrorInterfaceInRes }}
	ret{{ $i }} = ph.MakeErrorSerializable(ret{{ $i }})
{{ end }}

	err = serializeResults()
	if err != nil {
		return
	}

	return
}
{{ end }}


{{ range $f := .Functions }}
func INSCONSTRUCTOR_{{ $f.Name }}(ref insolar.Reference, data []byte) (state []byte, result []byte, err error) {
	ph := common.CurrentProxyCtx
	ph.SetSystemError(nil)

	{{ $f.ArgumentsZeroList }}
	err = ph.Deserialize(data, &args)
	if err != nil {
		err = &foundation.Error{ S: "[ Fake{{ $f.Name }} ] ( INSCONSTRUCTOR_* ) ( Generated Method ) Can't deserialize args.Arguments: " + err.Error() }
		return
	}

	{{ $f.ResultDefinitions }}

	serializeResults := func() error {
		return ph.Serialize(
			foundation.Result{Returns:[]interface{}{ ref, ret1 }},
			&result,
		)
	}

	needRecover := true
	defer func() {
		if !needRecover {
			return
		}
		if r := recover(); r != nil {
			recoveredError := errors.Wrap(errors.Errorf("%v", r), "Failed to execute constructor (panic)")
			recoveredError = ph.MakeErrorSerializable(recoveredError)

			if PanicIsLogicalError {
				ret1 = recoveredError

				state = data
				err = serializeResults()
			} else {
				err = recoveredError
			}
		}
	}()

	{{ $f.Results }} = {{ $f.Name }}( {{ $f.Arguments }} )

	needRecover = false

	ret1 = ph.MakeErrorSerializable(ret1)
	if ret0 == nil && ret1 == nil {
		ret1 = &foundation.Error{ S: "constructor returned nil" }
	}

	if ph.GetSystemError() != nil {
		err = ph.GetSystemError()
		return
	}

	err = serializeResults()
	if err != nil {
		return
	}

	if ret1 != nil {
		// logical error, the result should be registered with type RequestSideEffectNone
		state = nil
		return
	}

	err = ph.Serialize(ret0, &state)
	if err != nil {
		return
	}

	return
}
{{ end }}

{{ if $.GenerateInitialize -}}
func Initialize() insolar.ContractWrapper {
	return insolar.ContractWrapper{
		GetCode: INSMETHOD_GetCode,
		GetPrototype: INSMETHOD_GetPrototype,
		Methods: insolar.ContractMethods{
			{{ range $method := .Methods -}}
					"{{ $method.Name }}": INSMETHOD_{{ $method.Name }},
			{{ end }}
		},
		Constructors: insolar.ContractConstructors{
			{{ range $f := .Functions -}}
					"{{ $f.Name }}": INSCONSTRUCTOR_{{ $f.Name }},
			{{ end }}
		},
	}
}
{{- end }}
