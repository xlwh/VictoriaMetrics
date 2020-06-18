package opentsdbhttp

import (
	"fmt"
	"net/http"

	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert/common"
	parser "github.com/VictoriaMetrics/VictoriaMetrics/lib/protoparser/opentsdbhttp"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/writeconcurrencylimiter"
	"github.com/VictoriaMetrics/metrics"
)

var (
	rowsInserted  = metrics.NewCounter(`vm_rows_inserted_total{type="opentsdbhttp"}`)
	rowsPerInsert = metrics.NewHistogram(`vm_rows_per_insert{type="opentsdbhttp"}`)
)

// InsertHandler processes HTTP OpenTSDB put requests.
// See http://opentsdb.net/docs/build/html/api_http/put.html
func InsertHandler(req *http.Request) error {
	path := req.URL.Path
	switch path {
	case "/api/put":
		// 写入并发度限制
		return writeconcurrencylimiter.Do(func() error {
			return parser.ParseStream(req, insertRows)
		})
	default:
		return fmt.Errorf("unexpected path requested on HTTP OpenTSDB server: %q", path)
	}
}

func insertRows(rows []parser.Row) error {
	// 对象池
	ctx := common.GetInsertCtx()
	defer common.PutInsertCtx(ctx)

	ctx.Reset(len(rows))
	for i := range rows {
		r := &rows[i]
		ctx.Labels = ctx.Labels[:0]
		// 组装tag，把字符串都转换成为[]byte，减少对象和gc压力
		ctx.AddLabel("", r.Metric)
		for j := range r.Tags {
			// 这里为啥取tag指针？避免拷贝嘛？
			tag := &r.Tags[j]
			ctx.AddLabel(tag.Key, tag.Value)
		}
		// 写入数据点
		ctx.WriteDataPoint(nil, ctx.Labels, r.Timestamp, r.Value)
	}
	rowsInserted.Add(len(rows))
	rowsPerInsert.Update(float64(len(rows)))
	return ctx.FlushBufs()
}
