package vue

import (
	"html/template"
	"net/http"
)

var page = template.Must(template.New("eventackle").Parse(`<!DOCTYPE html>
<html>
<head>
	<meta charset=utf-8/>
	<meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
	<title>{{.title}}</title>
</head>
<body>
<style type="text/css">
	html { font-family: "Open Sans", sans-serif; overflow: hidden; }
	body { margin: 0; background: white; }
</style>
<div id="app">Vue app here!</div>
<script type="text/javascript">
	window.addEventListener('load', function (event) {
		// any js stuff 
	})
</script>
</body>
</html>
`))

func Index(title string, endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := page.Execute(w, map[string]string{
			"title":    title,
			"endpoint": endpoint,
			"version":  "1.4.3",
		})
		if err != nil {
			panic(err)
		}
	}
}
