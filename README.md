# Flow-Contract-Bot

An bot you can start to spy on testnet/mainnet deploys. 

Has the following configuration options
  - DISCORD_WEBHOOK_URL: an URL from a discrod server to post messages to. (required)
	- NETWORK: the flow network to listen for events in testnet|mainnet (required)
	- EVENTS_FILE: the name of the events file to track events in, defulat:<network>.events in same dir as binary
	- SLEEP: a string duration of how long to sleep between each run, default:1m


An example systemd file is provided
