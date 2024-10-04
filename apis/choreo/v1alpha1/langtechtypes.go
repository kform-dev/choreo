/*
Copyright 2024 Nokia.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

type LangTechType string

const (
	GoTemplate_LangTechType    LangTechType = "goTemplate"
	JinjaTemplate_LangTechType LangTechType = "jinjaTemplate"
	Invalid_LangTechType       LangTechType = "invalid"
)

func (r LangTechType) String() string {
	switch r {
	case GoTemplate_LangTechType:
		return "goTemplate"
	case JinjaTemplate_LangTechType:
		return "jinjaTemplate"
	default:
		return "invalid"
	}
}

type SoftwardTechnologyType string

const (
	SoftwardTechnologyType_Invalid       SoftwardTechnologyType = "invalid"
	SoftwardTechnologyType_Starlark      SoftwardTechnologyType = "starlark"
	SoftwardTechnologyType_Kform         SoftwardTechnologyType = "kform"
	SoftwardTechnologyType_GoTemplate    SoftwardTechnologyType = "gotemplate"
	SoftwardTechnologyType_JinjaTemplate SoftwardTechnologyType = "jinjatemplate"
	SoftwardTechnologyType_Internal      SoftwardTechnologyType = "internal"
)

func (r SoftwardTechnologyType) String() string {
	switch r {
	case SoftwardTechnologyType_Starlark:
		return "starlark"
	case SoftwardTechnologyType_Kform:
		return "kform"
	case SoftwardTechnologyType_GoTemplate:
		return "gotmplate"
	case SoftwardTechnologyType_JinjaTemplate:
		return "jinjatmplate"
	case SoftwardTechnologyType_Invalid:
		return "internal"
	default:
		return "invalid"
	}
}

func GetSoftwardTechnologyType(s string) SoftwardTechnologyType {
	switch s {
	case "starlark":
		return SoftwardTechnologyType_Starlark
	case "kform":
		return SoftwardTechnologyType_Kform
	default:
		return SoftwardTechnologyType_Invalid
	}
}
