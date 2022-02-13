| :exclamation:  **This software is currently in alpha phase.**   |
|-----------------------------------------------------------------|

<div align="center">
	<a href="https://github.com/pushbits/logo">
		<img height="200px" src="https://raw.githubusercontent.com/pushbits/server/master/.github/logo.png" />
	</a>
</div>

<div align="center">
	<h1>PushBits</h1>
	<h4 align="center">
		Receive your important notifications immediately, over <a href="https://matrix.org/">Matrix</a>.
	</h4>
	<p>PushBits enables you to send push notifications via a simple web API, and delivers them to your users.</p>
</div>

<p align="center">
	<a href="https://github.com/pushbits/server/actions"><img alt="Build status" src="https://img.shields.io/github/workflow/status/pushbits/server/Main"/></a>&nbsp;
	<a href="https://www.pushbits.io/docs/"><img alt="Documentation" src="https://img.shields.io/badge/docs-online-success"/></a>&nbsp;
	<a href="https://www.pushbits.io/api/"><img alt="API Documentation" src="https://img.shields.io/badge/api docs-online-success"/></a>&nbsp;
	<a href="https://matrix.to/#/#pushbits:matrix.org"><img alt="Matrix" src="https://img.shields.io/matrix/pushbits:matrix.org"/></a>&nbsp;
	<!--<a href="https://github.com/pushbits/server/releases/latest"><img alt="Latest release" src="https://img.shields.io/github/release/pushbits/server"/></a>&nbsp;-->
	<a href="https://github.com/pushbits/server/blob/master/LICENSE"><img alt="License" src="https://img.shields.io/github/license/pushbits/server"/></a>
</p>

## 💡&nbsp;About

