package main

const xmlTemplateBody = `<?xml version="1.0" ?>
<?xml-stylesheet href="/feed.xsl" type="text/xsl"?>
<rss xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd" xmlns:media="http://search.yahoo.com/mrss/" xmlns:creativeCommons="http://backend.userland.com/creativeCommonsRssModule" xmlns:atom="http://www.w3.org/2005/Atom" version="2.0">
	<channel>
		<title><![CDATA[{{.TitleName}}]]></title>
		<link><![CDATA[{{.TitleURL}}]]></link>
		<atom:link rel="self" href="{{.TitleURL}}" />
		<description><![CDATA[{{.TitleName}}]]></description>
		<itunes:type>serial</itunes:type>
		<itunes:author>AudioFeed</itunes:author>
		<itunes:summary><![CDATA[{{.TitleName}}]]></itunes:summary>
		<itunes:owner>
		<itunes:name>Audiofeed</itunes:name>
		<itunes:email>af@af.com</itunes:email>
		</itunes:owner>
		<itunes:category text="AudioBooks" />

		{{if .CoverURL}}
		<itunes:image href="{{.CoverURL}}" />
		<media:thumbnail url="{{.CoverURL}}" />
		<image>
			<title>{{.TitleName}}</title>
			<url>{{.CoverURL}}</url>
			<link>{{.TitleURL}}</link>
		</image>
		{{end}}

		{{ $TitleName := .TitleName }}
		{{ $TitleURL := .TitleURL }}
		{{ range $k, $v := .ChapterURLs }}
		<item>
			<title><![CDATA[{{$TitleName}}]]></title>
			<itunes:episode><![CDATA[{{$k}}]]></itunes:episode>
			<link><![CDATA[{{$TitleURL}}]]></link>
			<enclosure url="{{$v}}" length="0" type="audio/mpeg" />
			<itunes:explicit>No</itunes:explicit>
			<itunes:block>No</itunes:block>
			<media:content url="{{$v}}" type="audio/mpeg" />
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
