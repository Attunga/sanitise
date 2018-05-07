package main

import (
	"path/filepath"
)

// Function to detect the file type - in a seperate file because it is likely to be a bit busy.
func getDetectedFileType(filename string) string {

	//  Try to detect common types of files by extension before going to the OS or looking in the file

	// should we detect common file names?? ... .like maillog, messages etc .. get a starts with thingo

	// Get the file Extension from the file.
	// If no extension can be found then we return blank and will go off to OS level detection
	fileExtension := filepath.Ext(filename)

	switch fileExtension {
	case "":
		return detectFileTypeViaOS(filename)
	case ".docx":
		return "docx" // docx - process
	case "doc" :
	    return "doc" //  older doc - ask  users to convert to docx or do we convert?
	case ".log",".txt":
		return "txt" // Simple Text Files - will process
	case ".exe",",com",".dll":
		return "bin" // Binaries .. to be ignored
	case ".zip":
		return "zip" // Zip
	case ".gz":
		return "gzip" // Zip
	case ".bz2":
		return "bzip" // Zip
	default:
		return detectFileTypeViaOS(filename)  // no file extension
	}

	// not really needed but here as a safety anyway
	return detectFileTypeViaOS(filename)

}


