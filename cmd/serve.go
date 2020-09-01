/*
Copyright © 2020 Andrew Mobbs

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

type Slot struct {
	XMLName     xml.Name `xml:,any`
	SlotNumber  int      `xml:"mSlotNumber"`
	Status      int      `xml:"mStatus"`
	DiskState   int      `xml:"mDiskState"`
	ErrorCount  int      `xml:"mErrorCount"`
	Make        string   `xml:"mMake"`
	DiskFwRev   string   `xml:"mDiskFwRev"`
	Serial      string   `xml:"mSerial"`
	PhyCapacity int64    `xml:"mPhysicalCapacity"`
	RotSpeed    int      `xml:"RotationalSpeed"`
}

type Status struct {
	Serial        string `xml:"mSerial"`
	Name          string `xml:"mName"`
	Version       string `xml:"mVersion"`
	TotalCapacity int64  `xml:"mTotalCapacityProtected"`
	UsedCapacity  int64  `xml:"mUsedCapacityProtected"`
	FreeCapacity  int64  `xml:"mFreeCapacityProtected"`
	DNASStatus    int    `xml:"DNASStatus"`
	Status        int    `xml:"mStatus"`
	Slots         struct {
		Slots []Slot `xml:",any"`
	} `xml:"mSlotsExp"`
}

type Monitor struct {
	HaveGoodStatus    bool
	LastGoodFetchTime time.Time
	LastFetchTime     time.Time
	LastGoodStatus    Status
	LastFetchError    error
}

var m = new(Monitor)

// updateStatus gets the current status from the Drobo NASD status port
func (m *Monitor) updateStatus() error {
	if time.Since(m.LastFetchTime).Seconds() < 10 {
		return m.LastFetchError
	}
	m.LastFetchTime = time.Now()
	conn, err := net.Dial("tcp", drobo+":5000")
	if err != nil {
		log.Println("Dialling error" + err.Error())
		m.LastFetchError = err
		return err
	}
	defer conn.Close()
	buf := make([]byte, 0, 16384)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf, err = ioutil.ReadAll(conn)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// time out -- this is OK
			err = nil
		} else {
			log.Println("read error:", err)
			// some error else, do something else, for example create new conn
		}
	}

	if err == nil {
		i := bytes.Index(buf, []byte("<?xml"))
		if i >= 0 {
			data := &Status{}
			err = xml.Unmarshal(buf[i:], data)
			if err == nil {
				m.LastGoodStatus = *data
				m.HaveGoodStatus = true
				m.LastGoodFetchTime = time.Now()
			}
		} else {
			err = fmt.Errorf("XML Parsing Error")
		}
	}
	m.LastFetchError = err
	return err
}

//StatusHandler returns the full Drobo status object as JSON
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	err := m.updateStatus()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Could not get Status!"))
		return
	}
	j, err := json.Marshal(m.LastGoodStatus)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Error marshalling json!"))
		return
	}
	fmt.Fprintf(w, "%s\n", string(j))

}

// HealthHandler returns a health status object
// See https://github.com/droboports/droboports.github.io/wiki/NASD-XML-format
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	err := m.updateStatus()
	w.Header().Add("Content-Type", "application/health+json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{ \"status\" : \"fail\",\"notes\":\"Could not get status\"}"))
		return
	}
	switch m.LastGoodStatus.Status {
	case 32768:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ \"status\" : \"pass\"}"))
	case 32772, 32774:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ \"status\" : \"warn\"}"))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{ \"status\" : \"fail\",\"notes\":\"Could not get status\"}"))
	}
	return
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the Drobo health monitoring service",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/v1/drobomon/health", HealthHandler)
		http.HandleFunc("/v1/drobomon/status", StatusHandler)
		http.ListenAndServe(":"+strconv.Itoa(svcport), nil)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	//serveCmd.MarkPersistentFlagRequired("drobo")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
