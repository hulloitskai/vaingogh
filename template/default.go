package template

const defaultTemplate = `<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <meta name="go-import" content="{{ .Prefix }} {{ .VCSType }} {{ .ImportURL }}">
    <meta name="go-source" content="{{ .Prefix }} {{ .SourceURL }} {{ .SourceTreeURL }} {{ .SourceBlobURL }}">
    <meta http-equiv="refresh" content="0; url=https://godoc.org/{{ .Address }}">
  </head>
  <body>
    Nothing to see here; <a href="https://godoc.org/{{ .Address }}">move along</a>.
  </body>
</html>
`
