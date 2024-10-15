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

package git

import "fmt"

type FatalError struct {
	Message string
}

func (e *FatalError) Error() string {
	return fmt.Sprintf("Fatal: %s", e.Message)
}

// WarningError represents non-critical errors that serve as warnings.
type WarningError struct {
	Message string
}

func (e *WarningError) Error() string {
	return fmt.Sprintf("Warning: %s", e.Message)
}

// Convenience functions for creating errors
func NewFatalError(message string) error {
	return &FatalError{Message: message}
}

func NewWarningError(message string) error {
	return &WarningError{Message: message}
}

func IsFatalError(err error) bool {
	_, ok := err.(*FatalError)
	return ok
}

// IsWarningError checks if the error is a WarningError.
func IsWarningError(err error) bool {
	_, ok := err.(*WarningError)
	return ok
}
