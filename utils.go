package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
)

func getUuid() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func MapToJson(param map[string]string) string {
	dataType, _ := json.Marshal(param)
	dataString := string(dataType)
	return dataString
}

func parseValues(m1 map[string]string, values map[string][]string) {
	for i := range values {
		vs := values[i]
		if len(vs) == 0 {
			m1[i] = ""
		} else {
			m1[i] = vs[0]
		}
	}
}
