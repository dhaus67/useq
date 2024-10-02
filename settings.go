/*
 * Copyright 2024 dhaus67
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package useq

import (
	"encoding/json"
	"fmt"
	"os"
)

// Settings holds all the settings for the UseqAnalyzer.
type Settings struct {
	// The functions are fully qualified function names including the package (e.g. fmt.Printf).
	Functions []string `json:"functions"`
	// FunctionsPerPackage is a map of package names to the functions that should be checked.
	FunctionsPerPackage functionsPerPackage
}

// ReadSettingsFromFile reads the settings from the given path.
func ReadSettingsFromFile(f string) (Settings, error) {
	contents, err := os.ReadFile(f)
	if err != nil {
		return Settings{}, fmt.Errorf("reading file %q: %w", f, err)
	}

	var settings Settings
	if err := json.Unmarshal(contents, &settings); err != nil {
		return Settings{}, fmt.Errorf("parsing settings: %w", err)
	}

	return settings, nil
}
