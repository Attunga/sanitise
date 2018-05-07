// +build linux darwin

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
	"os/exec"
	"log"
	"strings"
)

func detectFileTypeViaOS(filename string) string {

	// TODO check whether file utility is installed .. warn if not found

	//

	out, err := exec.Command("file", "-b", filename).Output()
	if err != nil {
		println("File most likely not found in OS - Please install File Package")
		log.Fatal(err)
	}
	//fmt.Println(string(out))
	fileTypeDetected := strings.Split(string(out), " ")[0]

	//fmt.Println(fileTypeDetected)

	switch fileTypeDetected {
	case "ASCII":
		return "txt"
	case "ELF":
		return "bin"
	default:
		return "unknown"
	}

}
