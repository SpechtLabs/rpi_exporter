# rpi_exporter

> A Raspberry Pi CPU temperature exporter. - by **[Lukas Malkmus]** (_fork by [cedi]_)

[![go report_badge]][report]
[![release_badge]][release page]
[![license_badge]][license]

[![ci_badge]][ci]
[![release_workflow_badge]][release workflow]

---

## Table of Contents

1. [Introduction](#introduction)
2. [Usage](#usage)
3. [Contributing](#contributing)
4. [License](#license)

### Introduction

The _rpi_exporter_ is a simple server that scrapes the Raspberry Pi's CPU
temperature and exports it via HTTP for Prometheus consumption.

### Usage

#### Installation

The easiest way to run the _rpi_exporter_ is by grabbing the latest binary from
the [release page].

Do not forget to run _rpi_exporter_ using user in `video` group to get GPU
details from RPi.

The container image `ghcr.io/spechtlabs/rpi_exporter` is published for every
release (`latest` and the release tag) and for every commit to `main` (`main`):

```bash
docker run -d -p 9243:9243 ghcr.io/spechtlabs/rpi_exporter:latest
```

##### Building from source

The toolchain is pinned with [mise]:

```bash
git clone https://github.com/SpechtLabs/rpi_exporter.git
cd rpi_exporter
mise install
mise run build
```

`mise run check` runs the linters and tests CI runs.

#### Using the application

```bash
./rpi_exporter [flags]
```

Help on flags:

```bash
./rpi_exporter --help
```

### Contributing

Feel free to submit PRs or to fill Issues. Every kind of help is appreciated.

### License

© Lukas Malkmus, 2019

Distributed under Apache License (`Apache License, Version 2.0`).

See [LICENSE](LICENSE) for more information.

<!-- Links -->
[mise]: https://mise.jdx.dev
[Lukas Malkmus]: https://github.com/lukasmalkmus
[cedi]: https://github.com/cedi

<!-- Badges -->
[go report_badge]: https://goreportcard.com/badge/github.com/spechtlabs/rpi_exporter
[report]: https://goreportcard.com/report/github.com/spechtlabs/rpi_exporter
[release page]: https://github.com/SpechtLabs/rpi_exporter/releases
[release_badge]: https://img.shields.io/github/release/cedi/rpi_exporter.svg
[license]: https://opensource.org/licenses/Apache-2.0
[license_badge]: https://img.shields.io/badge/license-Apache-blue.svg
[ci_badge]: https://github.com/SpechtLabs/rpi_exporter/actions/workflows/ci.yaml/badge.svg
[ci]: https://github.com/SpechtLabs/rpi_exporter/actions/workflows/ci.yaml
[release_workflow_badge]: https://github.com/SpechtLabs/rpi_exporter/actions/workflows/release.yaml/badge.svg?branch=main
[release workflow]: https://github.com/SpechtLabs/rpi_exporter/actions/workflows/release.yaml
