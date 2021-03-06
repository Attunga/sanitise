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
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"encoding/csv"
)

// Need to somehow get this into a different file for neatness maybe?
type settingsStruct struct {
	filenameList       []string          // Name of Log Files to Sanitise
	sanitiseIPs        bool              // Option to not sanitise IPs in a log file
	sanitiseEmails     bool              // Option to not sanitise Emails in a log file
	recordChanges      bool              // (future reference) Option to output a list of changes to the file .. not sure if this should be on
	knownDataExists    bool              // A setting that is set when a known Data File is know to Exist
	knownDataList      map[string]string // Pointer to CSV File with List of known information - optional - loaded from
	devicesPrefixExist bool              // A setting that is set to true when a devices Prfix files exists
	devicesPrefix      map[string]string // Pointer to CSV File with List of prefixes --> [prefix]Naming ie lx Linux_Server##
	exclusionsExist    bool              // A setting that is set when an exclude List is detected as being passed
	excludeList        string            // Pointer to file with list of items we wish to exclude - we simple remove these from final map
	knownItems         string            // Filename with a custom list of known items.
	docx               bool              // Boolean to cover whether we are using a docx file.
}

// A struct which holds the file to be sanitised as well as it's type and other data
// It allows auto marking of file type and other metadata to be saved as well as allowing passing to functions to be
// simplified.
// ummm .. should it be also linked to it's change may saving parameters that need to be passed in for that file.
type fileDetails struct {
	filename       string            // The filename of the file passed
	fileContents   string            // The actual physical file contents .. may even make this an array spliting up the file into smaller sections
	fileType       string            // The type of file, for now docx or text
	fileChangesMap map[string]string // A Map of the changes that will be made to a file
}

const sanitiseVersion = "0.02 alpha"

func main() {

	startTime := time.Now()

	// Get Settings .... and do messy things like check command line arguments
	settings := initialiseSettings()

	log.WithFields(log.Fields{
		"time":     time.Now().String(),
		"function": "main",
		"Log File": settings.filenameList,
	}).Info("Into Main - Usng Log Level:", log.GetLevel())

	// Load a basic changeMap from a settings file that can be used for all looped change files.
	// Create a map that is used to store unique changes
	var changesMap = make(map[string]string)

	// Load in the known items to the changesMap so they are global
	changesMap = getKnownItems(changesMap, settings.knownItems)

	//Here .. should we also make an exclusions map at this stage and pass it as a parameter for processing
	// to save constantly reading from file??

	// Create a slice of File Structs,  just using the name and confirming that each file exists while trying to detect
	// the file time,  primarily whether it is word or not - but maybe also looking for Zips etc.
	fileList := make([]fileDetails, 0)
	fileList = getFileList(&fileList, settings.filenameList)

	// Create the final Exit Message
	exitMessage := ""

	// Process Log Files One by One
	for _, fileDetail := range fileList {

		fmt.Println("Filename:", fileDetail.filename)
		fmt.Println("Type:", fileDetail.fileType)

		// TODO - Do the file sanitising in paraellel - Doing maybe just 4 or so at a time
		switch fileDetail.fileType {
		case "txt":
			fmt.Println("Sanitising Text File", fileDetail.filename)
			exitMessage = exitMessage + sanitiseTextFile(fileDetail.filename, settings, changesMap) + "\n"
		case "docx":
			fmt.Println("Sanitising DocX File", fileDetail.filename)
			exitMessage = exitMessage + sanitiseDocXFile(fileDetail, settings, changesMap) + "\n"
		case "bin":
			fmt.Println("Skipping Binary File", fileDetail.filename)
		default:
			fmt.Println("Unknown File Type", fileDetail.filename)
		}


	}

	// TODO Function to search for known host hames maybe???? ...

	// TODO (Enable via options) Function to search for devices with start with a part of a name .. things like lx ws mg etc etc

	// (Enable via Options)Custom Function to search for usernames and passwords ... this might be text in a file,  things
	// like the firstwave database accounts and passwords etc...
	// Really just loads more stuff into changesMap

	// Display Changes to Screen if Option Passed
	log.WithFields(log.Fields{
		"time":     time.Now().String(),
		"function": "main",
	}).Info("The following changes were made to log files: \n", "changesString")

	// GEt the next file name to put through .. maybe generisize the name checker function as well

	// Write the sorted changes to disk .. using log file naming ... should this method be generisized???
	//writeChangesToDisk()

	// Show Exit  Message
	fmt.Println(exitMessage)

	log.WithFields(log.Fields{
		"time":     time.Now().String(),
		"function": "main",
	}).Debug("Total Time Taken to Complete: ", time.Now().Sub(startTime).String())

}
// ##### End of Main ############




