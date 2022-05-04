package atto

import "net/http"

// The RequestInterceptor is a function that is used to modify all HTTP
// requests that are sent to a node. If it is nil, requests are not
// modified.
//
// May be used, for example, to authenticate requests.
var RequestInterceptor func(request *http.Request) error
