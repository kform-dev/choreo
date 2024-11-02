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

package resource

import (
	"errors"
	"fmt"
)

var missingResourceError = fmt.Errorf(`you must provide one or more resources by argument or filename.
Example resource specifications include:
   '-f rsrc.yaml'
   '--filename=rsrc.json'`)

var StdinMultiUseError = errors.New("standard input cannot be used for multiple arguments")
