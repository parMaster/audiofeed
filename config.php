<?php
define(DIR_ROOT,        dirname(__FILE__).'/');
define(DIR_SYSTEM,      DIR_ROOT.'system/');
define(DIR_DATABASE,    DIR_SYSTEM.'database/');
define(DIR_DATA,        DIR_ROOT.'data/');
define(DIR_AUDIO,       DIR_ROOT.'audio/');
# Configurable:
define(HTTP_ADDRESS,    'http://books.mydomain.com/');
define(HTTP_AUDIO,      HTTP_ADDRESS.'audio/');
define(AUTHOR_NAME,     'Arrrrr');
define(AUTHOR_EMAIL,    'Arrrrr@gmail.com');




function z2($var, $trace = true) {

//	if ($_SERVER['HTTP_X_REAL_IP'] == '')
	{
		if (php_sapi_name() != "cli") {
			?><pre style="text-align:left; background: #FFFFFF"><?
		} else {
			$trace = false;
		}
		var_dump($var);

		if ($trace) {
			foreach((debug_backtrace()) as $stack)
			{
				if (isset($stack['file']) && isset($stack['line'])) {
					echo $stack['file'].'['.$stack['line'].']<br>';
				}
			}
			echo '<br>';
		}

		if (php_sapi_name() != "cli") {
			?></pre><?
		}
	}
}
