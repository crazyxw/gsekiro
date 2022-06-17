package main

import "encoding/json"

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (r *Response) toJson() []byte {
	result, err := json.Marshal(&r)
	if err != nil {
		return []byte{}
	}
	return result
}
