package parser

import (
	"regexp"
	"strings"
	"testing"

	"github.com/qiniu/logkit/conf"
	"github.com/qiniu/logkit/times"
	. "github.com/qiniu/logkit/utils/models"

	"github.com/stretchr/testify/assert"
)

func Test_QiniuLogRegex(t *testing.T) {

	tests := []struct {
		line string
		exp  bool
	}{
		{
			line: "2017/03/28 15:41:06 [Wm0AAPg-IUMW-68U][INFO] bdc.go:573: deleted: 67608",
			exp:  true,
		},
		{
			line: "2016/10/20 17:30:21.433423 [GE2owHck-Y4IWJHS][WARN] github.com/qiniu/http/rpcutil.v1/rpc_util.go:203:  ==> qiniu.com/streaming.v2/apiserver.go:1367: E18102: The specified repo does not exist under the provided appid ~",
			exp:  true,
		},
		{
			line: `[GE2owHck-Y4IWJHS]{"error":"No 	 such \t entry","reqid":"","details":null,"code":612}`,
			exp: false,
		},
		{
			line: "2016/10/20 18:20:30.642666 [ERROR] github.com/qiniu/logkit/queue/disk.go:241: DISKQUEUE(stream_local_save): readOne() error",
			exp:  true,
		},
		{
			line: "2016/10/20 17:20:30.642666 [GE2owHck-Y4IWJHS][INFO] disk.go github.com/qiniu/logkit/queue/disk.go:241: hello",
			exp:  true,
		},
		{
			line: "2016-10-20 17:20:30.642666 [GE2owHck-Y4IWJHS][INFO] disk.go github.com/qiniu/logkit/queue/disk.go:241: hello",
			exp:  false,
		},
		{
			line: "hello",
			exp:  false,
		},
		{
			line: "1234567/12/12e ax.log go",
			exp:  false,
		},
	}
	mp := "[1-9]\\d{3}/[0-1]\\d/[0-3]\\d [0-2]\\d:[0-6]\\d:[0-6]\\d(\\.\\d{6})?"
	for _, ti := range tests {
		got, err := regexp.MatchString("^"+mp, ti.line)
		assert.NoError(t, err)
		assert.Equal(t, ti.exp, got)
	}

	PREFIX := "QINIU"

	tests2 := []struct {
		line string
		exp  bool
	}{
		{
			line: "QINIU 2017/03/28 15:41:06 [Wm0AAPg-IUMW-68U][INFO] bdc.go:573: deleted: 67608",
			exp:  true,
		},
		{
			line: "QINIU 2016/10/20 17:30:21.433423 [GE2owHck-Y4IWJHS][WARN] github.com/qiniu/http/rpcutil.v1/rpc_util.go:203:  ==> qiniu.com/streaming.v2/apiserver.go:1367: E18102: The specified repo does not exist under the provided appid ~",
			exp:  true,
		},
		{
			line: `[GE2owHck-Y4IWJHS]{"error":"No 	 such \t entry","reqid":"","details":null,"code":612}`,
			exp: false,
		},
		{
			line: "QINIU 2016/10/20 18:20:30.642666 [ERROR] github.com/qiniu/logkit/queue/disk.go:241: DISKQUEUE(stream_local_save): readOne() error",
			exp:  true,
		},
		{
			line: "QINIU 2016/10/20 17:20:30.642666 [GE2owHck-Y4IWJHS][INFO] disk.go github.com/qiniu/logkit/queue/disk.go:241: hello",
			exp:  true,
		},
		{
			line: "QINIU 2016-10-20 17:20:30.642666 [GE2owHck-Y4IWJHS][INFO] disk.go github.com/qiniu/logkit/queue/disk.go:241: hello",
			exp:  false,
		},
		{
			line: "hello",
			exp:  false,
		},
		{
			line: "1234567/12/12e ax.log go",
			exp:  false,
		},
	}
	for _, ti := range tests2 {
		got, err := regexp.MatchString("^"+PREFIX+" "+mp, ti.line)
		assert.NoError(t, err)
		assert.Equal(t, ti.exp, got)
	}
}

