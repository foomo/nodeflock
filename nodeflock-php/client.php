<?php

/**
 * @link    www.foomo.org
 * @license www.gnu.org/licenses/lgpl.txt
 */
class Client {
	// --------------------------------------------------------------------------------------------
	// ~ Constants
	// --------------------------------------------------------------------------------------------

	const SOCKET_READ_WINDOW_SIZE = 8192;

	// --------------------------------------------------------------------------------------------
	// ~ Variables
	// --------------------------------------------------------------------------------------------

	/**
	 * @var string
	 */
	private $server;
	/**
	 * @var resource
	 */
	private $socket;

	// --------------------------------------------------------------------------------------------
	// ~ Constructor
	// --------------------------------------------------------------------------------------------

	/**
	 * @param string $server
	 * @throws \Exception
	 */
	public function __construct($server)
	{
		$this->server = $server;
		$urlParts = parse_url($this->server);
		if (!isset($urlParts['port'])) {
			trigger_error('you have to specify a port, because there is no std port for me', E_USER_ERROR);
		}
		if (!isset($urlParts['host'])) {
			trigger_error('i am missing a host to connect to in :' . $this->server, E_USER_ERROR);
		}
		$address = gethostbyname($urlParts['host']);
		$this->socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
		if ($this->socket === false) {
			trigger_error('failed to create socket: ' . socket_strerror(socket_last_error()), E_USER_ERROR);
		}
        $connected = socket_connect($this->socket, $address, $urlParts['port']);
        if ($connected === false) {
            trigger_error("could not connect to nodeflock server", E_USER_ERROR);
        }
	}

	/**
	 * @param $func
	 * @param $args
	 * @return mixed|void
	 */
	public function call($func, $args = [])
	{
        return $this->send([
            'func' => $func,
            'args' => $args
        ]);
	}

	/**
	 * @param $call
	 * @return void|mixed
	 */
	private function send($call)
	{
        $rawData = json_encode($call);
		$sendBytes = strlen($rawData) . $rawData;
		$bytesWritten = socket_write($this->socket, $sendBytes, strlen($sendBytes));
		if ($bytesWritten != strlen($sendBytes)) {
			trigger_error('failed to write my bytes', E_USER_ERROR);
		}
		$bytesRead = 0;
		$bytesToRead = -1;
		$msg = '';
		$window = 1;
		while (false !== $incoming = socket_read($this->socket, $window)) {
			if ($bytesToRead < 0) {
				if ($incoming == '{') {
					$bytesToRead = ((int) $msg) - 1;
					$msg = '{';
					$window = self::SOCKET_READ_WINDOW_SIZE;
				} else {
					$msg .= $incoming;
				}
			} else {
				$bytesRead += strlen($incoming);
				$msg .= $incoming;
				if ($bytesRead == $bytesToRead) {
					return json_decode($msg);
				}
			}
		}
        trigger_error("failed to read response", E_USER_ERROR);
	}

	// --------------------------------------------------------------------------------------------
	// ~ Magic methods
	// --------------------------------------------------------------------------------------------

	public function __destroy()
	{
		socket_close($this->socket);
	}
}




$start = microtime(true);

$client = new Client('tcp://127.0.0.1:8888');

for($i = 0;$i<100;$i++) {
	$sample = $client->call('MZG.Components.demo', [[
		"name" => "jan",
		"age" => 10+$i,
		"andisWille" => "ölalla"
	]]);
}
echo json_encode([
	'requests'  => $i,
	's total'   => $total = (microtime(true) - $start), 
	's per req' => $total / $i, $sample
], JSON_PRETTY_PRINT);	
