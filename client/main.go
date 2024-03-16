package main

import (
	"bufio"
	"io"
	"net/http"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

func main() {
	tracer, closer, _ := CreateTracer("UserinfoService")
	// 创建第一个 span A
	parentSpan := tracer.StartSpan("A")
	// 调用其它服务
	GetUserInfo(tracer, parentSpan)
	// 结束 A
	parentSpan.Finish()
	// 结束当前 tracer
	closer.Close()

	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadByte()
}

func CreateTracer(servieName string) (opentracing.Tracer, io.Closer, error) {
	var cfg = jaegercfg.Configuration{
		ServiceName: servieName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
			// 按实际情况替换你的 ip
			// CollectorEndpoint: "http://10.14.2.12:14268/api/traces",
			LocalAgentHostPort: "10.14.2.12:6831", //agent 地址

		},
	}

	jLogger := jaegerlog.StdLogger
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
	)
	return tracer, closer, err
}

// 请求远程服务，获得用户信息
func GetUserInfo(tracer opentracing.Tracer, parentSpan opentracing.Span) {
	// 继承上下文关系，创建子 span
	childSpan := tracer.StartSpan(
		"B",
		opentracing.ChildOf(parentSpan.Context()),
	)

	url := "http://127.0.0.1:8081/Get?username=痴者工良"
	req, _ := http.NewRequest("GET", url, nil)
	// 设置 tag，这个 tag 我们后面讲 span的tag
	ext.SpanKindRPCClient.Set(childSpan)
	ext.HTTPUrl.Set(childSpan, url)
	ext.HTTPMethod.Set(childSpan, "GET")
	tracer.Inject(childSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	resp, _ := http.DefaultClient.Do(req)
	_ = resp // 丢掉
	defer childSpan.Finish()
}
