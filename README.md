# rpmtool

A CLI for rpm packages.

## Building

Install `go`, `gcc`, `ld` and `rpm-devel`.

```sh
dnf install binutils golang gcc rpm-devel
```

Build `rpmtool` exectuable by running `go build`.

Also, there is a rpm spec available [here](https://gitlab.com/pgill/rpmtool-rpm).

## Usage

```sh
$ ./rpmtool tools --help
Check if required tools are available

Usage:
  rpmtool tools [flags]

Flags:
      --exclude-signature   exclude signature tools
  -h, --help                help for tools
```

```sh
$ rpmtool build --help
Build package from a .spec file

Usage:
  rpmtool build [spec] [flags]

Flags:
  -h, --help                         help for build
      --key string                   pgp key
      --key-id string                pgp key Id
      --key-passphrase-file string   pgp key passphrase
      --latest-deps                  install latest build dependencies
      --skip-deps                    skip build dependencies installation
      --srpm                         build srpm instead of rpm
```

Sequoia PGP's [`sq`](https://gitlab.com/sequoia-pgp/sequoia-sq) cli is required for signatures.

## License

See [LICENSE](LICENSE)
