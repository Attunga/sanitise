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
	"flag"
	"fmt"
	"os"
)

// This is a separate file so that I can complete rewrite the help output if required.
func overrideHelp() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "To use %s type sanitise, optional options and one or more files to be sanitised.\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s [Options] [file1] [file2] \n", os.Args[0])
		fmt.Println("Options must come before files to be processed")
		fmt.Println("Additional Optional Arguments")
		flag.PrintDefaults()
		os.Exit(0)
	}

}