func getFileList(fileList *[]fileDetails, filenameList []string) []fileDetails {

	for _, filename := range filenameList {

		fileDetailsNewFile := fileDetails{
			filename,
			"readFileContents().. maybe later as well due to size limitations,  may load on process",
			getDetectedFileType(filename),
			nil,
		}

		*fileList = append(*fileList, fileDetailsNewFile)
	}

	return *fileList
}

func getKnownItems(changesMap map[string]string, knownItems string) map[string]string {

	// Here we look for the convention of a known items csv list (knownitems.csv)
	// Does known items csv exist
	if fileExists("knownitems.csv") {
		changesMap = loadKnownItemsCSV(changesMap, "knownitems.csv")
	}

	// If the known items is empty the option was not passed and we just exit
	if knownItems == "" {
		println("No Custom items found")
		return changesMap
	}

	// If a known items filename was passwed we process it.
	// check file name exists,  if it is passed and does not exist then we need to exit with an exception.
	if fileExists(knownItems) {
		// Should we confirm csv extension??
		changesMap = loadKnownItemsCSV(changesMap, knownItems)
	} else {
		// A filename was passed but it does not exist,  we exit with an error message now
		println("A custom known items filename was passed but I could not find the file to read from")
		println("Custom Filename Passed:", knownItems)

		// Exit Application ....
		os.Exit(0)
	}

	return changesMap
}
func loadKnownItemsCSV(changesMap map[string]string, knownitemscsvfilename string) map[string]string {

	// Open CSV file
	f, err := os.Open(knownitemscsvfilename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		panic(err)
	}

	// Loop through lines & turn into object
	for _, line := range lines {
		// Duplicates not allowed but later settings will overwrite newer settings???,  Won't happen later though
		// might be ok at this stage
		changesMap[strings.Trim(line[0], " ")] = strings.Trim(line[1], " ")

	}
	return changesMap
}



func sanitiseTextFile(filename string, settings settingsStruct, changesMap map[string]string) string {

	// Read Log File into String ... into text
	var logFileString = getLogFileString(filename)
	//fmt.Println(logFileString)

	//Search through the log file for IP Addresses and return back a Map of Replacement IPs
	// Could be threaded later
	if settings.sanitiseIPs {
		changesMap = getIPAddressesFromLogFile(&logFileString, changesMap)
	}

	//Search through the log file for Email Addresses
	// Could be threaded later
	if settings.sanitiseEmails {
		changesMap = getEmailAddressesFromLogFile(&logFileString, changesMap)
	}

	// The final change to the changes map is the exclude list - it basically confirms the exclusion list is valid as
	// a filename format and then removes any of the excluded files from the final Changes map
	if settings.exclusionsExist {
		changesMap = processExclusions(settings.excludeList, changesMap)
	}

	// Pass off Final comparison string to process the log file
	var logFileProcessed = processLogFile(&logFileString, changesMap)

	//var currentTime = fmt.Sprint(int32(time.Now().Unix()))
	var processedLogFileName = getNextProcessedLogFileName(filename, 1) //"processed:" + currentTime + "-" + logfileName

	// Writes a lot file to disk and returns an exit message
	var exitMessage = writeProcessedLogFileToDisk(processedLogFileName, logFileProcessed)

	// Get a sorted String back from the changesMap that can be used to save to disk or print to screen for debugs
	//changesString := getChangesToString(changesMap)

	// Write Changes String to Disk using Filename Extension

	return exitMessage

}


