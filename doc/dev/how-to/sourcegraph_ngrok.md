# Run a local Sourcegraph instance behind Tunnelmole or ngrok

Sometimes it's useful to have the Sourcegraph instance you're running on your
local machine to be reachable over the internet. If you're testing webhooks, for
example, where a code host needs to be able to send requests to your instances.

We provide two alternatives for that:
- [Tunnelmole](https://github.com/robbie-cahill/tunnelmole-client) an open-source tunneling tool.
- [ngrok](https://ngrok.io/), a popular closed-source reverse proxy.
Both allow you to expose your instance to the internet.

## Setting up with Tunnelmole

1. Install `Tunnelmole`. For Linux, Mac and Windows Subsystem for Linux, use the following command:
```
curl -O https://tunnelmole.com/sh/install.sh && sudo bash install.sh
```
For Windows without WSL, [Download tmole.exe](https://tunnelmole.com/downloads/tmole.exe) and put it somewhere in your [PATH](https://www.wikihow.com/Change-the-PATH-Environment-Variable-on-Windows).

2. Start your Sourcegraph instance: `sg start`
3. Start `Tunnelmole` and point it at Sourcegraph: `tmole 3080`
4. Copy the `Forwarding` URL `Tunnelmole` displays. e.g.: `http://bvdo5f-ip-49-183-170-144.tunnelmole.net`
5. Edit your site-config (i.e. `../dev-private/enterprise/dev/site-config.json`) and update the `"externalURL"` to point to your Tunnelmole: `"externalURL": "http://bvdo5f-ip-49-183-170-144.tunnelmole.net"`
6. Open the Tunnelmole URL in a browser to make sure you see your instance
7. Done

## Setting up with ngrok
1. Install `ngrok`: `brew install ngrok`
2. Authenticate the `ngrok` if this is your first time running it (the token can be obtained from [Ngrok dashboard](https://dashboard.ngrok.com/get-started/setup)): `ngrok config add-authtoken <Your token>`
3. Start your Sourcegraph instance: `sg start`
4. Start `ngrok` and point it at Sourcegraph: `ngrok http --host-header=rewrite 3080`
5. Copy the `Forwarding` URL `ngrok` displays. e.g.: `https://630b-87-170-68-206.eu.ngrok.io`
6. Edit your site-config (i.e. `../dev-private/enterprise/dev/site-config.json`) and update the `"externalURL"` to point to your ngrok: `"externalURL": "https://630b-87-170-68-206.eu.ngrok.io"`
7. Open the ngrok URL in browser to make sure you see your instance
8. Done
