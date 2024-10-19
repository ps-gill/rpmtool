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
$ rpmtool build --help
Build package from a .spec file

Usage:
  rpmtool build [spec] [flags]

Flags:
  -h, --help          help for build
      --latest-deps   install latest build dependencies
      --skip-deps     skip build dependencies installation
      --srpm          build srpm instead of rpm
```

## License

See [LICENSE](LICENSE)
