package netsdk

// #cgo CFLAGS: -I ../include
// #cgo LDFLAGS: -L ../lib -lhcnetsdk

// #include <stdio.h>
// #include <stdlib.h>
// #include<string.h>
// #include "HCNetSDK.h"
import "C"

import (
	"log"
	"unsafe"

	"github.com/kr/pretty"
	"github.com/mattn/go-pointer"
)

//export export_MessageCallabck
func export_MessageCallabck(cmd int, alarm *C.NET_DVR_ALARMER, pBuf *C.char, l C.DWORD, userData C.long) {
	var (
		pAlarm *NET_DVR_ALARMER
		client *Client
		ok     bool
	)
	pAlarm = (*NET_DVR_ALARMER)(unsafe.Pointer(alarm))
	log.Printf("userdata %d", userData)

	if userData == 0 {
		return
	}

	if pAlarm.ST_byUserIDValid > 0 {
		clientsLock.Lock()
		if client, ok = clientsMap[int64(pAlarm.ST_lUserID)]; !ok {
			log.Printf("can;t get client user id %d", pAlarm.ST_lUserID)
		}
		clientsLock.Unlock()
	}

	v := pointer.Restore(unsafe.Pointer(uintptr(userData)))
	log.Printf("v %v", v)
	if visitor, ok := v.(MessageCallbackVisitor); ok {
		ccmd := CommAlarm(cmd)

		switch ccmd {
		case CommAlarmTfs:
			var (
				infol   = unsafe.Sizeof(NET_DVR_TFS_ALARM{})
				tfsInfo = NET_DVR_TFS_ALARM{
					ST_dwSize: uint32(unsafe.Sizeof(NET_DVR_TFS_ALARM{})),
				}
			)
			log.Printf("size %d == %d", unsafe.Sizeof(NET_DVR_TFS_ALARM{}), uint32(l))
			// info := (*C.NET_DVR_TFS_ALARM)(unsafe.Pointer(pBuf))
			C.memcpy(unsafe.Pointer(&tfsInfo), unsafe.Pointer(pBuf), C.ulong(infol))
			if client != nil && client.msgCb != nil {
				client.msgCb(ccmd, client, pAlarm, &tfsInfo)
			}

			if visitor.Callback != nil {
				visitor.Callback(ccmd, client, pAlarm, &tfsInfo)
			}
		case CommGisinfoUpload:
			var (
				gisInfo = NET_DVR_GIS_UPLOADINFO{
					ST_dwSize: uint32(unsafe.Sizeof(NET_DVR_GIS_UPLOADINFO{})),
				}
			)
			log.Printf("size %d == %d", unsafe.Sizeof(NET_DVR_GIS_UPLOADINFO{}), l)
			C.memcpy(unsafe.Pointer(&gisInfo), unsafe.Pointer(pBuf), C.ulong(l))
			if client != nil && client.msgCb != nil {
				client.msgCb(ccmd, client, pAlarm, &gisInfo)
			}

			if visitor.Callback != nil {
				visitor.Callback(ccmd, client, pAlarm, &gisInfo)
			}
		case CommAlarmAidV41:
		default:
			log.Printf("cmd 0x%0X", cmd)
			log.Printf("alarm % #v", pretty.Formatter(pAlarm))
		}
	} else {
		log.Printf("invalid messageCallabckVisitor")
	}

}

//export export_ExceptionCallBack
func export_ExceptionCallBack(dwType C.DWORD, lUserID C.LONG, lHandle C.LONG, pUser unsafe.Pointer) C.BOOL {
	log.Println("Enter fExceptionCallBack...")
	v := pointer.Restore(pUser)
	visitor, ok := v.(ExceptionVisitor)
	if !ok {
		return 0
	}

	b := visitor.Callback(int(dwType), visitor.UserData)
	if b {
		return 1
	} else {
		return 0
	}
}
