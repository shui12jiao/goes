package rpc

import (
	"fmt"
	"html/template"
	"net/http"
)

const debugText = `<html>
	<body>
	<title>RPC Services</title>
	{{range .}}
	<hr>
	Service {{.Name}}
	<hr>
		<table>
		<th align=center>Method</th><th align=center>Calls</th>
		{{range $name, $mtype := .Method}}
			<tr>
			<td align=left font=fixed>{{$name}}({{$mtype.ArgType}}, {{$mtype.ReplyType}}) error</td>
			<td align=center>{{$mtype.NumCalls}}</td>
			</tr>
		{{end}}
		</table>
	{{end}}
	</body>
	</html>`

var debugTempl = template.Must(template.New("debug").Parse(debugText))

type debugHTTP struct {
	*Server
}

type debugService struct {
	Name   string
	Method map[string]*methodType
}

func (d debugHTTP) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	services := make([]debugService, 0)
	d.serviceMap.Range(func(namei, svci any) bool {
		services = append(services, debugService{
			Name:   namei.(string),
			Method: svci.(*service).method,
		})
		return true
	})
	err := debugTempl.Execute(w, services)
	if err != nil {
		fmt.Fprintln(w, "rpc: debugTemplate execute error:", err)
	}
}