func getIPAddressesFromLogFile(logFileStringPtr *string, changesMap map[string]string) map[string]string {
	startTime := time.Now()
	var count int

	re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)

	submatchall := re.FindAllString(*logFileStringPtr, -1)

	for _, element := range submatchall {
		// I don't understand what the underscore does but element is true if it exists
		_, elementExists := changesMap[element]
		if !elementExists {
			count++
			changesMap[element] = "IP_Address_" + fmt.Sprintf("%04d", count)
		}

	}

	log.WithFields(log.Fields{
		"time":     time.Now().String(),
		"function": "getIPAddressesFromLogFile",
	}).Debug("Total Time Taken to Complete: ", time.Now().Sub(startTime).String())

	return changesMap
}

// this function might be changed to give another parameters with type of search and then from that we select the type
// of regex. This function very much duplicates the IP Address function
func getEmailAddressesFromLogFile(logFileStringPtr *string, changesMap map[string]string) map[string]string {

	startTime := time.Now()

	var count int

	//var regexString = "[\\w\\.><]+@[\\w\\.><]+\\.[\\w\\.><]+"
	var regexString = "[\\w\\.]+@[\\w\\.><]+\\.[\\w\\.]+"

	re := regexp.MustCompile(regexString)

	submatchall := re.FindAllString(*logFileStringPtr, -1)

	for _, element := range submatchall {
		// I don't understand what the underscore does but element is true if it exists
		_, elementExists := changesMap[element]
		if !elementExists {
			count++
			changesMap[element] = "Email_Address_" + fmt.Sprintf("%04d", count)
			//fmt.Println("Found Email: ", element)
		}

	}

	log.WithFields(log.Fields{
		"time":     time.Now().String(),
		"function": "getEmailAddressesFromLogFile",
	}).Debug("Total Time Taken to Complete: ", time.Now().Sub(startTime).String())

	return changesMap

}

func processLogFile(logFileStringPtr *string, changesMap map[string]string) string {

	totalChanges := len(changesMap)

	startTime := time.Now()

	// Kind of Redundant .. may remove .. or I might put some meta data in the header from a config file... ummm
	*logFileStringPtr = "\nProcessed log file .......\n" + *logFileStringPtr

	// This should be the most  efficient way .. if I could work out how to pass
	// a F$#Kings data object as a parameters .... GRRRRRRRR
	//myReplacer := strings.NewReplacer("lkjdsf", "ljdfljs")
	//logFileProcessedReturn = myReplacer.Replace(logFileString)

	// Total Time Calculation

	changeCount := 0

	// We do the less efficient way .. but it gets me there.
	for k, v := range changesMap {
		processStartTime := time.Now()
		*logFileStringPtr = strings.Replace(*logFileStringPtr, k, v, -1)
		changeCount++

		fmt.Printf("\rChanges Left to Process: %s Current Operation Took:  %s   Extimated Time to Completion:  %s ",
			fmt.Sprint(totalChanges-changeCount),
			time.Now().Sub(processStartTime).Round(time.Microsecond).String(),
			(time.Now().Sub(processStartTime) * time.Duration(totalChanges-changeCount)).Round(time.Second))
		//fmt.Printf("key[%s] value[%s]\n", k, v)
	}

	log.WithFields(log.Fields{
		"time":     time.Now().String(),
		"function": "processLogFile",
	}).Debug("Total Time Taken to Complete processLogFile: ", time.Now().Sub(startTime).String())

	return *logFileStringPtr
}