func Test_QiniulogParser(t *testing.T) {
	c := conf.MapConf{}
	c[KeyParserName] = "qiniulogparser"
	c[KeyParserType] = "qiniulog"
	c[KeyDisableRecordErrData] = "true"
	ps := NewParserRegistry()
	p, err := ps.NewLogParser(c)
	if err != nil {
		t.Error(err)
	}
	lines := []string{
		"2017/03/28 15:41:06 [Wm0AAPg-IUMW-68U][INFO] bdc.go:573: deleted: 67608",
		`2016/10/20 17:30:21.433423 [GE2owHck-Y4IWJHS][WARN] github.com/qiniu/http/rpcutil.v1/rpc_util.go:203:  ==> qiniu.com/streaming.v2/apiserver.go:1367: E18102: The specified repo does not exist under the provided appid ~
		[GE2owHck-Y4IWJHS]{"error":"No 	 such \t entry","reqid":"","details":null,"code":612}`,
		"2016/10/20 18:20:30.642666 [ERROR] github.com/qiniu/logkit/queue/disk.go:241: DISKQUEUE(stream_local_save): readOne() error",
		"2016/10/20 17:20:30.642666 [GE2owHck-Y4IWJHS][INFO] disk.go github.com/qiniu/logkit/queue/disk.go:241: hello",
		"",
	}
	dts, err := p.Parse(lines)
	if c, ok := err.(*StatsError); ok {
		err = c.ErrorDetail
		assert.Equal(t, int64(1), c.Errors)
	}
	if len(dts) != 4 {
		t.Fatalf("parse lines error expect 4 but %v", len(dts))
	}

	if dts[0]["reqid"] != "Wm0AAPg-IUMW-68U" {
		t.Errorf("parse reqid error exp Wm0AAPg-IUMW-68U but %v", dts[0]["reqid"])
	}
	if dts[1]["reqid"] != "GE2owHck-Y4IWJHS" {
		t.Errorf("parse reqid error exp GE2owHck-Y4IWJHS but %v", dts[1]["reqid"])
	}
	if dts[2]["reqid"] != "" {
		t.Errorf("parse reqid error exp  but %v", dts[2]["reqid"])
	}
	if dts[3]["reqid"] != "GE2owHck-Y4IWJHS" {
		t.Errorf("parse reqid error exp GE2owHck-Y4IWJHS but %v", dts[4]["reqid"])
	}
	_, zoneValue := times.GetTimeZone()
	exp := "2016/10/20 17:30:21.433423" + zoneValue
	if dts[1]["time"] != exp {
		t.Errorf("parse time error exp %v but %v", exp, dts[1]["time"])
	}
	rawlog := dts[1]["log"].(string)
	if strings.Contains(rawlog, "\n") || strings.Contains(rawlog, "\t") {
		t.Error("log should not contain \\n or \\t ")
	}
	newlines := []string{
		"2016/10/20 17:20:30.642666 [ERROR] disk.go github.com/qiniu/logkit/queue/disk.go:241: ",
		"2016/10/20 17:20:30.642662 [123][WARN] disk.go github.com/qiniu/logkit/queue/disk.go:241: 1",
	}
	dts, err = p.Parse(newlines)
	if c, ok := err.(*StatsError); ok {
		err = c.ErrorDetail
	}
	if err != nil {
		t.Error(err)
	}
	if len(dts) != 2 {
		t.Fatalf("parse lines error expect 2 but %v", len(dts))
	}
	if dts[0]["level"] != "ERROR" {
		t.Errorf("parse level error exp ERROR but %v", dts[0]["level"])
	}
	if dts[1]["level"] != "WARN" {
		t.Errorf("parse level error exp WARN but %v", dts[1]["level"])
	}
	if dts[1]["file"] != "disk.go github.com/qiniu/logkit/queue/disk.go:241:" {
		t.Errorf("parse level error exp disk.go github.com/qiniu/logkit/queue/disk.go:241: but %v", dts[0]["file"])
	}
	assert.EqualValues(t, "qiniulogparser", p.Name())
}

