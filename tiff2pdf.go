package tiff2pdf

/*
#cgo CFLAGS: -D_THREAD_SAFE -pthread -I../../vadz/libtiff/libtiff
#cgo LDFLAGS: -lm
#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include "c/libtiff.h"
#include "c/tiff2pdf.c"
#include "c/tif_golang.c"
*/
import "C"
import "errors"

//Config represents the command line tiff2pdf configuration
type Config struct {
	// PageSize sets the PDF page size, e.g. legal, letter or A4
	PageSize string
	// FullPage makes the tiff image fill the PDF page
	FullPage bool

	// Creator is the image software used to create the document
	Creator string

	// Author is the document name
	Author string
	// Subject is the document description
	Subject string
	// Title is the document name
	Title string
}

// DefaultConfig creates the default tiff2pdf configuration
func DefaultConfig() *Config {
	return &Config{
		PageSize: "A4",
		FullPage: true,
		Creator:  "go-tiff2pdf",
	}
}

func createTiff(tiff []byte, name, mode string) (*C.TIFF, error) {
	newFd := NewFd(tiff)
	tif := C.TIFFFdOpen(C.int(newFd.fd), C.CString(name), C.CString(mode))
	if tif == nil {
		return nil, ErrOpenFailed
	}
	return tif, nil
}

func configureT2p(t2p *C.T2P, config *Config) {
	// TODO page size

	if config.FullPage {
		t2p.pdf_image_fillpage = 1
	} else {
		t2p.pdf_image_fillpage = 0
	}

	// FIXME if len(config.Creator) == 0, is that "no flag" or "empty string"
	t2p.pdf_creator = stringTo512Cchar(config.Creator)
	t2p.pdf_author = stringTo512Cchar(config.Author)
	t2p.pdf_subject = stringTo512Cchar(config.Subject)
	t2p.pdf_title = stringTo512Cchar(config.Title)
}

func stringTo512Cchar(s string) [512]C.char {
	var cArr [512]C.char
	for i, c := range s {
		cArr[i] = C.char(c)
	}
	cArr[len(s)] = C.char(0)
	return cArr
}

// ConvertTiffToPDF converts an input TIFF byte array to an output PDF byte array
func ConvertTiffToPDF(tiff []byte, config *Config, inputName string, outputName string) ([]byte, error) {
	input, err := createTiff(tiff, inputName, "rw")
	if err != nil {
		return nil, err
	}

	output, err := createTiff([]byte{}, outputName, "w")
	if err != nil {
		return nil, err
	}
	GoTiffSeekProc(int(output.tif_fd), 0, 0)

	t2p := C.t2p_init()
	if t2p == nil {
		return nil, errors.New("Error: t2p_init!")
	}

	configureT2p(t2p, config)

	// t2p.outputfile = C.FILE(output.tif_fd)
	C.t2p_write_pdf(t2p, input, output)
	if t2p.t2p_error != 0 {
		C.t2p_free(t2p)
		return nil, errors.New("t2p_error")
	}

	C.t2p_free(t2p)

	return fdMap[int(output.tif_fd)].buffer, nil
}