func getLogFileString(logfile string) string {

	fileBytes, err := ioutil.ReadFile(logfile)
	if err != nil {
		// Process a log file name error here ...
		//log.Fatal(err)
		fmt.Println(err)
		os.Exit(0)
	}
	//defer ioutil.close(logfile)
	return string(fileBytes)
}

func getNextProcessedLogFileName(logfileName string, count int) string {

	// As we are using recursion in this function just do a sanity check on the number of files created to avoid unseen file system errors
	if count > 98 {
		fmt.Println("Too many sanitised files found for today - or file system error")
		os.Exit(0)
	}

	// I should really be trying to get the directory that the file was run in here to ensure it is written in the
	// correct location

	// Pad Integer
	var strCount = strconv.Itoa(count)
	if len(strCount) < 2 {
		strCount = "0" + strCount
	}

	// Get Todays Day in a String
	t := time.Now()
	dateString := t.Format("2006-01-02")

	// check if the filename exists
	filename := "sanitised_" + dateString + "_" + strCount + "_" + logfileName
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		//fmt.Println("File does not exist", filename)
		return filename
	}

	// Recursion back to the same function if the file exists
	count++
	return getNextProcessedLogFileName(logfileName, count)
}

func writeProcessedLogFileToDisk(processedLogFileName string, logFileProcessed string) string {

	err := ioutil.WriteFile(processedLogFileName, []byte(logFileProcessed), 0644)

	if err != nil {
		log.Fatal(err)
		return err.Error()
	}

	return "\nNew Filename  " + processedLogFileName + " written successfully"
}

func processExclusions(excludeListFileName string, changesMap map[string]string) map[string]string {

	newChangesMap := changesMap

	// Try to Open the file .... We should have already Checked that it exists in the Settings area using a function
	fileBytes, err := ioutil.ReadFile(excludeListFileName)
	if err != nil {
		// Process a log file error here
		//log.Fatal(err)
		fmt.Println(err)
		os.Exit(0)
	}

	// loop through files and get line
	scanner := bufio.NewScanner(strings.NewReader(string(fileBytes)))
	for scanner.Scan() {
		// if item exists in changes map then remove it
		foundItem := strings.Trim(scanner.Text(), " ")
		if foundItem != "" {
			delete(newChangesMap, foundItem)
		}
		// Print out the exclusion that has been found
		log.WithFields(log.Fields{
			"time":     time.Now().String(),
			"function": "processExclusions",
		}).Info("Exclusion Found:", scanner.Text())

	}

	return newChangesMap
}