PushBits is a relay server for push notifications.
It enables you to send notifications via a simple web API, and delivers them to you through [Matrix](https://matrix.org/).
This is similar to what [Pushover](https://pushover.net/) and [Gotify](https://gotify.net/) offer, but it does not require an additional app.

The vision is to have compatibility with Gotify on the sending side, while on the receiving side an established service is used.
This has the advantages that
- sending plugins written for Gotify (like those for [Watchtower](https://containrrr.dev/watchtower/) and [Jellyfin](https://jellyfin.org/)) as well as
- receiving clients written for Matrix
can be reused.

### Why Matrix instead of X?

This project totally would've used Signal if it would offer a proper API.
Sadly, neither [Signal](https://signal.org/) nor [WhatsApp](https://www.whatsapp.com/) come with an API (at the time of writing) through which PushBits could interact.

In [Telegram](https://telegram.org/) there is an API to run bots, but these are limited in that they cannot create chats by themselves.
If you insist on going with Telegram, have a look at [webhook2telegram](https://github.com/muety/webhook2telegram).

The idea of a federated, synchronized but yet end-to-end encrypted protocol is awesome, but its clients simply aren't really there yet.
Still, if you haven't tried it yet, we'd encourage you to check it out.

## 🤘&nbsp;Features

- [x] Multiple users and multiple channels (applications) per user
- [x] Compatibility with Gotify's API for sending messages
- [x] API and CLI for managing users and applications
- [x] Optional check for weak passwords using [HIBP](https://haveibeenpwned.com/)
- [x] Argon2 as KDF for password storage
- [ ] Two-factor authentication, [issue](https://github.com/pushbits/server/issues/19)
- [ ] Bi-directional key verification, [issue](https://github.com/pushbits/server/issues/20)

## 🚀&nbsp;Installation

PushBits is meant to be self-hosted.
That means you have to install it on your own server.

Currently, the only supported way of installing PushBits is via [Docker](https://www.docker.com/) or [Podman](https://podman.io/).
The image is hosted [here on Docker Hub](https://hub.docker.com/r/eikendev/pushbits).

| :warning:  **You are advised to install PushBits behind a reverse proxy and enable TLS.** Otherwise, your credentials will be transmitted unencrypted.   |
|----------------------------------------------------------------------------------------------------------------------------------------------------------|

## ⚙&nbsp;Configuration

To see what can be configured, have a look at the `config.sample.yml` file inside the root of the repository.

Settings can optionally be provided via environment variables.
The name of the environment variable is composed of a starting `PUSHBITS_`, followed by the keys of the setting, all
joined with `_`.
As an example, the HTTP port can be provided as an environment variable called `PUSHBITS_HTTP_PORT`.

To get started, here is a Docker Compose file you can use.
```yaml
version: '2'

services:
    server:
        image: eikendev/pushbits:latest
        ports:
            - 8080:8080
        environment:
            PUSHBITS_DATABASE_DIALECT: 'sqlite3'
            PUSHBITS_ADMIN_MATRIXID: '@your/matrix/username:matrix.org' # The Matrix account on which the admin will receive their notifications.
            PUSHBITS_ADMIN_PASSWORD: 'your/pushbits/password' # The login password of the admin account. Default username is 'admin'.
            PUSHBITS_MATRIX_USERNAME: 'your/matrix/username' # The Matrix account from which notifications are sent to all users.
            PUSHBITS_MATRIX_PASSWORD: 'your/matrix/password' # The password of the above account.
        volumes:
            - /etc/localtime:/etc/localtime:ro
            - /etc/timezone:/etc/timezone:ro
            - ./data:/data
```

In this example, the configuration file would be located at `./data/config.yml` on the host.
The SQLite database would be written to `./data/pushbits.db`.
**Don't forget to adjust the permissions** of the `./data` directory, otherwise PushBits will fail to operate.

## 📄&nbsp;Usage

Now, how can you interact with the server?
We provide [a little CLI tool called pbcli](https://github.com/PushBits/cli) to make basic API requests to the server.
It helps you to create new users and applications.
You will find further instructions in the linked repository.

At the time of writing, there is no fancy GUI built-in, and we're not sure if this is necessary at all.
Currently, we would like to avoid front end development, so if you want to contribute in this regard we're happy if you reach out!

After you have created a user and an application, you can use the API to send a push notification to your Matrix account.

```bash
curl \
	--header "Content-Type: application/json" \
	--request POST \
	--data '{"message":"my message","title":"my title"}' \
	"https://pushbits.example.com/message?token=$PB_TOKEN"
```

Note that the token is associated with your application and has to be kept secret.
You can retrieve the token using [pbcli](https://github.com/PushBits/cli) by running following command.

```bash
pbcli application show $PB_APPLICATION --url https://pushbits.example.com --username $PB_USERNAME
```

### Authentication

Pushbits offers you two methods of authenticating against the server:

* Basic authentication (`basic`)
* [Oauth 2.0](https://oauth.net/2/) (`oauth`)

You will find the corresponding setting in the security section.

```yaml
...
security:
    ...
    # The authentication method to use
    authentication: basic
...
```

#### Basic authentication

For [basic authentication](https://en.wikipedia.org/wiki/Basic_access_authentication) you have to provide your username and password in each request to the server. For example in curl you can do this with the `--user` flag:

```bash
curl -u myusername:totallysecretpassword
```

#### Oauth 2.0

[Oauth 2.0](https://en.wikipedia.org/wiki/OAuth) is a token based authentication method. Instead of passing your password with each request you request a token from an authorization server. With this token you are then able to authenticate yourself against the PushBits server. 

Make sure to setup the "oauth" section in the config file correctly.

##### Authenticating

For authentication use the ``/oauth2/auth` endpoint. E.g.:

```bash
curl \
	--header "Content-Type: application/json" \
	--request POST \
	"https://pushbits.example.com/oauth2/auth" -d "client_id=000000&username=admin&password=1233456&response_type=code&redirect_uri=https://myapp.example.com"
```

This will return a HTTP redirect with the status code `302` and an authentication code set as parameter:

```
HTTP/2 302
date: Sun, 23 May 2021 10:33:27 GMT
location: https://myapp.example.com?code=4T1TJXMBPTOS4NNGILBDYW
content-length: 0
```

Your app then needs to use this code to trade it for a access token. 

**Hint for command line users:** you can extract the authentication code from the redirect without the need of a running webserver.

##### Receiving an access token

You can get an access token from the `/oauth/token` endpoint. There are several methods, so called "grant types" for receiving a token. PushBits currently supports the following one's:

* Refresh 
* Authentication code

Oauth 2.0 authentication is based on "clients", thus you need to provide identifiers for a client with your request. These are the `client_id` and the `client_secret`. 

For your first token you will need a authentication code, see the section above. Then use it like this: 

```bash
curl \
	--header "Content-Type: application/json" \
	--request POST \
	"https://pushbits.example.com/oauth2/token" -d "grant_type=authorization_code&client_id=000000&client_secret=49gjg4js9&response_type=token&redirect_uri=https://myapp.example.com&code=OP1Q2UJEVL-RPR9GZAUURA"
```

This will then return an access token and refresh token for you. 

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIwMDAwMDAiLCJleHAiOjE2MjE4NTU3ODcsInN1YiI6IjEifQ.jMux7CBw6fY15Ohc8exEbcnUiMBVVgCowvq3rMrw7MQ",
  "expires_in": 86400,
  "refresh_token": "OP1Q2UJEVL-RPR9GZAUURA",
  "token_type": "Bearer"
}
```

The access token is short lived, the refresh token is long lived, but can not be used for authentication. If your access token runs out, you can use the refresh token to generate a new access token:

```bash
curl \
	--header "Content-Type: application/json" \
	--request POST \
	"https://pushbits.example.com/oauth2/token" -d "grant_type=refresh_token&client_id=000000&client_secret=49gjg4js9&response_type=token&refresh_token=OP1Q2UJEVL-RPR9GZAUURA"
```

##### Getting information about a access token

With a valid access token you can get information about it from `/oauth/tokeninfo`. This is meant for testing if a token is issued correctly.

```bash
curl \
	--header "Content-Type: application/json" \
	--request GET \
    --header "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIwMDAwMDAiLCJleHAiOjE2MjE4NTU3ODcsInN1YiI6IjEifQ.jMux7CBw6fY15Ohc8exEbcnUiMBVVgCowvq3rMrw7MQ" \
	"https://pushbits.example.com/oauth2/tokeninfo"
```

##### Revoking a token

Admin users are eligible to revoke tokens. This should not be necessary in normal operation, as tokens are only short lived. But there might be situations where attackers might have gotten knowledge about a token. 

```bash
curl \
	--header "Content-Type: application/json" \
	--request POST \
    --header "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIwMDAwMDAiLCJleHAiOjE2MjE4NTU3ODcsInN1YiI6IjEifQ.jMux7CBw6fY15Ohc8exEbcnUiMBVVgCowvq3rMrw7MQ" \
    --data '{"access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIwMDAwMDAiLCJleHAiOjE2MjE4NDg1MDYsInN1YiI6IjEifQ.cO0_8fqsJDG4KswjC0CSzc_EznntH-FDQejdolPAISo"}' \
	"https://pushbits.example.com/oauth2/revoke"
```

##### Requesting a longterm token

Longterm tokens are tokens that life for multiple years. They can be used for scripts and other software that access PushBits. So the other software does not need knowledge about the actuall password of the user. However be carefull with longterm tokens, if you loose one others might be able to perform actions on your user account.

```bash
curl \
    --request POST \
    --header "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIwMDAwMDAiLCJleHAiOjE2MjE4NTU3ODcsInN1YiI6IjEifQ.jMux7CBw6fY15Ohc8exEbcnUiMBVVgCowvq3rMrw7MQ" \
    --data '{"client_id": "000000", "client_secret": "49gjg4js9"}' \
    "https://push.remote.alexanderebhart.de/oauth2/longtermtoken" 
```

### Message options

Messages can be specified in three different syntaxes:

* `text/plain`
* `text/html`
* `text/markdown`

To set a specific syntax you need to set the `extras` parameter ([inspired by Gotify's message extras](https://gotify.net/docs/msgextras#clientdisplay)):

```bash
curl \
	--header "Content-Type: application/json" \
	--request POST \
	--data '{"message":"my message with\n\n**Markdown** _support_.","title":"my title","extras":{"client::display":{"contentType": "text/markdown"}}}' \
	"https://pushbits.example.com/message?token=$PB_TOKEN"
```

HTML content might not be fully rendered in your Matrix client; see the corresponding [Matrix specs](https://spec.matrix.org/unstable/client-server-api/#mroommessage-msgtypes).
This also holds for Markdown, as it is translated into the corresponding HTML syntax.

### Deleting a Message

You can delete a message, this will send a notification in response to the original message informing you that the message is "deleted".

To delete a message, you need its message ID  which is provided as part of the response when you send the message.
The ID might contain characters not valid in URIs.
We hence provide an additional `id_url_encoded` field for messages; you can directly use it when deleting a message without performing encoding yourself.

```bash
curl \
	--request DELETE \
	"https://pushbits.example.com/message/${MESSAGE_ID}?token=$PB_TOKEN"
```

## 👮&nbsp;Acknowledgments

The idea for this software and most parts of the initial source are heavily inspired by [Gotify](https://gotify.net/).
Many thanks to [jmattheis](https://jmattheis.de/) for his well-structured code.

## 💻&nbsp;Development

The source code is located on [GitHub](https://github.com/pushbits/server).
You can retrieve it by checking out the repository as follows.

```bash
git clone https://github.com/pushbits/server.git
```

[![Stargazers over time](https://starchart.cc/pushbits/server.svg)](https://starchart.cc/pushbits/server)
