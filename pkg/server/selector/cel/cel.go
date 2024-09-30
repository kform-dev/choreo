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

package cel

import (
	"fmt"

	"github.com/google/cel-go/cel"
)

func getCelEnv(varName string) (*cel.Env, error) {
	var opts []cel.EnvOption
	opts = append(opts, cel.EagerlyValidateDeclarations(true), cel.DefaultUTCTimeZone(true))
	//opts = append(opts, library.ExtensionLibs...)
	opts = append(opts, cel.Variable(varName, cel.DynType))
	return cel.NewEnv(opts...)
}

func GetValue(data map[string]any, celExpression string) (string, bool, error) {
	env, err := getCelEnv("input")
	if err != nil {
		return "", false, fmt.Errorf("cel environment initialization failed %s", err.Error())
	}
	// we add input to the cell variable
	expr := fmt.Sprintf("input.%s", celExpression)
	ast, iss := env.Compile(expr)
	if iss.Err() != nil {
		return "", false, fmt.Errorf("compilation failed for expression '%s': %s", celExpression, iss.Err().Error())
	}
	//_, err = cel.AstToCheckedExpr(ast)
	//if err != nil {
	//	return "", false, fmt.Errorf("ast to checked expression failed for expr %s, err: %s", celExpression, err.Error())
	//}
	prog, err := env.Program(ast, cel.EvalOptions(cel.OptOptimize))
	if err != nil {
		return "", false, fmt.Errorf("cel program creation failed for expression '%s': %s", celExpression, err.Error())
	}
	vars := map[string]any{"input": data}
	val, _, err := prog.Eval(vars)
	if err != nil {
		return "", false, nil
	}
	if val == nil {
		return "nil", false, nil
	}
	// we convert to string in comparison
	return fmt.Sprintf("%v", val.Value()), true, nil
}
