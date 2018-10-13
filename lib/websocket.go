package qniblib

import (
	"net/http"
	"strings"
)

func IsWebSocket(req *http.Request) bool {
	//log.Println("IsWebSocket called: ", req.URL.String())
	//log.Println("Connection", req.Header["Connection"])
	//log.Println("Upgrade:", req.Header["Upgrade"])

	conn_hdr := ""
	conn_hdrs := req.Header["Connection"]
	if len(conn_hdrs) > 0 {
		conn_hdr = conn_hdrs[0]
	}

	upgrade_websocket := false
	if strings.ToLower(conn_hdr) == "upgrade" {
		upgrade_hdrs := req.Header["Upgrade"]
		if len(upgrade_hdrs) > 0 {
			upgrade_websocket = (strings.ToLower(upgrade_hdrs[0]) == "websocket")
		}
	}

	return upgrade_websocket
}
