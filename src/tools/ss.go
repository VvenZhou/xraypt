package tools

import (

)

type SSShare struct{
	method string
	pwd string
	addr string
	port int
}

type SSUser struct{
	addr string `json:address`
	port int `json:port`
	method string `json:method`
	pwd string `json:password`
}

func SSLinkToOut(ss *Outbound, ssShareLink string){

}
