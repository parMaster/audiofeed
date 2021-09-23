<?php
require_once 'config.php';

$route = strval($_GET['route']);
$routeParts = explode('/', $route);

if ($routeParts[0]) {

	if ('index' == $routeParts[0]) {
		$index = glob(DIR_AUDIO.'*');
		foreach ($index AS $indexItem) {
			$fileNameParts = explode('/', $indexItem);
			$folder = $fileNameParts[sizeof($fileNameParts)-1];
			$fileUrl = HTTP_ADDRESS.$folder;
			echo '<a href="'.$fileUrl.'">'.$fileUrl.'<br>';
		}
		die();
	}

	$images = array();
	$folder = $routeParts[0];
	if (file_exists(DIR_AUDIO.$folder) && is_dir(DIR_AUDIO.$folder)) {

		$files = (glob(DIR_AUDIO.$folder.'/*.{mp3,m4b,mp4}', GLOB_BRACE));

		if ($images = glob(DIR_AUDIO.$folder.'/*.{jpg,jpeg,png}', GLOB_BRACE)) {
			$images = explode('/', $images[0]);
			$image = $images[sizeof($images)-1];
			$imageUrl = HTTP_AUDIO.$folder.'/'.$image;
		}

		if (sizeof($files)) {

			header('Content-Type: application/xml');

			echo '<?xml version="1.0" encoding="UTF-8"?>';
			?><rss version="2.0"
		           xmlns:content="http://purl.org/rss/1.0/modules/content/"
		           xmlns:wfw="http://wellformedweb.org/CommentAPI/"
		           xmlns:dc="http://purl.org/dc/elements/1.1/"
		           xmlns:atom="http://www.w3.org/2005/Atom"
		           xmlns:sy="http://purl.org/rss/1.0/modules/syndication/"
		           xmlns:slash="http://purl.org/rss/1.0/modules/slash/"
		           xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd"
		           xmlns:rawvoice="http://www.rawvoice.com/rawvoiceRssModule/"
			>
				<channel>
					<title><?=$folder?></title>
					<atom:link href="<?=HTTP_ADDRESS.$folder?>" rel="self" type="application/rss+xml" />
					<link><?=HTTP_ADDRESS.$folder?></link>
					<description><?=$folder?></description>
					<language>en-US</language>
					<sy:updatePeriod>hourly</sy:updatePeriod>
					<sy:updateFrequency>1</sy:updateFrequency>
					<itunes:summary>Audiobooks hosted feed</itunes:summary>
					<itunes:author><?=AUTHOR_NAME?></itunes:author>
					<itunes:owner>
						<itunes:name><?=AUTHOR_NAME?></itunes:name>
						<itunes:email><?=AUTHOR_EMAIL?></itunes:email>
					</itunes:owner>
					<itunes:subtitle><?=$folder?></itunes:subtitle>
					<itunes:category text="AudioBooks" />
					<? if ($imageUrl) { ?>
						<itunes:image href="<?=$imageUrl?>" />
					<? } ?>
					<image>
						<title><?=$folder?></title>
						<url><?=$imageUrl?></url>
						<link><?=HTTP_ADDRESS?></link>
					</image>
					<rawvoice:rating>TV-MA</rawvoice:rating>
					<rawvoice:location></rawvoice:location>
					<rawvoice:frequency>Weekly</rawvoice:frequency>
					<?

						foreach ($files AS $file) {
							$fileNameParts = explode('/', $file);
							$fileName = $fileNameParts[sizeof($fileNameParts)-1];
							$fileUrl = HTTP_AUDIO.$folder.'/'.($fileName);

							?>
							<item>
								<title><?=$fileName?></title>
								<link><?=HTTP_ADDRESS?></link>
								<comments><?=$fileName?></comments>
								<enclosure url="<?=($fileUrl)?>" type="audio/mpeg" />
							</item>
							<?
						}
					?>
				</channel>
			</rss>
			<?
		}
	}
}

