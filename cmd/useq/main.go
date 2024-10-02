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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dhaus67/useq"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	var settingsFile string
	flag.StringVar(&settingsFile, "config", "", "path to the config file to use")
	flag.Parse()

	settings := useq.Settings{}
	if settingsFile != "" {
		settingsFromFile, err := useq.ReadSettingsFromFile(settingsFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		settings = settingsFromFile
	}

	analyzer, err := useq.NewAnalyzer(settings)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	singlechecker.Main(analyzer)
}
