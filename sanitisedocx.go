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
	"github.com/nguyenthenguyen/docx"
	"time"
	"fmt"
)

func sanitiseDocXFile(fileDetail fileDetails, settings settingsStruct, changesMap map[string]string) string {

	// Read DocX Text Into fileDetails Struct
	fileDetail.fileContents = getDocXTextContent(fileDetail)

	//Search through the log file for IP Addresses and return back a Map of Replacement IPs
	// Could be threaded later
	if settings.sanitiseIPs {
		changesMap = getIPAddressesFromLogFile(&fileDetail.fileContents, changesMap)
	}

	//Search through the log file for Email Addresses
	// Could be threaded later
	if settings.sanitiseEmails {
		changesMap = getEmailAddressesFromLogFile(&fileDetail.fileContents, changesMap)
	}

	// The final change to the changes map is the exclude list - it basically confirms the exclusion list is valid as
	// a filename format and then removes any of the excluded files from the final Changes map
	if settings.exclusionsExist {
		changesMap = processExclusions(settings.excludeList, changesMap)
	}



	//var currentTime = fmt.Sprint(int32(time.Now().Unix()))
	var processedLogFileName = getNextProcessedLogFileName(fileDetail.filename, 1) //"processed:" + currentTime + "-" + logfileName


	// Pass off Final comparison string to process the log file and write to disk.   Returns a status message
	exitMessage := processDocxFile(&fileDetail, processedLogFileName, changesMap)


	// Get a sorted String back from the changesMap that can be used to save to disk or print to screen for debugs
	//changesString := getChangesToString(changesMap)

	// Write Changes String to Disk using Filename Extension


	return exitMessage


}


func getDocXTextContent(fileDetail fileDetails) string {

	// Read from docx file
	r, err := docx.ReadDocxFile(fileDetail.filename)
	if err != nil {
		panic(err)
	}

	docx1 := r.Editable()
	//fmt.Println("Showing Content ##########################################")
	//fmt.Println(docx1.GetDocXContent())
	//fmt.Println("Ending Content ##########################################")

	defer r.Close()

    return docx1.GetDocXContent()
}

func processDocxFile(fileDetail *fileDetails, newFileName string ,changesMap map[string]string) string {

	r, err := docx.ReadDocxFile(fileDetail.filename)
	if err != nil {
		panic(err)
	}

	docx1 := r.Editable()

	totalChanges := len(changesMap)

	//startTime := time.Now()

	// Kind of Redundant .. may remove .. or I might put some meta data in the header from a config file... ummm
	//*logFileStringPtr = "\nProcessed log file .......\n" + *logFileStringPtr

	// This should be the most  efficient way .. if I could work out how to pass
	// a F$#Kings data object as a parameters .... GRRRRRRRR
	//myReplacer := strings.NewReplacer("lkjdsf", "ljdfljs")
	//logFileProcessedReturn = myReplacer.Replace(logFileString)

	// Total Time Calculation

	changeCount := 0

	// We do the less efficient way .. but it gets me there.
	for k, v := range changesMap {
		processStartTime := time.Now()
		// do actual replacements of text
		docx1.Replace(k, v, -1)
		docx1.ReplaceHeader(k, v)
		docx1.ReplaceFooter(k, v)
		docx1.ReplaceLink(k, v, -1)
		changeCount++

		fmt.Printf("\rChanges Left to Process: %s Current Operation Took:  %s   Extimated Time to Completion:  %s ",
			fmt.Sprint(totalChanges-changeCount),
			time.Now().Sub(processStartTime).Round(time.Microsecond).String(),
			(time.Now().Sub(processStartTime) * time.Duration(totalChanges-changeCount)).Round(time.Second))
		//fmt.Printf("key[%s] value[%s]\n", k, v)
	}

	//log.WithFields(log.Fields{
	//	"time":     time.Now().String(),
	//	"function": "processLogFile",
	//}).Debug("Total Time Taken to Complete processLogFile: ", time.Now().Sub(startTime).String())

    // Write Changed DocX File to Disk
	docx1.WriteToFile(newFileName)

	// close our document
	defer r.Close()

	return "All Good"
}

