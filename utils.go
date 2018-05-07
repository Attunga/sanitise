/*
 * Log File Sanitiser
 * Copyright (c) Lindsay Steele - 2018.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"os"
	"sort"
)

// Just a bunch of utility functions.

// Function to Return true/false depending on whether a file exists
func fileExists(filename string) bool {

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		//File does not exist
		return false
	}
	// File does exist
	return true
}

// Method takes in a changesMap,  sorts it and then returns a string of Changes Values to Original Values
func getChangesToString(changesMap map[string]string) string {

	// My dodgy way of doing this .. might be a better way

	type changesStruct struct {
		changedValue  string // Changed Value
		originalValue string // Original Value Currently the unique key
	}

	// Store the keys in a slice of changesStructs
	changesSlice := []changesStruct{}
	for key := range changesMap {
		change := changesStruct{changesMap[key], key}
		changesSlice = append(changesSlice, change)
	}

	// Sort the slice of changes by the Changed Value
	sort.SliceStable(changesSlice, func(i, j int) bool { return changesSlice[i].changedValue < changesSlice[j].changedValue })

	// Iterate over the changed Values and Print Line by Line to a String
	changesString := ""
	for _, change := range changesSlice {
		changesString = changesString + change.changedValue + "    " + change.originalValue + "\n"
	}

	return changesString
}
