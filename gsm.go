// Author: Milan Nikolic <gen2brain@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gsm

// #cgo pkg-config: gammu
// #include <stdio.h>
// #include <gammu.h>
// extern void sendSMSCallback(GSM_StateMachine *sm, int status, int messageReference, void * user_data);
import "C"

import (
	"errors"
	"fmt"
	"log"
	"unsafe"
)

var smsSendStatus C.GSM_Error

const (
	ERR_NONE    = C.ERR_NONE
	ERR_UNKNOWN = C.ERR_UNKNOWN
	ERR_TIMEOUT = C.ERR_TIMEOUT
)

// Returns error message string
func errorString(e int) string {
	return C.GoString(C.GSM_ErrorString(C.GSM_Error(e)))
}

// Gammu GSM struct
type GSM struct {
	sm *C.GSM_StateMachine
}

// Returns new GSM
func NewGSM() (g *GSM, err error) {
	g = &GSM{}
	g.sm = C.GSM_AllocStateMachine()

	if g.sm == nil {
		err = errors.New("Cannot allocate state machine")
	}

	return
}

// Enables global debugging to stderr
func (g *GSM) EnableDebug() {
	debugInfo := C.GSM_GetGlobalDebug()
	C.GSM_SetDebugFileDescriptor(C.stderr, C.gboolean(1), debugInfo)
	C.GSM_SetDebugLevel(C.CString("textall"), debugInfo)
}

// Connects to phone
func (g *GSM) Connect() (err error) {
	e := C.GSM_InitConnection(g.sm, 1) // 1 means number of replies to wait for
	if e != ERR_NONE {
		err = errors.New(errorString(int(e)))
	}

	// set callback for message sending
	C.GSM_SetSendSMSStatusCallback(g.sm, (C.SendSMSStatusCallback)(unsafe.Pointer(C.sendSMSCallback)), nil)
	return
}

// Reads configuration file
func (g *GSM) SetConfig(config string) (err error) {
	path := C.CString(config)
	defer C.free(unsafe.Pointer(path))

	var cfg *C.INI_Section
	defer C.INI_Free(cfg)

	// find configuration file
	e := C.GSM_FindGammuRC(&cfg, path)
	if e != ERR_NONE {
		err = errors.New(errorString(int(e)))
		return
	}

	// read it
	e = C.GSM_ReadConfig(cfg, C.GSM_GetConfig(g.sm, 0), 0)
	if e != ERR_NONE {
		err = errors.New(errorString(int(e)))
		return
	}

	// we have one valid configuration
	C.GSM_SetConfigNum(g.sm, 1)
	return
}

// Sends message
func (g *GSM) SendSMS(text, number string) (err error) {
	var sms C.GSM_SMSMessage
	var smsc C.GSM_SMSC

	sms.PDU = C.SMS_Submit                           // submit message
	sms.UDH.Type = C.UDH_NoUDH                       // no UDH, just a plain message
	sms.Coding = C.SMS_Coding_Default_No_Compression // default coding for text
	sms.Class = 1                                    // class 1 message (normal)

	C.EncodeUnicode((*C.uchar)(unsafe.Pointer(&sms.Text)), C.CString(text), C.int(len(text)))
	C.EncodeUnicode((*C.uchar)(unsafe.Pointer(&sms.Number)), C.CString(number), C.int(len(number)))

	// we need to know SMSC number
	smsc.Location = 1
	e := C.GSM_GetSMSC(g.sm, &smsc)
	if e != ERR_NONE {
		err = errors.New(errorString(int(e)))
		return
	}

	// set SMSC number in message
	sms.SMSC.Number = smsc.Number

	// Set flag before callind SendSMS, some phones might give instant response
	smsSendStatus = ERR_TIMEOUT

	// send message
	e = C.GSM_SendSMS(g.sm, &sms)
	if e != ERR_NONE {
		err = errors.New(errorString(int(e)))
		return
	}

	// wait for network reply
	for {
		C.GSM_ReadDevice(g.sm, C.gboolean(1))
		if smsSendStatus == ERR_NONE {
			break
		}
		if smsSendStatus != ERR_TIMEOUT {
			err = errors.New(errorString(int(smsSendStatus)))
			break
		}
	}

	return
}

// Terminates connection and free memory
func (g *GSM) Terminate() (err error) {
	// terminate connection
	e := C.GSM_TerminateConnection(g.sm)
	if e != ERR_NONE {
		err = errors.New(errorString(int(e)))
	}

	// free up used memory
	C.GSM_FreeStateMachine(g.sm)
	return
}

// Checks if phone is connected
func (g *GSM) IsConnected() bool {
	return int(C.GSM_IsConnected(g.sm)) != 0
}

// Callback for message sending
//export sendSMSCallback
func sendSMSCallback(sm *C.GSM_StateMachine, status C.int, messageReference C.int, user_data unsafe.Pointer) {
	t := fmt.Sprintf("Sent SMS on device %s - ", C.GoString(C.GSM_GetConfig(sm, -1).Device))
	if int(status) == 0 {
		log.Printf(t + "OK\n")
		smsSendStatus = ERR_NONE
	} else {
		log.Printf(t+"ERROR %d\n", int(status))
		smsSendStatus = ERR_UNKNOWN
	}
}