func Test_QiniulogParserForErrData(t *testing.T) {
	c := conf.MapConf{}
	c[KeyParserName] = "qiniulogparser"
	c[KeyParserType] = "qiniulog"
	c[KeyDisableRecordErrData] = "false"
	ps := NewParserRegistry()
	p, err := ps.NewLogParser(c)
	if err != nil {
		t.Error(err)
	}
	lines := []string{
		"2017/03/28 15:41:06 [Wm0AAPg-IUMW-68U][INFO] bdc.go:573: deleted: 67608",
		"",
	}
	dts, err := p.Parse(lines)
	if c, ok := err.(*StatsError); ok {
		err = c.ErrorDetail
		assert.Equal(t, int64(1), c.Errors)
	}
	if len(dts) != 2 {
		t.Fatalf("parse lines error, expect 2 but %v", len(dts))
	}

	if dts[0]["reqid"] != "Wm0AAPg-IUMW-68U" {
		t.Errorf("parse reqid error exp Wm0AAPg-IUMW-68U but %v", dts[0]["reqid"])
	}
	assert.EqualValues(t, "qiniulogparser", p.Name())
}

func Test_QiniulogParserForTeapot(t *testing.T) {
	c := conf.MapConf{}
	c[KeyParserType] = "qiniulog"
	c[KeyLogHeaders] = "prefix,date,time,level,reqid,file"
	ps := NewParserRegistry()
	p, err := ps.NewLogParser(c)
	if err != nil {
		t.Error(err)
	}

	lines := []string{
		`2017/01/22 11:16:08.885550 [INFO][2pyKMgVp5EKg-ZsU]["github.com/teapots/request-logger/logger.go:75"] [REQ_END] 200 0.010k 3.792ms
		[WARN][SLdoIrCDZj7pmZsU]["qiniu.io/gaea/app/job/freeze.go:37"] <job.freezeDeamon> pop() failed: not found`,
		`2017/01/22 11:16:08.883870 [ERROR]["qiniu.io/gaea/app/providers/admin_login/admin_login.go:29"] current uid: 74121669`,
		`2017/01/22 11:15:54.947217 [INFO][2pyKMukqvwSd-ZsU]["qbox.us/biz/component/client/transport.go:109"] Service: POST 10.200.20.25:9100/user/info, Code: 200, Xlog: AC, Time: 1ms`,
	}

	dts, err := p.Parse(lines)
	if c, ok := err.(*StatsError); ok {
		err = c.ErrorDetail
	}
	if err != nil {
		t.Error(err)
	}

	if len(dts) != 3 {
		t.Fatalf("parse lines error expect 3 but %v", len(dts))
	}

	if dts[0]["reqid"] != "2pyKMgVp5EKg-ZsU" {
		t.Errorf("parse reqid error exp 2pyKMgVp5EKg-ZsU but %v", dts[0]["reqid"])
	}

	if dts[1]["reqid"] != "" {
		t.Errorf("parse reqid error exp  but %v", dts[1]["reqid"])
	}

	_, zoneValue := times.GetTimeZone()
	exp := "2017/01/22 11:16:08.885550" + zoneValue
	if dts[0]["time"] != exp {
		t.Errorf("parse time error exp %v but %v", exp, dts[0]["time"])
	}

	newlines := []string{
		`2017/01/22 12:14:14.072180 [ERROR][SLdoIlbiqLnL_JsU]["github.com/teapots/request-logger/logger.go:61"] hello`,
		`2017/01/22 12:12:10.065824 [WARN][SLdoIrCDZj7pmZsU]["qiniu.io/gaea/app/job/freeze.go:37"] <job.freezeDeamon> pop() failed: not found`,
	}

	dts, err = p.Parse(newlines)
	if c, ok := err.(*StatsError); ok {
		err = c.ErrorDetail
	}
	if err != nil {
		t.Error(err)
	}
	if len(dts) != 2 {
		t.Fatalf("parse lines error expect 2 but %v", len(dts))
	}
	if dts[0]["level"] != "ERROR" {
		t.Errorf("parse level error exp ERROR but %v", dts[0]["level"])
	}
	if dts[1]["level"] != "WARN" {
		t.Errorf("parse level error exp WARN but %v", dts[0]["level"])
	}
	assert.Equal(t, `"github.com/teapots/request-logger/logger.go:61"`, dts[0]["file"])
}
