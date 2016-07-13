# fast-cli

fast-cli estimates your current internet download speed by performing a series of downloads from Netflix's fast.com servers.


### Installing

###### Compile
This project requires go 1.6+ to compile. Just run `go get -u github.com/gesquive/fast-cli` and the executable should be built for you automatically in your `$GOPATH`.

###### Download
Alternately, you can download the latest release for your platform from [github](https://github.com/gesquive/fast-cli/releases).

Once you have an executable, make sure to copy it somewhere on your path like `/usr/local/bin` or `C:/Program Files/`.
If on a \*nix/mac system, make sure to run `chmod +x /path/to/fast-cli`.

## Usage

```console
fast-cli estimates your current internet download speed by performing a series of downloads from Netflix's fast.com servers.

Usage:
  fast-cli [flags]

Flags:
  -s, --use-https   Use HTTPS when connecting
      --version     Display the version number and exit
```
Optionally, a hidden debug flag is available in case you need additional output.
```console
Hidden Flags:
  -D, --debug                  Include debug statements in log output
```

## Documentation

This documentation can be found at github.com/gesquive/digitalocean-ddns

## License

This package is made available under an MIT-style license. See LICENSE.

## Contributing

PRs are always welcome!
