package bootstrapper

import (
	"archive/zip"
	"errors"
	"fmt"
	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/modules/consensus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var status = "N/A"
var skip = false

// Skip bootstrapping consensus from consensus.scpri.me
func Skip() {
	skip = true
}

// Progress returns the bootstrapper's progress as a percentage.
func Progress() string {
	return status
}

// Start begins the process of bootstrapping consensus from consensus.scpri.me.
func Start(dataDir string) {
	consensusDir := filepath.Join(dataDir, modules.ConsensusDir)
	consensusDb := filepath.Join(consensusDir, consensus.DatabaseFilename)
	_, err := os.Stat(consensusDir)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(consensusDir, os.ModePerm)
	}
	if err != nil {
		// Unable to create the consensus directory.
		// Return early and let the consensus module create the directory.
		return
	}
	_, err = os.Stat(consensusDb)
	if !errors.Is(err, os.ErrNotExist) {
		// Consensus database already exists so there is no need to download it.
		return
	}
	size, err := consensusSize()
	if size == 0 || err != nil {
		// Do not download consensus-latest.zip because something is wrong.
		return
	}
	tmp, err := ioutil.TempFile(os.TempDir(), "scprime-consensus")
	if err != nil {
		// Unable to create the temporary file to download the consensus database to.
		// Return early and let the consensus module create the directory from scratch.
		return
	}
	defer os.Remove(tmp.Name())
	var sem = make(chan int, 1)
	sem <- 1
	go func(filepath string) {
		consensusDownload(filepath)
		<-sem
	}(tmp.Name())
	// Updates the status.
	status = `0%`
	for i := 1; true; i++ {
		updateStatus(tmp.Name(), size)
		if len(sem) == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if skip {
		return
	}
	status = `99%`
	decompress(tmp.Name(), consensusDb)
	status = `100%`
}

// Decompress the zip archive; move consensus.db to the destination.
func decompress(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		if f.Name != "consensus.db" {
			continue
		}
		outFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		for {
			_, err := io.CopyN(outFile, rc, 1024)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()
	}
	return nil
}

// Returns the size of the latest consensus database in bytes.
func consensusSize() (int64, error) {
	resp, err := http.Head("https://consensus.scpri.me/releases/consensus-latest.zip")
	if err != nil {
		return 0, err
	}
	return resp.ContentLength, nil
}

// Downloads the consensus databse to a local file without loading the whole file into memory.
func consensusDownload(target string) error {
	// Get the data
	resp, err := http.Get("https://consensus.scpri.me/releases/consensus-latest.zip")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Create the file
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	// Write the body to file
	for {
		_, err := io.CopyN(out, resp.Body, 1024)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if skip {
			break
		}
	}
	out.Close()
	return nil
}

// updates the status.
func updateStatus(filepath string, size int64) {
	fi, _ := os.Stat(filepath)
	progress := int(float64(fi.Size()) / float64(size) * float64(100))
	if progress > 0 && progress < 99 {
		status = fmt.Sprintf("%d%%", progress)
	}
}