func initialiseSettings() settingsStruct {

	// Overrides the help to basically add a custom message about passing a parameter
	overrideHelp()

	settings := new(settingsStruct)

	sanitiseIPsPtr := flag.Bool("sanitiseips", true, "Locate and Sanitise IP addresses in Log files")
	sanitiseEmailsPtr := flag.Bool("sanitiseemails", true, "Locate and Sanitise email addresses in Log files")
	knownDataListPtr := flag.String("knowndatafile", "", "A CSV file with a list of known data and preferred naming - optional")
	devicesPrefixPtr := flag.String("devicesprefixfile", "", "CSV File with a list of device prefixes and preferred naming - optional")
	excludeListPtr := flag.String("exclude", "", "Simple file list with items that will not be sanitised - optional")
	loglevelPtr := flag.String("loglevel", "warn", "The Logging Level we wish to set (debug, info, warn, error) - Default - warn")
	logToStdOutPtr := flag.Bool("stdout", false, "Send logging messages to standard output instead of to system logging")
	docxPtr := flag.Bool("docx", false, "Process a Microsoft Word DocX file format")
	knownItems := flag.String("knownitems", "", "A CSV File of known items to be sanitised in format [item].[new_name]")
	versionPtr := flag.Bool("version", false, "Show sanitiser version")

	// Parse flags so that they can be seen
	flag.Parse()

	if *versionPtr {
		fmt.Println("Sanitise Version:", sanitiseVersion)
		os.Exit(0)
	}

	settings.sanitiseIPs = *sanitiseIPsPtr
	settings.sanitiseEmails = *sanitiseEmailsPtr
	//settings.devicesPrefix = getPrefixesMap(devicesPrefixPtr)
	//settings.knownDataList = getKnownDataListMap(*devicesPrefixPtr)
	settings.excludeList = *excludeListPtr
	settings.exclusionsExist = false
	settings.docx = *docxPtr
	settings.filenameList = flag.Args()
	settings.knownItems = *knownItems

	// Detect Options being passed after files to be processe,  then give a message and get out

	// No Command Line Arguments Provided or a log filename has not been given
	if len(os.Args) < 2 || len(settings.filenameList) == 0 {
		fmt.Printf("Please provide one or more log files to sanitise - %v [options] [file1] [file2] \n", os.Args[0])
		fmt.Printf("All optional parameters must come before files to be sanitised")
		fmt.Printf("For additional options use - %v -help\n", os.Args[0])
		os.Exit(0)
	}

	// Bug out message so that we can give multiple errors at once
	bugOutMessages := ""

	// Loop through filename parameters and check whether the the files exist and if the DocX parameter
	// has been passed whether it is a valid docx file.
	for _, filename := range settings.filenameList {

		// Make sure that the file does not start with a dash - this could mean it is a parameter that is being passed
		// after filenames
		// if filename.startsWith - then go but

		// Check that our Files given as parameters exist, bug out if there is a file passed as a parameter that does not exist.
		if filename != "" && !fileExists(filename) {
			bugOutMessages = bugOutMessages + "I could not find the log file file named: " + filename + "\n"
		}

		// Sanity Check the DocX parameters,  if a true is passed for the Docx then we make sure that the file extension is actually a docx
		if *docxPtr && filepath.Ext(filename) != ".docx" {
			fmt.Println("Docx option is passed but filename is not a docx file")
			fmt.Println("Please provide a docx file - or convert the document to docx format")
			os.Exit(0)
		}
	}

	// quit if there are file name errors
	if bugOutMessages != "" {
		fmt.Print(bugOutMessages)
		os.Exit(0)
	}

	if settings.excludeList != "" && !fileExists(settings.excludeList) {
		bugOutMessages = bugOutMessages + "I could not find the exclusions file named: " + settings.excludeList + "\n"
	} else if settings.excludeList == "" {
		settings.exclusionsExist = false
	} else {
		settings.exclusionsExist = true
	}
	if *knownDataListPtr != "" && !fileExists(*knownDataListPtr) {
		bugOutMessages = bugOutMessages + "I could not find the Known Data List file named: " + *knownDataListPtr + "\n"
	}
	if *devicesPrefixPtr != "" && !fileExists(*devicesPrefixPtr) {
		bugOutMessages = bugOutMessages + "I could not find the Devices Prefix file named: " + *devicesPrefixPtr + "\n"
	}
	if bugOutMessages != "" {
		fmt.Print(bugOutMessages)
		os.Exit(0)
	}

	// Sanity Check to confirm we actually have some work to do ... and if this gets any longer it is going into it's own function
	if !settings.sanitiseEmails &&
		!settings.sanitiseIPs &&
		*knownDataListPtr == "" &&
		*devicesPrefixPtr == "" {
		fmt.Println("I got nothing to sanitise bro??")
		os.Exit(0)
	}

	//Set Log Output instead of the default of STD Error .. need to confirm what this does
	if *logToStdOutPtr {
		log.SetOutput(os.Stdout)
	}

	// Set Up logging levels
	switch strings.ToLower(*loglevelPtr) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.WarnLevel)
		fmt.Println("Invalid Logging Level Used - Setting Logging to the Default Warning Level")
	}

	log.WithFields(log.Fields{
		"time":     time.Now().String(),
		"function": "initialiseSettings",
	}).Info("Finished Settings - Usng Log Level:", log.GetLevel())

	return *settings
}
