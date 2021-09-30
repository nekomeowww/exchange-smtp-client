# Microsoft Exchange SMTP Example based on OAuth of Microsoft Graph

## Implementation of XOAuth2 Authentication

### Implementation

```golang
// SMTPAuth SMTPAuth
type SMTPAuth struct {
	accessToken string
}

// Auth auth
func Auth(token string) smtp.Auth {
	return &SMTPAuth{token}
}

// Start start
func (a *SMTPAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "xoauth2", []byte(""), nil
}

// Next next
func (a *SMTPAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return []byte(a.accessToken), nil
	}
	return nil, nil
}
```

### Explaination

Start a connection by running command:
```shell
$ openssl s_client -starttls smtp -ign_eof -crlf -connect smtp.office365.com:587
```

#### 1. Connection established, send `HELO` or `EHLO` command to server

##### openssl connection example

```
[... SSL handshake...]

250 SMTPUTF8
HELO <your email host> # Send HELO command to server
250 BY3PR04CA0028.outlook.office365.com Hello [134.195.101.47]
```

##### client implementation explaination

The connection establishment has been implemented by net/smtp itself, we only need to implement the auth part by ourselves.

#### 2. Send `AUTH XOAUTH2` command to SMTP server

##### openssl connection example

```
AUTH XOAUTH2
334 # server sends code `334` with message `(empty)` to client
```

##### client implementation explaination

```golang
func (a *SMTPAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "xoauth2", []byte(""), nil // "xoauth2" is equivalent to "XOAUTH2"
}
```

#### 3. Send base64 encoded access token to SMTP server

##### openssl connection example

Code 334 represents: Server challenge - the text part contains the Base64-encoded challenge ([SMTP return code reference](https://en.wikipedia.org/wiki/List_of_SMTP_server_return_codes))

```
334
<base64 token>
235 2.7.0 Authentication successful
```

##### client implementation explaination

```golang
// at this point, SMTP server tells us that we need to take a challenge (provide our access token)
// fromServer represents the message follow along with "334"
// more tells us that one more step is required in order to complete the authentication (or flow)
func (a *SMTPAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
        // sends the access token to the server
		return []byte(a.accessToken), nil
	}
	return nil, nil
}
```

### Brief Summary

Entire authentication flow:
```
HELO <your email host>
250 BY3PR04CA0028.outlook.office365.com Hello [134.195.101.47] 
AUTH XOAUTH2
334
<base64 token>
235 2.7.0 Authentication successful
```
