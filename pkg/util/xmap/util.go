// Copyright 2020 Douyu
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

package xmap

import (
	"fmt"
	"reflect"

	xcast "github.com/spf13/cast"
)

// MergeStringMap merge two map
func MergeStringMap(dest, src map[string]any) {
	for sk, sv := range src {
		tv, ok := dest[sk]
		if !ok {
			// val不存在时，直接赋值
			dest[sk] = sv
			continue
		}

		svType := reflect.TypeOf(sv)
		tvType := reflect.TypeOf(tv)
		if svType != tvType {
			fmt.Println("continue, type is different")
			continue
		}

		switch ttv := tv.(type) {
		case map[any]any:
			tsv := sv.(map[any]any)
			ssv := ToMapStringInterface(tsv)
			stv := ToMapStringInterface(ttv)
			MergeStringMap(stv, ssv)
			dest[sk] = stv
		case map[string]any:
			MergeStringMap(ttv, sv.(map[string]any))
			dest[sk] = ttv
		default:
			dest[sk] = sv
		}
	}
}

// ToMapStringInterface cast map[any]any to map[string]any
func ToMapStringInterface(src map[any]any) map[string]any {
	tgt := map[string]any{}
	for k, v := range src {
		tgt[fmt.Sprintf("%v", k)] = v
	}
	return tgt
}

// DeepSearchInMap deep search in map
func DeepSearchInMap(m map[string]any, paths ...string) map[string]any {
	//深度拷贝
	mtmp := make(map[string]any)
	for k, v := range m {
		mtmp[k] = v
	}
	for _, k := range paths {
		m2, ok := mtmp[k]
		if !ok {
			m3 := make(map[string]any)
			mtmp[k] = m3
			mtmp = m3
			continue
		}

		m3, err := xcast.ToStringMapE(m2)
		if err != nil {
			m3 = make(map[string]any)
			mtmp[k] = m3
		}
		// continue search
		mtmp = m3
	}
	return mtmp
}
