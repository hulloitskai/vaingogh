package template

const defaultRawTpl = `<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <meta name="go-import" content="{{ .Prefix }} git {{ .ImportURL }}">
    <meta name="go-source" content="{{ .Prefix }} {{ .SourceURL }} {{ .SourceURL }}/tree/master{/dir} {{ .SourceURL }}/blob/master{/dir}/{file}#L{line}">
    <meta http-equiv="refresh" content="0; url=https://godoc.org/{{ .Address }}">
  </head>
  <body>
    Nothing to see here; <a href="https://godoc.org/{{ .Address }}">move along</a>.
  </body>
</html>
`
