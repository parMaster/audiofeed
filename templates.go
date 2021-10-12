package main

const xmlTemplateBody = `<?xml version="1.0" ?>
<?xml-stylesheet href="/feed.xsl" type="text/xsl"?>
<rss xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd" xmlns:media="http://search.yahoo.com/mrss/" xmlns:creativeCommons="http://backend.userland.com/creativeCommonsRssModule" xmlns:atom="http://www.w3.org/2005/Atom" version="2.0">
	<channel>
		<title><![CDATA[{{.Title.Name}}]]></title>
		<link><![CDATA[http://{{.HostName}}/{{.Title.Path}}]]></link>
		<atom:link rel="self" href="http://{{.HostName}}/{{.Title.Path}}" />
		<description><![CDATA[{{.Title.Name}}]]></description>
		<itunes:type>serial</itunes:type>
		<itunes:author>AudioFeed</itunes:author>
		<itunes:summary><![CDATA[{{.Title.Name}}]]></itunes:summary>
		<itunes:owner>
		<itunes:name>Audiofeed</itunes:name>
		<itunes:email>af@af.com</itunes:email>
		</itunes:owner>
		<itunes:category text="AudioBooks" />

		{{if .Title.CoverPath}}
		<itunes:image href="http://{{.HostName}}/{{.Title.CoverPath}}" />
		<media:thumbnail url="http://{{.HostName}}/{{.Title.CoverPath}}" />
		<image>
			<title>{{.Title.Name}}</title>
			<url>http://{{.HostName}}/{{.Title.CoverPath}}</url>
			<link>http://{{.HostName}}/{{.Title.Path}}</link>
		</image>
		{{end}}

		{{ $TitleName := .Title.Name }}
		{{ $HostName := .HostName }}
		{{ $TitlePath := .Title.Path }}
		{{ range $k, $v := .Title.Chapters }}
		<item>
			<title><![CDATA[{{$TitleName}}]]></title>
			<itunes:episode><![CDATA[{{$k}}]]></itunes:episode>
			<link><![CDATA[http://{{$HostName}}/{{$TitlePath}}]]></link>
			<enclosure url="http://{{$HostName}}/{{$v}}" length="0" type="audio/mpeg" />
			<itunes:explicit>No</itunes:explicit>
			<itunes:block>No</itunes:block>
			<media:content url="http://{{$HostName}}/{{$v}}" type="audio/mpeg" />
		</item>
		{{ end }} 

	</channel>
</rss>`

const titlesTemplateBody = `<html>
<body>
<h2>Audiobook feed</h3>
<h3>Titles available:</h4>
{{ range $k, $v := . }}
<li><a href="/title/{{$v}}">{{$v}}</a></li>
{{ end }} 
</body>
</html>`
