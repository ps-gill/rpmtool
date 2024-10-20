# rpmtool

A CLI for rpm packages.

## Building

Install `go`, `gcc` and `rpm-devel`.

```sh
dnf install golang gcc rpm-devel
```

Build `rpmtool` exectuable by running `go build`.

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
$ ./rpmtool build --help
Build package from a .spec file

Usage:
  rpmtool build [spec] [flags]

Flags:
      --gpg-key string              gpg key
      --gpg-key-id string           gpg key Id
      --gpg-key-passphrase string   gpg key passphrase
  -h, --help                        help for build
      --latest-deps                 install latest build dependencies
      --skip-deps                   skip build dependencies installation
      --srpm                        build srpm instead of rpm
```

## License

See [LICENSE](LICENSE)
