async function init() {
	/** @typedef {{action: "play" | "pause" | "sync", time: number}} MessageData */

	const conn = new WebSocket(
		`ws://${window.location.host}/ws/${window.location.pathname.split("/").at(-1)}`,
	);

	/** @param {MessageData} data  */
	function send(data) {
		console.log("Sending:", data);
		conn.send(JSON.stringify(data));
	}

	conn.addEventListener("open", () => {
		send({
			action: "sync",
			time: 0,
		});
	});

	const player = await document.querySelector("lite-youtube").getYTPlayer();

	let isRemoteUpdate = false;

	player.addEventListener("onStateChange", (e) => {
		if (!isRemoteUpdate) {
			switch (e.data) {
				case 1:
					send({ action: "play", time: player.getCurrentTime() });
					break;
				case 2:
					send({ action: "pause", time: player.getCurrentTime() });
					break;
			}
		}
	});

	conn.addEventListener("message", (e) => {
		/** @type {MessageData} msg */
		const msg = JSON.parse(e.data);
		console.log("Received:", msg);

		isRemoteUpdate = true;

		switch (msg.action) {
			case "play":
				player.seekTo(msg.time, true);
				player.playVideo();
				break;
			case "pause":
				player.seekTo(msg.time, true);
				player.pauseVideo();
				break;
			case "sync":
				player.seekTo(msg.time, true);
				player.pauseVideo();
				break;
		}

		setTimeout(() => {
			isRemoteUpdate = false;
		}, 100);
	});
}

init().catch(console.error);
