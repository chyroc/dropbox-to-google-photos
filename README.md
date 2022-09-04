# dropbox-to-google-photos
save dropbox pictures and videos to google photos

## Installation

```bash
go install github.com/chyroc/dropbox-to-google-photos@latest
```

## Usage

### Init config

```bash
dropbox-to-google-photos init
```

default config file is `~/.dropbox-to-google-photos/config.json`

open config file and fill in the blanks

```json
{
  "account": "someaccount@gmail.com",
  "google_photos": {
    "client_id": "client id",
    "client_secret": "client secret"
  },
  "dropbox":{
    "token": "dropbox token",
    "root_dir": "/"
  },
  "worker": 20
}
```

### Auth google photos

```bash
dropbox-to-google-photos auth
```

### Sync dropbox to google photos

```bash
dropbox-to-google-photos sync
```
