package upstream

import (
	"bytes"
	"coastline/consts"
	"coastline/ctx"
	"coastline/safeutil"
	"coastline/tlog"
	"coastline/vconfig"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"
)

type RespCode int

const (
	OK          RespCode = 1
	BadUpstream RespCode = 2
	LocalError  RespCode = 3
)

var httpClient *http.Client

func init() {
	httpClient = initHttpClient()
	//httpClient = initHttp2Client()
}

func initHttpClient() *http.Client {
	c := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, 1*time.Second)
			},
			MaxIdleConnsPerHost:   32,
			MaxConnsPerHost:       500,
			IdleConnTimeout:       10 * time.Minute,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 2 * time.Second,
		},
		Timeout: 20 * time.Second,
	}
	fmt.Println("init http_1_1 client successfully")
	return &c
}

func initHttp2Client() *http.Client {
	c := http.Client{
		Transport: &http2.Transport{
			// So http2.Transport doesn't complain the URL scheme isn't 'https'
			AllowHTTP: true,
			// Pretend we are dialing a TLS endpoint. (Note, we ignore the passed tls.Config)
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
		Timeout: time.Duration(20) * time.Second,
	}

	fmt.Println("init http_2 client successfully")
	return &c
}

func Invoke(gc *gin.Context) (*http.Response, RespCode) {
	c := ctx.DetachFrom(gc)

	var req *http.Request
	var err error
	if isMultipart(gc) {
		req, err = buildMultipartReq(gc)
	} else {
		req, err = buildDefaultReq(gc)
	}

	if err != nil {
		c.Errorln("build request err:", err)
		return nil, LocalError
	}
	c.Infof("upstream request headers: %v\n", req.Header)

	start := time.Now()
	resp, err := httpClient.Do(req)
	c.Infof("invoke upstream cost:[%d]ms", time.Now().UnixMilli()-start.UnixMilli())

	if err != nil {
		c.Errorln("invoke err:", err)
		return nil, BadUpstream
	}

	return resp, OK
}

func buildMultipartReq(gc *gin.Context) (*http.Request, error) {
	c := ctx.DetachFrom(gc)

	type part struct {
		io.Reader
		filename string
	}

	m := make(map[string]part)

	var err error
	for k, v := range gc.Request.MultipartForm.File {
		file, err := v[0].Open()
		if err != nil {
			c.Errorln("open file err:", err)
		} else {
			m[k] = part{
				Reader:   file,
				filename: v[0].Filename,
			}
		}
	}

	for k, v := range gc.Request.MultipartForm.Value {
		m[k] = part{
			Reader:   strings.NewReader(v[0]),
			filename: "",
		}
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for key, pt := range m {
		var fw io.Writer
		if x, ok := pt.Reader.(io.Closer); ok {
			defer x.Close()
		}

		// Add string part
		if _, ok := pt.Reader.(*strings.Reader); ok {
			if fw, err = w.CreateFormField(key); err != nil {
				return nil, err
			}
		} else {
			// Add file part
			if fw, err = w.CreateFormFile(key, pt.filename); err != nil {
				return nil, err
			}

		}
		if written, err := io.Copy(fw, pt); err != nil {
			return nil, err
		} else {
			c.Infof("written bytes[%d]", written)
		}
	}

	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	req, err := http.NewRequest("POST", buildUrl(gc), &b)
	if err != nil {
		return nil, err
	}

	for _, v := range ctx.DetachFrom(gc).Headers {
		req.Header[v.Key] = []string{v.Val}
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, nil
}

func isMultipart(gc *gin.Context) bool {
	return gc.Request.ParseMultipartForm(8<<20) == nil
}

func Auth(token string, c *ctx.Ctx) (*Result[AuthInfo], RespCode) {
	body := make(map[string]string)
	body[consts.HeaderToken] = token

	bs, err := json.Marshal(body)
	if err != nil {
		c.Errorln("marshal body err:", err)
		return nil, LocalError
	}

	var req *http.Request
	if tlog.IsDebug() {
		traceCtx := httptrace.WithClientTrace(context.Background(), newClientTrace(c))
		req, err = http.NewRequestWithContext(traceCtx, "POST", vconfig.Upstream().UrlAuth(), bytes.NewReader(bs))
	} else {
		req, err = http.NewRequest("POST", vconfig.Upstream().UrlAuth(), bytes.NewReader(bs))
	}

	if err != nil {
		c.Errorln("res new request err:", err)
		return nil, LocalError
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header[consts.HeaderTraceId] = []string{c.TraceId}

	start := time.Now()
	resp, err := httpClient.Do(req)
	c.Infof("upstream auth cost:[%d]ms", time.Now().UnixMilli()-start.UnixMilli())

	if err != nil {
		c.Errorln("auth request err:", err)
		return nil, BadUpstream
	}

	authInfo := NewAuthInfo()
	err = safeutil.Unmarshal(resp, authInfo)
	if err != nil {
		c.Errorln("resp unmarshal err:", err)
		return nil, BadUpstream
	}

	if authRespMissingData(authInfo) {
		return nil, BadUpstream
	}

	return authInfo, OK
}

func GetUserInfo(uid string, c *ctx.Ctx) (*Result[UserInfo], RespCode) {
	var req *http.Request
	var err error
	if tlog.IsDebug() {
		traceCtx := httptrace.WithClientTrace(context.Background(), newClientTrace(c))
		req, err = http.NewRequestWithContext(traceCtx, "GET",
			strings.Replace(vconfig.Upstream().UrlUserInfo(), "{uid}", uid, 1), nil)
	} else {
		req, err = http.NewRequest("GET",
			strings.Replace(vconfig.Upstream().UrlUserInfo(), "{uid}", uid, 1), nil)
	}

	if err != nil {
		c.Errorln("get user info request err: ", err)
		return nil, LocalError
	}

	req.Header[consts.HeaderTraceId] = []string{c.TraceId}
	req.Header[consts.HeaderUid] = []string{uid}

	start := time.Now()
	resp, err := httpClient.Do(req)
	c.Infof("upstream get user info cost:[%d]ms", time.Now().UnixMilli()-start.UnixMilli())

	if err != nil {
		c.Errorln("get user info err: ", err)
		return nil, BadUpstream
	}

	userInfo := NewUserInfo()
	err = safeutil.Unmarshal(resp, userInfo)
	if err != nil {
		c.Errorln("unmarshal err: ", err)
		return nil, BadUpstream
	}

	return userInfo, OK
}

func authRespMissingData(info *Result[AuthInfo]) bool {
	if info == nil {
		return true
	}
	if info.Success {
		if info.Data.UserInfo == nil {
			return true
		}
		if info.Data.TokenInfo == nil {
			return true
		}
	}
	return false
}

func buildDefaultReq(gc *gin.Context) (*http.Request, error) {
	c := ctx.DetachFrom(gc)

	var body io.Reader
	if isGetOrHead(gc) {
		body = nil
	} else {
		bs, err := io.ReadAll(gc.Request.Body)
		c.Infof("post request body: %s", bs)
		if err != nil {
			return nil, err
		}
		defer func() {
			if gc.Request.Body != nil {
				gc.Request.Body.Close()
			}
		}()
		body = bytes.NewReader(bs)
	}

	var req *http.Request
	var err error
	if tlog.IsDebug() {
		traceCtx := httptrace.WithClientTrace(context.Background(), newClientTrace(c))
		req, err = http.NewRequestWithContext(traceCtx, gc.Request.Method, buildUrl(gc), body)
	} else {
		req, err = http.NewRequest(gc.Request.Method, buildUrl(gc), body)
	}

	if err != nil {
		c.Errorln("new request err:", err)
		return nil, err
	}

	for _, v := range ctx.DetachFrom(gc).Headers {
		req.Header[v.Key] = []string{v.Val}
	}
	return req, nil
}

func isGetOrHead(gc *gin.Context) bool {
	return strings.EqualFold(gc.Request.Method, "GET") ||
		strings.EqualFold(gc.Request.Method, "HEAD")
}

func buildUrl(gc *gin.Context) string {
	c := ctx.DetachFrom(gc)

	//log lookup IP
	ip, err := net.LookupIP(c.Route.ServiceName)
	if err != nil {
		c.Errorf("lookup ip err host: %s, err: %v", c.Route.ServiceName, err)
	} else {
		c.Infof("lookup ip host %s -> ip %v", c.Route.ServiceName, ip)
	}

	url := c.Route.UpstreamUrl
	query := gc.Request.URL.RawQuery
	if len(query) > 0 {
		url = url + "?" + query
	}

	c.Infoln("build upstream url:", url)
	return url
}

func newClientTrace(c *ctx.Ctx) *httptrace.ClientTrace {
	cTrace := &httptrace.ClientTrace{}

	cTrace.GetConn = func(hostPort string) {
		c.Debugf("http trace GetConn %s", hostPort)
	}
	cTrace.GotConn = func(info httptrace.GotConnInfo) {
		c.Debugf("http trace GotConn %+v", info)
	}
	cTrace.GotFirstResponseByte = func() {
		c.Debugf("http trace GotFirstResponseByte")
	}
	cTrace.DNSStart = func(info httptrace.DNSStartInfo) {
		c.Debugf("http trace DNSStart %v", info)
	}
	cTrace.DNSDone = func(info httptrace.DNSDoneInfo) {
		c.Debugf("http trace DNSDone %+v", info)
	}
	cTrace.ConnectStart = func(network, addr string) {
		c.Debugf("http trace ConnectStart network %s | addr %s", network, addr)
	}
	cTrace.ConnectDone = func(network, addr string, err error) {
		c.Debugf("http trace ConnectDone network %s | addr %s | err %v", network, addr, err)
	}
	cTrace.WroteHeaderField = func(key string, value []string) {
		c.Debugf("http trace WroteHeaderField key %s val %v", key, value)
	}
	cTrace.WroteHeaders = func() {
		c.Debugf("http trace WroteHeaders")
	}
	cTrace.WroteRequest = func(info httptrace.WroteRequestInfo) {
		c.Debugf("http wrote request %v", info)
	}

	return cTrace
}
